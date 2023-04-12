package legacyevent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender"
)

var _ sender.Sender = &Sender{}

const format = events.LegacyFormat

// Sender sends legacy events.
type Sender struct {
	client *http.Client
	config *config.Config
	ctx    context.Context
}

func (s *Sender) Init(t *http.Transport, cfg *config.Config) {
	s.config = cfg
	s.client = &http.Client{
		Transport: t,
	}
}

func (s *Sender) Format() string {
	return events.LegacyFormat
}

func NewSender(conf *config.Config) *Sender {
	s := &Sender{config: conf}

	return s
}

func (s *Sender) SendEvent(evt *GenericEvent.Event, ack, nack, undelivered chan<- int) {

	seq := <-evt.Counter()

	// Build a http request out of the legacy event.
	le := evt.ToLegacyEvent(seq)
	b, err := json.Marshal(le)
	if err != nil {
		return
	}

	r := bytes.NewReader(b)
	legacyEndpoint := fmt.Sprintf("%s/%s/v1/events", s.config.PublishEndpoint, evt.Source())
	rq, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, legacyEndpoint, r)
	if err != nil {
		return
	}
	rq.Header.Add("Content-Type", "application/json")

	// Send the http request.
	if s.client == nil {
		undelivered <- 1
		return
	}
	resp, err := s.client.Do(rq)
	if err != nil {
		undelivered <- 1
		evt.Feedback() <- seq
		return
	}

	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck // we just to read the body as we want to reuse the connection

	// Evaluate the response.
	switch {
	case resp.StatusCode/100 == 2:
		{
			ack <- 1
			evt.Success() <- seq
		}
	default:
		{
			nack <- 1
			evt.Feedback() <- seq
		}
	}
}
