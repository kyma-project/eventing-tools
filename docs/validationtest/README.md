# Validationtest

This set of tools continuously publishes legacy and CloudEvents the Eventing publisher proxy in Kyma. In addition to the publisher, a set of subscribers is created with corresponding subscriptions.

## Application to event type mapping

Currently, it is hardcoded which application sends which event type:

| Application   | Event type                                       |
|---------------|--------------------------------------------------|
| commerce      | order.created                                    |
| appname       | DocuSing_BO.Account_DocuSign.Updated             |
| no-app        | New.Some-Other.Order-äöüÄÖÜβ.Final.C-r-e-a-t-e-d |

## Subscription

In total six subscribers will be created:
- **Three** microservice subscribers.
- **Three** Kyma function subscribers.

Those subscribers will be used as sinks in a set of six Kyma subscriptions.
Each of them has the following hardcoded filters:
   - `sap.kyma.custom.commerce.order.created.v1`.
   - `sap.kyma.custom.appname.DocuSing_BO.Account_DocuSign.Updated.v1`.
   - `sap.kyma.custom.no-app.New.Some-Other.Order-äöüÄÖÜβ.Final.C-r-e-a-t-e-d.v1`.

### Subscription -> Subscriber mapping

| Namespace       | Subscription            | Subscriber         |
|-----------------|-------------------------|--------------------|
| eventing-test   | event-subscription-0    | event-subscriber-0 |
| eventing-test   | event-subscription-1    | event-subscriber-1 |
| eventing-test   | event-subscription-2    | event-subscriber-2 |
| eventing-test   | function-subscription-0 | function-0         |
| eventing-test   | function-subscription-1 | function-1         |
| eventing-test   | function-subscription-2 | function-2         |


## Usage

### Deploy Validationtest 

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make validationtest-deploy
   ```

### Stop Validationtest

You can stop sending events to the Eventing publisher proxy while preserving the Eventing infrastructure inside the `eventing-test` Namespace with the following command:

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make validationtest-stop
   ```

   > Note: To continue sending events, deploy the publisher again.

### Logs

#### View Publisher Logs

1. Connect to a Kubernetes cluster.
2. Publisher logs:
   ```bash
   stern -n eventing-test publisher -c publisher --since 1s
   ```
#### View Subscriber Logs

1. Connect to a Kubernetes cluster.
2. Functions logs:
   ```bash
   stern -n eventing-test function -c function --since 1s | grep mode
   ```
3. Microservices logs:
   ```bash
   stern -n eventing-test subscriber -c event-subscriber --since 1s | grep mode
   ```

### Cleanup

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make validationtest-delete
   ```

## Future enhancements
- Add instructions for deploying the Eventing infrastructure with an empty prefix.
- Automate deploying the Eventing infrastructure based on the Eventing prefix configured in the connected Kyma cluster.
