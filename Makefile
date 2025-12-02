# Project variables
PROJECT_NAME := dra-deployer
BINARY_NAME := dra-deployer
BUILD_DIR := bin
CMD_DIR := cmd/dra-deployer
GO := go

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: all
all: build

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GO) mod tidy
	$(GO) mod vendor
	@echo "Dependencies updated"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

.PHONY: check
check: fmt vet test-unit ## Run fmt, vet, and test

.PHONY: release
release: clean test-unit build ## Clean, test, and build for release

.PHONY: version
version: ## Display version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(BUILD_DATE)"

.PHONY: deploy-dra-driver-memory
deploy-dra-driver-memory: build
	hack/deploy-dra-driver-memory.sh

.PHONY: setup-e2e-cluster
setup-e2e-cluster: ## Setup Minikube cluster with CRI-O and NRI for e2e tests
	@echo "Setting up e2e cluster..."
	hack/ci/setup-cluster.sh

.PHONY: test-e2e
test-e2e: build ## Run end-to-end tests
	@echo "Running e2e tests..."
	$(GO) run github.com/onsi/ginkgo/v2/ginkgo -v -timeout=10m ./test/e2e/...