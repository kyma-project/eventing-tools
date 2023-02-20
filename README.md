# Eventing Tools

This repository contains useful tools for Kyma Eventing:

- [Loadtest](./docs/loadtest/README.md) for load-testing the Eventing components by continuously sending and receiving cloudevents.
- [Publisher](./docs/publisher/README.md) for publishing legacy and cloudevents continuously.
- [Subscriber](./docs/subscriber/README.md) for consuming cloudevents.

> Note: The `publisher` and the `subscriber` are using the same Eventing infrastructure (e.g. Kyma subscriptions, subscribers, ...etc).
> This could be useful when verifying Eventing on any cluster by deploying both applications and watching their logs.

## Prerequisites

- [ko](https://github.com/google/ko).
- [stern](https://github.com/stern/stern).
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl).

## Future enhancements

- Instead of duplicating and hard-coding the common Eventing infrastructure (e.g. Kyma subscriptions, subscribers, ...etc) between the `publisher` and the `subscriber`, have a central place for common configurations where both applications can use at runtime.
- Use a structured logger for all applications.
