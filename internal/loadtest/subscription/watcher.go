package subscription

import (
	"fmt"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha2"
	"github.com/kyma-project/eventing-tools/internal/logger"
)

const (
	defaultSize   = 10
	defaultResync = time.Minute
	loadtestLabel = "eventing-loadtest"
)

// compile-time check for interfaces implementation.
var (
	_ DeleteNotifiable = &Watcher{}
)

type Watcher struct {
	client               dynamic.Interface
	addNotifiableList    []AddNotifiable
	updateNotifiableList []UpdateNotifiable
	deleteNotifiableList []DeleteNotifiable
	stopCh               chan struct{}
}

func NewWatcher(client dynamic.Interface) *Watcher {
	return &Watcher{
		client:               client,
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

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		w.client,
		defaultResync,
		metav1.NamespaceAll,
		func(o *metav1.ListOptions) {
			o.LabelSelector = loadtestLabel
		},
	)
	si := factory.ForResource(v1alpha2.GroupVersion.WithResource(v1alpha2.Resource))
	subInf := si.Informer()
	_, err := subInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.addFunc,
		UpdateFunc: w.updateFunc,
		DeleteFunc: w.deleteFunc,
	})
	if err != nil {
		runtime.HandleError(err)
	}

	subInf.Run(w.stopCh)
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

func (w *Watcher) OnDeleteSubscription(_ *unstructured.Unstructured) {
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
	if u, ok := o.(*unstructured.Unstructured); ok {
		for _, n := range w.addNotifiableList {
			n.OnNewSubscription(u)
		}
	}
}

func (w *Watcher) updateFunc(o interface{}, n interface{}) {
	var (
		ok bool
		ou *unstructured.Unstructured
		nu *unstructured.Unstructured
	)

	if ou, ok = o.(*unstructured.Unstructured); !ok {
		logger.LogIfError(fmt.Errorf("cannot convert old object to unstructured"))
		return
	}
	if nu, ok = n.(*unstructured.Unstructured); !ok {
		logger.LogIfError(fmt.Errorf("cannot convert new object to unstructured"))
		return
	}

	if !reflect.DeepEqual(ou.Object["spec"], nu.Object["spec"]) {
		for _, n := range w.updateNotifiableList {
			n.OnChangedSubscription(nu)
		}
	}
}

func (w *Watcher) deleteFunc(o interface{}) {
	if u, ok := o.(*unstructured.Unstructured); ok {
		for _, n := range w.deleteNotifiableList {
			n.OnDeleteSubscription(u)
		}
	}
}
