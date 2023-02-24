package config

import (
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/eventing-tools/internal/logger"
)

const (
	defaultSize         = 10
	defaultResync       = time.Minute
	fieldSelectorFormat = "metadata.name=%s"
)

// compile-time check for interfaces implementation.
var (
	_ DeleteNotifiable = &Watcher{}
)

type Watcher struct {
	client               kubernetes.Interface
	namespace            string
	name                 string
	addNotifiableList    []AddNotifiable
	updateNotifiableList []UpdateNotifiable
	deleteNotifiableList []DeleteNotifiable
	stopCh               chan struct{}
}

func NewWatcher(client kubernetes.Interface, namespace, name string) *Watcher {
	return &Watcher{
		client:               client,
		namespace:            namespace,
		name:                 name,
		addNotifiableList:    make([]AddNotifiable, 0, defaultSize),
		updateNotifiableList: make([]UpdateNotifiable, 0, defaultSize),
		deleteNotifiableList: make([]DeleteNotifiable, 0, defaultSize),
	}
}

func (w *Watcher) init() {
	w.stopCh = make(chan struct{})
}

func (w *Watcher) start() {
	defer runtime.HandleCrash()

	factory := informers.NewSharedInformerFactoryWithOptions(
		w.client,
		defaultResync,
		informers.WithNamespace(w.namespace),
		informers.WithTweakListOptions(func(o *metav1.ListOptions) {
			o.FieldSelector = fmt.Sprintf(fieldSelectorFormat, w.name)
		}),
	)
	configMapsInformer := factory.Core().V1().ConfigMaps().Informer()
	_, err := configMapsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.addFunc,
		UpdateFunc: w.updateFunc,
		DeleteFunc: w.deleteFunc,
	})
	if err != nil {
		runtime.HandleError(err)
	}

	factory.Start(w.stopCh)
	factory.WaitForCacheSync(w.stopCh)
	if !cache.WaitForCacheSync(w.stopCh, configMapsInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timeout waiting for caches to sync"))
		return
	}
}

func (w *Watcher) stop() {
	// recover from closing already closed channels
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic: ", r)
		}
	}()

	w.stopCh <- struct{}{}
	close(w.stopCh)
}

func (w *Watcher) Watch() {
	w.init()
	w.start()
}

func (w *Watcher) NotifyDelete(*corev1.ConfigMap) {
	w.stop()
	w.Watch()
}

func (w *Watcher) OnAddNotify(notifiable AddNotifiable) *Watcher {
	w.addNotifiableList = append(w.addNotifiableList, notifiable)
	return w
}

func (w *Watcher) OnUpdateNotify(notifiable UpdateNotifiable) *Watcher {
	w.updateNotifiableList = append(w.updateNotifiableList, notifiable)
	return w
}

func (w *Watcher) OnDeleteNotify(notifiable DeleteNotifiable) *Watcher {
	w.deleteNotifiableList = append(w.deleteNotifiableList, notifiable)
	return w
}

func (w *Watcher) OnDeleteNotifyMe() *Watcher {
	return w.OnDeleteNotify(w)
}

func (w *Watcher) addFunc(o interface{}) {
	if cm, ok := o.(*corev1.ConfigMap); ok {
		for _, n := range w.addNotifiableList {
			n.NotifyAdd(cm)
		}
	}
}

func (w *Watcher) updateFunc(o interface{}, n interface{}) {
	var (
		ok    bool
		oldCM *corev1.ConfigMap
		newCM *corev1.ConfigMap
	)

	if oldCM, ok = o.(*corev1.ConfigMap); !ok {
		logger.LogIfError(fmt.Errorf("cannot convert old object to configmap"))
		return
	}
	if newCM, ok = n.(*corev1.ConfigMap); !ok {
		logger.LogIfError(fmt.Errorf("cannot convert new object to configmap"))
		return
	}

	if !reflect.DeepEqual(oldCM.Data, newCM.Data) {
		for _, n := range w.updateNotifiableList {
			n.NotifyUpdate(newCM)
		}
	}
}

func (w *Watcher) deleteFunc(o interface{}) {
	if cm, ok := o.(*corev1.ConfigMap); ok {
		for _, n := range w.deleteNotifiableList {
			n.NotifyDelete(cm)
		}
	}
}
