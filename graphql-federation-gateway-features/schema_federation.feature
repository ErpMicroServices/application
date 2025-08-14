Feature: Schema Federation
  As a GraphQL client
  I want to access a unified schema composed from multiple services
  So that I can query data from different domains seamlessly

  Background:
    Given the gateway is running
    And the following services are available:
      | service                          | url                                    |
      | people_and_organizations         | http://people-service:8080/graphql     |
      | e_commerce                       | http://ecommerce-service:8080/graphql  |
      | products                         | http://products-service:8080/graphql   |

  Scenario: Gateway composes schemas from multiple services
    Given each service exposes a valid GraphQL schema
    When the gateway starts up
    Then the gateway should successfully compose a unified schema
    And the unified schema should include entities from all services
    And the schema should be valid according to Apollo Federation v2 rules

  Scenario: Query single service through gateway
    Given the unified schema is available
    When I execute a query for people data only:
      """
      query {
        people {
          id
          name
          email
        }
      }
      """
    Then the query should be routed to the people_and_organizations service
    And I should receive a successful response with people data

  Scenario: Query federated entities across services
    Given the unified schema includes federated entities
    When I execute a cross-service query:
      """
      query {
        person(id: "123") {
          id
          name
          orders {
            id
            status
            items {
              product {
                name
                price
              }
            }
          }
        }
      }
      """
    Then the gateway should create an execution plan across multiple services
    And the person data should be fetched from people_and_organizations service
    And the orders data should be fetched from e_commerce service
    And the product data should be fetched from products service
    And I should receive a unified response with all data composed correctly

  Scenario: Handle service unavailability gracefully
    Given the unified schema is available
    And the products service is unavailable
    When I execute a query that includes product data:
      """
      query {
        person(id: "123") {
          name
          orders {
            items {
              product {
                name
              }
            }
          }
        }
      }
      """
    Then the gateway should return partial data for available services
    And the response should include an error indicating the products service is unavailable
    And the person and order data should still be returned successfully

  Scenario: Validate schema composition on service changes
    Given the gateway has composed schemas from all services
    When a service updates its schema
    And the service becomes available with the new schema
    Then the gateway should detect the schema change
    And the gateway should recompose the unified schema
    And the new unified schema should be valid
    And clients should be notified of schema updates if subscribed