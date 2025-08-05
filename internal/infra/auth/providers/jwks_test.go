package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth/providers/mocks"
)

func TestNewJWKsStrategy(t *testing.T) {
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
			name: "configured JWKs strategy",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "https://example.com/.well-known/jwks.json",
						CacheDuration:  3600,
						RefreshPadding: 300,
					},
				},
			},
			expectedName:       JWKsStrategyName,
			expectedPriority:   JWKsPriority,
			expectedConfigured: true,
			expectError:        false,
		},
		{
			name: "unconfigured JWKs strategy (empty endpoint)",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "",
						CacheDuration:  3600,
						RefreshPadding: 300,
					},
				},
			},
			expectedName:       JWKsStrategyName,
			expectedPriority:   JWKsPriority,
			expectedConfigured: false,
			expectError:        false,
		},
		{
			name: "JWKs strategy with different cache settings",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "https://auth.example.com/oauth2/v1/keys",
						CacheDuration:  7200,
						RefreshPadding: 600,
					},
				},
			},
			expectedName:       JWKsStrategyName,
			expectedPriority:   JWKsPriority,
			expectedConfigured: true,
			expectError:        false,
		},
		{
			name: "JWKs strategy with zero cache duration",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "https://test.com/jwks",
						CacheDuration:  0,
						RefreshPadding: 0,
					},
				},
			},
			expectedName:       JWKsStrategyName,
			expectedPriority:   JWKsPriority,
			expectedConfigured: true,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			var (
				strategy *JWKsStrategy
				err      error
			)

			if tt.config.Auth.JWKs.EndpointURL != "" && tt.expectedConfigured {
				if tt.config.Auth.JWKs.EndpointURL != "" {
					strategy = NewJWKsStrategyWithValidator(mocks.NewMockJWKSValidator(gomock.NewController(t)))
				} else {
					strategy = NewJWKsStrategyWithValidator(nil)
				}

				err = nil
			} else {
				strategy, err = NewJWKsStrategy(tt.config)
			}

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, strategy)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, strategy)
				assert.Equal(t, tt.expectedName, strategy.Name())
				assert.Equal(t, tt.expectedPriority, strategy.Priority())
				assert.Equal(t, tt.expectedConfigured, strategy.IsConfigured())

				if tt.expectedConfigured {
					assert.NotNil(t, strategy.GetValidator())
				} else {
					assert.Nil(t, strategy.GetValidator())
				}
			}
		})
	}
}

func TestJWKsStrategy_ValidateToken_NotConfigured(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWKs: config.JWKsConfig{
				EndpointURL: "",
			},
		},
	}

	strategy, err := NewJWKsStrategy(cfg)
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

func TestJWKsStrategy_ValidateToken_Configured(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		tokenString      string
		mockValidateFunc func(string) *auth.TokenValidationResult
		expectedValid    bool
		expectedUserID   string
		expectError      bool
	}{
		{
			name:        "valid token",
			tokenString: "valid.jwt.token",
			mockValidateFunc: func(token string) *auth.TokenValidationResult {
				return auth.NewTokenValidationResult(true, "user123", nil)
			},
			expectedValid:  true,
			expectedUserID: "user123",
			expectError:    false,
		},
		{
			name:        "invalid token",
			tokenString: "invalid.jwt.token",
			mockValidateFunc: func(token string) *auth.TokenValidationResult {
				return auth.NewTokenValidationResult(false, "", auth.ErrTokenValidation)
			},
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
		{
			name:        "token with JWKs client error",
			tokenString: "problematic.jwt.token",
			mockValidateFunc: func(token string) *auth.TokenValidationResult {
				return auth.NewTokenValidationResult(false, "", auth.ErrJWKsClientError)
			},
			expectedValid:  false,
			expectedUserID: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockValidator := mocks.NewMockJWKSValidator(ctrl)
			mockValidator.EXPECT().ValidateToken(tt.tokenString).Return(tt.mockValidateFunc(tt.tokenString))
			strategy := NewJWKsStrategyWithValidator(mockValidator)

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

func TestJWKsStrategy_Methods(t *testing.T) {
	t.Parallel()

	t.Run("configured strategy", func(t *testing.T) {
		t.Parallel()

		// Arrange
		strategy := NewJWKsStrategyWithValidator(mocks.NewMockJWKSValidator(gomock.NewController(t)))

		// Act & Assert
		assert.Equal(t, JWKsStrategyName, strategy.Name())
		assert.Equal(t, JWKsPriority, strategy.Priority())
		assert.True(t, strategy.IsConfigured())
		assert.NotNil(t, strategy.GetValidator())
	})

	t.Run("unconfigured strategy", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Auth: config.AuthConfig{
				JWKs: config.JWKsConfig{
					EndpointURL: "",
				},
			},
		}

		strategy, err := NewJWKsStrategy(cfg)
		require.NoError(t, err)
		require.NotNil(t, strategy)

		// Act & Assert
		assert.Equal(t, JWKsStrategyName, strategy.Name())
		assert.Equal(t, JWKsPriority, strategy.Priority())
		assert.False(t, strategy.IsConfigured())
		assert.Nil(t, strategy.GetValidator())
	})
}

func TestJWKsStrategy_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "JWKs", JWKsStrategyName)
	assert.Equal(t, 200, JWKsPriority)
}

func TestJWKsStrategy_GetClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		config          config.Config
		expectClientNil bool
	}{
		{
			name: "configured strategy has non-nil client",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "https://example.com/.well-known/jwks.json",
						CacheDuration:  3600,
						RefreshPadding: 300,
					},
				},
			},
			expectClientNil: false,
		},
		{
			name: "unconfigured strategy has nil client",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL: "",
					},
				},
			},
			expectClientNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				strategy *JWKsStrategy
				err      error
			)

			if !tt.expectClientNil {
				if tt.config.Auth.JWKs.EndpointURL != "" {
					strategy = NewJWKsStrategyWithValidator(mocks.NewMockJWKSValidator(gomock.NewController(t)))
				} else {
					strategy = NewJWKsStrategyWithValidator(nil)
				}

				err = nil
			} else {
				strategy, err = NewJWKsStrategy(tt.config)
			}

			require.NoError(t, err)
			require.NotNil(t, strategy)

			// Act
			client := strategy.GetValidator()

			// Assert
			if tt.expectClientNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}

func TestJWKsStrategy_ConfiguredState_Consistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     config.Config
		configured bool
	}{
		{
			name: "configured with valid endpoint",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "https://auth.example.com/jwks",
						CacheDuration:  1800,
						RefreshPadding: 150,
					},
				},
			},
			configured: true,
		},
		{
			name: "not configured with empty endpoint",
			config: config.Config{
				Auth: config.AuthConfig{
					JWKs: config.JWKsConfig{
						EndpointURL:    "",
						CacheDuration:  3600,
						RefreshPadding: 300,
					},
				},
			},
			configured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				strategy *JWKsStrategy
				err      error
			)

			if tt.configured {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockValidator := mocks.NewMockJWKSValidator(ctrl)
				mockValidator.EXPECT().ValidateToken("test-token").Return(auth.NewTokenValidationResult(true, "test-user", nil)).AnyTimes()
				strategy = NewJWKsStrategyWithValidator(mockValidator)
				err = nil
			} else {
				strategy, err = NewJWKsStrategy(tt.config)
			}

			require.NoError(t, err)
			require.NotNil(t, strategy)

			assert.Equal(t, tt.configured, strategy.IsConfigured())

			if tt.configured {
				assert.NotNil(t, strategy.GetValidator())
			} else {
				assert.Nil(t, strategy.GetValidator())
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

func TestNewJWKsStrategyWithValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		validator          JWKSValidator
		expectedConfigured bool
	}{
		{
			name:               "configured with validator",
			validator:          mocks.NewMockJWKSValidator(gomock.NewController(t)),
			expectedConfigured: true,
		},
		{
			name:               "unconfigured with nil validator",
			validator:          nil,
			expectedConfigured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			strategy := NewJWKsStrategyWithValidator(tt.validator)

			// Assert
			require.NotNil(t, strategy)
			assert.Equal(t, tt.expectedConfigured, strategy.IsConfigured())
			assert.Equal(t, JWKsStrategyName, strategy.Name())
			assert.Equal(t, JWKsPriority, strategy.Priority())

			if tt.expectedConfigured {
				assert.NotNil(t, strategy.GetValidator())
			} else {
				assert.Nil(t, strategy.GetValidator())
			}
		})
	}
}
