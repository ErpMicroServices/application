# GraphQL Federation Gateway Features

This repository contains BDD (Behavior-Driven Development) scenarios for the GraphQL Federation Gateway written in Go.

## Overview

The features define the expected behavior of the GraphQL federation gateway that orchestrates multiple ERP microservices into a unified GraphQL API.

## Feature Files

### Core Federation Features
- **`schema_federation.feature`** - Schema composition and federation across services
- **`query_routing.feature`** - Query planning, routing, and optimization
- **`error_handling.feature`** - Error handling, resilience, and circuit breakers

### Security & Real-time Features  
- **`authentication.feature`** - JWT authentication and role-based authorization
- **`subscriptions.feature`** - WebSocket subscriptions and real-time updates

### Monitoring & Performance
- **`performance_monitoring.feature`** - Observability, metrics, and performance tracking

## Services Under Test

The gateway federates the following ERP microservices:

| Service | GraphQL Endpoint | Domain |
|---------|-----------------|---------|
| people_and_organizations | http://people-service:8080/graphql | People, contacts, relationships |
| e_commerce | http://ecommerce-service:8080/graphql | Orders, shopping, user preferences |
| products | http://products-service:8080/graphql | Product catalog, inventory |
| accounting_and_budgeting | http://accounting-service:8080/graphql | Financial data, budgets |
| human_resources | http://hr-service:8080/graphql | Employee management |
| work_effort | http://work-effort-service:8080/graphql | Projects, tasks, time tracking |

## Testing Approach

### BDD Methodology
1. **Scenarios First**: All features are defined as Gherkin scenarios before implementation
2. **Behavior Specification**: Each scenario describes expected system behavior from user perspective  
3. **Acceptance Criteria**: Scenarios serve as acceptance criteria for implementation

### Integration with Go Implementation
- Feature files drive the creation of Go test suites
- Step definitions will be implemented using Go BDD frameworks (like Godog)
- Tests will run against the actual Go implementation

## Scenario Categories

### Happy Path Scenarios
- Successful schema federation
- Normal query execution 
- Valid authentication flows
- Working subscriptions

### Error Handling Scenarios  
- Service unavailability
- Invalid queries and tokens
- Timeout conditions
- Circuit breaker activation

### Performance Scenarios
- Query complexity limits
- Response time requirements
- Memory usage monitoring
- Load testing conditions

### Security Scenarios
- Authentication validation
- Authorization enforcement
- Token expiration handling
- Business context rules

## Usage

These feature files will be used to:

1. **Drive Implementation** - Guide the Go implementation to ensure all requirements are met
2. **Acceptance Testing** - Validate that the implementation meets business requirements
3. **Regression Testing** - Ensure changes don't break existing functionality
4. **Documentation** - Serve as living documentation of system behavior

## Dependencies

The test scenarios assume:
- Apollo Federation v2 compatibility
- JWT authentication using auth-go module
- Integration with common-go utilities
- WebSocket support for subscriptions
- Prometheus metrics collection
- Circuit breaker patterns for resilience

## Running Tests

(Will be implemented with Go test runners)

```bash
# Run all BDD scenarios
make test-bdd

# Run specific feature
make test-feature FEATURE=schema_federation

# Run with coverage
make test-bdd-coverage
```

## Contributing

When adding new scenarios:

1. Follow Gherkin syntax strictly
2. Use realistic test data
3. Include both positive and negative test cases  
4. Consider edge cases and error conditions
5. Ensure scenarios are independent and can run in any order