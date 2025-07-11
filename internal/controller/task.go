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

func (t *Task) GetAllTasks(ctx context.Context) ([]*domain.Task, error) {
	tasks, err := t.taskRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *Task) CreateTask(ctx context.Context, title string) (*domain.Task, error) {
	if title == "" {
		panic("Title must not be empty")
	}

	task, err := t.taskRepo.Create(ctx, title)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (t *Task) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		panic("ID must not be empty")
	}

	err := t.taskRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) UpdateTask(ctx context.Context, id, title string) (*domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}
	if title == "" {
		panic("Title must not be empty")
	}

	task, err := t.taskRepo.Update(ctx, id, title)
	if err != nil {
		return nil, err
	}

	return task, nil
}
