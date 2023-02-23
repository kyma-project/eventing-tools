# Loadtest

This tool is used to generate continues load on the Eventing components.
It does that by sending Cloudevents to the Eventing publisher proxy forever and consume them inside Kyma functions.
Ideally, it should be used when JetStream is used as the active Eventing backend.

## Eventing infrastructure

For testing against **Legacy Events** set `use_legacy events` to `true`. Leave it to `false` to test against **Cloud Events**.

The load test will create:

1. **Two** Kyma function subscribers.
2. **Two** Kyma subscriptions.

| Namespace      | Subscription   | Subscriber              |
|----------------|----------------|-------------------------|
| eventing-test  | subscription-0 | loadtest-subscriber-0   |
| eventing-test  | subscription-1 | loadtest-subscriber-1   |

> Note: Each Kyma subscription is configured with `N` unique event types.

## Event types naming conventions

1. **There are two event names**:
    - The `subscription-0` has `order.created`.
    - The `subscription-1` has `order.updated`.

2. **Events carry the `eps` information**:
 - EPS **50**: `order.created.v0050`.
 - EPS **90**: `order.created.v0090`.

*Cloud Events* have the `eps` encoded at the end as part of the `version`.
*Legacy Events* have the `eps`  body in the `event-type` field of the body.

> Note: Encoding the EPS in the event type version is used only for debugging purposes and is not a production use-case.

## Configurations

Command-line arguments:

| Arguments | Description                                                                        | Default  |
|-----------|------------------------------------------------------------------------------------|----------|
| addr      | The loadtest server address used by the Kubernetes liveness and readiness probes.  | :8888    |

[ConfigMap](../../resources/loadtest/300-configmap.yaml) to change the loadtest behaviour at runtime:

| Config                  | Description                                                                                                             | Default                                             |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| publish_endpoint        | The Eventing publisher proxy CloudEvents endpoint.                                                                      | http://eventing-publisher-proxy.kyma-system/publish |
| use_legacy_events:      | Use `legacy events` or `CloudEvents`.                                                                                  | false                                               |
| event_source:           | The event source for both subscriptions.                                                                                | "noapp"                                             |
| version_format:         | The format string used to create a the event-version for both subscriptions.                                            | v%04d                                               |
| max_inflight_messages_0 | The max inflight messages for the first subscription.                                                                   | 10                                                  |
| max_inflight_messages_1 | The max inflight messages for the second subscription.                                                                  | 10                                                  |
| event_name_0            | The event name used in the event-type for the first subscription.                                                       | order.created                                       |
| event_name_1            | The event name used in the event-type for the second subscription.                                                      | order.updated                                       |
| generate_count_0        | The number of event types to generate for the first subscription.                                                       | 1                                                   |
| generate_count_1        | The number of event types to generate for the second subscription.                                                      | 1                                                   |
| eps_start_0             | The EPS start for the first event type of the first subscription.                                                       | 1                                                   |
| eps_start_1             | The EPS start for the first event type of the second subscription.                                                      | 1                                                   |
| eps_increment_0         | The EPS increment for every event type (after the first) of the first subscription.                                     | 1                                                   |
| eps_increment_1         | The EPS increment for every event type (after the first) of the second subscription.                                    | 1                                                   |
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

- Generate subscriptions in a more generic way. Instead of creating only `2` subscriptions, create `N` subscriptions.
- Implement a graceful shutdown to delete Kyma subscriptions in the `eventing-test` namespace when the loadtest receives a
  termination signal. 
