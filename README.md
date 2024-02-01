[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/eventing-tools)](https://api.reuse.software/info/github.com/kyma-project/eventing-tools)
# Eventing Tools

This repository contains useful tools for Kyma Eventing:

- [Loadtest](./docs/loadtest/README.md) for load-testing the Eventing components by continuously sending and receiving cloudevents.
- [Publisher](./docs/publisher/README.md) for publishing legacy and cloudevents continuously.
- [Subscriber](./docs/subscriber/README.md) for consuming cloudevents.
- [Linter settings for Kyma modules](https://github.tools.sap/skydivingtunas/how-to/blob/master/golangci-lint-configuration.md)
> Note: The `publisher` and the `subscriber` are using the same Eventing infrastructure (e.g. Kyma subscriptions, subscribers, ...etc).
> This could be useful when verifying Eventing on any cluster by deploying both applications and watching their logs.

## Prerequisites

- [ko](https://github.com/google/ko).
- [stern](https://github.com/stern/stern).
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl).

## Contributing
<!--- mandatory section - do not change this! --->

See [CONTRIBUTING.md](CONTRIBUTING.md)

## Code of Conduct
<!--- mandatory section - do not change this! --->

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Licensing
<!--- mandatory section - do not change this! --->

See the [LICENSE file](./LICENSE)
