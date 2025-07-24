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

func (t *Task) GetTaskById(ctx context.Context, userID, id string) (*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}
	if id == "" {
		panic("ID must not be empty")
	}

	taskItem, err := t.taskRepo.FindById(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}

func (t *Task) GetAllTasks(ctx context.Context, userID string) ([]*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}

	tasks, err := t.taskRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *Task) CreateTask(ctx context.Context, userID, title string) (*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}
	if title == "" {
		panic("Title must not be empty")
	}

	taskItem, err := t.taskRepo.Create(ctx, userID, title)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}

func (t *Task) DeleteTask(ctx context.Context, userID, id string) error {
	if userID == "" {
		panic("UserID must not be empty")
	}
	if id == "" {
		panic("ID must not be empty")
	}

	err := t.taskRepo.Delete(ctx, userID, id)
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) UpdateTask(ctx context.Context, userID, id, title string) (*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}
	if id == "" {
		panic("ID must not be empty")
	}
	if title == "" {
		panic("Title must not be empty")
	}

	taskItem, err := t.taskRepo.Update(ctx, userID, id, title)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}
