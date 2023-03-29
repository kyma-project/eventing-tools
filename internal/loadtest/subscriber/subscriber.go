package subscriber

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/eventing-tools/internal/loadtest/events"

	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
	"github.com/kyma-project/eventing-tools/internal/tree"
)

var evtChn chan *event.Event
var received map[string]*tree.Node

func Start(port int) {
	http.HandleFunc("/", handler) // sink
	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)

	evtChn = make(chan *event.Event, 100000)
	received = make(map[string]*tree.Node)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go processEvents(ctx)

	logger.LogIfError(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		log.Printf("failed to parse CloudEvent from request: %v", err)
		return
	}
	evtChn <- event
	w.WriteHeader(http.StatusOK)
}

func processEvents(ctx context.Context) {
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case e := <-evtChn:
			d := &events.DTO{}
			err := e.DataAs(d)
			if err != nil {
				log.Print(err)
				continue
			}
			received[fmt.Sprintf("%v.%v", d.Start, e.Type())] = tree.InsertInt(received[fmt.Sprintf("%v.%v", d.Start, e.Type())], d.Value)
		case <-timer.C:
			printStats()
		case <-ctx.Done():
			return
		}
	}
}

func printStats() {
	if len(received) == 0 {
		fmt.Println("Nothing received")
		return
	}
	fmt.Println("--------")
	var keys []string
	for k := range received {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		t := received[k]
		fmt.Printf("%v: %v\n", k, t)
	}
}
