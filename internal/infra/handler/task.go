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

func (t *TaskServer) TaskGetAllTasks(c *gin.Context) {}

func (t *TaskServer) TaskCreateTask(c *gin.Context) {}

func (t *TaskServer) TaskDeleteTask(c *gin.Context, taskId string) {}

func (t *TaskServer) TaskGetTask(c *gin.Context, taskId string) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "ID must not be empty"})
		return
	}

	task, err := t.controller.GetTaskById(c.Request.Context(), id)
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

func (t *TaskServer) TaskUpdateTask(c *gin.Context, taskId string) {}
