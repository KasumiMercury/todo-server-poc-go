package providers

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

func TestNewSecretStrategy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		config             config.Config
		expectedName       string
		expectedPriority   int
		expectedConfigured bool
		expectedSecretKey  string
	}{
		{
			name: "configured secret strategy",
			config: config.Config{
				Auth: config.AuthConfig{
					JWTSecret: "test-secret-key",
				},
			},
			expectedName:       SecretStrategyName,
			expectedPriority:   SecretPriority,
			expectedConfigured: true,
			expectedSecretKey:  "test-secret-key",
		},
		{
			name: "unconfigured secret strategy",
			config: config.Config{
				Auth: config.AuthConfig{
					JWTSecret: "",
				},
			},
			expectedName:       SecretStrategyName,
			expectedPriority:   SecretPriority,
			expectedConfigured: false,
			expectedSecretKey:  "",
		},
		{
			name: "strategy with complex secret key",
			config: config.Config{
				Auth: config.AuthConfig{
					JWTSecret: "very-complex-secret-key-with-special-chars!@#$%^&*()",
				},
			},
			expectedName:       SecretStrategyName,
			expectedPriority:   SecretPriority,
			expectedConfigured: true,
			expectedSecretKey:  "very-complex-secret-key-with-special-chars!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			strategy, err := NewSecretStrategy(tt.config)

			// Assert
			assert.NoError(t, err)
			require.NotNil(t, strategy)
			assert.Equal(t, tt.expectedName, strategy.Name())
			assert.Equal(t, tt.expectedPriority, strategy.Priority())
			assert.Equal(t, tt.expectedConfigured, strategy.IsConfigured())
			assert.Equal(t, tt.expectedSecretKey, strategy.GetSecretKey())
		})
	}
}

func TestSecretStrategy_ValidateToken_NotConfigured(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "",
		},
	}

	strategy, err := NewSecretStrategy(cfg)
	require.NoError(t, err)
	require.NotNil(t, strategy)

	// Act
	result := strategy.ValidateToken("any-token")

	// Assert
	require.NotNil(t, result)
	assert.False(t, result.IsValid())
	assert.Empty(t, result.UserID())
	assert.ErrorIs(t, result.Error(), auth.ErrProviderNotConfigured)
}

func TestSecretStrategy_ValidateToken_ValidToken(t *testing.T) {
	t.Parallel()

	// Arrange
	secretKey := "test-secret-key"
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret: secretKey,
		},
	}

	strategy, err := NewSecretStrategy(cfg)
	require.NoError(t, err)
	require.NotNil(t, strategy)

	claims := jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validTokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// Act
	result := strategy.ValidateToken(validTokenString)

	// Assert
	require.NotNil(t, result)
	assert.True(t, result.IsValid())
	assert.Equal(t, "user123", result.UserID())
	assert.NoError(t, result.Error())
}

func TestSecretStrategy_ValidateToken_InvalidTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	secretKey := "test-secret-key"
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret: secretKey,
		},
	}

	strategy, err := NewSecretStrategy(cfg)
	require.NoError(t, err)
	require.NotNil(t, strategy)

	tests := []struct {
		name           string
		tokenString    string
		expectedValid  bool
		expectedUserID string
		expectError    bool
	}{
		{
			name:           "malformed token",
			tokenString:    "invalid-token-format",
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:           "empty token",
			tokenString:    "",
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:           "token with wrong signing method",
			tokenString:    createRSAToken(t),
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:           "token signed with wrong secret",
			tokenString:    createHMACTokenWithSecret(t, "wrong-secret", "user456"),
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:           "expired token",
			tokenString:    createExpiredHMACToken(t, secretKey, "user789"),
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := strategy.ValidateToken(tt.tokenString)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedValid, result.IsValid())
			assert.Equal(t, tt.expectedUserID, result.UserID())

			if tt.expectError {
				assert.Error(t, result.Error())
			} else {
				assert.NoError(t, result.Error())
			}
		})
	}
}

func TestSecretStrategy_ValidateToken_TokenWithoutSubClaim(t *testing.T) {
	t.Parallel()

	// Arrange
	secretKey := "test-secret-key"
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret: secretKey,
		},
	}

	strategy, err := NewSecretStrategy(cfg)
	require.NoError(t, err)
	require.NotNil(t, strategy)

	claims := jwt.MapClaims{
		"exp":          time.Now().Add(time.Hour).Unix(),
		"iat":          time.Now().Unix(),
		"custom_claim": "custom_value",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// Act
	result := strategy.ValidateToken(tokenString)

	// Assert
	require.NotNil(t, result)
	assert.True(t, result.IsValid())
	assert.Empty(t, result.UserID())
	assert.NoError(t, result.Error())
}

func TestSecretStrategy_Methods(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "test-secret",
		},
	}

	strategy, err := NewSecretStrategy(cfg)
	require.NoError(t, err)
	require.NotNil(t, strategy)

	// Act & Assert
	assert.Equal(t, SecretStrategyName, strategy.Name())
	assert.Equal(t, SecretPriority, strategy.Priority())
	assert.True(t, strategy.IsConfigured())
	assert.Equal(t, "test-secret", strategy.GetSecretKey())
}

func TestSecretStrategy_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Secret", SecretStrategyName)
	assert.Equal(t, 100, SecretPriority)
}

func createHMACTokenWithSecret(t *testing.T, secret, userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	return tokenString
}

func createExpiredHMACToken(t *testing.T, secret, userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	return tokenString
}

func createRSAToken(t *testing.T) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	return tokenString
}
