package cloudevents

import (
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"

	"github.com/kyma-project/eventing-tools/internal/logger"
)

func NewClientOrDie(transport *http.Transport) client.Client {
	p, err := cloudevents.NewHTTP(cloudevents.WithRoundTripper(transport))
	logger.FatalIfError(err)

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	logger.FatalIfError(err)

	return c
}
