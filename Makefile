GO=go

.DEFAULT_GOAL := install

GOBIN := $(shell pwd)/bin
PATH := $(GOBIN):$(PATH)

IMAGE_REGISTRY ?= localhost
IMAGE_NAME ?= datactl
IMAGE_TAG ?= latest

export PATH
export GOBIN

.PHONY: tag
tag:
	git tag $(svu next)

.PHONY: licenses
licenses:
	find . -type f -name "*.go" | xargs addlicense -c "IBM Corporation."

.PHONY: mod
mod:
	go mod tidy
	go mod download

.PHONY: test
test:
	ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --race --show-node-events

.PHONY: generate
generate: tools
	go generate ./...

.PHONY: install
install: goreleaser
	goreleaser build --skip-validate --single-target --id datactl --clean
	cp $(shell find dist -type f -name datactl | xargs) /usr/local/bin/

.PHONY: test-release
test-release: goreleaser
	goreleaser release --skip-publish --skip-announce --skip-validate --clean

.PHONY: release
release: goreleaser
	goreleaser release --clean

.PHONY: goreleaser
goreleaser:
	go install github.com/goreleaser/goreleaser/v2@v2.16.0

tools:
	go mod download
	go install "k8s.io/code-generator/cmd/conversion-gen@v0.27.7"	
	go install "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.12.1"

go-licenses:
	go install "github.com/google/go-licenses@latest"

licenses-check: go-licenses
	go-licenses check --include_tests ./...

licenses-save: go-licenses
	go-licenses save --include_tests ./... --save_path=licenses

docker-build:
	docker build -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)