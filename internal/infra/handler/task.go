package handler

import (
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
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
		c.JSON(500, gin.H{"error": "failed to get all tasks: " + err.Error()})
		return
	}

	var res []struct {
		Id    string `json:"id"`
		Title string `json:"title"`
	}

	for _, task := range tasks {
		res = append(res, struct {
			Id    string `json:"id"`
			Title string `json:"title"`
		}{
			Id:    task.ID(),
			Title: task.Title(),
		})
	}

	c.JSON(200, res)
}

func (t *TaskServer) TaskCreateTask(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(400, gin.H{"error": "title must not be empty"})
		return
	}

	task, err := t.controller.CreateTask(c.Request.Context(), req.Title)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create task: " + err.Error()})
		return
	}

	res := struct {
		Id    string `json:"id"`
		Title string `json:"title"`
	}{
		Id:    task.ID(),
		Title: task.Title(),
	}

	c.JSON(201, res)
}

func (t *TaskServer) TaskDeleteTask(c *gin.Context, taskId string) {
	if taskId == "" {
		c.JSON(400, gin.H{"error": "ID must not be empty"})
		return
	}

	err := t.controller.DeleteTask(c.Request.Context(), taskId)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete task: " + err.Error()})
		return
	}

	c.JSON(204, nil)
}

func (t *TaskServer) TaskGetTask(c *gin.Context, taskId string) {
	if taskId == "" {
		c.JSON(400, gin.H{"error": "ID must not be empty"})
		return
	}

	task, err := t.controller.GetTaskById(c.Request.Context(), taskId)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get task by ID: " + err.Error()})
		return
	}

	if task == nil {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	res := struct {
		Id    string `json:"id"`
		Title string `json:"title"`
	}{
		Id:    task.ID(),
		Title: task.Title(),
	}

	c.JSON(200, res)
}

func (t *TaskServer) TaskUpdateTask(c *gin.Context, taskId string) {
	if taskId == "" {
		c.JSON(400, gin.H{"error": "ID must not be empty"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(400, gin.H{"error": "title must not be empty"})
		return
	}

	task, err := t.controller.UpdateTask(c.Request.Context(), taskId, req.Title)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update task: " + err.Error()})
		return
	}

	res := struct {
		Id    string `json:"id"`
		Title string `json:"title"`
	}{
		Id:    task.ID(),
		Title: task.Title(),
	}

	c.JSON(200, res)
}
