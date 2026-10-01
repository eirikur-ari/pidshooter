# pidshooter Makefile

BINARY_NAME := pidshooter
BUILD_DIR := build
BIN_DIR := bin

LDFLAGS := -s -w

WORKSPACE_FOLDER := $(CURDIR)
LABEL_FILTER      := label=devcontainer.local_folder=$(WORKSPACE_FOLDER)
DC_SHELL          ?= zsh

## help: Show this help
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

.PHONY: all
all: clean vet lint build test coverage-html

## build: Compile the binary into bin/
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/pidshooter

## vet: Run static analyzer
.PHONY: vet
vet:
	go vet -tags "integration acceptance" ./...

## lint: Run all non-mutating source and dependency checks
.PHONY: lint
lint: fmt-lint tidy-lint

## fmt-lint: Check source code is formatted, without changing anything
.PHONY: fmt-lint
fmt-lint:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## tidy-lint: Check go.mod and go.sum are tidy, without changing anything
.PHONY: tidy-lint
tidy-lint:
	go mod tidy -diff

## fmt: Format source code
.PHONY: fmt
fmt:
	go fmt ./...

## tidy: Sync go.mod and go.sum with the code's actual imports
.PHONY: tidy
tidy:
	go mod tidy

## test: Run all tests with coverage
.PHONY: test
test: coverage

## test-unit: Run only unit test
.PHONY: test-unit
test-unit:
	go test -v -count=1 ./...

## test-integration: Run only integration tests
.PHONY: test-integration
test-integration:
	go test -v -tags integration -run TestIntegration -count=1 ./...

## test-acceptance: Run only acceptance tests
.PHONY: test-acceptance
test-acceptance:
	go test -v -tags acceptance -run TestAcceptance -count=1 ./...

## test-short: Run tests without verbose output
.PHONY: test-short
test-short:
	go test -tags integration -count=1 ./...

## test-race: Run tests with the race detector
.PHONY: test-race
test-race:
	go test -tags integration -race -count=1 ./...

## coverage: Run tests, then print a function coverage report
.PHONY: coverage
coverage:
	@mkdir -p $(BUILD_DIR)
	go test -v -tags integration -race -count=1 -coverprofile=$(BUILD_DIR)/coverage.out -covermode=atomic ./...
	go tool cover -func=$(BUILD_DIR)/coverage.out

## coverage-html: Generate an HTML report from an existing coverage profile
.PHONY: coverage-html
coverage-html:
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

## clean: Remove build artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(BUILD_DIR)

## run: Build and run (override ARGS= to customize)
ARGS ?= sleep
.PHONY: run
run: build
	./$(BIN_DIR)/$(BINARY_NAME) "$(ARGS)"

## dev-start: Start the devcontainer
.PHONY: dev-start
dev-start:
	devcontainer up --workspace-folder $(WORKSPACE_FOLDER)

## dev-rebuild: Build the devcontainer
.PHONY: dev-rebuild
dev-rebuild:
	devcontainer up --workspace-folder $(WORKSPACE_FOLDER) --remove-existing-container --build-no-cache

## dev-stop: Stop the devcontainer
.PHONY: dev-stop
dev-stop:
	@id=$$(docker ps -q --filter "$(LABEL_FILTER)"); \
	if [ -n "$$id" ]; then \
		docker stop $$id; \
	else \
		echo "No running devcontainer found."; \
	fi

## dev-shell: Open a shell in the devcontainer
.PHONY: dev-shell
dev-shell:
	devcontainer exec --workspace-folder $(WORKSPACE_FOLDER) $(DC_SHELL)

## dev-destroy: Remove the devcontainer and its image
.PHONY: dev-destroy
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
.PHONY: dev-status
dev-status:
	@docker ps -a --filter "$(LABEL_FILTER)" \
		--format 'table {{.ID}}\t{{.Status}}\t{{.Image}}'