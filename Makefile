.PHONY: build test swag lint

## build: compile all packages
build:
	go build ./...

## test: run all tests with the race detector
test:
	go test -race ./... -count=1

## swag: generate Swagger documentation from source annotations
swag:
	swag init --parseDependency --parseInternal -g ./cmd/api/main.go --output ./docs --outputTypes json,yaml

## lint: run golangci-lint on the entire module
lint:
	golangci-lint run ./...
