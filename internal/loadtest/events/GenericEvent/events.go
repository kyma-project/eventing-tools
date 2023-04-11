package GenericEvent

import (
	"context"
	"fmt"
	"log"
	"sync"
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
	successes *tree.Node
	eventtype string
	ce        cev2.Event
	wg        *sync.WaitGroup
	running   bool
	stopper   sync.Mutex
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
		wg:        &sync.WaitGroup{},
	}
	ce := cev2.NewEvent()
	ce.SetType(e.eventtype)
	ce.SetSource(source)
	e.ce = ce
	return &e
}

func (e *Event) handleSuccess(ctx context.Context) {
	defer e.wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("%v.%v: %v\n", e.Starttime, e.EventName, e.successes)
			fmt.Printf("DONE success %v.%v\n", e.source, e.eventtype)
			return
		case val := <-e.success:
			e.successes = tree.InsertInt(e.successes, val)
		}
	}
}

func (e *Event) PrintStats() {
	fmt.Printf("%v.%v.%v.%v: %v\n", e.Starttime, e.source, e.EventName, e.version, e.successes)
}

func (e *Event) fillCounter(ctx context.Context) {
	defer e.wg.Done()
	var c int
	var next int
	for {
		select {
		case next = <-e.feedback:
			break
		default:
			next = c
			c++
		}
		select {
		case <-ctx.Done():
			fmt.Printf("DONE counter %v.%v\n", e.source, e.eventtype)
			return
		case e.counter <- next:
			break

		}
	}
}

func (e *Event) queueEvent(ctx context.Context) {
	defer e.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from in queueEvent: ", r)
		}
	}()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	// queue event immediately
	for {
		select {
		case <-t.C:
			for i := 0; i < e.eps; i++ {
				select {
				case <-ctx.Done():
					close(e.events)
					fmt.Printf("DONE queue %v.%v\n", e.source, e.eventtype)
					return
				case e.events <- e:
					continue
				}
			}
		case <-ctx.Done():
			close(e.events)
			fmt.Printf("DONE queue %v.%v\n", e.source, e.eventtype)
			return
		}
	}
}

func (e *Event) Stop() {
	e.stopper.Lock()
	if !e.running {
		return
	}
	e.cancel()
	fmt.Printf("waiting for %v.%v\n", e.source, e.eventtype)
	e.wg.Wait()
	fmt.Printf("DONE waiting for %v.%v\n", e.source, e.eventtype)
	e.running = false
	e.stopper.Unlock()
}

func (e *Event) Start() {
	if e.running {
		return
	}
	e.running = true
	e.events = make(chan *Event, e.eps)
	e.counter = make(chan int, e.eps*4)
	e.feedback = make(chan int, e.eps*4)
	e.success = make(chan int, e.eps*4)
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.successes = nil
	e.wg.Add(1)
	go e.fillCounter(ctx)
	e.wg.Add(1)
	go e.handleSuccess(ctx)
	e.wg.Add(1)
	go e.queueEvent(ctx)
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

	d := payload.DTO{
		Start: e.Starttime,
		Value: seq,
	}
	err := e.ce.SetData(cev2.ApplicationJSON, d)
	return e.ce, err
}
