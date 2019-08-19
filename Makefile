ORG=jsanda
NAMESPACE=tlp-stress
PROJECT=tlp-stress-operator
REG=docker.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/jsanda/tlp-stress-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

.PHONY: run
code/run:
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
build-image:
	@operator-sdk build ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: push-image
push-image:
	docker push ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: unit-test
unit-test:
	@echo Running tests:
	go test -v -race -cover ./pkg/...