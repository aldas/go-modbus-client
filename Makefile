PROJECT_NAME := go-modbus-client
PKG := "github.com/aldas/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)

.PHONY: init lint test coverage coverhtml

.DEFAULT_GOAL := check

check: lint vet race ## check project

init:
	git config core.hooksPath ./scripts/.githooks
	@go get -u golang.org/x/lint/golint

lint: ## Lint the files
	@golint -set_exit_status ${PKG_LIST}

vet: ## Vet the files
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

goversion ?= "1.16"
test_version: ## Run tests inside Docker with given version (defaults to 1.16 oldest supported). Example: make test_version goversion=1.16
	@docker run --rm -it -v $(shell pwd):/project golang:$(goversion) /bin/sh -c "cd /project && make init check"

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

benchmark: ## Run benchmarks
	@go test -run="-" -bench=".*" ${PKG_LIST}

coverage: ## Generate global code coverage report
	./scripts/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	./scripts/coverage.sh html;


help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
