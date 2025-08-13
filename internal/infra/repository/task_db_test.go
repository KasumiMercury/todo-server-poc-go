package repository

import (
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewTaskDB(t *testing.T) {
	t.Parallel()

	// Arrange
	mockDb := &gorm.DB{}

	// Act
	taskDB := NewTaskDB(mockDb)

	// Assert
	assert.NotNil(t, taskDB)
	assert.Equal(t, mockDb, taskDB.db)
}

func TestTaskModel_TableName(t *testing.T) {
	t.Parallel()

	// Arrange
	taskModel := TaskModel{}

	// Act
	tableName := taskModel.TableName()

	// Assert
	assert.Equal(t, "tasks", tableName)
}

func TestTaskModel_ToDomain(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for test data
	testID1 := uuid.New().String()
	testUserID1 := uuid.New().String()
	testID2 := uuid.New().String()
	testUserID2 := uuid.New().String()

	// Create domain types for expectations
	domainTaskID1, _ := task.NewTaskID(testID1)
	domainUserID1, _ := user.NewUserID(testUserID1)
	domainTaskID2, _ := task.NewTaskID(testID2)
	domainUserID2, _ := user.NewUserID(testUserID2)

	tests := []struct {
		name        string
		taskModel   TaskModel
		expected    *task.Task
		expectError bool
	}{
		{
			name: "convert task model to domain",
			taskModel: TaskModel{
				ID:        testID1,
				Title:     "Test Task",
				UserID:    testUserID1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected:    task.NewTaskWithoutValidation(domainTaskID1, "Test Task", domainUserID1),
			expectError: false,
		},
		{
			name: "convert with invalid task ID",
			taskModel: TaskModel{
				ID:        "invalid-uuid",
				Title:     "Test Task",
				UserID:    testUserID1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "convert with invalid user ID",
			taskModel: TaskModel{
				ID:        testID1,
				Title:     "Test Task",
				UserID:    "invalid-uuid",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "convert with special characters",
			taskModel: TaskModel{
				ID:        testID2,
				Title:     "タスク with émojis 🚀",
				UserID:    testUserID2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected:    task.NewTaskWithoutValidation(domainTaskID2, "タスク with émojis 🚀", domainUserID2),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			domainTask, err := tt.taskModel.ToDomain()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, domainTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, domainTask)
				assert.Equal(t, tt.expected.ID(), domainTask.ID())
				assert.Equal(t, tt.expected.Title(), domainTask.Title())
				assert.Equal(t, tt.expected.UserID(), domainTask.UserID())
			}
		})
	}
}
