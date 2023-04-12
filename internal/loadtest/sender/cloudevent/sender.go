package cloudevent

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"

	"github.com/kyma-project/eventing-tools/internal/client/cloudevents"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender"
)

// compile-time check for interfaces implementation.
var _ sender.Sender = &Sender{}

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

func (s *Sender) Format() string {
	return events.CloudeventFormat
}

func (s *Sender) Init(t *http.Transport, cfg *config.Config) {
	s.config = cfg
	s.client = cloudevents.NewClientOrDie(t)
}

func NewSender(conf *config.Config) *Sender {
	s := &Sender{config: conf}
	return s
}

func (s *Sender) SendEvent(evt *GenericEvent.Event, ack chan<- int, nack chan<- int, undelivered chan<- int) {

	seq := <-evt.Counter()

	ce, err := evt.ToCloudEvent(seq)
	if err != nil {
		return
	}

	endpoint := fmt.Sprintf("%v/publish", s.config.PublishEndpoint)
	ctx := cev2.WithEncodingStructured(cev2.ContextWithTarget(context.Background(), endpoint))
	resp := s.client.Send(ctx, ce)
	switch {
	case cev2.IsUndelivered(resp):
		{
			undelivered <- 1
			evt.Feedback() <- seq
		}
	case cev2.IsACK(resp):
		{
			ack <- 1
			evt.Success() <- seq
		}
	case cev2.IsNACK(resp):
		{
			nack <- 1
			evt.Feedback() <- seq
		}
	}
}
