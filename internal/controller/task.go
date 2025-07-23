package controller

import (
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"golang.org/x/net/context"
)

type Task struct {
	taskRepo task.TaskRepository
}

func NewTask(taskRepo task.TaskRepository) *Task {
	return &Task{
		taskRepo: taskRepo,
	}
}

func (t *Task) GetTaskById(ctx context.Context, id string) (*task.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}

	taskItem, err := t.taskRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}

func (t *Task) GetAllTasks(ctx context.Context) ([]*task.Task, error) {
	tasks, err := t.taskRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *Task) CreateTask(ctx context.Context, title string) (*task.Task, error) {
	if title == "" {
		panic("Title must not be empty")
	}

	taskItem, err := t.taskRepo.Create(ctx, title)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
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

func (t *Task) UpdateTask(ctx context.Context, id, title string) (*task.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}
	if title == "" {
		panic("Title must not be empty")
	}

	taskItem, err := t.taskRepo.Update(ctx, id, title)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}
