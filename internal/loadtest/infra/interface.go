package infra

import "github.com/kyma-project/eventing-tools/internal/loadtest/config"

type InfraInterface interface {
	config.AddNotifiable
	config.UpdateNotifiable
	config.DeleteNotifiable
}
