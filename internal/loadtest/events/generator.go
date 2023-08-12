package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-project/eventing-tools/internal/loadtest/api/subscription/v1alpha2"
)

type Generator struct {
	source    string
	eps       int
	starttime string
	cancel    context.CancelFunc
	eventtype string
	id        int
	sink      string
	c         chan<- Event
	wg        sync.WaitGroup
	format    EventFormat
	lock      sync.Mutex
}

type Event struct {
	EventType string
	Source    string
	Sink      string
	ID        int
	StartTime string
	Format    EventFormat
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

func updateGeneratorFormat(sub *v1alpha2.Subscription, gen *Generator) {
	f := EventFormatFromString(sub.GetLabels()[formatLabel])
	if gen.format != f {
		gen.Update(f)
	}
}

func (e *Generator) fillChan(ctx context.Context, c chan<- Event) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	remaining := e.eps
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("DONE counter %v.%v\n", e.source, e.eventtype)
			return
		case <-t.C:
			remaining = e.eps
		default:
			if remaining > 0 {
				// ensure nowone resets the id atm
				e.lock.Lock()
				id := e.id
				e.id++
				e.lock.Unlock()

				c <- Event{
					EventType: e.eventtype,
					Source:    e.source,
					Sink:      e.sink,
					ID:        id,
					StartTime: e.starttime,
					Format:    e.format,
				}
				remaining--
			}
		}
	}
}

func (e *Generator) Update(format EventFormat) {
	// let's stop all concurrency for a while
	e.lock.Lock()
	e.format = format
	e.id = 0
	e.starttime = time.Now().Format("2006-01-02T15:04:05")
	e.lock.Unlock()
}

func (e *Generator) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.fillChan(ctx, e.c)
	}()
}

func (e *Generator) Stop() {
	e.cancel()
	e.wg.Wait()
}

func EventFormatFromString(format string) EventFormat {
	if format == "legacy" {
		return Legacy
	}
	return CloudEvent
}

type EventFormat int

func (e EventFormat) String() string {
	if e == Legacy {
		return "legacy"
	}
	return "cloudevent"
}

const (
	Legacy EventFormat = iota
	CloudEvent
)
