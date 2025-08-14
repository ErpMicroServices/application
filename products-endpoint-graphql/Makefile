# Products GraphQL API Makefile

# Variables
BINARY_NAME=products-graphql-api
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*" -not -path "./graph/generated/*")
DOCKER_IMAGE=erp-microservices/products-graphql-api
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
LDFLAGS=-ldflags="-w -s -X main.Version=${VERSION}"

# Default target
.DEFAULT_GOAL := help

## Build targets

.PHONY: build
build: ## Build the application binary
	@echo "Building ${BINARY_NAME}..."
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/server

.PHONY: build-linux
build-linux: ## Build the application binary for Linux
	@echo "Building ${BINARY_NAME} for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux ./cmd/server

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	go clean

## Development targets

.PHONY: run
run: ## Run the application locally
	@echo "Running ${BINARY_NAME}..."
	go run ./cmd/server

.PHONY: dev
dev: ## Run in development mode with hot reload
	@echo "Starting development server with hot reload..."
	which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

.PHONY: generate
generate: ## Generate GraphQL code
	@echo "Generating GraphQL code..."
	go generate ./...

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: vendor
vendor: ## Create vendor directory
	@echo "Creating vendor directory..."
	go mod vendor

## Testing targets

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -race -short -coverprofile=coverage-unit.out ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v -race -run Integration -coverprofile=coverage-integration.out ./...

.PHONY: test-bdd
test-bdd: ## Run BDD tests using Cucumber
	@echo "Running BDD tests..."
	cd ../products-features && npm test

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-coverage-func
test-coverage-func: test ## Show function-level coverage
	go tool cover -func=coverage.out

.PHONY: bench
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race ./...

## Quality targets

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	gofmt -s -w ${GO_FILES}

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@diff -u <(echo -n) <(gofmt -d ${GO_FILES})

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: sec
sec: ## Run security scan
	@echo "Running security scan..."
	which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...

.PHONY: quality
quality: fmt-check vet lint sec test ## Run all quality checks

## Docker targets

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t ${DOCKER_IMAGE}:${VERSION} -t ${DOCKER_IMAGE}:latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 ${DOCKER_IMAGE}:latest

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image..."
	docker push ${DOCKER_IMAGE}:${VERSION}
	docker push ${DOCKER_IMAGE}:latest

.PHONY: docker-compose-up
docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop all services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

## Database targets

.PHONY: db-migrate-up
db-migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate -path ./migrations -database "postgresql://products_user:products_password@localhost:5432/products_db?sslmode=disable" up

.PHONY: db-migrate-down
db-migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	migrate -path ./migrations -database "postgresql://products_user:products_password@localhost:5432/products_db?sslmode=disable" down

.PHONY: db-migrate-create
db-migrate-create: ## Create new database migration (usage: make db-migrate-create name=migration_name)
	@echo "Creating new migration: ${name}..."
	migrate create -ext sql -dir ./migrations ${name}

.PHONY: db-reset
db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	docker-compose exec postgres psql -U products_user -d postgres -c "DROP DATABASE IF EXISTS products_db;"
	docker-compose exec postgres psql -U products_user -d postgres -c "CREATE DATABASE products_db;"

## Performance targets

.PHONY: profile
profile: ## Run performance profiling
	@echo "Starting profiling server..."
	go test -cpuprofile cpu.prof -memprofile mem.prof -bench . ./...

.PHONY: profile-cpu
profile-cpu: ## Analyze CPU profile
	go tool pprof cpu.prof

.PHONY: profile-mem
profile-mem: ## Analyze memory profile
	go tool pprof mem.prof

## Load testing

.PHONY: load-test
load-test: ## Run load tests
	@echo "Running load tests..."
	which k6 > /dev/null || echo "Please install k6 for load testing"
	k6 run scripts/load-test.js

## Documentation targets

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Documentation server starting at http://localhost:6060"
	godoc -http=:6060

.PHONY: swagger
swagger: ## Generate Swagger documentation for REST endpoints
	@echo "Generating Swagger documentation..."
	which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g ./cmd/server/main.go

## CI/CD targets

.PHONY: ci
ci: quality test-coverage ## Run CI pipeline
	@echo "CI pipeline completed successfully"

.PHONY: release
release: clean build-linux docker-build ## Prepare release artifacts
	@echo "Release artifacts prepared for version: ${VERSION}"

## Utility targets

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/99designs/gqlgen@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install golang.org/x/tools/cmd/godoc@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: mod-check
mod-check: ## Check if go.mod is tidy
	@echo "Checking go.mod..."
	go mod tidy
	git diff --exit-code go.mod go.sum

.PHONY: version
version: ## Show version
	@echo ${VERSION}

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Configuration targets

.PHONY: config-check
config-check: ## Validate configuration
	@echo "Validating configuration..."
	go run ./cmd/server -config-check

.PHONY: config-generate
config-generate: ## Generate sample configuration file
	@echo "Generating sample configuration..."
	cp config/config.example.yaml config/config.yaml
	@echo "Configuration generated: config/config.yaml"

# Development shortcuts

.PHONY: all
all: clean deps generate quality test build ## Run full development cycle

.PHONY: quick
quick: fmt test ## Quick development check

.PHONY: up
up: docker-compose-up ## Quick start with docker-compose

.PHONY: down
down: docker-compose-down ## Quick stop with docker-compose