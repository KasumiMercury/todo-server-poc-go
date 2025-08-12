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
	UserID    string    `gorm:"not null;type:varchar(255);index"` // Note: This stores userId from JWT sub claim, which may change in the future
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

	userID, err := user.NewUserID(t.UserID)
	if err != nil {
		return nil, err
	}

	return task.NewTask(taskID, t.Title, userID), nil
}

// TaskDB implements the TaskRepository interface using GORM for database operations.
type TaskDB struct {
	db *gorm.DB
}

// NewTaskDB creates a new TaskDB instance with the provided GORM database connection.
func NewTaskDB(db *gorm.DB) *TaskDB {
	return &TaskDB{db: db}
}

func (t *TaskDB) FindById(ctx context.Context, userID user.UserID, id task.TaskID) (*task.Task, error) {
	if userID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
		return nil, task.ErrTaskIDEmpty
	}

	taskRecord, err := gorm.G[TaskModel](t.db).Where("id = ? AND user_id = ?", id.String(), userID.String()).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, task.ErrTaskNotFound
		}

		return nil, err
	}

	return taskRecord.ToDomain()
}

func (t *TaskDB) FindAllByUserID(ctx context.Context, userID user.UserID) ([]*task.Task, error) {
	if userID.IsEmpty() {
		return nil, user.ErrUserIDEmpty
	}

	taskRecords, err := gorm.G[TaskModel](t.db).Where("user_id = ?", userID.String()).Find(ctx)
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
		ID:     taskEntity.ID().String(),
		Title:  taskEntity.Title(),
		UserID: taskEntity.UserID().String(),
	}

	result := gorm.WithResult()
	if err := gorm.G[TaskModel](t.db, result).Create(ctx, taskModel); err != nil {
		return nil, err
	}

	return taskModel.ToDomain()
}

func (t *TaskDB) Delete(ctx context.Context, userID user.UserID, id task.TaskID) error {
	if userID.IsEmpty() {
		return user.ErrUserIDEmpty
	}

	if id.IsEmpty() {
		return task.ErrTaskIDEmpty
	}

	if _, err := gorm.G[TaskModel](t.db).Where("id = ? AND user_id = ?", id.String(), userID.String()).Delete(ctx); err != nil {
		return err
	}

	return nil
}

func (t *TaskDB) Update(ctx context.Context, taskEntity *task.Task) (*task.Task, error) {
	taskModel := &TaskModel{ //nolint:exhaustruct
		ID:     taskEntity.ID().String(),
		Title:  taskEntity.Title(),
		UserID: taskEntity.UserID().String(),
	}

	if _, err := gorm.G[TaskModel](t.db).Where("id = ? AND user_id = ?", taskEntity.ID().String(), taskEntity.UserID().String()).Updates(ctx, *taskModel); err != nil {
		return nil, err
	}

	return taskModel.ToDomain()
}
