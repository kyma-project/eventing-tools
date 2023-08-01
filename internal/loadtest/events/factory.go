package events

import (
	"regexp"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha2"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
)

type eventGenerator map[string]*Generator

type Factory struct {
	generators map[NamespaceName]eventGenerator
	senderC    chan<- Event
}

func (f *Factory) OnNewSubscription(subscription *unstructured.Unstructured) {
	f.FromSubscription(subscription)
}

func (f *Factory) OnChangedSubscription(subscription *unstructured.Unstructured) {
	f.FromSubscription(subscription)
}

func (f *Factory) OnDeleteSubscription(subscription *unstructured.Unstructured) {
	f.FromSubscription(subscription)
}

func NewGeneratorFactory(senderC chan<- Event) *Factory {
	return &Factory{
		generators: map[NamespaceName]eventGenerator{},
		senderC:    senderC,
	}
}

type NamespaceName struct {
	Name, Namespace string
}

const (
	formatLabel = "eventing-loadtest"
)

var _ subscription.Notifiable = &Factory{}

func (f *Factory) FromSubscription(subscription *unstructured.Unstructured) error {
	sub, err := v1alpha2.ToSubscription(subscription)
	if err != nil {
		return err
	}
	return f.reconcile(sub)
}

func (f *Factory) reconcile(sub *v1alpha2.Subscription) error {
	// for now we support only type matching standard
	if sub.Spec.TypeMatching != v1alpha2.Standard {
		return nil
	}
	nn := NamespaceName{
		Name:      sub.Name,
		Namespace: sub.Namespace,
	}

	// delete subscription
	if sub.DeletionTimestamp != nil {
		if g, ok := f.generators[nn]; ok {
			f.stopGenerators(g)
			delete(f.generators, nn)
			return nil
		}
	}

	// create default eventGenerator (empty)
	if _, ok := f.generators[nn]; !ok {
		f.generators[nn] = eventGenerator{}
	}

	eg := f.generators[nn]

	// check if eventtypes have been removed
currentGenerator:
	for etgen, gen := range eg {
		for _, et := range EventTypeFromSubscription(sub) {
			if et == etgen {
				continue currentGenerator
			}
		}
		// remove generators for removed eventType
		gen.Stop()
		delete(eg, etgen)
	}

	// handle adding EventTypes
currentEventType:
	for _, et := range EventTypeFromSubscription(sub) {
		//for etgen, gen := range eg {
		for etgen, gen := range eg {
			if et == etgen {
				// update the eventFormat to the one specified in the subscription
				updateGeneratorFormat(sub, gen)
				continue currentEventType
			}
		}
		// create a new EventGenerator for the found eventType and start it
		gen := f.ConfigureAndStartGenerator(sub, et)
		if gen != nil {
			eg[et] = gen
		}
	}
	return nil
}

func (f *Factory) Stop() {
	for _, gens := range f.generators {
		for _, gen := range gens {
			gen.Stop()
		}
	}
}

func (f *Factory) ConfigureAndStartGenerator(sub *v1alpha2.Subscription, eventType string) *Generator {
	re := regexp.MustCompile(`.v(\d+)$`)
	if !re.MatchString(eventType) {
		return nil
	}
	rs := re.FindStringSubmatch(eventType)[1]
	rate, err := strconv.Atoi(rs)
	if err != nil {
		return nil
	}
	gen := NewGenerator(eventType, sub.Spec.Source, rate, sub.GetLabels()[formatLabel], f.senderC) // create a clean `v<eps>` version indicator.
	gen.Start()
	return gen
}

func (f *Factory) stopGenerators(generators eventGenerator) {
	for _, g := range generators {
		g.Stop()
	}
}

func EventTypeFromSubscription(sub *v1alpha2.Subscription) []string {
	var ets []string
	for _, t := range sub.Spec.Types {
		ets = append(ets, t)
	}
	return ets
}
