package cloudevent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/eventing-tools/internal/loadtest/sender"

	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"

	"github.com/kyma-project/eventing-tools/internal/client/cloudevents"
	"github.com/kyma-project/eventing-tools/internal/client/transport"
)

const (
	// buffer the event types to be sent by workers.
	buffer = 1_000_000
)

// compile-time check for interfaces implementation.
var _ sender.Sender = &Sender{}

// Sender sends legacy events.
type Sender struct {
	ctx         context.Context
	cancel      context.CancelFunc
	client      client.Client
	config      *config.Config
	events      []events.Event
	endpoint    string
	process     chan bool
	queue       chan events.Event
	cleanup     sync.WaitGroup
	running     bool
	undelivered int32
	ack         int32
	nack        int32
	// event payload size in bytes
	eventSize int32
}

func NewSender(conf *config.Config) *Sender {
	return &Sender{config: conf}
}

func (s *Sender) NotifyAdd(cm *corev1.ConfigMap) {
	s.stop()
	config.Map(cm, s.config)
	if s.config.UseLegacyEvents {
		return
	}
	log.Printf("Starting Cloud Event Sender")
	s.init()
	// delay for few seconds
	time.Sleep(10 * time.Second)
	s.start()
}

func (s *Sender) NotifyUpdate(cm *corev1.ConfigMap) {
	s.stop()
	config.Map(cm, s.config)
	if s.config.UseLegacyEvents {
		return
	}
	log.Printf("Starting Cloud Event Sender")
	s.init()
	time.Sleep(10 * time.Second)
	s.start()
}

func (s *Sender) NotifyDelete(*corev1.ConfigMap) {
	s.stop()
}

func (s *Sender) init() {
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = cloudevents.NewClientOrDie(t.Clone())
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.events = events.Generate(s.config)
	s.eventSize = 0
	s.process = make(chan bool, s.config.EpsLimit)
	s.queue = make(chan events.Event, buffer)
	s.undelivered = 0
	s.ack = 0
	s.nack = 0
	s.endpoint = fmt.Sprintf("%s/publish", s.config.PublishEndpoint)
	log.Printf(
		"cloud event sender endpoint: %s",
		s.endpoint,
	)
}

func (s *Sender) start() {
	s.running = true
	s.queueEventsAsync()
	s.sendEventsAsync()
	s.reportUsageAsync(time.Second)
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
	for _, e := range s.events {
		e.Stop()
	}
	close(s.process)
	close(s.queue)
	s.cleanup.Wait()
}

func (s *Sender) queueEventsAsync() {
	for _, e := range s.events {
		go s.queueEvent(e)
	}
}

func (s *Sender) sendEventsAsync() {
	for i := 0; i < s.config.Workers; i++ {
		go s.sendEvents()
	}
}

func (s *Sender) reportUsageAsync(d time.Duration) {
	targetEPS := s.config.ComputeTotalEventsPerSecond()

	go func() {
		defer func() {
			s.cleanup.Done()
		}()

		s.cleanup.Add(1)

		t := time.NewTicker(d)
		defer t.Stop()

		for s.running {
			<-t.C
			log.Printf(
				"cloud events: | eps:%04d | undelivered:%04d | ack:%04d | nack:%04d | sum:%04d | ack payload size:%d bytes |",
				targetEPS, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack, s.eventSize,
			)
			// reset counts for last report
			atomic.StoreInt32(&s.undelivered, 0)
			atomic.StoreInt32(&s.ack, 0)
			atomic.StoreInt32(&s.nack, 0)
		}
	}()
}

func (s *Sender) queueEvent(evt events.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from: ", r)
		}
		s.cleanup.Done()
	}()

	s.cleanup.Add(1)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	// queue event immediately
	for ; s.running; <-t.C {
		for i := 0; i < evt.Eps; i++ {
			if !s.running {
				return
			}
			s.queue <- evt
		}
	}
}

func (s *Sender) sendEvents() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from: ", r)
		}
		s.cleanup.Done()
	}()

	s.cleanup.Add(1)

	for s.running {
		e := <-s.queue
		if !s.running {
			return
		}

		s.process <- true
		go s.sendEvent(e)
	}
}

func (s *Sender) sendEvent(evt events.Event) {
	defer func() {
		if !s.running {
			return
		}

		<-s.process
	}()

	seq := <-evt.Counter

	ce, err := evt.ToCloudEvent(seq, s.config.EventSource)
	if err != nil {
		return
	}

	singleEventSize := len(ce.Data())

	ctx := cev2.ContextWithTarget(s.ctx, s.endpoint)
	resp := s.client.Send(ctx, ce)
	switch {
	case cev2.IsUndelivered(resp):
		{
			atomic.AddInt32(&s.undelivered, 1)
			evt.Feedback <- seq
		}
	case cev2.IsACK(resp):
		{
			atomic.AddInt32(&s.ack, 1)
			s.eventSize = int32(singleEventSize) * s.ack
			evt.Success <- seq
		}
	case cev2.IsNACK(resp):
		{
			atomic.AddInt32(&s.nack, 1)
			evt.Feedback <- seq
		}
	}
}
