# Loadtest

This version of the load tester can test subscription `v1alpha2` only. For `v1alpha1` we have a [legacy branch](https://github.com/kyma-project/eventing-tools/tree/loadtest-subscription-v1alpha1).
This tool is used to generate continues load on the Eventing components.
It does that by sending Cloudevents to the Eventing publisher proxy forever and consume them inside Kyma functions.
Ideally, it should be used when JetStream is used as the active Eventing backend.

## Subscriptions

The publisher watches all subscriptions in the cluster.
To activate sending events for one subscription you have to label that subscription with one of these labels:
- eventing-loadtest: legacy 
- eventing-loadtest: cloudevent

The load tester will then create events using the event types configured in the subscription and send them with the format configured in the label.

The load generated per event type will be extracted from the event type. To do this the event type must match the following pattern:
`<event-name>.v<Number of events per second>`, eg. `order.created.v500` will create events of type order.create.v500 and the sender will try to publish 500 events per second.

> Note: Encoding the EPS in the event type version is used only for debugging purposes and is not a production use-case.

> There is no special handling of subscriptions otherwise. The `sink` and also the `maxInFlight` settings will be respected.

This publisher works best in combination with the loadtest-subscriber. Two instances of the subscriber are deployed using the kustomize deployment. Configure the subscriptions to use those by setting the sink to `loadtest-subscriber-0.eventing-test.svc.cluster.local` or `loadtest-subscriber-1.eventing-test.svc.cluster.local`

## Configurations

Command-line arguments for the subscriber:

| Arguments | Description                                                                         | Default  |
|-----------|-------------------------------------------------------------------------------------|----------|
| --listen-port      | The loadtest server address used by the Kubernetes liveness and readiness probes. | :8888    |

[ConfigMap](../../resources/loadtest/300-configmap.yaml) to change the loadtest behaviour at runtime:

| Config                  | Description                                                                                                             | Default                                             |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| publish_endpoint        | The Eventing publisher proxy cloudevents endpoint.                                                                      | http://eventing-publisher-proxy.kyma-system/publish |
| eps_limit               | The EPS limit for the total publish requests sent in parallel.                                                          | 2000                                                |
| workers                 | The number of workers to publish events in parallel.                                                                    | 12                                                  |
| max_idle_conns          | The maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.                              | 100                                                 |
| max_conns_per_host      | The the total number of connections per host limit. Zero means no limit.                                                | 100                                                 |
| max_idle_conns_per_host | The maximum idle (keep-alive) connections to keep per-host. Zero means, use the default value which is 2.               | 100                                                 |
| idle_conn_timeout       | The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit. | 1m0s                                                |



## Usage

### Deploy

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make loadtest-deploy
   ```

### Change the loadtest behaviour at runtime:

1. Update the loadtest ConfigMap values:
   ```bash
   kubectl edit cm -n eventing-test loadtest
   ```

   > Note: After updating the loadtest-publisher  ConfigMap, its behaviour will change
   > automatically at runtime without the need to restart it.

### Stop

This is useful in case we want to stop sending events to the Eventing publisher proxy, but preserve the Eventing
infrastructure inside the `eventing-test` namespace.

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make loadtest-stop
   ```

   > Note: To continue sending events, deploy the loadtest again.

### Logs

1. Connect to a Kubernetes cluster.
2. Loadtest logs:
   ```bash
   stern -n eventing-test loadtest -c loadtest --since 1s
   ```
3. Subscriber logs:
   ```bash
   stern -n eventing-test sink -c function --since 1s | grep type
   ```

### Cleanup

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make loadtest-delete
   ```

## Future enhancements

- Implement a graceful shutdown to delete Kyma subscriptions in the `eventing-test` namespace when the loadtest receives a
  termination signal. 
