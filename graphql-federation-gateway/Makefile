# GraphQL Federation Gateway (Go) - Makefile
# Strict BDD/TDD Development with Quality Gates

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Application parameters
BINARY_NAME=graphql-federation-gateway
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd/gateway

# Coverage requirements
COVERAGE_THRESHOLD=85
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Quality gate parameters
LINT_TIMEOUT=10m

# Docker parameters
DOCKER_IMAGE=erp/graphql-federation-gateway
DOCKER_TAG=latest

.PHONY: all build test clean run deps fmt vet lint coverage quality-gates

all: quality-gates build

# Development commands
deps:
	@echo "📦 Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies installed"

fmt:
	@echo "🎨 Formatting code..."
	$(GOFMT) ./...
	@echo "✅ Code formatted"

vet:
	@echo "🔍 Running go vet..."
	$(GOVET) ./...
	@echo "✅ go vet passed"

lint:
	@echo "🧹 Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	}
	golangci-lint run --timeout=$(LINT_TIMEOUT)
	@echo "✅ Linting passed"

# Testing commands
test:
	@echo "🧪 Running unit tests..."
	$(GOTEST) -v -race ./...
	@echo "✅ Unit tests passed"

test-short:
	@echo "🧪 Running short tests..."
	$(GOTEST) -short -race ./...
	@echo "✅ Short tests passed"

test-bdd:
	@echo "🥒 Running BDD scenarios..."
	@command -v godog >/dev/null 2>&1 || { \
		echo "Installing godog..."; \
		$(GOGET) github.com/cucumber/godog/cmd/godog@latest; \
	}
	godog ../graphql-federation-gateway-features/*.feature
	@echo "✅ BDD scenarios passed"

test-integration:
	@echo "🔗 Running integration tests..."
	$(GOTEST) -v -tags=integration ./test/integration/...
	@echo "✅ Integration tests passed"

coverage:
	@echo "📊 Generating test coverage..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "✅ Coverage report generated"

coverage-html: coverage
	@echo "🌐 Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✅ HTML coverage report: $(COVERAGE_HTML)"

coverage-check: coverage
	@echo "🎯 Checking coverage threshold ($(COVERAGE_THRESHOLD)%)..."
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $${COVERAGE}% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

# Quality gates - CRITICAL: Must pass before any merge
quality-gates: deps fmt vet lint test coverage-check
	@echo "🚪 All quality gates passed! ✅"

# Build commands
build: quality-gates
	@echo "🏗️  Building application..."
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)
	@echo "✅ Build complete: $(BINARY_NAME)"

build-linux: quality-gates
	@echo "🏗️  Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(MAIN_PATH)
	@echo "✅ Linux build complete: $(BINARY_UNIX)"

# Run commands
run: build
	@echo "🚀 Starting application..."
	./$(BINARY_NAME)

run-dev:
	@echo "🚀 Starting in development mode..."
	$(GOCMD) run $(MAIN_PATH)/main.go

# Docker commands
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: docker-build
	@echo "🐳 Running Docker container..."
	docker run -p 4000:4000 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Kubernetes commands
k8s-deploy:
	@echo "☸️  Deploying to Kubernetes..."
	kubectl apply -f deployments/kubernetes/
	@echo "✅ Kubernetes deployment complete"

k8s-delete:
	@echo "☸️  Deleting from Kubernetes..."
	kubectl delete -f deployments/kubernetes/
	@echo "✅ Kubernetes resources deleted"

# Database and migration commands (if needed)
db-migrate:
	@echo "🗄️  Running database migrations..."
	# Add migration commands here if database is used
	@echo "✅ Database migrations complete"

# Clean commands
clean:
	@echo "🧹 Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)
	@echo "✅ Clean complete"

# Security and audit commands
security-audit:
	@echo "🔒 Running security audit..."
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "Installing govulncheck..."; \
		$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	govulncheck ./...
	@echo "✅ Security audit passed"

# Development workflow commands
pre-commit: quality-gates security-audit
	@echo "✅ Pre-commit checks passed - ready to commit!"

ci: quality-gates security-audit test-integration test-bdd
	@echo "✅ CI pipeline passed!"

# Help command
help:
	@echo "GraphQL Federation Gateway - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run linter"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run unit tests"
	@echo "  test-short     - Run short tests"
	@echo "  test-bdd       - Run BDD scenarios"
	@echo "  test-integration - Run integration tests"
	@echo "  coverage       - Generate coverage report"
	@echo "  coverage-html  - Generate HTML coverage report"
	@echo "  coverage-check - Check coverage threshold ($(COVERAGE_THRESHOLD)%)"
	@echo ""
	@echo "Quality Gates (REQUIRED):"
	@echo "  quality-gates  - Run all quality checks"
	@echo "  pre-commit     - Run pre-commit checks"
	@echo "  ci             - Run full CI pipeline"
	@echo ""
	@echo "Build & Run:"
	@echo "  build          - Build application"
	@echo "  run            - Build and run application"
	@echo "  run-dev        - Run in development mode"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Build and run Docker container"
	@echo ""
	@echo "Deployment:"
	@echo "  k8s-deploy     - Deploy to Kubernetes"
	@echo "  k8s-delete     - Delete from Kubernetes"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean          - Clean build artifacts"
	@echo "  security-audit - Run security audit"

# Default target
.DEFAULT_GOAL := help