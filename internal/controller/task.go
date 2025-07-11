package controller

import (
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/repository"
	"golang.org/x/net/context"
)

type Task struct {
	taskRepo repository.Task
}

func NewTask(taskRepo repository.Task) *Task {
	return &Task{
		taskRepo: taskRepo,
	}
}

func (t *Task) GetTaskById(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}

	task, err := t.taskRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	return task, nil
}
