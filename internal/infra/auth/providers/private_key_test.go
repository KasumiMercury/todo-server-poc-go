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

func TestNewPrivateKeyStrategy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		config             config.Config
		expectedName       string
		expectedPriority   int
		expectedConfigured bool
		expectError        bool
	}{
		{
			name: "unconfigured private key strategy (empty path)",
			config: config.Config{
				Auth: config.AuthConfig{
					PrivateKeyFilePath: "",
				},
			},
			expectedName:       PrivateKeyStrategyName,
			expectedPriority:   PrivateKeyPriority,
			expectedConfigured: false,
			expectError:        false,
		},
		{
			name: "configured private key strategy with invalid path",
			config: config.Config{
				Auth: config.AuthConfig{
					PrivateKeyFilePath: "/nonexistent/path/to/key.pem",
				},
			},
			expectedName:       PrivateKeyStrategyName,
			expectedPriority:   PrivateKeyPriority,
			expectedConfigured: false,
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			strategy, err := NewPrivateKeyStrategy(tt.config)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				// Strategy might be nil or might have been created before the error
			} else {
				assert.NoError(t, err)
				require.NotNil(t, strategy)
				assert.Equal(t, tt.expectedName, strategy.Name())
				assert.Equal(t, tt.expectedPriority, strategy.Priority())
				assert.Equal(t, tt.expectedConfigured, strategy.IsConfigured())

				if !tt.expectedConfigured {
					assert.Equal(t, auth.KeyFormatUnknown, strategy.GetKeyFormat())
				}
			}
		})
	}
}

func TestPrivateKeyStrategy_ValidateToken_NotConfigured(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		Auth: config.AuthConfig{
			PrivateKeyFilePath: "",
		},
	}

	strategy, err := NewPrivateKeyStrategy(cfg)
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

func TestPrivateKeyStrategy_ValidateToken_ValidToken(t *testing.T) {
	t.Parallel()

	// Generate a test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	loadedKey := auth.NewLoadedPrivateKey(privateKey, auth.KeyFormatRSAPEM)
	strategy := &PrivateKeyStrategy{
		name:             PrivateKeyStrategyName,
		configured:       true,
		loadedPrivateKey: loadedKey,
	}

	claims := jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	validTokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Act
	result := strategy.ValidateToken(validTokenString)

	// Assert
	require.NotNil(t, result)
	assert.True(t, result.IsValid())
	assert.Equal(t, "user123", result.UserID())
	assert.NoError(t, result.Error())
}

func TestPrivateKeyStrategy_ValidateToken_InvalidTokens(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	loadedKey := auth.NewLoadedPrivateKey(privateKey, auth.KeyFormatRSAPEM)
	strategy := &PrivateKeyStrategy{
		name:             PrivateKeyStrategyName,
		configured:       true,
		loadedPrivateKey: loadedKey,
	}

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
			name:           "token with wrong signing method (HMAC)",
			tokenString:    createHMACTokenWithSecret(t, "secret", "user456"),
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:           "token signed with different RSA key",
			tokenString:    createTokenWithDifferentRSAKey(t, "user789"),
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:           "expired RSA token",
			tokenString:    createExpiredRSAToken(t, privateKey, "user101"),
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

func TestPrivateKeyStrategy_ValidateToken_TokenWithoutSubClaim(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	loadedKey := auth.NewLoadedPrivateKey(privateKey, auth.KeyFormatRSAPEM)
	strategy := &PrivateKeyStrategy{
		name:             PrivateKeyStrategyName,
		configured:       true,
		loadedPrivateKey: loadedKey,
	}

	// Create a valid token without 'sub' claim
	claims := jwt.MapClaims{
		"exp":          time.Now().Add(time.Hour).Unix(),
		"iat":          time.Now().Unix(),
		"custom_claim": "custom_value",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Act
	result := strategy.ValidateToken(tokenString)

	// Assert
	require.NotNil(t, result)
	assert.True(t, result.IsValid())
	assert.Empty(t, result.UserID()) // Should be empty since no 'sub' claim
	assert.NoError(t, result.Error())
}

func TestPrivateKeyStrategy_Methods(t *testing.T) {
	t.Parallel()

	t.Run("unconfigured strategy", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Auth: config.AuthConfig{
				PrivateKeyFilePath: "",
			},
		}

		strategy, err := NewPrivateKeyStrategy(cfg)
		require.NoError(t, err)
		require.NotNil(t, strategy)

		// Act & Assert
		assert.Equal(t, PrivateKeyStrategyName, strategy.Name())
		assert.Equal(t, PrivateKeyPriority, strategy.Priority())
		assert.False(t, strategy.IsConfigured())
		assert.Equal(t, auth.KeyFormatUnknown, strategy.GetKeyFormat())
	})

	t.Run("configured strategy", func(t *testing.T) {
		t.Parallel()

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		loadedKey := auth.NewLoadedPrivateKey(privateKey, auth.KeyFormatPKCS8PEM)
		strategy := &PrivateKeyStrategy{
			name:             PrivateKeyStrategyName,
			configured:       true,
			loadedPrivateKey: loadedKey,
		}

		// Act & Assert
		assert.Equal(t, PrivateKeyStrategyName, strategy.Name())
		assert.Equal(t, PrivateKeyPriority, strategy.Priority())
		assert.True(t, strategy.IsConfigured())
		assert.Equal(t, auth.KeyFormatPKCS8PEM, strategy.GetKeyFormat())
	})
}

func TestPrivateKeyStrategy_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "PrivateKey", PrivateKeyStrategyName)
	assert.Equal(t, 300, PrivateKeyPriority)
}

func TestPrivateKeyStrategy_GetKeyFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		loadedKey      *auth.LoadedPrivateKey
		expectedFormat auth.KeyFormat
	}{
		{
			name:           "nil loaded key returns unknown format",
			loadedKey:      nil,
			expectedFormat: auth.KeyFormatUnknown,
		},
		{
			name: "RSA PEM format",
			loadedKey: func() *auth.LoadedPrivateKey {
				key, _ := rsa.GenerateKey(rand.Reader, 2048)

				return auth.NewLoadedPrivateKey(key, auth.KeyFormatRSAPEM)
			}(),
			expectedFormat: auth.KeyFormatRSAPEM,
		},
		{
			name: "PKCS8 PEM format",
			loadedKey: func() *auth.LoadedPrivateKey {
				key, _ := rsa.GenerateKey(rand.Reader, 2048)

				return auth.NewLoadedPrivateKey(key, auth.KeyFormatPKCS8PEM)
			}(),
			expectedFormat: auth.KeyFormatPKCS8PEM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			strategy := &PrivateKeyStrategy{
				name:             PrivateKeyStrategyName,
				configured:       tt.loadedKey != nil,
				loadedPrivateKey: tt.loadedKey,
			}

			// Act
			format := strategy.GetKeyFormat()

			// Assert
			assert.Equal(t, tt.expectedFormat, format)
		})
	}
}

func TestPrivateKeyStrategy_ConfiguredState_Consistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured bool
		hasKey     bool
	}{
		{
			name:       "unconfigured strategy has no key",
			configured: false,
			hasKey:     false,
		},
		{
			name:       "configured strategy has key",
			configured: true,
			hasKey:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var loadedKey *auth.LoadedPrivateKey

			if tt.hasKey {
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)

				loadedKey = auth.NewLoadedPrivateKey(privateKey, auth.KeyFormatRSAPEM)
			}

			// Arrange
			strategy := &PrivateKeyStrategy{
				name:             PrivateKeyStrategyName,
				configured:       tt.configured,
				loadedPrivateKey: loadedKey,
			}

			// Assert consistency
			assert.Equal(t, tt.configured, strategy.IsConfigured())

			if tt.hasKey {
				assert.NotEqual(t, auth.KeyFormatUnknown, strategy.GetKeyFormat())
			} else {
				assert.Equal(t, auth.KeyFormatUnknown, strategy.GetKeyFormat())
			}

			result := strategy.ValidateToken("test-token")
			require.NotNil(t, result)

			if !tt.configured {
				assert.ErrorIs(t, result.Error(), auth.ErrProviderNotConfigured)
			} else {
				if result.Error() != nil {
					assert.NotErrorIs(t, result.Error(), auth.ErrProviderNotConfigured)
				}
			}
		})
	}
}

// Helper functions for creating test tokens

func createTokenWithDifferentRSAKey(t *testing.T, userID string) string {
	// Generate a different RSA key
	differentKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(differentKey)
	require.NoError(t, err)

	return tokenString
}

func createExpiredRSAToken(t *testing.T, privateKey *rsa.PrivateKey, userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	return tokenString
}
