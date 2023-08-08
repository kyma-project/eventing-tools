package sender

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/eventing-tools/internal/loadtest/config"
	"github.com/kyma-project/eventing-tools/internal/loadtest/events"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/cloudevent"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/interface"
	"github.com/kyma-project/eventing-tools/internal/loadtest/sender/legacyevent"
)

// compile-time check for interfaces implementation.
var _ config.Notifiable = &EventSender{}

// Sender sends cloud factories.
type EventSender struct {
	cfg                       config.Config
	ctx                       context.Context
	cancel                    context.CancelFunc
	ackC, nackC, undeliveredC chan events.Event
	wg                        sync.WaitGroup
	events                    chan events.Event
	senders                   []_interface.Sender
	limitC                    chan any
	writer                    *tabwriter.Writer
}

func (s *EventSender) NotifyAdd(configMap *corev1.ConfigMap) {
	// TODO update config
	config.Map(configMap, &s.cfg)
	s.senders = append(s.senders, cloudevent.NewSender(s.cfg, s.ackC, s.nackC, s.undeliveredC))
	s.senders = append(s.senders, legacyevent.NewSender(s.cfg, s.ackC, s.nackC, s.undeliveredC))

}

func (s *EventSender) NotifyUpdate(configMap *corev1.ConfigMap) {
	//TODO implement me
	config.Map(configMap, &s.cfg)
	s.senders = make([]_interface.Sender, 0)
	s.senders = append(s.senders, cloudevent.NewSender(s.cfg, s.ackC, s.nackC, s.undeliveredC))
	s.senders = append(s.senders, legacyevent.NewSender(s.cfg, s.ackC, s.nackC, s.undeliveredC))
}

func (s *EventSender) NotifyDelete(_ *corev1.ConfigMap) {
	s.senders = make([]_interface.Sender, 0)
}

func NewSender() (*EventSender, chan<- events.Event) {
	eventsC := make(chan events.Event)
	ackC := make(chan events.Event)
	nackC := make(chan events.Event)
	undeliveredC := make(chan events.Event)
	s := &EventSender{
		writer:       new(tabwriter.Writer),
		events:       eventsC,
		ackC:         ackC,
		nackC:        nackC,
		undeliveredC: undeliveredC,
	}
	s.writer.Init(os.Stdout, 8, 8, 0, '\t', tabwriter.AlignRight)
	return s, eventsC
}

func (s *EventSender) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.limitC = make(chan any, 3000)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.doAccounting()
	}()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.sendEvents()
	}()
}

func (s *EventSender) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *EventSender) sendEvents() {
	for {
		select {
		case e := <-s.events:
			// here we have to actually send messages to the sink
			for _, es := range s.senders {
				if es.Format() == e.Format {
					s.limitC <- struct{}{}
					go func() {
						es.SendEvent(e)
						<-s.limitC
					}()
					break
				}
			}
		case <-s.ctx.Done():
			return
		}
	}
}

type stat struct {
	acks, nacks, undelivered int
}

type stats map[string]stat

func (s *EventSender) doAccounting() {
	dur := 10 * time.Second
	tickOne := time.NewTicker(1 * time.Second)
	tickTen := time.NewTicker(dur)
	defer tickOne.Stop()
	defer tickTen.Stop()
	cs := make(stats)
	var all stat
	fmt.Fprintf(s.writer, "%s\t%s\t%s", "ACK", "NACK", "UNDELIVERED")
	fmt.Fprintf(s.writer, "\n%s\t%s\t%s", "----", "----", "----")
	s.writer.Flush()

	for {
		select {
		case <-tickOne.C:
			fmt.Fprintf(s.writer, "\n%v\t%v\t%v", all.acks, all.nacks, all.undelivered)
			s.writer.Flush()
			all = stat{}
		case <-tickTen.C:
			s.printStats(cs, dur)
			cs = make(stats)
		case e := <-s.ackC:
			all.acks++
			st := cs["["+e.Format.String()+"]"+e.Source+"/"+e.EventType]
			st.acks++
			cs["["+e.Format.String()+"]"+e.Source+"/"+e.EventType] = st
		case e := <-s.nackC:
			all.nacks++
			st := cs["["+e.Format.String()+"]"+e.Source+"/"+e.EventType]
			st.nacks++
			cs["["+e.Format.String()+"]"+e.Source+"/"+e.EventType] = st
		case e := <-s.undeliveredC:
			all.undelivered++
			st := cs["["+e.Format.String()+"]"+e.Source+"/"+e.EventType]
			st.undelivered++
			cs["["+e.Format.String()+"]"+e.Source+"/"+e.EventType] = st
		case <-s.ctx.Done():
			fmt.Fprintf(s.writer, "\n%v\t%v\t%v\n", all.acks, all.nacks, all.undelivered)
			s.writer.Flush()
			s.printStats(cs, dur)
			return
		}
	}
}

func (s *EventSender) printStats(cs stats, interval time.Duration) {
	// initialize tabwriter

	fmt.Fprint(s.writer, "\n--------------------------------------------")
	s.writer.Flush()

	fmt.Fprintf(s.writer, "\n%s\t%s\t%s\t%s", "TYPE", "ACK", "NACK", "UNDELIVERED")
	fmt.Fprintf(s.writer, "\n%s\t%s\t%s\t%s\t", "----", "----", "----", "----")

	var ak []string
	for k := range cs {
		ak = append(ak, k)
	}
	sort.Strings(ak)
	for _, k := range ak {
		fmt.Fprintf(s.writer, "\n%v\t%v\t%v\t%v", k, float64(cs[k].acks)/interval.Seconds(), float64(cs[k].nacks)/interval.Seconds(), float64(cs[k].undelivered)/interval.Seconds())
	}
	s.writer.Flush()
	fmt.Fprintln(s.writer, "\n--------------------------------------------")
	s.writer.Flush()
	fmt.Fprintf(s.writer, "%s\t%s\t%s\t", "ACK", "NACK", "UNDELIVERED")
	fmt.Fprintf(s.writer, "\n%s\t%s\t%s\t", "----", "----", "----")
}
