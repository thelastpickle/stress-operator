ORG=jsanda
NAMESPACE=tlp-stress
PROJECT=tlp-stress-operator
REG=docker.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/jsanda/tlp-stress-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

.PHONY: code/run
code/run:
	@operator-sdk up local --namespace=${NAMESPACE}

.PHONY: code/compile
code/compile:
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o=$(COMPILE_TARGET) ./cmd/manager

.PHONY: code/gen
code/gen:
	operator-sdk generate k8s

.PHONY: image/build
image/build: code/compile
	@operator-sdk build ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: image/push
image/push:
	docker push ${REG}/${ORG}/${PROJECT}:${TAG}