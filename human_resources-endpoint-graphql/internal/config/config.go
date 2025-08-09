package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	GraphQL  GraphQLConfig  `mapstructure:"graphql"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Environment  string        `mapstructure:"environment"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	Name              string        `mapstructure:"name"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	SSLMode           string        `mapstructure:"ssl_mode"`
	MaxOpenConns      int           `mapstructure:"max_open_conns"`
	MaxIdleConns      int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime   time.Duration `mapstructure:"conn_max_idle_time"`
}

// GraphQLConfig holds GraphQL server configuration
type GraphQLConfig struct {
	Playground       bool `mapstructure:"playground"`
	Introspection    bool `mapstructure:"introspection"`
	ComplexityLimit  int  `mapstructure:"complexity_limit"`
	DepthLimit       int  `mapstructure:"depth_limit"`
	EnableDataLoader bool `mapstructure:"enable_data_loader"`
	EnableTracing    bool `mapstructure:"enable_tracing"`
	CacheEnabled     bool `mapstructure:"cache_enabled"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	JWTSecret    string `mapstructure:"jwt_secret"`
	OAuth2       OAuth2Config `mapstructure:"oauth2"`
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level         string `mapstructure:"level"`
	Format        string `mapstructure:"format"`
	EnableConsole bool   `mapstructure:"enable_console"`
}

// GetDSN returns the database connection string
func (db DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode)
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/human-resources-api")

	// Set defaults
	setDefaults()

	// Enable environment variable support
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "human_resources_db")
	viper.SetDefault("database.user", "human_resources_user")
	viper.SetDefault("database.password", "human_resources_password")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "300s")
	viper.SetDefault("database.conn_max_idle_time", "300s")

	// GraphQL defaults
	viper.SetDefault("graphql.playground", true)
	viper.SetDefault("graphql.introspection", true)
	viper.SetDefault("graphql.complexity_limit", 1000)
	viper.SetDefault("graphql.depth_limit", 10)
	viper.SetDefault("graphql.enable_data_loader", true)
	viper.SetDefault("graphql.enable_tracing", true)
	viper.SetDefault("graphql.cache_enabled", true)

	// Auth defaults
	viper.SetDefault("auth.enabled", false)
	viper.SetDefault("auth.jwt_secret", "your-secret-key")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "console")
	viper.SetDefault("logging.enable_console", true)
}