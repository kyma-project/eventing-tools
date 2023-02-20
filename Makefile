# =============================================================================
# Common
# =============================================================================

DIR := $(realpath $(shell pwd))

# =============================================================================
# Loadtest
# =============================================================================

.PHONY: loadtest-deploy
loadtest-deploy:
	@kubectl apply -f ${DIR}/resources/common/100-namespace.yaml -f ${DIR}/resources/loadtest/400-rbac.yaml -f ${DIR}/resources/loadtest/200-service.yaml
	@kubectl create -f ${DIR}/resources/loadtest/300-configmap.yaml || true
	@ko apply -f ${DIR}/resources/loadtest/500-sender.yaml -f ${DIR}/resources/loadtest/100-loadsubscriber.yaml

.PHONY: loadtest-start
loadtest-start:
	@kubectl scale deployment -n eventing-test loadtest-publisher --replicas 1

.PHONY: loadtest-stop
loadtest-stop:
	@kubectl scale deployment -n eventing-test loadtest-publisher --replicas 0

.PHONY: loadtest-delete
loadtest-delete:
	@ko delete -f ${DIR}/resources/loadtest
	@kubectl delete -f ${DIR}/resources/loadtest/300-configmap.yaml || true
	@kubectl delete subscriptions.eventing.kyma-project.io -n eventing-test -l 'app=loadtest'

# =============================================================================
# Publisher
# =============================================================================

.PHONY: publisher-deploy
publisher-deploy:
	@kubectl apply -f ${DIR}/resources/common/100-namespace.yaml
	@ko apply -f ${DIR}/resources/publisher

.PHONY: publisher-start
publisher-start:
	@kubectl scale deployment -n eventing-test publisher --replicas 1

.PHONY: publisher-stop
publisher-stop:
	@kubectl scale deployment -n eventing-test publisher --replicas 0

.PHONY: publisher-delete
publisher-delete:
	@ko delete -f ${DIR}/resources/publisher

# =============================================================================
# subscriber
# =============================================================================

.PHONY: subscriber-apply
subscriber-apply:
	@kubectl apply -f ${DIR}/resources/common/100-namespace.yaml -f ${DIR}/resources/subscriber/100-functions.yaml -f ${DIR}/resources/subscriber/200-service.yaml
	@ko apply -f ${DIR}/resources/subscriber/300-deployments.yaml

.PHONY: subscriber-wait
subscriber-wait:
	@echo "Waiting for Kyma functions to be ready  (if error happens, it should self-heal)"
	@kubectl wait --for=condition=ready pod -n eventing-test -l 'serverless.kyma-project.io/function-name=function-0,serverless.kyma-project.io/managed-by=function-controller,serverless.kyma-project.io/resource=deployment' --timeout=120s || sleep 20
	@kubectl wait --for=condition=ready pod -n eventing-test -l 'serverless.kyma-project.io/function-name=function-0,serverless.kyma-project.io/managed-by=function-controller,serverless.kyma-project.io/resource=deployment' --timeout=120s
	@kubectl wait --for=condition=ready pod -n eventing-test -l 'serverless.kyma-project.io/function-name=function-1,serverless.kyma-project.io/managed-by=function-controller,serverless.kyma-project.io/resource=deployment' --timeout=120s
	@kubectl wait --for=condition=ready pod -n eventing-test -l 'serverless.kyma-project.io/function-name=function-2,serverless.kyma-project.io/managed-by=function-controller,serverless.kyma-project.io/resource=deployment' --timeout=120s

.PHONY: subscriber-apply-with-default-prefix
subscriber-apply-with-default-prefix:
	@sed 's/EVENT_TYPE_PREFIX/sap.kyma.custom/g' ${DIR}/resources/subscriber/400-subscriptions.yaml | kubectl apply -f -

.PHONY: subscriber-apply-with-empty-prefix
subscriber-apply-with-empty-prefix:
	@sed 's/EVENT_TYPE_PREFIX.//g' ${DIR}/resources/subscriber/400-subscriptions.yaml | kubectl apply -f -

.PHONY: subscriber-deploy
subscriber-deploy: subscriber-apply subscriber-apply-with-default-prefix

.PHONY: subscriber-deploy-wait
subscriber-deploy-wait: subscriber-apply subscriber-wait subscriber-apply-with-default-prefix

.PHONY: subscriber-deploy-empty-prefix
subscriber-deploy-empty-prefix: subscriber-apply subscriber-apply-with-empty-prefix

.PHONY: subscriber-deploy-empty-prefix-wait
subscriber-deploy-empty-prefix-wait: subscriber-apply subscriber-wait subscriber-apply-with-empty-prefix

.PHONY: subscriber-delete
subscriber-delete:
	@ko delete -f ${DIR}/resources/subscriber
