package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

// APIServer handles HTTP requests for all API operations.
// It implements the ServerInterface and delegates to specialized handlers.
type APIServer struct {
	taskHandler   *TaskHandler
	healthHandler *HealthHandler
}

// NewAPIServer creates a new APIServer with the provided handlers.
func NewAPIServer(taskController controller.Task, healthService service.HealthService) *APIServer {
	return &APIServer{
		taskHandler:   NewTaskHandler(taskController),
		healthHandler: NewHealthHandler(healthService),
	}
}

// HealthGetHealth implements the ServerInterface for health endpoint by delegating to HealthHandler
func (s *APIServer) HealthGetHealth(c echo.Context) error {
	return s.healthHandler.GetHealth(c)
}

// TaskGetAllTasks implements the ServerInterface for task operations by delegating to TaskHandler
func (s *APIServer) TaskGetAllTasks(c echo.Context) error {
	return s.taskHandler.GetAllTasks(c)
}

// TaskCreateTask implements the ServerInterface for task creation by delegating to TaskHandler
func (s *APIServer) TaskCreateTask(c echo.Context) error {
	return s.taskHandler.CreateTask(c)
}

// TaskDeleteTask implements the ServerInterface for task deletion by delegating to TaskHandler
func (s *APIServer) TaskDeleteTask(c echo.Context, taskId string) error {
	return s.taskHandler.DeleteTask(c, taskId)
}

// TaskGetTask implements the ServerInterface for getting a specific task by delegating to TaskHandler
func (s *APIServer) TaskGetTask(c echo.Context, taskId string) error {
	return s.taskHandler.GetTask(c, taskId)
}

// TaskUpdateTask implements the ServerInterface for task updates by delegating to TaskHandler
func (s *APIServer) TaskUpdateTask(c echo.Context, taskId string) error {
	return s.taskHandler.UpdateTask(c, taskId)
}
