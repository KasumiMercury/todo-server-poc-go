package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	taskHandler "github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	healthService service.HealthService
}

// NewHealthHandler creates a new health handler instance
func NewHealthHandler(healthService service.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// GetHealth handles GET /health requests and returns OpenAPI-compliant response
func (h *HealthHandler) GetHealth(c echo.Context) error {
	ctx := c.Request().Context()

	healthStatus := h.healthService.CheckHealth(ctx)

	// Convert service model to generated response model
	components := taskHandler.HealthStatus{
		Status:    taskHandler.HealthStatusStatus(healthStatus.Status),
		Timestamp: healthStatus.Timestamp,
		Components: struct { //nolint:exhaustruct
			Database *taskHandler.HealthComponent `json:"database,omitempty"`
		}{},
	}

	if dbComponent, exists := healthStatus.Components["database"]; exists {
		components.Components.Database = &taskHandler.HealthComponent{
			Status:  taskHandler.HealthComponentStatus(dbComponent.Status),
			Details: &dbComponent.Details,
		}
	}

	// Return appropriate HTTP status code
	statusCode := http.StatusOK
	if healthStatus.Status == "DOWN" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, components)
}
