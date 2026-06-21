# pidshooter Makefile

BINARY_NAME := pidshooter
BUILD_DIR := bin

LDFLAGS := -s -w

.PHONY: all build test test-short test-race coverage vet fmt clean run help

all: build

## build: Compile the binary into bin/
build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/pidshooter

## test: Run all tests
test:
	go test -v -count=1 ./...

## test-short: Run tests without verbose output
test-short:
	go test -count=1 ./...

## test-race: Run tests with the race detector
test-race:
	go test -race -count=1 ./...

## coverage: Run tests with coverage report
coverage:
	@mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -func=$(BUILD_DIR)/coverage.out

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format source code
fmt:
	go fmt ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## run: Build and run (override ARGS= to customize)
ARGS ?= sleep
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) "$(ARGS)"

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
