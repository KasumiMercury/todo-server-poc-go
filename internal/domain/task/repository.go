package task

import (
	"context"
)

// TaskRepository defines the interface for task data persistence operations.
type TaskRepository interface {
	FindById(ctx context.Context, userID, id string) (*Task, error)
	FindAllByUserID(ctx context.Context, userID string) ([]*Task, error)
	Create(ctx context.Context, userID, title string) (*Task, error)
	Delete(ctx context.Context, userID, id string) error
	Update(ctx context.Context, userID, id, title string) (*Task, error)
}
