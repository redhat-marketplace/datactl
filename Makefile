.DEFAULT_GOAL := install

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
generate:
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
