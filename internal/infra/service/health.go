package service

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// HealthService defines the interface for health check operations
type HealthService interface {
	CheckHealth(ctx context.Context) HealthStatus
}

// HealthStatus represents the overall health status of the application
type HealthStatus struct {
	Status     string                       `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Components map[string]HealthComponent  `json:"components"`
}

// HealthComponent represents the health status of an individual component
type HealthComponent struct {
	Status  string                 `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthServiceImpl implements the HealthService interface
type HealthServiceImpl struct {
	db *gorm.DB
}

// NewHealthService creates a new health service instance
func NewHealthService(db *gorm.DB) HealthService {
	return &HealthServiceImpl{
		db: db,
	}
}

// CheckHealth performs health checks on all components
func (h *HealthServiceImpl) CheckHealth(ctx context.Context) HealthStatus {
	timestamp := time.Now()
	components := make(map[string]HealthComponent)
	
	// Check database health
	dbHealth := h.checkDatabaseHealth(ctx)
	components["database"] = dbHealth
	
	// Determine overall status
	overallStatus := "UP"
	if dbHealth.Status == "DOWN" {
		overallStatus = "DOWN"
	}
	
	return HealthStatus{
		Status:     overallStatus,
		Timestamp:  timestamp,
		Components: components,
	}
}

// checkDatabaseHealth checks the health of the database connection
func (h *HealthServiceImpl) checkDatabaseHealth(ctx context.Context) HealthComponent {
	start := time.Now()
	
	// Get the underlying sql.DB to check connection
	sqlDB, err := h.db.DB()
	if err != nil {
		return HealthComponent{
			Status: "DOWN",
			Details: map[string]interface{}{
				"error": "Failed to get database connection",
			},
		}
	}
	
	// Ping the database with context
	if err := sqlDB.PingContext(ctx); err != nil {
		return HealthComponent{
			Status: "DOWN",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}
	
	responseTime := time.Since(start)
	
	return HealthComponent{
		Status: "UP",
		Details: map[string]interface{}{
			"connection":   "PostgreSQL",
			"responseTime": responseTime.String(),
		},
	}
}