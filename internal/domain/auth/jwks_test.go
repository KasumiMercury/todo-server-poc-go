package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWKsEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		url           string
		expectedError error
	}{
		{
			name:          "valid URL",
			url:           "https://example.com/.well-known/jwks.json",
			expectedError: nil,
		},
		{
			name:          "valid localhost URL",
			url:           "http://localhost:8080/jwks",
			expectedError: nil,
		},
		{
			name:          "empty URL",
			url:           "",
			expectedError: ErrInvalidJWKsEndpoint,
		},
		{
			name:          "URL with special characters",
			url:           "https://auth.example.com/oauth2/v1/keys?tenant=test",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			endpoint, err := NewJWKsEndpoint(tt.url)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, endpoint)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, endpoint)
				assert.Equal(t, tt.url, endpoint.URL())
			}
		})
	}
}

func TestJWKsEndpoint_URL(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedURL := "https://auth.example.com/.well-known/jwks.json"
	endpoint, err := NewJWKsEndpoint(expectedURL)
	require.NoError(t, err)
	require.NotNil(t, endpoint)

	// Act
	actualURL := endpoint.URL()

	// Assert
	assert.Equal(t, expectedURL, actualURL)
}

func TestNewJWKsCacheConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		cacheDuration   time.Duration
		refreshPadding  time.Duration
		expectedCache   time.Duration
		expectedRefresh time.Duration
	}{
		{
			name:            "typical configuration",
			cacheDuration:   5 * time.Minute,
			refreshPadding:  30 * time.Second,
			expectedCache:   5 * time.Minute,
			expectedRefresh: 30 * time.Second,
		},
		{
			name:            "zero durations",
			cacheDuration:   0,
			refreshPadding:  0,
			expectedCache:   0,
			expectedRefresh: 0,
		},
		{
			name:            "large durations",
			cacheDuration:   24 * time.Hour,
			refreshPadding:  1 * time.Hour,
			expectedCache:   24 * time.Hour,
			expectedRefresh: 1 * time.Hour,
		},
		{
			name:            "refresh padding larger than cache duration",
			cacheDuration:   1 * time.Minute,
			refreshPadding:  5 * time.Minute,
			expectedCache:   1 * time.Minute,
			expectedRefresh: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			config := NewJWKsCacheConfig(tt.cacheDuration, tt.refreshPadding)

			// Assert
			require.NotNil(t, config)
			assert.Equal(t, tt.expectedCache, config.CacheDuration())
			assert.Equal(t, tt.expectedRefresh, config.RefreshPadding())
		})
	}
}

func TestJWKsCacheConfig_Methods(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedCacheDuration := 10 * time.Minute
	expectedRefreshPadding := 2 * time.Minute
	config := NewJWKsCacheConfig(expectedCacheDuration, expectedRefreshPadding)

	// Act & Assert
	assert.Equal(t, expectedCacheDuration, config.CacheDuration())
	assert.Equal(t, expectedRefreshPadding, config.RefreshPadding())
}

func TestNewTokenValidationResult(t *testing.T) {
	t.Parallel()

	testError := errors.New("test validation error")

	tests := []struct {
		name           string
		isValid        bool
		userID         string
		err            error
		expectedValid  bool
		expectedUserID string
		expectedError  error
	}{
		{
			name:           "valid token result",
			isValid:        true,
			userID:         "user123",
			err:            nil,
			expectedValid:  true,
			expectedUserID: "user123",
			expectedError:  nil,
		},
		{
			name:           "invalid token result with error",
			isValid:        false,
			userID:         "",
			err:            testError,
			expectedValid:  false,
			expectedUserID: "",
			expectedError:  testError,
		},
		{
			name:           "invalid token result without error",
			isValid:        false,
			userID:         "",
			err:            nil,
			expectedValid:  false,
			expectedUserID: "",
			expectedError:  nil,
		},
		{
			name:           "valid token with UUID user ID",
			isValid:        true,
			userID:         "550e8400-e29b-41d4-a716-446655440000",
			err:            nil,
			expectedValid:  true,
			expectedUserID: "550e8400-e29b-41d4-a716-446655440000",
			expectedError:  nil,
		},
		{
			name:           "token validation error with user ID",
			isValid:        false,
			userID:         "user456",
			err:            ErrTokenExpired,
			expectedValid:  false,
			expectedUserID: "user456",
			expectedError:  ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := NewTokenValidationResult(tt.isValid, tt.userID, tt.err)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedValid, result.IsValid())
			assert.Equal(t, tt.expectedUserID, result.UserID())
			if tt.expectedError != nil {
				assert.ErrorIs(t, result.Error(), tt.expectedError)
			} else {
				assert.NoError(t, result.Error())
			}
		})
	}
}

func TestTokenValidationResult_Methods(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedValid := true
	expectedUserID := "test-user-123"
	expectedError := ErrInvalidTokenSignature

	result := NewTokenValidationResult(expectedValid, expectedUserID, expectedError)

	// Act & Assert
	assert.Equal(t, expectedValid, result.IsValid())
	assert.Equal(t, expectedUserID, result.UserID())
	assert.ErrorIs(t, result.Error(), expectedError)
}

func TestTokenValidationResult_NoError(t *testing.T) {
	t.Parallel()

	// Arrange
	result := NewTokenValidationResult(true, "user123", nil)

	// Act & Assert
	assert.True(t, result.IsValid())
	assert.Equal(t, "user123", result.UserID())
	assert.NoError(t, result.Error())
}
