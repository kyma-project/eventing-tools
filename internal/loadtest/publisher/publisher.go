package publisher

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/eventing-tools/internal/k8s"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	sender2 "github.com/kyma-project/eventing-tools/internal/loadtest/sender"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
)

const (
	Namespace = "eventing-test"

	ConfigMapName = "loadtest-publisher"
)

func Start(port int) {
	//appConfig := config.New()
	k8sConfig := k8s.ConfigOrDie()
	k8sClient := k8s.ClientOrDie(k8sConfig)
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)

	sender, senderC := sender2.NewSender()
	factory := events.NewGeneratorFactory(senderC)

	config.NewWatcher(k8sClient, Namespace, ConfigMapName).
		OnAddNotify(sender).
		OnUpdateNotify(sender).
		OnDeleteNotify(sender).
		Watch()

	sender.Start()

	subscription.NewWatcher(dynamicClient).
		OnAddNotify(factory).
		OnUpdateNotify(factory).
		OnDeleteNotify(factory).
		Watch()

	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
