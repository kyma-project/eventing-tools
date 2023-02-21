package v1alpha1

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v3"
	"github.com/kyma-project/eventing-tools/internal/loadtest/infra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha1"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"

	"github.com/kyma-project/eventing-tools/internal/k8s"
	"github.com/kyma-project/eventing-tools/internal/logger"
)

const (
	labelKey   = "app"
	labelValue = "loadtest"
)

// compile-time check for interfaces implementation.
var _ infra.InfraInterface = &Infra{}

type Infra struct {
	config         *config.Config
	v1alpha1Client *v1alpha1.Client
	k8sClient      kubernetes.Interface
}

func New(config *config.Config, k8sConfig *rest.Config) *Infra {
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	subscriptionClient := v1alpha1.NewClient(dynamicClient)
	return &Infra{
		config:         config,
		v1alpha1Client: subscriptionClient,
		k8sClient:      k8s.ClientOrDie(k8sConfig),
	}
}

func (i *Infra) NotifyAdd(cm *corev1.ConfigMap) {
	i.apply(cm)
}

func (i *Infra) NotifyUpdate(cm *corev1.ConfigMap) {
	i.apply(cm)
}

func (i *Infra) NotifyDelete(cm *corev1.ConfigMap) {
	fmt.Println("ConfigMap was deleted")

	// wait for the ConfigMap object to be completely deleted
	_ = retry.Do(func() error {
		_, err := i.k8sClient.CoreV1().ConfigMaps(cm.Namespace).Get(context.TODO(), cm.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("waiting for the ConfigMap to be deleted")
	},
		retry.Attempts(5),
		retry.Delay(time.Second*5),
		retry.DelayType(retry.FixedDelay),
	)

	// recreate the ConfigMap
	cm = cm.DeepCopy()
	cm.ResourceVersion = ""
	if _, err := i.k8sClient.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
		logger.LogIfError(err)
		return
	}

	fmt.Println("ConfigMap is restored")
}

func (i *Infra) apply(cm *corev1.ConfigMap) {
	config.Map(cm, i.config)
	format0 := fmt.Sprintf("sap.kyma.custom.%s.%s.%s", i.config.EventSource, i.config.EventName0, i.config.VersionFormat)
	if err := i.configureSubscription(
		infra.Namespace, infra.SubscriptionName0, infra.Sink0, format0,
		i.config.MaxInflightMessages0, i.config.GenerateCount0, i.config.EpsStart0, i.config.EpsIncrement0,
	); err != nil {
		logger.LogIfError(err)
	}

	format1 := fmt.Sprintf("sap.kyma.custom.%s.%s.%s", i.config.EventSource, i.config.EventName1, i.config.VersionFormat)
	if err := i.configureSubscription(
		infra.Namespace, infra.SubscriptionName1, infra.Sink1, format1,
		i.config.MaxInflightMessages1, i.config.GenerateCount1, i.config.EpsStart1, i.config.EpsIncrement1,
	); err != nil {
		logger.LogIfError(err)
	}
}

func (i *Infra) configureSubscription(namespace, name, sink, format string, maxInflight, count, start, increment int) error {
	if err := i.v1alpha1Client.Delete(context.TODO(), namespace, name); err != nil && !errors.IsNotFound(err) {
		return err
	}

	filters := v1alpha1.GenerateFilters(format, count, start, increment)
	opts := []v1alpha1.SubscriptionOption{
		v1alpha1.WithSink(sink),
		v1alpha1.WithFilters(filters),
		v1alpha1.WithMaxInFlightMessages(maxInflight),
		v1alpha1.WithLabels(map[string]string{labelKey: labelValue}),
	}
	for _, err := i.v1alpha1Client.Get(context.TODO(), namespace, name); !errors.IsNotFound(err); _, err = i.v1alpha1Client.Get(context.Background(), namespace, name) {
		time.Sleep(time.Second)
	}

	subscription := v1alpha1.NewSubscription(namespace, name, opts...)
	if _, err := i.v1alpha1Client.Create(context.TODO(), subscription); err != nil {
		return err
	}

	return nil
}
