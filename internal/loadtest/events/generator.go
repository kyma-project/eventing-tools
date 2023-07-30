package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha2"
)

type Generator struct {
	source      string
	version     string
	name        string
	eps         int
	starttime   string
	cancel      context.CancelFunc
	eventtype   string
	running     bool
	counter     int
	counterLock sync.Mutex
	id          int
	sink        string
	c           chan<- Event
	wg          sync.WaitGroup
	format      EventFormat
}

type Event struct {
	eventtype string
	source    string
	sink      string
	id        int
	startTime string
	format    EventFormat
}

func NewGenerator(eventType, source string, eps int, format string, senderC chan<- Event) *Generator {
	e := Generator{
		eps:       eps,
		starttime: time.Now().Format("2006-01-02T15:04:05"),
		source:    source,
		eventtype: eventType,
		format:    EventFormatFromString(format),
		c:         senderC,
	}
	return &e
}

type EventStats struct {
	eventtype, source, startTime string
	sent                         int
}

func updateGeneratorFormat(sub *v1alpha2.Subscription, gen *Generator) {
	f := EventFormatFromString(sub.GetLabels()[formatLabel])
	if gen.format != f {
		gen.format = f
		gen.starttime = time.Now().Format("2006-01-02T15:04:05")
	}
}

func (e *Generator) fillChan(ctx context.Context, c chan<- Event) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	remaining := e.eps
	sent := 0
	id := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("DONE counter %v.%v\n", e.source, e.eventtype)
			return
		case <-t.C:
			remaining = e.eps
			//stats <- EventStats{
			//	eventtype: e.eventtype,
			//	source:    e.source,
			//	startTime: e.starttime,
			//	sent:      sent,
			//}
			sent = 0
		default:
			if remaining > 0 {
				c <- Event{
					eventtype: e.eventtype,
					source:    e.source,
					sink:      e.sink,
					id:        id,
					startTime: e.starttime,
					format:    e.format,
				}
				sent++
				remaining--
			}
		}
	}
}

func (e *Generator) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	go func() {
		e.wg.Add(1)
		defer e.wg.Done()
		e.fillChan(ctx, e.c)
	}()
}

func (e *Generator) Stop() {
	e.cancel()
	e.wg.Wait()
}

// func (e *Generator) ToLegacyEvent(seq int) payload.LegacyEvent {
// 	d := payload.DTO{
// 		Start: e.starttime,
// 		Value: seq,
// 	}
// 	return payload.LegacyEvent{
// 		Data:             d,
// 		EventType:        e.name,
// 		EventTypeVersion: e.version,
// 		EventTime:        time.Now().Format("2006-01-02T15:04:05.000Z"),
// 		EventTracing:     true,
// 	}
// }
//
// func (e *Generator) ToCloudEvent(seq int) (cev2.Event, error) {
// 	ce := cev2.NewEvent()
// 	ce.SetType(e.eventtype)
// 	ce.SetSource(e.source)
// 	d := payload.DTO{
// 		Start: e.starttime,
// 		Value: seq,
// 	}
// 	err := ce.SetData(cev2.ApplicationJSON, d)
// 	return ce, err
// }

func EventFormatFromString(format string) EventFormat {
	if format == "legacy" {
		return Legacy
	}
	return CloudEvent
}

type EventFormat int

const (
	Legacy EventFormat = iota
	CloudEvent
)
