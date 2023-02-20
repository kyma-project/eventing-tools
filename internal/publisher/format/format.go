package format

import (
	"fmt"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/google/uuid"
)

const (
	keyApp  = "app"
	keyMode = "mode"
	keyType = "type"

	version = "v1"
)

func LegacyPublishEndpoint(format, application string) string {
	return fmt.Sprintf(format, application)
}

func LegacyEventData(application, eventType string) string {
	return `{\"` + keyApp + `\":\"` + application + `\",\"` + keyMode + `\":\"legacy\",\"` + keyType + `\":\"` + eventType + `\"}`
}
func LegacyEventPayload(application, eventType string) string {
	data := LegacyEventData(application, eventType)
	return `{"data":"` + data + `","event-id":"` + uuid.New().String() + `","event-type":"` + eventType + `","event-time":"2020-04-02T21:37:00Z","event-type-version":"` + version + `"}`
}

func CloudEventMode(encoding binding.Encoding) string {
	return fmt.Sprintf("ce-%s", encoding.String())
}

func CloudEventData(application, eventType string, encoding binding.Encoding) map[string]interface{} {
	return map[string]interface{}{keyApp: application, keyMode: CloudEventMode(encoding), keyType: eventType}
}

func CloudEventType(prefix, application, eventType string) string {
	return fmt.Sprintf("%s.%s.%s.%s", prefix, application, eventType, version)
}
