package task

import (
	"strings"
)

// MaxTitleLength defines the maximum allowed length for task titles.
const MaxTitleLength = 255

// Task represents a task entity in the domain layer.
// It encapsulates task data and business logic for task management.
type Task struct {
	id     string
	title  string
	userId string // Note: This represents the userId from JWT sub claim, which may change in the future
}

// NewTask creates a new Task instance with the provided parameters.
// It does not perform validation on the input parameters.
func NewTask(id, title, userId string) *Task {
	return &Task{
		id:     id,
		title:  title,
		userId: userId,
	}
}

// NewTaskWithValidation creates a new Task instance with title validation.
// It returns an error if the title is invalid according to business rules.
func NewTaskWithValidation(id, title, userId string) (*Task, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	return &Task{
		id:     id,
		title:  title,
		userId: userId,
	}, nil
}

func validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleEmpty
	}

	if len(title) > MaxTitleLength {
		return ErrTitleTooLong
	}

	return nil
}

func (t *Task) ID() string {
	return t.id
}

func (t *Task) Title() string {
	return t.title
}

func (t *Task) UserID() string {
	return t.userId
}

func (t *Task) UpdateTitle(title string) error {
	if err := validateTitle(title); err != nil {
		return err
	}

	t.title = title

	return nil
}
