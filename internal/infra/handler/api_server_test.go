package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/mocks"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

// Helper functions for creating domain objects from strings in tests
func createTaskID(s string) task.TaskID {
	taskID, err := task.NewTaskID(s)
	if err != nil {
		panic("failed to create task ID: " + err.Error())
	}

	return taskID
}

func createUserID(s string) user.UserID {
	userID, err := user.NewUserID(s)
	if err != nil {
		panic("failed to create user ID: " + err.Error())
	}

	return userID
}

func TestNewAPIServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		setupMocks          func(ctrl *gomock.Controller) (controller.Task, service.HealthService)
		expectedNil         bool
		expectedFieldsValid bool
	}{
		{
			name: "valid dependencies",
			setupMocks: func(ctrl *gomock.Controller) (controller.Task, service.HealthService) {
				mockRepo := mocks.NewMockTaskRepository(ctrl)
				taskController := controller.NewTask(mockRepo)
				mockHealthService := mocks.NewMockHealthService(ctrl)

				return *taskController, mockHealthService
			},
			expectedNil:         false,
			expectedFieldsValid: true,
		},
		{
			name: "nil task controller",
			setupMocks: func(ctrl *gomock.Controller) (controller.Task, service.HealthService) {
				mockHealthService := mocks.NewMockHealthService(ctrl)

				return controller.Task{}, mockHealthService
			},
			expectedNil:         false,
			expectedFieldsValid: false,
		},
		{
			name: "nil health service",
			setupMocks: func(ctrl *gomock.Controller) (controller.Task, service.HealthService) {
				mockRepo := mocks.NewMockTaskRepository(ctrl)
				taskController := controller.NewTask(mockRepo)

				return *taskController, nil
			},
			expectedNil:         false,
			expectedFieldsValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			taskController, healthService := tt.setupMocks(ctrl)

			// Act
			apiServer := NewAPIServer(taskController, healthService)

			// Assert
			if tt.expectedNil {
				assert.Nil(t, apiServer)
			} else {
				assert.NotNil(t, apiServer)

				if tt.expectedFieldsValid {
					assert.NotNil(t, apiServer.taskHandler)
					assert.NotNil(t, apiServer.healthHandler)
				}
			}
		})
	}
}

func TestAPIServer_HealthGetHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		healthStatus       service.HealthStatus
		expectedStatusCode int
		expectedStatus     string
	}{
		{
			name: "healthy system",
			healthStatus: service.HealthStatus{
				Status:    "UP",
				Timestamp: time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
				Components: map[string]service.HealthComponent{
					"database": {
						Status: "UP",
						Details: map[string]interface{}{
							"connection":   "established",
							"responseTime": "5ms",
						},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedStatus:     "UP",
		},
		{
			name: "unhealthy system",
			healthStatus: service.HealthStatus{
				Status:    "DOWN",
				Timestamp: time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
				Components: map[string]service.HealthComponent{
					"database": {
						Status: "DOWN",
						Details: map[string]interface{}{
							"error": "connection timeout",
						},
					},
				},
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedStatus:     "DOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTaskRepository(ctrl)
			taskController := controller.NewTask(mockRepo)
			mockHealthService := mocks.NewMockHealthService(ctrl)

			mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(tt.healthStatus)

			apiServer := NewAPIServer(*taskController, mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Act
			err := apiServer.HealthGetHealth(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			var response generated.HealthStatus

			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, string(response.Status))
			assert.Equal(t, tt.healthStatus.Timestamp, response.Timestamp)
		})
	}
}

func TestAPIServer_TaskGetAllTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		userID             string
		setupMock          func(*mocks.MockTaskRepository, user.UserID)
		expectedStatusCode int
		expectedTaskCount  int
	}{
		{
			name:   "successful retrieval of tasks",
			userID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				task1 := task.NewTask(createTaskID(uuid.New().String()), "Task 1", userID)
				task2 := task.NewTask(createTaskID(uuid.New().String()), "Task 2", userID)
				mockRepo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return([]*task.Task{task1, task2}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedTaskCount:  2,
		},
		{
			name:   "empty task list",
			userID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				mockRepo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return([]*task.Task{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedTaskCount:  0,
		},
		{
			name:   "missing user_id in context",
			userID: "",
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				// No mock expectations - should fail before repository call
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedTaskCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTaskRepository(ctrl)
			mockHealthService := mocks.NewMockHealthService(ctrl)
			taskController := controller.NewTask(mockRepo)

			var domainUserID user.UserID
			if tt.userID != "" {
				domainUserID = createUserID(tt.userID)
			}

			tt.setupMock(mockRepo, domainUserID)

			apiServer := NewAPIServer(*taskController, mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Act
			err := apiServer.TaskGetAllTasks(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode == http.StatusOK {
				var tasks []generated.Task

				err = json.Unmarshal(rec.Body.Bytes(), &tasks)
				require.NoError(t, err)
				assert.Len(t, tasks, tt.expectedTaskCount)
			}
		})
	}
}

func TestAPIServer_TaskCreateTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		userID             string
		requestBody        string
		setupMock          func(*mocks.MockTaskRepository, user.UserID)
		expectedStatusCode int
		expectedTitle      string
	}{
		{
			name:        "successful task creation",
			userID:      uuid.New().String(),
			requestBody: `{"title": "New Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				taskID := createTaskID(uuid.New().String())
				createdTask := task.NewTask(taskID, "New Task", userID)
				mockRepo.EXPECT().Create(gomock.Any(), userID, "New Task").Return(createdTask, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedTitle:      "New Task",
		},
		{
			name:        "invalid JSON request",
			userID:      uuid.New().String(),
			requestBody: `{"title": "Invalid JSON"`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				// No mock expectations - should fail before repository call
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "empty title validation",
			userID:      uuid.New().String(),
			requestBody: `{"title": ""}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				// No mock expectations - should fail validation
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "missing user_id in context",
			userID:      "",
			requestBody: `{"title": "Test Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID) {
				// No mock expectations - should fail authentication
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTaskRepository(ctrl)
			mockHealthService := mocks.NewMockHealthService(ctrl)
			taskController := controller.NewTask(mockRepo)

			var domainUserID user.UserID
			if tt.userID != "" {
				domainUserID = createUserID(tt.userID)
			}

			tt.setupMock(mockRepo, domainUserID)

			apiServer := NewAPIServer(*taskController, mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Act
			err := apiServer.TaskCreateTask(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode == http.StatusCreated {
				var responseTask generated.Task

				err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTitle, responseTask.Title)
			}
		})
	}
}

func TestAPIServer_TaskGetTask(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()

	tests := []struct {
		name               string
		userID             string
		taskID             string
		setupMock          func(*mocks.MockTaskRepository, user.UserID, task.TaskID)
		expectedStatusCode int
		expectedTitle      string
	}{
		{
			name:   "successful task retrieval",
			userID: testUserID,
			taskID: testTaskID,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				existingTask := task.NewTask(taskID, "Test Task", userID)
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskID).Return(existingTask, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedTitle:      "Test Task",
		},
		{
			name:   "task not found",
			userID: testUserID,
			taskID: testTaskID,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskID).Return(nil, task.ErrTaskNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "missing user_id in context",
			userID: "",
			taskID: testTaskID,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				// No mock expectations - should fail authentication
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTaskRepository(ctrl)
			mockHealthService := mocks.NewMockHealthService(ctrl)
			taskController := controller.NewTask(mockRepo)

			var (
				domainUserID user.UserID
				domainTaskID task.TaskID
			)

			if tt.userID != "" {
				domainUserID = createUserID(tt.userID)
			}

			if tt.taskID != "" {
				domainTaskID = createTaskID(tt.taskID)
			}

			tt.setupMock(mockRepo, domainUserID, domainTaskID)

			apiServer := NewAPIServer(*taskController, mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Act
			err := apiServer.TaskGetTask(c, tt.taskID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode == http.StatusOK {
				var responseTask generated.Task

				err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTitle, responseTask.Title)
				assert.Equal(t, tt.taskID, responseTask.Id)
			}
		})
	}
}

func TestAPIServer_TaskUpdateTask(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()

	tests := []struct {
		name               string
		userID             string
		taskID             string
		requestBody        string
		setupMock          func(*mocks.MockTaskRepository, user.UserID, task.TaskID)
		expectedStatusCode int
		expectedTitle      string
	}{
		{
			name:        "successful task update",
			userID:      testUserID,
			taskID:      testTaskID,
			requestBody: `{"title": "Updated Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				existingTask := task.NewTask(taskID, "Original Task", userID)
				updatedTask := task.NewTask(taskID, "Updated Task", userID)
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskID).Return(existingTask, nil)
				mockRepo.EXPECT().Update(gomock.Any(), userID, taskID, "Updated Task").Return(updatedTask, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedTitle:      "Updated Task",
		},
		{
			name:        "task not found",
			userID:      testUserID,
			taskID:      testTaskID,
			requestBody: `{"title": "Updated Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskID).Return(nil, task.ErrTaskNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:        "invalid JSON request",
			userID:      testUserID,
			taskID:      testTaskID,
			requestBody: `{"title": }`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				// No mock expectations - should fail before repository call
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "missing user_id in context",
			userID:      "",
			taskID:      testTaskID,
			requestBody: `{"title": "Updated Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				// No mock expectations - should fail authentication
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTaskRepository(ctrl)
			mockHealthService := mocks.NewMockHealthService(ctrl)
			taskController := controller.NewTask(mockRepo)

			var (
				domainUserID user.UserID
				domainTaskID task.TaskID
			)

			if tt.userID != "" {
				domainUserID = createUserID(tt.userID)
			}

			if tt.taskID != "" {
				domainTaskID = createTaskID(tt.taskID)
			}

			tt.setupMock(mockRepo, domainUserID, domainTaskID)

			apiServer := NewAPIServer(*taskController, mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Act
			err := apiServer.TaskUpdateTask(c, tt.taskID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode == http.StatusOK {
				var responseTask generated.Task

				err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTitle, responseTask.Title)
			}
		})
	}
}

func TestAPIServer_TaskDeleteTask(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()

	tests := []struct {
		name               string
		userID             string
		taskID             string
		setupMock          func(*mocks.MockTaskRepository, user.UserID, task.TaskID)
		expectedStatusCode int
	}{
		{
			name:   "successful task deletion",
			userID: testUserID,
			taskID: testTaskID,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				existingTask := task.NewTask(taskID, "Task to Delete", userID)
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskID).Return(existingTask, nil)
				mockRepo.EXPECT().Delete(gomock.Any(), userID, taskID).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:   "task not found",
			userID: testUserID,
			taskID: testTaskID,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskID).Return(nil, task.ErrTaskNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "missing user_id in context",
			userID: "",
			taskID: testTaskID,
			setupMock: func(mockRepo *mocks.MockTaskRepository, userID user.UserID, taskID task.TaskID) {
				// No mock expectations - should fail authentication
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTaskRepository(ctrl)
			mockHealthService := mocks.NewMockHealthService(ctrl)
			taskController := controller.NewTask(mockRepo)

			var (
				domainUserID user.UserID
				domainTaskID task.TaskID
			)

			if tt.userID != "" {
				domainUserID = createUserID(tt.userID)
			}

			if tt.taskID != "" {
				domainTaskID = createTaskID(tt.taskID)
			}

			tt.setupMock(mockRepo, domainUserID, domainTaskID)

			apiServer := NewAPIServer(*taskController, mockHealthService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Act
			err := apiServer.TaskDeleteTask(c, tt.taskID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode == http.StatusNoContent {
				// Body might be empty or contain "null" - both are acceptable for NoContent
				body := strings.TrimSpace(rec.Body.String())
				assert.True(t, body == "" || body == "null", "Expected empty body or null, got: %s", body)
			}
		})
	}
}

func TestAPIServer_IntegrationBehavior(t *testing.T) {
	t.Parallel()

	t.Run("handler delegation works correctly", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockTaskRepository(ctrl)
		mockHealthService := mocks.NewMockHealthService(ctrl)
		taskController := controller.NewTask(mockRepo)

		healthStatus := service.HealthStatus{
			Status:    "UP",
			Timestamp: time.Now(),
			Components: map[string]service.HealthComponent{
				"database": {Status: "UP"},
			},
		}
		mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(healthStatus)

		apiServer := NewAPIServer(*taskController, mockHealthService)

		// Act & Assert
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := apiServer.HealthGetHealth(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("error propagation from handlers", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockTaskRepository(ctrl)
		mockHealthService := mocks.NewMockHealthService(ctrl)
		taskController := controller.NewTask(mockRepo)

		apiServer := NewAPIServer(*taskController, mockHealthService)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		err := apiServer.TaskGetAllTasks(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestAPIServer_NilHandlerScenarios(t *testing.T) {
	t.Parallel()

	t.Run("api server with minimal valid setup", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockTaskRepository(ctrl)
		mockHealthService := mocks.NewMockHealthService(ctrl)
		taskController := controller.NewTask(mockRepo)

		// Act
		apiServer := NewAPIServer(*taskController, mockHealthService)

		// Assert
		assert.NotNil(t, apiServer)
		assert.NotNil(t, apiServer.taskHandler)
		assert.NotNil(t, apiServer.healthHandler)
	})
}
