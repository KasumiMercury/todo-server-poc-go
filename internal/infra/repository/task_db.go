package repository

import (
	"context"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	Title     string    `gorm:"not null;type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

func (t *TaskModel) ToDomain() *domain.Task {
	return domain.NewTask(t.ID, t.Title)
}

func NewTaskModelFromDomain(task *domain.Task) *TaskModel {
	return &TaskModel{
		ID:    task.ID(),
		Title: task.Title(),
	}
}

type TaskDB struct {
	db *gorm.DB
}

func NewTaskDB(db *gorm.DB) *TaskDB {
	return &TaskDB{db: db}
}

func (t *TaskDB) FindById(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}

	var taskModel TaskModel
	if err := t.db.WithContext(ctx).First(&taskModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return taskModel.ToDomain(), nil
}

func (t *TaskDB) FindAll(ctx context.Context) ([]*domain.Task, error) {
	var taskModels []TaskModel
	if err := t.db.WithContext(ctx).Find(&taskModels).Error; err != nil {
		return nil, err
	}

	tasks := make([]*domain.Task, len(taskModels))
	for i, taskModel := range taskModels {
		tasks[i] = taskModel.ToDomain()
	}

	return tasks, nil
}

func (t *TaskDB) Create(ctx context.Context, title string) (*domain.Task, error) {
	if title == "" {
		panic("Title must not be empty")
	}

	taskModel := &TaskModel{
		ID:    uuid.New().String(),
		Title: title,
	}

	if err := t.db.WithContext(ctx).Create(taskModel).Error; err != nil {
		return nil, err
	}

	return taskModel.ToDomain(), nil
}

func (t *TaskDB) Delete(ctx context.Context, id string) error {
	if id == "" {
		panic("ID must not be empty")
	}

	result := t.db.WithContext(ctx).Delete(&TaskModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (t *TaskDB) Update(ctx context.Context, id, title string) (*domain.Task, error) {
	if id == "" {
		panic("ID must not be empty")
	}
	if title == "" {
		panic("Title must not be empty")
	}

	var taskModel TaskModel
	if err := t.db.WithContext(ctx).First(&taskModel, "id = ?", id).Error; err != nil {
		return nil, err
	}

	taskModel.Title = title
	if err := t.db.WithContext(ctx).Save(&taskModel).Error; err != nil {
		return nil, err
	}

	return taskModel.ToDomain(), nil
}