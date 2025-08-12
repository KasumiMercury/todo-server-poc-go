package task

import "errors"

var (
	ErrTitleEmpty   = errors.New("task title cannot be empty")
	ErrTitleTooLong = errors.New("task title cannot exceed 255 characters")
	ErrTaskNotFound = errors.New("task not found")
	ErrUserIDEmpty  = errors.New("user ID cannot be empty")
	ErrTaskIDEmpty  = errors.New("task ID cannot be empty")
)
