package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebhookAuth defines the Webhook called by an active subscription in BEB.
type WebhookAuth struct {
	// Type defines type of authentication
	// +optional
	Type string `json:"type,omitempty"`

	// GrantType defines grant type for OAuth2
	GrantType string `json:"grantType"`

	// ClientID defines clientID for OAuth2
	ClientID string `json:"clientId"`

	// ClientSecret defines client secret for OAuth2
	ClientSecret string `json:"clientSecret"`

	// TokenURL defines token URL for OAuth2
	TokenURL string `json:"tokenUrl"`

	// Scope defines scope for OAuth2
	Scope []string `json:"scope,omitempty"`
}

// ProtocolSettings defines the CE protocol setting specification implementation.
type ProtocolSettings struct {
	// ContentMode defines content mode for eventing based on BEB.
	// +optional
	ContentMode *string `json:"contentMode,omitempty"`

	// ExemptHandshake defines whether exempt handshake for eventing based on BEB.
	// +optional
	ExemptHandshake *bool `json:"exemptHandshake,omitempty"`

	// Qos defines quality of service for eventing based on BEB.
	// +optional
	Qos *string `json:"qos,omitempty"`

	// WebhookAuth defines the Webhook called by an active subscription in BEB.
	// +optional
	WebhookAuth *WebhookAuth `json:"webhookAuth,omitempty"`
}

// Filter defines the CE filter element.
type Filter struct {
	// Type defines the type of the filter
	// +optional
	Type string `json:"type,omitempty"`

	// Property defines the property of the filter
	Property string `json:"property"`

	// Value defines the value of the filter
	Value string `json:"value"`
}

// BEBFilter defines the BEB filter element as a combination of two CE filter elements.
type BEBFilter struct {
	// EventSource defines the source of CE filter
	EventSource *Filter `json:"eventSource"`

	// EventType defines the type of CE filter
	EventType *Filter `json:"eventType"`
}

// BEBFilters defines the list of BEB filters.
type BEBFilters struct {
	// +optional
	Dialect string `json:"dialect,omitempty"`

	Filters []*BEBFilter `json:"filters"`
}

type SubscriptionConfig struct {
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxInFlightMessages int `json:"maxInFlightMessages,omitempty"`
}

// SubscriptionSpec defines the desired state of Subscription.
type SubscriptionSpec struct {
	// ID is the unique identifier of Subscription, read-only.
	// +optional
	ID string `json:"id,omitempty"`

	// Protocol defines the CE protocol specification implementation
	// +optional
	Protocol string `json:"protocol,omitempty"`

	// ProtocolSettings defines the CE protocol setting specification implementation
	// +optional
	ProtocolSettings *ProtocolSettings `json:"protocolsettings,omitempty"`

	// Sink defines endpoint of the subscriber
	Sink string `json:"sink"`

	// Filter defines the list of filters
	Filter *BEBFilters `json:"filter"`

	// Config defines the configurations that can be applied to the eventing backend when creating this subscription
	// +optional
	Config *SubscriptionConfig `json:"config,omitempty"`
}

// Subscription is the Schema for the subscriptions API.
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SubscriptionSpec `json:"spec,omitempty"`
}

// SubscriptionList contains a list of Subscription.
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}
