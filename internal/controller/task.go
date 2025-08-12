package controller

import (
	"errors"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"golang.org/x/net/context"
)

// Task represents the task controller that handles business logic for task operations.
type Task struct {
	taskRepo task.TaskRepository
}

// NewTask creates a new Task controller with the provided repository.
func NewTask(taskRepo task.TaskRepository) *Task {
	return &Task{
		taskRepo: taskRepo,
	}
}

// GetTaskById retrieves a specific task by its ID for the given user.
func (t *Task) GetTaskById(ctx context.Context, userID, id string) (*task.Task, error) {
	if userID == "" {
		return nil, task.ErrUserIDEmpty
	}

	if id == "" {
		return nil, task.ErrTaskIDEmpty
	}

	taskItem, err := t.taskRepo.FindById(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}

// GetAllTasks retrieves all tasks for the given user.
// It returns an empty slice if no tasks are found.
func (t *Task) GetAllTasks(ctx context.Context, userID string) ([]*task.Task, error) {
	if userID == "" {
		return nil, task.ErrUserIDEmpty
	}

	tasks, err := t.taskRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			return []*task.Task{}, nil
		}

		return nil, err
	}

	return tasks, nil
}

// CreateTask creates a new task with the provided title for the given user.
// It validates the title using domain validation rules.
func (t *Task) CreateTask(ctx context.Context, userID, title string) (*task.Task, error) {
	if userID == "" {
		return nil, task.ErrUserIDEmpty
	}

	// Validate title using domain validation
	_, err := task.NewTaskWithValidation("", title, userID)
	if err != nil {
		return nil, err
	}

	taskItem, err := t.taskRepo.Create(ctx, userID, title)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}

// DeleteTask removes a task by its ID for the given user.
func (t *Task) DeleteTask(ctx context.Context, userID, id string) error {
	if userID == "" {
		return task.ErrUserIDEmpty
	}

	if id == "" {
		return task.ErrTaskIDEmpty
	}

	err := t.taskRepo.Delete(ctx, userID, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateTask updates a task's title for the given user.
// It validates the new title using domain validation rules.
func (t *Task) UpdateTask(ctx context.Context, userID, id, title string) (*task.Task, error) {
	if userID == "" {
		return nil, task.ErrUserIDEmpty
	}

	if id == "" {
		return nil, task.ErrTaskIDEmpty
	}

	// Validate title using domain validation
	_, err := task.NewTaskWithValidation("", title, userID)
	if err != nil {
		return nil, err
	}

	taskItem, err := t.taskRepo.Update(ctx, userID, id, title)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}
