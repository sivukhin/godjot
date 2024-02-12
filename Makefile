.PHONY: test lint

TEST_FORMAT ?= standard-quiet
LINT_FORMAT ?= colored-line-number

all: test lint

test:
	go install gotest.tools/gotestsum@latest
	gotestsum --format $(FORMAT) -- -tags=test ./...
lint:
	golangci-lint run --out-format $(LINT_FORMAT) --verbose
