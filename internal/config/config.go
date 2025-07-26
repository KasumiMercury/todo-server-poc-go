package config

import (
	"os"
	"strconv"
	"strings"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWKsConfig struct {
	EndpointURL    string
	CacheDuration  int // seconds
	RefreshPadding int // seconds
}

type Config struct {
	Database           DatabaseConfig
	JWTSecret          string
	JWKs               JWKsConfig
	PrivateKeyFilePath string
	AllowOrigins       []string
	ServiceName        string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "taskdb"),
		},
		JWTSecret: getEnv("JWT_SECRET", "secret-key-for-testing"),
		JWKs: JWKsConfig{
			EndpointURL:    getEnv("JWKS_ENDPOINT_URL", ""),
			CacheDuration:  getIntEnv("JWKS_CACHE_DURATION", 3600), // 1 hour
			RefreshPadding: getIntEnv("JWKS_REFRESH_PADDING", 300), // 5 minutes
		},
		PrivateKeyFilePath: getEnv("JWT_PRIVATE_KEY_FILE", ""),
		AllowOrigins:       strings.Split(getEnv("ALLOW_ORIGINS", "http://localhost:5173,http://localhsot:3000"), ","),
		ServiceName:        getEnv("SERVICE_NAME", "todo-server"),
	}
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
