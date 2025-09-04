package config

import (
	"errors"
	"os"
	"testing"
)

func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: DatabaseConfig{
				Host:     "",
				Port:     "5432",
				User:     "user",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "missing port",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "",
				User:     "user",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database port is required",
		},
		{
			name: "missing user",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database user is required",
		},
		{
			name: "missing password",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database password is required",
		},
		{
			name: "missing name",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "password",
				Name:     "",
			},
			wantErr: true,
			errMsg:  "database name is required",
		},
		{
			name: "invalid port - not a number",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "invalid",
				User:     "user",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database port must be a valid number between 1 and 65535: invalid",
		},
		{
			name: "invalid port - zero",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "0",
				User:     "user",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database port must be a valid number between 1 and 65535: 0",
		},
		{
			name: "invalid port - too high",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "65536",
				User:     "user",
				Password: "password",
				Name:     "dbname",
			},
			wantErr: true,
			errMsg:  "database port must be a valid number between 1 and 65535: 65536",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.config.Validate()

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("DatabaseConfig.Validate() expected error, got nil")

					return
				}

				if err.Error() != tt.errMsg {
					t.Errorf("DatabaseConfig.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("DatabaseConfig.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestJWKsConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  JWKsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: JWKsConfig{
				EndpointURL:    "https://example.com/.well-known/jwks.json",
				CacheDuration:  3600,
				RefreshPadding: 300,
			},
			wantErr: false,
		},
		{
			name: "valid config with zero values",
			config: JWKsConfig{
				EndpointURL:    "",
				CacheDuration:  0,
				RefreshPadding: 0,
			},
			wantErr: false,
		},
		{
			name: "negative cache duration",
			config: JWKsConfig{
				EndpointURL:    "https://example.com/.well-known/jwks.json",
				CacheDuration:  -1,
				RefreshPadding: 300,
			},
			wantErr: true,
			errMsg:  "JWKs cache duration must be non-negative",
		},
		{
			name: "negative refresh padding",
			config: JWKsConfig{
				EndpointURL:    "https://example.com/.well-known/jwks.json",
				CacheDuration:  3600,
				RefreshPadding: -1,
			},
			wantErr: true,
			errMsg:  "JWKs refresh padding must be non-negative",
		},
		{
			name: "refresh padding equals cache duration",
			config: JWKsConfig{
				EndpointURL:    "https://example.com/.well-known/jwks.json",
				CacheDuration:  3600,
				RefreshPadding: 3600,
			},
			wantErr: true,
			errMsg:  "JWKs refresh padding must be less than cache duration",
		},
		{
			name: "refresh padding greater than cache duration",
			config: JWKsConfig{
				EndpointURL:    "https://example.com/.well-known/jwks.json",
				CacheDuration:  300,
				RefreshPadding: 600,
			},
			wantErr: true,
			errMsg:  "JWKs refresh padding must be less than cache duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.config.Validate()

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("JWKsConfig.Validate() expected error, got nil")

					return
				}

				if err.Error() != tt.errMsg {
					t.Errorf("JWKsConfig.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("JWKsConfig.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestAuthConfig_Validate(t *testing.T) {
	// Create temporary test file for private key testing
	tmpFile, err := os.CreateTemp("", "test_private_key_*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	tests := []struct {
		name    string
		config  AuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with JWT secret",
			config: AuthConfig{
				JWTSecret: "test-secret",
				JWKs: JWKsConfig{
					EndpointURL:    "",
					CacheDuration:  3600,
					RefreshPadding: 300,
				},
				PrivateKeyFilePath: "",
			},
			wantErr: false,
		},
		{
			name: "valid config with JWKs",
			config: AuthConfig{
				JWTSecret: "",
				JWKs: JWKsConfig{
					EndpointURL:    "https://example.com/.well-known/jwks.json",
					CacheDuration:  3600,
					RefreshPadding: 300,
				},
				PrivateKeyFilePath: "",
			},
			wantErr: false,
		},
		{
			name: "valid config with private key file",
			config: AuthConfig{
				JWTSecret: "",
				JWKs: JWKsConfig{
					EndpointURL:    "",
					CacheDuration:  3600,
					RefreshPadding: 300,
				},
				PrivateKeyFilePath: tmpFile.Name(),
			},
			wantErr: false,
		},
		{
			name: "no auth method configured",
			config: AuthConfig{
				JWTSecret: "",
				JWKs: JWKsConfig{
					EndpointURL:    "",
					CacheDuration:  3600,
					RefreshPadding: 300,
				},
				PrivateKeyFilePath: "",
			},
			wantErr: true,
			errMsg:  "at least one authentication method must be configured: JWT_SECRET, JWKS_ENDPOINT_URL, or JWT_PRIVATE_KEY_FILE",
		},
		{
			name: "invalid JWKs configuration",
			config: AuthConfig{
				JWTSecret: "test-secret",
				JWKs: JWKsConfig{
					EndpointURL:    "",
					CacheDuration:  -1,
					RefreshPadding: 300,
				},
				PrivateKeyFilePath: "",
			},
			wantErr: true,
			errMsg:  "JWKs cache duration must be non-negative",
		},
		{
			name: "private key file does not exist",
			config: AuthConfig{
				JWTSecret: "",
				JWKs: JWKsConfig{
					EndpointURL:    "",
					CacheDuration:  3600,
					RefreshPadding: 300,
				},
				PrivateKeyFilePath: "/nonexistent/path/private.key",
			},
			wantErr: true,
			errMsg:  "private key file does not exist: /nonexistent/path/private.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - test setup already done in table

			// Act
			err := tt.config.Validate()

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("AuthConfig.Validate() expected error, got nil")

					return
				}

				if err.Error() != tt.errMsg {
					t.Errorf("AuthConfig.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("AuthConfig.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_private_key_*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	validDatabaseConfig := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "password",
		Name:     "dbname",
	}

	validAuthConfig := AuthConfig{
		JWTSecret: "test-secret",
		JWKs: JWKsConfig{
			EndpointURL:    "",
			CacheDuration:  3600,
			RefreshPadding: 300,
		},
		PrivateKeyFilePath: "",
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Database:     validDatabaseConfig,
				Auth:         validAuthConfig,
				AllowOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
				ServiceName:  "todo-server",
			},
			wantErr: false,
		},
		{
			name: "empty allow origins",
			config: Config{
				Database:     validDatabaseConfig,
				Auth:         validAuthConfig,
				AllowOrigins: []string{},
				ServiceName:  "todo-server",
			},
			wantErr: false,
		},
		{
			name: "invalid database config",
			config: Config{
				Database: DatabaseConfig{
					Host:     "",
					Port:     "5432",
					User:     "user",
					Password: "password",
					Name:     "dbname",
				},
				Auth:         validAuthConfig,
				AllowOrigins: []string{"http://localhost:3000"},
				ServiceName:  "todo-server",
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "invalid auth config",
			config: Config{
				Database: validDatabaseConfig,
				Auth: AuthConfig{
					JWTSecret: "",
					JWKs: JWKsConfig{
						EndpointURL:    "",
						CacheDuration:  3600,
						RefreshPadding: 300,
					},
					PrivateKeyFilePath: "",
				},
				AllowOrigins: []string{"http://localhost:3000"},
				ServiceName:  "todo-server",
			},
			wantErr: true,
			errMsg:  "at least one authentication method must be configured: JWT_SECRET, JWKS_ENDPOINT_URL, or JWT_PRIVATE_KEY_FILE",
		},
		{
			name: "empty service name",
			config: Config{
				Database:     validDatabaseConfig,
				Auth:         validAuthConfig,
				AllowOrigins: []string{"http://localhost:3000"},
				ServiceName:  "",
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "empty origin in allow origins",
			config: Config{
				Database:     validDatabaseConfig,
				Auth:         validAuthConfig,
				AllowOrigins: []string{"http://localhost:3000", "", "http://localhost:5173"},
				ServiceName:  "todo-server",
			},
			wantErr: true,
			errMsg:  "allowed origin is empty at index 1",
		},
		{
			name: "whitespace only origin in allow origins",
			config: Config{
				Database:     validDatabaseConfig,
				Auth:         validAuthConfig,
				AllowOrigins: []string{"http://localhost:3000", "   ", "http://localhost:5173"},
				ServiceName:  "todo-server",
			},
			wantErr: true,
			errMsg:  "allowed origin is empty at index 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.config.Validate()

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error, got nil")

					return
				}

				if err.Error() != tt.errMsg {
					t.Errorf("Config.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		setup   func() func()
		wantErr bool
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"DB_HOST":              "",
				"DB_PORT":              "",
				"DB_USER":              "",
				"DB_PASSWORD":          "",
				"DB_NAME":              "",
				"JWT_SECRET":           "default-secret",
				"JWKS_ENDPOINT_URL":    "",
				"JWKS_CACHE_DURATION":  "",
				"JWKS_REFRESH_PADDING": "",
				"JWT_PRIVATE_KEY_FILE": "",
				"ALLOW_ORIGINS":        "",
				"SERVICE_NAME":         "",
				"PORT":                 "",
				"METRICS_PORT":         "",
			},
			wantErr: false,
		},
		{
			name: "custom environment variables",
			envVars: map[string]string{
				"DB_HOST":              "custom-host",
				"DB_PORT":              "3306",
				"DB_USER":              "custom-user",
				"DB_PASSWORD":          "custom-password",
				"DB_NAME":              "custom-db",
				"JWT_SECRET":           "custom-secret",
				"JWKS_CACHE_DURATION":  "7200",
				"JWKS_REFRESH_PADDING": "600",
				"ALLOW_ORIGINS":        "http://example.com,https://app.example.com",
				"SERVICE_NAME":         "custom-service",
				"PORT":                 "9000",
				"METRICS_PORT":         "9001",
			},
			wantErr: false,
		},
		{
			name: "invalid port in environment",
			envVars: map[string]string{
				"DB_HOST":      "localhost",
				"DB_PORT":      "invalid-port",
				"DB_USER":      "user",
				"DB_PASSWORD":  "password",
				"DB_NAME":      "dbname",
				"JWT_SECRET":   "secret",
				"SERVICE_NAME": "todo-server",
				"PORT":         "8080",
				"METRICS_PORT": "8081",
			},
			wantErr: true,
		},
		{
			name: "missing authentication method",
			envVars: map[string]string{
				"DB_HOST":              "localhost",
				"DB_PORT":              "5432",
				"DB_USER":              "user",
				"DB_PASSWORD":          "password",
				"DB_NAME":              "dbname",
				"JWT_SECRET":           "",
				"JWKS_ENDPOINT_URL":    "",
				"JWT_PRIVATE_KEY_FILE": "",
				"SERVICE_NAME":         "todo-server",
				"PORT":                 "8080",
				"METRICS_PORT":         "8081",
			},
			wantErr: true,
		},
		{
			name: "invalid JWKs cache duration",
			envVars: map[string]string{
				"DB_HOST":             "localhost",
				"DB_PORT":             "5432",
				"DB_USER":             "user",
				"DB_PASSWORD":         "password",
				"DB_NAME":             "dbname",
				"JWT_SECRET":          "secret",
				"JWKS_CACHE_DURATION": "-100",
				"SERVICE_NAME":        "todo-server",
				"PORT":                "8080",
				"METRICS_PORT":        "8081",
			},
			wantErr: true,
		},
		{
			name: "nonexistent private key file",
			envVars: map[string]string{
				"DB_HOST":              "localhost",
				"DB_PORT":              "5432",
				"DB_USER":              "user",
				"DB_PASSWORD":          "password",
				"DB_NAME":              "dbname",
				"JWT_SECRET":           "",
				"JWT_PRIVATE_KEY_FILE": "/nonexistent/private.key",
				"SERVICE_NAME":         "todo-server",
				"PORT":                 "8080",
				"METRICS_PORT":         "8081",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			originalEnv := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			defer func() {
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup()
				defer cleanup()
			}

			// Act & Assert
			if tt.wantErr {
				config, err := Load()
				if err == nil {
					t.Errorf("Load() expected error, got nil")
				}

				if config != nil {
					t.Errorf("Load() expected nil config on error, got %+v", config)
				}

				if !errors.Is(err, ErrConfigValidation) {
					t.Errorf("Load() error should wrap ErrConfigValidation, got %v", err)
				}
			} else {
				config, err := Load()
				if err != nil {
					t.Errorf("Load() unexpected error: %v", err)
				}

				if config == nil {
					t.Errorf("Load() returned nil config")
				}

				if config.ServiceName == "" {
					t.Errorf("Load() ServiceName is empty")
				}

				if len(config.AllowOrigins) == 0 {
					t.Errorf("Load() AllowOrigins is empty")
				}

				if config.Database.Host == "" {
					t.Errorf("Load() Database.Host is empty")
				}

				if config.Port == "" {
					t.Errorf("Load() Port is empty")
				}

				if config.MetricsPort == "" {
					t.Errorf("Load() MetricsPort is empty")
				}
			}
		})
	}
}

func TestLoadPortConfiguration(t *testing.T) {
	tests := []struct {
		name                string
		envVars             map[string]string
		expectedPort        string
		expectedMetricsPort string
	}{
		{
			name: "default port values",
			envVars: map[string]string{
				"JWT_SECRET":   "test-secret",
				"PORT":         "",
				"METRICS_PORT": "",
			},
			expectedPort:        "8080",
			expectedMetricsPort: "8081",
		},
		{
			name: "custom port values",
			envVars: map[string]string{
				"JWT_SECRET":   "test-secret",
				"PORT":         "9000",
				"METRICS_PORT": "9001",
			},
			expectedPort:        "9000",
			expectedMetricsPort: "9001",
		},
		{
			name: "only PORT set",
			envVars: map[string]string{
				"JWT_SECRET":   "test-secret",
				"PORT":         "3000",
				"METRICS_PORT": "",
			},
			expectedPort:        "3000",
			expectedMetricsPort: "8081",
		},
		{
			name: "only METRICS_PORT set",
			envVars: map[string]string{
				"JWT_SECRET":   "test-secret",
				"PORT":         "",
				"METRICS_PORT": "4000",
			},
			expectedPort:        "8080",
			expectedMetricsPort: "4000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			originalEnv := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			defer func() {
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			// Act
			config, err := Load()

			// Assert
			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
			}

			if config == nil {
				t.Errorf("Load() returned nil config")

				return
			}

			if config.Port != tt.expectedPort {
				t.Errorf("Load() Port = %v, want %v", config.Port, tt.expectedPort)
			}

			if config.MetricsPort != tt.expectedMetricsPort {
				t.Errorf("Load() MetricsPort = %v, want %v", config.MetricsPort, tt.expectedMetricsPort)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable is empty",
			key:          "EMPTY_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			originalValue := os.Getenv(tt.key)

			defer func() {
				if originalValue == "" {
					os.Unsetenv(tt.key)
				} else {
					os.Setenv(tt.key, originalValue)
				}
			}()

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			// Act
			result := getEnv(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("getEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid integer environment variable",
			key:          "TEST_INT_KEY",
			defaultValue: 100,
			envValue:     "200",
			expected:     200,
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_INT_KEY",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
		{
			name:         "invalid integer environment variable",
			key:          "INVALID_INT_KEY",
			defaultValue: 100,
			envValue:     "invalid",
			expected:     100,
		},
		{
			name:         "empty environment variable",
			key:          "EMPTY_INT_KEY",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
		{
			name:         "zero value environment variable",
			key:          "ZERO_INT_KEY",
			defaultValue: 100,
			envValue:     "0",
			expected:     0,
		},
		{
			name:         "negative integer environment variable",
			key:          "NEGATIVE_INT_KEY",
			defaultValue: 100,
			envValue:     "-50",
			expected:     -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			originalValue := os.Getenv(tt.key)

			defer func() {
				if originalValue == "" {
					os.Unsetenv(tt.key)
				} else {
					os.Setenv(tt.key, originalValue)
				}
			}()

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			// Act
			result := getIntEnv(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("getIntEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}
