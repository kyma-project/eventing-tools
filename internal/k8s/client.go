package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/eventing-tools/internal/logger"
)

func ClientOrDie(cfg *rest.Config) kubernetes.Interface {
	c, err := kubernetes.NewForConfig(cfg)
	logger.FatalIfError(err)
	return c
}
