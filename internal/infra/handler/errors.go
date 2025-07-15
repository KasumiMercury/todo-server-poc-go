package handler

import "github.com/KasumiMercury/todo-server-poc-go/internal/task"

// stringPtr creates a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

// NewError creates a new Error using the generated Error struct
func NewError(code int, message string, details *string) task.Error {
	return task.Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewBadRequestError creates a 400 Bad Request error
func NewBadRequestError(message string, details *string) task.Error {
	return NewError(400, message, details)
}

// NewUnauthorizedError creates a 401 Unauthorized error
func NewUnauthorizedError(message string, details *string) task.Error {
	return NewError(401, message, details)
}

// NewNotFoundError creates a 404 Not Found error
func NewNotFoundError(message string) task.Error {
	return NewError(404, message, nil)
}

// NewInternalServerError creates a 500 Internal Server Error
func NewInternalServerError(message string, details *string) task.Error {
	return NewError(500, message, details)
}
