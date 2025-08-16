package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
)

// TaskModel represents the database model for tasks.
// It defines the structure for task data persistence in the database.
type TaskModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	Title     string    `gorm:"not null;type:varchar(255)"`
	CreatorID string    `gorm:"not null;type:varchar(255);index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the database table name for TaskModel.
func (TaskModel) TableName() string {
	return "tasks"
}

// ToDomain converts a TaskModel to a domain Task entity.
func (t TaskModel) ToDomain() (*task.Task, error) {
	taskID, err := task.NewTaskID(t.ID)
	if err != nil {
		return nil, err
	}

	creatorID, err := user.NewUserID(t.CreatorID)
	if err != nil {
		return nil, err
	}

	return task.NewTaskWithoutValidation(taskID, t.Title, creatorID), nil
}

// TaskDB implements the TaskRepository interface using GORM for database operations.
type TaskDB struct {
	db *gorm.DB
}

// NewTaskDB creates a new TaskDB instance with the provided GORM database connection.
func NewTaskDB(db *gorm.DB) *TaskDB {
	return &TaskDB{db: db}
}

func (t *TaskDB) FindById(ctx context.Context, creatorID user.UserID, id task.TaskID) (*task.Task, error) {
	if creatorID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
		return nil, task.ErrTaskIDEmpty
	}

	taskRecord, err := gorm.G[TaskModel](t.db).Where("id = ? AND creator_id = ?", id.String(), creatorID.String()).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, task.ErrTaskNotFound
		}

		return nil, err
	}

	return taskRecord.ToDomain()
}

func (t *TaskDB) FindAllByUserID(ctx context.Context, creatorID user.UserID) ([]*task.Task, error) {
	if creatorID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	taskRecords, err := gorm.G[TaskModel](t.db).Where("creator_id = ?", creatorID.String()).Find(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, task.ErrTaskNotFound
		}

		return nil, err
	}

	if len(taskRecords) == 0 {
		return nil, task.ErrTaskNotFound
	}

	tasks := make([]*task.Task, len(taskRecords))
	for i, record := range taskRecords {
		domainTask, err := record.ToDomain()
		if err != nil {
			return nil, err
		}

		tasks[i] = domainTask
	}

	return tasks, nil
}

func (t *TaskDB) Create(ctx context.Context, taskEntity *task.Task) (*task.Task, error) {
	taskModel := &TaskModel{ //nolint:exhaustruct
		ID:        taskEntity.ID().String(),
		Title:     taskEntity.Title(),
		CreatorID: taskEntity.UserID().String(),
	}

	result := gorm.WithResult()
	if err := gorm.G[TaskModel](t.db, result).Create(ctx, taskModel); err != nil {
		return nil, err
	}

	return taskModel.ToDomain()
}

func (t *TaskDB) Delete(ctx context.Context, creatorID user.UserID, id task.TaskID) error {
	if creatorID.IsEmpty() {
		return user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
		return task.ErrTaskIDEmpty
	}

	if _, err := gorm.G[TaskModel](t.db).Where("id = ? AND creator_id = ?", id.String(), creatorID.String()).Delete(ctx); err != nil {
		return err
	}

	return nil
}

func (t *TaskDB) Update(ctx context.Context, taskEntity *task.Task) (*task.Task, error) {
	taskModel := &TaskModel{ //nolint:exhaustruct
		ID:        taskEntity.ID().String(),
		Title:     taskEntity.Title(),
		CreatorID: taskEntity.UserID().String(),
	}

	if _, err := gorm.G[TaskModel](t.db).Where("id = ? AND creator_id = ?", taskEntity.ID().String(), taskEntity.UserID().String()).Updates(ctx, *taskModel); err != nil {
		return nil, err
	}

	return taskModel.ToDomain()
}
