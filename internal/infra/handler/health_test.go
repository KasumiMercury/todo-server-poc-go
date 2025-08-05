package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/mocks"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

//go:generate go run go.uber.org/mock/mockgen -source=../service/health.go -destination=mocks/mock_health_service.go -package=mocks

func TestNewHealthHandler(t *testing.T) {
	t.Parallel()

	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHealthService := mocks.NewMockHealthService(ctrl)

	// Act
	handler := NewHealthHandler(mockHealthService)

	// Assert
	require.NotNil(t, handler)
	assert.Equal(t, mockHealthService, handler.healthService)
}

func TestHealthHandler_GetHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		healthStatus       service.HealthStatus
		expectedStatusCode int
		expectedStatus     string
		expectedComponents map[string]HealthComponentResponse
	}{
		{
			name: "healthy system",
			healthStatus: service.HealthStatus{
				Status:    "UP",
				Timestamp: time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
				Components: map[string]service.HealthComponent{
					"database": {
						Status: "UP",
						Details: map[string]interface{}{
							"connection":   "established",
							"responseTime": "5ms",
						},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedStatus:     "UP",
			expectedComponents: map[string]HealthComponentResponse{
				"database": {
					Status: "UP",
					Details: map[string]interface{}{
						"connection":   "established",
						"responseTime": "5ms",
					},
				},
			},
		},
		{
			name: "unhealthy system",
			healthStatus: service.HealthStatus{
				Status:    "DOWN",
				Timestamp: time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
				Components: map[string]service.HealthComponent{
					"database": {
						Status: "DOWN",
						Details: map[string]interface{}{
							"error": "connection timeout",
						},
					},
				},
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedStatus:     "DOWN",
			expectedComponents: map[string]HealthComponentResponse{
				"database": {
					Status: "DOWN",
					Details: map[string]interface{}{
						"error": "connection timeout",
					},
				},
			},
		},
		{
			name: "mixed component status",
			healthStatus: service.HealthStatus{
				Status:    "DOWN",
				Timestamp: time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
				Components: map[string]service.HealthComponent{
					"database": {
						Status: "DOWN",
						Details: map[string]interface{}{
							"error": "connection failed",
						},
					},
					"cache": {
						Status: "UP",
						Details: map[string]interface{}{
							"connection": "ok",
						},
					},
				},
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedStatus:     "DOWN",
			expectedComponents: map[string]HealthComponentResponse{
				"database": {
					Status: "DOWN",
					Details: map[string]interface{}{
						"error": "connection failed",
					},
				},
				"cache": {
					Status: "UP",
					Details: map[string]interface{}{
						"connection": "ok",
					},
				},
			},
		},
		{
			name: "no components",
			healthStatus: service.HealthStatus{
				Status:     "UP",
				Timestamp:  time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
				Components: map[string]service.HealthComponent{},
			},
			expectedStatusCode: http.StatusOK,
			expectedStatus:     "UP",
			expectedComponents: map[string]HealthComponentResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHealthService := mocks.NewMockHealthService(ctrl)
			mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(tt.healthStatus)

			handler := NewHealthHandler(mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Act
			err := handler.GetHealth(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			var response HealthResponse

			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, response.Status)
			assert.Equal(t, tt.healthStatus.Timestamp, response.Timestamp)
			assert.Equal(t, tt.expectedComponents, response.Components)
		})
	}
}

func TestHealthHandler_GetHealth_ContextHandling(t *testing.T) {
	t.Parallel()

	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHealthService := mocks.NewMockHealthService(ctrl)
	mockHealthService.EXPECT().CheckHealth(gomock.Any()).DoAndReturn(func(ctx context.Context) service.HealthStatus {
		assert.NotNil(t, ctx)

		return service.HealthStatus{
			Status:     "UP",
			Timestamp:  time.Now(),
			Components: map[string]service.HealthComponent{},
		}
	})

	handler := NewHealthHandler(mockHealthService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetHealth(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthHandler_ResponseStructure(t *testing.T) {
	t.Parallel()

	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedTimestamp := time.Date(2023, 12, 1, 15, 30, 45, 0, time.UTC)
	mockHealthService := mocks.NewMockHealthService(ctrl)
	mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(service.HealthStatus{
		Status:    "UP",
		Timestamp: expectedTimestamp,
		Components: map[string]service.HealthComponent{
			"database": {
				Status: "UP",
				Details: map[string]interface{}{
					"version":    "14.5",
					"connection": "active",
					"pool_size":  10,
				},
			},
		},
	})

	handler := NewHealthHandler(mockHealthService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetHealth(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response HealthResponse

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "UP", response.Status)
	assert.Equal(t, expectedTimestamp, response.Timestamp)
	assert.NotNil(t, response.Components)

	dbComponent, exists := response.Components["database"]
	assert.True(t, exists)
	assert.Equal(t, "UP", dbComponent.Status)
	assert.NotNil(t, dbComponent.Details)

	assert.Equal(t, "14.5", dbComponent.Details["version"])
	assert.Equal(t, "active", dbComponent.Details["connection"])
	assert.Equal(t, float64(10), dbComponent.Details["pool_size"]) // JSON unmarshaling converts numbers to float64
}

func TestHealthHandler_StatusCodeMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		systemStatus   string
		expectedStatus int
	}{
		{
			name:           "UP status returns 200",
			systemStatus:   "UP",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DOWN status returns 503",
			systemStatus:   "DOWN",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "DEGRADED status returns 200",
			systemStatus:   "DEGRADED",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "UNKNOWN status returns 200",
			systemStatus:   "UNKNOWN",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHealthService := mocks.NewMockHealthService(ctrl)
			mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(service.HealthStatus{
				Status:     tt.systemStatus,
				Timestamp:  time.Now(),
				Components: map[string]service.HealthComponent{},
			})

			handler := NewHealthHandler(mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Act
			err := handler.GetHealth(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestHealthHandler_ComponentDetailsMapping(t *testing.T) {
	t.Parallel()

	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHealthService := mocks.NewMockHealthService(ctrl)
	mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(service.HealthStatus{
		Status:    "UP",
		Timestamp: time.Now(),
		Components: map[string]service.HealthComponent{
			"database": {
				Status: "UP",
				Details: map[string]interface{}{
					"connection_count": 5,
					"max_connections":  100,
					"response_time":    "2ms",
					"last_check":       "2023-12-01T12:00:00Z",
				},
			},
			"cache": {
				Status: "UP",
				Details: map[string]interface{}{
					"memory_usage": "65%",
					"hit_ratio":    0.95,
				},
			},
		},
	})

	handler := NewHealthHandler(mockHealthService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetHealth(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response HealthResponse

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	dbComponent := response.Components["database"]
	assert.Equal(t, "UP", dbComponent.Status)
	assert.Equal(t, float64(5), dbComponent.Details["connection_count"])
	assert.Equal(t, float64(100), dbComponent.Details["max_connections"])
	assert.Equal(t, "2ms", dbComponent.Details["response_time"])
	assert.Equal(t, "2023-12-01T12:00:00Z", dbComponent.Details["last_check"])

	cacheComponent := response.Components["cache"]
	assert.Equal(t, "UP", cacheComponent.Status)
	assert.Equal(t, "65%", cacheComponent.Details["memory_usage"])
	assert.Equal(t, 0.95, cacheComponent.Details["hit_ratio"])
}
