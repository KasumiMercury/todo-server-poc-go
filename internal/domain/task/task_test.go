package task

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskWithValidation(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testID := uuid.New().String()
	testUserID := uuid.New().String()

	tests := []struct {
		name          string
		id            string
		title         string
		userID        string
		expectedError error
	}{
		{
			name:          "valid task",
			id:            testID,
			title:         "Valid Task Title",
			userID:        testUserID,
			expectedError: nil,
		},
		{
			name:          "empty title",
			id:            testID,
			title:         "",
			userID:        testUserID,
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "whitespace only title",
			id:            testID,
			title:         "   \n\t   ",
			userID:        testUserID,
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "title exactly max length",
			id:            testID,
			title:         strings.Repeat("a", MaxTitleLength),
			userID:        testUserID,
			expectedError: nil,
		},
		{
			name:          "title too long",
			id:            testID,
			title:         strings.Repeat("a", MaxTitleLength+1),
			userID:        testUserID,
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "title with leading/trailing spaces (valid after trim)",
			id:            testID,
			title:         "  Valid Title  ",
			userID:        testUserID,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			task, err := NewTaskWithValidation(tt.id, tt.title, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, tt.id, task.ID())
				assert.Equal(t, tt.title, task.Title())
				assert.Equal(t, tt.userID, task.UserID())
			}
		})
	}
}

func TestTaskUpdateTitle(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testID := uuid.New().String()
	testUserID := uuid.New().String()

	tests := []struct {
		name          string
		initialTitle  string
		newTitle      string
		expectedError error
		expectedTitle string
	}{
		{
			name:          "valid title update",
			initialTitle:  "Old Title",
			newTitle:      "New Title",
			expectedError: nil,
			expectedTitle: "New Title",
		},
		{
			name:          "empty new title",
			initialTitle:  "Old Title",
			newTitle:      "",
			expectedError: ErrTitleEmpty,
			expectedTitle: "Old Title", // should remain unchanged
		},
		{
			name:          "whitespace only new title",
			initialTitle:  "Old Title",
			newTitle:      "   \t\n   ",
			expectedError: ErrTitleEmpty,
			expectedTitle: "Old Title", // should remain unchanged
		},
		{
			name:          "new title too long",
			initialTitle:  "Old Title",
			newTitle:      strings.Repeat("b", MaxTitleLength+1),
			expectedError: ErrTitleTooLong,
			expectedTitle: "Old Title", // should remain unchanged
		},
		{
			name:          "new title exactly max length",
			initialTitle:  "Old Title",
			newTitle:      strings.Repeat("c", MaxTitleLength),
			expectedError: nil,
			expectedTitle: strings.Repeat("c", MaxTitleLength),
		},
		{
			name:          "new title with leading/trailing spaces",
			initialTitle:  "Old Title",
			newTitle:      "  Trimmed Title  ",
			expectedError: nil,
			expectedTitle: "  Trimmed Title  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			task := NewTask(testID, tt.initialTitle, testUserID)

			// Act
			err := task.UpdateTitle(tt.newTitle)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedTitle, task.Title())
		})
	}
}

func TestValidateTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		title         string
		expectedError error
	}{
		{
			name:          "valid title",
			title:         "Valid Title",
			expectedError: nil,
		},
		{
			name:          "empty title",
			title:         "",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "whitespace only title",
			title:         "   \t\n   ",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "title exactly max length",
			title:         strings.Repeat("x", MaxTitleLength),
			expectedError: nil,
		},
		{
			name:          "title too long",
			title:         strings.Repeat("x", MaxTitleLength+1),
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "title with special characters",
			title:         "タスク with émojis 🚀 and symbols !@#$%",
			expectedError: nil,
		},
		{
			name:          "single character title",
			title:         "x",
			expectedError: nil,
		},
		{
			name:          "title with leading/trailing spaces",
			title:         "  Valid Title  ",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			err := validateTitle(tt.title)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskGetters(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedID := uuid.New().String()
	expectedTitle := "Test Task Title"
	expectedUserID := uuid.New().String()

	task := NewTask(expectedID, expectedTitle, expectedUserID)

	// Act & Assert
	assert.Equal(t, expectedID, task.ID())
	assert.Equal(t, expectedTitle, task.Title())
	assert.Equal(t, expectedUserID, task.UserID())
}

func TestMaxTitleLength(t *testing.T) {
	t.Parallel()

	// Assert that the constant is properly defined
	assert.Equal(t, 255, MaxTitleLength)
}
