package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubscriptionOption func(*Subscription)

func (s *Subscription) WithOptions(opts ...SubscriptionOption) {
	for _, opt := range opts {
		opt(s)
	}
}

func NewSubscription(namespace, name string, opts ...SubscriptionOption) *Subscription {
	subscription := &Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       Kind,
			APIVersion: GroupVersion.String(),
		},
		Spec: SubscriptionSpec{
			Config: &SubscriptionConfig{
				MaxInFlightMessages: 10,
			},
			Filter: &BEBFilters{
				Dialect: "nats",
				Filters: []*BEBFilter{},
			},
			Protocol: "",
			ProtocolSettings: &ProtocolSettings{
				ExemptHandshake: ptrBool(true),
				Qos:             ptrString("AT-LEAST-ONCE"),
			},
		},
	}
	for _, opt := range opts {
		opt(subscription)
	}
	return subscription
}

func NewFilter(eventType string) *BEBFilter {
	return &BEBFilter{
		EventSource: &Filter{Type: "source", Property: "exact", Value: ""},
		EventType:   &Filter{Type: "type", Property: "exact", Value: eventType},
	}
}

func GenerateFilters(format string, count, start, increment int) []*BEBFilter {
	filters := make([]*BEBFilter, 0, count)
	for i, v := 0, start; i < count; i, v = i+1, start+(increment*(i+1)) {
		eventType := fmt.Sprintf(format, v)
		filters = append(filters, NewFilter(eventType))
	}
	return filters
}

func WithMaxInFlightMessages(maxInFlightMessages int) SubscriptionOption {
	return func(subscription *Subscription) {
		subscription.Spec.Config.MaxInFlightMessages = maxInFlightMessages
	}
}

func WithFilters(filters []*BEBFilter) SubscriptionOption {
	return func(subscription *Subscription) {
		subscription.Spec.Filter.Filters = filters
	}
}

func WithSink(sink string) SubscriptionOption {
	return func(subscription *Subscription) {
		subscription.Spec.Sink = sink
	}
}

func WithLabels(labels map[string]string) SubscriptionOption {
	return func(subscription *Subscription) {
		subscription.Labels = labels
	}
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrString(s string) *string {
	return &s
}
