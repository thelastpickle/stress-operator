ORG=thelastpickle
PROJECT=stress-operator
REG=docker.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/thelastpickle/stress-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
REV=$(shell git rev-parse --short=12 HEAD)

IMAGE_BASE=$(REG)/$(ORG)/$(PROJECT)
PRE_TEST_TAG=$(BRANCH)-$(REV)-TEST
POST_TEST_TAG=$(BRANCH)-$(REV)
E2E_IMAGE=$(IMAGE_BASE):$(PRE_TEST_TAG)
BRANCH_REV_IMAGE=$(IMAGE_BASE):$(POST_TEST_TAG)
REV_IMAGE=$(IMAGE_BASE):$(REV)
BRANCH_LATEST_IMAGE=$(IMAGE_BASE):$(BRANCH)-latest
LATEST_IMAGE=$(IMAGE_BASE):latest

ifeq ($(CIRCLE_BRANCH),master)
	PUSH_LATEST := true
endif

DEV_NS ?= tlpstress
E2E_NS?=tlpstress-e2e

TEMPLATE_DIR := $(shell echo "`pwd`/templates")

.PHONY: clean
clean:
	rm -rf build/_output

.PHONY: run
run:
	@operator-sdk up local --namespace=${DEV_NS}

.PHONY: build
build:
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o=$(COMPILE_TARGET) ./cmd/manager

.PHONY: code-gen
code-gen:
	operator-sdk generate k8s

.PHONY: openapi-gen
openapi-gen:
	operator-sdk generate openapi

.PHONY: build-e2e-image
build-e2e-image:
	@echo Building ${E2E_IMAGE}
	@operator-sdk build ${E2E_IMAGE}

.PHONY: push-e2e-image
push-e2e-image:
	docker push ${E2E_IMAGE}

.PHONY: build-image
build-image: code-gen openapi-gen
	@operator-sdk build ${REG}/${ORG}/${PROJECT}:${POST_TEST_TAG}

.PHONY: push-image
push-image:
	@echo Pushing ${BRANCH_REV_IMAGE}
	docker tag ${E2E_IMAGE} ${BRANCH_REV_IMAGE}
	docker push ${BRANCH_REV_IMAGE}
	@echo Pushing ${REV_IMAGE}
	docker tag ${BRANCH_REV_IMAGE} ${REV_IMAGE}
	docker push ${REV_IMAGE}
ifdef CIRCLE_BRANCH
	@echo Pushing ${BRANCH_LATEST_IMAGE}
	docker tag ${BRANCH_REV_IMAGE} ${BRANCH_LATEST_IMAGE}
	docker push ${BRANCH_LATEST_IMAGE}
endif
ifdef PUSH_LATEST
	@echo PUSHING ${LATEST_IMAGE}
	docker tag ${BRANCH_REV_IMAGE} ${LATEST_IMAGE}
	docker push ${IMAGE_BASE}:latest
endif

.PHONY: unit-test
unit-test: export TEMPLATE_PATH=$(TEMPLATE_DIR)
unit-test:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONY: do-deploy-casskop
do-deploy-casskop:
	kubectl -n $(CASSKOP_NS) apply -f config/casskop/casskop.yaml

.PHONY: deploy-casskop
deploy-casskop: CASSKOP_NS ?= $(DEV_NS)
deploy-casskop: do-deploy-casskop

.PHONY: do-deploy-grafana-operator
do-deploy-grafana-operator:
	@echo GRAFANA_NS = $(GRAFANA_NS)
	kubectl -n $(GRAFANA_NS) apply -f config/grafana/grafana-operator.yaml

.PHONY: deploy-grafana-operator
deploy-grafana-operator: GRAFANA_NS ?= $(DEV_NS)
deploy-grafana-operator: do-deploy-grafana-operator

.PHONY: do-deploy-grafana
do-deploy-grafana: deploy-grafana-operator
	until kubectl get crd grafanas.integreatly.org > /dev/null 2>&1; do \
    	echo "Waiting for grafanas.integreatly.org CRD to be deployed"; \
    	sleep 1; \
    done;
	echo "Installing Grafana"
    # Temporarily adding a sleep call here due to intermittent failures on CircleCI.
	sleep 2
	kubectl -n $(GRAFANA_NS) apply -f config/grafana/02_grafana.yaml
	kubectl -n $(GRAFANA_NS) apply -f config/grafana/prometheus-datasource.yaml

.PHONY: deploy-grafana
deploy-grafana: GRAFANA_NS ?= $(DEV_NS)
deploy-grafana: do-deploy-grafana

.PHONY: deploy-prometheus-operator
deploy-prometheus-operator:
	kubectl apply -f config/prometheus-operator/prometheus-operator.yaml

.PHONY: do-deploy-prometheus
do-deploy-prometheus: deploy-prometheus-operator
	until kubectl get crd prometheuses.monitoring.coreos.com > /dev/null 2>&1; do \
		echo "Waiting for prometheuses.monitoring.coreos.com CRD to be deployed"; \
		sleep 1; \
	done;
	echo "Installing Prometheus"
	# Temporarily adding a sleep call here due to intermittent failures on CircleCI.
	sleep 2
	kubectl -n $(PROMETHEUS_NS) apply -f config/prometheus/prometheus.yaml

.PHONY: deploy-prometheus
deploy-prometheus: PROMETHEUS_NS ?= $(DEV_NS)
deploy-prometheus: do-deploy-prometheus

.PHONY: create-e2e-ns
create-e2e-ns:
	./scripts/create-ns.sh $(E2E_NS)

.PHONY: e2e-setup
e2e-setup: PROMETHEUS_NS = $(E2E_NS)
e2e-setup: CASSKOP_NS = $(E2E_NS)
e2e-setup: GRAFANA_NS = $(E2E_NS)
e2e-setup: create-e2e-ns deploy-prometheus-operator do-deploy-prometheus do-deploy-casskop do-deploy-grafana-operator
	kubectl apply -n $(E2E_NS) -f config/casskop

.PHONY: e2e-test
e2e-test: e2e-setup
	@echo Running e2e tests
	operator-sdk test local ./test/e2e --image $(E2E_IMAGE) --namespace $(E2E_NS)

.PHONY: e2e-cleanup
e2e-cleanup:
	kubectl -n $(E2E_NS) delete cassandracluster --all
	kubectl -n $(E2E_NS) delete sa tstress-operator
	kubectl -n $(E2E_NS) delete role tstress-operator
	kubectl -n $(E2E_NS) delete rolebinding tstress-operator
	kubectl -n $(E2E_NS) delete deployment tstress-operator

.PHONY: create-dev-ns
create-dev-ns:
	./scripts/create-ns.sh $(DEV_NS)

.PHONY: deploy
deploy: create-dev-ns
	kubectl -n $(DEV_NS) apply -f deploy/crds/
	kubectl -n $(DEV_NS) apply -f deploy/service_account.yaml
	kubectl -n $(DEV_NS) apply -f deploy/role.yaml
	kubectl -n $(DEV_NS) apply -f deploy/role_binding.yaml
	kubectl -n $(DEV_NS) apply -f deploy/operator.yaml

.PHONY: deploy-all
deploy-all: CASSKOP_NS = $(DEV_NS)
deploy-all: PROMETHEUS_NS = $(DEV_NS)
deploy-all: GRAFANA_NS = $(DEV_NS)
deploy-all: create-dev-ns do-deploy-casskop deploy-prometheus deploy-grafana deploy

.PHONY: init-kind-kubeconfig
init-kind-kubeconfig:
	export KUBECONFIG=$(shell kind get kubeconfig-path --name="tlpstress")

.PHONY: create-kind-cluster
create-kind-cluster:
	kind create -f config/kind/multi-node.yaml
	export KUBECONFIG="$(kind get kubeconfig-path --name="tlpstress")"
	kubectl apply -f config/kind/hostpath-provisioner.yaml
	kubectl delete storageclass default