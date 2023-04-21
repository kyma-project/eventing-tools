package publisher

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/eventing-tools/internal/k8s"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	sender2 "github.com/kyma-project/eventing-tools/internal/loadtest/sender"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/cloudevent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/legacyevent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/subscription"
	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
)

const (
	Namespace = "eventing-test"

	ConfigMapName = "loadtest-publisher"
)

func Start(port int) {
	appConfig := config.New()
	k8sConfig := k8s.ConfigOrDie()
	k8sClient := k8s.ClientOrDie(k8sConfig)
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)

	legacySender := legacyevent.NewSender(appConfig)
	legacyEventSender := sender2.NewSender(appConfig, legacySender)

	ceSender := cloudevent.NewSender(appConfig)
	ceEventSender := sender2.NewSender(appConfig, ceSender)

	config.NewWatcher(k8sClient, Namespace, ConfigMapName).
		OnAddNotify(legacyEventSender).
		OnUpdateNotify(legacyEventSender).
		OnDeleteNotify(legacyEventSender).
		OnAddNotify(ceEventSender).
		OnUpdateNotify(ceEventSender).
		OnDeleteNotify(ceEventSender).
		OnDeleteNotifyMe().
		Watch()

	subscription.NewWatcher(dynamicClient).
		OnAddNotify(ceEventSender).
		OnUpdateNotify(ceEventSender).
		OnDeleteNotify(ceEventSender).
		OnAddNotify(legacyEventSender).
		OnUpdateNotify(legacyEventSender).
		OnDeleteNotify(legacyEventSender).
		Watch()

	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
