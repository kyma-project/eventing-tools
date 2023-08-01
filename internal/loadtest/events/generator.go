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

type EventStats struct {
	eventtype, source, startTime string
	sent                         int
}

func updateGeneratorFormat(sub *v1alpha2.Subscription, gen *Generator) {
	f := EventFormatFromString(sub.GetLabels()[formatLabel])
	if gen.format != f {
		gen.format = f
		gen.id = 0
		gen.starttime = time.Now().Format("2006-01-02T15:04:05")
	}
}

func (e *Generator) fillChan(ctx context.Context, c chan<- Event) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	remaining := e.eps
	id := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("DONE counter %v.%v\n", e.source, e.eventtype)
			return
		case <-t.C:
			remaining = e.eps
		default:
			if remaining > 0 {
				c <- Event{
					EventType: e.eventtype,
					Source:    e.source,
					Sink:      e.sink,
					ID:        id,
					StartTime: e.starttime,
					Format:    e.format,
				}
				remaining--
				id++
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
