package task

import (
	"context"
)

type TaskRepository interface {
	FindById(ctx context.Context, id string) (*Task, error)
	FindAll(ctx context.Context) ([]*Task, error)
	Create(ctx context.Context, name string) (*Task, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id, name string) (*Task, error)
}

//type TaskMemory struct{}
//
//func NewTaskMemory() *TaskMemory {
//	return &TaskMemory{}
//}
//
//func (t *TaskMemory) FindById(ctx context.Context, id string) (*domain.Task, error) {
//	if id == "" {
//		panic("ID must not be empty")
//	}
//
//	return domain.NewTask(
//		id,
//		"Sample Task",
//	), nil
//}
//
//func (t *TaskMemory) FindAll(ctx context.Context) ([]*domain.Task, error) {
//	return []*domain.Task{
//		domain.NewTask("1", "Task 1"),
//		domain.NewTask("2", "Task 2"),
//		domain.NewTask("3", "Task 3"),
//	}, nil
//}
//
//func (t *TaskMemory) Create(ctx context.Context, title string) (*domain.Task, error) {
//	if title == "" {
//		panic("Title must not be empty")
//	}
//
//	return domain.NewTask("new-task-id", title), nil
//}
//
//func (t *TaskMemory) Delete(ctx context.Context, id string) error {
//	if id == "" {
//		panic("ID must not be empty")
//	}
//
//	return nil
//}
//
//func (t *TaskMemory) Update(ctx context.Context, id, title string) (*domain.Task, error) {
//	if id == "" {
//		panic("ID must not be empty")
//	}
//	if title == "" {
//		panic("Title must not be empty")
//	}
//
//	return domain.NewTask(id, title), nil
//}
