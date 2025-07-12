package main

import (
	"fmt"
	"log"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := initDB(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	router := gin.Default()

	// Add CORS middleware globally
	router.Use(handler.CORSMiddleware())

	taskRepo := repository.NewTaskDB(db)
	taskController := controller.NewTask(taskRepo)

	taskServer := handler.NewTaskServer(
		*taskController,
	)

	// Register handlers with JWT middleware
	jwtMiddleware := handler.JWTMiddleware(cfg.JWTSecret)
	handler.RegisterHandlersWithOptions(router, taskServer, handler.GinServerOptions{
		Middlewares: []handler.MiddlewareFunc{
			handler.MiddlewareFunc(jwtMiddleware),
		},
	})

	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
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
