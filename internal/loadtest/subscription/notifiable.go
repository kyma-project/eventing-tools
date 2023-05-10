package subscription

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type AddNotifiable interface {
	OnNewSubscription(subscription *unstructured.Unstructured)
}

type UpdateNotifiable interface {
	OnChangedSubscription(subscription *unstructured.Unstructured)
}

type DeleteNotifiable interface {
	OnDeleteSubscription(subscription *unstructured.Unstructured)
}

type Notifiable interface {
	AddNotifiable
	UpdateNotifiable
	DeleteNotifiable
}
