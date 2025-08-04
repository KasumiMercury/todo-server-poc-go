package repository

import (
	"context"
	"errors"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
func (t TaskModel) ToDomain() *task.Task {
	return task.NewTask(t.ID, t.Title, t.UserID)
}

// TaskDB implements the TaskRepository interface using GORM for database operations.
type TaskDB struct {
	db *gorm.DB
}

// NewTaskDB creates a new TaskDB instance with the provided GORM database connection.
func NewTaskDB(db *gorm.DB) *TaskDB {
	return &TaskDB{db: db}
}

func (t *TaskDB) FindById(ctx context.Context, userID, id string) (*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}

	if id == "" {
		panic("ID must not be empty")
	}

	taskRecord, err := gorm.G[TaskModel](t.db).Where("id = ? AND user_id = ?", id, userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, task.ErrTaskNotFound
		}

		return nil, err
	}

	return taskRecord.ToDomain(), nil
}

func (t *TaskDB) FindAllByUserID(ctx context.Context, userID string) ([]*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}

	taskRecords, err := gorm.G[TaskModel](t.db).Where("user_id = ?", userID).Find(ctx)
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
	for i, taskModel := range taskRecords {
		tasks[i] = taskModel.ToDomain()
	}

	return tasks, nil
}

func (t *TaskDB) Create(ctx context.Context, userID, title string) (*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}

	if title == "" {
		panic("Title must not be empty")
	}

	taskModel := &TaskModel{
		ID:        uuid.New().String(),
		Title:     title,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := gorm.WithResult()
	if err := gorm.G[TaskModel](t.db, result).Create(ctx, taskModel); err != nil {
		return nil, err
	}

	return taskModel.ToDomain(), nil
}

func (t *TaskDB) Delete(ctx context.Context, userID, id string) error {
	if userID == "" {
		panic("UserID must not be empty")
	}

	if id == "" {
		panic("ID must not be empty")
	}

	if _, err := gorm.G[TaskModel](t.db).Where("id = ? AND user_id = ?", id, userID).Delete(ctx); err != nil {
		return err
	}

	return nil
}

func (t *TaskDB) Update(ctx context.Context, userID, id, title string) (*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}

	if id == "" {
		panic("ID must not be empty")
	}

	if title == "" {
		panic("Title must not be empty")
	}

	taskModel := &TaskModel{ //nolint:exhaustruct
		ID:        id,
		Title:     title,
		UserID:    userID,
		UpdatedAt: time.Now(),
	}

	if _, err := gorm.G[TaskModel](t.db).Where("id = ? AND user_id = ?", id, userID).Updates(ctx, *taskModel); err != nil {
		return nil, err
	}

	return taskModel.ToDomain(), nil
}
