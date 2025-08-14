Feature: Real-time Subscriptions
  As a client application
  I want to receive real-time updates through GraphQL subscriptions
  So that I can provide live data to users

  Background:
    Given the gateway is running with subscription support
    And WebSocket connections are enabled
    And the following services support subscriptions:
      | service              | subscription_endpoint               |
      | people_and_organizations | ws://people-service:8080/graphql    |
      | e_commerce           | ws://ecommerce-service:8080/graphql |

  Scenario: Establish WebSocket connection for subscriptions
    Given I am a client application
    When I connect to the gateway's WebSocket endpoint at "ws://gateway:8080/graphql"
    Then the connection should be established successfully
    And I should receive a connection acknowledgment message
    And the connection should be ready to handle subscription operations

  Scenario: Subscribe to single service updates
    Given I have an active WebSocket connection
    When I send a subscription operation:
      """
      subscription {
        personUpdated(id: "123") {
          id
          name
          email
          updatedAt
        }
      }
      """
    Then the gateway should establish a subscription with the people_and_organizations service
    And I should receive a subscription acknowledgment
    When the person with ID "123" is updated in the people service
    Then I should receive a real-time update message:
      """
      {
        "data": {
          "personUpdated": {
            "id": "123",
            "name": "John Doe Updated",
            "email": "john.doe.updated@example.com",
            "updatedAt": "2024-01-15T10:30:00Z"
          }
        }
      }
      """

  Scenario: Subscribe to federated entity updates
    Given I have an active WebSocket connection
    When I send a subscription for federated data:
      """
      subscription {
        orderUpdated(personId: "123") {
          id
          status
          person {
            id
            name
          }
          items {
            product {
              id
              name
            }
          }
        }
      }
      """
    Then the gateway should establish subscriptions with multiple services
    And updates from any related service should trigger the subscription
    When an order is updated in the e_commerce service
    Then the gateway should fetch related person data from people service
    And I should receive a unified subscription update with all federated data

  Scenario: Handle subscription authentication
    Given I have a WebSocket connection with authentication
    And I provide a valid JWT token during connection:
      """
      {
        "type": "connection_init",
        "payload": {
          "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        }
      }
      """
    When I subscribe to protected data:
      """
      subscription {
        userOrderUpdates {
          id
          total
          items {
            product {
              name
            }
          }
        }
      }
      """
    Then the subscription should be authenticated successfully
    And I should only receive updates for orders belonging to the authenticated user

  Scenario: Reject unauthenticated subscriptions to protected resources
    Given I have a WebSocket connection without authentication
    When I attempt to subscribe to protected data:
      """
      subscription {
        userOrderUpdates {
          id
          total
        }
      }
      """
    Then the subscription should be rejected
    And I should receive an authentication error:
      """
      {
        "type": "error",
        "payload": {
          "message": "Authentication required for subscription",
          "extensions": {
            "code": "UNAUTHENTICATED"
          }
        }
      }
      """

  Scenario: Handle service disconnection gracefully
    Given I have an active subscription to order updates
    And the e_commerce service becomes unavailable
    When an order update should be triggered
    Then the gateway should detect the service disconnection
    And I should receive a service unavailable error:
      """
      {
        "type": "error",
        "payload": {
          "message": "Subscription service temporarily unavailable",
          "extensions": {
            "code": "SERVICE_UNAVAILABLE",
            "service": "e_commerce"
          }
        }
      }
      """
    And the subscription should remain active for automatic reconnection
    When the e_commerce service becomes available again
    Then the subscription should automatically resume
    And I should receive pending updates that occurred during disconnection

  Scenario: Handle subscription multiplexing
    Given I have multiple active subscriptions on the same WebSocket connection:
      | subscription_id | operation |
      | sub1           | personUpdated(id: "123") |
      | sub2           | orderUpdated(personId: "123") |
      | sub3           | productUpdated(id: "456") |
    When updates occur for different subscriptions simultaneously
    Then each update should be delivered with the correct subscription ID
    And the updates should not interfere with each other
    And the client should be able to distinguish between different subscription responses

  Scenario: Clean up subscriptions on disconnect
    Given I have multiple active subscriptions
    When my WebSocket connection is closed unexpectedly
    Then the gateway should detect the disconnection
    And all subscriptions associated with that connection should be cleaned up
    And the gateway should notify downstream services to stop sending updates
    And no memory leaks should occur from abandoned subscriptions