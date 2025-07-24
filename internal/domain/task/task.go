package task

import (
	"errors"
	"strings"
)

const MaxTitleLength = 255

var (
	ErrTitleEmpty   = errors.New("task title cannot be empty")
	ErrTitleTooLong = errors.New("task title cannot exceed 255 characters")
)

type Task struct {
	id     string
	title  string
	userId string // Note: This represents the userId from JWT sub claim, which may change in the future
}

func NewTask(id, title, userId string) *Task {
	return &Task{
		id:     id,
		title:  title,
		userId: userId,
	}
}

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
