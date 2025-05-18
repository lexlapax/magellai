# ABOUTME: Makefile for building, testing, and managing the Magellai project
# ABOUTME: Provides commands for building, testing, linting, and documentation
.PHONY: all build test test-integration test-all test-race test-coverage test-sqlite clean clean-cache clean-testcache clean-modcache clean-all install fmt lint vet help docker-build docker-test release-build release docs

# Build variables
BINARY_NAME=magellai
BUILD_DIR=bin
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_VET=$(GO_CMD) vet
GO_FMT=$(GO_CMD) fmt
GO_INSTALL=$(GO_CMD) install
GO_CLEAN=$(GO_CMD) clean
GO_MOD=$(GO_CMD) mod
GOPATH=$(shell go env GOPATH)

# Version information
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
all: build

## build: Build the binary (default with no database support)
build:
	@echo "Building $(BINARY_NAME) (default, no database support)..."
	$(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/magellai/...

## build-race: Build with race detection
build-race:
	@echo "Building with race detection..."
	$(GO_BUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-race ./cmd/magellai/...

## build-sqlite: Build with SQLite database support
build-sqlite:
	@echo "Building with SQLite database support..."
	$(GO_BUILD) -tags="sqlite db" $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-sqlite ./cmd/magellai/...

## build-db: Build with all database support (SQLite and PostgreSQL)
build-db:
	@echo "Building with all database support..."
	$(GO_BUILD) -tags="sqlite postgresql db" $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-db ./cmd/magellai/...

## install: Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GO_INSTALL) $(LDFLAGS) ./cmd/magellai

## test: Run unit tests only
test:
	@echo "Running unit tests..."
	$(GO_TEST) -short ./...

## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	$(GO_TEST) -tags=integration ./...

## test-all: Run all tests (unit and integration)
test-all:
	@echo "Running all tests..."
	$(GO_TEST) -tags=integration ./...

## test-race: Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GO_TEST) -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO_TEST) -cover ./...

## test-coverage-html: Generate HTML coverage report
test-coverage-html:
	@echo "Generating coverage report..."
	$(GO_TEST) -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

## test-sqlite: Run tests with SQLite support
test-sqlite:
	@echo "Running tests with SQLite support..."
	$(GO_TEST) -tags="sqlite" ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO_VET) ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -rf $(BUILD_DIR)/* coverage.out coverage.html

## clean-cache: Clean Go build cache
clean-cache:
	@echo "Cleaning Go build cache..."
	$(GO_CLEAN) -cache

## clean-testcache: Clean Go test cache
clean-testcache:
	@echo "Cleaning Go test cache..."
	$(GO_CLEAN) -testcache

## clean-modcache: Clean Go module cache
clean-modcache:
	@echo "Cleaning Go module cache..."
	$(GO_CLEAN) -modcache

## clean-all: Clean everything (artifacts, build cache, test cache, module cache)
clean-all: clean clean-cache clean-testcache clean-modcache
	@echo "All caches and artifacts cleaned"

## deps: Manage dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO_MOD) download

## deps-tidy: Clean up dependencies
deps-tidy:
	@echo "Tidying dependencies..."
	$(GO_MOD) tidy

## deps-update: Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GO_CMD) get -u ./...
	$(GO_MOD) tidy

## docs: Generate documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server on http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Run: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

## bench: Run all benchmarks
bench:
	@echo "Running all benchmarks..."
	$(GO_TEST) -bench=. -benchmem ./...

## bench-storage: Run storage backend benchmarks
bench-storage:
	@echo "Running storage backend benchmarks..."
	$(GO_TEST) -bench=. -benchmem ./pkg/repl -run=^$

## bench-storage-db: Run storage backend benchmarks with database support
bench-storage-db:
	@echo "Running storage backend benchmarks with database support..."
	$(GO_TEST) -tags="sqlite db" -bench=. -benchmem ./pkg/repl -run=^$

## bench-compare: Compare storage backends performance
bench-compare:
	@echo "Comparing storage backend performance..."
	@echo "FileSystem backend:"
	$(GO_TEST) -bench=BenchmarkSessionSave -benchmem ./pkg/repl -run=^$ | grep -E "BenchmarkSessionSave|ns/op|B/op|allocs/op"
	@echo ""
	@echo "SQLite backend:"
	$(GO_TEST) -tags="sqlite db" -bench=BenchmarkSessionSave -benchmem ./pkg/repl -run=^$ | grep -E "BenchmarkSessionSave|ns/op|B/op|allocs/op"

## release-build: Build releases for multiple platforms
release-build:
	@echo "Building releases..."
	@mkdir -p $(BUILD_DIR)/release
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 cmd/magellai/main.go
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 cmd/magellai/main.go
	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 cmd/magellai/main.go
	# Darwin ARM64
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 cmd/magellai/main.go
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe cmd/magellai/main.go

## tools: Install development tools
tools:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing godoc..."
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest

## pre-commit: Run pre-commit checks (uses unit tests only for speed)
pre-commit: fmt vet lint test

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

## docker-test: Run tests in Docker
docker-test:
	@echo "Running tests in Docker..."
	docker run --rm $(BINARY_NAME):$(VERSION) make test-all

## help: Show this help
help:
	@echo "Available targets:"
	@grep -h -E '^##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Catch-all target
%:
	@echo "Unknown target '$@'"
	@echo "Run 'make help' for available targets"
	@exit 1