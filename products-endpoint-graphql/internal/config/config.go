package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	GraphQL  GraphQLConfig  `mapstructure:"graphql"`
	Auth     AuthConfig     `mapstructure:"auth"`
	AI       AIConfig       `mapstructure:"ai"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port" default:"8080"`
	Host         string        `mapstructure:"host" default:"0.0.0.0"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"30s"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" default:"120s"`
	Environment  string        `mapstructure:"environment" default:"development"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled" default:"false"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host                 string        `mapstructure:"host" default:"localhost"`
	Port                 int           `mapstructure:"port" default:"5432"`
	Name                 string        `mapstructure:"name" default:"products_db"`
	Username             string        `mapstructure:"username" default:"products_user"`
	Password             string        `mapstructure:"password" default:"products_password"`
	SSLMode              string        `mapstructure:"ssl_mode" default:"disable"`
	MaxOpenConns         int           `mapstructure:"max_open_conns" default:"25"`
	MaxIdleConns         int           `mapstructure:"max_idle_conns" default:"25"`
	ConnMaxLifetime      time.Duration `mapstructure:"conn_max_lifetime" default:"5m"`
	ConnMaxIdleTime      time.Duration `mapstructure:"conn_max_idle_time" default:"5m"`
	MigrationPath        string        `mapstructure:"migration_path" default:"file://migrations"`
	EnableQueryLogging   bool          `mapstructure:"enable_query_logging" default:"false"`
	SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" default:"1s"`
}

// GraphQLConfig holds GraphQL server configuration
type GraphQLConfig struct {
	Playground         bool          `mapstructure:"playground" default:"true"`
	Introspection      bool          `mapstructure:"introspection" default:"true"`
	ComplexityLimit    int           `mapstructure:"complexity_limit" default:"1000"`
	DepthLimit         int           `mapstructure:"depth_limit" default:"15"`
	EnableDataLoader   bool          `mapstructure:"enable_dataloader" default:"true"`
	DataLoaderWait     time.Duration `mapstructure:"dataloader_wait" default:"10ms"`
	DataLoaderMaxBatch int           `mapstructure:"dataloader_max_batch" default:"100"`
	QueryTimeout       time.Duration `mapstructure:"query_timeout" default:"30s"`
	EnableMetrics      bool          `mapstructure:"enable_metrics" default:"true"`
	EnableTracing      bool          `mapstructure:"enable_tracing" default:"false"`
	CacheEnabled       bool          `mapstructure:"cache_enabled" default:"true"`
	CacheTTL           time.Duration `mapstructure:"cache_ttl" default:"5m"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled            bool          `mapstructure:"enabled" default:"true"`
	JWTSecret          string        `mapstructure:"jwt_secret"`
	JWTExpiration      time.Duration `mapstructure:"jwt_expiration" default:"24h"`
	OAuth2Issuer       string        `mapstructure:"oauth2_issuer"`
	OAuth2Audience     string        `mapstructure:"oauth2_audience"`
	OAuth2ClientID     string        `mapstructure:"oauth2_client_id"`
	OAuth2ClientSecret string        `mapstructure:"oauth2_client_secret"`
	OAuth2RedirectURL  string        `mapstructure:"oauth2_redirect_url"`
	RequiredScopes     []string      `mapstructure:"required_scopes"`
	EnableRBAC         bool          `mapstructure:"enable_rbac" default:"true"`
	CacheTokens        bool          `mapstructure:"cache_tokens" default:"true"`
	TokenCacheTTL      time.Duration `mapstructure:"token_cache_ttl" default:"5m"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	Enabled                    bool          `mapstructure:"enabled" default:"true"`
	CategorizationEnabled      bool          `mapstructure:"categorization_enabled" default:"true"`
	RecommendationEnabled      bool          `mapstructure:"recommendation_enabled" default:"true"`
	ImageAnalysisEnabled       bool          `mapstructure:"image_analysis_enabled" default:"true"`
	PricingOptimizationEnabled bool          `mapstructure:"pricing_optimization_enabled" default:"true"`
	DemandForecastEnabled      bool          `mapstructure:"demand_forecast_enabled" default:"true"`
	CompetitorAnalysisEnabled  bool          `mapstructure:"competitor_analysis_enabled" default:"true"`
	InventoryOptimizationEnabled bool        `mapstructure:"inventory_optimization_enabled" default:"true"`
	
	// Model Configuration
	ModelProvider         string        `mapstructure:"model_provider" default:"openai"`
	ModelName             string        `mapstructure:"model_name" default:"gpt-4-vision-preview"`
	ModelAPIKey           string        `mapstructure:"model_api_key"`
	ModelEndpoint         string        `mapstructure:"model_endpoint"`
	ModelTimeout          time.Duration `mapstructure:"model_timeout" default:"30s"`
	ModelMaxRetries       int           `mapstructure:"model_max_retries" default:"3"`
	ModelRetryDelay       time.Duration `mapstructure:"model_retry_delay" default:"1s"`
	
	// Processing Configuration
	BatchSize             int           `mapstructure:"batch_size" default:"10"`
	BatchTimeout          time.Duration `mapstructure:"batch_timeout" default:"2m"`
	MaxConcurrentJobs     int           `mapstructure:"max_concurrent_jobs" default:"5"`
	QueueSize             int           `mapstructure:"queue_size" default:"100"`
	
	// Confidence Thresholds
	CategoryConfidenceThreshold      float64 `mapstructure:"category_confidence_threshold" default:"0.8"`
	RecommendationConfidenceThreshold float64 `mapstructure:"recommendation_confidence_threshold" default:"0.6"`
	ImageAnalysisConfidenceThreshold  float64 `mapstructure:"image_analysis_confidence_threshold" default:"0.7"`
	
	// Cache Configuration
	CacheEnabled      bool          `mapstructure:"cache_enabled" default:"true"`
	CacheTTL          time.Duration `mapstructure:"cache_ttl" default:"1h"`
	CacheKeyPrefix    string        `mapstructure:"cache_key_prefix" default:"ai:products:"`
	
	// Feature Flags
	EnableFeedbackLearning    bool `mapstructure:"enable_feedback_learning" default:"true"`
	EnableModelVersioning     bool `mapstructure:"enable_model_versioning" default:"true"`
	EnablePerformanceMetrics  bool `mapstructure:"enable_performance_metrics" default:"true"`
	EnableAuditLogging        bool `mapstructure:"enable_audit_logging" default:"true"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level         string `mapstructure:"level" default:"info"`
	Format        string `mapstructure:"format" default:"json"`
	Output        string `mapstructure:"output" default:"stdout"`
	EnableConsole bool   `mapstructure:"enable_console" default:"false"`
	EnableFile    bool   `mapstructure:"enable_file" default:"true"`
	FilePath      string `mapstructure:"file_path" default:"logs/products-api.log"`
	MaxSize       int    `mapstructure:"max_size" default:"100"`
	MaxBackups    int    `mapstructure:"max_backups" default:"3"`
	MaxAge        int    `mapstructure:"max_age" default:"30"`
	Compress      bool   `mapstructure:"compress" default:"true"`
}

// RedisConfig holds Redis configuration for caching
type RedisConfig struct {
	Enabled      bool          `mapstructure:"enabled" default:"true"`
	Host         string        `mapstructure:"host" default:"localhost"`
	Port         int           `mapstructure:"port" default:"6379"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database" default:"0"`
	MaxRetries   int           `mapstructure:"max_retries" default:"3"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" default:"5s"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" default:"3s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"3s"`
	PoolSize     int           `mapstructure:"pool_size" default:"10"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout" default:"4s"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" default:"5m"`
	KeyPrefix    string        `mapstructure:"key_prefix" default:"products:"`
}

// Load loads configuration from various sources
func Load() (*Config, error) {
	config := &Config{}

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.environment", "development")
	
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "products_db")
	viper.SetDefault("database.username", "products_user")
	viper.SetDefault("database.password", "products_password")
	viper.SetDefault("database.ssl_mode", "disable")
	
	viper.SetDefault("graphql.playground", true)
	viper.SetDefault("graphql.introspection", true)
	viper.SetDefault("graphql.complexity_limit", 1000)
	
	viper.SetDefault("auth.enabled", true)
	viper.SetDefault("auth.enable_rbac", true)
	
	viper.SetDefault("ai.enabled", true)
	viper.SetDefault("ai.model_provider", "openai")
	viper.SetDefault("ai.batch_size", 10)
	
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	
	viper.SetDefault("redis.enabled", true)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)

	// Configuration file settings
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/products-api")
	viper.AddConfigPath("$HOME/.products-api")

	// Environment variable settings
	viper.SetEnvPrefix("PRODUCTS_API")
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; ignore error if desired
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Auth.Enabled && c.Auth.JWTSecret == "" && c.Auth.OAuth2ClientID == "" {
		return fmt.Errorf("authentication is enabled but no JWT secret or OAuth2 client ID provided")
	}

	if c.AI.Enabled && c.AI.ModelAPIKey == "" {
		return fmt.Errorf("AI is enabled but no model API key provided")
	}

	if c.AI.CategoryConfidenceThreshold < 0 || c.AI.CategoryConfidenceThreshold > 1 {
		return fmt.Errorf("invalid category confidence threshold: %f", c.AI.CategoryConfidenceThreshold)
	}

	if c.GraphQL.ComplexityLimit < 1 {
		return fmt.Errorf("GraphQL complexity limit must be positive")
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production" || c.Server.Environment == "prod"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development" || c.Server.Environment == "dev"
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}