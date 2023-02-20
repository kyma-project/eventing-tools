package config

import corev1 "k8s.io/api/core/v1"

type AddNotifiable interface {
	NotifyAdd(*corev1.ConfigMap)
}

type UpdateNotifiable interface {
	NotifyUpdate(*corev1.ConfigMap)
}

type DeleteNotifiable interface {
	NotifyDelete(*corev1.ConfigMap)
}
