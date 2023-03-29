package publisher

import (
	"context"
	"net/http"

	"github.com/kyma-project/eventing-tools/internal/client/cloudevents"
	pkghttp "github.com/kyma-project/eventing-tools/internal/client/http"
	"github.com/kyma-project/eventing-tools/internal/client/transport"
	"github.com/kyma-project/eventing-tools/internal/logger"
	"github.com/kyma-project/eventing-tools/internal/probes"
	"github.com/kyma-project/eventing-tools/internal/publisher"
	"github.com/kyma-project/eventing-tools/internal/publisher/config"
)

func Start() {
	conf := config.ProcessOrDie()
	t := transport.New(conf.MaxIdleConns, conf.MaxConnsPerHost, conf.MaxIdleConnsPerHost, conf.IdleConnTimeout)
	clientHTTP := pkghttp.NewClient(t.Clone())
	clientCE := cloudevents.NewClientOrDie(t.Clone())
	publisher.Start(context.Background(), clientCE, clientHTTP, conf)
	http.HandleFunc(probes.EndpointReadyz, probes.DefaultHandler)
	http.HandleFunc(probes.EndpointHealthz, probes.DefaultHandler)
	logger.LogIfError(http.ListenAndServe(conf.ServerAddress, nil))
}
