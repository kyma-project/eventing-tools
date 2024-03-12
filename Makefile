
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# =============================================================================
# Lint
# =============================================================================

GOLANG_CI_LINT_VERSION ?= v1.55.2
.PHONY: golangci-lint
golangci-lint:
	test -s $(LOCALBIN)/golangci-lint && $(LOCALBIN)/golangci-lint version | grep -q $(GOLANG_CI_LINT_VERSION) || \
		GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)

.PHONY: lint
lint: golangci-lint## Check lint issues using `golangci-lint`
	$(LOCALBIN)/golangci-lint run --timeout 5m

.PHONY: lint-compact
lint-compact: golangci-lint## Check lint issues using `golangci-lint` in compact result format
	$(LOCALBIN)/golangci-lint run --timeout 5m --print-issued-lines=false

.PHONY: lint-fix
lint-fix: golangci-lint## Check and fix lint issues using `golangci-lint`
	$(LOCALBIN)/golangci-lint run --fix --timeout 5m


# =============================================================================
# Loadtest
# =============================================================================

.PHONY: loadtest-deploy
loadtest-deploy:
	@kubectl apply -k resources/loadtest/base

.PHONY: loadtest-deploy-ko
loadtest-deploy-ko:
	@kustomize build resources/loadtest/ko | ko apply -f -

.PHONY: loadtest-start
loadtest-start:
	@kubectl scale deployment -n eventing-test loadtest-publisher --replicas 1

.PHONY: loadtest-stop
loadtest-stop:
	@kubectl scale deployment -n eventing-test loadtest-publisher --replicas 0

.PHONY: loadtest-delete
loadtest-delete:
	@kubectl delete -k resources/loadtest/base
	@kubectl delete subscriptions.eventing.kyma-project.io -n eventing-test -l 'app=loadtest' || true
