# Human Resources GraphQL API

A GraphQL API service for managing human resources data in the ERP microservices system.

## Features

- **Employee Management**: CRUD operations for employee records
- **Position Management**: Job positions and role definitions
- **Department Management**: Organizational structure management
- **Authentication**: JWT-based authentication with role-based access control
- **Authorization**: Role-based permissions (HR_ADMIN, HR_MANAGER, HR_USER)
- **Apollo Federation**: Compatible with GraphQL federation gateway
- **Health Checks**: Built-in health and readiness endpoints
- **Observability**: Structured logging with zerolog
- **Database**: PostgreSQL with connection pooling
- **Caching**: Redis integration ready

## Quick Start

### Prerequisites

- Go 1.23 or later
- PostgreSQL 15+
- Redis (optional, for caching)

### Development Setup

1. **Clone and navigate to the project:**
   ```bash
   cd human_resources-endpoint-graphql
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Generate GraphQL code:**
   ```bash
   make generate
   ```

4. **Run the development server:**
   ```bash
   make run
   ```

   The API will be available at:
   - GraphQL Endpoint: http://localhost:8080/graphql
   - GraphQL Playground: http://localhost:8080/playground
   - Health Check: http://localhost:8080/health

### Docker Setup

1. **Build and run with Docker Compose:**
   ```bash
   make docker-up
   ```

2. **Stop services:**
   ```bash
   make docker-down
   ```

## Configuration

Configuration can be provided via:

- Environment variables
- YAML config file (`config.yaml`)
- Command line flags

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server host address |
| `SERVER_PORT` | `8080` | Server port |
| `DATABASE_HOST` | `localhost` | PostgreSQL host |
| `DATABASE_PORT` | `5432` | PostgreSQL port |
| `DATABASE_NAME` | `human_resources_db` | Database name |
| `DATABASE_USER` | `human_resources_user` | Database user |
| `DATABASE_PASSWORD` | `human_resources_password` | Database password |
| `AUTH_ENABLED` | `false` | Enable authentication |
| `AUTH_JWT_SECRET` | `your-secret-key` | JWT secret key |
| `GRAPHQL_PLAYGROUND` | `true` | Enable GraphQL playground |
| `LOGGING_LEVEL` | `info` | Log level (debug, info, warn, error) |

## API Schema

### Types

- **Employee**: Employee records with personal and employment information
- **Position**: Job positions with requirements and responsibilities
- **Department**: Organizational departments with hierarchy support

### Queries

```graphql
query {
  employees {
    id
    employeeId
    firstName
    lastName
    email
    position {
      title
    }
    department {
      name
    }
  }
}
```

### Mutations

```graphql
mutation {
  createEmployee(input: {
    employeeId: "EMP001"
    firstName: "John"
    lastName: "Doe"
    email: "john.doe@company.com"
    hireDate: "2024-01-15T00:00:00Z"
  }) {
    id
    employeeId
    firstName
    lastName
  }
}
```

## Development

### Project Structure

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── graph/          # GraphQL resolvers
│   ├── repositories/   # Data access layer
│   └── services/       # Business logic
├── pkg/
│   ├── directives/     # GraphQL directives
│   ├── middleware/     # HTTP middleware
│   └── models/         # Data models
├── graph/              # Generated GraphQL code
├── schema.graphql      # GraphQL schema definition
└── gqlgen.yml         # gqlgen configuration
```

### Available Commands

```bash
make help              # Show all available commands
make deps              # Install dependencies
make generate          # Generate GraphQL code
make build             # Build the application
make run               # Run the application
make dev               # Run with live reload
make test              # Run tests
make test-coverage     # Run tests with coverage
make lint              # Lint code
make docker-build      # Build Docker image
make ci                # Run full CI pipeline
```

### Adding New Features

1. **Update GraphQL Schema**: Modify `schema.graphql`
2. **Generate Code**: Run `make generate`
3. **Implement Resolvers**: Add resolver logic in `graph/`
4. **Add Business Logic**: Implement services in `internal/services/`
5. **Add Data Access**: Implement repositories in `internal/repositories/`
6. **Write Tests**: Add tests for new functionality

## Authentication & Authorization

The API supports JWT-based authentication with role-based access control:

- **HR_ADMIN**: Full access to all HR operations
- **HR_MANAGER**: Access to employee and position management
- **HR_USER**: Read-only access to employee information

### Example JWT Claims

```json
{
  "sub": "user123",
  "roles": ["HR_MANAGER"],
  "exp": 1640995200
}
```

## Deployment

### Production Build

```bash
make ci                # Run full CI pipeline
make docker-build      # Build production image
```

### Kubernetes

Deploy using the provided Kubernetes manifests in the `deployments/` directory:

```bash
kubectl apply -f deployments/
```

## Monitoring & Observability

### Health Checks

- `/health` - Basic health check
- `/ready` - Readiness check with dependency validation
- `/metrics` - Basic metrics endpoint

### Logging

Structured JSON logging with configurable levels:

```bash
# Set log level
export LOGGING_LEVEL=debug

# Enable console output
export LOGGING_FORMAT=console
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the CI pipeline: `make ci`
6. Submit a pull request

## License

This project is licensed under the Apache License 2.0.

## Support

For questions and support, please refer to the main ERP microservices documentation.