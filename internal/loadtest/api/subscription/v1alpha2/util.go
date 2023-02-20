package v1alpha2

import (
	"fmt"

	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultNamespace = "default"
	noapp            = "noapp"
)

type SubscriptionOption func(*Subscription)

// New helps to build a Subscription uitilizing SubscriptionOptions.
func New(opts ...SubscriptionOption) *Subscription {
	subscription := &Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      uuid.NewString(),
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       Kind,
			APIVersion: GroupVersion.String(),
		},
		Spec: SubscriptionSpec{
			Config:       map[string]string{},
			Types:        []string{},
			Source:       noapp,
			TypeMatching: Standard,
		},
	}
	for _, opt := range opts {
		opt(subscription)
	}
	return subscription
}

// WithName is a SubscriptionOption that allow to define the name of a Subscription.
// Be adviced that the default name of a Subscription is a auto-generated UUI.
func WithName(name string) SubscriptionOption {
	return func(s *Subscription) {
		s.SetName(name)
	}
}

// WithNamespace is a SubscriptionOption that allows to define the namespace of a Subscription.
// Be adviced that the defaut namespace of a Subscription is `default`.
func WithNamespace(namespace string) SubscriptionOption {
	return func(s *Subscription) {
		s.SetNamespace(namespace)
	}
}

// WithSource is a SusbscriptionOption that allows to define the source of a Subscription.
func WithSource(source string) SubscriptionOption {
	return func(s *Subscription) {
		s.Spec.Source = source
	}
}

// WithTypeMatching is a Subscription to define the TypeMatching of a Subscription.
// Be advissed that the default value for a Subscription is `standard`.
func WithTypeMatching(typeMatching TypeMatching) SubscriptionOption {
	return func(s *Subscription) {
		s.Spec.TypeMatching = typeMatching
	}
}

// WithSink is  a SubscriptionOption that allows to define the sink of a Subscription.
func WithSink(sink string) SubscriptionOption {
	return func(s *Subscription) {
		s.Spec.Sink = sink
	}
}

// WithLabel is a SubscriptionOption that allows to add one label (key and value) at a time.
func WithLabel(key, value string) SubscriptionOption {
	return func(subscription *Subscription) {
		if subscription.Labels == nil {
			subscription.Labels = map[string]string{}
		}
		subscription.Labels[key] = value
	}
}

// WithConfig is a SubscriptionOption that allows to add one Config (key and value) at a time.
func WithConfig(key, value string) SubscriptionOption {
	return func(subscription *Subscription) {
		if subscription.Spec.Config == nil {
			subscription.Spec.Config = map[string]string{}
		}
		subscription.Spec.Config[key] = value
	}
}

// WithType is a SubscriptionOption that allows to add one Type at a time.
func WithType(eventType string) SubscriptionOption {
	return func(subscription *Subscription) {
		if subscription.Spec.Types == nil {
			subscription.Spec.Types = []string{}
		}
		subscription.Spec.Types = append(subscription.Spec.Types, eventType)
	}
}

// WithTypes is a SubscriptionOption that allows to add a number of Types to a Subscription at once.
func WithTypes(types ...string) SubscriptionOption {
	return func(s *Subscription) {
		if s.Spec.Types == nil {
			s.Spec.Types = []string{}
		}
		s.Spec.Types = append(s.Spec.Types, types...)
	}
}

// WithGeneratedTypes is a SubscriptionOption that automatically generates and adds a number of types of the
// form `<event>.v<n>` where <n> is a calculated number; it starts at <start>, gets incremented by <increase>
// and stops at a total of <count>.
func WithGeneratedTypes(format string, count, start, increment int) SubscriptionOption {
	eventTypes := make([]string, 0, count)
	for i, v := 0, start; i < count; i, v = i+1, start+(increment*(i+1)) {
		eventType := fmt.Sprintf(format, v)
		eventTypes = append(eventTypes, eventType)
	}

	return func(subscription *Subscription) {
		subscription.Spec.Types = eventTypes
	}
}

// containsString checks if a string is contained in a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
