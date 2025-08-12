package task

import "errors"

var (
	ErrTitleEmpty          = errors.New("task title cannot be empty")
	ErrTitleTooLong        = errors.New("task title cannot exceed 255 characters")
	ErrTaskNotFound        = errors.New("task not found")
	ErrTaskIDEmpty         = errors.New("task ID cannot be empty")
	ErrInvalidTaskIDFormat = errors.New("task ID must be a valid UUID format")
)
