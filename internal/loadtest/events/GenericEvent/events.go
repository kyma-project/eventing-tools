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
	source        string
	VersionFormat string
	EventName     string
	eps           int
	Starttime     string
	feedback      chan int
	counter       chan int
	success       chan int
	events        chan *Event
	cancel        context.CancelFunc
	ctx           context.Context
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
		VersionFormat: format,
		EventName:     name,
		eps:           eps,
		Starttime:     time.Now().Format("2006-01-02T15:04:05"),
		source:        source,
	}
	e.Start()
	return &e
}

func (e *Event) handleSuccess() {
	var successes *tree.Node
	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-e.ctx.Done():
			fmt.Printf("%v.%v: %v\n", e.Starttime, e.EventName, successes)
			return
		case val := <-e.success:
			successes = tree.InsertInt(successes, val)
		case <-t.C:
			fmt.Printf("%v.%v.%v: %v\n", e.Starttime, e.EventName, e.VersionFormat, successes)
		}
	}
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
			log.Print(e.eps)
			for i := 0; i < e.eps; i++ {
				e.events <- e
			}
		case <-e.ctx.Done():
			close(e.events)
			return
		}
	}
}

// func Generate(conf *config.Config) []Event {
// 	count := conf.ComputeEventsCount()
// 	events := make([]Event, 0, count)
//
// 	generate := func(format, name string, start, increment, count int) {
// 		// for i, Eps := 0, start; i < count; i, Eps = i+1, start+(increment*(i+1)) {
// 		// 	// event := newEvent(format, name, Eps)
// 		// 	// events = append(events, event)
// 		// }
// 	}
//
// 	if conf.IsVersionFormatEmpty() {
// 		return events
// 	}
// 	if !conf.IsEventName0Empty() {
// 		generate(conf.VersionFormat, conf.EventName0, conf.EpsStart0, conf.EpsIncrement0, conf.GenerateCount0)
// 	}
// 	if !conf.IsEventName1Empty() {
// 		generate(conf.VersionFormat, conf.EventName1, conf.EpsStart1, conf.EpsIncrement1, conf.GenerateCount1)
// 	}
//
// 	return events
// }

func (e *Event) Stop() {
	e.cancel()
}

func (e *Event) Start() {
	e.events = make(chan *Event, e.eps)
	e.counter = make(chan int, e.eps*4)
	e.feedback = make(chan int, e.eps*4)
	e.success = make(chan int, e.eps*4)
	e.ctx, e.cancel = context.WithCancel(context.Background())
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
		EventTypeVersion: e.version(),
		EventTime:        time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventTracing:     true,
	}
}

func (e *Event) ToCloudEvent(seq int, evtSrc string) (cev2.Event, error) {
	ce := cev2.NewEvent()
	ce.SetType(e.EventType())
	ce.SetSource(evtSrc)

	d := payload.DTO{
		Start: e.Starttime,
		Value: seq,
	}
	err := ce.SetData(cev2.ApplicationJSON, d)
	return ce, err
}

func (e *Event) version() string {
	return fmt.Sprintf(e.VersionFormat, e.Eps)
}

func (e *Event) EventType() string {
	return fmt.Sprintf("%s.%s", e.EventName, e.version())
}
