package controller

import (
	task2 "github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"golang.org/x/net/context"
)

type Task struct {
	taskRepo task2.TaskRepository
}

func NewTask(taskRepo task2.TaskRepository) *Task {
	return &Task{
		taskRepo: taskRepo,
	}
}

func (t *Task) GetTaskById(ctx context.Context, id string) (*task2.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}

	task, err := t.taskRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (t *Task) GetAllTasks(ctx context.Context) ([]*task2.Task, error) {
	tasks, err := t.taskRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *Task) CreateTask(ctx context.Context, name string) (*task2.Task, error) {
	if name == "" {
		panic("Name must not be empty")
	}

	task, err := t.taskRepo.Create(ctx, name)
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

func (t *Task) UpdateTask(ctx context.Context, id, name string) (*task2.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}
	if name == "" {
		panic("Name must not be empty")
	}

	task, err := t.taskRepo.Update(ctx, id, name)
	if err != nil {
		return nil, err
	}

	return task, nil
}
