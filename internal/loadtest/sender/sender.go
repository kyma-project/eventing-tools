package sender

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
)

// compile-time check for interfaces implementation.

// Sender sends cloud factories.
type EventSender struct {
	ctx                       context.Context
	cancel                    context.CancelFunc
	undelivered               int32
	ack                       int32
	nack                      int32
	mapLock                   sync.RWMutex
	wg                        sync.WaitGroup
	stopper                   sync.Mutex
	sender                    Sender
	acks, nacks, undelivereds chan int
	events                    chan events.Event
	cnclCtx                   context.Context
}

func NewSender() (*EventSender, chan<- events.Event) {
	eventsC := make(chan events.Event)
	s := &EventSender{
		events: eventsC,
	}
	return s, eventsC
	//s.undelivered = 0
	//s.ack = 0
	//s.nack = 0
	//s.factories = make(map[string][]*events.Generator)
	//s.factory = Factory.New()
	//s.sender = sender
	//s.acks = make(chan int)
	//s.nacks = make(chan int)
	//s.undelivereds = make(chan int)
	//
	//return s
}

func (s *EventSender) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()
		s.sendEvents()
	}()
}

func (s *EventSender) Stop() {
	s.cancel()
	s.wg.Wait()
}

//func (s *EventSender) NotifyAdd(cm *corev1.ConfigMap) {
//	s.stopper.Lock()
//	defer s.stopper.Unlock()
//	s.stop()
//	config.Map(cm, s.config)
//	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
//	s.sender.Init(t, s.config)
//	s.ctx, s.cancel = context.WithCancel(context.TODO())
//	s.process = make(chan bool, s.config.EpsLimit)
//	s.start()
//}
//
//func (s *EventSender) NotifyUpdate(cm *corev1.ConfigMap) {
//	s.stopper.Lock()
//	defer s.stopper.Unlock()
//	s.stop()
//	config.Map(cm, s.config)
//	t := transport.New(s.config.MaxIdleConns, s.config.MaxConnsPerHost, s.config.MaxIdleConnsPerHost, s.config.IdleConnTimeout)
//	s.sender.Init(t, s.config)
//	s.ctx, s.cancel = context.WithCancel(context.TODO())
//	s.process = make(chan bool, s.config.EpsLimit)
//	s.start()
//}
//
//func (s *EventSender) NotifyDelete(*corev1.ConfigMap) {
//	s.stopper.Lock()
//	defer s.stopper.Unlock()
//	s.stop()
//}

func (s *EventSender) sendEvents() {
	for {
		select {
		case e := <-s.events:
			fmt.Sprintf("%+v", e)
		case <-s.ctx.Done():
			return
		}
	}
}

//func (s *EventSender) ComputeTotalEventsPerSecond() int {
//	eps := 0
//	s.mapLock.RLock()
//	for _, subs := range s.factories {
//		for _, e := range subs {
//			eps += e.Eps()
//		}
//	}
//	s.mapLock.RUnlock()
//	return eps
//}

//func (s *EventSender) refillMaxEps(d time.Duration) {
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Println("Recovered in refillMaxEps", r)
//		}
//	}()
//
//	t := time.NewTicker(d)
//	for {
//		select {
//		case <-t.C:
//			for i := 0; i < s.config.EpsLimit; i++ {
//				s.process <- true
//			}
//		case <-s.ctx.Done():
//			return
//		}
//	}
//}
