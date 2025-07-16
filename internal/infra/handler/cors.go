package handler

import (
	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"time"
)

// CORSMiddleware returns a Gin middleware that handles CORS headers
func CORSMiddleware(cfg config.Config) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	})
}
