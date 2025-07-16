package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	taskHandler "github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

type TaskServer struct {
	controller    controller.Task
	healthService service.HealthService
}

func NewTaskServer(ctr controller.Task, healthService service.HealthService) *TaskServer {
	return &TaskServer{
		controller:    ctr,
		healthService: healthService,
	}
}

func (t *TaskServer) TaskGetAllTasks(c echo.Context) error {
	tasks, err := t.controller.GetAllTasks(c.Request().Context())
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return err
	}

	var res []taskHandler.Task

	for _, task := range tasks {
		res = append(res, taskHandler.Task{
			Id:   task.ID(),
			Name: task.Name(),
		})
	}

	c.JSON(200, res)
	return nil
}

func (t *TaskServer) TaskCreateTask(c echo.Context) error {
	var req taskHandler.TaskCreate

	if err := c.Bind(&req); err != nil {
		details := err.Error()
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return err
	}

	if req.Name == "" {
		details := "name field is required and cannot be empty"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return nil
	}

	task, err := t.controller.CreateTask(c.Request().Context(), req.Name)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return err
	}

	res := taskHandler.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(201, res)
	return nil
}

func (t *TaskServer) TaskDeleteTask(c echo.Context, taskId string) error {
	if taskId == "" {
		details := "taskId path parameter is required"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return nil
	}

	err := t.controller.DeleteTask(c.Request().Context(), taskId)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return err
	}

	c.JSON(204, nil)
	return nil
}

func (t *TaskServer) TaskGetTask(c echo.Context, taskId string) error {
	if taskId == "" {
		details := "taskId path parameter is required"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return nil
	}

	task, err := t.controller.GetTaskById(c.Request().Context(), taskId)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return err
	}

	if task == nil {
		c.JSON(404, NewNotFoundError("Task not found"))
		return nil
	}

	res := taskHandler.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(200, res)
	return nil
}

func (t *TaskServer) TaskUpdateTask(c echo.Context, taskId string) error {
	if taskId == "" {
		details := "taskId path parameter is required"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return nil
	}

	var req taskHandler.TaskUpdate

	if err := c.Bind(&req); err != nil {
		details := err.Error()
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return err
	}

	task, err := t.controller.GetTaskById(c.Request().Context(), taskId)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return err
	}

	if task == nil {
		c.JSON(404, NewNotFoundError("Task not found"))
		return nil
	}

	name := task.Name()
	if req.Name != nil {
		name = *req.Name
	}

	task, err = t.controller.UpdateTask(c.Request().Context(), taskId, name)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return err
	}

	res := taskHandler.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(200, res)
	return nil
}

// HealthGetHealth implements the ServerInterface for health endpoint
func (t *TaskServer) HealthGetHealth(c echo.Context) error {
	ctx := c.Request().Context()

	// Handle potential panics from health service
	defer func() {
		if r := recover(); r != nil {
			details := "Health check service unavailable"
			c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal Server Error", &details))
		}
	}()

	healthStatus := t.healthService.CheckHealth(ctx)

	// Convert service model to generated response model
	components := taskHandler.HealthStatus{
		Status:    taskHandler.HealthStatusStatus(healthStatus.Status),
		Timestamp: healthStatus.Timestamp,
		Components: struct {
			Database *taskHandler.HealthComponent `json:"database,omitempty"`
		}{},
	}

	if dbComponent, exists := healthStatus.Components["database"]; exists {
		components.Components.Database = &taskHandler.HealthComponent{
			Status:  taskHandler.HealthComponentStatus(dbComponent.Status),
			Details: &dbComponent.Details,
		}
	}

	// Return appropriate HTTP status code
	statusCode := http.StatusOK
	if healthStatus.Status == "DOWN" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, components)

	return nil
}
