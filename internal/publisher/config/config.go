package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/kyma-project/eventing-tools/internal/logger"
)

type Config struct {
	ServerAddress               string        `envconfig:"SERVER_ADDRESS" required:"true" default:":8888"`
	PublishEndpointCloudEvents  string        `envconfig:"PUBLISH_ENDPOINT_CLOUDEVENTS" required:"true" default:"http://eventing-publisher-proxy.kyma-system/publish"`
	PublishEndpointLegacyEvents string        `envconfig:"PUBLISH_ENDPOINT_LEGACY_EVENTS" required:"true" default:"http://eventing-publisher-proxy.kyma-system/%s/v1/events"`
	PublishInterval             time.Duration `envconfig:"PUBLISH_INTERVAL" required:"true" default:"10s"`
	EventSource                 string        `envconfig:"EVENT_SOURCE" required:"true" default:"/default/sap.kyma/tunas-develop"`
	EventTypePrefix             string        `envconfig:"EVENT_TYPE_PREFIX" required:"true" default:"sap.kyma.custom"`
	MaxIdleConns                int           `envconfig:"MAX_IDLE_CONNS" required:"true" default:"10"`
	MaxConnsPerHost             int           `envconfig:"MAX_CONNS_PER_HOST" required:"true" default:"10"`
	MaxIdleConnsPerHost         int           `envconfig:"MAX_IDLE_CONNS_PER_HOST" required:"true" default:"10"`
	IdleConnTimeout             time.Duration `envconfig:"IDLE_CONN_TIMEOUT" required:"true" default:"1m0s"`
}

func ProcessOrDie() *Config {
	config := &Config{}
	err := envconfig.Process("", config)
	logger.FatalIfError(err)
	return config
}
