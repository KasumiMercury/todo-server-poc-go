package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth/mocks"
)

func TestNewAuthenticationService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*gomock.Controller) []auth.AuthenticationStrategy
		expectedError error
		expectedCount int
		expectedNames []string
	}{
		{
			name: "service with secret strategy only",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				mockStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				mockStrategy.EXPECT().Name().Return("Secret").AnyTimes()
				mockStrategy.EXPECT().Priority().Return(100).AnyTimes()
				mockStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()
				return []auth.AuthenticationStrategy{mockStrategy}
			},
			expectedError: nil,
			expectedCount: 1,
			expectedNames: []string{"Secret"},
		},
		{
			name: "service with multiple strategies",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				jwksStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				jwksStrategy.EXPECT().Name().Return("JWKs").AnyTimes()
				jwksStrategy.EXPECT().Priority().Return(200).AnyTimes()
				jwksStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

				secretStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				secretStrategy.EXPECT().Name().Return("Secret").AnyTimes()
				secretStrategy.EXPECT().Priority().Return(100).AnyTimes()
				secretStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

				return []auth.AuthenticationStrategy{jwksStrategy, secretStrategy}
			},
			expectedError: nil,
			expectedCount: 2,
			expectedNames: []string{"JWKs", "Secret"},
		},
		{
			name: "service with no configured strategies",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				return []auth.AuthenticationStrategy{}
			},
			expectedError: auth.ErrNoValidProvider,
			expectedCount: 0,
			expectedNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			strategies := tt.setupMocks(ctrl)

			// Act
			service, err := NewAuthenticationServiceWithProviders(strategies)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, service)
				assert.Equal(t, tt.expectedCount, service.GetProviderCount())
				assert.Equal(t, tt.expectedNames, service.GetConfiguredProviders())
			}
		})
	}
}

func TestAuthenticationService_ValidateToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
	mockStrategy.EXPECT().Name().Return("Secret").AnyTimes()
	mockStrategy.EXPECT().Priority().Return(100).AnyTimes()
	mockStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

	service, err := NewAuthenticationServiceWithProviders([]auth.AuthenticationStrategy{mockStrategy})
	require.NoError(t, err)
	require.NotNil(t, service)

	tests := []struct {
		name             string
		tokenString      string
		setupMock        func(*mocks.MockAuthenticationStrategy)
		expectedValid    bool
		expectedUserID   string
		expectedError    error
		expectedStrategy string
	}{
		{
			name:             "empty token string",
			tokenString:      "",
			setupMock:        func(m *mocks.MockAuthenticationStrategy) {},
			expectedValid:    false,
			expectedUserID:   "",
			expectedError:    auth.ErrInvalidTokenFormat,
			expectedStrategy: "",
		},
		{
			name:        "valid token",
			tokenString: "valid-token",
			setupMock: func(m *mocks.MockAuthenticationStrategy) {
				m.EXPECT().ValidateToken("valid-token").Return(auth.NewTokenValidationResult(true, "test-user", nil))
			},
			expectedValid:    true,
			expectedUserID:   "test-user",
			expectedError:    nil,
			expectedStrategy: "Secret",
		},
		{
			name:        "invalid token",
			tokenString: "invalid-token",
			setupMock: func(m *mocks.MockAuthenticationStrategy) {
				m.EXPECT().ValidateToken("invalid-token").Return(auth.NewTokenValidationResult(false, "", auth.ErrInvalidTokenFormat))
			},
			expectedValid:    false,
			expectedUserID:   "",
			expectedError:    auth.ErrInvalidTokenFormat,
			expectedStrategy: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			tt.setupMock(mockStrategy)

			// Act
			result := service.ValidateToken(tt.tokenString)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedValid, result.IsValid())
			assert.Equal(t, tt.expectedUserID, result.UserID())

			if tt.expectedError != nil {
				assert.ErrorIs(t, result.Error(), tt.expectedError)
			} else if !tt.expectedValid {
				assert.Error(t, result.Error())
			}

			if tt.expectedStrategy != "" {
				assert.Equal(t, tt.expectedStrategy, result.StrategyName())
			}
		})
	}
}

func TestAuthenticationService_ExtractTokenFromHeader(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
	mockStrategy.EXPECT().Name().Return("Secret").AnyTimes()
	mockStrategy.EXPECT().Priority().Return(100).AnyTimes()
	mockStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

	service, err := NewAuthenticationServiceWithProviders([]auth.AuthenticationStrategy{mockStrategy})
	require.NoError(t, err)
	require.NotNil(t, service)

	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectedError error
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token",
			expectedError: nil,
		},
		{
			name:          "bearer token with extra spaces",
			authHeader:    "Bearer   test-token-with-spaces   ",
			expectedToken: "test-token-with-spaces",
			expectedError: nil,
		},
		{
			name:          "empty authorization header",
			authHeader:    "",
			expectedToken: "",
			expectedError: auth.ErrMissingAuthorizationHeader,
		},
		{
			name:          "missing bearer prefix",
			authHeader:    "test-token-without-bearer",
			expectedToken: "",
			expectedError: auth.ErrInvalidAuthorizationFormat,
		},
		{
			name:          "bearer without token",
			authHeader:    "Bearer",
			expectedToken: "",
			expectedError: auth.ErrInvalidAuthorizationFormat,
		},
		{
			name:          "bearer with only spaces",
			authHeader:    "Bearer   ",
			expectedToken: "",
			expectedError: auth.ErrInvalidAuthorizationFormat,
		},
		{
			name:          "case sensitive bearer",
			authHeader:    "bearer token",
			expectedToken: "",
			expectedError: auth.ErrInvalidAuthorizationFormat,
		},
		{
			name:          "bearer with complex token",
			authHeader:    "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.signature",
			expectedToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.signature",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			token, err := service.ExtractTokenFromHeader(tt.authHeader)

			// Assert
			assert.Equal(t, tt.expectedToken, token)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthenticationService_GetConfiguredProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		setupMocks        func(*gomock.Controller) []auth.AuthenticationStrategy
		expectedCount     int
		expectedProviders []string
	}{
		{
			name: "single secret provider",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				mockStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				mockStrategy.EXPECT().Name().Return("Secret").AnyTimes()
				mockStrategy.EXPECT().Priority().Return(100).AnyTimes()
				mockStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()
				return []auth.AuthenticationStrategy{mockStrategy}
			},
			expectedCount:     1,
			expectedProviders: []string{"Secret"},
		},
		{
			name: "multiple providers sorted by priority (mocked)",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				secretStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				secretStrategy.EXPECT().Name().Return("Secret").AnyTimes()
				secretStrategy.EXPECT().Priority().Return(100).AnyTimes()
				secretStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

				jwksStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				jwksStrategy.EXPECT().Name().Return("JWKs").AnyTimes()
				jwksStrategy.EXPECT().Priority().Return(200).AnyTimes()
				jwksStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

				return []auth.AuthenticationStrategy{secretStrategy, jwksStrategy}
			},
			expectedCount:     2,
			expectedProviders: []string{"JWKs", "Secret"}, // JWKs has higher priority (200 > 100)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			strategies := tt.setupMocks(ctrl)

			// Arrange
			service, err := NewAuthenticationServiceWithProviders(strategies)
			require.NoError(t, err)
			require.NotNil(t, service)

			// Act
			providers := service.GetConfiguredProviders()
			count := service.GetProviderCount()

			// Assert
			assert.Equal(t, tt.expectedCount, count)
			assert.Equal(t, tt.expectedProviders, providers)
		})
	}
}

func TestAuthenticationService_GetProviderCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMocks    func(*gomock.Controller) []auth.AuthenticationStrategy
		expectedCount int
		expectError   bool
	}{
		{
			name: "no providers configured",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				return []auth.AuthenticationStrategy{}
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "single provider",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				mockStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				mockStrategy.EXPECT().Name().Return("Secret").AnyTimes()
				mockStrategy.EXPECT().Priority().Return(100).AnyTimes()
				mockStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()
				return []auth.AuthenticationStrategy{mockStrategy}
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "multiple providers (mocked)",
			setupMocks: func(ctrl *gomock.Controller) []auth.AuthenticationStrategy {
				secretStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				secretStrategy.EXPECT().Name().Return("Secret").AnyTimes()
				secretStrategy.EXPECT().Priority().Return(100).AnyTimes()
				secretStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

				jwksStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
				jwksStrategy.EXPECT().Name().Return("JWKs").AnyTimes()
				jwksStrategy.EXPECT().Priority().Return(200).AnyTimes()
				jwksStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

				return []auth.AuthenticationStrategy{secretStrategy, jwksStrategy}
			},
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			strategies := tt.setupMocks(ctrl)

			// Arrange
			service, err := NewAuthenticationServiceWithProviders(strategies)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, auth.ErrNoValidProvider)
				assert.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)

				// Assert
				assert.Equal(t, tt.expectedCount, service.GetProviderCount())
			}
		})
	}
}

func TestAuthenticationService_PriorityOrdering(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create strategies with different priorities to test ordering
	secretStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
	secretStrategy.EXPECT().Name().Return("Secret").AnyTimes()
	secretStrategy.EXPECT().Priority().Return(100).AnyTimes()
	secretStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

	jwksStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
	jwksStrategy.EXPECT().Name().Return("JWKs").AnyTimes()
	jwksStrategy.EXPECT().Priority().Return(200).AnyTimes()
	jwksStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()

	strategies := []auth.AuthenticationStrategy{secretStrategy, jwksStrategy}

	// Act
	service, err := NewAuthenticationServiceWithProviders(strategies)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, service)

	providers := service.GetConfiguredProviders()
	require.Len(t, providers, 2)

	assert.Equal(t, "JWKs", providers[0])
	assert.Equal(t, "Secret", providers[1])
}

func TestAuthenticationService_InitializeProviders_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	strategies := []auth.AuthenticationStrategy{}

	service, err := NewAuthenticationServiceWithProviders(strategies)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNoValidProvider)
	assert.Nil(t, service)
}

func TestAuthenticationService_ValidateToken_AllProvidersFail(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStrategy := mocks.NewMockAuthenticationStrategy(ctrl)
	mockStrategy.EXPECT().Name().Return("Secret").AnyTimes()
	mockStrategy.EXPECT().Priority().Return(100).AnyTimes()
	mockStrategy.EXPECT().IsConfigured().Return(true).AnyTimes()
	mockStrategy.EXPECT().ValidateToken("invalid.jwt.token").Return(auth.NewTokenValidationResult(false, "", auth.ErrInvalidTokenFormat))

	service, err := NewAuthenticationServiceWithProviders([]auth.AuthenticationStrategy{mockStrategy})
	require.NoError(t, err)
	require.NotNil(t, service)

	invalidToken := "invalid.jwt.token"

	// Act
	result := service.ValidateToken(invalidToken)

	// Assert
	require.NotNil(t, result)
	assert.False(t, result.IsValid())
	assert.Empty(t, result.UserID())
	assert.Error(t, result.Error())

	assert.Nil(t, result.Strategy)
}

func TestNewAuthenticationService_WithConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        config.Config
		expectedError error
		expectedCount int
	}{
		{
			name: "service with secret strategy only",
			config: config.Config{
				Auth: config.AuthConfig{
					JWTSecret: "test-secret-key",
				},
			},
			expectedError: nil,
			expectedCount: 1,
		},
		{
			name: "service with no configured strategies",
			config: config.Config{
				Auth: config.AuthConfig{},
			},
			expectedError: auth.ErrNoValidProvider,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			service, err := NewAuthenticationService(tt.config)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, service)
				assert.Equal(t, tt.expectedCount, service.GetProviderCount())
			}
		})
	}
}
