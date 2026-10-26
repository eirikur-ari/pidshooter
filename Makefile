# pidshooter Makefile

BINARY_NAME := pidshooter
BUILD_DIR := build
BIN_DIR := bin
DIST_DIR := $(BUILD_DIR)/dist
PACKAGE_DIR := $(BUILD_DIR)/package

LDFLAGS := -s -w

WORKSPACE_FOLDER := $(CURDIR)
DEVCONTAINER      := WORKSPACE_FOLDER=$(WORKSPACE_FOLDER) bash scripts/devcontainer.sh

## help: Show this help
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

## all: Run every check and produce the release archives from a clean state
.PHONY: all
all: clean vet lint build coverage coverage-html test-acceptance licenses package checksums

## build: Compile the binary, optionally stamped with a release version
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	VERSION_FLAG="-X github.com/eirikur-ari/pidshooter/internal/composition.releaseVersion=$(RELEASE_VERSION)"; \
	go build -ldflags "$(LDFLAGS) $$VERSION_FLAG" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/pidshooter

## vet: Run go vet on all packages, including the integration and acceptance tests
.PHONY: vet
vet:
	go vet -tags "integration acceptance" ./...

## lint: Run all non-mutating source and dependency checks
.PHONY: lint
lint: fmt-lint tidy-lint shell-lint readme-lint

## fmt-lint: Check source code is formatted, without changing anything
.PHONY: fmt-lint
fmt-lint:
	@echo "Checking source formatting"
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## vuln-lint: Check dependencies and the Go toolchain for known vulnerabilities
.PHONY: vuln-lint
vuln-lint:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -show verbose ./...

## tidy-lint: Check go.mod and go.sum are tidy, without changing anything
.PHONY: tidy-lint
tidy-lint:
	@echo "Checking module tidiness"
	@go mod tidy -diff

## shell-lint: Run shellcheck on the shell scripts
.PHONY: shell-lint
shell-lint:
	find scripts .github/scripts .devcontainer -name '*.sh' ! -path .devcontainer/files/claude/claude-install.sh -print0 | xargs -0 shellcheck

## readme-lint: Run all README checks
.PHONY: readme-lint
readme-lint: readme-badge-go-lint readme-badge-release-lint

## readme-badge-go-lint: Check the Go version badge in README.md matches the go version in go.mod
.PHONY: readme-badge-go-lint
readme-badge-go-lint:
	bash scripts/readme-lint.sh go-badge

## readme-badge-release-lint: Check the release badge in README.md shows the current release version
.PHONY: readme-badge-release-lint
readme-badge-release-lint:
	RELEASE_VERSION=$(RELEASE_VERSION) bash scripts/readme-lint.sh release-badge

## fmt: Format source code
.PHONY: fmt
fmt:
	go fmt ./...

## tidy: Sync go.mod and go.sum with the code's actual imports
.PHONY: tidy
tidy:
	go mod tidy

## test: Run unit and integration tests with the race detector
.PHONY: test
test: test-race

## test-unit: Run only unit tests, verbosely
.PHONY: test-unit
test-unit:
	go test -v -count=1 ./...

## test-integration: Run only integration tests
.PHONY: test-integration
test-integration:
	go test -v -tags integration -run TestIntegration -count=1 ./...

## test-acceptance: Build the binary, then run only acceptance tests against it
.PHONY: test-acceptance
test-acceptance: build
	PIDSHOOTER_BINARY=$(WORKSPACE_FOLDER)/$(BIN_DIR)/$(BINARY_NAME) go test -v -tags acceptance -run TestAcceptance -count=1 ./...

## test-race: Run unit and integration tests with the race detector
.PHONY: test-race
test-race:
	go test -tags integration -race -count=1 ./...

## coverage: Run unit and integration tests with the race detector, then print a function coverage report
.PHONY: coverage
coverage:
	@mkdir -p $(BUILD_DIR)
	go test -v -tags integration -race -count=1 -coverprofile=$(BUILD_DIR)/coverage.out -covermode=atomic ./...
	go tool cover -func=$(BUILD_DIR)/coverage.out

## coverage-summary: Print the total coverage percentage of an existing coverage profile
.PHONY: coverage-summary
coverage-summary:
	@go tool cover -func=$(BUILD_DIR)/coverage.out | awk '/^total:/ {print $$3}'

## coverage-html: Generate an HTML report from an existing coverage profile
.PHONY: coverage-html
coverage-html:
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

## licenses: Collect the licenses of the third-party software compiled into the binary
.PHONY: licenses
licenses:
	WORKSPACE_FOLDER=$(WORKSPACE_FOLDER) PACKAGE_DIR=$(PACKAGE_DIR) bash scripts/third-party-licenses.sh

## package: Build the binary and archive it with its documentation and licenses
.PHONY: package
package: licenses build
	WORKSPACE_FOLDER=$(WORKSPACE_FOLDER) BIN_DIR=$(BIN_DIR) BINARY_NAME=$(BINARY_NAME) PACKAGE_DIR=$(PACKAGE_DIR) DIST_DIR=$(DIST_DIR) bash scripts/package.sh

## checksums: Write the SHA-256 checksums of the release archives
.PHONY: checksums
checksums:
	cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt

## release-notes: Write the release notes of a version from the changelog to a file, then print them
.PHONY: release-notes
release-notes: release-notes-file
	cat $(BUILD_DIR)/release-notes.md

## release-notes-content: Print the release notes of a version from the changelog, failing if there are none
.PHONY: release-notes-content
release-notes-content:
	RELEASE_VERSION=$(RELEASE_VERSION) bash scripts/release-notes.sh

## release-notes-file: Write the release notes of a version from the changelog to a file, failing if there are none
.PHONY: release-notes-file
release-notes-file:
	@mkdir -p $(BUILD_DIR)
	RELEASE_VERSION=$(RELEASE_VERSION) bash scripts/release-notes.sh > $(BUILD_DIR)/release-notes.md || { rm -f $(BUILD_DIR)/release-notes.md; exit 1; }

## clean: Remove build artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(BUILD_DIR)

## run: Build and run (override ARGS= to customize)
ARGS ?= play sleep
.PHONY: run
run: build
	./$(BIN_DIR)/$(BINARY_NAME) $(ARGS)

## dev-start: Start the devcontainer
.PHONY: dev-start
dev-start:
	$(DEVCONTAINER) start

## dev-rebuild: Rebuild the devcontainer from scratch
.PHONY: dev-rebuild
dev-rebuild:
	$(DEVCONTAINER) rebuild

## dev-stop: Stop the devcontainer
.PHONY: dev-stop
dev-stop:
	@$(DEVCONTAINER) stop

## dev-shell: Open a shell in the devcontainer
.PHONY: dev-shell
dev-shell:
	$(DEVCONTAINER) shell

## dev-copy: Copy a file or directory into the devcontainer
.PHONY: dev-copy
dev-copy:
	$(DEVCONTAINER) copy "$(SOURCE)" $(if $(DESTINATION),"$(DESTINATION)")

## dev-destroy: Remove the devcontainer and its image
.PHONY: dev-destroy
dev-destroy:
	@$(DEVCONTAINER) destroy

## dev-status: Show the status of the devcontainer
.PHONY: dev-status
dev-status:
	@$(DEVCONTAINER) status
