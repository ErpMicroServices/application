package federation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedError bool
		assertions    func(t *testing.T, gateway *Gateway)
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Services: config.ServicesConfig{
					PeopleOrganizations: config.ServiceConfig{
						Name:    "people-organizations",
						URL:     "http://localhost:8081/graphql",
						Enabled: true,
					},
					ECommerce: config.ServiceConfig{
						Name:    "e-commerce",
						URL:     "http://localhost:8082/graphql",
						Enabled: true,
					},
				},
			},
			expectedError: false,
			assertions: func(t *testing.T, gateway *Gateway) {
				assert.NotNil(t, gateway)
				assert.NotNil(t, gateway.services)
				assert.Len(t, gateway.services, 8) // All enabled services from GetEnabledServices
			},
		},
		{
			name: "empty configuration",
			config: &config.Config{
				Services: config.ServicesConfig{},
			},
			expectedError: false,
			assertions: func(t *testing.T, gateway *Gateway) {
				assert.NotNil(t, gateway)
				assert.NotNil(t, gateway.services)
				assert.Empty(t, gateway.services)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway, err := New(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, gateway)
			} else {
				require.NoError(t, err)
				require.NotNil(t, gateway)
				if tt.assertions != nil {
					tt.assertions(t, gateway)
				}
			}
		})
	}
}

func TestServiceClientInitialization(t *testing.T) {
	cfg := &config.Config{
		Services: config.ServicesConfig{
			PeopleOrganizations: config.ServiceConfig{
				Name:    "people-organizations",
				URL:     "http://localhost:8081/graphql",
				Enabled: true,
			},
			ECommerce: config.ServiceConfig{
				Name:    "e-commerce",
				URL:     "http://localhost:8082/graphql",
				Enabled: false, // Disabled service
			},
		},
	}

	gateway, err := New(cfg)
	require.NoError(t, err)

	// Test enabled services
	services := gateway.GetServices()
	assert.Contains(t, services, "people-organizations")
	
	peopleService := services["people-organizations"]
	assert.Equal(t, "people-organizations", peopleService.Name)
	assert.Equal(t, "http://localhost:8081/graphql", peopleService.URL)
	assert.True(t, peopleService.Enabled)

	// Test service enablement check
	assert.True(t, gateway.IsServiceEnabled("people-organizations"))
	assert.False(t, gateway.IsServiceEnabled("nonexistent-service"))
}

func TestComposeSchema(t *testing.T) {
	cfg := createTestConfig()
	gateway, err := New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This should currently return nil (not implemented)
	err = gateway.ComposeSchema(ctx)
	
	// For now, we expect no error since it's just a placeholder
	// When implementation is added, this test should verify actual schema composition
	assert.NoError(t, err, "Schema composition should not error (placeholder implementation)")
}

func TestExecuteQuery(t *testing.T) {
	cfg := createTestConfig()
	gateway, err := New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `query { person(id: "1") { id name } }`
	variables := map[string]interface{}{"id": "1"}

	result, err := gateway.ExecuteQuery(ctx, query, variables)
	
	// Should return error since not implemented yet
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestGetServices(t *testing.T) {
	cfg := createTestConfig()
	gateway, err := New(cfg)
	require.NoError(t, err)

	services := gateway.GetServices()
	assert.NotNil(t, services)
	assert.IsType(t, map[string]*ServiceClient{}, services)
	
	// Should contain enabled services
	enabledCount := 0
	for _, service := range services {
		if service.Enabled {
			enabledCount++
		}
	}
	assert.Greater(t, enabledCount, 0, "Should have at least one enabled service")
}

func TestIsServiceEnabled(t *testing.T) {
	cfg := createTestConfig()
	gateway, err := New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name        string
		serviceName string
		expected    bool
	}{
		{
			name:        "enabled service",
			serviceName: "people-organizations",
			expected:    true,
		},
		{
			name:        "nonexistent service",
			serviceName: "nonexistent",
			expected:    false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gateway.IsServiceEnabled(tt.serviceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		Environment: "test",
		Services: config.ServicesConfig{
			PeopleOrganizations: config.ServiceConfig{
				Name:    "people-organizations",
				URL:     "http://localhost:8081/graphql",
				Timeout: 30 * time.Second,
				Retries: 3,
				Enabled: true,
			},
			ECommerce: config.ServiceConfig{
				Name:    "e-commerce", 
				URL:     "http://localhost:8082/graphql",
				Timeout: 30 * time.Second,
				Retries: 3,
				Enabled: true,
			},
			Products: config.ServiceConfig{
				Name:    "products",
				URL:     "http://localhost:8084/graphql",
				Timeout: 30 * time.Second,
				Retries: 3,
				Enabled: true,
			},
		},
		Federation: config.FederationConfig{
			QueryTimeout:            30 * time.Second,
			MaxQueryComplexity:      1000,
			EnableQueryValidation:   true,
			EnableComplexityAnalysis: true,
		},
	}
}