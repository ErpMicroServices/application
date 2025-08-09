# Integration Examples

This document provides comprehensive examples of how to integrate the auth-go module with various ERP microservices.

## Table of Contents
- [Spring Boot Authorization Server Integration](#spring-boot-authorization-server-integration)
- [GraphQL Service Integration](#graphql-service-integration)
- [REST API Service Integration](#rest-api-service-integration)
- [Database Service Integration](#database-service-integration)
- [Frontend Integration](#frontend-integration)
- [Service-to-Service Communication](#service-to-service-communication)

## Spring Boot Authorization Server Integration

### 1. Configure OAuth2 Client in Authorization Server

Add the following configuration to your Spring Boot authorization server:

```yaml
# application.yml
spring:
  security:
    oauth2:
      authorizationserver:
        client:
          erp-microservices-client:
            registration:
              client-id: erp-microservices-client
              client-secret: your-client-secret-here
              client-authentication-methods:
                - client_secret_basic
                - client_secret_post
              authorization-grant-types:
                - authorization_code
                - client_credentials
                - refresh_token
              redirect-uris:
                - http://localhost:8080/auth/callback
                - http://localhost:3000/auth/callback
              scopes:
                - read
                - write
                - profile
                - openid
            require-authorization-consent: false
```

### 2. Custom Claims Configuration

```java
// JwtCustomizerConfig.java
@Configuration
public class JwtCustomizerConfig {

    @Bean
    public OAuth2TokenCustomizer<JwtEncodingContext> jwtCustomizer() {
        return context -> {
            if (context.getTokenType() == OAuth2TokenType.ACCESS_TOKEN) {
                Authentication principal = context.getPrincipal();
                
                // Add user roles
                Set<String> authorities = principal.getAuthorities().stream()
                    .map(GrantedAuthority::getAuthority)
                    .collect(Collectors.toSet());
                context.getClaims().claim("roles", authorities);
                context.getClaims().claim("authorities", authorities);
                
                // Add ERP-specific claims
                if (principal instanceof UsernamePasswordAuthenticationToken) {
                    UserDetails userDetails = (UserDetails) principal.getPrincipal();
                    if (userDetails instanceof ERPUserDetails) {
                        ERPUserDetails erpUser = (ERPUserDetails) userDetails;
                        context.getClaims().claim("organization_id", erpUser.getOrganizationId());
                        context.getClaims().claim("department_id", erpUser.getDepartmentId());
                        context.getClaims().claim("employee_id", erpUser.getEmployeeId());
                        context.getClaims().claim("tenant_id", erpUser.getTenantId());
                    }
                }
            }
        };
    }
}
```

## GraphQL Service Integration

### 1. Basic GraphQL Server Setup

```go
package main

import (
    "net/http"
    
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    
    "github.com/erpmicroservices/auth-go/internal/config"
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
    "github.com/erpmicroservices/auth-go/pkg/middleware"
    "github.com/erpmicroservices/auth-go/pkg/directives"
    "github.com/erpmicroservices/auth-go/pkg/cache"
    "github.com/erpmicroservices/auth-go/pkg/jwt"
    
    "your-service/graph"
    "your-service/graph/generated"
)

func main() {
    cfg := config.Load()
    
    // Initialize auth components
    tokenCache := cache.NewInMemoryTokenCache(cfg.Cache.TTL, cfg.Cache.CleanupInterval)
    oauth2Client := oauth2.NewClient(&cfg.OAuth2, tokenCache)
    jwtParser := jwt.NewParser(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.SigningKey)
    validator := oauth2.NewValidator(oauth2Client, cfg.OAuth2.JWKSURL, cfg.JWT.Issuer, cfg.JWT.Audience)
    
    // Setup authentication middleware
    authConfig := middleware.DefaultAuthConfig()
    authConfig.SkipPaths = []string{"/", "/graphql-playground"}
    authMiddleware := middleware.NewAuthMiddleware(oauth2Client, jwtParser, validator, authConfig)
    
    // Configure GraphQL server with directives
    c := generated.Config{Resolvers: &graph.Resolver{}}
    c.Directives.Auth = directives.Auth(oauth2Client)
    c.Directives.HasRole = directives.HasRole(oauth2Client)
    c.Directives.HasAuthority = directives.HasAuthority(oauth2Client)
    c.Directives.HasPermission = directives.HasPermission(oauth2Client)
    
    srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))
    
    http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
    http.Handle("/graphql", authMiddleware.OptionalAuth(srv))
    
    log.Fatal(http.ListenAndServe(":"+cfg.Service.Port, nil))
}
```

### 2. GraphQL Resolver with Authentication

```go
package graph

import (
    "context"
    "fmt"
    
    "github.com/erpmicroservices/auth-go/pkg/directives"
    "your-service/graph/model"
)

type Resolver struct{}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    // Get authenticated user from context
    authCtx, err := directives.GetAuthenticatedUser(ctx)
    if err != nil {
        return nil, fmt.Errorf("authentication required: %w", err)
    }
    
    // Fetch user from database using subject
    user, err := r.userService.GetBySubject(ctx, authCtx.Subject)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }
    
    return user, nil
}

func (r *queryResolver) Users(ctx context.Context, limit *int, offset *int) ([]*model.User, error) {
    // This resolver is protected by @hasRole directive in schema
    // No need for additional auth checks here
    
    l := 10
    if limit != nil {
        l = *limit
    }
    
    o := 0
    if offset != nil {
        o = *offset
    }
    
    return r.userService.ListUsers(ctx, l, o)
}

func (r *mutationResolver) UpdateMyProfile(ctx context.Context, input model.UpdateUserProfileInput) (*model.UserProfile, error) {
    authCtx := directives.MustGetAuthenticatedUser(ctx)
    
    return r.userService.UpdateProfile(ctx, authCtx.Subject, input)
}

func (r *queryResolver) SensitiveData(ctx context.Context) (string, error) {
    // This resolver requires both ADMIN role AND READ_SENSITIVE authority
    // due to @hasPermission directive
    authCtx := directives.MustGetAuthenticatedUser(ctx)
    
    return fmt.Sprintf("Sensitive data for user %s", authCtx.GetDisplayName()), nil
}
```

## REST API Service Integration

### 1. HTTP Server with Authentication Middleware

```go
package main

import (
    "encoding/json"
    "net/http"
    
    "github.com/erpmicroservices/auth-go/pkg/middleware"
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
)

func main() {
    // ... initialize auth components
    
    mux := http.NewServeMux()
    
    // Public endpoints
    mux.HandleFunc("/health", healthHandler)
    
    // Protected endpoints
    mux.Handle("/api/profile", authMiddleware.RequireAuth(
        http.HandlerFunc(profileHandler)))
    
    // Role-protected endpoints
    mux.Handle("/api/admin", authMiddleware.RequireAuth(
        rbacMiddleware.RequireRole("ADMIN")(
            http.HandlerFunc(adminHandler))))
    
    // Authority-protected endpoints
    mux.Handle("/api/service-data", authMiddleware.RequireAuth(
        rbacMiddleware.RequireAuthority("SERVICE")(
            http.HandlerFunc(serviceDataHandler))))
    
    // Combined role and authority protection
    mux.Handle("/api/sensitive", authMiddleware.RequireAuth(
        rbacMiddleware.RequireRoleOrAuthority(
            []string{"ADMIN", "DATA_ADMIN"}, 
            []string{"READ_SENSITIVE"})(
            http.HandlerFunc(sensitiveDataHandler))))
    
    http.ListenAndServe(":8080", mux)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    authCtx := middleware.MustGetAuthContext(r.Context())
    
    profile := map[string]interface{}{
        "subject":         authCtx.Subject,
        "name":           authCtx.Name,
        "email":          authCtx.Email,
        "roles":          authCtx.Roles,
        "authorities":    authCtx.Authorities,
        "organization_id": authCtx.OrganizationID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(profile)
}
```

### 2. Service Layer with Authorization Checks

```go
package service

import (
    "context"
    "fmt"
    
    "github.com/erpmicroservices/auth-go/pkg/middleware"
)

type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    authCtx, ok := middleware.GetAuthContext(ctx)
    if !ok {
        return nil, fmt.Errorf("authentication required")
    }
    
    // Check if user is accessing their own data or has admin privileges
    if authCtx.Subject != userID && !authCtx.HasRole("ADMIN") {
        return nil, fmt.Errorf("access denied")
    }
    
    return s.repo.GetByID(ctx, userID)
}

func (s *UserService) ListUsers(ctx context.Context) ([]*User, error) {
    if !middleware.HasAnyRole(ctx, "USER_MANAGER", "ADMIN") {
        return nil, fmt.Errorf("insufficient privileges")
    }
    
    return s.repo.ListAll(ctx)
}

func (s *UserService) UpdateSalary(ctx context.Context, employeeID string, salary float64) error {
    if !middleware.HasRole(ctx, "HR") {
        return fmt.Errorf("HR role required")
    }
    
    return s.repo.UpdateSalary(ctx, employeeID, salary)
}
```

## Database Service Integration

### 1. Securing Database Admin Endpoints

```go
package main

import (
    "net/http"
    "database/sql"
    
    "github.com/erpmicroservices/auth-go/pkg/middleware"
)

type DatabaseHandler struct {
    db *sql.DB
    authMiddleware *middleware.AuthMiddleware
    rbacMiddleware *middleware.RBACMiddleware
}

func (h *DatabaseHandler) SetupRoutes() *http.ServeMux {
    mux := http.NewServeMux()
    
    // Public endpoints
    mux.HandleFunc("/health", h.healthHandler)
    
    // Admin-only database operations
    mux.Handle("/admin/migrate", h.authMiddleware.RequireAuth(
        h.rbacMiddleware.RequireAuthority("DATABASE_ADMIN")(
            http.HandlerFunc(h.migrateHandler))))
    
    mux.Handle("/admin/backup", h.authMiddleware.RequireAuth(
        h.rbacMiddleware.RequireAuthority("DATABASE_ADMIN")(
            http.HandlerFunc(h.backupHandler))))
    
    // Service-to-service endpoints
    mux.Handle("/api/query", h.authMiddleware.RequireAuth(
        h.rbacMiddleware.RequireAuthority("SERVICE")(
            http.HandlerFunc(h.queryHandler))))
    
    return mux
}

func (h *DatabaseHandler) migrateHandler(w http.ResponseWriter, r *http.Request) {
    authCtx := middleware.MustGetAuthContext(r.Context())
    
    log.Info().
        Str("user", authCtx.Subject).
        Str("action", "database_migration").
        Msg("Executing database migration")
    
    // Perform migration
    err := h.runMigrations()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Migration completed successfully"))
}
```

## Frontend Integration

### 1. React Application with OAuth2 Flow

```typescript
// auth.service.ts
import axios from 'axios';

interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_at: string;
  refresh_token?: string;
}

export class AuthService {
  private readonly authServerUrl = 'http://localhost:9090';
  private readonly clientId = 'erp-microservices-client';
  private readonly redirectUri = 'http://localhost:3000/auth/callback';

  getAuthorizationUrl(): string {
    const state = this.generateState();
    localStorage.setItem('oauth_state', state);
    
    const params = new URLSearchParams({
      response_type: 'code',
      client_id: this.clientId,
      redirect_uri: this.redirectUri,
      scope: 'read write profile openid',
      state: state,
    });

    return `${this.authServerUrl}/oauth2/authorize?${params}`;
  }

  async exchangeCodeForToken(code: string, state: string): Promise<TokenResponse> {
    const storedState = localStorage.getItem('oauth_state');
    if (storedState !== state) {
      throw new Error('Invalid state parameter');
    }

    const response = await axios.post(`${this.authServerUrl}/oauth2/token`, {
      grant_type: 'authorization_code',
      client_id: this.clientId,
      code: code,
      redirect_uri: this.redirectUri,
    }, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    });

    return response.data;
  }

  private generateState(): string {
    return Math.random().toString(36).substring(7);
  }
}
```

### 2. React Component with Authentication

```tsx
// AuthenticatedApp.tsx
import React, { useEffect, useState } from 'react';
import { useQuery } from '@apollo/client';
import gql from 'graphql-tag';

const ME_QUERY = gql`
  query Me {
    me {
      id
      name
      email
      roles
      profile {
        picture
        biography
      }
    }
  }
`;

const USERS_QUERY = gql`
  query Users($limit: Int, $offset: Int) {
    users(limit: $limit, offset: $offset) {
      id
      name
      email
      status
    }
  }
`;

export const AuthenticatedApp: React.FC = () => {
  const { data: userData, loading: userLoading, error: userError } = useQuery(ME_QUERY);
  const { data: usersData, loading: usersLoading, error: usersError } = useQuery(USERS_QUERY, {
    variables: { limit: 10, offset: 0 },
    // This query will fail with INSUFFICIENT_PRIVILEGES if user doesn't have USER_MANAGER or ADMIN role
    errorPolicy: 'ignore',
  });

  if (userLoading) return <div>Loading...</div>;
  if (userError) return <div>Error: {userError.message}</div>;

  const user = userData?.me;
  const canViewUsers = user?.roles?.some((role: string) => 
    ['USER_MANAGER', 'ADMIN'].includes(role)
  );

  return (
    <div>
      <header>
        <h1>Welcome, {user?.name || user?.email}</h1>
        <p>Roles: {user?.roles?.join(', ')}</p>
      </header>

      <main>
        <section>
          <h2>My Profile</h2>
          <img src={user?.profile?.picture} alt="Profile" />
          <p>{user?.profile?.biography}</p>
        </section>

        {canViewUsers && (
          <section>
            <h2>User Management</h2>
            {usersLoading ? (
              <p>Loading users...</p>
            ) : usersError ? (
              <p>Error loading users: {usersError.message}</p>
            ) : (
              <ul>
                {usersData?.users?.map((user: any) => (
                  <li key={user.id}>
                    {user.name} ({user.email}) - {user.status}
                  </li>
                ))}
              </ul>
            )}
          </section>
        )}
      </main>
    </div>
  );
};
```

### 3. Apollo Client Setup with Auth Headers

```typescript
// apollo-client.ts
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

const httpLink = createHttpLink({
  uri: 'http://localhost:8080/graphql',
});

const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem('access_token');
  
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    }
  }
});

export const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache(),
  defaultOptions: {
    errorPolicy: 'all', // Handle auth errors gracefully
  },
});
```

## Service-to-Service Communication

### 1. Client Service Making Authenticated Requests

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
)

type OrderService struct {
    oauth2Client *oauth2.Client
    httpClient   *http.Client
    inventoryURL string
    paymentURL   string
}

func (s *OrderService) CheckInventory(ctx context.Context, productID string, quantity int) (bool, error) {
    // Get service-to-service token
    token, err := s.oauth2Client.GetClientCredentialsToken(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to get service token: %w", err)
    }
    
    // Make authenticated request to inventory service
    req, err := http.NewRequestWithContext(ctx, "POST", s.inventoryURL+"/api/check", nil)
    if err != nil {
        return false, err
    }
    
    req.Header.Set("Authorization", "Bearer "+token.AccessToken)
    req.Header.Set("Content-Type", "application/json")
    
    payload := map[string]interface{}{
        "product_id": productID,
        "quantity":   quantity,
    }
    
    payloadBytes, _ := json.Marshal(payload)
    req.Body = ioutil.NopCloser(bytes.NewReader(payloadBytes))
    
    resp, err := s.httpClient.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusUnauthorized {
        return false, fmt.Errorf("service authentication failed")
    }
    
    var result struct {
        Available bool `json:"available"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    return result.Available, err
}
```

### 2. Service Discovery and Load Balancing

```go
package client

import (
    "context"
    "fmt"
    "math/rand"
    "net/http"
    "time"
    
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
)

type ServiceClient struct {
    oauth2Client *oauth2.Client
    httpClient   *http.Client
    serviceURLs  map[string][]string // service name -> URLs
}

func (c *ServiceClient) CallService(ctx context.Context, serviceName, endpoint string, payload interface{}) (*http.Response, error) {
    // Get service URLs
    urls, exists := c.serviceURLs[serviceName]
    if !exists || len(urls) == 0 {
        return nil, fmt.Errorf("no URLs found for service %s", serviceName)
    }
    
    // Simple load balancing
    url := urls[rand.Intn(len(urls))]
    
    // Get service token with caching
    token, err := c.oauth2Client.GetClientCredentialsToken(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get service token: %w", err)
    }
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, "POST", url+endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    // Set authentication headers
    req.Header.Set("Authorization", "Bearer "+token.AccessToken)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "erp-service-client/1.0")
    
    // Add payload if provided
    if payload != nil {
        payloadBytes, _ := json.Marshal(payload)
        req.Body = ioutil.NopCloser(bytes.NewReader(payloadBytes))
    }
    
    return c.httpClient.Do(req)
}
```

### 3. Circuit Breaker with Authentication

```go
package client

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/erpmicroservices/auth-go/pkg/oauth2"
)

type CircuitBreakerClient struct {
    oauth2Client  *oauth2.Client
    client        *ServiceClient
    states        map[string]*CircuitState
    mutex         sync.RWMutex
}

type CircuitState struct {
    FailureCount    int
    LastFailureTime time.Time
    State          string // CLOSED, OPEN, HALF_OPEN
}

func (c *CircuitBreakerClient) CallWithCircuitBreaker(ctx context.Context, serviceName, endpoint string, payload interface{}) (*http.Response, error) {
    c.mutex.RLock()
    state, exists := c.states[serviceName]
    if !exists {
        state = &CircuitState{State: "CLOSED"}
        c.states[serviceName] = state
    }
    c.mutex.RUnlock()
    
    // Check circuit breaker state
    if state.State == "OPEN" {
        if time.Since(state.LastFailureTime) < 30*time.Second {
            return nil, fmt.Errorf("circuit breaker open for service %s", serviceName)
        }
        state.State = "HALF_OPEN"
    }
    
    // Make the call
    resp, err := c.client.CallService(ctx, serviceName, endpoint, payload)
    
    c.mutex.Lock()
    if err != nil || (resp != nil && resp.StatusCode >= 500) {
        state.FailureCount++
        state.LastFailureTime = time.Now()
        
        if state.FailureCount >= 3 {
            state.State = "OPEN"
        }
    } else {
        state.FailureCount = 0
        state.State = "CLOSED"
    }
    c.mutex.Unlock()
    
    return resp, err
}
```

## Error Handling and Monitoring

### 1. Authentication Error Handling

```go
package middleware

import (
    "encoding/json"
    "net/http"
    
    "github.com/rs/zerolog/log"
)

type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message"`
    Code    string                 `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (am *AuthMiddleware) handleAuthError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
    // Log the error
    log.Warn().
        Err(err).
        Str("path", r.URL.Path).
        Str("method", r.Method).
        Str("remote_addr", r.RemoteAddr).
        Str("user_agent", r.UserAgent()).
        Int("status_code", statusCode).
        Msg("Authentication error")
    
    // Determine error response based on error type
    var response ErrorResponse
    
    switch statusCode {
    case http.StatusUnauthorized:
        response = ErrorResponse{
            Error:   "unauthorized",
            Message: "Authentication required",
            Code:    "AUTH_REQUIRED",
        }
    case http.StatusForbidden:
        response = ErrorResponse{
            Error:   "forbidden",
            Message: "Insufficient privileges",
            Code:    "INSUFFICIENT_PRIVILEGES",
        }
    default:
        response = ErrorResponse{
            Error:   "authentication_error",
            Message: err.Error(),
        }
    }
    
    // Add debug information in development
    if am.config.DebugMode {
        response.Details = map[string]interface{}{
            "timestamp": time.Now().Unix(),
            "request_id": r.Header.Get("X-Request-ID"),
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("WWW-Authenticate", "Bearer")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

### 2. Metrics and Monitoring

```go
package monitoring

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
)

var (
    authRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_requests_total",
            Help: "Total number of authentication requests",
        },
        []string{"status", "method"},
    )
    
    authDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "auth_duration_seconds",
            Help: "Authentication request duration",
        },
        []string{"method"},
    )
    
    tokenCacheHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "token_cache_hits_total",
            Help: "Total number of token cache hits",
        },
        []string{"type"},
    )
)

func RecordAuthRequest(status, method string, duration time.Duration) {
    authRequestsTotal.WithLabelValues(status, method).Inc()
    authDuration.WithLabelValues(method).Observe(duration.Seconds())
}

func RecordCacheHit(cacheType string) {
    tokenCacheHits.WithLabelValues(cacheType).Inc()
}
```

This comprehensive integration guide shows how to use the auth-go module across different components of the ERP microservices system, from GraphQL APIs to service-to-service communication, with proper error handling and monitoring.