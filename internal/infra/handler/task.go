package handler

import (
	"net/http"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	taskHandler "github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
	"github.com/gin-gonic/gin"
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

func (t *TaskServer) TaskGetAllTasks(c *gin.Context) {
	tasks, err := t.controller.GetAllTasks(c.Request.Context())
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	var res []taskHandler.Task

	for _, task := range tasks {
		res = append(res, taskHandler.Task{
			Id:   task.ID(),
			Name: task.Name(),
		})
	}

	c.JSON(200, res)
}

func (t *TaskServer) TaskCreateTask(c *gin.Context) {
	var req taskHandler.TaskCreate

	if err := c.ShouldBindJSON(&req); err != nil {
		details := err.Error()
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return
	}

	if req.Name == "" {
		details := "name field is required and cannot be empty"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return
	}

	task, err := t.controller.CreateTask(c.Request.Context(), req.Name)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	res := taskHandler.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(201, res)
}

func (t *TaskServer) TaskDeleteTask(c *gin.Context, taskId string) {
	if taskId == "" {
		details := "taskId path parameter is required"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return
	}

	err := t.controller.DeleteTask(c.Request.Context(), taskId)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	c.JSON(204, nil)
}

func (t *TaskServer) TaskGetTask(c *gin.Context, taskId string) {
	if taskId == "" {
		details := "taskId path parameter is required"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return
	}

	task, err := t.controller.GetTaskById(c.Request.Context(), taskId)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	if task == nil {
		c.JSON(404, NewNotFoundError("Task not found"))
		return
	}

	res := taskHandler.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(200, res)
}

func (t *TaskServer) TaskUpdateTask(c *gin.Context, taskId string) {
	if taskId == "" {
		details := "taskId path parameter is required"
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return
	}

	var req taskHandler.TaskUpdate

	if err := c.ShouldBindJSON(&req); err != nil {
		details := err.Error()
		c.JSON(400, NewBadRequestError("Bad request", &details))
		return
	}

	task, err := t.controller.GetTaskById(c.Request.Context(), taskId)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	if task == nil {
		c.JSON(404, NewNotFoundError("Task not found"))
		return
	}

	name := task.Name()
	if req.Name != nil {
		name = *req.Name
	}

	task, err = t.controller.UpdateTask(c.Request.Context(), taskId, name)
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	res := taskHandler.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(200, res)
}

// HealthGetHealth implements the ServerInterface for health endpoint
func (t *TaskServer) HealthGetHealth(c *gin.Context) {
	ctx := c.Request.Context()
	
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
}
