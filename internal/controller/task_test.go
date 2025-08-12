package controller

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTaskRepository implements task.TaskRepository for testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) FindAllByUserID(ctx context.Context, userID user.UserID) ([]*task.Task, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*task.Task), args.Error(1)
}

func (m *MockTaskRepository) FindById(ctx context.Context, userID user.UserID, id task.TaskID) (*task.Task, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*task.Task), args.Error(1)
}

func (m *MockTaskRepository) Create(ctx context.Context, userID user.UserID, title string) (*task.Task, error) {
	args := m.Called(ctx, userID, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*task.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, userID user.UserID, id task.TaskID, title string) (*task.Task, error) {
	args := m.Called(ctx, userID, id, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*task.Task), args.Error(1)
}

func (m *MockTaskRepository) Delete(ctx context.Context, userID user.UserID, id task.TaskID) error {
	args := m.Called(ctx, userID, id)

	return args.Error(0)
}

func TestNewTask(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := &MockTaskRepository{}

	// Act
	controller := NewTask(mockRepo)

	// Assert
	assert.NotNil(t, controller)
	assert.Equal(t, mockRepo, controller.taskRepo)
}

func TestTaskController_GetTaskById(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := user.GenerateUserID()
	testTaskID := task.GenerateTaskID()
	otherUserID := user.GenerateUserID()
	otherTaskID := task.GenerateTaskID()

	tests := []struct {
		name          string
		userID        user.UserID
		taskID        task.TaskID
		mockReturn    *task.Task
		mockError     error
		expectedTask  *task.Task
		expectedError error
	}{
		{
			name:          "successful retrieval",
			userID:        testUserID,
			taskID:        testTaskID,
			mockReturn:    task.NewTask(testTaskID, "Test Task", testUserID),
			mockError:     nil,
			expectedTask:  task.NewTask(testTaskID, "Test Task", testUserID),
			expectedError: nil,
		},
		{
			name:          "task not found",
			userID:        otherUserID,
			taskID:        otherTaskID,
			mockReturn:    nil,
			mockError:     task.ErrTaskNotFound,
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
		},
		{
			name:          "repository error",
			userID:        otherUserID,
			taskID:        otherTaskID,
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			mockRepo.On("FindById", ctx, tt.userID, tt.taskID).Return(tt.mockReturn, tt.mockError)

			// Act & Assert
			result, err := controller.GetTaskById(ctx, tt.userID, tt.taskID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTask.ID(), result.ID())
				assert.Equal(t, tt.expectedTask.Title(), result.Title())
				assert.Equal(t, tt.expectedTask.UserID(), result.UserID())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskController_GetAllTasks(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := user.GenerateUserID()
	task1ID := task.GenerateTaskID()
	task2ID := task.GenerateTaskID()

	tests := []struct {
		name          string
		userID        user.UserID
		mockReturn    []*task.Task
		mockError     error
		expectedTasks []*task.Task
		expectedError error
	}{
		{
			name:   "successful retrieval with tasks",
			userID: testUserID,
			mockReturn: []*task.Task{
				task.NewTask(task1ID, "Task 1", testUserID),
				task.NewTask(task2ID, "Task 2", testUserID),
			},
			mockError: nil,
			expectedTasks: []*task.Task{
				task.NewTask(task1ID, "Task 1", testUserID),
				task.NewTask(task2ID, "Task 2", testUserID),
			},
			expectedError: nil,
		},
		{
			name:          "no tasks found - returns empty slice",
			userID:        testUserID,
			mockReturn:    nil,
			mockError:     task.ErrTaskNotFound,
			expectedTasks: []*task.Task{},
			expectedError: nil,
		},
		{
			name:          "repository error",
			userID:        testUserID,
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTasks: nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			mockRepo.On("FindAllByUserID", ctx, tt.userID).Return(tt.mockReturn, tt.mockError)

			// Act
			result, err := controller.GetAllTasks(ctx, tt.userID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedTasks), len(result))

				for i, expectedTask := range tt.expectedTasks {
					assert.Equal(t, expectedTask.ID(), result[i].ID())
					assert.Equal(t, expectedTask.Title(), result[i].Title())
					assert.Equal(t, expectedTask.UserID(), result[i].UserID())
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskController_CreateTask(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := user.GenerateUserID()
	newTaskID := task.GenerateTaskID()

	tests := []struct {
		name          string
		userID        user.UserID
		title         string
		mockReturn    *task.Task
		mockError     error
		expectedTask  *task.Task
		expectedError error
	}{
		{
			name:          "successful creation",
			userID:        testUserID,
			title:         "New Task",
			mockReturn:    task.NewTask(newTaskID, "New Task", testUserID),
			mockError:     nil,
			expectedTask:  task.NewTask(newTaskID, "New Task", testUserID),
			expectedError: nil,
		},
		{
			name:          "empty title should fail validation",
			userID:        testUserID,
			title:         "",
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleEmpty,
		},
		{
			name:          "title too long should fail validation",
			userID:        testUserID,
			title:         strings.Repeat("a", task.MaxTitleLength+1),
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleTooLong,
		},
		{
			name:          "repository error",
			userID:        testUserID,
			title:         "Valid Title",
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			// Only set up mock expectations if we expect the repository to be called
			if tt.expectedError != task.ErrTitleEmpty && tt.expectedError != task.ErrTitleTooLong {
				mockRepo.On("Create", ctx, tt.userID, tt.title).Return(tt.mockReturn, tt.mockError)
			}

			// Act
			result, err := controller.CreateTask(ctx, tt.userID, tt.title)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTask.ID(), result.ID())
				assert.Equal(t, tt.expectedTask.Title(), result.Title())
				assert.Equal(t, tt.expectedTask.UserID(), result.UserID())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskController_DeleteTask(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := user.GenerateUserID()
	testTaskID := task.GenerateTaskID()

	tests := []struct {
		name          string
		userID        user.UserID
		taskID        task.TaskID
		mockError     error
		expectedError error
	}{
		{
			name:          "successful deletion",
			userID:        testUserID,
			taskID:        testTaskID,
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "task not found",
			userID:        testUserID,
			taskID:        testTaskID,
			mockError:     task.ErrTaskNotFound,
			expectedError: task.ErrTaskNotFound,
		},
		{
			name:          "repository error",
			userID:        testUserID,
			taskID:        testTaskID,
			mockError:     errors.New("database error"),
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			mockRepo.On("Delete", ctx, tt.userID, tt.taskID).Return(tt.mockError)

			// Act
			err := controller.DeleteTask(ctx, tt.userID, tt.taskID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskController_UpdateTask(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := user.GenerateUserID()
	testTaskID := task.GenerateTaskID()

	tests := []struct {
		name          string
		userID        user.UserID
		taskID        task.TaskID
		title         string
		mockReturn    *task.Task
		mockError     error
		expectedTask  *task.Task
		expectedError error
	}{
		{
			name:          "successful update",
			userID:        testUserID,
			taskID:        testTaskID,
			title:         "Updated Task",
			mockReturn:    task.NewTask(testTaskID, "Updated Task", testUserID),
			mockError:     nil,
			expectedTask:  task.NewTask(testTaskID, "Updated Task", testUserID),
			expectedError: nil,
		},
		{
			name:          "empty title should fail validation",
			userID:        testUserID,
			taskID:        testTaskID,
			title:         "",
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleEmpty,
		},
		{
			name:          "title too long should fail validation",
			userID:        testUserID,
			taskID:        testTaskID,
			title:         strings.Repeat("a", task.MaxTitleLength+1),
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleTooLong,
		},
		{
			name:          "task not found",
			userID:        testUserID,
			taskID:        testTaskID,
			title:         "Valid Title",
			mockReturn:    nil,
			mockError:     task.ErrTaskNotFound,
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
		},
		{
			name:          "repository error",
			userID:        testUserID,
			taskID:        testTaskID,
			title:         "Valid Title",
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			// Only set up mock expectations if we expect the repository to be called
			if tt.expectedError != task.ErrTitleEmpty && tt.expectedError != task.ErrTitleTooLong {
				mockRepo.On("Update", ctx, tt.userID, tt.taskID, tt.title).Return(tt.mockReturn, tt.mockError)
			}

			// Act
			result, err := controller.UpdateTask(ctx, tt.userID, tt.taskID, tt.title)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTask.ID(), result.ID())
				assert.Equal(t, tt.expectedTask.Title(), result.Title())
				assert.Equal(t, tt.expectedTask.UserID(), result.UserID())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
