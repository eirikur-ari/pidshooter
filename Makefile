# pidshooter Makefile

BINARY_NAME := pidshooter
BUILD_DIR := build
BIN_DIR := bin
DIST_DIR := $(BUILD_DIR)/dist

GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BINARY_VERSION = $(shell $(BIN_DIR)/$(BINARY_NAME) --version | awk '{print $$3}')
ARCHIVE_NAME = $(BINARY_NAME)_$(BINARY_VERSION)_$(GOOS)_$(GOARCH).tar.gz

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

## all: Clean, vet, lint, build, run tests with coverage, generate the HTML coverage report, run the acceptance tests, package, and write the checksums
.PHONY: all
all: clean vet lint build coverage coverage-html test-acceptance package checksums

## build: Compile the binary into bin/ (set RELEASE_VERSION= to make it report that version)
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
lint: fmt-lint tidy-lint readme-lint

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

## readme-lint: Run all README checks
.PHONY: readme-lint
readme-lint: readme-badge-go-lint readme-badge-release-lint

## readme-badge-go-lint: Check the Go version badge in README.md matches the go version in go.mod
.PHONY: readme-badge-go-lint
readme-badge-go-lint:
	@echo "Checking README Go version badge"
	@go_version=$$(awk '/^go /{print $$2}' go.mod); \
	if ! grep -q "badge/Go-$$go_version-" README.md; then \
		echo "README.md Go version badge does not match go $$go_version in go.mod"; \
		exit 1; \
	fi

## readme-badge-release-lint: Check the release badge in README.md shows RELEASE_VERSION, or, when it is unset, the latest v* tag on the current branch or the newest version in CHANGELOG.md
.PHONY: readme-badge-release-lint
readme-badge-release-lint:
	@echo "Checking README release badge"
	@if [ -n "$(RELEASE_VERSION)" ]; then \
		expected="$(RELEASE_VERSION)"; \
	else \
		expected="$$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null | sed 's/^v//') \
			$$(sed -n 's/^## \[\([0-9][^]]*\)\].*/\1/p' CHANGELOG.md | head -n 1)"; \
	fi; \
	if [ -z "$$(echo $$expected)" ]; then \
		echo "RELEASE_VERSION is not set, and neither a v* tag nor a version in CHANGELOG.md was found"; \
		exit 1; \
	fi; \
	for version in $$expected; do \
		if grep -qF "badge/release-v$$version-" README.md; then \
			exit 0; \
		fi; \
	done; \
	echo "README.md release badge does not show any of: $$expected"; \
	exit 1

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

## package: Build the binary, then archive it with the license, readme and changelog into build/dist/
.PHONY: package
package: build
	@mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(ARCHIVE_NAME) -C $(BIN_DIR) $(BINARY_NAME) -C $(WORKSPACE_FOLDER) LICENSE README.md CHANGELOG.md

## checksums: Write the SHA-256 checksums of the archives in build/dist/ to build/dist/checksums.txt
.PHONY: checksums
checksums:
	cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt

## release-notes: Write the changelog section of RELEASE_VERSION to build/release-notes.md, then print it
.PHONY: release-notes
release-notes: release-notes-file
	cat $(BUILD_DIR)/release-notes.md

## release-notes-content: Print the changelog section of RELEASE_VERSION, failing if it is missing or empty
.PHONY: release-notes-content
release-notes-content:
	@test -n "$(RELEASE_VERSION)" || { echo "RELEASE_VERSION is required"; exit 1; }
	awk -v version="$(RELEASE_VERSION)" '\
		/^## \[/ { if (found) exit; found = (index($$0, "## [" version "]") == 1); next } \
		found && /^\[[^]]+\]: / { exit } \
		found { body = body $$0 "\n" } \
		END { \
			if (body !~ /[^[:space:]]/) { print "CHANGELOG.md has no entries for " version > "/dev/stderr"; exit 1 } \
			printf "%s", body \
		}' CHANGELOG.md

## release-notes-file: Write the changelog section of RELEASE_VERSION to build/release-notes.md, failing if it is missing or empty
.PHONY: release-notes-file
release-notes-file:
	@mkdir -p $(BUILD_DIR)
	$(MAKE) --silent release-notes-content > $(BUILD_DIR)/release-notes.md || { rm -f $(BUILD_DIR)/release-notes.md; exit 1; }

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
	devcontainer up --workspace-folder $(WORKSPACE_FOLDER)

## dev-rebuild: Rebuild the devcontainer from scratch
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