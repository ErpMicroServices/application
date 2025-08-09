package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the auth-go module
type Config struct {
	OAuth2  OAuth2Config
	JWT     JWTConfig
	Cache   CacheConfig
	Service ServiceConfig
	Logging LoggingConfig
	CORS    CORSConfig
	RateLimit RateLimitConfig
}

// OAuth2Config holds OAuth2 specific configuration
type OAuth2Config struct {
	ClientID                string
	ClientSecret           string
	AuthorizationServerURL string
	TokenURL               string
	AuthorizeURL           string
	UserInfoURL            string
	IntrospectURL          string
	JWKSURL                string
	RedirectURL            string
	Scopes                 []string
	Timeout                time.Duration
}

// JWTConfig holds JWT specific configuration
type JWTConfig struct {
	Issuer     string
	Audience   string
	SigningKey string
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name string
	Port string
	Host string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		OAuth2: OAuth2Config{
			ClientID:                getEnvOrDefault("OAUTH2_CLIENT_ID", "erp-microservices-client"),
			ClientSecret:           getEnvOrDefault("OAUTH2_CLIENT_SECRET", ""),
			AuthorizationServerURL: getEnvOrDefault("OAUTH2_AUTHORIZATION_SERVER_URL", "http://localhost:9090"),
			TokenURL:               getEnvOrDefault("OAUTH2_TOKEN_URL", "http://localhost:9090/oauth2/token"),
			AuthorizeURL:           getEnvOrDefault("OAUTH2_AUTHORIZE_URL", "http://localhost:9090/oauth2/authorize"),
			UserInfoURL:            getEnvOrDefault("OAUTH2_USERINFO_URL", "http://localhost:9090/oauth2/userinfo"),
			IntrospectURL:          getEnvOrDefault("OAUTH2_INTROSPECT_URL", "http://localhost:9090/oauth2/introspect"),
			JWKSURL:                getEnvOrDefault("OAUTH2_JWKS_URL", "http://localhost:9090/oauth2/jwks"),
			RedirectURL:            getEnvOrDefault("OAUTH2_REDIRECT_URL", "http://localhost:8080/auth/callback"),
			Scopes:                 getEnvAsSlice("OAUTH2_SCOPES", []string{"read", "write"}),
			Timeout:                getEnvAsDuration("OAUTH2_TIMEOUT", 30*time.Second),
		},
		JWT: JWTConfig{
			Issuer:     getEnvOrDefault("JWT_ISSUER", "http://localhost:9090"),
			Audience:   getEnvOrDefault("JWT_AUDIENCE", "erp-microservices"),
			SigningKey: getEnvOrDefault("JWT_SIGNING_KEY", ""),
		},
		Cache: CacheConfig{
			TTL:             getEnvAsDuration("TOKEN_CACHE_TTL", 3600*time.Second),
			CleanupInterval: getEnvAsDuration("TOKEN_CACHE_CLEANUP_INTERVAL", 600*time.Second),
		},
		Service: ServiceConfig{
			Name: getEnvOrDefault("SERVICE_NAME", "auth-go"),
			Port: getEnvOrDefault("SERVICE_PORT", "8080"),
			Host: getEnvOrDefault("SERVICE_HOST", "localhost"),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvAsDuration("RATE_LIMIT_WINDOW", 60*time.Second),
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.OAuth2.ClientID == "" {
		return &ConfigError{Field: "OAUTH2_CLIENT_ID", Message: "client ID is required"}
	}
	if c.OAuth2.ClientSecret == "" {
		return &ConfigError{Field: "OAUTH2_CLIENT_SECRET", Message: "client secret is required"}
	}
	if c.OAuth2.AuthorizationServerURL == "" {
		return &ConfigError{Field: "OAUTH2_AUTHORIZATION_SERVER_URL", Message: "authorization server URL is required"}
	}
	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error in field " + e.Field + ": " + e.Message
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}