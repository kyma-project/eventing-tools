package sender

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/eventing-tools/internal/client/transport"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericFactory"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
)

// compile-time check for interfaces implementation.
var _ subscription.Notifiable = &EventSender{}

// Sender sends cloud events.
type EventSender struct {
	ctx                       context.Context
	cancel                    context.CancelFunc
	config                    *config.Config
	events                    map[string][]*GenericEvent.Event
	factory                   events.EventFactory
	process                   chan bool
	running                   bool
	undelivered               int32
	ack                       int32
	nack                      int32
	mapLock                   sync.RWMutex
	wg                        sync.WaitGroup
	stopper                   sync.Mutex
	sender                    Sender
	acks, nacks, undelivereds chan int
}

func (s *EventSender) FormatName() string {
	// TODO implement me
	panic("implement me")
}

func NewSender(conf *config.Config, sender Sender) *EventSender {
	s := &EventSender{config: conf}
	s.undelivered = 0
	s.ack = 0
	s.nack = 0
	s.events = make(map[string][]*GenericEvent.Event)
	s.factory = GenericFactory.New()
	s.sender = sender
	s.acks = make(chan int)
	s.nacks = make(chan int)
	s.undelivereds = make(chan int)

	return s
}

func (s *EventSender) NotifyAdd(cm *corev1.ConfigMap) {
	s.stopper.Lock()
	defer s.stopper.Unlock()
	s.stop()
	config.Map(cm, s.config)
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.sender.Init(t, s.config)
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.process = make(chan bool, s.config.EpsLimit)
	s.start()
}

func (s *EventSender) NotifyUpdate(cm *corev1.ConfigMap) {
	s.stopper.Lock()
	defer s.stopper.Unlock()
	s.stop()
	config.Map(cm, s.config)
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.sender.Init(t, s.config)
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.process = make(chan bool, s.config.EpsLimit)
	s.start()
}

func (s *EventSender) NotifyDelete(*corev1.ConfigMap) {
	s.stopper.Lock()
	defer s.stopper.Unlock()
	s.stop()
}

func (s *EventSender) OnNewSubscription(subscription *unstructured.Unstructured) {
	ne := s.factory.FromSubscription(subscription, s.sender.Format())
	if len(ne) == 0 {
		return
	}
	s.stopper.Lock()
	defer s.stopper.Unlock()

	// s.queue = make(chan events.Event, buffer)
	for _, e := range ne {
		e.Start()
	}
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	s.events[fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName())] = ne
}

func (s *EventSender) OnChangedSubscription(subscription *unstructured.Unstructured) {
	if subscription.GetDeletionTimestamp() != nil {
		return
	}
	s.stopper.Lock()
	defer s.stopper.Unlock()
	for _, e := range s.events[fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName())] {
		e.Stop()
	}
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	delete(s.events, fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName()))

	ne := s.factory.FromSubscription(subscription, s.sender.Format())
	if len(ne) == 0 {
		return
	}

	for _, e := range ne {
		e.Start()
	}
	s.events[fmt.Sprintf("%v/%v", subscription.GetNamespace(), subscription.GetName())] = ne
}

func (s *EventSender) OnDeleteSubscription(sub *unstructured.Unstructured) {
	s.stopper.Lock()
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	defer s.stopper.Unlock()
	for _, e := range s.events[fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName())] {
		e.Stop()
	}
	delete(s.events, fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName()))
}

func (s *EventSender) init() {
}

func (s *EventSender) start() {
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

func (s *EventSender) stop() {
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

func (s *EventSender) sendEventsAsync() {
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go s.sendEvents()
	}
}

func (s *EventSender) reportUsageAsync(send, success time.Duration) {

	defer func() {
		s.wg.Done()
	}()

	sendt := time.NewTicker(send)
	defer sendt.Stop()
	succt := time.NewTicker(success)
	defer succt.Stop()

	for {
		select {
		case na := <-s.acks:
			atomic.AddInt32(&s.ack, int32(na))
		case nn := <-s.nacks:
			atomic.AddInt32(&s.nack, int32(nn))
		case nu := <-s.undelivereds:
			atomic.AddInt32(&s.undelivered, int32(nu))

		case <-s.ctx.Done():
			targetEPS := s.ComputeTotalEventsPerSecond()
			log.Printf(
				"%v: | target_eps:% 4d (% 4d)| undelivered:% 4d | ack:% 4d | nack:% 4d | sum:% 4d |",
				s.sender.Format(), targetEPS, s.config.EpsLimit, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack,
			)
			return
		case <-sendt.C:
			targetEPS := s.ComputeTotalEventsPerSecond()
			if targetEPS == 0 {
				continue
			}
			log.Printf(
				"%v: | target_eps:% 4d (% 4d)| undelivered:% 4d | ack:% 4d | nack:% 4d | sum:% 4d |",
				s.sender.Format(), targetEPS, s.config.EpsLimit, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack,
			)
			// reset counts for last report
			atomic.StoreInt32(&s.undelivered, 0)
			atomic.StoreInt32(&s.ack, 0)
			atomic.StoreInt32(&s.nack, 0)
		case <-succt.C:
			s.mapLock.RLock()
			stats := []string{fmt.Sprintf("%v:", s.sender.Format())}
			for _, subs := range s.events {
				for _, e := range subs {
					stats = append(stats, e.PrintStats())
				}
			}
			fmt.Println(strings.Join(stats, "\n\t"))
			s.mapLock.RUnlock()
		}
	}
}

func (s *EventSender) sendEvents() {
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
		go s.sender.SendEvent(e, s.acks, s.nacks, s.undelivereds)
	}
}

func (s *EventSender) ComputeTotalEventsPerSecond() int {
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

func (s *EventSender) refillMaxEps(d time.Duration) {
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
