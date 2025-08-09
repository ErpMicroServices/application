package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
	"github.com/erpmicroservices/graphql-federation-gateway/internal/federation"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		federation    *federation.Gateway
		expectedError bool
	}{
		{
			name:          "valid configuration",
			config:        createTestConfig(),
			federation:    createTestFederation(),
			expectedError: false,
		},
		{
			name:          "nil configuration",
			config:        nil,
			federation:    createTestFederation(),
			expectedError: true,
		},
		{
			name:          "nil federation",
			config:        createTestConfig(),
			federation:    nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(tt.config, tt.federation)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, server)
			} else {
				require.NoError(t, err)
				require.NotNil(t, server)
				assert.NotNil(t, server.config)
				assert.NotNil(t, server.federation)
				assert.NotNil(t, server.httpServer)
			}
		})
	}
}

func TestServerRouterSetup(t *testing.T) {
	cfg := createTestConfig()
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	router := server.setupRouter()
	require.NotNil(t, router)

	// Test route setup by making test requests
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "health check endpoint",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "graphql endpoint - not implemented",
			method:         "POST",
			path:           "/graphql",
			expectedStatus: http.StatusNotImplemented,
			expectedHeader: "application/json",
		},
		{
			name:           "playground endpoint - test environment",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedHeader: "text/html",
		},
		{
			name:           "metrics endpoint - not implemented",
			method:         "GET",
			path:           "/metrics",
			expectedStatus: http.StatusNotImplemented,
			expectedHeader: "text/plain",
		},
		{
			name:           "non-existent endpoint",
			method:         "GET",
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, testServer.URL+tt.path, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestHealthCheckHandler(t *testing.T) {
	cfg := createTestConfig()
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	
	// Should contain valid JSON with status and timestamp
	body := w.Body.String()
	assert.Contains(t, body, "healthy")
	assert.Contains(t, body, "timestamp")
}

func TestGraphQLHandler(t *testing.T) {
	cfg := createTestConfig()
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/graphql", nil)
	w := httptest.NewRecorder()

	server.handleGraphQL(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return not implemented for now
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	
	body := w.Body.String()
	assert.Contains(t, body, "not yet implemented")
}

func TestPlaygroundHandler(t *testing.T) {
	cfg := createTestConfig()
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handlePlayground(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
	
	body := w.Body.String()
	assert.Contains(t, body, "GraphQL Federation Gateway")
	assert.Contains(t, body, "/graphql")
}

func TestMetricsHandler(t *testing.T) {
	cfg := createTestConfig()
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return not implemented for now
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	
	body := w.Body.String()
	assert.Contains(t, body, "not yet implemented")
}

func TestServerLifecycle(t *testing.T) {
	cfg := createTestConfig()
	cfg.Server.Port = "0" // Use random available port
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	// Should not error even if server isn't running
	assert.NoError(t, err)
}

func TestCORSConfiguration(t *testing.T) {
	cfg := createTestConfig()
	cfg.Server.EnableCORS = true
	cfg.Server.CORSOrigins = []string{"https://example.com"}
	
	fedGateway := createTestFederation()
	
	server, err := NewServer(cfg, fedGateway)
	require.NoError(t, err)

	router := server.setupRouter()
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Test CORS preflight request
	req, err := http.NewRequest("OPTIONS", testServer.URL+"/graphql", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// CORS headers should be present
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Origin"), "example.com")
}

// Helper functions

func createTestConfig() *config.Config {
	return &config.Config{
		Environment: "test",
		Server: config.ServerConfig{
			Port:                "4004",
			Host:                "localhost",
			ReadTimeout:         10 * time.Second,
			WriteTimeout:        10 * time.Second,
			IdleTimeout:         60 * time.Second,
			EnablePlayground:    true,
			EnableIntrospection: true,
			EnableCORS:          true,
			CORSOrigins:         []string{"*"},
		},
		Monitoring: config.MonitoringConfig{
			EnableMetrics:   true,
			HealthCheckPath: "/health",
		},
	}
}

func createTestFederation() *federation.Gateway {
	cfg := &config.Config{
		Services: config.ServicesConfig{
			PeopleOrganizations: config.ServiceConfig{
				Name:    "people-organizations",
				URL:     "http://localhost:8081/graphql",
				Enabled: true,
			},
		},
	}
	
	gateway, err := federation.New(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test federation: %v", err))
	}
	
	return gateway
}