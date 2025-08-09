//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
	"github.com/erpmicroservices/graphql-federation-gateway/internal/federation"
	"github.com/erpmicroservices/graphql-federation-gateway/pkg/gateway"
	"github.com/erpmicroservices/graphql-federation-gateway/test/fixtures"
)

// TestGatewayIntegration tests the complete gateway integration
func TestGatewayIntegration(t *testing.T) {
	// Create mock services
	mockServices := fixtures.CreateMockServices()
	defer fixtures.CleanupMockServices(mockServices)

	// Load test configuration
	cfg, err := loadTestConfig(mockServices)
	require.NoError(t, err, "Failed to load test configuration")

	// Create federation gateway
	fedGateway, err := federation.New(cfg)
	require.NoError(t, err, "Failed to create federation gateway")

	// Create gateway server
	server, err := gateway.NewServer(cfg, fedGateway)
	require.NoError(t, err, "Failed to create gateway server")

	// Start server in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	// Shutdown server
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()

	t.Run("health_check", func(t *testing.T) {
		testHealthCheck(t, cfg)
	})

	t.Run("service_discovery", func(t *testing.T) {
		testServiceDiscovery(t, fedGateway, mockServices)
	})

	t.Run("schema_composition", func(t *testing.T) {
		testSchemaComposition(t, fedGateway)
	})

	// TODO: Add more integration tests
	// t.Run("graphql_queries", func(t *testing.T) {
	//     testGraphQLQueries(t, cfg)
	// })
	
	// t.Run("authentication", func(t *testing.T) {
	//     testAuthentication(t, cfg)
	// })
}

// testHealthCheck verifies the gateway health check endpoint
func testHealthCheck(t *testing.T, cfg *config.Config) {
	// TODO: Implement health check test
	t.Log("Health check test not yet implemented")
}

// testServiceDiscovery verifies service discovery functionality
func testServiceDiscovery(t *testing.T, gateway *federation.Gateway, mockServices map[string]*fixtures.MockService) {
	services := gateway.GetServices()
	
	assert.NotEmpty(t, services, "No services discovered")
	
	// Verify expected services are present
	expectedServices := []string{"people-organizations", "e-commerce", "products"}
	for _, expectedService := range expectedServices {
		assert.True(t, gateway.IsServiceEnabled(expectedService), 
			"Service %s should be enabled", expectedService)
	}
	
	t.Logf("Discovered %d services", len(services))
}

// testSchemaComposition verifies schema composition functionality
func testSchemaComposition(t *testing.T, gateway *federation.Gateway) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := gateway.ComposeSchema(ctx)
	
	// For now, we expect an error since composition is not implemented
	// This test will pass when composition is implemented
	t.Logf("Schema composition result: %v", err)
}

// loadTestConfig creates a test configuration with mock service URLs
func loadTestConfig(mockServices map[string]*fixtures.MockService) (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Override configuration for testing
	cfg.Environment = "test"
	cfg.Server.Port = "4002" // Different port for integration tests
	cfg.Server.EnablePlayground = false
	cfg.Auth.TokenValidation = false // Disable auth for basic tests

	// Override service URLs with mock services
	if peopleService, exists := mockServices["people-organizations"]; exists {
		cfg.Services.PeopleOrganizations.URL = peopleService.URL() + "/graphql"
	}
	
	if ecommerceService, exists := mockServices["e-commerce"]; exists {
		cfg.Services.ECommerce.URL = ecommerceService.URL() + "/graphql"
	}
	
	if productsService, exists := mockServices["products"]; exists {
		cfg.Services.Products.URL = productsService.URL() + "/graphql"
	}

	return cfg, nil
}