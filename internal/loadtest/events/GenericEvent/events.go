package GenericEvent

import (
	"context"
	"fmt"
	"log"
	"time"

	cev2 "github.com/cloudevents/sdk-go/v2"

	"github.com/kyma-project/eventing-tools/internal/loadtest/events/payload"
	"github.com/kyma-project/eventing-tools/internal/tree"
)

type Event struct {
	source    string
	version   string
	EventName string
	eps       int
	Starttime string
	feedback  chan int
	counter   chan int
	success   chan int
	events    chan *Event
	cancel    context.CancelFunc
	ctx       context.Context
	successes *tree.Node
	eventtype string
	ce        cev2.Event
}

func (e *Event) Events() <-chan *Event {
	return e.events
}

func (e *Event) Source() string {
	return e.source
}

func (e *Event) Feedback() chan<- int {
	return e.feedback
}

func (e *Event) Success() chan<- int {
	return e.success
}

func (e *Event) Eps() int {
	return e.eps
}

func (e *Event) Counter() <-chan int {
	return e.counter

}

func NewEvent(format, name, source string, eps int) *Event {
	e := Event{
		version:   format,
		EventName: name,
		eps:       eps,
		Starttime: time.Now().Format("2006-01-02T15:04:05"),
		source:    source,
		eventtype: fmt.Sprintf("%s.%s", name, format),
	}
	ce := cev2.NewEvent()
	ce.SetType(e.eventtype)
	ce.SetSource(source)
	e.ce = ce
	e.Start()
	return &e
}

func (e *Event) handleSuccess() {
	for {
		select {
		case <-e.ctx.Done():
			fmt.Printf("%v.%v: %v\n", e.Starttime, e.EventName, e.successes)
			return
		case val := <-e.success:
			e.successes = tree.InsertInt(e.successes, val)
		}
	}
}

func (e *Event) PrintStats() {
	fmt.Printf("%v.%v.%v: %v\n", e.Starttime, e.EventName, e.version, e.successes)
}

func (e *Event) fillCounter() {
	var c int
	var cur int
	list := make([]int, 0)
	for {
		select {
		case <-e.ctx.Done():
			return
		case val := <-e.feedback:
			list = append(list, val)
		default:
			if len(list) > 0 {
				cur, list = list[0], list[1:]
				e.counter <- cur
				continue
			}
			e.counter <- c
			c++
		}
	}
}

func (e *Event) queueEvent() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from: ", r)
		}
	}()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	// queue event immediately
	for {
		select {
		case <-t.C:
			for i := 0; i < e.eps; i++ {
				e.events <- e
			}
		case <-e.ctx.Done():
			close(e.events)
			return
		}
	}
}

func (e *Event) Stop() {
	e.cancel()
}

func (e *Event) Start() {
	e.events = make(chan *Event, e.eps)
	e.counter = make(chan int, e.eps*4)
	e.feedback = make(chan int, e.eps*4)
	e.success = make(chan int, e.eps*4)
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.successes = nil
	go e.fillCounter()
	go e.handleSuccess()
	go e.queueEvent()
}

func (e *Event) ToLegacyEvent(seq int) payload.LegacyEvent {
	d := payload.DTO{
		Start: e.Starttime,
		Value: seq,
	}
	return payload.LegacyEvent{
		Data:             d,
		EventType:        e.EventName,
		EventTypeVersion: e.version,
		EventTime:        time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventTracing:     true,
	}
}

func (e *Event) ToCloudEvent(seq int) (cev2.Event, error) {

	// d := payload.DTO{
	// 	Start: e.Starttime,
	// 	Value: seq,
	// }
	// err := e.ce.SetData(cev2.ApplicationJSON, d)
	// return e.ce, err
	return e.ce, nil
}
