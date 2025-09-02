package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
)

func TestNewError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		code            int
		message         string
		details         *string
		expectedCode    int
		expectedMessage string
		expectedDetails *string
	}{
		{
			name:            "error with details",
			code:            400,
			message:         "Bad Request",
			details:         stringPtr("Invalid input provided"),
			expectedCode:    400,
			expectedMessage: "Bad Request",
			expectedDetails: stringPtr("Invalid input provided"),
		},
		{
			name:            "error without details",
			code:            500,
			message:         "Internal Server Error",
			details:         nil,
			expectedCode:    500,
			expectedMessage: "Internal Server Error",
			expectedDetails: nil,
		},
		{
			name:            "empty message",
			code:            404,
			message:         "",
			details:         nil,
			expectedCode:    404,
			expectedMessage: "",
			expectedDetails: nil,
		},
		{
			name:            "custom error code",
			code:            422,
			message:         "Unprocessable Entity",
			details:         stringPtr("Validation failed"),
			expectedCode:    422,
			expectedMessage: "Unprocessable Entity",
			expectedDetails: stringPtr("Validation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := NewError(tt.code, tt.message, tt.details)

			// Assert
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMessage, result.Message)

			if tt.expectedDetails != nil {
				assert.NotNil(t, result.Details)
				assert.Equal(t, *tt.expectedDetails, *result.Details)
			} else {
				assert.Nil(t, result.Details)
			}
		})
	}
}

func TestNewBadRequestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		details         *string
		expectedCode    int
		expectedMessage string
		expectedDetails *string
	}{
		{
			name:            "bad request with details",
			message:         "Invalid request format",
			details:         stringPtr("JSON parsing failed"),
			expectedCode:    400,
			expectedMessage: "Invalid request format",
			expectedDetails: stringPtr("JSON parsing failed"),
		},
		{
			name:            "bad request without details",
			message:         "Missing required fields",
			details:         nil,
			expectedCode:    400,
			expectedMessage: "Missing required fields",
			expectedDetails: nil,
		},
		{
			name:            "validation error",
			message:         "Validation failed",
			details:         stringPtr("Field 'title' is required"),
			expectedCode:    400,
			expectedMessage: "Validation failed",
			expectedDetails: stringPtr("Field 'title' is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := NewBadRequestError(tt.message, tt.details)

			// Assert
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMessage, result.Message)

			if tt.expectedDetails != nil {
				assert.NotNil(t, result.Details)
				assert.Equal(t, *tt.expectedDetails, *result.Details)
			} else {
				assert.Nil(t, result.Details)
			}
		})
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		details         *string
		expectedCode    int
		expectedMessage string
		expectedDetails *string
	}{
		{
			name:            "unauthorized with token error",
			message:         "Unauthorized",
			details:         stringPtr("Invalid token"),
			expectedCode:    401,
			expectedMessage: "Unauthorized",
			expectedDetails: stringPtr("Invalid token"),
		},
		{
			name:            "unauthorized without details",
			message:         "Authentication required",
			details:         nil,
			expectedCode:    401,
			expectedMessage: "Authentication required",
			expectedDetails: nil,
		},
		{
			name:            "expired token error",
			message:         "Token expired",
			details:         stringPtr("Please re-authenticate"),
			expectedCode:    401,
			expectedMessage: "Token expired",
			expectedDetails: stringPtr("Please re-authenticate"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := NewUnauthorizedError(tt.message, tt.details)

			// Assert
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMessage, result.Message)

			if tt.expectedDetails != nil {
				assert.NotNil(t, result.Details)
				assert.Equal(t, *tt.expectedDetails, *result.Details)
			} else {
				assert.Nil(t, result.Details)
			}
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "resource not found",
			message:         "Task not found",
			expectedCode:    404,
			expectedMessage: "Task not found",
		},
		{
			name:            "endpoint not found",
			message:         "Endpoint not found",
			expectedCode:    404,
			expectedMessage: "Endpoint not found",
		},
		{
			name:            "user not found",
			message:         "User not found",
			expectedCode:    404,
			expectedMessage: "User not found",
		},
		{
			name:            "empty message",
			message:         "",
			expectedCode:    404,
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := NewNotFoundError(tt.message)

			// Assert
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMessage, result.Message)
			assert.Nil(t, result.Details)
		})
	}
}

func TestNewInternalServerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		details         *string
		expectedCode    int
		expectedMessage string
		expectedDetails *string
	}{
		{
			name:            "database error",
			message:         "Internal Server Error",
			details:         stringPtr("Database connection failed"),
			expectedCode:    500,
			expectedMessage: "Internal Server Error",
			expectedDetails: stringPtr("Database connection failed"),
		},
		{
			name:            "generic internal error",
			message:         "Something went wrong",
			details:         nil,
			expectedCode:    500,
			expectedMessage: "Something went wrong",
			expectedDetails: nil,
		},
		{
			name:            "service unavailable",
			message:         "Service temporarily unavailable",
			details:         stringPtr("External service timeout"),
			expectedCode:    500,
			expectedMessage: "Service temporarily unavailable",
			expectedDetails: stringPtr("External service timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := NewInternalServerError(tt.message, tt.details)

			// Assert
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMessage, result.Message)

			if tt.expectedDetails != nil {
				assert.NotNil(t, result.Details)
				assert.Equal(t, *tt.expectedDetails, *result.Details)
			} else {
				assert.Nil(t, result.Details)
			}
		})
	}
}

func TestErrorStruct_Structure(t *testing.T) {
	t.Parallel()

	t.Run("error struct fields", func(t *testing.T) {
		t.Parallel()

		// Arrange
		details := "test details"
		err := generated.ErrorResponse{
			Code:    400,
			Message: "test message",
			Details: &details,
		}

		// Assert
		assert.Equal(t, 400, err.Code)
		assert.Equal(t, "test message", err.Message)
		assert.NotNil(t, err.Details)
		assert.Equal(t, "test details", *err.Details)
	})

	t.Run("error struct with nil details", func(t *testing.T) {
		t.Parallel()

		// Arrange
		err := generated.ErrorResponse{
			Code:    404,
			Message: "not found",
			Details: nil,
		}

		// Assert
		assert.Equal(t, 404, err.Code)
		assert.Equal(t, "not found", err.Message)
		assert.Nil(t, err.Details)
	})
}

func TestErrorFunctions_HTTPStatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		errorFunc      func() generated.ErrorResponse
		expectedCode   int
		expectedStatus string
	}{
		{
			name: "bad request",
			errorFunc: func() generated.ErrorResponse {
				return NewBadRequestError("test", nil)
			},
			expectedCode:   400,
			expectedStatus: "Bad Request",
		},
		{
			name: "unauthorized",
			errorFunc: func() generated.ErrorResponse {
				return NewUnauthorizedError("test", nil)
			},
			expectedCode:   401,
			expectedStatus: "Unauthorized",
		},
		{
			name: "not found",
			errorFunc: func() generated.ErrorResponse {
				return NewNotFoundError("test")
			},
			expectedCode:   404,
			expectedStatus: "Not Found",
		},
		{
			name: "internal server error",
			errorFunc: func() generated.ErrorResponse {
				return NewInternalServerError("test", nil)
			},
			expectedCode:   500,
			expectedStatus: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := tt.errorFunc()

			// Assert
			assert.Equal(t, tt.expectedCode, result.Code)

			assert.True(t, isValidHTTPStatusCode(result.Code), "HTTP status code %d should be valid", result.Code)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func isValidHTTPStatusCode(code int) bool {
	return code >= 100 && code < 600
}
