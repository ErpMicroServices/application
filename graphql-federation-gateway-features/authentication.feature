Feature: Authentication and Authorization
  As a secure GraphQL gateway
  I want to validate user authentication and authorize requests
  So that only authorized users can access protected resources

  Background:
    Given the gateway is running
    And authentication is configured to use auth-go module
    And JWT validation is enabled
    And the following protected resources exist:
      | resource      | required_role |
      | person.email  | user          |
      | person.orders | user          |
      | admin_queries | admin         |

  Scenario: Successful authentication with valid JWT token
    Given I have a valid JWT token with user role
    When I make a request with the Authorization header:
      """
      Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
      """
    And I execute a query:
      """
      query {
        person(id: "123") {
          id
          name
        }
      }
      """
    Then the token should be validated successfully
    And the request should be processed
    And I should receive the person data

  Scenario: Reject request with invalid JWT token
    Given I have an invalid JWT token
    When I make a request with the Authorization header:
      """
      Authorization: Bearer invalid.token.here
      """
    And I execute a query:
      """
      query {
        person(id: "123") {
          name
        }
      }
      """
    Then the token validation should fail
    And the request should be rejected with status 401
    And I should receive an authentication error:
      """
      {
        "errors": [{
          "message": "Invalid authentication token",
          "extensions": {
            "code": "UNAUTHENTICATED"
          }
        }]
      }
      """

  Scenario: Reject request without authentication token
    Given I do not provide an Authorization header
    When I execute a query requiring authentication:
      """
      query {
        person(id: "123") {
          email
        }
      }
      """
    Then the request should be rejected with status 401
    And I should receive an authentication error:
      """
      {
        "errors": [{
          "message": "Authentication required",
          "extensions": {
            "code": "UNAUTHENTICATED"
          }
        }]
      }
      """

  Scenario: Authorize access to protected fields
    Given I have a valid JWT token with user role
    When I execute a query requesting protected fields:
      """
      query {
        person(id: "123") {
          id
          name
          email
          orders {
            id
            total
          }
        }
      }
      """
    Then the auth-go module should validate the token
    And the user role should be extracted from the token
    And access to email and orders fields should be granted
    And I should receive the complete person data with protected fields

  Scenario: Deny access to fields above user's permission level
    Given I have a valid JWT token with user role
    When I execute a query requesting admin-only data:
      """
      query {
        adminStats {
          totalUsers
          revenue
        }
      }
      """
    Then the authorization should fail
    And I should receive an authorization error:
      """
      {
        "errors": [{
          "message": "Insufficient permissions for field 'adminStats'",
          "extensions": {
            "code": "FORBIDDEN"
          }
        }]
      }
      """

  Scenario: Handle expired JWT token
    Given I have an expired JWT token
    When I make a request with the expired token:
      """
      Authorization: Bearer expired.jwt.token
      """
    And I execute a query:
      """
      query {
        person(id: "123") {
          name
        }
      }
      """
    Then the token validation should fail due to expiration
    And I should receive an authentication error indicating token expiration:
      """
      {
        "errors": [{
          "message": "Authentication token has expired",
          "extensions": {
            "code": "UNAUTHENTICATED",
            "reason": "TOKEN_EXPIRED"
          }
        }]
      }
      """

  Scenario: Context-based authorization with business rules
    Given I have a valid JWT token for user "john.doe@example.com"
    When I execute a query to access another user's private data:
      """
      query {
        person(id: "456") {
          email
          orders {
            id
          }
        }
      }
      """
    Then the business context authorization should be evaluated
    And access should be denied if the user is not the data owner or admin
    And I should receive an authorization error:
      """
      {
        "errors": [{
          "message": "Access denied: cannot view another user's private data",
          "extensions": {
            "code": "FORBIDDEN"
          }
        }]
      }
      """