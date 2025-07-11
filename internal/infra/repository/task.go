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

func (t *Task) FindAll(ctx context.Context) ([]*domain.Task, error) {
	return []*domain.Task{
		domain.NewTask("1", "Task 1"),
		domain.NewTask("2", "Task 2"),
		domain.NewTask("3", "Task 3"),
	}, nil
}

func (t *Task) Create(ctx context.Context, title string) (*domain.Task, error) {
	if title == "" {
		panic("Title must not be empty")
	}

	return domain.NewTask("new-task-id", title), nil
}

func (t *Task) Delete(ctx context.Context, id string) error {
	if id == "" {
		panic("ID must not be empty")
	}

	return nil
}

func (t *Task) Update(ctx context.Context, id, title string) (*domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}
	if title == "" {
		panic("Title must not be empty")
	}

	return domain.NewTask(id, title), nil
}
