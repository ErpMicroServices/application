package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	NATS     NATSConfig     `json:"nats"`
	App      AppConfig      `json:"app"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Database        string `json:"database"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	SSLMode         string `json:"ssl_mode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

// NATSConfig holds NATS streaming configuration
type NATSConfig struct {
	URL       string `json:"url"`
	ClusterID string `json:"cluster_id"`
	ClientID  string `json:"client_id"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment   string `json:"environment"`
	LogLevel      string `json:"log_level"`
	GraphQLPath   string `json:"graphql_path"`
	PlaygroundPath string `json:"playground_path"`
	Version       string `json:"version"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getIntEnv("SERVER_PORT", 8080),
			Host:         getStringEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getIntEnv("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getIntEnv("SERVER_WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:            getStringEnv("DB_HOST", "localhost"),
			Port:            getIntEnv("DB_PORT", 5432),
			Database:        getStringEnv("DB_NAME", "order_fulfillment_db"),
			Username:        getStringEnv("DB_USER", "order_fulfillment_user"),
			Password:        getStringEnv("DB_PASSWORD", "order_fulfillment_password"),
			SSLMode:         getStringEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getIntEnv("DB_CONN_MAX_LIFETIME", 300), // 5 minutes
			ConnMaxIdleTime: getIntEnv("DB_CONN_MAX_IDLE_TIME", 30), // 30 seconds
		},
		Redis: RedisConfig{
			Host:         getStringEnv("REDIS_HOST", "localhost"),
			Port:         getIntEnv("REDIS_PORT", 6379),
			Password:     getStringEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
		},
		NATS: NATSConfig{
			URL:       getStringEnv("NATS_URL", "nats://localhost:4222"),
			ClusterID: getStringEnv("NATS_CLUSTER_ID", "order-fulfillment-cluster"),
			ClientID:  getStringEnv("NATS_CLIENT_ID", "order-fulfillment-api"),
		},
		App: AppConfig{
			Environment:    getStringEnv("APP_ENV", "development"),
			LogLevel:       getStringEnv("LOG_LEVEL", "info"),
			GraphQLPath:    getStringEnv("GRAPHQL_PATH", "/graphql"),
			PlaygroundPath: getStringEnv("PLAYGROUND_PATH", "/playground"),
			Version:        getStringEnv("APP_VERSION", "v0.0.1-SNAPSHOT"),
		},
	}
}

// Validate performs basic validation on the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	if c.Database.Username == "" {
		return fmt.Errorf("database username cannot be empty")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	return nil
}

// getStringEnv retrieves a string environment variable with a default fallback
func getStringEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv retrieves an integer environment variable with a default fallback
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}