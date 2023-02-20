package publisher

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/publisher/config"
	"github.com/kyma-project/eventing-tools/internal/publisher/format"
	"github.com/kyma-project/eventing-tools/internal/publisher/mapping"
)

type publisher struct {
	ctx        context.Context
	clientCE   client.Client
	clientHTTP *http.Client
	conf       *config.Config
}

func Start(ctx context.Context, clientCE client.Client, clientHTTP *http.Client, conf *config.Config) {
	p := &publisher{
		ctx:        ctx,
		clientCE:   clientCE,
		clientHTTP: clientHTTP,
		conf:       conf,
	}

	log.Printf("Publisher starting on port:[%s]", conf.ServerAddress)
	p.publishAsync(conf.PublishInterval)
}

func (p *publisher) publishAsync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// publish events immediately
		for ; true; <-ticker.C {
			for application, eventType := range mapping.ApplicationEventType {
				p.sendLegacyEvent(application, eventType)
				p.sendCloudEvent(application, eventType, binding.EncodingBinary)
				p.sendCloudEvent(application, eventType, binding.EncodingStructured)
			}
		}
	}()
}

func (p *publisher) sendLegacyEvent(application, eventType string) {
	url := format.LegacyPublishEndpoint(p.conf.PublishEndpointLegacyEvents, application)
	payload := format.LegacyEventPayload(application, eventType)

	req, err1 := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(payload)))
	if err1 != nil {
		log.Printf("Failed to create HTTP request with error:[%s]", err1)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err2 := p.clientHTTP.Do(req)
	if err2 != nil {
		log.Printf("Failed to send legacy-event with error:[%s]", err2)
		return
	}
	defer func() { logger.LogIfError(resp.Body.Close()) }()

	if is2XX(resp.StatusCode) {
		log.Printf("Sent legacy-event [%s]", eventType)
		return
	}

	if body, err3 := io.ReadAll(resp.Body); err3 != nil {
		log.Printf("Failed to read response body with error:[%s]", err3)
	} else {
		log.Printf("Failed to send legacy-event:[%s] response:[%d] body:[%s]", eventType, resp.StatusCode, string(body))
	}
}

func (p *publisher) sendCloudEvent(application, eventType string, encoding binding.Encoding) {
	ce := cloudevents.NewEvent()
	eventType = format.CloudEventType(p.conf.EventTypePrefix, application, eventType)
	data := format.CloudEventData(application, eventType, encoding)
	ce.SetType(eventType)
	ce.SetSource(p.conf.EventSource)
	if err := ce.SetData(cloudevents.ApplicationJSON, data); err != nil {
		log.Printf("Failed to set cloudevent-%s data with error:[%s]", encoding.String(), err)
		return
	}

	ctx := cloudevents.ContextWithTarget(p.ctx, p.conf.PublishEndpointCloudEvents)
	switch encoding {
	case binding.EncodingBinary:
		{
			ctx = binding.WithForceBinary(ctx)
		}
	case binding.EncodingStructured:
		{
			ctx = binding.WithForceStructured(ctx)
		}
	default:
		{
			log.Printf("Failed to use unsupported cloudevent encoding:[%s]", encoding.String())
			return
		}
	}

	result := p.clientCE.Send(ctx, ce)
	switch {
	case cloudevents.IsUndelivered(result):
		{
			log.Printf("Failed to send cloudevent-%s undelivered:[%s] response:[%s]", encoding.String(), eventType, result)
			return
		}
	case cloudevents.IsNACK(result):
		{
			log.Printf("Failed to send cloudevent-%s nack:[%s] response:[%s]", encoding.String(), eventType, result)
			return
		}
	case cloudevents.IsACK(result):
		{
			log.Printf("Sent cloudevent-%s [%s]", encoding.String(), eventType)
			return
		}
	default:
		{
			log.Printf("Failed to send cloudevent-%s unknown:[%s] response:[%s]", encoding.String(), eventType, result)
			return
		}
	}
}

func is2XX(statusCode int) bool {
	return http.StatusOK <= statusCode && statusCode <= http.StatusIMUsed
}
