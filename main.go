package main

import (
	"fmt"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"log"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/repository"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := initDB(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize Echo router
	router := echo.New()

	// Add panic recovery middleware
	router.Use(middleware.Recover())

	// Initialize metrics service
	metricsService := service.NewMetricsService(*cfg)
	metricsService.SetupMiddleware(router)
	metricsService.StartMetricsServer()

	// Setup OpenTelemetry middleware
	router.Use(otelecho.Middleware(cfg.ServiceName))

	// Add CORS middleware globally
	router.Use(handler.CORSMiddleware(*cfg))

	taskRepo := repository.NewTaskDB(db)
	taskController := controller.NewTask(taskRepo)

	// Initialize health service
	healthService := service.NewHealthService(db)

	taskServer := handler.NewTaskServer(
		*taskController,
		healthService,
	)

	// Setup authentication service with strategy pattern
	authService, err := auth.NewAuthenticationService(*cfg)
	if err != nil {
		log.Fatal("Failed to initialize authentication service:", err)
	}

	// Create authentication middleware
	authMiddleware := handler.NewAuthenticationMiddleware(authService)
	authMiddlewareFunc := authMiddleware.MiddlewareFunc()

	// Create wrapper for generated handlers
	wrapper := generated.ServerInterfaceWrapper{
		Handler: taskServer,
	}

	// Register health endpoint without authentication
	router.GET("/health", wrapper.HealthGetHealth)

	// Create a group for protected task endpoints
	taskGroup := router.Group("/tasks")
	taskGroup.Use(authMiddlewareFunc)

	// Register task endpoints with authentication middleware
	taskGroup.GET("", wrapper.TaskGetAllTasks)
	taskGroup.POST("", wrapper.TaskCreateTask)
	taskGroup.GET("/:taskId", wrapper.TaskGetTask)
	taskGroup.PUT("/:taskId", wrapper.TaskUpdateTask)
	taskGroup.DELETE("/:taskId", wrapper.TaskDeleteTask)

	if err := router.Start(":8080"); err != nil {
		log.Fatal("Failed to start server:", err.Error())
	}
}

func initDB(dbConfig config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&repository.TaskModel{}) //nolint:exhaustruct
	if err != nil {
		return nil, err
	}

	return db, nil
}
