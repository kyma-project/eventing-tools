package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Kind is the custom resource kind.
	Kind = "Subscription"

	// Resource is the custom resource plural name.
	Resource = "subscriptions"

	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "eventing.kyma-project.io", Version: "v1alpha1"}
)
