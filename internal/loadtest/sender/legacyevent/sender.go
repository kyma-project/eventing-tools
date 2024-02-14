package legacyevent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/eventing-tools/internal/client/transport"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/payload"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/interface"
)

var _ _interface.Sender = &Sender{}

// Sender sends legacy events.
type Sender struct {
	ackC, nackC, undeliveredC chan<- events.Event
	client                    *http.Client
	config                    config.Config
}

func (s *Sender) Format() events.EventFormat {
	return events.Legacy
}

func NewSender(cfg config.Config, ackC, nackC, undeliveredC chan<- events.Event) *Sender {
	s := &Sender{
		config: cfg,
		client: &http.Client{
			Transport: transport.New(cfg.MaxIdleConns, cfg.MaxConnsPerHost, cfg.MaxIdleConnsPerHost, cfg.IdleConnTimeout),
		},
		ackC:         ackC,
		nackC:        nackC,
		undeliveredC: undeliveredC,
	}

	return s
}

func ToLegacyEvent(event events.Event) payload.LegacyEvent {
	d := payload.DTO{
		Start: event.StartTime,
		Value: event.ID,
	}
	eventtype, version := splitEventType(event.EventType)
	return payload.LegacyEvent{
		Data:             d,
		EventType:        eventtype,
		EventTypeVersion: version,
		EventTime:        time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventTracing:     true,
	}
}

func splitEventType(eventType string) (string, string) {
	i := strings.LastIndex(eventType, ".")
	return eventType[0:i], eventType[i+1:]
}

func (s *Sender) SendEvent(event events.Event) {
	// Build a http request out of the legacy event.
	le := ToLegacyEvent(event)
	b, err := json.Marshal(le)
	if err != nil {
		return
	}

	r := bytes.NewReader(b)
	legacyEndpoint := fmt.Sprintf("%s/%s/v1/events", s.config.PublishHost, event.Source)
	rq, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, legacyEndpoint, r)
	if err != nil {
		return
	}
	rq.Header.Add("Content-Type", "application/json")

	// Send the http request.
	if s.client == nil {
		s.undeliveredC <- event
		return
	}
	resp, err := s.client.Do(rq)
	if err != nil {
		log.Println(err)
		s.undeliveredC <- event
		return
	}

	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck // we just to read the body as we want to reuse the connection

	// Evaluate the response.
	switch {
	case resp.StatusCode/100 == 2:
		{
			s.ackC <- event
		}
	default:
		{
			s.nackC <- event
		}
	}
}
