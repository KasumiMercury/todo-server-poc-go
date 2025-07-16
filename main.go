package main

import (
	"fmt"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/labstack/echo/v4"
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

	//router := gin.Default()
	router := echo.New()

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

	// Register health endpoint without authentication
	//router.GET("/health", taskServer.HealthGetHealth)

	// TODO: setup JWT middleware
	// Register task endpoints with JWT middleware
	//jwtMiddleware := handler.JWTMiddleware(cfg.JWTSecret)

	// Register individual task endpoints with JWT middleware
	//taskWrapper := generated.ServerInterfaceWrapper{
	//	Handler: taskServer,
	//	HandlerMiddlewares: []generated.MiddlewareFunc{
	//		generated.MiddlewareFunc(jwtMiddleware),
	//	},
	//}

	//router.GET("/tasks", taskWrapper.TaskGetAllTasks)
	//router.POST("/tasks", taskWrapper.TaskCreateTask)
	//router.GET("/tasks/:taskId", taskWrapper.TaskGetTask)
	//router.PUT("/tasks/:taskId", taskWrapper.TaskUpdateTask)
	//router.DELETE("/tasks/:taskId", taskWrapper.TaskDeleteTask)

	generated.RegisterHandlers(router, taskServer)

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
