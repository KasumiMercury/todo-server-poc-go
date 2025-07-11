package handler

import (
	"fmt"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/gin-gonic/gin"
)

type Task struct {
	ctr controller.Task
}

func NewTask(ctr controller.Task) *Task {
	return &Task{
		ctr: ctr,
	}
}

func (t *Task) Register(r *gin.Engine) {
	r.GET("/task/:id", func(c *gin.Context) {
		if err := t.GetTaskById(c); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
	})

}

func (t *Task) GetTaskById(c *gin.Context) error {
	id := c.Param("id")
	if id == "" {
		return fmt.Errorf("ID must not be empty")
	}

	task, err := t.ctr.GetTaskById(c.Request.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to get task by ID: %w", err)
	}

	if task == nil {
		return fmt.Errorf("task not found")
	}

	res := struct {
		Id    string `json:"id"`
		Title string `json:"title"`
	}{
		Id:    task.ID(),
		Title: task.Title(),
	}

	c.JSON(200, res)
	return nil
}

//func (t *Task) GetTaskById(ctx context.Context, id string) (domain.Task, error) {
//	if id == "" {
//		panic("ID must not be empty")
//	}
//
//	task, err := t.ctr.GetTaskById(ctx, id)
//	if err != nil {
//		return domain.Task{}, err
//	}
//
//	return *task, nil
//}
