package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrDatabaseHostRequired     = errors.New("database host is required")
	ErrDatabasePortRequired     = errors.New("database port is required")
	ErrDatabaseUserRequired     = errors.New("database user is required")
	ErrDatabasePasswordRequired = errors.New("database password is required")
	ErrDatabaseNameRequired     = errors.New("database name is required")
	ErrDatabasePortInvalid      = errors.New("database port must be a valid number between 1 and 65535")

	ErrJWKsCacheDurationNegative  = errors.New("JWKs cache duration must be non-negative")
	ErrJWKsRefreshPaddingNegative = errors.New("JWKs refresh padding must be non-negative")
	ErrJWKsRefreshPaddingTooLarge = errors.New("JWKs refresh padding must be less than cache duration")

	ErrAuthMethodRequired     = errors.New("at least one authentication method must be configured: JWT_SECRET, JWKS_ENDPOINT_URL, or JWT_PRIVATE_KEY_FILE")
	ErrPrivateKeyFileNotFound = errors.New("private key file does not exist")

	ErrServiceNameRequired = errors.New("service name is required")
	ErrAllowOriginEmpty    = errors.New("allowed origin is empty")

	ErrConfigValidation = errors.New("configuration validation failed")
)

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// Validate validates the database configuration
func (dc DatabaseConfig) Validate() error {
	if dc.Host == "" {
		return ErrDatabaseHostRequired
	}

	if dc.Port == "" {
		return ErrDatabasePortRequired
	}

	if dc.User == "" {
		return ErrDatabaseUserRequired
	}

	if dc.Password == "" {
		return ErrDatabasePasswordRequired
	}

	if dc.Name == "" {
		return ErrDatabaseNameRequired
	}

	if port, err := strconv.Atoi(dc.Port); err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("%w: %s", ErrDatabasePortInvalid, dc.Port)
	}

	return nil
}

// JWKsConfig holds JSON Web Key Set configuration for JWT validation.
type JWKsConfig struct {
	EndpointURL    string
	CacheDuration  int // seconds
	RefreshPadding int // seconds
}

// Validate validates the JWKs configuration
func (jc JWKsConfig) Validate() error {
	if jc.CacheDuration < 0 {
		return ErrJWKsCacheDurationNegative
	}

	if jc.RefreshPadding < 0 {
		return ErrJWKsRefreshPaddingNegative
	}

	if jc.RefreshPadding >= jc.CacheDuration && jc.CacheDuration > 0 {
		return ErrJWKsRefreshPaddingTooLarge
	}

	return nil
}

// AuthConfig holds authentication and JWT configuration.
type AuthConfig struct {
	JWTSecret          string
	JWKs               JWKsConfig
	PrivateKeyFilePath string
}

// Validate validates the auth configuration
func (ac AuthConfig) Validate() error {
	// Validate JWKs configuration
	if err := ac.JWKs.Validate(); err != nil {
		return err
	}

	// At least one authentication method must be configured
	hasJWTSecret := ac.JWTSecret != ""
	hasJWKs := ac.JWKs.EndpointURL != ""
	hasPrivateKey := ac.PrivateKeyFilePath != ""

	if !hasJWTSecret && !hasJWKs && !hasPrivateKey {
		return ErrAuthMethodRequired
	}

	if hasPrivateKey {
		if _, err := os.Stat(ac.PrivateKeyFilePath); os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrPrivateKeyFileNotFound, ac.PrivateKeyFilePath)
		}
	}

	return nil
}

// Config represents the application configuration loaded from environment variables.
type Config struct {
	Database     DatabaseConfig
	Auth         AuthConfig
	AllowOrigins []string
	ServiceName  string
}

// Validate validates the entire configuration
func (c Config) Validate() error {
	// Validate database configuration
	if err := c.Database.Validate(); err != nil {
		return err
	}

	// Validate auth configuration
	if err := c.Auth.Validate(); err != nil {
		return err
	}

	if c.ServiceName == "" {
		return ErrServiceNameRequired
	}

	for i, origin := range c.AllowOrigins {
		if strings.TrimSpace(origin) == "" {
			return fmt.Errorf("%w at index %d", ErrAllowOriginEmpty, i)
		}
	}

	return nil
}

// Load creates and returns a new Config instance with values loaded from environment variables.
func Load() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "taskdb"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", ""),
			JWKs: JWKsConfig{
				EndpointURL:    getEnv("JWKS_ENDPOINT_URL", ""),
				CacheDuration:  getIntEnv("JWKS_CACHE_DURATION", 3600), // 1 hour
				RefreshPadding: getIntEnv("JWKS_REFRESH_PADDING", 300), // 5 minutes
			},
			PrivateKeyFilePath: getEnv("JWT_PRIVATE_KEY_FILE", ""),
		},
		AllowOrigins: strings.Split(getEnv("ALLOW_ORIGINS", "http://localhost:5173,http://localhost:3000"), ","),
		ServiceName:  getEnv("SERVICE_NAME", "todo-server"),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigValidation, err)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return defaultValue
}
