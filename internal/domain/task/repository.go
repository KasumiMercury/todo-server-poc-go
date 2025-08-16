package task

import (
	"context"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
)

// TaskRepository defines the interface for task data persistence operations.
type TaskRepository interface {
	FindById(ctx context.Context, creatorID user.UserID, id TaskID) (*Task, error)
	FindAllByUserID(ctx context.Context, creatorID user.UserID) ([]*Task, error)
	Create(ctx context.Context, task *Task) (*Task, error)
	Delete(ctx context.Context, creatorID user.UserID, id TaskID) error
	Update(ctx context.Context, task *Task) (*Task, error)
}
