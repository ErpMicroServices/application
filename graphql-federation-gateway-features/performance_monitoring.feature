Feature: Performance Monitoring and Observability
  As a system operator
  I want to monitor the performance and health of the GraphQL gateway
  So that I can maintain optimal system performance and troubleshoot issues

  Background:
    Given the gateway is running with monitoring enabled
    And metrics collection is configured
    And the following monitoring endpoints are available:
      | endpoint           | purpose                    |
      | /health           | Health check               |
      | /metrics          | Prometheus metrics         |
      | /debug/pprof      | Go profiling data          |

  Scenario: Collect query performance metrics
    Given query performance monitoring is enabled
    When I execute various GraphQL queries:
      | query_name        | complexity | execution_time |
      | simple_person     | 50         | 45ms          |
      | federated_orders  | 200        | 150ms         |
      | complex_analytics | 800        | 450ms         |
    Then the gateway should record the following metrics:
      | metric_name                    | value |
      | graphql_query_duration_seconds | varied by query |
      | graphql_query_complexity_score | varied by query |
      | graphql_requests_total         | 3     |
    And the metrics should be available at the /metrics endpoint
    And the metrics should include query labels for analysis

  Scenario: Monitor service health and availability
    Given health monitoring is configured for all federated services
    When the health check endpoint is queried
    Then I should receive a health status response:
      """
      {
        "status": "healthy",
        "services": {
          "people_and_organizations": {
            "status": "healthy",
            "response_time": "12ms",
            "last_check": "2024-01-15T10:30:00Z"
          },
          "e_commerce": {
            "status": "healthy", 
            "response_time": "8ms",
            "last_check": "2024-01-15T10:30:00Z"
          },
          "products": {
            "status": "degraded",
            "response_time": "2500ms",
            "last_check": "2024-01-15T10:30:00Z"
          }
        }
      }
      """
    And the overall status should reflect the worst service status

  Scenario: Track query resolution performance
    Given detailed query tracing is enabled
    When I execute a complex federated query:
      """
      query ComplexQuery($personId: ID!) {
        person(id: $personId) {
          name
          orders {
            id
            items {
              product {
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
    Then the gateway should record execution traces with:
      | phase                    | duration | service |
      | query_parsing           | 2ms      | gateway |
      | query_validation        | 3ms      | gateway |
      | query_planning          | 5ms      | gateway |
      | person_fetch            | 45ms     | people  |
      | orders_fetch            | 67ms     | e_commerce |
      | products_fetch          | 123ms    | products |
      | result_composition      | 8ms      | gateway |
    And the trace should include service call details
    And the trace should be available for performance analysis

  Scenario: Monitor memory usage and garbage collection
    Given memory monitoring is enabled
    When the gateway processes queries continuously for 10 minutes
    Then memory usage metrics should be recorded:
      | metric                     | description |
      | go_memstats_heap_inuse_bytes | Heap memory in use |
      | go_memstats_gc_duration_seconds | GC pause duration |
      | go_goroutines              | Number of goroutines |
    And memory growth should be monitored for leaks
    And GC pressure should be within acceptable limits

  Scenario: Alert on performance degradation
    Given alerting thresholds are configured:
      | metric                        | threshold |
      | avg_query_duration           | 500ms     |
      | error_rate                   | 5%        |
      | service_availability         | 95%       |
    When average query duration exceeds 500ms for 5 minutes
    Then an alert should be generated:
      """
      {
        "alert": "high_query_latency",
        "message": "Average query duration exceeded threshold",
        "current_value": "642ms",
        "threshold": "500ms",
        "duration": "5 minutes"
      }
      """
    And the alert should be sent to monitoring system

  Scenario: Profile resource usage under load
    Given profiling is enabled
    When the gateway is under high load (1000 concurrent requests)
    Then CPU profiling data should be available at /debug/pprof/profile
    And memory profiling data should be available at /debug/pprof/heap
    And goroutine profiles should show no goroutine leaks
    And the profiling data should help identify performance bottlenecks

  Scenario: Monitor subscription connection health
    Given subscription monitoring is enabled
    And I have 100 active WebSocket connections with subscriptions
    When monitoring data is collected
    Then the following subscription metrics should be recorded:
      | metric                          | value |
      | websocket_connections_active    | 100   |
      | subscription_events_sent_total  | varied |
      | subscription_errors_total       | minimal |
      | websocket_connection_duration   | varied |
    And connection lifecycle events should be tracked
    And subscription performance should be monitored per service

  Scenario: Track query complexity and prevent abuse
    Given query complexity analysis is enabled with limits
    When queries with varying complexity are executed:
      | query_type          | complexity_score |
      | simple_lookup      | 25              |
      | moderate_join      | 150             |
      | complex_analytics  | 400             |
      | abusive_deep_query | 1200            |
    Then complexity metrics should be recorded
    And queries exceeding complexity limits should be blocked
    And complexity trends should be available for analysis
    And potential abuse patterns should be detectable

  Scenario: Correlate errors with performance impacts
    Given error correlation tracking is enabled
    When service errors occur during query execution
    Then errors should be correlated with performance metrics:
      | error_type           | performance_impact |
      | timeout_error       | increased_latency  |
      | service_unavailable | partial_results    |
      | validation_error    | no_impact         |
    And error patterns should be analyzable
    And the correlation data should help with root cause analysis