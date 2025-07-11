package repository

import (
	"context"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain"
)

type Task struct{}

func NewTask() *Task {
	return &Task{}
}

func (t *Task) FindById(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}

	return domain.NewTask(
		id,
		"Sample Task",
	), nil
}
