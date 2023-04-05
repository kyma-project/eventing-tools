package legacyevent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
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
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
)

const (
	// buffer the event types to be sent by workers.
	buffer = 1_000_000
)

// compile-time check for interfaces implementation.
var _ sender.Sender = &Sender{}
var _ subscription.Notifiable = &Sender{}

// Sender sends legacy events.
type Sender struct {
	ctx         context.Context
	cancel      context.CancelFunc
	client      http.Client
	config      *config.Config
	events      map[string][]*GenericEvent.Event
	factory     events.EventFactory
	endpoint    string
	process     chan bool
	cleanup     sync.WaitGroup
	running     bool
	undelivered int32
	ack         int32
	nack        int32
	mapLock     sync.RWMutex
}

func NewSender(conf *config.Config) *Sender {
	s := &Sender{config: conf}
	s.undelivered = 0
	s.ack = 0
	s.nack = 0
	s.events = make(map[string][]*GenericEvent.Event)
	s.factory = GenericFactory.New(s)

	return s
}

func (s *Sender) NotifyAdd(cm *corev1.ConfigMap) {
	s.stop()
	config.Map(cm, s.config)
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = http.Client{
		Transport: t,
	}
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.process = make(chan bool, s.config.EpsLimit)
	s.start()
}

func (s *Sender) NotifyUpdate(cm *corev1.ConfigMap) {
	s.stop()
	config.Map(cm, s.config)
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = http.Client{
		Transport: t,
	}
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.process = make(chan bool, s.config.EpsLimit)
	s.start()
}

func (s *Sender) NotifyDelete(*corev1.ConfigMap) {
	s.stop()
}

func (s *Sender) OnNewSubscription(sub *unstructured.Unstructured) {
	log.Printf("Starting Legacy Event Sender")
	e := s.factory.FromSubscription(sub, events.LegacyFormat)
	if len(e) == 0 {
		return
	}
	// s.queue = make(chan events.Event, buffer)
	s.mapLock.Lock()
	s.events[fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName())] = e
	s.mapLock.Unlock()
}

func (s *Sender) OnChangedSubscription(sub *unstructured.Unstructured) {
	ne := s.factory.FromSubscription(sub, events.LegacyFormat)
	if len(ne) == 0 {
		return
	}
	s.mapLock.RLock()
	for _, e := range s.events[fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName())] {
		e.Stop()
	}
	s.mapLock.RUnlock()

	s.mapLock.Lock()
	s.events[fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName())] = ne
	s.mapLock.Unlock()
}

func (s *Sender) OnDeleteSubscription(sub *unstructured.Unstructured) {
	s.mapLock.RLock()
	for _, e := range s.events[fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName())] {
		e.Stop()
	}
	s.mapLock.RUnlock()
	s.mapLock.Lock()
	delete(s.events, fmt.Sprintf("%v/%v", sub.GetNamespace(), sub.GetName()))
	s.mapLock.Unlock()
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
	s.refillMaxEps(time.Second)
	s.reportUsageAsync(time.Second, 20*time.Second)
}

func (s *Sender) stop() {
	// recover from closing already closed channels
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from: ", r)
		}
	}()

	s.running = false
	s.cancel()
	s.mapLock.RLock()
	for _, subs := range s.events {
		for _, e := range subs {
			e.Stop()
		}
	}
	s.mapLock.RUnlock()
	close(s.process)
	s.cleanup.Wait()
}

func (s *Sender) sendEventsAsync() {
	for i := 0; i < s.config.Workers; i++ {
		go s.sendEvents(i)
	}
}

func (s *Sender) reportUsageAsync(send, success time.Duration) {

	go func() {
		defer func() {
			s.cleanup.Done()
		}()

		s.cleanup.Add(1)

		sendt := time.NewTicker(send)
		defer sendt.Stop()
		succt := time.NewTicker(success)
		defer succt.Stop()

		for {
			select {
			case <-s.ctx.Done():
				targetEPS := s.ComputeTotalEventsPerSecond()
				log.Printf(
					"legacy events: | eps:%04d | undelivered:%04d | ack:%04d | nack:%04d | sum:%04d |",
					targetEPS, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack,
				)
				return
			case <-sendt.C:
				targetEPS := s.ComputeTotalEventsPerSecond()
				if targetEPS == 0 {
					continue
				}
				log.Printf(
					"legacy events: | eps:%04d | undelivered:%04d | ack:%04d | nack:%04d | sum:%04d |",
					targetEPS, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack,
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
	}()
}

func (s *Sender) sendEvents(id int) {
	for {
		cases := []reflect.SelectCase{}
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

	// Build a http request out of the legacy event.
	le := evt.ToLegacyEvent(seq)
	b, err := json.Marshal(le)
	if err != nil {
		return
	}

	r := bytes.NewReader(b)
	legacyEndpoint := fmt.Sprintf("%s/%s/v1/events", s.config.PublishEndpoint, evt.Source())
	rq, err := http.NewRequestWithContext(s.ctx, http.MethodPost, legacyEndpoint, r)
	if err != nil {
		return
	}
	rq.Header.Add("Content-Type", "application/json")

	// Send the http request.
	resp, err := s.client.Do(rq)
	if err != nil {
		atomic.AddInt32(&s.undelivered, 1)
		evt.Feedback() <- seq
		return
	}

	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck // we just to read the body as we want to reuse the connection

	// Evaluate the response.
	switch {
	case resp.StatusCode/100 == 2:
		{
			atomic.AddInt32(&s.ack, 1)
			evt.Success() <- seq
		}
	default:
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
	go func() {
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
	}()
}
