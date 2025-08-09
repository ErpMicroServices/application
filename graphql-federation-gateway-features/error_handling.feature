Feature: Error Handling and Resilience
  As a robust GraphQL gateway
  I want to handle errors gracefully and maintain system resilience
  So that clients receive meaningful error responses and the system remains stable

  Background:
    Given the gateway is running
    And error handling middleware is configured
    And circuit breaker patterns are enabled

  Scenario: Handle GraphQL syntax errors
    Given the gateway is receiving requests
    When I send a query with invalid GraphQL syntax:
      """
      query {
        person(id: "123" {
          name
        }
      }
      """
    Then the query should be rejected during parsing
    And I should receive a syntax error response:
      """
      {
        "errors": [{
          "message": "Syntax Error: Expected ')', found '{'",
          "locations": [{"line": 2, "column": 23}],
          "extensions": {
            "code": "GRAPHQL_PARSE_FAILED"
          }
        }]
      }
      """
    And no downstream services should be contacted

  Scenario: Handle validation errors
    Given the unified schema is available
    When I send a query requesting non-existent fields:
      """
      query {
        person(id: "123") {
          id
          nonExistentField
        }
      }
      """
    Then the query should fail validation
    And I should receive a validation error:
      """
      {
        "errors": [{
          "message": "Cannot query field 'nonExistentField' on type 'Person'",
          "locations": [{"line": 4, "column": 11}],
          "extensions": {
            "code": "GRAPHQL_VALIDATION_FAILED"
          }
        }]
      }
      """

  Scenario: Handle downstream service errors gracefully
    Given the people service returns an internal server error
    When I execute a query targeting the people service:
      """
      query {
        person(id: "123") {
          id
          name
        }
      }
      """
    Then the gateway should catch the service error
    And I should receive a structured error response:
      """
      {
        "data": null,
        "errors": [{
          "message": "Internal error occurred in people service",
          "path": ["person"],
          "extensions": {
            "code": "DOWNSTREAM_ERROR",
            "service": "people_and_organizations",
            "timestamp": "2024-01-15T10:30:00Z"
          }
        }]
      }
      """
    And the error should be logged for monitoring

  Scenario: Return partial results when some services fail
    Given I have a federated query requiring multiple services
    And the products service is returning errors
    When I execute the federated query:
      """
      query {
        person(id: "123") {
          id
          name
          orders {
            id
            items {
              quantity
              product {
                id
                name
              }
            }
          }
        }
      }
      """
    Then the gateway should return successful data from available services
    And the response should include partial results and errors:
      """
      {
        "data": {
          "person": {
            "id": "123",
            "name": "John Doe",
            "orders": [{
              "id": "order-456",
              "items": [{
                "quantity": 2,
                "product": null
              }]
            }]
          }
        },
        "errors": [{
          "message": "Failed to fetch product data",
          "path": ["person", "orders", 0, "items", 0, "product"],
          "extensions": {
            "code": "DOWNSTREAM_ERROR",
            "service": "products"
          }
        }]
      }
      """

  Scenario: Implement circuit breaker for failing services
    Given the products service has been failing consistently
    And the circuit breaker threshold is set to 5 failures in 60 seconds
    When the products service fails for the 5th time within the threshold period
    Then the circuit breaker should trip to OPEN state
    And subsequent requests to products service should be blocked immediately
    And I should receive a circuit breaker error:
      """
      {
        "data": null,
        "errors": [{
          "message": "Service temporarily unavailable due to repeated failures",
          "extensions": {
            "code": "CIRCUIT_BREAKER_OPEN",
            "service": "products"
          }
        }]
      }
      """

  Scenario: Circuit breaker recovery
    Given the circuit breaker for products service is in OPEN state
    And the recovery timeout period has elapsed
    When the circuit breaker transitions to HALF_OPEN state
    And a test request to products service succeeds
    Then the circuit breaker should transition to CLOSED state
    And normal request processing should resume
    And subsequent requests should be processed normally

  Scenario: Handle timeout errors
    Given query timeout is configured to 10 seconds
    And the e_commerce service has a response delay of 15 seconds
    When I execute a query requiring e_commerce service:
      """
      query {
        person(id: "123") {
          orders {
            id
          }
        }
      }
      """
    Then the request should timeout after 10 seconds
    And I should receive a timeout error:
      """
      {
        "data": {
          "person": {
            "orders": null
          }
        },
        "errors": [{
          "message": "Request timeout: service did not respond within 10 seconds",
          "path": ["person", "orders"],
          "extensions": {
            "code": "TIMEOUT",
            "service": "e_commerce",
            "timeout": 10000
          }
        }]
      }
      """

  Scenario: Handle rate limiting errors
    Given rate limiting is configured to 100 requests per minute per user
    And I have exceeded the rate limit
    When I attempt to make another request
    Then the request should be rejected
    And I should receive a rate limit error:
      """
      {
        "errors": [{
          "message": "Rate limit exceeded: too many requests",
          "extensions": {
            "code": "RATE_LIMITED",
            "limit": 100,
            "window": "1 minute",
            "retryAfter": 45
          }
        }]
      }
      """

  Scenario: Sanitize sensitive information in errors
    Given error sanitization is enabled
    When a downstream service returns an error containing sensitive information:
      """
      Database connection failed: password='secret123' host='internal-db'
      """
    Then the gateway should sanitize the error message
    And I should receive a generic error response:
      """
      {
        "errors": [{
          "message": "Internal service error occurred",
          "extensions": {
            "code": "INTERNAL_ERROR",
            "timestamp": "2024-01-15T10:30:00Z"
          }
        }]
      }
      """
    And the detailed error should be logged securely for debugging