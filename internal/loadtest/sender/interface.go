package sender

import "github.com/kyma-project/eventing-tools/internal/loadtest/config"

type Sender interface {
	config.AddNotifiable
	config.UpdateNotifiable
	config.DeleteNotifiable
}
