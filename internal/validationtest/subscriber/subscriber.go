package subscriber

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
	"github.com/kyma-project/eventing-tools/internal/subscriber"
)

func Start(port int) {
	http.HandleFunc("/", subscriber.Handler)
	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
