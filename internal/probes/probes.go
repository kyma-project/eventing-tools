package probes

import "net/http"

const (
	EndpointReadyz  = "/readyz"
	EndpointHealthz = "/healthz"
)

func DefaultHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
