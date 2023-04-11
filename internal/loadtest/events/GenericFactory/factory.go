package GenericFactory

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha2"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
)

type GenericEventFactory struct {
}

const (
	formatLabel = "eventing-loadtest"
)

func New() *GenericEventFactory {
	return &GenericEventFactory{}
}

func (g *GenericEventFactory) FromSubscription(subscription *unstructured.Unstructured, eventFormat string) []*GenericEvent.Event {
	sub, err := v1alpha2.ToSubscription(subscription)
	if err != nil {
		return nil
	}
	events := []*GenericEvent.Event{}

	// for now we support only type matching standard
	if sub.Spec.TypeMatching != v1alpha2.Standard {
		return events
	}
	if !sub.Status.Ready {
		return events
	}

	if sub.GetLabels()[formatLabel] != eventFormat {
		return events
	}

	re := regexp.MustCompile(`.v(\d+)$`)
	for _, et := range sub.Spec.Types {
		if !re.MatchString(et) {
			continue
		}
		rss := re.FindStringSubmatch(et)
		for _, rs := range rss {
			rate, err := strconv.Atoi(rs)
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(et, fmt.Sprintf(".v%v", rs))                                             // trim the `.v<eps>` to get a clean event type without trailing `.`.
			events = append(events, GenericEvent.NewEvent(fmt.Sprintf("v%v", rs), name, sub.Spec.Source, rate)) // create a clean `v<eps>` version indicator.
		}
	}
	return events
}
