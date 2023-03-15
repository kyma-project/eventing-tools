package publisher

import (
	"net/http"

	"github.com/kyma-project/eventing-tools/internal/k8s"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/infra"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/cloudevent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/legacyevent"
	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
)

func Start() {
	appConfig := config.New()
	k8sConfig := k8s.ConfigOrDie()
	k8sClient := k8s.ClientOrDie(k8sConfig)

	legacyEventSender := legacyevent.NewSender(appConfig)
	cloudEventSender := cloudevent.NewSender(appConfig)
	infraInstance := infra.New(appConfig, k8sConfig)

	config.NewWatcher(k8sClient, infra.Namespace, infra.ConfigMapName).
		OnAddNotify(infraInstance).
		OnUpdateNotify(infraInstance).
		OnDeleteNotify(infraInstance).
		OnAddNotify(legacyEventSender).
		OnUpdateNotify(legacyEventSender).
		OnDeleteNotify(legacyEventSender).
		OnAddNotify(cloudEventSender).
		OnUpdateNotify(cloudEventSender).
		OnDeleteNotify(cloudEventSender).
		OnDeleteNotifyMe().
		Watch()

	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(appConfig.ServerAddress, nil))
}
