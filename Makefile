# Variables
BINARY_NAME=lazyoc
VERSION?=$(shell cat VERSION)
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DIR=bin
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/lazyoc

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "Installing ${BINARY_NAME}..."
	@go install ${LDFLAGS} ./cmd/lazyoc

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}
	@rm -f ${BINARY_NAME}

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy and verify modules
	@echo "Tidying modules..."
	@go mod tidy
	@go mod verify

.PHONY: vendor
vendor: ## Create vendor directory
	@echo "Creating vendor directory..."
	@go mod vendor

.PHONY: run
run: build ## Build and run the application
	@echo "Running ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

.PHONY: dev
dev: ## Run in development mode
	@echo "Running in development mode..."
	@go run ./cmd/lazyoc

.PHONY: version
version: ## Show version
	@echo ${VERSION}

.PHONY: all
all: clean fmt vet test build ## Run all checks and build