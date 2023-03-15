package legacyevent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/eventing-tools/internal/client/transport"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender"
)

const (
	buffer = 1_000_000
)

// Compile-time check of interface implementation.
var _ sender.Sender = &Sender{}

type Sender struct {
	ctx         context.Context
	cancel      context.CancelFunc
	config      *config.Config
	events      []events.Event
	client      http.Client
	endpoint    string
	process     chan bool
	queue       chan events.Event
	cleanup     sync.WaitGroup
	running     bool
	undelivered int32
	ack         int32
	nack        int32
	eventSize   int // event size in bytes
}

func NewSender(conf *config.Config) *Sender {
	return &Sender{config: conf}
}

func (s *Sender) NotifyAdd(cm *corev1.ConfigMap) {
	s.stop()
	config.Map(cm, s.config)
	if !s.config.UseLegacyEvents {
		return
	}
	log.Printf("Starting Legacy Event Sender")
	s.init()
	// Delay for few seconds.
	time.Sleep(10 * time.Second)
	s.start()
}

func (s *Sender) NotifyUpdate(cm *corev1.ConfigMap) {
	s.stop()
	config.Map(cm, s.config)
	if !s.config.UseLegacyEvents {
		return
	}
	log.Printf("Starting Legacy Event Sender")
	s.init()
	time.Sleep(10 * time.Second)
	s.start()
}

func (s *Sender) NotifyDelete(*corev1.ConfigMap) {
	s.stop()
}

func (s *Sender) init() {
	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
	s.client = http.Client{
		Transport: t.Clone(),
	}
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.events = events.Generate(s.config)
	s.eventSize = 0
	s.process = make(chan bool, s.config.EpsLimit)
	s.queue = make(chan events.Event, buffer)
	s.undelivered = 0
	s.ack = 0
	s.nack = 0
	s.endpoint = fmt.Sprintf("%s/%s/v1/events", s.config.PublishEndpoint, s.config.EventSource)
	log.Printf(
		"legacy event sender endpoint: %s",
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
	// Recover from closing already closed channels.
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from:", r)
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
				"legacy events: | eps:%04d | undelivered:%04d | ack:%04d | nack:%04d | sum:%04d | event size:%d bytes |",
				targetEPS, s.undelivered, s.ack, s.nack, s.undelivered+s.ack+s.nack, s.eventSize,
			)

			// Reset counts for last report.
			atomic.StoreInt32(&s.undelivered, 0)
			atomic.StoreInt32(&s.ack, 0)
			atomic.StoreInt32(&s.nack, 0)
		}
	}()
}

func (s *Sender) queueEvent(evt events.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from:", r)
		}
		s.cleanup.Done()
	}()

	s.cleanup.Add(1)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	// Queue event immediately.
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
			log.Println("Recovered from:", r)
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
	// Check if the sender is still running.
	defer func() {
		if !s.running {
			return
		}
		<-s.process
	}()

	seq := <-evt.Counter

	// Build a http request out of the legacy event.
	le := evt.ToLegacyEvent(seq)
	b, err := json.Marshal(le)
	if err != nil {
		return
	}

	s.eventSize = len(b)

	r := bytes.NewReader(b)
	rq, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.endpoint, r)
	if err != nil {
		return
	}
	rq.Header.Add("Content-Type", "application/json")

	// Send the http request.
	resp, err := s.client.Do(rq)
	if err != nil {
		atomic.AddInt32(&s.undelivered, 1)
		evt.Feedback <- seq
		return
	}

	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck // we just to read the body as we want to reuse the connection

	// Evaluate the response.
	switch {
	case resp.StatusCode/100 == 2:
		{
			atomic.AddInt32(&s.ack, 1)
			evt.Success <- seq
		}
	default:
		{
			atomic.AddInt32(&s.nack, 1)
			evt.Feedback <- seq
		}
	}
}
