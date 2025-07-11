package handler

import (
	"context"
	"fmt"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain"
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
		id := c.Param("id")
		task, err := t.GetTaskById(c.Request.Context(), id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(task)

		res := struct {
			Id    string `json:"id"`
			Title string `json:"title"`
		}{
			Id:    task.ID(),
			Title: task.Title(),
		}

		c.JSON(200, res)
	})

}

func (t *Task) GetTaskById(ctx context.Context, id string) (domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}

	task, err := t.ctr.GetTaskById(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	return *task, nil
}
