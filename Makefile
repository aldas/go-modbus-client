PROJECT_NAME := go-modbus-client
PKG := "github.com/aldas/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)

.PHONY: init lint test coverage coverhtml

.DEFAULT_GOAL := check

check: staticcheck revive vet security race ## check project

init:
	# installing staticcheck
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	# installing revive
	@go install github.com/mgechev/revive@latest
	# installing gosec
	@go install github.com/securego/gosec/v2/cmd/gosec@latest

staticcheck: ## Lint the files with staticcheck
	@staticcheck ${PKG_LIST}
	@revive ${PKG_LIST}

revive: ## Lint the files with revive
	@staticcheck ${PKG_LIST}
	@revive ${PKG_LIST}

vet: ## Vet the files
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

# disable `G115 (CWE-190): integer overflow conversion int -> uint16` at the moment
security: ## Run Gosec static code security analyzer
	@gosec -quiet -exclude=G115 -exclude-dir=.cache ./...

goversion ?= "1.27"
test_version: ## Run tests inside Docker with given version (defaults to 1.27). Example: make test_version goversion=1.27
	@docker run --rm -it -v $(shell pwd):/project golang:$(goversion) /bin/sh -c "cd /project && make init check"

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

benchmark: ## Run benchmarks
	@go test -run="-" -bench=".*" ${PKG_LIST}

coverage: ## Generate global code coverage report
	./scripts/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	./scripts/coverage.sh html;

build-poller: ## build modbus-poller for current machine Arch
	@go build -o modbus-poller cmd/modbus-poller/main.go

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
