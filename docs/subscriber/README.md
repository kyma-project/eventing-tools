# Subscriber

This tool is used to create Kyma subscriptions and their corresponding subscribers.

## Eventing infrastructure

1. **Three** microservice subscribers.
2. **Three** Kyma function subscribers.
3. **Six** Kyma subscriptions each of them has the following hardcoded filters:
   - `sap.kyma.custom.commerce.order.created.v1`.
   - `sap.kyma.custom.appname.DocuSing_BO.Account_DocuSign.Updated.v1`.
   - `sap.kyma.custom.no-app.New.Some-Other.Order-äöüÄÖÜβ.Final.C-r-e-a-t-e-d.v1`.

| Namespace       | Subscription            | Subscriber         |
|-----------------|-------------------------|--------------------|
| eventing-test   | event-subscription-0    | event-subscriber-0 |
| eventing-test   | event-subscription-1    | event-subscriber-1 |
| eventing-test   | event-subscription-2    | event-subscriber-2 |
| eventing-test   | function-subscription-0 | function-0         |
| eventing-test   | function-subscription-1 | function-1         |
| eventing-test   | function-subscription-2 | function-2         |

## Usage

### Deploy

1. Connect to a Kubernetes cluster.
2. Execute:
   ```bash
   make subscriber-deploy
   ```

### Logs

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
   make subscriber-delete
   ```

## Future enhancements

- Add instructions for deploying the Eventing infrastructure with an empty prefix.
- Automate deploying the Eventing infrastructure based on the Eventing prefix configured in the connected Kyma cluster.
