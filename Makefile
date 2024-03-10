PROJECT_NAME := go-icann-czds-client
GOLANGCI_LINT_VER := v1.56.2

.PHONY: go-lint
go-lint:
	docker run \
		--rm \
		--volume "$$PWD":/src/github.com/martinsirbe/$(PROJECT_NAME) \
		--workdir /src/github.com/martinsirbe/$(PROJECT_NAME) \
		golangci/golangci-lint:$(GOLANGCI_LINT_VER) \
		/bin/bash -c "golangci-lint run -v --config=/src/github.com/martinsirbe/$(PROJECT_NAME)/.golangci.yml"

.PHONY: tests
tests:
	@go test -v -cover -race -coverprofile=coverage.txt -covermode=atomic ./...
