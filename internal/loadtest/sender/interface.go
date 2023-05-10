package sender

import (
	"net/http"

	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
)

type ConfigHandler interface {
	config.AddNotifiable
	config.UpdateNotifiable
	config.DeleteNotifiable
}

type Sender interface {
	SendEvent(e *GenericEvent.Event, ack chan<- int, nack chan<- int, undelivered chan<- int)
	Format() string
	Init(t *http.Transport, cfg *config.Config)
}
