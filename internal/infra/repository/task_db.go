package repository

import (
	"context"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	Title     string    `gorm:"not null;type:varchar(255)"`
	UserID    string    `gorm:"not null;type:varchar(255);index"` // Note: This stores userId from JWT sub claim, which may change in the future
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

func (t *TaskModel) ToDomain() *task.Task {
	return task.NewTask(t.ID, t.Title, t.UserID)
}

func NewTaskModelFromDomain(task *task.Task) *TaskModel {
	return &TaskModel{
		ID:     task.ID(),
		Title:  task.Title(),
		UserID: task.UserID(),
	}
}

type TaskDB struct {
	db *gorm.DB
}

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

	var taskModel TaskModel
	if err := t.db.WithContext(ctx).First(&taskModel, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return taskModel.ToDomain(), nil
}

func (t *TaskDB) FindAllByUserID(ctx context.Context, userID string) ([]*task.Task, error) {
	if userID == "" {
		panic("UserID must not be empty")
	}

	var taskModels []TaskModel
	if err := t.db.WithContext(ctx).Find(&taskModels, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	tasks := make([]*task.Task, len(taskModels))
	for i, taskModel := range taskModels {
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
		ID:     uuid.New().String(),
		Title:  title,
		UserID: userID,
	}

	if err := t.db.WithContext(ctx).Create(taskModel).Error; err != nil {
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

	result := t.db.WithContext(ctx).Delete(&TaskModel{}, "id = ? AND user_id = ?", id, userID)
	if result.Error != nil {
		return result.Error
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

	var taskModel TaskModel
	if err := t.db.WithContext(ctx).First(&taskModel, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}

	taskModel.Title = title
	if err := t.db.WithContext(ctx).Save(&taskModel).Error; err != nil {
		return nil, err
	}

	return taskModel.ToDomain(), nil
}
