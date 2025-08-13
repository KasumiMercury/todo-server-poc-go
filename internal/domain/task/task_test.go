package task

import (
	"strings"
	"testing"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskWithValidation(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testID := GenerateTaskID()
	testUserID := user.GenerateUserID()

	tests := []struct {
		name          string
		id            TaskID
		title         string
		userID        user.UserID
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
			task, err := NewTask(tt.id, tt.title, tt.userID)

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
	testID := GenerateTaskID()
	testUserID := user.GenerateUserID()

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
			task := NewTaskWithoutValidation(testID, tt.initialTitle, testUserID)

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
		{
			name:          "chinese characters exactly max length",
			title:         strings.Repeat("你", MaxTitleLength),
			expectedError: nil,
		},
		{
			name:          "chinese characters over max length",
			title:         strings.Repeat("你", MaxTitleLength+1),
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "korean characters exactly max length",
			title:         strings.Repeat("안", MaxTitleLength),
			expectedError: nil,
		},
		{
			name:          "korean characters over max length",
			title:         strings.Repeat("안", MaxTitleLength+1),
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "arabic characters exactly max length",
			title:         strings.Repeat("م", MaxTitleLength),
			expectedError: nil,
		},
		{
			name:          "arabic characters over max length",
			title:         strings.Repeat("م", MaxTitleLength+1),
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "zero-width joiner emoji with boundary consideration",
			title:         strings.Repeat("👨‍💻", 85), // 85 * 3 runes = 255 runes
			expectedError: nil,
		},
		{
			name:          "zero-width joiner emoji over boundary",
			title:         strings.Repeat("👨‍💻", 86), // 86 * 3 runes = 258 runes > 255
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "family emoji with boundary consideration",
			title:         strings.Repeat("👨‍👩‍👧‍👦", 36), // 36 * 7 runes = 252 runes
			expectedError: nil,
		},
		{
			name:          "family emoji over boundary",
			title:         strings.Repeat("👨‍👩‍👧‍👦", 37), // 37 * 7 runes = 259 runes > 255
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "skin tone emoji with boundary consideration",
			title:         strings.Repeat("👋🏽", 127), // 127 * 2 runes = 254 runes
			expectedError: nil,
		},
		{
			name:          "skin tone emoji over boundary",
			title:         strings.Repeat("👋🏽", 128), // 128 * 2 runes = 256 runes > 255
			expectedError: ErrTitleTooLong,
		},
		{
			name:          "unicode non-breaking space only",
			title:         "\u00A0\u00A0\u00A0",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "unicode thin space only",
			title:         "\u2009\u2009\u2009",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "ideographic space only",
			title:         "　　　",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "zero-width space only",
			title:         "​​​",
			expectedError: ErrTitleEmpty, // Zero-width spaces are now trimmed
		},
		{
			name:          "mixed unicode whitespace only",
			title:         "\u00A0\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u3000",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "line separator character only",
			title:         "\u2028\u2028",
			expectedError: ErrTitleEmpty,
		},
		{
			name:          "paragraph separator character only",
			title:         "\u2029\u2029",
			expectedError: ErrTitleEmpty,
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
	expectedID := GenerateTaskID()
	expectedTitle := "Test Task Title"
	expectedUserID := user.GenerateUserID()

	task := NewTaskWithoutValidation(expectedID, expectedTitle, expectedUserID)

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
