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
