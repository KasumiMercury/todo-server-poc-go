package user

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// UserID represents a unique identifier for a user.
// It ensures type safety and domain-specific validation for user identifiers.
type UserID struct {
	value uuid.UUID
}

var (
	ErrUserIDEmpty         = errors.New("user ID cannot be empty")
	ErrInvalidUserIDFormat = errors.New("user ID format is invalid")
)

// NewUserID creates a new UserID from a string value.
func NewUserID(id string) (UserID, error) {
	if id == "" {
		return UserID{}, ErrUserIDEmpty
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return UserID{}, ErrUserIDEmpty
	}

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return UserID{}, ErrInvalidUserIDFormat
	}

	return UserID{value: parsedUUID}, nil
}

// GenerateUserID creates a new UserID with a generated UUID.
func GenerateUserID() UserID {
	return UserID{value: uuid.New()}
}

// String returns the string representation of the UserID.
func (u UserID) String() string {
	return u.value.String()
}

// IsEmpty returns true if the UserID is empty.
func (u UserID) IsEmpty() bool {
	return u.value == uuid.Nil
}
