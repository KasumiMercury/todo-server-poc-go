package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
)

func TestCORSMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		config                   config.Config
		requestMethod            string
		requestOrigin            string
		requestHeaders           string
		expectedAllowOrigin      string
		expectedAllowMethods     string
		expectedAllowHeaders     string
		expectedAllowCredentials string
		expectedMaxAge           string
		expectedStatusCode       int
	}{
		{
			name: "preflight request with allowed origin",
			config: config.Config{
				AllowOrigins: []string{"http://localhost:3000", "https://example.com"},
			},
			requestMethod:            http.MethodOptions,
			requestOrigin:            "http://localhost:3000",
			requestHeaders:           "Authorization, Content-Type",
			expectedAllowOrigin:      "http://localhost:3000",
			expectedAllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			expectedAllowHeaders:     "Authorization,Content-Type",
			expectedAllowCredentials: "true",
			expectedMaxAge:           "43200",
			expectedStatusCode:       http.StatusNoContent,
		},
		{
			name: "preflight request with disallowed origin",
			config: config.Config{
				AllowOrigins: []string{"https://allowed-domain.com"},
			},
			requestMethod:       http.MethodOptions,
			requestOrigin:       "http://malicious-site.com",
			requestHeaders:      "Authorization, Content-Type",
			expectedAllowOrigin: "",
			expectedStatusCode:  http.StatusNoContent,
		},
		{
			name: "simple request with allowed origin",
			config: config.Config{
				AllowOrigins: []string{"https://example.com"},
			},
			requestMethod:            http.MethodGet,
			requestOrigin:            "https://example.com",
			expectedAllowOrigin:      "https://example.com",
			expectedAllowCredentials: "true",
			expectedStatusCode:       http.StatusOK,
		},
		{
			name: "wildcard origin configuration",
			config: config.Config{
				AllowOrigins: []string{"*"},
			},
			requestMethod:            http.MethodGet,
			requestOrigin:            "https://any-domain.com",
			expectedAllowOrigin:      "*",
			expectedAllowCredentials: "",
			expectedStatusCode:       http.StatusOK,
		},
		{
			name: "multiple allowed origins",
			config: config.Config{
				AllowOrigins: []string{
					"http://localhost:3000",
					"http://localhost:8080",
					"https://staging.example.com",
					"https://production.example.com",
				},
			},
			requestMethod:            http.MethodPost,
			requestOrigin:            "https://staging.example.com",
			expectedAllowOrigin:      "https://staging.example.com",
			expectedAllowCredentials: "true",
			expectedStatusCode:       http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			e := echo.New()
			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			if tt.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
			}
			if tt.requestMethod == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", "POST")
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			next := func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			}

			corsMiddleware := CORSMiddleware(tt.config)
			handler := corsMiddleware(next)

			// Act
			err := handler(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			// Check CORS headers
			if tt.expectedAllowOrigin != "" {
				assert.Equal(t, tt.expectedAllowOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.expectedAllowMethods != "" {
				assert.Equal(t, tt.expectedAllowMethods, rec.Header().Get("Access-Control-Allow-Methods"))
			}

			if tt.expectedAllowHeaders != "" {
				assert.Equal(t, tt.expectedAllowHeaders, rec.Header().Get("Access-Control-Allow-Headers"))
			}

			if tt.expectedAllowCredentials != "" {
				assert.Equal(t, tt.expectedAllowCredentials, rec.Header().Get("Access-Control-Allow-Credentials"))
			}

			if tt.expectedMaxAge != "" {
				assert.Equal(t, tt.expectedMaxAge, rec.Header().Get("Access-Control-Max-Age"))
			}
		})
	}
}

func TestCORSMiddleware_Configuration(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		AllowOrigins: []string{"https://test.example.com"},
	}

	// Act
	corsMiddleware := CORSMiddleware(cfg)

	// Assert
	require.NotNil(t, corsMiddleware)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://test.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	handler := corsMiddleware(next)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://test.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_MaxAge(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		AllowOrigins: []string{"https://example.com"},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	corsMiddleware := CORSMiddleware(cfg)
	handler := corsMiddleware(next)

	// Act
	err := handler(c)

	// Assert
	require.NoError(t, err)

	maxAgeHeader := rec.Header().Get("Access-Control-Max-Age")
	expectedMaxAge := int((12 * time.Hour).Seconds()) // 43200 seconds
	assert.Equal(t, "43200", maxAgeHeader)

	assert.Equal(t, 43200, expectedMaxAge)
}

func TestCORSMiddleware_AllowedMethods(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := config.Config{
		AllowOrigins: []string{"https://example.com"},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	corsMiddleware := CORSMiddleware(cfg)
	handler := corsMiddleware(next)

	// Act
	err := handler(c)

	// Assert
	require.NoError(t, err)

	allowedMethods := rec.Header().Get("Access-Control-Allow-Methods")

	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	for _, method := range expectedMethods {
		assert.Contains(t, allowedMethods, method, "Expected method %s to be allowed", method)
	}
}

func TestCORSMiddleware_AllowedHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		requestHeaders  string
		expectedHeaders []string
	}{
		{
			name:            "authorization and content-type headers",
			requestHeaders:  "Authorization, Content-Type",
			expectedHeaders: []string{"Authorization", "Content-Type"},
		},
		{
			name:            "only authorization header",
			requestHeaders:  "Authorization",
			expectedHeaders: []string{"Authorization"},
		},
		{
			name:            "only content-type header",
			requestHeaders:  "Content-Type",
			expectedHeaders: []string{"Content-Type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cfg := config.Config{
				AllowOrigins: []string{"https://example.com"},
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodOptions, "/test", nil)
			req.Header.Set("Origin", "https://example.com")
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			next := func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}

			corsMiddleware := CORSMiddleware(cfg)
			handler := corsMiddleware(next)

			// Act
			err := handler(c)

			// Assert
			require.NoError(t, err)

			allowedHeaders := rec.Header().Get("Access-Control-Allow-Headers")
			for _, header := range tt.expectedHeaders {
				assert.Contains(t, allowedHeaders, header, "Expected header %s to be allowed", header)
			}
		})
	}
}
