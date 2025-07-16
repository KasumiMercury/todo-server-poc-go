package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"time"

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

// HealthResponse represents the response format for health endpoint
type HealthResponse struct {
	Status     string                             `json:"status"`
	Timestamp  time.Time                          `json:"timestamp"`
	Components map[string]HealthComponentResponse `json:"components"`
}

// HealthComponentResponse represents the response format for health component
type HealthComponentResponse struct {
	Status  string                 `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// GetHealth handles GET /health requests
func (h *HealthHandler) GetHealth(c echo.Context) error {
	ctx := c.Request().Context()

	healthStatus := h.healthService.CheckHealth(ctx)

	// Convert service model to response model
	response := HealthResponse{
		Status:     healthStatus.Status,
		Timestamp:  healthStatus.Timestamp,
		Components: make(map[string]HealthComponentResponse),
	}

	for name, component := range healthStatus.Components {
		response.Components[name] = HealthComponentResponse{
			Status:  component.Status,
			Details: component.Details,
		}
	}

	// Return appropriate HTTP status code
	statusCode := http.StatusOK
	if healthStatus.Status == "DOWN" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, response)
}
