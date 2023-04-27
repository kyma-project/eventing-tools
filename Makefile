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


# =============================================================================
# Validationtest
# =============================================================================

.PHONY: validationtest-deploy
validationtest-deploy:
	@kubectl apply -k resources/validationtest/base

.PHONY: validationtest-start
validationtest-start:
	@kubectl scale deployment -n eventing-test publisher --replicas 1

.PHONY: validationtest-stop
validationtest-stop:
	@kubectl scale deployment -n eventing-test publisher --replicas 0

.PHONY: validationtest-delete
validationtest-delete:
	@kubectl delete -k resources/validationtest/base
