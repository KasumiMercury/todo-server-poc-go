package handler

import (
	taskDomain "github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/oapi-codegen/runtime/types"
)

// UUIDAdapter handles conversion between OpenAPI UUID types and domain TaskID
type UUIDAdapter struct{}

// NewUUIDAdapter creates a new UUIDAdapter instance
func NewUUIDAdapter() *UUIDAdapter {
	return &UUIDAdapter{}
}

// ToDomainTaskID converts openapi_types.UUID to domain TaskID
func (a *UUIDAdapter) ToDomainTaskID(apiUUID types.UUID) (taskDomain.TaskID, error) {
	return taskDomain.NewTaskID(apiUUID.String())
}
