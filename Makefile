GO=go
GO_MAJOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MINIMUM_SUPPORTED_GO_MAJOR_VERSION = 1
MINIMUM_SUPPORTED_GO_MINOR_VERSION = 17
MAXIMUM_SUPPORTED_GO_MINOR_VERSION = 17
GO_VERSION_VALIDATION_ERR_MSG = Golang version is not supported, please update to least $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION).$(MINIMUM_SUPPORTED_GO_MINOR_VERSION)

.DEFAULT_GOAL := install

validate-go-version: ## Validates the installed version of go against Mattermost's minimum requirement.
	@if [ $(GO_MAJOR_VERSION) -gt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		exit 0 ;\
	elif [ $(GO_MAJOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -gt $(MAXIMUM_SUPPORTED_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	fi

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
	ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress

.PHONY: generate
generate: validate-go-version
	go generate ./...

.PHONY: install
install: goreleaser
	goreleaser build --skip-validate --single-target --id datactl --rm-dist
	cp $(shell find dist -type f -name datactl | xargs) /usr/local/bin/

.PHONY: test-release
test-release: goreleaser
	goreleaser release --skip-publish --skip-announce --skip-validate --rm-dist

.PHONY: release
release: goreleaser
	goreleaser release --rm-dist

.PHONY: goreleaser
goreleaser:
	go install github.com/goreleaser/goreleaser@v1.1.0
