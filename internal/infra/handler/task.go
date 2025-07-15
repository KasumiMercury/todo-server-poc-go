package handler

import (
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/task"
	"github.com/gin-gonic/gin"
)

type TaskServer struct {
	controller controller.Task
}

func NewTaskServer(ctr controller.Task) *TaskServer {
	return &TaskServer{
		controller: ctr,
	}
}

func (t *TaskServer) TaskGetAllTasks(c *gin.Context) {
	tasks, err := t.controller.GetAllTasks(c.Request.Context())
	if err != nil {
		details := err.Error()
		c.JSON(500, NewInternalServerError("Internal server error", &details))
		return
	}

	var res []task.Task

	for _, task := range tasks {
		res = append(res, task.Task{
			Id:   task.ID(),
			Name: task.Name(),
		})
	}

	c.JSON(200, res)
}

func (t *TaskServer) TaskCreateTask(c *gin.Context) {
	var req task.TaskCreate

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

	res := task.Task{
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

	res := task.Task{
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

	var req task.TaskUpdate

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

	res := task.Task{
		Id:   task.ID(),
		Name: task.Name(),
	}

	c.JSON(200, res)
}
