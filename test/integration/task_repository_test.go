package integration

import (
	"context"
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&repository.TaskModel{})
	require.NoError(t, err)

	return db, func() {
		postgresContainer.Terminate(ctx)
	}
}

func TestTaskDB_Integration_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := repository.NewTaskDB(db)
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        string
		title         string
		expectedError error
	}{
		{
			name:          "successful creation",
			userID:        uuid.New().String(),
			title:         "Integration Test Task",
			expectedError: nil,
		},
		{
			name:          "create multiple tasks for same user",
			userID:        uuid.New().String(),
			title:         "Second Task",
			expectedError: nil,
		},
		{
			name:          "create task for different user",
			userID:        uuid.New().String(),
			title:         "Other User Task",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			createdTask, err := taskRepo.Create(ctx, tt.userID, tt.title)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, createdTask)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, createdTask)
				assert.NotEmpty(t, createdTask.ID())
				assert.Equal(t, tt.title, createdTask.Title())
				assert.Equal(t, tt.userID, createdTask.UserID())
			}
		})
	}
}

func TestTaskDB_Integration_FindById(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := repository.NewTaskDB(db)
	ctx := context.Background()

	// Create test task
	testUserID := uuid.New().String()
	createdTask, err := taskRepo.Create(ctx, testUserID, "Test Task")
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	tests := []struct {
		name          string
		userID        string
		taskID        string
		expectedTask  *task.Task
		expectedError error
	}{
		{
			name:          "find existing task",
			userID:        testUserID,
			taskID:        createdTask.ID(),
			expectedTask:  createdTask,
			expectedError: nil,
		},
		{
			name:          "task not found for different user",
			userID:        uuid.New().String(),
			taskID:        createdTask.ID(),
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
		},
		{
			name:          "task not found with nonexistent ID",
			userID:        testUserID,
			taskID:        "nonexistent-id",
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			foundTask, err := taskRepo.FindById(ctx, tt.userID, tt.taskID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, foundTask)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, foundTask)
				assert.Equal(t, tt.expectedTask.ID(), foundTask.ID())
				assert.Equal(t, tt.expectedTask.Title(), foundTask.Title())
				assert.Equal(t, tt.expectedTask.UserID(), foundTask.UserID())
			}
		})
	}
}

func TestTaskDB_Integration_FindAllByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := repository.NewTaskDB(db)
	ctx := context.Background()

	// Create test users
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()

	// Create test tasks for user1
	_, err := taskRepo.Create(ctx, user1ID, "Task 1")
	require.NoError(t, err)
	_, err = taskRepo.Create(ctx, user1ID, "Task 2")
	require.NoError(t, err)

	// Create task for different user
	_, err = taskRepo.Create(ctx, user2ID, "Other User Task")
	require.NoError(t, err)

	tests := []struct {
		name           string
		userID         string
		expectedCount  int
		expectedTitles []string
		expectedError  error
	}{
		{
			name:           "find tasks for user with multiple tasks",
			userID:         user1ID,
			expectedCount:  2,
			expectedTitles: []string{"Task 1", "Task 2"},
			expectedError:  nil,
		},
		{
			name:           "find tasks for user with single task",
			userID:         user2ID,
			expectedCount:  1,
			expectedTitles: []string{"Other User Task"},
			expectedError:  nil,
		},
		{
			name:           "find tasks for user with no tasks",
			userID:         uuid.New().String(),
			expectedCount:  0,
			expectedTitles: nil,
			expectedError:  task.ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			foundTasks, err := taskRepo.FindAllByUserID(ctx, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, foundTasks)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, foundTasks)
				assert.Len(t, foundTasks, tt.expectedCount)

				// Check that all tasks belong to the correct user
				for _, foundTask := range foundTasks {
					assert.Equal(t, tt.userID, foundTask.UserID())
				}

				// Check titles (order may vary)
				foundTitles := make([]string, len(foundTasks))
				for i, foundTask := range foundTasks {
					foundTitles[i] = foundTask.Title()
				}

				assert.ElementsMatch(t, tt.expectedTitles, foundTitles)
			}
		})
	}
}

func TestTaskDB_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := repository.NewTaskDB(db)
	ctx := context.Background()

	// Create test task
	testUserID := uuid.New().String()
	createdTask, err := taskRepo.Create(ctx, testUserID, "Original Title")
	require.NoError(t, err)

	tests := []struct {
		name          string
		userID        string
		taskID        string
		newTitle      string
		expectedError error
	}{
		{
			name:          "successful update",
			userID:        testUserID,
			taskID:        createdTask.ID(),
			newTitle:      "Updated Title",
			expectedError: nil,
		},
		{
			name:          "update nonexistent task",
			userID:        uuid.New().String(),
			taskID:        "nonexistent-id",
			newTitle:      "Updated Title",
			expectedError: nil, // GORM Updates doesn't return error for no rows affected
		},
		{
			name:          "update task for different user",
			userID:        uuid.New().String(),
			taskID:        createdTask.ID(),
			newTitle:      "Updated Title",
			expectedError: nil, // GORM Updates doesn't return error for no rows affected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			updatedTask, err := taskRepo.Update(ctx, tt.userID, tt.taskID, tt.newTitle)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, updatedTask)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, updatedTask)
				assert.Equal(t, tt.taskID, updatedTask.ID())
				assert.Equal(t, tt.newTitle, updatedTask.Title())
				assert.Equal(t, tt.userID, updatedTask.UserID())

				// Verify the update persisted
				if tt.userID == testUserID && tt.taskID == createdTask.ID() {
					foundTask, err := taskRepo.FindById(ctx, tt.userID, tt.taskID)
					assert.NoError(t, err)
					require.NotNil(t, foundTask)
					assert.Equal(t, tt.newTitle, foundTask.Title())
				}
			}
		})
	}
}

func TestTaskDB_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := repository.NewTaskDB(db)
	ctx := context.Background()

	// Create test tasks
	testUserID := uuid.New().String()
	task1, err := taskRepo.Create(ctx, testUserID, "Task to Delete")
	require.NoError(t, err)
	task2, err := taskRepo.Create(ctx, testUserID, "Task to Keep")
	require.NoError(t, err)

	tests := []struct {
		name          string
		userID        string
		taskID        string
		expectedError error
	}{
		{
			name:          "successful deletion",
			userID:        testUserID,
			taskID:        task1.ID(),
			expectedError: nil,
		},
		{
			name:          "delete nonexistent task",
			userID:        testUserID,
			taskID:        "nonexistent-id",
			expectedError: nil,
		},
		{
			name:          "delete task for different user",
			userID:        uuid.New().String(),
			taskID:        task2.ID(),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := taskRepo.Delete(ctx, tt.userID, tt.taskID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)

				// Verify deletion for the successful case
				if tt.userID == testUserID && tt.taskID == task1.ID() {
					// Task should not be found
					_, err := taskRepo.FindById(ctx, tt.userID, tt.taskID)
					assert.Error(t, err)
					assert.ErrorIs(t, err, task.ErrTaskNotFound)

					// Other task should still exist
					foundTask, err := taskRepo.FindById(ctx, testUserID, task2.ID())
					assert.NoError(t, err)
					assert.NotNil(t, foundTask)
					assert.Equal(t, "Task to Keep", foundTask.Title())
				}
			}
		})
	}
}

func TestTaskDB_Integration_UserIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskRepo := repository.NewTaskDB(db)
	ctx := context.Background()

	// Create tasks for different users
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()

	user1Task1, err := taskRepo.Create(ctx, user1ID, "User 1 Task 1")
	require.NoError(t, err)
	_, err = taskRepo.Create(ctx, user1ID, "User 1 Task 2")
	require.NoError(t, err)

	user2Task1, err := taskRepo.Create(ctx, user2ID, "User 2 Task 1")
	require.NoError(t, err)

	// Test that users can only access their own tasks
	t.Run("user 1 can find their tasks", func(t *testing.T) {
		tasks, err := taskRepo.FindAllByUserID(ctx, user1ID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		taskItem, err := taskRepo.FindById(ctx, user1ID, user1Task1.ID())
		assert.NoError(t, err)
		assert.Equal(t, "User 1 Task 1", taskItem.Title())
	})

	t.Run("user 2 can find their tasks", func(t *testing.T) {
		tasks, err := taskRepo.FindAllByUserID(ctx, user2ID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)

		taskItem, err := taskRepo.FindById(ctx, user2ID, user2Task1.ID())
		assert.NoError(t, err)
		assert.Equal(t, "User 2 Task 1", taskItem.Title())
	})

	t.Run("users cannot access other users' tasks", func(t *testing.T) {
		// User 1 trying to access User 2's task
		_, err := taskRepo.FindById(ctx, user1ID, user2Task1.ID())
		assert.Error(t, err)
		assert.ErrorIs(t, err, task.ErrTaskNotFound)

		// User 2 trying to access User 1's task
		_, err = taskRepo.FindById(ctx, user2ID, user1Task1.ID())
		assert.Error(t, err)
		assert.ErrorIs(t, err, task.ErrTaskNotFound)
	})

	t.Run("users cannot update other users' tasks", func(t *testing.T) {
		// This should silently fail (no rows affected) but not return an error
		updatedTask, err := taskRepo.Update(ctx, user1ID, user2Task1.ID(), "Hacked Title")
		assert.NoError(t, err)
		assert.NotNil(t, updatedTask) // Returns the task with provided values, not from DB

		// Verify the original task was not actually updated
		originalTask, err := taskRepo.FindById(ctx, user2ID, user2Task1.ID())
		assert.NoError(t, err)
		assert.Equal(t, "User 2 Task 1", originalTask.Title()) // Still original title
	})

	t.Run("users cannot delete other users' tasks", func(t *testing.T) {
		// This should silently fail (no rows affected) but not return an error
		err := taskRepo.Delete(ctx, user1ID, user2Task1.ID())
		assert.NoError(t, err)

		// Verify the taskItem still exists for the correct user
		taskItem, err := taskRepo.FindById(ctx, user2ID, user2Task1.ID())
		assert.NoError(t, err)
		assert.Equal(t, "User 2 Task 1", taskItem.Title())
	})
}
