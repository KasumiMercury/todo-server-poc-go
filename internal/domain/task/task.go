package task

import (
	"strings"
	"unicode"
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

// NewTask creates a new Task instance with title validation.
// It returns an error if the title is invalid according to business rules.
func NewTask(id TaskID, title string, userID user.UserID) (*Task, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	return &Task{
		id:     id,
		title:  title,
		userID: userID,
	}, nil
}

// NewTaskWithoutValidation creates a new Task instance with the provided parameters.
// It does not perform validation on the input parameters.
func NewTaskWithoutValidation(id TaskID, title string, userID user.UserID) *Task {
	return &Task{
		id:     id,
		title:  title,
		userID: userID,
	}
}

// trimSpaceAndZeroWidth removes leading/trailing spaces and zero-width characters
func trimSpaceAndZeroWidth(s string) string {
	// First trim standard spaces
	s = strings.TrimSpace(s)

	// Remove zero-width characters from both ends
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) ||
			r == '\u200B' || // Zero-width space
			r == '\u200C' || // Zero-width non-joiner
			r == '\u200D' || // Zero-width joiner
			r == '\u200E' || // Left-to-right mark
			r == '\u200F' || // Right-to-left mark
			r == '\uFEFF' // Zero-width no-break space (BOM)
	})
}

func validateTitle(title string) error {
	if trimSpaceAndZeroWidth(title) == "" {
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
