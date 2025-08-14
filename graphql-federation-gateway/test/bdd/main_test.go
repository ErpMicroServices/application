package bdd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
	"github.com/erpmicroservices/graphql-federation-gateway/internal/federation"
	"github.com/erpmicroservices/graphql-federation-gateway/pkg/gateway"
)

// TestSuite holds the test suite context
type TestSuite struct {
	config     *config.Config
	gateway    *gateway.Server
	federation *federation.Gateway
	httpClient *http.Client
	lastResponse *http.Response
	lastError  error
	context    context.Context
	cancel     context.CancelFunc
}

// NewTestSuite creates a new BDD test suite
func NewTestSuite() *TestSuite {
	return &TestSuite{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TestMain runs the BDD scenarios
func TestMain(m *testing.M) {
	// Setup logging for tests
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	status := godog.TestSuite{
		Name:                 "GraphQL Federation Gateway BDD Tests",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../graphql-federation-gateway-features"},
			TestingT: nil, // Not using testing.T for now
		},
	}.Run()

	os.Exit(status)
}

// InitializeTestSuite initializes the test suite
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		fmt.Println("🚀 Starting GraphQL Federation Gateway BDD Test Suite")
	})

	ctx.AfterSuite(func() {
		fmt.Println("✅ Completed GraphQL Federation Gateway BDD Test Suite")
	})
}

// InitializeScenario initializes each scenario with step definitions
func InitializeScenario(ctx *godog.ScenarioContext) {
	suite := NewTestSuite()

	// Setup and teardown hooks
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.context, suite.cancel = context.WithTimeout(context.Background(), 60*time.Second)
		return ctx, suite.setupScenario()
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		suite.teardownScenario()
		if suite.cancel != nil {
			suite.cancel()
		}
		return ctx, nil
	})

	// Register step definitions
	suite.registerStepDefinitions(ctx)
}

// setupScenario sets up the test scenario
func (ts *TestSuite) setupScenario() error {
	var err error

	// Load test configuration
	ts.config, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with test values
	ts.config.Environment = "test"
	ts.config.Server.Port = "4001" // Use different port for tests
	ts.config.Server.EnablePlayground = false

	// Initialize federation gateway
	ts.federation, err = federation.New(ts.config)
	if err != nil {
		return fmt.Errorf("failed to create federation gateway: %w", err)
	}

	// Create gateway server
	ts.gateway, err = gateway.NewServer(ts.config, ts.federation)
	if err != nil {
		return fmt.Errorf("failed to create gateway server: %w", err)
	}

	return nil
}

// teardownScenario cleans up after each scenario
func (ts *TestSuite) teardownScenario() {
	if ts.gateway != nil {
		// Shutdown gateway if it's running
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.gateway.Shutdown(ctx)
	}

	// Reset state
	ts.lastResponse = nil
	ts.lastError = nil
}

// registerStepDefinitions registers all step definitions for BDD scenarios
func (ts *TestSuite) registerStepDefinitions(ctx *godog.ScenarioContext) {
	// Gateway setup steps
	ctx.Step(`^the gateway is running$`, ts.theGatewayIsRunning)
	ctx.Step(`^the gateway is running with federated services$`, ts.theGatewayIsRunningWithFederatedServices)

	// Service availability steps
	ctx.Step(`^the following services are available:$`, ts.theFollowingServicesAreAvailable)
	ctx.Step(`^each service exposes a valid GraphQL schema$`, ts.eachServiceExposesValidGraphQLSchema)

	// Schema composition steps
	ctx.Step(`^the gateway starts up$`, ts.theGatewayStartsUp)
	ctx.Step(`^the gateway should successfully compose a unified schema$`, ts.theGatewayShouldComposeUnifiedSchema)
	ctx.Step(`^the unified schema should include entities from all services$`, ts.theUnifiedSchemaShouldIncludeEntities)
	ctx.Step(`^the schema should be valid according to Apollo Federation v2 rules$`, ts.theSchemaShouldBeValidFederationV2)

	// Query execution steps
	ctx.Step(`^I execute a query for people data only:$`, ts.iExecuteQueryForPeopleDataOnly)
	ctx.Step(`^I execute a cross-service query:$`, ts.iExecuteCrossServiceQuery)
	ctx.Step(`^the query should be routed to the (.+) service$`, ts.theQueryShouldBeRoutedToService)
	ctx.Step(`^I should receive a successful response with people data$`, ts.iShouldReceiveSuccessfulResponseWithPeopleData)

	// Authentication steps
	ctx.Step(`^I have a valid JWT token with user role$`, ts.iHaveValidJWTTokenWithUserRole)
	ctx.Step(`^I make a request with the Authorization header:$`, ts.iMakeRequestWithAuthorizationHeader)
	ctx.Step(`^the token should be validated successfully$`, ts.theTokenShouldBeValidatedSuccessfully)
	ctx.Step(`^the request should be processed$`, ts.theRequestShouldBeProcessed)

	// Error handling steps
	ctx.Step(`^the (.+) service is unavailable$`, ts.theServiceIsUnavailable)
	ctx.Step(`^the gateway should return partial data for available services$`, ts.theGatewayShouldReturnPartialData)

	// General response steps
	ctx.Step(`^I should receive a successful response$`, ts.iShouldReceiveSuccessfulResponse)
	ctx.Step(`^I should receive an error$`, ts.iShouldReceiveError)
	ctx.Step(`^the response should contain (.+)$`, ts.theResponseShouldContain)

	// Configuration steps
	ctx.Step(`^the unified schema is available$`, ts.theUnifiedSchemaIsAvailable)
	ctx.Step(`^authentication is configured to use auth-go module$`, ts.authenticationIsConfigured)
	ctx.Step(`^JWT validation is enabled$`, ts.jwtValidationIsEnabled)

	// WebSocket subscription steps
	ctx.Step(`^the gateway is running with subscription support$`, ts.theGatewayIsRunningWithSubscriptionSupport)
	ctx.Step(`^WebSocket connections are enabled$`, ts.webSocketConnectionsAreEnabled)

	// Monitoring steps  
	ctx.Step(`^the gateway is running with monitoring enabled$`, ts.theGatewayIsRunningWithMonitoringEnabled)
	ctx.Step(`^metrics collection is configured$`, ts.metricsCollectionIsConfigured)
}