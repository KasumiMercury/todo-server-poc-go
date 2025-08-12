package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	taskDomain "github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
	taskHandler "github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
)

// TaskHandler handles HTTP requests for task operations.
type TaskHandler struct {
	controller controller.Task
}

// NewTaskHandler creates a new TaskHandler with the provided controller.
func NewTaskHandler(ctr controller.Task) *TaskHandler {
	return &TaskHandler{
		controller: ctr,
	}
}

// isDomainValidationError checks if the error is a domain validation error
func isDomainValidationError(err error) bool {
	return errors.Is(err, taskDomain.ErrTitleEmpty) ||
		errors.Is(err, taskDomain.ErrTitleTooLong) ||
		errors.Is(err, user.ErrUserIDEmpty) ||
		errors.Is(err, taskDomain.ErrTaskIDEmpty) ||
		errors.Is(err, taskDomain.ErrInvalidTaskIDFormat) ||
		errors.Is(err, user.ErrInvalidUserIDFormat)
}

// extractUserID extracts user ID from JWT context
func (t *TaskHandler) extractUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in token")
	}

	return userID, nil
}

func (t *TaskHandler) GetAllTasks(c echo.Context) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, err := t.extractUserID(c)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
	}

	// Convert string userID to domain UserID
	domainUserID, err := user.NewUserID(userID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid user ID format", &details))
	}

	tasks, err := t.controller.GetAllTasks(c.Request().Context(), domainUserID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	res := make([]taskHandler.Task, 0, len(tasks))

	for _, task := range tasks {
		res = append(res, taskHandler.Task{
			Id:    task.ID().String(),
			Title: task.Title(),
		})
	}

	return c.JSON(http.StatusOK, res)
}

func (t *TaskHandler) CreateTask(c echo.Context) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, err := t.extractUserID(c)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
	}

	var req taskHandler.TaskCreate

	if err := c.Bind(&req); err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
	}

	if req.Title == "" {
		details := "title field is required and cannot be empty"

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
	}

	// Convert string userID to domain UserID
	domainUserID, err := user.NewUserID(userID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid user ID format", &details))
	}

	task, err := t.controller.CreateTask(c.Request().Context(), domainUserID, req.Title)
	if err != nil {
		details := err.Error()
		if isDomainValidationError(err) {
			return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
		}

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	res := taskHandler.Task{
		Id:    task.ID().String(),
		Title: task.Title(),
	}

	return c.JSON(http.StatusCreated, res)
}

func (t *TaskHandler) DeleteTask(c echo.Context, taskId string) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, err := t.extractUserID(c)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
	}

	if taskId == "" {
		details := "taskId path parameter is required"

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
	}

	// Convert string userID to domain UserID
	domainUserID, err := user.NewUserID(userID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid user ID format", &details))
	}

	// Convert string taskId to domain TaskID
	domainTaskID, err := taskDomain.NewTaskID(taskId)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid task ID format", &details))
	}

	// Check if task exists before deletion
	task, err := t.controller.GetTaskById(c.Request().Context(), domainUserID, domainTaskID)
	if err != nil {
		if errors.Is(err, taskDomain.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, NewNotFoundError("Task not found"))
		}

		details := err.Error()

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	if task == nil {
		return c.JSON(http.StatusNotFound, NewNotFoundError("Task not found"))
	}

	err = t.controller.DeleteTask(c.Request().Context(), domainUserID, domainTaskID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	return c.JSON(http.StatusNoContent, nil)
}

func (t *TaskHandler) GetTask(c echo.Context, taskId string) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, err := t.extractUserID(c)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
	}

	if taskId == "" {
		details := "taskId path parameter is required"

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
	}

	// Convert string userID to domain UserID
	domainUserID, err := user.NewUserID(userID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid user ID format", &details))
	}

	// Convert string taskId to domain TaskID
	domainTaskID, err := taskDomain.NewTaskID(taskId)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid task ID format", &details))
	}

	task, err := t.controller.GetTaskById(c.Request().Context(), domainUserID, domainTaskID)
	if err != nil {
		if errors.Is(err, taskDomain.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, NewNotFoundError("Task not found"))
		}

		details := err.Error()

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	if task == nil {
		return c.JSON(http.StatusNotFound, NewNotFoundError("Task not found"))
	}

	res := taskHandler.Task{
		Id:    task.ID().String(),
		Title: task.Title(),
	}

	return c.JSON(http.StatusOK, res)
}

func (t *TaskHandler) UpdateTask(c echo.Context, taskId string) error {
	// Extract userID from JWT sub claim (set by JWTMiddleware in jwt.go)
	userID, err := t.extractUserID(c)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
	}

	if taskId == "" {
		details := "taskId path parameter is required"

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
	}

	var req taskHandler.TaskUpdate

	if err := c.Bind(&req); err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
	}

	// Convert string userID to domain UserID
	domainUserID, err := user.NewUserID(userID)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid user ID format", &details))
	}

	// Convert string taskId to domain TaskID
	domainTaskID, err := taskDomain.NewTaskID(taskId)
	if err != nil {
		details := err.Error()

		return c.JSON(http.StatusBadRequest, NewBadRequestError("Invalid task ID format", &details))
	}

	task, err := t.controller.GetTaskById(c.Request().Context(), domainUserID, domainTaskID)
	if err != nil {
		if errors.Is(err, taskDomain.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, NewNotFoundError("Task not found"))
		}

		details := err.Error()

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	if task == nil {
		return c.JSON(http.StatusNotFound, NewNotFoundError("Task not found"))
	}

	title := task.Title()
	if req.Title != nil {
		title = *req.Title
	}

	task, err = t.controller.UpdateTask(c.Request().Context(), domainUserID, domainTaskID, title)
	if err != nil {
		details := err.Error()
		if isDomainValidationError(err) {
			return c.JSON(http.StatusBadRequest, NewBadRequestError("Bad request", &details))
		}

		return c.JSON(http.StatusInternalServerError, NewInternalServerError("Internal server error", &details))
	}

	res := taskHandler.Task{
		Id:    task.ID().String(),
		Title: task.Title(),
	}

	return c.JSON(http.StatusOK, res)
}
