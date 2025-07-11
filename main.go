package main

import (
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	taskRepo := repository.NewTask()
	taskController := controller.NewTask(*taskRepo)

	taskServer := handler.NewTaskServer(
		*taskController,
	)

	handler.RegisterHandlers(router, taskServer)

	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
