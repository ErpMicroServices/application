package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the GraphQL federation gateway
type Config struct {
	Environment string        `json:"environment"`
	Server      ServerConfig  `json:"server"`
	Services    ServicesConfig `json:"services"`
	Auth        AuthConfig    `json:"auth"`
	Monitoring  MonitoringConfig `json:"monitoring"`
	Federation  FederationConfig `json:"federation"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port                string        `json:"port"`
	Host                string        `json:"host"`
	ReadTimeout         time.Duration `json:"read_timeout"`
	WriteTimeout        time.Duration `json:"write_timeout"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	ShutdownTimeout     time.Duration `json:"shutdown_timeout"`
	EnablePlayground    bool          `json:"enable_playground"`
	EnableIntrospection bool          `json:"enable_introspection"`
	EnableCORS          bool          `json:"enable_cors"`
	CORSOrigins         []string      `json:"cors_origins"`
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int    `json:"retries"`
	Enabled  bool   `json:"enabled"`
}

// ServicesConfig holds all federated services configuration
type ServicesConfig struct {
	PeopleOrganizations ServiceConfig `json:"people_organizations"`
	ECommerce          ServiceConfig `json:"e_commerce"`
	Products           ServiceConfig `json:"products"`
	Accounting         ServiceConfig `json:"accounting"`
	Orders             ServiceConfig `json:"orders"`
	Invoices           ServiceConfig `json:"invoices"`
	Shipments          ServiceConfig `json:"shipments"`
	HumanResources     ServiceConfig `json:"human_resources"`
	WorkEffort         ServiceConfig `json:"work_effort"`
	Analytics          ServiceConfig `json:"analytics"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string        `json:"jwt_secret"`
	JWTIssuer          string        `json:"jwt_issuer"`
	JWTAudience        string        `json:"jwt_audience"`
	TokenValidation    bool          `json:"token_validation"`
	AuthServiceURL     string        `json:"auth_service_url"`
	PublicKeyURL       string        `json:"public_key_url"`
	TokenCacheDuration time.Duration `json:"token_cache_duration"`
}

// MonitoringConfig holds monitoring and observability configuration
type MonitoringConfig struct {
	EnableMetrics     bool          `json:"enable_metrics"`
	MetricsPort       string        `json:"metrics_port"`
	EnableTracing     bool          `json:"enable_tracing"`
	TracingEndpoint   string        `json:"tracing_endpoint"`
	HealthCheckPath   string        `json:"health_check_path"`
	LogLevel          string        `json:"log_level"`
	LogFormat         string        `json:"log_format"`
	RequestLogging    bool          `json:"request_logging"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"`
}

// FederationConfig holds Apollo Federation specific configuration
type FederationConfig struct {
	SchemaPollingInterval    time.Duration `json:"schema_polling_interval"`
	QueryPlanCacheSize      int           `json:"query_plan_cache_size"`
	QueryPlanCacheTTL       time.Duration `json:"query_plan_cache_ttl"`
	MaxQueryComplexity      int           `json:"max_query_complexity"`
	QueryTimeout            time.Duration `json:"query_timeout"`
	EnableQueryValidation   bool          `json:"enable_query_validation"`
	EnableComplexityAnalysis bool          `json:"enable_complexity_analysis"`
	BatchingEnabled         bool          `json:"batching_enabled"`
	BatchTimeout            time.Duration `json:"batch_timeout"`
	MaxBatchSize            int           `json:"max_batch_size"`
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:                getEnv("SERVER_PORT", "4000"),
			Host:                getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:         getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:        getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:         getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:     getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			EnablePlayground:    getBoolEnv("SERVER_ENABLE_PLAYGROUND", true),
			EnableIntrospection: getBoolEnv("SERVER_ENABLE_INTROSPECTION", true),
			EnableCORS:          getBoolEnv("SERVER_ENABLE_CORS", true),
			CORSOrigins:         getStringSliceEnv("SERVER_CORS_ORIGINS", []string{"*"}),
		},
		Services: ServicesConfig{
			PeopleOrganizations: ServiceConfig{
				Name:    "people-organizations",
				URL:     getEnv("PEOPLE_ORGS_SERVICE_URL", "http://localhost:8081/graphql"),
				Timeout: getDurationEnv("PEOPLE_ORGS_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("PEOPLE_ORGS_RETRIES", 3),
				Enabled: getBoolEnv("PEOPLE_ORGS_ENABLED", true),
			},
			ECommerce: ServiceConfig{
				Name:    "e-commerce",
				URL:     getEnv("ECOMMERCE_SERVICE_URL", "http://localhost:8082/graphql"),
				Timeout: getDurationEnv("ECOMMERCE_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("ECOMMERCE_RETRIES", 3),
				Enabled: getBoolEnv("ECOMMERCE_ENABLED", true),
			},
			Products: ServiceConfig{
				Name:    "products",
				URL:     getEnv("PRODUCTS_SERVICE_URL", "http://localhost:8084/graphql"),
				Timeout: getDurationEnv("PRODUCTS_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("PRODUCTS_RETRIES", 3),
				Enabled: getBoolEnv("PRODUCTS_ENABLED", true),
			},
			Accounting: ServiceConfig{
				Name:    "accounting-budgeting",
				URL:     getEnv("ACCOUNTING_SERVICE_URL", "http://localhost:8083/graphql"),
				Timeout: getDurationEnv("ACCOUNTING_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("ACCOUNTING_RETRIES", 3),
				Enabled: getBoolEnv("ACCOUNTING_ENABLED", true),
			},
			Orders: ServiceConfig{
				Name:    "orders",
				URL:     getEnv("ORDERS_SERVICE_URL", "http://localhost:8085/graphql"),
				Timeout: getDurationEnv("ORDERS_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("ORDERS_RETRIES", 3),
				Enabled: getBoolEnv("ORDERS_ENABLED", true),
			},
			Invoices: ServiceConfig{
				Name:    "invoices",
				URL:     getEnv("INVOICES_SERVICE_URL", "http://localhost:8086/graphql"),
				Timeout: getDurationEnv("INVOICES_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("INVOICES_RETRIES", 3),
				Enabled: getBoolEnv("INVOICES_ENABLED", true),
			},
			Shipments: ServiceConfig{
				Name:    "shipments",
				URL:     getEnv("SHIPMENTS_SERVICE_URL", "http://localhost:8087/graphql"),
				Timeout: getDurationEnv("SHIPMENTS_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("SHIPMENTS_RETRIES", 3),
				Enabled: getBoolEnv("SHIPMENTS_ENABLED", true),
			},
			HumanResources: ServiceConfig{
				Name:    "human-resources",
				URL:     getEnv("HR_SERVICE_URL", "http://localhost:8088/graphql"),
				Timeout: getDurationEnv("HR_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("HR_RETRIES", 3),
				Enabled: getBoolEnv("HR_ENABLED", true),
			},
			WorkEffort: ServiceConfig{
				Name:    "work-effort",
				URL:     getEnv("WORK_EFFORT_SERVICE_URL", "http://localhost:8089/graphql"),
				Timeout: getDurationEnv("WORK_EFFORT_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("WORK_EFFORT_RETRIES", 3),
				Enabled: getBoolEnv("WORK_EFFORT_ENABLED", true),
			},
			Analytics: ServiceConfig{
				Name:    "analytics",
				URL:     getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8090/graphql"),
				Timeout: getDurationEnv("ANALYTICS_TIMEOUT", 30*time.Second),
				Retries: getIntEnv("ANALYTICS_RETRIES", 3),
				Enabled: getBoolEnv("ANALYTICS_ENABLED", false), // Optional service
			},
		},
		Auth: AuthConfig{
			JWTSecret:           getEnv("JWT_SECRET", ""),
			JWTIssuer:          getEnv("JWT_ISSUER", "erp-microservices"),
			JWTAudience:        getEnv("JWT_AUDIENCE", "erp-federation-gateway"),
			TokenValidation:    getBoolEnv("TOKEN_VALIDATION", true),
			AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8080"),
			PublicKeyURL:       getEnv("PUBLIC_KEY_URL", ""),
			TokenCacheDuration: getDurationEnv("TOKEN_CACHE_DURATION", 5*time.Minute),
		},
		Monitoring: MonitoringConfig{
			EnableMetrics:      getBoolEnv("ENABLE_METRICS", true),
			MetricsPort:        getEnv("METRICS_PORT", "9090"),
			EnableTracing:      getBoolEnv("ENABLE_TRACING", true),
			TracingEndpoint:    getEnv("TRACING_ENDPOINT", "http://localhost:14268/api/traces"),
			HealthCheckPath:    getEnv("HEALTH_CHECK_PATH", "/health"),
			LogLevel:           getEnv("LOG_LEVEL", "info"),
			LogFormat:          getEnv("LOG_FORMAT", "json"),
			RequestLogging:     getBoolEnv("REQUEST_LOGGING", true),
			SlowQueryThreshold: getDurationEnv("SLOW_QUERY_THRESHOLD", 1*time.Second),
		},
		Federation: FederationConfig{
			SchemaPollingInterval:    getDurationEnv("SCHEMA_POLLING_INTERVAL", 30*time.Second),
			QueryPlanCacheSize:      getIntEnv("QUERY_PLAN_CACHE_SIZE", 1000),
			QueryPlanCacheTTL:       getDurationEnv("QUERY_PLAN_CACHE_TTL", 5*time.Minute),
			MaxQueryComplexity:      getIntEnv("MAX_QUERY_COMPLEXITY", 1000),
			QueryTimeout:            getDurationEnv("QUERY_TIMEOUT", 30*time.Second),
			EnableQueryValidation:   getBoolEnv("ENABLE_QUERY_VALIDATION", true),
			EnableComplexityAnalysis: getBoolEnv("ENABLE_COMPLEXITY_ANALYSIS", true),
			BatchingEnabled:         getBoolEnv("BATCHING_ENABLED", true),
			BatchTimeout:            getDurationEnv("BATCH_TIMEOUT", 16*time.Millisecond),
			MaxBatchSize:           getIntEnv("MAX_BATCH_SIZE", 100),
		},
	}

	// Validate required configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate ensures required configuration values are present
func (c *Config) validate() error {
	if c.Auth.TokenValidation && c.Auth.JWTSecret == "" && c.Auth.PublicKeyURL == "" {
		return fmt.Errorf("JWT_SECRET or PUBLIC_KEY_URL must be set when token validation is enabled")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT cannot be empty")
	}

	return nil
}

// GetEnabledServices returns a list of enabled services
func (c *Config) GetEnabledServices() []ServiceConfig {
	var services []ServiceConfig

	allServices := []ServiceConfig{
		c.Services.PeopleOrganizations,
		c.Services.ECommerce,
		c.Services.Products,
		c.Services.Accounting,
		c.Services.Orders,
		c.Services.Invoices,
		c.Services.Shipments,
		c.Services.HumanResources,
		c.Services.WorkEffort,
		c.Services.Analytics,
	}

	for _, service := range allServices {
		if service.Enabled {
			services = append(services, service)
		}
	}

	return services
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		// Simple comma-separated parsing
		// For more complex parsing, consider using a proper CSV parser
		return []string{value} // Simplified for now
	}
	return defaultValue
}