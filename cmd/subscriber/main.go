package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
	"github.com/kyma-project/eventing-tools/internal/subscriber"
)

func main() {
	http.HandleFunc("/", subscriber.Handler)
	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(readAddress(), nil))
}

func readAddress() (addr string) {
	flag.StringVar(&addr, "addr", ":8888", "HTTP Server listen address.")
	flag.Parse()
	log.Printf("Subscriber starting on port [%s]", addr)
	return addr
}
