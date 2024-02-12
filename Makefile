FORMAT=standard-quiet
test:
	go install gotest.tools/gotestsum@latest
	gotestsum -f $(FORMAT) -- -tags=test ./...
lint:
	golangci-lint run -v
