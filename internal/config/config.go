package config

import (
	"os"
	"strings"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type Config struct {
	Database     DatabaseConfig
	JWTSecret    string
	AllowOrigins []string
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
		JWTSecret:    getEnv("JWT_SECRET", "secret-key-for-testing"),
		AllowOrigins: strings.Split(getEnv("ALLOW_ORIGINS", "http://localhost:5173,http://localhsot:3000"), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
