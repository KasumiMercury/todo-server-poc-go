package controller

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTaskRepository implements task.TaskRepository for testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) FindAllByUserID(ctx context.Context, userID string) ([]*task.Task, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*task.Task), args.Error(1)
}

func (m *MockTaskRepository) FindById(ctx context.Context, userID, id string) (*task.Task, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*task.Task), args.Error(1)
}

func (m *MockTaskRepository) Create(ctx context.Context, userID, title string) (*task.Task, error) {
	args := m.Called(ctx, userID, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*task.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, userID, id, title string) (*task.Task, error) {
	args := m.Called(ctx, userID, id, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*task.Task), args.Error(1)
}

func (m *MockTaskRepository) Delete(ctx context.Context, userID, id string) error {
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
	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()
	otherUserID := uuid.New().String()
	otherTaskID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		taskID        string
		mockReturn    *task.Task
		mockError     error
		expectedTask  *task.Task
		expectedError error
		shouldPanic   bool
	}{
		{
			name:          "successful retrieval",
			userID:        testUserID,
			taskID:        testTaskID,
			mockReturn:    task.NewTask(testTaskID, "Test Task", testUserID),
			mockError:     nil,
			expectedTask:  task.NewTask(testTaskID, "Test Task", testUserID),
			expectedError: nil,
			shouldPanic:   false,
		},
		{
			name:          "task not found",
			userID:        otherUserID,
			taskID:        "nonexistent",
			mockReturn:    nil,
			mockError:     task.ErrTaskNotFound,
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
			shouldPanic:   false,
		},
		{
			name:          "repository error",
			userID:        otherUserID,
			taskID:        otherTaskID,
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTask:  nil,
			expectedError: errors.New("database error"),
			shouldPanic:   false,
		},
		{
			name:        "empty userID should panic",
			userID:      "",
			taskID:      otherTaskID,
			shouldPanic: true,
		},
		{
			name:        "empty taskID should panic",
			userID:      testUserID,
			taskID:      "",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			if !tt.shouldPanic {
				mockRepo.On("FindById", ctx, tt.userID, tt.taskID).Return(tt.mockReturn, tt.mockError)
			}

			// Act & Assert
			if tt.shouldPanic {
				assert.Panics(t, func() {
					controller.GetTaskById(ctx, tt.userID, tt.taskID)
				})
			} else {
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
			}
		})
	}
}

func TestTaskController_GetAllTasks(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := uuid.New().String()
	task1ID := uuid.New().String()
	task2ID := uuid.New().String()
	otherUserID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		mockReturn    []*task.Task
		mockError     error
		expectedTasks []*task.Task
		expectedError error
		shouldPanic   bool
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
			shouldPanic:   false,
		},
		{
			name:          "no tasks found - returns empty slice",
			userID:        otherUserID,
			mockReturn:    nil,
			mockError:     task.ErrTaskNotFound,
			expectedTasks: []*task.Task{},
			expectedError: nil,
			shouldPanic:   false,
		},
		{
			name:          "repository error",
			userID:        otherUserID,
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTasks: nil,
			expectedError: errors.New("database error"),
			shouldPanic:   false,
		},
		{
			name:        "empty userID should panic",
			userID:      "",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			if !tt.shouldPanic {
				mockRepo.On("FindAllByUserID", ctx, tt.userID).Return(tt.mockReturn, tt.mockError)
			}

			// Act & Assert
			if tt.shouldPanic {
				assert.Panics(t, func() {
					controller.GetAllTasks(ctx, tt.userID)
				})
			} else {
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
			}
		})
	}
}

func TestTaskController_CreateTask(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := uuid.New().String()
	newTaskID := uuid.New().String()
	otherUserID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		title         string
		mockReturn    *task.Task
		mockError     error
		expectedTask  *task.Task
		expectedError error
		shouldPanic   bool
	}{
		{
			name:          "successful creation",
			userID:        testUserID,
			title:         "New Task",
			mockReturn:    task.NewTask(newTaskID, "New Task", testUserID),
			mockError:     nil,
			expectedTask:  task.NewTask(newTaskID, "New Task", testUserID),
			expectedError: nil,
			shouldPanic:   false,
		},
		{
			name:          "empty title should return validation error",
			userID:        otherUserID,
			title:         "",
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleEmpty,
			shouldPanic:   false,
		},
		{
			name:          "title too long should return validation error",
			userID:        otherUserID,
			title:         strings.Repeat("a", task.MaxTitleLength+1),
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleTooLong,
			shouldPanic:   false,
		},
		{
			name:          "repository error",
			userID:        otherUserID,
			title:         "Valid Task",
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTask:  nil,
			expectedError: errors.New("database error"),
			shouldPanic:   false,
		},
		{
			name:        "empty userID should panic",
			userID:      "",
			title:       "Valid Task",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			if !tt.shouldPanic && tt.expectedError != task.ErrTitleEmpty && tt.expectedError != task.ErrTitleTooLong {
				mockRepo.On("Create", ctx, tt.userID, tt.title).Return(tt.mockReturn, tt.mockError)
			}

			// Act & Assert
			if tt.shouldPanic {
				assert.Panics(t, func() {
					controller.CreateTask(ctx, tt.userID, tt.title)
				})
			} else {
				result, err := controller.CreateTask(ctx, tt.userID, tt.title)

				// Assert
				if tt.expectedError != nil {
					assert.Error(t, err)
					if errors.Is(tt.expectedError, task.ErrTitleEmpty) || errors.Is(tt.expectedError, task.ErrTitleTooLong) {
						assert.ErrorIs(t, err, tt.expectedError)
					} else {
						assert.Equal(t, tt.expectedError.Error(), err.Error())
					}
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					require.NotNil(t, result)
					assert.Equal(t, tt.expectedTask.ID(), result.ID())
					assert.Equal(t, tt.expectedTask.Title(), result.Title())
					assert.Equal(t, tt.expectedTask.UserID(), result.UserID())
				}

				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestTaskController_UpdateTask(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()
	otherUserID := uuid.New().String()
	otherTaskID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		taskID        string
		title         string
		mockReturn    *task.Task
		mockError     error
		expectedTask  *task.Task
		expectedError error
		shouldPanic   bool
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
			shouldPanic:   false,
		},
		{
			name:          "empty title should return validation error",
			userID:        otherUserID,
			taskID:        otherTaskID,
			title:         "",
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleEmpty,
			shouldPanic:   false,
		},
		{
			name:          "title too long should return validation error",
			userID:        otherUserID,
			taskID:        otherTaskID,
			title:         strings.Repeat("b", task.MaxTitleLength+1),
			mockReturn:    nil,
			mockError:     nil,
			expectedTask:  nil,
			expectedError: task.ErrTitleTooLong,
			shouldPanic:   false,
		},
		{
			name:          "task not found",
			userID:        otherUserID,
			taskID:        "nonexistent",
			title:         "Updated Task",
			mockReturn:    nil,
			mockError:     task.ErrTaskNotFound,
			expectedTask:  nil,
			expectedError: task.ErrTaskNotFound,
			shouldPanic:   false,
		},
		{
			name:          "repository error",
			userID:        otherUserID,
			taskID:        otherTaskID,
			title:         "Updated Task",
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedTask:  nil,
			expectedError: errors.New("database error"),
			shouldPanic:   false,
		},
		{
			name:        "empty userID should panic",
			userID:      "",
			taskID:      otherTaskID,
			title:       "Updated Task",
			shouldPanic: true,
		},
		{
			name:        "empty taskID should panic",
			userID:      testUserID,
			taskID:      "",
			title:       "Updated Task",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			if !tt.shouldPanic && tt.expectedError != task.ErrTitleEmpty && tt.expectedError != task.ErrTitleTooLong {
				mockRepo.On("Update", ctx, tt.userID, tt.taskID, tt.title).Return(tt.mockReturn, tt.mockError)
			}

			// Act & Assert
			if tt.shouldPanic {
				assert.Panics(t, func() {
					controller.UpdateTask(ctx, tt.userID, tt.taskID, tt.title)
				})
			} else {
				result, err := controller.UpdateTask(ctx, tt.userID, tt.taskID, tt.title)

				// Assert
				if tt.expectedError != nil {
					assert.Error(t, err)
					if errors.Is(tt.expectedError, task.ErrTitleEmpty) || errors.Is(tt.expectedError, task.ErrTitleTooLong) {
						assert.ErrorIs(t, err, tt.expectedError)
					} else {
						assert.Equal(t, tt.expectedError.Error(), err.Error())
					}
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					require.NotNil(t, result)
					assert.Equal(t, tt.expectedTask.ID(), result.ID())
					assert.Equal(t, tt.expectedTask.Title(), result.Title())
					assert.Equal(t, tt.expectedTask.UserID(), result.UserID())
				}

				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestTaskController_DeleteTask(t *testing.T) {
	t.Parallel()

	// Generate UUIDs for consistent test data
	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()
	otherUserID := uuid.New().String()
	otherTaskID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		taskID        string
		mockError     error
		expectedError error
		shouldPanic   bool
	}{
		{
			name:          "successful deletion",
			userID:        testUserID,
			taskID:        testTaskID,
			mockError:     nil,
			expectedError: nil,
			shouldPanic:   false,
		},
		{
			name:          "repository error",
			userID:        otherUserID,
			taskID:        otherTaskID,
			mockError:     errors.New("database error"),
			expectedError: errors.New("database error"),
			shouldPanic:   false,
		},
		{
			name:        "empty userID should panic",
			userID:      "",
			taskID:      otherTaskID,
			shouldPanic: true,
		},
		{
			name:        "empty taskID should panic",
			userID:      testUserID,
			taskID:      "",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockRepo := &MockTaskRepository{}
			controller := NewTask(mockRepo)
			ctx := context.Background()

			if !tt.shouldPanic {
				mockRepo.On("Delete", ctx, tt.userID, tt.taskID).Return(tt.mockError)
			}

			// Act & Assert
			if tt.shouldPanic {
				assert.Panics(t, func() {
					controller.DeleteTask(ctx, tt.userID, tt.taskID)
				})
			} else {
				err := controller.DeleteTask(ctx, tt.userID, tt.taskID)

				// Assert
				if tt.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				} else {
					assert.NoError(t, err)
				}

				mockRepo.AssertExpectations(t)
			}
		})
	}
}
