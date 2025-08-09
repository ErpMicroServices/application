# GraphQL Federation Gateway (Go)

Revolutionary Apollo Federation v2 Gateway built in Go following strict BDD/TDD methodology for the ERP microservices system.

## Overview

This gateway serves as the central API orchestration layer, federating multiple GraphQL services into a unified schema while providing authentication, authorization, monitoring, and real-time subscriptions.

## Features

### Core Federation
- **Apollo Federation v2** - Full specification compliance
- **Schema Composition** - Dynamic service discovery and schema composition
- **Query Planning** - Intelligent query execution across services
- **Entity Resolution** - Cross-service entity relationships

### Security & Authentication
- **JWT Authentication** - Integration with auth-go module
- **Role-based Authorization** - Business context permission system
- **Token Validation** - Cached token validation with refresh
- **Security Headers** - CORS, rate limiting, and security middleware

### Real-time Features
- **GraphQL Subscriptions** - WebSocket-based real-time updates
- **Subscription Federation** - Multiplexed subscriptions across services
- **Connection Management** - Graceful connection handling and cleanup

### Observability & Monitoring
- **Prometheus Metrics** - Comprehensive performance metrics
- **Distributed Tracing** - Request tracing across federated services
- **Health Checks** - Service health monitoring and reporting
- **Logging** - Structured logging with zerolog

### Performance & Resilience
- **Query Complexity Analysis** - Configurable complexity limits
- **Circuit Breakers** - Service failure protection
- **Request Batching** - Efficient request batching and caching
- **Connection Pooling** - Optimized HTTP client pooling

## Architecture

### Directory Structure

```
graphql-federation-gateway/
├── cmd/gateway/            # Application entry point
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── federation/        # Apollo Federation logic
│   ├── resolvers/         # GraphQL resolvers
│   ├── middleware/        # HTTP/GraphQL middleware
│   └── subscriptions/     # WebSocket subscription handling
├── pkg/                   # Public packages
│   ├── gateway/           # Gateway server interface
│   └── client/            # Client utilities
├── api/                   # API definitions
├── test/                  # Test suites
│   ├── integration/       # Integration tests
│   └── fixtures/          # Test data and mocks
└── deployments/           # Deployment configurations
    ├── kubernetes/        # K8s manifests
    └── docker/            # Docker configurations
```

### Federated Services

| Service | Port | Domain |
|---------|------|---------|
| people_and_organizations | 8081 | People, contacts, relationships |
| e_commerce | 8082 | Orders, shopping, user preferences |
| products | 8084 | Product catalog, inventory |
| accounting_and_budgeting | 8083 | Financial data, budgets |
| orders | 8085 | Order management |
| invoices | 8086 | Invoice processing |
| shipments | 8087 | Shipping and logistics |
| human_resources | 8088 | Employee management |
| work_effort | 8089 | Projects, tasks, time tracking |
| analytics | 8090 | Business intelligence (optional) |

## Development

### Prerequisites

- **Go 1.21+** - Latest Go version
- **Docker** - For containerization
- **Make** - For build automation
- **golangci-lint** - For code quality (auto-installed)
- **godog** - For BDD testing (auto-installed)

### Quick Start

1. **Clone and Setup**
   ```bash
   git clone <repo-url>
   cd graphql-federation-gateway
   cp .env.gateway.example .env
   ```

2. **Install Dependencies**
   ```bash
   make deps
   ```

3. **Run Quality Gates**
   ```bash
   make quality-gates
   ```

4. **Build and Run**
   ```bash
   make build
   make run
   ```

5. **Development Mode**
   ```bash
   make run-dev
   ```

### BDD/TDD Development Workflow

This project follows strict **BDD (Behavior-Driven Development)** and **TDD (Test-Driven Development)** methodology:

#### 1. BDD Scenarios First
```bash
# BDD features are in ../graphql-federation-gateway-features/
make test-bdd
```

#### 2. Failing Tests First
```bash
# Write failing unit tests before implementation
make test
```

#### 3. Quality Gates (85% Coverage)
```bash
# REQUIRED: All quality gates must pass
make quality-gates
```

#### 4. CI/CD Pipeline
```bash
# Full CI pipeline with all checks
make ci
```

### Testing

#### Unit Tests
```bash
# Run all unit tests
make test

# Run with coverage
make coverage

# Check coverage threshold (85%)
make coverage-check

# Generate HTML coverage report
make coverage-html
```

#### BDD Tests
```bash
# Run all BDD scenarios
make test-bdd

# BDD features are defined in:
# ../graphql-federation-gateway-features/*.feature
```

#### Integration Tests
```bash
# Run integration tests with real services
make test-integration
```

### Quality Gates

**CRITICAL**: These quality gates must pass before any commit:

| Gate | Requirement | Command |
|------|-------------|---------|
| **Formatting** | `gofmt` compliant | `make fmt` |
| **Vetting** | `go vet` clean | `make vet` |
| **Linting** | golangci-lint clean | `make lint` |
| **Unit Tests** | All tests pass | `make test` |
| **Coverage** | ≥85% test coverage | `make coverage-check` |

```bash
# Run all quality gates
make quality-gates
```

## Configuration

### Environment Variables

Copy `.env.gateway.example` to `.env` and configure:

#### Server Configuration
- `SERVER_PORT` - Server port (default: 4000)
- `SERVER_HOST` - Bind host (default: 0.0.0.0)
- `SERVER_*_TIMEOUT` - Various timeout settings

#### Service URLs
- `*_SERVICE_URL` - Individual service GraphQL endpoints
- `*_TIMEOUT` - Per-service timeout settings
- `*_ENABLED` - Enable/disable individual services

#### Authentication
- `JWT_SECRET` - JWT signing secret
- `AUTH_SERVICE_URL` - Authentication service URL
- `TOKEN_VALIDATION` - Enable/disable token validation

#### Monitoring
- `ENABLE_METRICS` - Prometheus metrics
- `ENABLE_TRACING` - Distributed tracing
- `LOG_LEVEL` - Logging level

### Federation Configuration
- `MAX_QUERY_COMPLEXITY` - Query complexity limit
- `QUERY_TIMEOUT` - Maximum query execution time
- `BATCHING_ENABLED` - Enable request batching

## Deployment

### Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

### Kubernetes

```bash
# Deploy to Kubernetes
make k8s-deploy

# Delete deployment
make k8s-delete
```

### Health Checks

- **Health**: `GET /health` - Service health status
- **Metrics**: `GET /metrics` - Prometheus metrics
- **GraphQL**: `POST /graphql` - GraphQL endpoint
- **Playground**: `GET /` - GraphQL playground (dev only)

## Integration Points

### auth-go Module
```go
import "github.com/erpmicroservices/auth-go"
```
- JWT token validation
- Role-based authorization
- User context extraction

### common-go Module
```go
import "github.com/erpmicroservices/common-go"
```
- Shared utilities
- Error handling
- Logging infrastructure

### graphql-foundation Module
```go
import "github.com/erpmicroservices/graphql-foundation"
```
- Base GraphQL utilities
- Common middleware
- Schema helpers

## Performance

### Benchmarks

- **Response Time**: <100ms for simple queries
- **Throughput**: 1000+ concurrent requests
- **Memory Usage**: <512MB baseline
- **Query Complexity**: Configurable limits

### Monitoring

- **Prometheus Metrics**: Query duration, error rates, service health
- **Distributed Tracing**: Request flow across services
- **Custom Metrics**: Business-specific KPIs

## Security

### Authentication Flow
1. Client sends request with JWT token
2. Gateway validates token with auth-go module
3. User context extracted and passed to resolvers
4. Business rules applied based on user roles

### Authorization
- Field-level authorization
- Business context validation
- Rate limiting per user/IP
- Query complexity analysis

## Contributing

### Development Standards

1. **BDD First**: Write Gherkin scenarios before code
2. **TDD Approach**: Write failing tests, then implement
3. **Quality Gates**: 85% coverage, all lints pass
4. **Code Review**: All changes require review
5. **Documentation**: Keep README and docs current

### Commit Workflow

```bash
# Before committing
make pre-commit

# This runs: quality-gates + security-audit
```

### Branch Strategy

- `main` - Production ready code
- `develop` - Integration branch
- `feature/*` - Feature branches
- `bugfix/*` - Bug fix branches

## Troubleshooting

### Common Issues

1. **Service Connection Failures**
   - Check service URLs in `.env`
   - Verify services are running
   - Check network connectivity

2. **Authentication Errors**
   - Verify JWT_SECRET configuration
   - Check auth-go module integration
   - Validate token format

3. **Performance Issues**
   - Check query complexity limits
   - Review service response times
   - Monitor memory usage

4. **Build Failures**
   - Run `make clean && make deps`
   - Check Go version compatibility
   - Verify all quality gates pass

### Debugging

```bash
# Run with debug logging
LOG_LEVEL=debug make run-dev

# Check service health
curl http://localhost:4000/health

# View metrics
curl http://localhost:9090/metrics

# Test GraphQL endpoint
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ __schema { queryType { name } } }"}'
```

## License

Apache 2.0 - See LICENSE file for details.

## Support

- **Documentation**: This README and inline code documentation
- **BDD Scenarios**: `../graphql-federation-gateway-features/`
- **Integration Examples**: `test/integration/`
- **Configuration Examples**: `.env.gateway.example`

---

**Built with revolutionary BDD/TDD methodology ensuring 85%+ test coverage and production-ready quality gates.**