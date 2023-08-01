package cloudevent

import (
	"context"
	"fmt"

	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"

	"github.com/kyma-project/eventing-tools/internal/client/cloudevents"
	"github.com/kyma-project/eventing-tools/internal/client/transport"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/payload"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/interface"
)

// compile-time check for interfaces implementation.
var _ _interface.Sender = &Sender{}

// Sender sends cloud events.
type Sender struct {
	ackC, nackC, undeliveredC chan<- events.Event
	client                    client.Client
	config                    config.Config
}

func (s *Sender) Format() events.EventFormat {
	return events.CloudEvent
}

func NewSender(conf config.Config, ackC, nackC, undeliveredC chan<- events.Event) *Sender {
	t := transport.New(conf.MaxIdleConns, conf.MaxConnsPerHost, conf.MaxIdleConnsPerHost, conf.IdleConnTimeout)
	s := &Sender{
		config:       conf,
		ackC:         ackC,
		nackC:        nackC,
		undeliveredC: undeliveredC,
	}
	s.client = cloudevents.NewClientOrDie(t)
	return s
}

func ToCloudEvent(event events.Event) (cev2.Event, error) {
	ce := cev2.NewEvent()
	ce.SetType(event.EventType)
	ce.SetSource(event.Source)
	d := payload.DTO{
		Start: event.StartTime,
		Value: event.ID,
	}
	err := ce.SetData(cev2.ApplicationJSON, d)
	return ce, err
}

func (s *Sender) SendEvent(event events.Event) {
	ce, err := ToCloudEvent(event)
	if err != nil {
		return
	}

	endpoint := fmt.Sprintf("%v/publish", s.config.PublishHost)
	ctx := cev2.WithEncodingStructured(cev2.ContextWithTarget(context.Background(), endpoint))
	resp := s.client.Send(ctx, ce)
	switch {
	case cev2.IsUndelivered(resp):
		{
			s.undeliveredC <- event
		}
	case cev2.IsACK(resp):
		{
			s.ackC <- event
		}
	case cev2.IsNACK(resp):
		{
			s.nackC <- event
		}
	}
}
