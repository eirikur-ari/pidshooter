# pidshooter Makefile

BINARY_NAME := pidshooter
BUILD_DIR := build
BIN_DIR := bin

LDFLAGS := -s -w

WORKSPACE_FOLDER := $(CURDIR)
LABEL_FILTER      := label=devcontainer.local_folder=$(WORKSPACE_FOLDER)
DC_SHELL          ?= zsh

.PHONY: all build test test-unit test-integration test-acceptance test-short test-race coverage coverage-html vet fmt \
        clean run help dev-start dev-stop dev-shell dev-destroy dev-status dev-rebuild

all: clean test build coverage-html

## build: Compile the binary into bin/
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/pidshooter

## test: Run all tests
test:
	go test -v -tags integration -count=1 ./...

# test-unit: Run only unit test
test-unit:
	go test -v -count=1 ./...

# test-integration: Run only integration tests
test-integration:
	go test -v -tags integration -run TestIntegration -count=1 ./...

# test-acceptance: Run only acceptance tests
test-acceptance:
	go test -v -tags acceptance -run TestAcceptance -count=1 ./...

## test-short: Run tests without verbose output
test-short:
	go test -tags integration -count=1 ./...

## test-race: Run tests with the race detector
test-race:
	go test -tags integration -race -count=1 ./...

## coverage: Run tests with coverage report
coverage:
	@mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -func=$(BUILD_DIR)/coverage.out

## coverage-html: Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format source code
fmt:
	go fmt ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BIN_DIR) $(BUILD_DIR)

## run: Build and run (override ARGS= to customize)
ARGS ?= sleep
run: build
	./$(BIN_DIR)/$(BINARY_NAME) "$(ARGS)"

### dev-start: Start the devcontainer
dev-start:
	devcontainer up --workspace-folder $(WORKSPACE_FOLDER)

## dev-rebuild: Build the devcontainer
dev-rebuild:
	devcontainer up --workspace-folder $(WORKSPACE_FOLDER) --remove-existing-container --build-no-cache

## dev-stop: Stop the devcontainer
dev-stop:
	@id=$$(docker ps -q --filter "$(LABEL_FILTER)"); \
	if [ -n "$$id" ]; then \
		docker stop $$id; \
	else \
		echo "No running devcontainer found."; \
	fi

## dev-shell: Open a shell in the devcontainer
dev-shell:
	devcontainer exec --workspace-folder $(WORKSPACE_FOLDER) $(DC_SHELL)

## dev-destroy: Remove the devcontainer and its image
dev-destroy:
	@id=$$(docker ps -aq --filter "$(LABEL_FILTER)"); \
	if [ -n "$$id" ]; then \
		docker rm -f $$id; \
		echo "Removed container $$id"; \
	else \
		echo "No devcontainer found."; \
	fi
	@img=$$(docker images -q --filter "$(LABEL_FILTER)"); \
	if [ -n "$$img" ]; then \
		docker rmi -f $$img; \
		echo "Removed image $$img"; \
	fi

## dev-status: Show the status of the devcontainer
dev-status:
	@docker ps -a --filter "$(LABEL_FILTER)" \
		--format 'table {{.ID}}\t{{.Status}}\t{{.Image}}'

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
