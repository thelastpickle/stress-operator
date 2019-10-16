ORG=jsanda
NAMESPACE=tlpstress-dev
PROJECT=tlp-stress-operator
REG=docker.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/jsanda/tlp-stress-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

# This currently has to be set to tlpstress-e2e. The manifest files in test/manifests
# declare the namespace to be tlpstresss-e2e
E2E_NAMESPACE=tlpstress-e2e

PROMETHEUS_OPERATOR_NAMESPACE=prometheus-operator

.PHONY: clean
clean:
	rm -rf build/_output

.PHONY: run
run:
	@operator-sdk up local --namespace=${NAMESPACE}

.PHONY: build
build:
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o=$(COMPILE_TARGET) ./cmd/manager

.PHONY: code-gen
code-gen:
	operator-sdk generate k8s

.PHONE: openapi-gen
openapi-gen:
	operator-sdk generate openapi

.PHONY: build-image
build-image: code-gen openapi-gen
	@operator-sdk build ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: push-image
push-image:
	docker push ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: unit-test
unit-test:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONE: deploy-casskop
deploy-casskop:
	kubectl apply -f config/casskop

.PHONY: deploy-prometheus-operator
deploy-prometheus-operator:
	kubectl apply -f config/prometheus-operator/bundle.yaml

.PHONY: deploy-prometheus
deploy-prometheus: deploy-prometheus-operator
	until kubectl get crd prometheuses.monitoring.coreos.com > /dev/null 2>&1; do \
		echo "Waiting for prometheuses.monitoring.coreos.com CRD to be deployed"; \
		sleep 1; \
	done;
	kubectl apply -f config/prometheus/bundle.yaml

.PHONY: e2e-setup
e2e-setup:
	./scripts/create-ns.sh $(E2E_NAMESPACE)
	./scripts/create-ns.sh $(PROMETHEUS_OPERATOR_NAMESPACE)
	kubectl apply -n $(E2E_NAMESPACE) -f config/casskop
	kubectl apply -n $(PROMETHEUS_OPERATOR_NAMESPACE) -f config/prometheus-operator/bundle.yaml

.PHONY: e2e-test
e2e-test: e2e-setup
	@echo Running e2e tests
	operator-sdk test local ./test/e2e --namespace $(E2E_NAMESPACE)

.PHONY: e2e-cleanup
e2e-cleanup:
#	kubectl -n $(E2E_NAMESPACE) delete tlpstress --all
	kubectl -n $(E2E_NAMESPACE) delete cassandracluster --all
	kubectl -n $(E2E_NAMESPACE) delete sa tlp-stress-operator
	kubectl -n $(E2E_NAMESPACE) delete role tlp-stress-operator
	kubectl -n $(E2E_NAMESPACE) delete rolebinding tlp-stress-operator
	kubectl -n $(E2E_NAMESPACE) delete deployment tlp-stress-operator

.PHONY: init-kind-kubeconfig
init-kind-kubeconfig:
	export KUBECONFIG=$(shell kind get kubeconfig-path --name="tlpstress")

.PHONY: create-kind-cluster
create-kind-cluster:
	kind create -f config/kind/multi-node.yaml
	export KUBECONFIG="$(kind get kubeconfig-path --name="tlpstress")"
	kubectl apply -f config/kind/hostpath-provisioner.yaml
	kubectl delete storageclass default