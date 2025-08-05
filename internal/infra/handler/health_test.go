package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	generated "github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/mocks"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

//go:generate go run go.uber.org/mock/mockgen -source=../service/health.go -destination=mocks/mock_health_service.go -package=mocks

func TestNewHealthHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHealthService := mocks.NewMockHealthService(ctrl)
	handler := NewHealthHandler(mockHealthService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockHealthService, handler.healthService)
}

func TestHealthHandler_GetHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		healthStatus       service.HealthStatus
		expectedStatusCode int
		expectedStatus     string
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHealthService := mocks.NewMockHealthService(ctrl)
			handler := NewHealthHandler(mockHealthService)

			mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(tt.healthStatus)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.GetHealth(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			var response generated.HealthStatus

			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, string(response.Status))
			assert.Equal(t, tt.healthStatus.Timestamp, response.Timestamp)

			if dbComponent, exists := tt.healthStatus.Components["database"]; exists {
				assert.NotNil(t, response.Components.Database)
				assert.Equal(t, dbComponent.Status, string(response.Components.Database.Status))

				if dbComponent.Details != nil {
					assert.Equal(t, dbComponent.Details, *response.Components.Database.Details)
				}
			}
		})
	}
}

func TestHealthHandler_GetHealth_NoComponents(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHealthService := mocks.NewMockHealthService(ctrl)
	handler := NewHealthHandler(mockHealthService)

	healthStatus := service.HealthStatus{
		Status:     "UP",
		Timestamp:  time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
		Components: map[string]service.HealthComponent{},
	}

	mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(healthStatus)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetHealth(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response generated.HealthStatus

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "UP", string(response.Status))
	assert.Nil(t, response.Components.Database)
}
