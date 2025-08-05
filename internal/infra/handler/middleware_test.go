package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	infraAuth "github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth/mocks"
)

//go:generate go run go.uber.org/mock/mockgen -source=../../domain/auth/token_validator.go -destination=mocks/mock_auth_strategy.go -package=mocks

func TestNewAuthenticationMiddleware(t *testing.T) {
	t.Parallel()

	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	strategy := mocks.NewMockAuthenticationStrategy(ctrl)
	strategy.EXPECT().Name().Return("TestStrategy").AnyTimes()
	strategy.EXPECT().IsConfigured().Return(true).AnyTimes()
	strategy.EXPECT().Priority().Return(50).AnyTimes()

	authService, err := infraAuth.NewAuthenticationServiceWithProviders([]auth.AuthenticationStrategy{strategy})
	require.NoError(t, err)

	// Act
	middleware := NewAuthenticationMiddleware(authService)

	// Assert
	require.NotNil(t, middleware)
	assert.Equal(t, authService, middleware.authService)
}

func TestAuthenticationMiddleware_MiddlewareFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		authHeader           string
		validationResult     *auth.TokenValidationResult
		expectedStatusCode   int
		expectedUserID       string
		expectedAuthStrategy string
		expectNextCalled     bool
	}{
		{
			name:                 "successful authentication",
			authHeader:           "Bearer valid-token",
			validationResult:     auth.NewTokenValidationResult(true, "user123", nil),
			expectedStatusCode:   http.StatusOK,
			expectedUserID:       "user123",
			expectedAuthStrategy: "TestStrategy",
			expectNextCalled:     true,
		},
		{
			name:               "missing authorization header",
			authHeader:         "",
			validationResult:   nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectNextCalled:   false,
		},
		{
			name:               "invalid authorization format",
			authHeader:         "InvalidFormat token",
			validationResult:   nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectNextCalled:   false,
		},
		{
			name:               "invalid token",
			authHeader:         "Bearer invalid-token",
			validationResult:   auth.NewTokenValidationResult(false, "", auth.ErrTokenValidation),
			expectedStatusCode: http.StatusUnauthorized,
			expectNextCalled:   false,
		},
		{
			name:               "token validation error",
			authHeader:         "Bearer expired-token",
			validationResult:   auth.NewTokenValidationResult(false, "", auth.ErrTokenExpired),
			expectedStatusCode: http.StatusUnauthorized,
			expectNextCalled:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			nextCalled := false
			next := func(c echo.Context) error {
				nextCalled = true

				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			}

			strategy := mocks.NewMockAuthenticationStrategy(ctrl)
			strategy.EXPECT().Name().Return("TestStrategy").AnyTimes()
			strategy.EXPECT().IsConfigured().Return(true).AnyTimes()
			strategy.EXPECT().Priority().Return(50).AnyTimes()

			if tt.validationResult != nil {
				strategy.EXPECT().ValidateToken(gomock.Any()).Return(tt.validationResult).AnyTimes()
			}

			authService, err := infraAuth.NewAuthenticationServiceWithProviders([]auth.AuthenticationStrategy{strategy})
			require.NoError(t, err)

			middleware := NewAuthenticationMiddleware(authService)
			middlewareFunc := middleware.MiddlewareFunc()
			handler := middlewareFunc(next)

			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Act
			err = handler(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)
			assert.Equal(t, tt.expectNextCalled, nextCalled)

			if tt.expectNextCalled {
				userID := c.Get("user_id")
				authStrategy := c.Get("auth_strategy")

				assert.Equal(t, tt.expectedUserID, userID)
				assert.Equal(t, tt.expectedAuthStrategy, authStrategy)
			}
		})
	}
}

func TestAuthenticationMiddleware_ErrorResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		authHeader         string
		validationResult   *auth.TokenValidationResult
		expectedMessage    string
		expectDetailsField bool
	}{
		{
			name:               "missing authorization header error",
			authHeader:         "",
			validationResult:   nil,
			expectedMessage:    "Unauthorized",
			expectDetailsField: true,
		},
		{
			name:               "invalid authorization format error",
			authHeader:         "InvalidFormat token",
			validationResult:   nil,
			expectedMessage:    "Unauthorized",
			expectDetailsField: true,
		},
		{
			name:               "token validation error",
			authHeader:         "Bearer invalid-token",
			validationResult:   auth.NewTokenValidationResult(false, "", auth.ErrTokenValidation),
			expectedMessage:    "Unauthorized",
			expectDetailsField: true,
		},
		{
			name:               "token expired error",
			authHeader:         "Bearer expired-token",
			validationResult:   auth.NewTokenValidationResult(false, "", auth.ErrTokenExpired),
			expectedMessage:    "Unauthorized",
			expectDetailsField: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			strategy := mocks.NewMockAuthenticationStrategy(ctrl)
			strategy.EXPECT().Name().Return("TestStrategy").AnyTimes()
			strategy.EXPECT().IsConfigured().Return(true).AnyTimes()
			strategy.EXPECT().Priority().Return(50).AnyTimes()

			if tt.validationResult != nil {
				strategy.EXPECT().ValidateToken(gomock.Any()).Return(tt.validationResult).AnyTimes()
			}

			authService, err := infraAuth.NewAuthenticationServiceWithProviders([]auth.AuthenticationStrategy{strategy})
			require.NoError(t, err)

			middleware := NewAuthenticationMiddleware(authService)
			middlewareFunc := middleware.MiddlewareFunc()
			handler := middlewareFunc(func(c echo.Context) error {
				return c.JSON(http.StatusOK, "should not reach here")
			})

			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Act
			err = handler(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			assert.Contains(t, rec.Body.String(), tt.expectedMessage)

			if tt.expectDetailsField {
				assert.Contains(t, rec.Body.String(), "details")
			}
		})
	}
}

func TestStrPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "non-empty string",
			input:    "test string",
			expected: "test string",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with special characters",
			input:    "hello@#$%^&*()",
			expected: "hello@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := strPtr(tt.input)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}
