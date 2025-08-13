package integration

import (
	"context"
	"strings"
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
			var (
				createdTask *task.Task
				err         error
			)

			// Handle validation errors based on test case

			if tt.expectedError == task.ErrTitleEmpty || tt.expectedError == task.ErrTitleTooLong {
				// For title validation errors, the error comes from domain validation
				_, err = task.NewTask(task.GenerateTaskID(), tt.title, tt.userID)
				createdTask = nil
			} else {
				// For valid cases or other errors, create domain model and call repository
				taskEntity, validationErr := task.NewTask(task.GenerateTaskID(), tt.title, tt.userID)
				if validationErr != nil {
					err = validationErr
					createdTask = nil
				} else {
					createdTask, err = taskRepo.Create(ctx, taskEntity)
				}
			}

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
	taskEntity, err := task.NewTask(task.GenerateTaskID(), "Test Task", testUserID)
	require.NoError(t, err)
	createdTask, err := taskRepo.Create(ctx, taskEntity)
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

	taskEntity1, err := task.NewTask(task.GenerateTaskID(), "User 1 Task 1", user1ID)
	require.NoError(t, err)
	task1, err := taskRepo.Create(ctx, taskEntity1)
	require.NoError(t, err)
	taskEntity2, err := task.NewTask(task.GenerateTaskID(), "User 1 Task 2", user1ID)
	require.NoError(t, err)
	task2, err := taskRepo.Create(ctx, taskEntity2)
	require.NoError(t, err)

	taskEntity3, err := task.NewTask(task.GenerateTaskID(), "User 2 Task 1", user2ID)
	require.NoError(t, err)
	task3, err := taskRepo.Create(ctx, taskEntity3)
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
				for _, item := range tasks {
					assert.Equal(t, tt.userID, item.UserID())
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

func TestTaskDB_Integration_MultiByteTitles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	taskRepo := repository.NewTaskDB(db)

	userID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "chinese greeting exactly 255 characters",
			title: strings.Repeat("你", 255),
		},
		{
			name:  "korean greeting exactly 255 characters",
			title: strings.Repeat("안", 255),
		},
		{
			name:  "arabic greeting exactly 255 characters",
			title: strings.Repeat("م", 255),
		},
		{
			name:  "mixed multibyte characters",
			title: "你好世界 안녕하세요 مرحبا بالعالم नमस्कार संसार สวัสดีชาวโลก שלום עולם Привет мир",
		},
		{
			name:  "japanese with various scripts",
			title: "日本語ひらがなカタカナ漢字テスト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create task entity with multibyte title
			taskEntity, err := task.NewTask(task.GenerateTaskID(), tt.title, userID)
			require.NoError(t, err)

			// Act - Create task in database
			createdTask, err := taskRepo.Create(ctx, taskEntity)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.title, createdTask.Title())
			assert.Equal(t, userID, createdTask.UserID())

			// Act - Retrieve task from database
			retrievedTask, err := taskRepo.FindById(ctx, userID, createdTask.ID())

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.title, retrievedTask.Title())
			assert.Equal(t, createdTask.ID(), retrievedTask.ID())
			assert.Equal(t, userID, retrievedTask.UserID())
		})
	}
}

func TestTaskDB_Integration_ZeroWidthJoinerTitles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	taskRepo := repository.NewTaskDB(db)

	userID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "profession emoji within boundary",
			title: strings.Repeat("👨‍💻", 85), // 85 * 3 = 255 runes
		},
		{
			name:  "family emoji within boundary",
			title: strings.Repeat("👨‍👩‍👧‍👦", 36), // 36 * 7 = 252 runes
		},
		{
			name:  "skin tone emoji within boundary",
			title: strings.Repeat("👋🏽", 127), // 127 * 2 = 254 runes
		},
		{
			name:  "mixed complex emojis",
			title: "👨‍💻👩‍🔬👨‍🍳🧑‍🦰👋🏽👨‍👩‍👧‍👦",
		},
		{
			name:  "gendered activity emojis",
			title: "🏃‍♂️🏃‍♀️🚴‍♂️🚴‍♀️🏊‍♂️🏊‍♀️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create task entity with ZWJ emoji title
			taskEntity, err := task.NewTask(task.GenerateTaskID(), tt.title, userID)
			require.NoError(t, err)

			// Act - Create task in database
			createdTask, err := taskRepo.Create(ctx, taskEntity)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.title, createdTask.Title())
			assert.Equal(t, userID, createdTask.UserID())

			// Act - Retrieve task from database
			retrievedTask, err := taskRepo.FindById(ctx, userID, createdTask.ID())

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.title, retrievedTask.Title())
			assert.Equal(t, createdTask.ID(), retrievedTask.ID())
			assert.Equal(t, userID, retrievedTask.UserID())
		})
	}
}

func TestTaskDB_Integration_BoundaryTitles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	taskRepo := repository.NewTaskDB(db)

	userID, err := user.NewUserID(uuid.New().String())
	require.NoError(t, err)

	tests := []struct {
		name          string
		title         string
		shouldSucceed bool
		expectedError error
	}{
		{
			name:          "title exactly 255 ASCII characters",
			title:         strings.Repeat("a", 255),
			shouldSucceed: true,
		},
		{
			name:          "title exactly 255 Chinese characters",
			title:         strings.Repeat("你", 255),
			shouldSucceed: true,
		},
		{
			name:          "title exactly 255 ZWJ emoji runes",
			title:         strings.Repeat("👨‍💻", 85), // 85 * 3 = 255 runes
			shouldSucceed: true,
		},
		{
			name:          "title 256 ASCII characters (should fail)",
			title:         strings.Repeat("a", 256),
			shouldSucceed: false,
			expectedError: task.ErrTitleTooLong,
		},
		{
			name:          "title 256 Chinese characters (should fail)",
			title:         strings.Repeat("你", 256),
			shouldSucceed: false,
			expectedError: task.ErrTitleTooLong,
		},
		{
			name:          "title 258 ZWJ emoji runes (should fail)",
			title:         strings.Repeat("👨‍💻", 86), // 86 * 3 = 258 runes
			shouldSucceed: false,
			expectedError: task.ErrTitleTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - Create task entity with boundary title
			taskEntity, err := task.NewTask(task.GenerateTaskID(), tt.title, userID)

			if tt.shouldSucceed {
				// Assert successful creation
				require.NoError(t, err)

				// Act - Create task in database
				createdTask, err := taskRepo.Create(ctx, taskEntity)

				// Assert
				require.NoError(t, err)
				assert.Equal(t, tt.title, createdTask.Title())

				// Act - Retrieve task from database to ensure roundtrip works
				retrievedTask, err := taskRepo.FindById(ctx, userID, createdTask.ID())

				// Assert
				require.NoError(t, err)
				assert.Equal(t, tt.title, retrievedTask.Title())
			} else {
				// Assert validation failure
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, taskEntity)
			}
		})
	}
}
