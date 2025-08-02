package service

import (
	"log"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

type MetricsService struct {
	cfg        config.Config
	metricPort string
}

func NewMetricsService(cfg config.Config) *MetricsService {
	return &MetricsService{
		cfg:        cfg,
		metricPort: ":8081",
	}
}

func (m *MetricsService) SetupMiddleware(router *echo.Echo) {
	router.Use(echoprometheus.NewMiddleware(m.cfg.ServiceName))
}

func (m *MetricsService) StartMetricsServer() {
	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())

		if err := metrics.Start(m.metricPort); err != nil {
			log.Fatal("Failed to start metrics server:", err)
		}
	}()
}

func (m *MetricsService) SetMetricsPort(port string) {
	m.metricPort = port
}
