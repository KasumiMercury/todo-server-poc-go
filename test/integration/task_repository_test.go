package integration

import (
	"context"
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
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

	userID1, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)
	userID2, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)
	userID3, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	tests := []struct {
		name          string
		userID        user.UserID
		title         string
		expectedError error
	}{
		{
			name:          "successful creation",
			userID:        userID1,
			title:         "Integration Test Task",
			expectedError: nil,
		},
		{
			name:          "create multiple tasks for same user",
			userID:        userID2,
			title:         "Second Task",
			expectedError: nil,
		},
		{
			name:          "create task for different user",
			userID:        userID3,
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
				assert.False(t, createdTask.ID().IsEmpty())
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
	testUserID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)
	createdTask, err := taskRepo.Create(ctx, testUserID, "Test Task")
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	diffUserID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	nonExistentTaskID, err := task.NewTaskID(uuid.New().String())
	require.NoError(t, err)

	tests := []struct {
		name          string
		userID        user.UserID
		taskID        task.TaskID
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
			userID:        diffUserID,
			taskID:        createdTask.ID(),
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
		},
		{
			name:          "task not found - nonexistent ID",
			userID:        testUserID,
			taskID:        nonExistentTaskID,
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

	// Create test data
	user1ID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)
	user2ID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	task1, err := taskRepo.Create(ctx, user1ID, "User 1 Task 1")
	require.NoError(t, err)
	task2, err := taskRepo.Create(ctx, user1ID, "User 1 Task 2")
	require.NoError(t, err)

	task3, err := taskRepo.Create(ctx, user2ID, "User 2 Task 1")
	require.NoError(t, err)

	noTasksUserID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	tests := []struct {
		name          string
		userID        user.UserID
		expectedTasks int
		expectedError error
	}{
		{
			name:          "find tasks for user with multiple tasks",
			userID:        user1ID,
			expectedTasks: 2,
			expectedError: nil,
		},
		{
			name:          "find tasks for user with single task",
			userID:        user2ID,
			expectedTasks: 1,
			expectedError: nil,
		},
		{
			name:          "find tasks for user with no tasks",
			userID:        noTasksUserID,
			expectedTasks: 0,
			expectedError: task.ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			tasks, err := taskRepo.FindAllByUserID(ctx, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tt.expectedTasks)

				// Verify all tasks belong to the correct user
				for _, task := range tasks {
					assert.Equal(t, tt.userID, task.UserID())
				}
			}
		})
	}

	// Verify specific task content for user1
	t.Run("verify user1 task content", func(t *testing.T) {
		tasks, err := taskRepo.FindAllByUserID(ctx, user1ID)
		require.NoError(t, err)
		require.Len(t, tasks, 2)

		// Check that we have both tasks
		taskIDs := []task.TaskID{tasks[0].ID(), tasks[1].ID()}
		assert.Contains(t, taskIDs, task1.ID())
		assert.Contains(t, taskIDs, task2.ID())
	})

	// Verify specific task content for user2
	t.Run("verify user2 task content", func(t *testing.T) {
		tasks, err := taskRepo.FindAllByUserID(ctx, user2ID)
		require.NoError(t, err)
		require.Len(t, tasks, 1)

		assert.Equal(t, task3.ID(), tasks[0].ID())
		assert.Equal(t, "User 2 Task 1", tasks[0].Title())
	})
}
