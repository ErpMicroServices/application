# Makefile for ERP Common Go Library

.PHONY: help build test lint clean fmt vet deps example audit

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build all packages"
	@echo "  test      - Run all tests"
	@echo "  lint      - Run linters (requires golangci-lint)"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  clean     - Clean build artifacts"
	@echo "  deps      - Download dependencies"
	@echo "  example   - Run the basic usage example"
	@echo "  audit     - Audit dependencies for security issues"

# Build all packages
build:
	@echo "Building all packages..."
	@go build ./...

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Benchmark tests
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run linters (requires golangci-lint to be installed)
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/" && exit 1)
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@go clean ./...
	@rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run the basic usage example
example:
	@echo "Running basic usage example..."
	@go run ./examples/basic_usage.go

# Audit dependencies for security vulnerabilities
audit:
	@echo "Auditing dependencies..."
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Generate documentation
docs:
	@echo "Generating documentation..."
	@which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Visit http://localhost:6060/pkg/github.com/erpmicroservices/common-go/"
	@godoc -http=:6060

# Check for outdated dependencies
check-updates:
	@echo "Checking for outdated dependencies..."
	@go list -u -m all

# Run all quality checks
quality: fmt vet lint test
	@echo "All quality checks passed!"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Tools installed successfully!"
	@echo "To install golangci-lint, visit: https://golangci-lint.run/"

# Create a new package template
new-package:
	@read -p "Enter package name: " pkg; \
	mkdir -p pkg/$$pkg; \
	echo "// Package $$pkg provides functionality for the ERP microservices system." > pkg/$$pkg/$$pkg.go; \
	echo "package $$pkg" >> pkg/$$pkg/$$pkg.go; \
	echo "Package $$pkg created successfully!"

# Run security scan
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec ./...

# Profile CPU usage of tests
profile-cpu:
	@echo "Profiling CPU usage..."
	@go test -cpuprofile=cpu.prof -bench=. ./...
	@go tool pprof cpu.prof

# Profile memory usage of tests  
profile-mem:
	@echo "Profiling memory usage..."
	@go test -memprofile=mem.prof -bench=. ./...
	@go tool pprof mem.prof

# Docker build (if Dockerfile exists)
docker-build:
	@if [ -f Dockerfile ]; then \
		echo "Building Docker image..."; \
		docker build -t erp-common-go .; \
	else \
		echo "Dockerfile not found"; \
	fi

# Release preparation
pre-release: clean fmt vet lint test audit
	@echo "Pre-release checks completed successfully!"
	@echo "Ready for release!"