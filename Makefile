.PHONY: check fmt install-hooks lint test test-race vet

check: fmt lint vet test

install-hooks:
	git config core.hooksPath .githooks

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

lint:
	golangci-lint run ./...

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...
