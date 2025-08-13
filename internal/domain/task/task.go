package task

import (
	"strings"
	"unicode/utf8"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
	"github.com/google/uuid"
)

// MaxTitleLength defines the maximum allowed length for task titles.
const MaxTitleLength = 255

// TaskID represents a unique identifier for a task.
type TaskID struct {
	value uuid.UUID
}

// NewTaskID creates a new TaskID from a string value.
func NewTaskID(id string) (TaskID, error) {
	if id == "" {
		return TaskID{}, ErrTaskIDEmpty
	}

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return TaskID{}, ErrInvalidTaskIDFormat
	}

	return TaskID{value: parsedUUID}, nil
}

// GenerateTaskID creates a new TaskID with a generated UUID.
func GenerateTaskID() TaskID {
	return TaskID{value: uuid.New()}
}

// String returns the string representation of the TaskID.
func (t TaskID) String() string {
	return t.value.String()
}

// IsEmpty returns true if the TaskID is empty.
func (t TaskID) IsEmpty() bool {
	return t.value == uuid.Nil
}

// Task represents a task entity in the domain layer.
// It encapsulates task data and business logic for task management.
type Task struct {
	id     TaskID
	title  string
	userID user.UserID
}

// NewTask creates a new Task instance with the provided parameters.
// It does not perform validation on the input parameters.
func NewTask(id TaskID, title string, userID user.UserID) *Task {
	return &Task{
		id:     id,
		title:  title,
		userID: userID,
	}
}

// NewTaskWithValidation creates a new Task instance with title validation.
// It returns an error if the title is invalid according to business rules.
func NewTaskWithValidation(id TaskID, title string, userID user.UserID) (*Task, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	return &Task{
		id:     id,
		title:  title,
		userID: userID,
	}, nil
}

func validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleEmpty
	}

	if utf8.RuneCountInString(title) > MaxTitleLength {
		return ErrTitleTooLong
	}

	return nil
}

func (t *Task) ID() TaskID {
	return t.id
}

func (t *Task) Title() string {
	return t.title
}

func (t *Task) UserID() user.UserID {
	return t.userID
}

func (t *Task) UpdateTitle(title string) error {
	if err := validateTitle(title); err != nil {
		return err
	}

	t.title = title

	return nil
}
