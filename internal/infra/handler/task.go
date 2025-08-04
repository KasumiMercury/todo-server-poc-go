package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	taskDomain "github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	taskHandler "github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

// TaskServer handles HTTP requests for task operations.
type TaskServer struct {
	controller    controller.Task
	healthService service.HealthService
}

// NewTaskServer creates a new TaskServer with the provided controller and health service.
func NewTaskServer(ctr controller.Task, healthService service.HealthService) *TaskServer {
	return &TaskServer{
		controller:    ctr,
		healthService: healthService,
	}
}

// isDomainValidationError checks if the error is a domain validation error
func isDomainValidationError(err error) bool {
	return errors.Is(err, taskDomain.ErrTitleEmpty) || errors.Is(err, taskDomain.ErrTitleTooLong)
}

func (t *TaskServer) TaskGetAllTasks(c echo.Context) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		details := "User ID not found in token"

		return c.JSON(401, NewUnauthorizedError("Unauthorized", &details))
	}

	tasks, err := t.controller.GetAllTasks(c.Request().Context(), userID)
	if err != nil {
		details := err.Error()

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	res := make([]taskHandler.Task, 0, len(tasks))

	for _, task := range tasks {
		res = append(res, taskHandler.Task{
			Id:    task.ID(),
			Title: task.Title(),
		})
	}

	return c.JSON(200, res)
}

func (t *TaskServer) TaskCreateTask(c echo.Context) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		details := "User ID not found in token"

		return c.JSON(401, NewUnauthorizedError("Unauthorized", &details))
	}

	var req taskHandler.TaskCreate

	if err := c.Bind(&req); err != nil {
		details := err.Error()

		return c.JSON(400, NewBadRequestError("Bad request", &details))
	}

	if req.Title == "" {
		details := "title field is required and cannot be empty"

		return c.JSON(400, NewBadRequestError("Bad request", &details))
	}

	task, err := t.controller.CreateTask(c.Request().Context(), userID, req.Title)
	if err != nil {
		details := err.Error()
		if isDomainValidationError(err) {
			return c.JSON(400, NewBadRequestError("Bad request", &details))
		}

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	res := taskHandler.Task{
		Id:    task.ID(),
		Title: task.Title(),
	}

	return c.JSON(201, res)
}

func (t *TaskServer) TaskDeleteTask(c echo.Context, taskId string) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		details := "User ID not found in token"

		return c.JSON(401, NewUnauthorizedError("Unauthorized", &details))
	}

	if taskId == "" {
		details := "taskId path parameter is required"

		return c.JSON(400, NewBadRequestError("Bad request", &details))
	}

	// Check if task exists before deletion
	task, err := t.controller.GetTaskById(c.Request().Context(), userID, taskId)
	if err != nil {
		if errors.Is(err, taskDomain.ErrTaskNotFound) {
			return c.JSON(404, NewNotFoundError("Task not found"))
		}

		details := err.Error()

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	if task == nil {
		return c.JSON(404, NewNotFoundError("Task not found"))
	}

	err = t.controller.DeleteTask(c.Request().Context(), userID, taskId)
	if err != nil {
		details := err.Error()

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	return c.JSON(204, nil)
}

func (t *TaskServer) TaskGetTask(c echo.Context, taskId string) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		details := "User ID not found in token"

		return c.JSON(401, NewUnauthorizedError("Unauthorized", &details))
	}

	if taskId == "" {
		details := "taskId path parameter is required"

		return c.JSON(400, NewBadRequestError("Bad request", &details))
	}

	task, err := t.controller.GetTaskById(c.Request().Context(), userID, taskId)
	if err != nil {
		if errors.Is(err, taskDomain.ErrTaskNotFound) {
			return c.JSON(404, NewNotFoundError("Task not found"))
		}

		details := err.Error()

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	if task == nil {
		return c.JSON(404, NewNotFoundError("Task not found"))
	}

	res := taskHandler.Task{
		Id:    task.ID(),
		Title: task.Title(),
	}

	return c.JSON(200, res)
}

func (t *TaskServer) TaskUpdateTask(c echo.Context, taskId string) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		details := "User ID not found in token"

		return c.JSON(401, NewUnauthorizedError("Unauthorized", &details))
	}

	if taskId == "" {
		details := "taskId path parameter is required"

		return c.JSON(400, NewBadRequestError("Bad request", &details))
	}

	var req taskHandler.TaskUpdate

	if err := c.Bind(&req); err != nil {
		details := err.Error()

		return c.JSON(400, NewBadRequestError("Bad request", &details))
	}

	task, err := t.controller.GetTaskById(c.Request().Context(), userID, taskId)
	if err != nil {
		if errors.Is(err, taskDomain.ErrTaskNotFound) {
			return c.JSON(404, NewNotFoundError("Task not found"))
		}

		details := err.Error()

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	if task == nil {
		return c.JSON(404, NewNotFoundError("Task not found"))
	}

	title := task.Title()
	if req.Title != nil {
		title = *req.Title
	}

	task, err = t.controller.UpdateTask(c.Request().Context(), userID, taskId, title)
	if err != nil {
		details := err.Error()
		if isDomainValidationError(err) {
			return c.JSON(400, NewBadRequestError("Bad request", &details))
		}

		return c.JSON(500, NewInternalServerError("Internal server error", &details))
	}

	res := taskHandler.Task{
		Id:    task.ID(),
		Title: task.Title(),
	}

	return c.JSON(200, res)
}

// HealthGetHealth implements the ServerInterface for health endpoint
func (t *TaskServer) HealthGetHealth(c echo.Context) error {
	ctx := c.Request().Context()

	healthStatus := t.healthService.CheckHealth(ctx)

	// Convert service model to generated response model
	components := taskHandler.HealthStatus{
		Status:    taskHandler.HealthStatusStatus(healthStatus.Status),
		Timestamp: healthStatus.Timestamp,
		Components: struct { //nolint:exhaustruct
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

	return c.JSON(statusCode, components)
}
