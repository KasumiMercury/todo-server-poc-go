package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/mocks"
)

//go:generate go run go.uber.org/mock/mockgen -source=../../domain/task/repository.go -destination=mocks/mock_task_repository.go -package=mocks

func setupTestServer(ctrl *gomock.Controller) (*TaskHandler, *mocks.MockTaskRepository) {
	mockRepo := mocks.NewMockTaskRepository(ctrl)
	taskController := controller.NewTask(mockRepo)
	handler := NewTaskHandler(*taskController)

	return handler, mockRepo
}

func TestTaskGetAllTasks(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	userID := createUserID(testUserID)

	task1 := task.NewTask(createTaskID(uuid.New().String()), "Task 1", userID)
	task2 := task.NewTask(createTaskID(uuid.New().String()), "Task 2", userID)
	mockRepo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return([]*task.Task{task1, task2}, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := handler.GetAllTasks(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var tasks []generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &tasks)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestTaskCreateTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()
	userID := createUserID(testUserID)

	createdTask := task.NewTask(createTaskID(taskID), "New Task", userID)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdTask, nil)

	e := echo.New()
	requestBody := `{"title": "New Task"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := handler.CreateTask(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var responseTask generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
	require.NoError(t, err)
	assert.Equal(t, taskID, responseTask.Id)
	assert.Equal(t, "New Task", responseTask.Title)
}

func TestTaskGetTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()
	userID := createUserID(testUserID)
	taskDomainID := createTaskID(taskID)

	existingTask := task.NewTask(taskDomainID, "Test Task", userID)
	mockRepo.EXPECT().FindById(gomock.Any(), userID, taskDomainID).Return(existingTask, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := handler.GetTask(c, taskID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var responseTask generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
	require.NoError(t, err)
	assert.Equal(t, taskID, responseTask.Id)
	assert.Equal(t, "Test Task", responseTask.Title)
}

func TestTaskUpdateTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()
	userID := createUserID(testUserID)
	taskDomainID := createTaskID(taskID)

	existingTask := task.NewTask(taskDomainID, "Original Task", userID)
	updatedTask := task.NewTask(taskDomainID, "Updated Task", userID)

	mockRepo.EXPECT().FindById(gomock.Any(), userID, taskDomainID).Return(existingTask, nil)
	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(updatedTask, nil)

	e := echo.New()
	requestBody := `{"title": "Updated Task"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := handler.UpdateTask(c, taskID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var responseTask generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
	require.NoError(t, err)
	assert.Equal(t, taskID, responseTask.Id)
	assert.Equal(t, "Updated Task", responseTask.Title)
}

func TestTaskDeleteTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()
	userID := createUserID(testUserID)
	taskDomainID := createTaskID(taskID)

	mockRepo.EXPECT().Delete(gomock.Any(), userID, taskDomainID).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := handler.DeleteTask(c, taskID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAuthenticationRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := setupTestServer(ctrl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetAllTasks(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		setupMock      func()
	}{
		{"empty title", `{"title": ""}`, http.StatusBadRequest, func() {}},
		{"whitespace title", `{"title": "   "}`, http.StatusBadRequest, func() {}},
		{"title too long", `{"title": "` + strings.Repeat("a", 256) + `"}`, http.StatusBadRequest, func() {}},
		{"valid title", `{"title": "Valid Task"}`, http.StatusCreated, func() {
			taskID := uuid.New().String()
			userID := createUserID(testUserID)
			createdTask := task.NewTask(createTaskID(taskID), "Valid Task", userID)
			mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdTask, nil)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", testUserID)

			// Act
			err := handler.CreateTask(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestTaskNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	nonExistentTaskID := uuid.New().String()
	userID := createUserID(testUserID)
	taskDomainID := createTaskID(nonExistentTaskID)

	tests := []struct {
		name      string
		setupMock func()
		operation func() error
	}{
		{
			name: "get non-existent task",
			setupMock: func() {
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskDomainID).Return(nil, task.ErrTaskNotFound)
			},
			operation: func() error {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/tasks/"+nonExistentTaskID, nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user_id", testUserID)

				err := handler.GetTask(c, nonExistentTaskID)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, rec.Code)

				return nil
			},
		},
		{
			name: "update non-existent task",
			setupMock: func() {
				mockRepo.EXPECT().FindById(gomock.Any(), userID, taskDomainID).Return(nil, task.ErrTaskNotFound)
			},
			operation: func() error {
				e := echo.New()
				requestBody := `{"title": "Updated Task"}`
				req := httptest.NewRequest(http.MethodPut, "/tasks/"+nonExistentTaskID, strings.NewReader(requestBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user_id", testUserID)

				err := handler.UpdateTask(c, nonExistentTaskID)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, rec.Code)

				return nil
			},
		},
		{
			name: "delete non-existent task",
			setupMock: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), userID, taskDomainID).Return(nil)
			},
			operation: func() error {
				e := echo.New()
				req := httptest.NewRequest(http.MethodDelete, "/tasks/"+nonExistentTaskID, nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user_id", testUserID)

				err := handler.DeleteTask(c, nonExistentTaskID)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNoContent, rec.Code)

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.setupMock()
			err := tt.operation()
			assert.NoError(t, err)
		})
	}
}

func TestInvalidJSONRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()

	tests := []struct {
		name        string
		requestBody string
		method      string
		url         string
	}{
		{
			name:        "invalid JSON in create task",
			requestBody: `{"title": "Invalid JSON"`,
			method:      http.MethodPost,
			url:         "/tasks",
		},
		{
			name:        "invalid JSON in update task",
			requestBody: `{"title": }`,
			method:      http.MethodPut,
			url:         "/tasks/" + uuid.New().String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", testUserID)

			var err error

			switch tt.method {
			case http.MethodPost:
				err = handler.CreateTask(c)
			case http.MethodPut:
				parts := strings.Split(tt.url, "/")
				taskID := parts[len(parts)-1]
				err = handler.UpdateTask(c, taskID)
			}

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestTaskRepositoryErrors(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New().String()
	userID := createUserID(testUserID)

	tests := []struct {
		name               string
		operation          string
		setupMock          func(*mocks.MockTaskRepository)
		requestBody        string
		taskID             string
		expectedStatusCode int
		expectedErrorType  string
	}{
		{
			name:      "database connection timeout on GetAllTasks",
			operation: "GetAllTasks",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return(nil, context.DeadlineExceeded)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "timeout",
		},
		{
			name:      "database connection error on GetAllTasks",
			operation: "GetAllTasks",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.EXPECT().FindAllByUserID(gomock.Any(), userID).Return(nil, fmt.Errorf("database connection failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "database",
		},
		{
			name:        "constraint violation on CreateTask",
			operation:   "CreateTask",
			requestBody: `{"title": "Duplicate Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("UNIQUE constraint failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "constraint",
		},
		{
			name:        "transaction rollback on CreateTask",
			operation:   "CreateTask",
			requestBody: `{"title": "Transaction Failed"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("transaction rolled back"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "transaction",
		},
		{
			name:      "network failure on GetTask",
			operation: "GetTask",
			taskID:    uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.EXPECT().FindById(gomock.Any(), userID, gomock.Any()).Return(nil, fmt.Errorf("network I/O timeout"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "network",
		},
		{
			name:        "optimistic locking failure on UpdateTask",
			operation:   "UpdateTask",
			taskID:      uuid.New().String(),
			requestBody: `{"title": "Updated Task"}`,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				existingTask := task.NewTask(createTaskID(uuid.New().String()), "Original Task", userID)
				mockRepo.EXPECT().FindById(gomock.Any(), userID, gomock.Any()).Return(existingTask, nil)
				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("optimistic locking failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "conflict",
		},
		{
			name:      "foreign key constraint on DeleteTask",
			operation: "DeleteTask",
			taskID:    uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.EXPECT().Delete(gomock.Any(), userID, gomock.Any()).Return(fmt.Errorf("FOREIGN KEY constraint failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockRepo := setupTestServer(ctrl)
			tt.setupMock(mockRepo)

			e := echo.New()

			var (
				req *http.Request
				err error
			)

			switch tt.operation {
			case "GetAllTasks":
				req = httptest.NewRequest(http.MethodGet, "/tasks", nil)
			case "CreateTask":
				req = httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			case "GetTask":
				req = httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)
			case "UpdateTask":
				req = httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID, strings.NewReader(tt.requestBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			case "DeleteTask":
				req = httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID, nil)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", testUserID)

			// Act
			switch tt.operation {
			case "GetAllTasks":
				err = handler.GetAllTasks(c)
			case "CreateTask":
				err = handler.CreateTask(c)
			case "GetTask":
				err = handler.GetTask(c, tt.taskID)
			case "UpdateTask":
				err = handler.UpdateTask(c, tt.taskID)
			case "DeleteTask":
				err = handler.DeleteTask(c, tt.taskID)
			}

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode >= 400 {
				var errorResponse generated.Error

				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				assert.NoError(t, unmarshalErr)
				assert.NotEmpty(t, errorResponse.Message)
			}
		})
	}
}

func TestTaskEdgeCaseValidation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)
	testUserID := uuid.New().String()
	userID := createUserID(testUserID)

	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		setupMock          func()
	}{
		{
			name:               "unicode title",
			requestBody:        `{"title": "タスク 🚀 Test ñoño ënd"}`,
			expectedStatusCode: http.StatusCreated,
			setupMock: func() {
				taskID := uuid.New().String()
				createdTask := task.NewTask(createTaskID(taskID), "タスク 🚀 Test ñoño ënd", userID)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdTask, nil)
			},
		},
		{
			name:               "special characters in title",
			requestBody:        `{"title": "Task with <script>alert('xss')</script> & special chars"}`,
			expectedStatusCode: http.StatusCreated,
			setupMock: func() {
				taskID := uuid.New().String()
				title := "Task with <script>alert('xss')</script> & special chars"
				createdTask := task.NewTask(createTaskID(taskID), title, userID)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdTask, nil)
			},
		},
		{
			name:               "maximum length title (255 chars)",
			requestBody:        fmt.Sprintf(`{"title": "%s"}`, strings.Repeat("a", 255)),
			expectedStatusCode: http.StatusCreated,
			setupMock: func() {
				taskID := uuid.New().String()
				title := strings.Repeat("a", 255)
				createdTask := task.NewTask(createTaskID(taskID), title, userID)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdTask, nil)
			},
		},
		{
			name:               "title with newlines and tabs",
			requestBody:        `{"title": "Task with\nnewlines\tand\ttabs"}`,
			expectedStatusCode: http.StatusCreated,
			setupMock: func() {
				taskID := uuid.New().String()
				title := "Task with\nnewlines\tand\ttabs"
				createdTask := task.NewTask(createTaskID(taskID), title, userID)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdTask, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			tt.setupMock()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", testUserID)

			// Act
			err := handler.CreateTask(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode == http.StatusCreated {
				var responseTask generated.Task

				err = json.Unmarshal(rec.Body.Bytes(), &responseTask)
				require.NoError(t, err)
				assert.NotEmpty(t, responseTask.Id)
				assert.NotEmpty(t, responseTask.Title)
			}
		})
	}
}

func TestTaskMalformedRequests(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := setupTestServer(ctrl)
	testUserID := uuid.New().String()

	tests := []struct {
		name               string
		requestBody        string
		contentType        string
		expectedStatusCode int
	}{
		{
			name:               "wrong content type",
			requestBody:        `{"title": "Test Task"}`,
			contentType:        "text/plain",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "XML instead of JSON",
			requestBody:        `<task><title>Test Task</title></task>`,
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty request body",
			requestBody:        "",
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "null JSON",
			requestBody:        "null",
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "JSON array instead of object",
			requestBody:        `[{"title": "Task 1"}, {"title": "Task 2"}]`,
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "malformed JSON with trailing comma",
			requestBody:        `{"title": "Test Task",}`,
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, tt.contentType)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", testUserID)

			// Act
			err := handler.CreateTask(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			if tt.expectedStatusCode >= 400 {
				var errorResponse generated.Error

				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				assert.NoError(t, unmarshalErr)
				assert.NotEmpty(t, errorResponse.Message)
			}
		})
	}
}

func TestTaskContextErrors(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := setupTestServer(ctrl)

	tests := []struct {
		name               string
		setupContext       func(echo.Context)
		setupMock          func()
		operation          string
		expectedStatusCode int
	}{
		{
			name: "empty user_id string",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "")
			},
			setupMock:          func() {},
			operation:          "GetAllTasks",
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			tt.setupMock()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			tt.setupContext(c)

			// Act
			var err error

			switch tt.operation {
			case "GetAllTasks":
				err = handler.GetAllTasks(c)
			}

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, rec.Code)
		})
	}
}

func TestTaskConcurrentOperations(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := setupTestServer(ctrl)
	testUserID := uuid.New().String()
	testTaskID := uuid.New().String()
	userID := createUserID(testUserID)
	taskDomainID := createTaskID(testTaskID)

	existingTask := task.NewTask(taskDomainID, "Original Task", userID)

	// First call succeeds
	mockRepo.EXPECT().FindById(gomock.Any(), userID, taskDomainID).Return(existingTask, nil).Times(1)
	// Second call fails due to concurrent modification
	mockRepo.EXPECT().FindById(gomock.Any(), userID, taskDomainID).Return(nil, fmt.Errorf("record modified by another transaction")).Times(1)

	e := echo.New()

	req1 := httptest.NewRequest(http.MethodGet, "/tasks/"+testTaskID, nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	c1.Set("user_id", testUserID)

	req2 := httptest.NewRequest(http.MethodGet, "/tasks/"+testTaskID, nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.Set("user_id", testUserID)

	// Act
	err1 := handler.GetTask(c1, testTaskID)
	err2 := handler.GetTask(c2, testTaskID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.True(t, (rec1.Code == http.StatusOK && rec2.Code == http.StatusInternalServerError) ||
		(rec1.Code == http.StatusInternalServerError && rec2.Code == http.StatusOK))
}
