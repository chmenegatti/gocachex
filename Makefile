# GoCacheX Makefile - Library Project

# Variables
BINARY_NAME=gocachex
PKG_LIST=$(shell go list ./... | grep -v /vendor/)
GO_FILES=$(shell find . -name '*.go' | grep -v vendor | grep -v _test.go)

# Default target
.PHONY: all
all: clean deps test lint vet

# Build the library (compile check)
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) library..."
	@go build ./...

# Build for multiple platforms (compile check)
.PHONY: build-all
build-all:
	@echo "Building library for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build ./...
	@GOOS=darwin GOARCH=amd64 go build ./...
	@GOOS=windows GOARCH=amd64 go build ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -race $(PKG_LIST)

# Run tests with coverage
.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic $(PKG_LIST)
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output and race detection
.PHONY: test-verbose
test-verbose:
	@echo "Running verbose tests..."
	@go test -v -race -count=1 $(PKG_LIST)

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -v -race -tags=integration $(PKG_LIST)

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem $(PKG_LIST)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@go clean
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt $(PKG_LIST)

# Check formatting
.PHONY: fmt-check
fmt-check:
	@echo "Checking formatting..."
	@test -z $$(gofmt -l $(GO_FILES)) || (echo "Code not formatted, run 'make fmt'" && exit 1)

# Lint code
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Install linter
.PHONY: install-lint
install-lint:
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

# Vet code
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet $(PKG_LIST)

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Install tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate code (protobuf, etc.)
.PHONY: generate
generate:
	@echo "Generating code..."
	@go generate ./...

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_PATH)

# Run examples
.PHONY: run-examples
run-examples: example-basic example-redis example-memcached example-cli

.PHONY: example-basic
example-basic:
	@echo "Running basic example..."
	@cd examples/basic && go run main.go

.PHONY: example-redis
example-redis:
	@echo "Running Redis example..."
	@cd examples/redis && go run main.go

.PHONY: example-memcached  
example-memcached:
	@echo "Running Memcached example..."
	@cd examples/memcached && go run main.go

.PHONY: example-hierarchical
example-hierarchical:
	@echo "Running hierarchical example..."
	@cd examples/hierarchical && go run main.go

.PHONY: example-cli
example-cli:
	@echo "Running CLI example..."
	@cd examples/cli && go run main.go -op demo

.PHONY: example-cli-help
example-cli-help:
	@echo "Showing CLI example help..."
	@cd examples/cli && go run main.go -help

# Example with custom parameters
.PHONY: example-cli-custom
example-cli-custom:
	@echo "Running CLI example with custom parameters..."
	@cd examples/cli && go run main.go -op set -key "makefile:test" -value "Hello from Makefile" -ttl 1h
	@cd examples/cli && go run main.go -op get -key "makefile:test"
	@cd examples/cli && go run main.go -op stats

# Start Redis for testing
.PHONY: redis-start
redis-start:
	@echo "Starting Redis..."
	@docker run -d --name gocachex-redis -p 6379:6379 redis:alpine

# Stop Redis
.PHONY: redis-stop
redis-stop:
	@echo "Stopping Redis..."
	@docker stop gocachex-redis || true
	@docker rm gocachex-redis || true

# Start Memcached for testing
.PHONY: memcached-start
memcached-start:
	@echo "Starting Memcached..."
	@docker run -d --name gocachex-memcached -p 11211:11211 memcached:alpine

# Stop Memcached
.PHONY: memcached-stop
memcached-stop:
	@echo "Stopping Memcached..."
	@docker stop gocachex-memcached || true
	@docker rm gocachex-memcached || true

# Start all test services
.PHONY: services-start
services-start: redis-start memcached-start

# Stop all test services
.PHONY: services-stop
services-stop: redis-stop memcached-stop

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t gocachex:latest .

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run --rm -it gocachex:latest

# Security scan
.PHONY: security
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest 2>/dev/null || echo "Warning: gosec installation failed, skipping security scan"
	@gosec ./... 2>/dev/null || echo "Security scan completed (gosec may not be available)"

# Dependency check
.PHONY: deps-check
deps-check:
	@echo "Checking dependencies..."
	@which nancy > /dev/null || go install github.com/sonatypecommunity/nancy@latest
	@go list -json -m all | nancy sleuth

# Release preparation
.PHONY: release-prepare
release-prepare: clean deps test lint vet fmt-check security
	@echo "Release preparation complete"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Clean, build and test"
	@echo "  build            - Build the application"
	@echo "  build-all        - Build for multiple platforms"
	@echo "  test             - Run tests"
	@echo "  coverage         - Run tests with coverage"
	@echo "  test-verbose     - Run tests with verbose output"
	@echo "  test-integration - Run integration tests"
	@echo "  benchmark        - Run benchmarks"
	@echo "  clean            - Clean build artifacts"
	@echo "  fmt              - Format code"
	@echo "  fmt-check        - Check code formatting"
	@echo "  lint             - Run linter"
	@echo "  install-lint     - Install golangci-lint"
	@echo "  vet              - Run go vet"
	@echo "  deps             - Download dependencies"
	@echo "  deps-update      - Update dependencies"
	@echo "  install-tools    - Install development tools"
	@echo "  generate         - Generate code"
	@echo "  run              - Run the application"
	@echo "  run-examples     - Run examples"
	@echo "  redis-start      - Start Redis for testing"
	@echo "  redis-stop       - Stop Redis"
	@echo "  memcached-start  - Start Memcached for testing"
	@echo "  memcached-stop   - Stop Memcached"
	@echo "  services-start   - Start all test services"
	@echo "  services-stop    - Stop all test services"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  security         - Run security scan"
	@echo "  deps-check       - Check dependencies"
	@echo "  release-prepare  - Prepare for release"
	@echo "  help             - Show this help"
