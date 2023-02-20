# Publisher

This tool is used to publish continues legacy and cloudevents the Eventing publisher proxy Kyma.

## Application to event type mapping

Currently, it is hardcoded which application sends which event type as follows:

| Application   | Event type                                       |
|---------------|--------------------------------------------------|
| commerce      | order.created                                    |
| appname       | DocuSing_BO.Account_DocuSign.Updated             |
| no-app        | New.Some-Other.Order-äöüÄÖÜβ.Final.C-r-e-a-t-e-d |

## Usage

### Deploy

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make publisher-deploy
   ```

### Stop

This is useful in case we want to stop sending events to the Eventing publisher proxy, but preserve the Eventing
infrastructure inside the `eventing-test` namespace.

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make publisher-stop
   ```

   > Note: To continue sending events, deploy the publisher again.

### Logs

1. Connect to a Kubernetes cluster.
2. Publisher logs:
   ```bash
   stern -n eventing-test publisher -c publisher --since 1s
   ```

### Cleanup

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make publisher-delete
   ```

## Future enhancements

- Add instructions for publishing events with an empty prefix.
- Automate publishing based on the Eventing prefix configured in the connected Kyma cluster.
