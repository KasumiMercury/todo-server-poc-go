package controller

import (
	"errors"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
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
func (t *Task) GetTaskById(ctx context.Context, userID user.UserID, id task.TaskID) (*task.Task, error) {
	if userID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
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
func (t *Task) GetAllTasks(ctx context.Context, userID user.UserID) ([]*task.Task, error) {
	if userID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
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
func (t *Task) CreateTask(ctx context.Context, userID user.UserID, title string) (*task.Task, error) {
	if userID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	taskEntity, err := task.NewTask(task.GenerateTaskID(), title, userID)
	if err != nil {
		return nil, err
	}

	taskItem, err := t.taskRepo.Create(ctx, taskEntity)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}

// DeleteTask removes a task by its ID for the given user.
func (t *Task) DeleteTask(ctx context.Context, userID user.UserID, id task.TaskID) error {
	if userID.IsEmpty() {
		return user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
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
func (t *Task) UpdateTask(ctx context.Context, userID user.UserID, id task.TaskID, title string) (*task.Task, error) {
	if userID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
		return nil, task.ErrTaskIDEmpty
	}

	taskEntity, err := task.NewTask(id, title, userID)
	if err != nil {
		return nil, err
	}

	taskItem, err := t.taskRepo.Update(ctx, taskEntity)
	if err != nil {
		return nil, err
	}

	return taskItem, nil
}
