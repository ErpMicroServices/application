# Auth-Go - OAuth2 Authentication Module for ERP Microservices

A comprehensive OAuth2 authentication module for the ERP microservices system, providing secure integration with the existing Spring Boot authorization server.

## Features

- **OAuth2 Integration**: Full support for Authorization Code and Client Credentials flows
- **JWT Token Validation**: Parse and validate JWT tokens from the authorization server
- **Service-to-Service Authentication**: Secure inter-service communication
- **Role-Based Access Control (RBAC)**: Comprehensive authorization middleware
- **GraphQL Directives**: Built-in authentication and authorization directives
- **Token Caching**: Efficient token storage and refresh mechanisms
- **Configuration Management**: Environment-based configuration
- **Comprehensive Logging**: Structured logging with multiple levels
- **Test Coverage**: Extensive unit and integration tests

## Quick Start

### 1. Copy Environment Configuration

```bash
cp .env.example .env
# Edit .env with your specific configuration
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run Example Application

```bash
cd cmd/example
go run main.go
```

## Architecture

### Package Structure

```
auth-go/
├── pkg/
│   ├── oauth2/         # OAuth2 client implementation
│   ├── jwt/           # JWT token parsing and validation
│   ├── middleware/    # HTTP middleware for authentication
│   ├── directives/    # GraphQL directives
│   └── cache/         # Token caching mechanisms
├── internal/
│   └── config/        # Configuration management
└── cmd/
    └── example/       # Example application
```

### Core Components

#### OAuth2 Client (`pkg/oauth2/`)
- **Client**: Main OAuth2 client for interacting with authorization server
- **Token**: Token management and refresh logic
- **Validation**: Token validation and user information retrieval

#### JWT Handler (`pkg/jwt/`)
- **Parser**: JWT token parsing and validation
- **Claims**: Custom claims handling for ERP system

#### Middleware (`pkg/middleware/`)
- **Auth**: Authentication middleware for HTTP requests
- **RBAC**: Role-based access control middleware

#### GraphQL Directives (`pkg/directives/`)
- **@auth**: Require authentication for GraphQL fields
- **@hasRole**: Require specific roles for field access

## Usage

### Basic Authentication Middleware

```go
package main

import (
    "net/http"
    
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
    "github.com/erpmicroservices/auth-go/pkg/middleware"
    "github.com/erpmicroservices/auth-go/internal/config"
)

func main() {
    cfg := config.Load()
    
    oauth2Client := oauth2.NewClient(cfg.OAuth2)
    authMiddleware := middleware.NewAuth(oauth2Client)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/protected", protectedHandler)
    
    // Apply authentication middleware
    handler := authMiddleware.RequireAuth(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### GraphQL Integration

```go
package main

import (
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/erpmicroservices/auth-go/pkg/directives"
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
)

func main() {
    cfg := config.Load()
    oauth2Client := oauth2.NewClient(cfg.OAuth2)
    
    srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
        Resolvers: &graph.Resolver{},
        Directives: generated.DirectiveRoot{
            Auth:    directives.Auth(oauth2Client),
            HasRole: directives.HasRole(oauth2Client),
        },
    }))
    
    http.Handle("/graphql", srv)
}
```

### Service-to-Service Authentication

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
)

func main() {
    cfg := config.Load()
    oauth2Client := oauth2.NewClient(cfg.OAuth2)
    
    // Get client credentials token for service-to-service communication
    token, err := oauth2Client.GetClientCredentialsToken(context.Background())
    if err != nil {
        // Handle error
    }
    
    // Use token in HTTP requests
    req, _ := http.NewRequest("GET", "http://other-service/api/data", nil)
    req.Header.Set("Authorization", "Bearer "+token.AccessToken)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    // Handle response
}
```

## Configuration

### Environment Variables

All configuration is handled through environment variables. See `.env.example` for all available options.

### OAuth2 Configuration

The module expects the Spring Boot authorization server to be running with the following endpoints:

- **Token Endpoint**: `/oauth2/token`
- **Authorization Endpoint**: `/oauth2/authorize`
- **User Info Endpoint**: `/oauth2/userinfo`
- **Token Introspection**: `/oauth2/introspect`
- **JWKS Endpoint**: `/oauth2/jwks`

### Required Client Configuration

In your Spring Boot authorization server, ensure you have a registered client with:

- Client ID matching `OAUTH2_CLIENT_ID`
- Client secret matching `OAUTH2_CLIENT_SECRET`
- Redirect URI matching `OAUTH2_REDIRECT_URL`
- Grant types: `authorization_code`, `client_credentials`, `refresh_token`
- Scopes: `read`, `write` (or as needed for your application)

## Integration with ERP Microservices

### Common Patterns

This module is designed to integrate seamlessly with the existing ERP microservices architecture:

1. **Database Services**: Use for securing admin endpoints
2. **GraphQL Endpoints**: Integrate directives for field-level security
3. **UI Services**: Handle user authentication flows
4. **Service-to-Service**: Secure internal API communications

### Role-Based Access Control

The module supports the role/authority structure from the authorization server:

```go
// Example roles from authorization server
roles := []string{"ADMIN", "USER", "SERVICE_ACCOUNT"}

// Use in middleware
rbacMiddleware := middleware.NewRBAC(oauth2Client)
handler := rbacMiddleware.RequireRole("ADMIN")(protectedHandler)
```

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/oauth2/
```

### Integration Tests

Integration tests require the authorization server to be running:

```bash
# Start authorization server (from authorization_server directory)
./gradlew bootRun --args='--spring.profiles.active=development'

# Run integration tests
go test -tags=integration ./...
```

## Docker Support

### Build Image

```bash
docker build -t erp-microservices/auth-go .
```

### Run Container

```bash
docker run -p 8080:8080 --env-file .env erp-microservices/auth-go
```

## Contributing

1. Follow Go conventions and formatting (`go fmt`, `go vet`)
2. Write tests for new functionality
3. Update documentation for API changes
4. Ensure integration tests pass

## Security Considerations

- **Token Storage**: Tokens are cached in memory by default
- **HTTPS**: Always use HTTPS in production
- **Secret Management**: Use secure secret management for production
- **Token Rotation**: Implement proper token refresh mechanisms
- **Rate Limiting**: Configure appropriate rate limits for your use case

## License

This project follows the same license as the parent ERP microservices project.