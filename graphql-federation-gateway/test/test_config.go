package test

import (
	"os"
	"time"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
)

// LoadTestConfig creates a configuration suitable for testing
func LoadTestConfig() (*config.Config, error) {
	// Set test environment variables
	testEnv := map[string]string{
		"ENVIRONMENT":                  "test",
		"SERVER_PORT":                  "4003",
		"SERVER_HOST":                  "localhost",
		"SERVER_ENABLE_PLAYGROUND":     "false",
		"SERVER_ENABLE_INTROSPECTION":  "false",
		"LOG_LEVEL":                    "error",
		"TOKEN_VALIDATION":             "false",
		"ENABLE_METRICS":               "true",
		"ENABLE_TRACING":               "false",
		"PEOPLE_ORGS_ENABLED":          "true",
		"ECOMMERCE_ENABLED":            "true",
		"PRODUCTS_ENABLED":             "true",
		"ACCOUNTING_ENABLED":           "false",
		"ORDERS_ENABLED":               "false",
		"INVOICES_ENABLED":             "false",
		"SHIPMENTS_ENABLED":            "false",
		"HR_ENABLED":                   "false",
		"WORK_EFFORT_ENABLED":          "false",
		"ANALYTICS_ENABLED":            "false",
	}

	// Set environment variables for test
	for key, value := range testEnv {
		os.Setenv(key, value)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Override specific test values
	cfg.Federation.QueryTimeout = 5 * time.Second
	cfg.Server.ReadTimeout = 10 * time.Second
	cfg.Server.WriteTimeout = 10 * time.Second

	return cfg, nil
}

// CleanupTestConfig removes test environment variables
func CleanupTestConfig() {
	testKeys := []string{
		"ENVIRONMENT",
		"SERVER_PORT",
		"SERVER_HOST",
		"SERVER_ENABLE_PLAYGROUND",
		"SERVER_ENABLE_INTROSPECTION",
		"LOG_LEVEL",
		"TOKEN_VALIDATION",
		"ENABLE_METRICS",
		"ENABLE_TRACING",
		"PEOPLE_ORGS_ENABLED",
		"ECOMMERCE_ENABLED",
		"PRODUCTS_ENABLED",
		"ACCOUNTING_ENABLED",
		"ORDERS_ENABLED",
		"INVOICES_ENABLED",
		"SHIPMENTS_ENABLED",
		"HR_ENABLED",
		"WORK_EFFORT_ENABLED",
		"ANALYTICS_ENABLED",
	}

	for _, key := range testKeys {
		os.Unsetenv(key)
	}
}