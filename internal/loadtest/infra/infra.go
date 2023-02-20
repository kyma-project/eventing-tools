package infra

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	subscription "github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha2"
	"github.com/kyma-project/eventing-tools/internal/loadtest/config"

	"github.com/kyma-project/eventing-tools/internal/k8s"
	"github.com/kyma-project/eventing-tools/internal/logger"
)

const (
	labelKey   = "app"
	labelValue = "loadtest"
)

// Compile-time check for the implementation of interfaces.
var (
	_ config.AddNotifiable    = &Infra{}
	_ config.UpdateNotifiable = &Infra{}
	_ config.DeleteNotifiable = &Infra{}
)

type Infra struct {
	config    *config.Config
	subClient *subscription.Client
	k8sClient kubernetes.Interface
}

func New(config *config.Config, k8sConfig *rest.Config) *Infra {
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	subClient := subscription.NewClient(dynamicClient)
	return &Infra{
		config:    config,
		subClient: subClient,
		k8sClient: k8s.ClientOrDie(k8sConfig),
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

	// wait for the ConfigMap object to be completely deleted.
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

	// recreate the ConfigMap.
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

	typeFormat0 := fmt.Sprintf("%s.%s", i.config.EventName0, i.config.VersionFormat)
	if err := i.configureSubscription(
		Namespace, SubscriptionName0, Sink0, typeFormat0,
		i.config.MaxInflightMessages0, i.config.GenerateCount0, i.config.EpsStart0, i.config.EpsIncrement0,
	); err != nil {
		logger.LogIfError(err)
	}

	typeFormat1 := fmt.Sprintf("%s.%s", i.config.EventName1, i.config.VersionFormat)
	if err := i.configureSubscription(
		Namespace, SubscriptionName1, Sink1, typeFormat1,
		i.config.MaxInflightMessages1, i.config.GenerateCount1, i.config.EpsStart1, i.config.EpsIncrement1,
	); err != nil {
		logger.LogIfError(err)
	}
}

func (i *Infra) configureSubscription(namespace, name, sink, typeFormat, maxInflight string, count, start, increment int) error {
	if err := i.subClient.Delete(context.TODO(), namespace, name); err != nil && !errors.IsNotFound(err) {
		return err
	}

	s := subscription.New(
		subscription.WithName(name),
		subscription.WithNamespace(namespace),
		subscription.WithLabel(labelKey, labelValue),
		subscription.WithSink(sink),
		subscription.WithTypeMatching(subscription.Standard),
		subscription.WithConfig(subscription.MaxInFlightMessages, maxInflight),
		subscription.WithGeneratedTypes(typeFormat, count, start, increment),
	)
	for _, err := i.subClient.Get(context.TODO(), namespace, name); !errors.IsNotFound(err); _, err = i.subClient.Get(context.Background(), namespace, name) {
		time.Sleep(time.Second)
	}

	log.Printf("creating subscription %s with sink %s and %v types", s.Name, s.Spec.Sink, len(s.Spec.Types))
	if _, err := i.subClient.Create(context.TODO(), s); err != nil {
		log.Printf("error while creating subscription %s: %v", s.Name, err.Error())
		return err
	}

	return nil
}
