package events

import (
	"context"
	"fmt"
	"time"

	cev2 "github.com/cloudevents/sdk-go/v2"

	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/tree"
)

type Event struct {
	VersionFormat string
	EventName     string
	Eps           int
	Start         string
	Feedback      chan int
	Counter       chan int
	Success       chan int
	cancel        context.CancelFunc
}

type LegacyEvent struct {
	Data             DTO    `json:"Data"`
	EventType        string `json:"Event-Type"`
	EventTypeVersion string `json:"Event-Type-Version"`
	EventTime        string `json:"Event-Time"`
	EventTracing     bool   `json:"Event-Tracing"`
}

type DTO struct {
	Start string `json:"StartTime"`
	Value int    `json:"Value"`
}

func newEvent(format, name string, eps int) Event {
	counter := make(chan int, eps*4)
	ctx, cancel := context.WithCancel(context.Background())
	feedback := make(chan int, eps*4)
	success := make(chan int, eps*4)
	go fillCounter(ctx, counter, feedback)
	e := Event{
		VersionFormat: format,
		EventName:     name,
		Eps:           eps,
		Counter:       counter,
		cancel:        cancel,
		Start:         time.Now().Format("2006-01-02T15:04:05"),
		Feedback:      feedback,
		Success:       success,
	}
	go e.handleSuccess(ctx, success)
	return e
}

func (e *Event) handleSuccess(ctx context.Context, success chan int) {
	var successes *tree.Node
	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("%v.%v: %v\n", e.Start, e.EventName, successes)
			return
		case val := <-success:
			successes = tree.InsertInt(successes, val)
		case <-t.C:
			fmt.Printf("%v.%v: %v\n", e.Start, e.EventName, successes)
		}
	}
}

func fillCounter(ctx context.Context, counter chan int, feedback chan int) {
	var c int
	var cur int
	list := make([]int, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case val := <-feedback:
			list = append(list, val)
		default:
			if len(list) > 0 {
				cur, list = list[0], list[1:]
				counter <- cur
				continue
			}
			counter <- c
			c++
		}
	}
}

func Generate(conf *config.Config) []Event {
	count := conf.ComputeEventsCount()
	events := make([]Event, 0, count)

	generate := func(format, name string, start, increment, count int) {
		for i, Eps := 0, start; i < count; i, Eps = i+1, start+(increment*(i+1)) {
			event := newEvent(format, name, Eps)
			events = append(events, event)
		}
	}

	if conf.IsVersionFormatEmpty() {
		return events
	}
	if !conf.IsEventName0Empty() {
		generate(conf.VersionFormat, conf.EventName0, conf.EpsStart0, conf.EpsIncrement0, conf.GenerateCount0)
	}
	if !conf.IsEventName1Empty() {
		generate(conf.VersionFormat, conf.EventName1, conf.EpsStart1, conf.EpsIncrement1, conf.GenerateCount1)
	}

	return events
}

func (e *Event) Stop() {
	e.cancel()
}

func (e *Event) ToLegacyEvent(seq int) LegacyEvent {
	d := DTO{
		Start: e.Start,
		Value: seq,
	}
	return LegacyEvent{
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

	d := DTO{
		Start: e.Start,
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
