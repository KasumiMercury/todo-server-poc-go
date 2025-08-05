package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing interfaces

type mockTokenValidator struct {
	name             string
	validationResult *TokenValidationResult
}

func (m *mockTokenValidator) ValidateToken(tokenString string) *TokenValidationResult {
	return m.validationResult
}

func (m *mockTokenValidator) Name() string {
	return m.name
}

type mockAuthenticationStrategy struct {
	mockTokenValidator
	isConfigured bool
	priority     int
}

func (m *mockAuthenticationStrategy) IsConfigured() bool {
	return m.isConfigured
}

func (m *mockAuthenticationStrategy) Priority() int {
	return m.priority
}

func TestNewAuthenticationResult(t *testing.T) {
	t.Parallel()

	// Arrange
	mockStrategy := &mockAuthenticationStrategy{
		mockTokenValidator: mockTokenValidator{
			name:             "TestStrategy",
			validationResult: NewTokenValidationResult(true, "user123", nil),
		},
		isConfigured: true,
		priority:     10,
	}

	validationResult := NewTokenValidationResult(true, "user123", nil)

	tests := []struct {
		name             string
		strategy         AuthenticationStrategy
		result           *TokenValidationResult
		expectedValid    bool
		expectedUserID   string
		expectedError    error
		expectedStrategy string
	}{
		{
			name:             "successful authentication",
			strategy:         mockStrategy,
			result:           validationResult,
			expectedValid:    true,
			expectedUserID:   "user123",
			expectedError:    nil,
			expectedStrategy: "TestStrategy",
		},
		{
			name:             "failed authentication",
			strategy:         mockStrategy,
			result:           NewTokenValidationResult(false, "", ErrTokenExpired),
			expectedValid:    false,
			expectedUserID:   "",
			expectedError:    ErrTokenExpired,
			expectedStrategy: "TestStrategy",
		},
		{
			name:             "authentication with validation error",
			strategy:         mockStrategy,
			result:           NewTokenValidationResult(false, "", ErrInvalidTokenSignature),
			expectedValid:    false,
			expectedUserID:   "",
			expectedError:    ErrInvalidTokenSignature,
			expectedStrategy: "TestStrategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			authResult := NewAuthenticationResult(tt.strategy, tt.result)

			// Assert
			require.NotNil(t, authResult)
			assert.Equal(t, tt.expectedValid, authResult.IsValid())
			assert.Equal(t, tt.expectedUserID, authResult.UserID())
			assert.Equal(t, tt.expectedStrategy, authResult.StrategyName())

			if tt.expectedError != nil {
				assert.ErrorIs(t, authResult.Error(), tt.expectedError)
			} else {
				assert.NoError(t, authResult.Error())
			}
		})
	}
}

func TestAuthenticationResult_Methods(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedUserID := "test-user-456"
	expectedError := ErrTokenValidation
	expectedStrategyName := "JWKSStrategy"

	mockStrategy := &mockAuthenticationStrategy{
		mockTokenValidator: mockTokenValidator{
			name: expectedStrategyName,
		},
	}

	validationResult := NewTokenValidationResult(false, expectedUserID, expectedError)
	authResult := NewAuthenticationResult(mockStrategy, validationResult)

	// Act & Assert
	assert.False(t, authResult.IsValid())
	assert.Equal(t, expectedUserID, authResult.UserID())
	assert.ErrorIs(t, authResult.Error(), expectedError)
	assert.Equal(t, expectedStrategyName, authResult.StrategyName())
}

func TestAuthenticationResult_ValidScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		validationIsValid  bool
		validationUserID   string
		validationError    error
		strategyName       string
		strategyConfigured bool
		strategyPriority   int
	}{
		{
			name:               "JWKS strategy success",
			validationIsValid:  true,
			validationUserID:   "jwks-user-123",
			validationError:    nil,
			strategyName:       "JWKSStrategy",
			strategyConfigured: true,
			strategyPriority:   100,
		},
		{
			name:               "private key strategy success",
			validationIsValid:  true,
			validationUserID:   "pk-user-456",
			validationError:    nil,
			strategyName:       "PrivateKeyStrategy",
			strategyConfigured: true,
			strategyPriority:   50,
		},
		{
			name:               "secret strategy failure",
			validationIsValid:  false,
			validationUserID:   "",
			validationError:    ErrInvalidTokenFormat,
			strategyName:       "SecretStrategy",
			strategyConfigured: true,
			strategyPriority:   25,
		},
		{
			name:               "unconfigured strategy",
			validationIsValid:  false,
			validationUserID:   "",
			validationError:    ErrProviderNotConfigured,
			strategyName:       "UnconfiguredStrategy",
			strategyConfigured: false,
			strategyPriority:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockStrategy := &mockAuthenticationStrategy{
				mockTokenValidator: mockTokenValidator{
					name: tt.strategyName,
				},
				isConfigured: tt.strategyConfigured,
				priority:     tt.strategyPriority,
			}

			validationResult := NewTokenValidationResult(tt.validationIsValid, tt.validationUserID, tt.validationError)

			// Act
			authResult := NewAuthenticationResult(mockStrategy, validationResult)

			// Assert
			assert.Equal(t, tt.validationIsValid, authResult.IsValid())
			assert.Equal(t, tt.validationUserID, authResult.UserID())
			assert.Equal(t, tt.strategyName, authResult.StrategyName())
			assert.Equal(t, tt.strategyConfigured, authResult.Strategy.IsConfigured())
			assert.Equal(t, tt.strategyPriority, authResult.Strategy.Priority())

			if tt.validationError != nil {
				assert.ErrorIs(t, authResult.Error(), tt.validationError)
			} else {
				assert.NoError(t, authResult.Error())
			}
		})
	}
}

func TestTokenValidator_Interface(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedName := "MockValidator"
	expectedResult := NewTokenValidationResult(true, "user789", nil)

	validator := &mockTokenValidator{
		name:             expectedName,
		validationResult: expectedResult,
	}

	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token"

	// Act
	result := validator.ValidateToken(testToken)
	name := validator.Name()

	// Assert
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedName, name)
}

func TestAuthenticationStrategy_Interface(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedName := "MockStrategy"
	expectedConfigured := true
	expectedPriority := 75
	expectedResult := NewTokenValidationResult(false, "", ErrTokenExpired)

	strategy := &mockAuthenticationStrategy{
		mockTokenValidator: mockTokenValidator{
			name:             expectedName,
			validationResult: expectedResult,
		},
		isConfigured: expectedConfigured,
		priority:     expectedPriority,
	}

	testToken := "invalid.jwt.token"

	// Act
	result := strategy.ValidateToken(testToken)
	name := strategy.Name()
	configured := strategy.IsConfigured()
	priority := strategy.Priority()

	// Assert
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedConfigured, configured)
	assert.Equal(t, expectedPriority, priority)
}

func TestAuthenticationResult_StrategyProperties(t *testing.T) {
	t.Parallel()

	// Arrange
	highPriorityStrategy := &mockAuthenticationStrategy{
		mockTokenValidator: mockTokenValidator{
			name: "HighPriorityStrategy",
		},
		isConfigured: true,
		priority:     100,
	}

	lowPriorityStrategy := &mockAuthenticationStrategy{
		mockTokenValidator: mockTokenValidator{
			name: "LowPriorityStrategy",
		},
		isConfigured: false,
		priority:     10,
	}

	validResult := NewTokenValidationResult(true, "user123", nil)
	invalidResult := NewTokenValidationResult(false, "", ErrAllProvidersFailed)

	tests := []struct {
		name               string
		strategy           AuthenticationStrategy
		result             *TokenValidationResult
		expectedConfigured bool
		expectedPriority   int
		expectedValid      bool
	}{
		{
			name:               "high priority configured strategy",
			strategy:           highPriorityStrategy,
			result:             validResult,
			expectedConfigured: true,
			expectedPriority:   100,
			expectedValid:      true,
		},
		{
			name:               "low priority unconfigured strategy",
			strategy:           lowPriorityStrategy,
			result:             invalidResult,
			expectedConfigured: false,
			expectedPriority:   10,
			expectedValid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			authResult := NewAuthenticationResult(tt.strategy, tt.result)

			// Assert
			assert.Equal(t, tt.expectedConfigured, authResult.Strategy.IsConfigured())
			assert.Equal(t, tt.expectedPriority, authResult.Strategy.Priority())
			assert.Equal(t, tt.expectedValid, authResult.IsValid())
		})
	}
}

func TestAuthenticationResult_ErrorPropagation(t *testing.T) {
	t.Parallel()

	testErrors := []error{
		ErrTokenValidation,
		ErrInvalidTokenFormat,
		ErrTokenExpired,
		ErrInvalidTokenSignature,
		ErrNoValidProvider,
		ErrProviderNotConfigured,
		ErrAllProvidersFailed,
		errors.New("custom validation error"),
	}

	for _, expectedError := range testErrors {
		t.Run(expectedError.Error(), func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockStrategy := &mockAuthenticationStrategy{
				mockTokenValidator: mockTokenValidator{
					name: "ErrorTestStrategy",
				},
			}

			validationResult := NewTokenValidationResult(false, "", expectedError)

			// Act
			authResult := NewAuthenticationResult(mockStrategy, validationResult)

			// Assert
			assert.False(t, authResult.IsValid())
			assert.ErrorIs(t, authResult.Error(), expectedError)
		})
	}
}

func TestAuthenticationResult_StrategyNaming(t *testing.T) {
	t.Parallel()

	strategyNames := []string{
		"JWKSAuthStrategy",
		"PrivateKeyAuthStrategy",
		"SecretAuthStrategy",
		"TestStrategy",
		"",
		"Strategy-With-Special_Characters.123",
	}

	for _, strategyName := range strategyNames {
		t.Run("strategy_name_"+strategyName, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockStrategy := &mockAuthenticationStrategy{
				mockTokenValidator: mockTokenValidator{
					name: strategyName,
				},
			}

			validationResult := NewTokenValidationResult(true, "user", nil)

			// Act
			authResult := NewAuthenticationResult(mockStrategy, validationResult)

			// Assert
			assert.Equal(t, strategyName, authResult.StrategyName())
		})
	}
}
