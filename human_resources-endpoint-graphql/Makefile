# Human Resources GraphQL API Makefile
.PHONY: build run test clean generate docker-build docker-run deps tidy

# Variables
APP_NAME := human-resources-endpoint-graphql
DOCKER_IMAGE := erp-microservices/$(APP_NAME)
PORT := 8080

# Default target
all: deps tidy generate build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Generate GraphQL code
generate:
	@echo "Generating GraphQL code..."
	go run github.com/99designs/gqlgen generate

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) ./cmd/server

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run ./cmd/server

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@echo "Starting development server with live reload..."
	air

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE)

# Docker compose up
docker-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

# Docker compose down
docker-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

# Database migration (placeholder - implement based on your migration tool)
migrate-up:
	@echo "Running database migrations..."
	# Add your migration command here

migrate-down:
	@echo "Rolling back database migrations..."
	# Add your migration rollback command here

# Security check
sec-check:
	@echo "Running security checks..."
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Full CI pipeline
ci: deps tidy generate fmt lint test-coverage sec-check build

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Install deps, tidy, generate, and build"
	@echo "  deps          - Install Go module dependencies"
	@echo "  tidy          - Tidy Go module dependencies"
	@echo "  generate      - Generate GraphQL code using gqlgen"
	@echo "  build         - Build the application binary"
	@echo "  run           - Run the application"
	@echo "  dev           - Run with live reload (requires air)"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format Go code"
	@echo "  lint          - Lint Go code (requires golangci-lint)"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-up     - Start services with docker-compose"
	@echo "  docker-down   - Stop services with docker-compose"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  sec-check     - Run security vulnerability checks"
	@echo "  ci            - Run full CI pipeline"
	@echo "  help          - Show this help message"