# EdgeX MessageBus Client Makefile

.PHONY: help build test clean lint fmt vet deps examples install-tools

# Default target
help: ## Show this help message
	@echo "EdgeX MessageBus Client - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build all examples
	@echo "Building examples..."
	@go build -o bin/basic-example ./example/main.go
	@go build -o bin/advanced-example ./example/advanced/main.go
	@echo "✅ Build completed"

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@go test -v -race ./...
	@echo "✅ Race tests completed"

# Code quality targets
lint: install-tools ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "✅ Linting completed"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ Vet completed"

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated"

# Example targets
examples: build ## Build and run examples
	@echo "Running basic example..."
	@./bin/basic-example || echo "Basic example requires MessageBus server"
	@echo ""
	@echo "Running advanced example..."
	@./bin/advanced-example || echo "Advanced example requires MessageBus server"

# Development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all . > docs.txt
	@echo "✅ Documentation generated: docs.txt"

# Clean targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html docs.txt
	@echo "✅ Clean completed"

# Release targets
tag: ## Create a new git tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag VERSION=v1.0.0"; exit 1; fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "✅ Tag $(VERSION) created and pushed"

# Docker targets (optional)
docker-build: ## Build Docker image for examples
	@echo "Building Docker image..."
	@docker build -t edgex-messagebus-client-example .
	@echo "✅ Docker image built"

# Integration test targets
test-integration: ## Run integration tests (requires EdgeX environment)
	@echo "Running integration tests..."
	@go test -tags=integration -v ./...
	@echo "✅ Integration tests completed"

# Benchmark targets
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...
	@echo "✅ Benchmarks completed"

# Security scan
security: install-tools ## Run security scan
	@echo "Running security scan..."
	@gosec ./...
	@echo "✅ Security scan completed"

# All quality checks
check: fmt vet lint test ## Run all quality checks

# CI/CD target
ci: deps check test-coverage ## Run CI pipeline

# Development setup
setup: deps install-tools ## Setup development environment
	@echo "✅ Development environment setup completed"

# Show project information
info: ## Show project information
	@echo "EdgeX MessageBus Client"
	@echo "======================"
	@echo "Go version: $(shell go version)"
	@echo "Module: $(shell go list -m)"
	@echo "Dependencies:"
	@go list -m all | grep -v "$(shell go list -m)" | head -10
	@echo ""
	@echo "Project structure:"
	@find . -name "*.go" -not -path "./vendor/*" | head -10
