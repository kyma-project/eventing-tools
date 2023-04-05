package events

import (
	cev2 "github.com/cloudevents/sdk-go/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/eventing-tools/internal/loadtest/events/GenericEvent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events/payload"
)

const (
	LegacyFormat     = "legacy"
	CloudeventFormat = "cloudevent"
)

type Event interface {
	ToCloudEvent(int, string) (cev2.Event, error)
	ToLegacyEvent(int) payload.LegacyEvent
	Stop()
	Eps() int
	Counter() <-chan int
	Source() string
	Feedback() chan<- int
	Success() chan<- int
	Events() <-chan Event
}

type EventFactory interface {
	FromSubscription(*unstructured.Unstructured, string) []*GenericEvent.Event
}
