package infra

const (
	Namespace = "eventing-test"

	ConfigMapName = "loadtest-publisher"

	SubscriptionName0 = "subscription-0"
	SubscriptionName1 = "subscription-1"

	Sink0 = "http://loadtest-subscriber-0.eventing-test.svc.cluster.local"
	Sink1 = "http://loadtest-subscriber-1.eventing-test.svc.cluster.local"
)
