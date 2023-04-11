package cloudevent

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/eventing-tools/internal/client/cloudevents"
	"github.com/kyma-project/eventing-tools/internal/client/transport"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericFactory"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
)

// compile-time check for interfaces implementation.
var _ sender.Sender = &Sender{}
var _ subscription.Notifiable = &Sender{}

// Sender sends cloud events.
type Sender struct {
	ctx         context.Context
	cancel      context.CancelFunc
	client      client.Client
	config      *config.Config
	events      map[string][]*GenericEvent.Event
	factory     events.EventFactory
	endpoint    string
	process     chan bool
	running     bool
	undelivered int32
	ack         int32
	nack        int32
	mapLock     sync.RWMutex
	wg          sync.WaitGroup
	stopper     sync.Mutex
}

func NewSender(conf *config.Config) *Sender {
	s := &Sender{config: conf}
	s.undelivered = 0
	s.ack = 0
	s.nack = 0
	s.events = make(map[string][]*GenericEvent.Event)
	s.factory = GenericFactory.New()

	return s
}

func (s *Sender) NotifyAdd(cm *corev1.ConfigMap) {
	s.stopper.Lock()
	defer s.stopper.Unlock()
	s.stop()
	config.Map(cm, s.config)
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = cloudevents.NewClientOrDie(t)
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.process = make(chan bool, s.config.EpsLimit)
	s.endpoint = fmt.Sprintf("%s/publish", s.config.PublishEndpoint)
	s.start()
}

func (s *Sender) NotifyUpdate(cm *corev1.ConfigMap) {
	s.stopper.Lock()
	defer s.stopper.Unlock()
	s.stop()
	config.Map(cm, s.config)
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = cloudevents.NewClientOrDie(t)
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.process = make(chan bool, s.config.EpsLimit)
	s.endpoint = fmt.Sprintf("%s/publish", s.config.PublishEndpoint)
	s.start()
}

func (s *Sender) NotifyDelete(*corev1.ConfigMap) {
	s.stopper.Lock()
	defer s.stopper.Unlock()
	s.stop()
}

func (s *Sender) OnNewSubscription(subscription *unstructured.Unstructured) {
	if subscription.GetDeletionTimestamp() != nil {
		return
	}
	fmt.Print(subscription)
	ne := s.factory.FromSubscription(subscription, events.CloudeventFormat)
	if len(ne) == 0 {
		return
	}
	s.stopper.Lock()
	defer s.stopper.Unlock()

	log.Printf("Starting Cloud Event Sender")
	defer log.Println("done starting")

	// s.queue = make(chan events.Event, buffer)
	for _, e := range ne {
		e.Start()
	}
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	s.events[fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName())] = ne
}

func (s *Sender) OnChangedSubscription(subscription *unstructured.Unstructured) {
	if subscription.GetDeletionTimestamp() != nil {
		return
	}
	s.stopper.Lock()
	defer s.stopper.Unlock()
	log.Printf("updating Cloud Event Sender")
	defer log.Println("done updating")
	for _, e := range s.events[fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName())] {
		e.Stop()
	}
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	delete(s.events, fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName()))

	ne := s.factory.FromSubscription(subscription, events.CloudeventFormat)
	if len(ne) == 0 {
		return
	}

	for _, e := range ne {
		e.Start()
	}
	s.events[fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName())] = ne
}

func (s *Sender) OnDeleteSubscription(sub *unstructured.Unstructured) {
	s.stopper.Lock()
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	defer s.stopper.Unlock()
	for _, e := range s.events[fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName())] {
		e.Stop()
	}
	delete(s.events, fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName()))
}

func (s *Sender) init() {
}

func (s *Sender) start() {
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.mapLock.RLock()
	for _, subs := range s.events {
		for _, e := range subs {
			e.Start()
		}
	}
	s.mapLock.RUnlock()
	s.sendEventsAsync()
	s.wg.Add(1)
	go s.refillMaxEps(time.Second)
	s.wg.Add(1)
	go s.reportUsageAsync(time.Second, 20*time.Second)
}

func (s *Sender) stop() {
	// recover from closing already closed channels
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from: ", r)
		}
	}()
	s.mapLock.RLock()
	for _, subs := range s.events {
		for _, e := range subs {
			e.Stop()
		}
	}
	s.mapLock.RUnlock()

	s.running = false
	s.cancel()
	close(s.process)
	s.wg.Wait()
}

func (s *Sender) sendEventsAsync() {
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go s.sendEvents()
	}
}

func (s *Sender) reportUsageAsync(send, success time.Duration) {

	defer func() {
		s.wg.Done()
	}()

	sendt := time.NewTicker(send)
	defer sendt.Stop()
	succt := time.NewTicker(success)
	defer succt.Stop()

	for {
		select {
		case <-s.ctx.Done():
			targetEPS := s.ComputeTotalEventsPerSecond()
			log.Printf(
				"cloud events: | target_eps:% 4d (% 4d) | undelivered:% 4d | ack:% 4d | nack:% 4d | sum:% 4d |",
				targetEPS, s.config.EpsLimit, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack,
			)
			return
		case <-sendt.C:
			targetEPS := s.ComputeTotalEventsPerSecond()
			if targetEPS == 0 {
				continue
			}
			log.Printf(
				"cloud events: | target_eps:% 4d (% 4d)| undelivered:% 4d | ack:% 4d | nack:% 4d | sum:% 4d |",
				targetEPS, s.config.EpsLimit, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack,
			)
			// reset counts for last report
			atomic.StoreInt32(&s.undelivered, 0)
			atomic.StoreInt32(&s.ack, 0)
			atomic.StoreInt32(&s.nack, 0)
		case <-succt.C:
			s.mapLock.RLock()
			for _, subs := range s.events {
				for _, e := range subs {
					e.PrintStats()
				}
			}
			s.mapLock.RUnlock()
		}
	}
}

func (s *Sender) sendEvents() {
	for {
		var cases []reflect.SelectCase
		s.mapLock.RLock()
		for _, subs := range s.events {
			for _, e := range subs {
				if e.Events() != nil {
					cases = append(cases, reflect.SelectCase{
						Dir:  reflect.SelectRecv,
						Chan: reflect.ValueOf(e.Events()),
					})
				}
			}
		}
		s.mapLock.RUnlock()
		if len(cases) == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		_, value, ok := reflect.Select(cases)
		if !ok {
			continue
		}

		e := value.Interface().(*GenericEvent.Event)
		<-s.process
		go s.sendEvent(e)
	}
}

func (s *Sender) sendEvent(evt *GenericEvent.Event) {

	seq := <-evt.Counter()

	ce, err := evt.ToCloudEvent(seq)
	if err != nil {
		return
	}

	ctx := cev2.WithEncodingStructured(cev2.ContextWithTarget(s.ctx, s.endpoint))
	resp := s.client.Send(ctx, ce)
	switch {
	case cev2.IsUndelivered(resp):
		{
			atomic.AddInt32(&s.undelivered, 1)
			evt.Feedback() <- seq
		}
	case cev2.IsACK(resp):
		{
			atomic.AddInt32(&s.ack, 1)
			evt.Success() <- seq
		}
	case cev2.IsNACK(resp):
		{
			atomic.AddInt32(&s.nack, 1)
			evt.Feedback() <- seq
		}
	}
}

func (s *Sender) ComputeTotalEventsPerSecond() int {
	eps := 0
	s.mapLock.RLock()
	for _, subs := range s.events {
		for _, e := range subs {
			eps += e.Eps()
		}
	}
	s.mapLock.RUnlock()
	return eps
}

func (s *Sender) refillMaxEps(d time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in refillMaxEps", r)
		}
	}()

	t := time.NewTicker(d)
	for {
		select {
		case <-t.C:
			for i := 0; i < s.config.EpsLimit; i++ {
				s.process <- true
			}
		case <-s.ctx.Done():
			return
		}
	}
}
