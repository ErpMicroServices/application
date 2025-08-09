# Common Go Library for ERP Microservices

A comprehensive shared library for the ERP microservices system, providing common utilities, types, and helpers for Go-based services.

## Features

- **UUID Utilities**: UUID generation, validation, and conversion helpers
- **Audit Fields**: Standardized audit trail fields (created_at, updated_at, created_by, updated_by)
- **Pagination**: GraphQL-compatible pagination helpers with cursor-based and offset-based pagination
- **Error Handling**: Custom error types and standardized error handling utilities
- **Logging**: Structured logging using zerolog with ERP-specific context
- **GraphQL Scalars**: Custom scalar types for DateTime, UUID, Money, Date
- **Database Helpers**: Transaction management, null handling, and common database patterns
- **Validation**: Common validation functions for business logic
- **Type Definitions**: Shared entity types and interfaces
- **HTTP Middleware**: Common middleware for authentication, logging, and request handling

## Installation

```bash
go get github.com/erpmicroservices/common-go
```

## Quick Start

```go
package main

import (
    "github.com/erpmicroservices/common-go/pkg/logging"
    "github.com/erpmicroservices/common-go/pkg/uuid"
)

func main() {
    // Initialize structured logging
    logger := logging.NewLogger("my-service")
    
    // Generate UUID
    id := uuid.New()
    logger.Info().Str("id", id.String()).Msg("Generated new ID")
}
```

## Package Documentation

### UUID Package
Provides UUID utilities with validation and conversion helpers:
- `New()` - Generate new UUID v4
- `Parse(string)` - Parse UUID from string with validation
- `IsValid(string)` - Validate UUID string format

### Audit Package
Standardized audit fields for all entities:
- `AuditFields` struct with created/updated timestamps and user IDs
- `WithAudit()` - Add audit fields to entities
- Automatic timestamp management

### Pagination Package
GraphQL-compatible pagination:
- Cursor-based pagination with forward/backward navigation
- Offset-based pagination for simple use cases
- Connection types following GraphQL Cursor Connections Specification

### Error Handling
Structured error handling with context:
- `BusinessError` - Domain-specific business logic errors
- `ValidationError` - Input validation errors
- `NotFoundError` - Resource not found errors
- Error wrapping with stack traces and context

### Logging Package
Structured logging with ERP context:
- Service-level loggers with consistent formatting
- Request tracing and correlation IDs
- Performance monitoring helpers
- Configurable log levels and output formats

### GraphQL Scalars
Custom scalar types:
- `DateTime` - RFC3339 timestamp with timezone
- `UUID` - UUID v4 with validation
- `Money` - Decimal-based monetary values
- `Date` - Date-only values (YYYY-MM-DD)

### Database Package
Database interaction helpers:
- Transaction management utilities
- Null value handling for optional fields
- Common query patterns
- Connection pool management

### Validation Package
Business logic validation:
- Email validation
- Phone number validation
- Address validation
- Custom validator registration

### Types Package
Shared entity definitions:
- Base entity interfaces
- Common enums and constants
- Status types
- Address and contact mechanism types

### Middleware Package
HTTP middleware for common concerns:
- Authentication and authorization
- Request logging and tracing
- CORS handling
- Rate limiting
- Error handling

## Architecture Patterns

This library follows these architectural patterns:

1. **Domain-Driven Design**: Types and validation align with business domains
2. **Clean Architecture**: Clear separation of concerns with interfaces
3. **SOLID Principles**: Single responsibility, dependency inversion
4. **Error Handling**: Explicit error types with context
5. **Observability**: Built-in logging, tracing, and metrics

## Configuration

Many packages support configuration through environment variables or config structs:

```go
// Logging configuration
logger := logging.NewLogger("service-name").
    WithLevel(logging.InfoLevel).
    WithFormat(logging.JSONFormat)

// Database configuration  
db := database.NewConnection(&database.Config{
    Host:     "localhost",
    Port:     5432,
    Database: "erp_db",
    Username: "user",
    Password: "pass",
})
```

## Contributing

1. Follow Go best practices and idioms
2. Add comprehensive tests for new functionality
3. Update documentation for API changes
4. Use conventional commit messages
5. Ensure backwards compatibility

## License

Apache License 2.0 - see LICENSE file for details.

## Dependencies

- Go 1.21+
- github.com/google/uuid
- github.com/rs/zerolog
- github.com/shopspring/decimal
- github.com/99designs/gqlgen
- go.uber.org/multierr