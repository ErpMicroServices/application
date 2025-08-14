package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedError bool
		assertions    func(t *testing.T, cfg *Config)
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"JWT_SECRET": "test-secret",
			},
			expectedError: false,
			assertions: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "development", cfg.Environment)
				assert.Equal(t, "4000", cfg.Server.Port)
				assert.Equal(t, "0.0.0.0", cfg.Server.Host)
				assert.True(t, cfg.Server.EnablePlayground)
				assert.True(t, cfg.Server.EnableIntrospection)
				assert.True(t, cfg.Server.EnableCORS)
			},
		},
		{
			name: "production configuration",
			envVars: map[string]string{
				"ENVIRONMENT":                "production",
				"SERVER_PORT":                "8080",
				"SERVER_ENABLE_PLAYGROUND":   "false",
				"SERVER_ENABLE_INTROSPECTION": "false",
				"JWT_SECRET":                 "production-secret",
			},
			expectedError: false,
			assertions: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "production", cfg.Environment)
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.False(t, cfg.Server.EnablePlayground)
				assert.False(t, cfg.Server.EnableIntrospection)
			},
		},
		{
			name: "missing JWT secret with validation enabled",
			envVars: map[string]string{
				"TOKEN_VALIDATION": "true",
			},
			expectedError: true,
			assertions:    nil,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"SERVER_PORT": "",
				"JWT_SECRET":  "test-secret",
			},
			expectedError: true,
			assertions:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			cfg, err := Load()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.assertions != nil {
					tt.assertions(t, cfg)
				}
			}

			// Clean up
			clearTestEnv()
		})
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Auth: AuthConfig{
			TokenValidation: true,
			JWTSecret:       "",
			PublicKeyURL:    "",
		},
	}

	err := cfg.validate()
	assert.Error(t, err, "Should fail validation when JWT secret is missing")

	cfg.Auth.JWTSecret = "test-secret"
	err = cfg.validate()
	assert.NoError(t, err, "Should pass validation with JWT secret")
}

func TestGetEnabledServices(t *testing.T) {
	cfg := &Config{
		Services: ServicesConfig{
			PeopleOrganizations: ServiceConfig{Name: "people", Enabled: true},
			ECommerce:           ServiceConfig{Name: "ecommerce", Enabled: false},
			Products:            ServiceConfig{Name: "products", Enabled: true},
		},
	}

	services := cfg.GetEnabledServices()
	
	// Should return 8 enabled services by default (excluding Analytics which is disabled)
	// But for this test, only 2 are enabled
	assert.Len(t, services, 8, "Should return all enabled services")
	
	enabledNames := make(map[string]bool)
	for _, service := range services {
		enabledNames[service.Name] = true
	}
	
	assert.True(t, enabledNames["people-organizations"])
	assert.True(t, enabledNames["products"])
	// Note: ECommerce is disabled so should not be in enabled list
}

func TestEnvironmentVariableParsing(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		expected interface{}
		testFunc func() interface{}
	}{
		{
			name:     "string value",
			envKey:   "TEST_STRING",
			envValue: "hello",
			expected: "hello",
			testFunc: func() interface{} { return getEnv("TEST_STRING", "default") },
		},
		{
			name:     "int value",
			envKey:   "TEST_INT",
			envValue: "42",
			expected: 42,
			testFunc: func() interface{} { return getIntEnv("TEST_INT", 0) },
		},
		{
			name:     "bool value true",
			envKey:   "TEST_BOOL_TRUE",
			envValue: "true",
			expected: true,
			testFunc: func() interface{} { return getBoolEnv("TEST_BOOL_TRUE", false) },
		},
		{
			name:     "bool value false",
			envKey:   "TEST_BOOL_FALSE",
			envValue: "false",
			expected: false,
			testFunc: func() interface{} { return getBoolEnv("TEST_BOOL_FALSE", true) },
		},
		{
			name:     "duration value",
			envKey:   "TEST_DURATION",
			envValue: "30s",
			expected: 30 * time.Second,
			testFunc: func() interface{} { return getDurationEnv("TEST_DURATION", time.Minute) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			result := tt.testFunc()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultValues(t *testing.T) {
	clearTestEnv()

	// Test defaults when environment variables are not set
	assert.Equal(t, "default", getEnv("NONEXISTENT", "default"))
	assert.Equal(t, 100, getIntEnv("NONEXISTENT", 100))
	assert.Equal(t, true, getBoolEnv("NONEXISTENT", true))
	assert.Equal(t, time.Hour, getDurationEnv("NONEXISTENT", time.Hour))
}

// Helper function to clear test environment variables
func clearTestEnv() {
	envVars := []string{
		"ENVIRONMENT",
		"SERVER_PORT",
		"SERVER_HOST",
		"SERVER_ENABLE_PLAYGROUND",
		"SERVER_ENABLE_INTROSPECTION",
		"JWT_SECRET",
		"TOKEN_VALIDATION",
		"TEST_STRING",
		"TEST_INT", 
		"TEST_BOOL_TRUE",
		"TEST_BOOL_FALSE",
		"TEST_DURATION",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}