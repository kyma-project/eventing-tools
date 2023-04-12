package publisher

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/eventing-tools/internal/k8s"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/infra"
	sender2 "github.com/kyma-project/eventing-tools/internal/loadtest/sender"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/cloudevent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/legacyevent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
)

func Start(port int) {
	appConfig := config.New()
	k8sConfig := k8s.ConfigOrDie()
	k8sClient := k8s.ClientOrDie(k8sConfig)
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)

	legacyEventSender := legacyevent.NewSender(appConfig)
	legacySender := sender2.NewSender(appConfig, legacyEventSender)

	cloudEventSender := cloudevent.NewSender(appConfig)
	cloudSender := sender2.NewSender(appConfig, cloudEventSender)

	config.NewWatcher(k8sClient, infra.Namespace, infra.ConfigMapName).
		// OnAddNotify(infraInstance).
		// OnUpdateNotify(infraInstance).
		// OnDeleteNotify(infraInstance).
		OnAddNotify(legacySender).
		OnUpdateNotify(legacySender).
		OnDeleteNotify(legacySender).
		OnAddNotify(cloudSender).
		OnUpdateNotify(cloudSender).
		OnDeleteNotify(cloudSender).
		OnDeleteNotifyMe().
		Watch()

	subscription.NewWatcher(dynamicClient).
		OnAddNotify(cloudSender).
		OnUpdateNotify(cloudSender).
		OnDeleteNotify(cloudSender).
		OnAddNotify(legacySender).
		OnUpdateNotify(legacySender).
		OnDeleteNotify(legacySender).
		Watch()

	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
