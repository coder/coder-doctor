PROJECT_ROOT := $(shell git rev-parse --show-toplevel)
GO_FILES := $(shell $(PROJECT_ROOT)/scripts/depfind/go.sh)

fmt/go:
# Skip this step in CI, golangci-lint will check formatting
ifndef CI
	@echo "--- goimports"
	git ls-files '*.go' | xargs -I % -n 16 -P 16 goimports -w -local=coder.com,cdr.dev,go.coder.com,github.com/cdr
endif
.PHONY: fmt/go

fmt: fmt/go
.PHONY: fmt

lint/go: lint/golangci-lint
.PHONY: lint/go

lint/golangci-lint:
	@echo "--- golangci-lint"
	golangci-lint run
.PHONY: lint/golangci-lint

lint/shellcheck: $(shell scripts/depfind/sh.sh)
	@echo "--- shellcheck"
	shellcheck $^
.PHONY: lint/shellcheck

lint: lint/go lint/shellcheck
.PHONY: lint

test: test/go
.PHONY: test

test/go:
	@echo "--- go test"
	$(shell scripts/test_go.sh)
.PHONY: test/go
