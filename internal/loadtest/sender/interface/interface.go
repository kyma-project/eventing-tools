package _interface

import (
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
)

type ConfigHandler interface {
	config.AddNotifiable
	config.UpdateNotifiable
	config.DeleteNotifiable
}

type Sender interface {
	SendEvent(event events.Event)
	Format() events.EventFormat
}
