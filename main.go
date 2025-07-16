package main

import (
	"fmt"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"log"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
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

	// Setup Prometheus metrics
	router.Use(echoprometheus.NewMiddleware(cfg.ServiceName))
	go func() {
		// Start metrics server on a separated port
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(":8081"); err != nil {
			log.Fatal("Failed to start metrics server:", err)
		}
	}()

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

	// Setup JWT middleware for protected endpoints
	jwtMiddleware := handler.JWTMiddleware(cfg.JWTSecret)

	// Create wrapper for generated handlers
	wrapper := generated.ServerInterfaceWrapper{
		Handler: taskServer,
	}

	// Register health endpoint without authentication
	router.GET("/health", wrapper.HealthGetHealth)

	// Create a group for protected task endpoints
	taskGroup := router.Group("/tasks")
	taskGroup.Use(jwtMiddleware)

	// Register task endpoints with JWT middleware
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

	err = db.AutoMigrate(&repository.TaskModel{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
