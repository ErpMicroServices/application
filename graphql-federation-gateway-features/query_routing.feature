Feature: Query Routing and Planning
  As a GraphQL gateway
  I want to efficiently route queries to the appropriate services
  So that I can minimize latency and resource usage

  Background:
    Given the gateway is running with federated services
    And query planning is enabled
    And query complexity analysis is configured

  Scenario: Route simple query to single service
    Given I have a query that targets only one service:
      """
      query GetPersonDetails($id: ID!) {
        person(id: $id) {
          id
          name
          email
          phoneNumber
        }
      }
      """
    When the query planner analyzes the query
    Then the query should be identified as single-service
    And the query should be routed directly to people_and_organizations service
    And no additional query planning overhead should be incurred

  Scenario: Plan complex federated query
    Given I have a complex federated query:
      """
      query GetPersonWithOrderHistory($personId: ID!, $limit: Int!) {
        person(id: $personId) {
          id
          name
          email
          orders(limit: $limit) {
            id
            date
            status
            total
            items {
              quantity
              price
              product {
                id
                name
                category {
                  name
                }
              }
            }
          }
        }
      }
      """
    When the query planner analyzes the query
    Then the planner should create an execution plan with the following steps:
      | step | service                    | query                        | depends_on |
      | 1    | people_and_organizations   | Get person by ID             | none       |
      | 2    | e_commerce                 | Get orders for person        | step 1     |
      | 3    | products                   | Get product details for items| step 2     |
    And the execution plan should optimize for parallel execution where possible
    And the plan should include proper error handling for each step

  Scenario: Handle query complexity limits
    Given query complexity analysis is configured with a limit of 1000
    When I execute a query with complexity score of 1500:
      """
      query VeryComplexQuery {
        people {
          orders {
            items {
              product {
                reviews {
                  author {
                    orders {
                      items {
                        product {
                          name
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
      """
    Then the query should be rejected before execution
    And I should receive an error indicating complexity limit exceeded
    And the error should include the actual complexity score
    And no downstream services should be queried

  Scenario: Optimize query with field selection
    Given I have a query with specific field selection:
      """
      query OptimizedQuery($personId: ID!) {
        person(id: $personId) {
          name
          orders {
            id
            status
          }
        }
      }
      """
    When the query planner creates the execution plan
    Then the sub-queries should only request the needed fields
    And the people service query should only request "id" and "name" fields
    And the e_commerce service query should only request "id" and "status" fields
    And unnecessary data transfer should be minimized

  Scenario: Handle query timeout
    Given query timeout is configured to 30 seconds
    And the products service has a 45-second response delay
    When I execute a query that requires data from the products service:
      """
      query TimeoutTest($personId: ID!) {
        person(id: $personId) {
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
    Then the query should timeout after 30 seconds
    And I should receive a timeout error
    And partial results from faster services should be included if available
    And the slow service should be marked as degraded in monitoring