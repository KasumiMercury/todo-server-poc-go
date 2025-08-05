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
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/mocks"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
)

//go:generate go run go.uber.org/mock/mockgen -source=../../domain/task/repository.go -destination=mocks/mock_task_repository.go -package=mocks

func setupTestServer(ctrl *gomock.Controller) (*TaskServer, *mocks.MockTaskRepository, *mocks.MockHealthService) {
	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockHealthService := mocks.NewMockHealthService(ctrl)
	taskController := controller.NewTask(mockRepo)
	server := NewTaskServer(*taskController, mockHealthService)

	return server, mockRepo, mockHealthService
}

func TestTaskGetAllTasks(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, mockRepo, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()

	task1 := task.NewTask(uuid.New().String(), "Task 1", testUserID)
	task2 := task.NewTask(uuid.New().String(), "Task 2", testUserID)
	mockRepo.EXPECT().FindAllByUserID(gomock.Any(), testUserID).Return([]*task.Task{task1, task2}, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := server.TaskGetAllTasks(c)

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

	server, mockRepo, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()

	createdTask := task.NewTask(taskID, "New Task", testUserID)
	mockRepo.EXPECT().Create(gomock.Any(), testUserID, "New Task").Return(createdTask, nil)

	e := echo.New()
	requestBody := `{"title": "New Task"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := server.TaskCreateTask(c)

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

	server, mockRepo, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()

	existingTask := task.NewTask(taskID, "Test Task", testUserID)
	mockRepo.EXPECT().FindById(gomock.Any(), testUserID, taskID).Return(existingTask, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := server.TaskGetTask(c, taskID)

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

	server, mockRepo, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()

	existingTask := task.NewTask(taskID, "Original Task", testUserID)
	updatedTask := task.NewTask(taskID, "Updated Task", testUserID)

	mockRepo.EXPECT().FindById(gomock.Any(), testUserID, taskID).Return(existingTask, nil)
	mockRepo.EXPECT().Update(gomock.Any(), testUserID, taskID, "Updated Task").Return(updatedTask, nil)

	e := echo.New()
	requestBody := `{"title": "Updated Task"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := server.TaskUpdateTask(c, taskID)

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

	server, mockRepo, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	taskID := uuid.New().String()

	existingTask := task.NewTask(taskID, "Task to Delete", testUserID)

	mockRepo.EXPECT().FindById(gomock.Any(), testUserID, taskID).Return(existingTask, nil)
	mockRepo.EXPECT().Delete(gomock.Any(), testUserID, taskID).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err := server.TaskDeleteTask(c, taskID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHealthGetHealth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, _, mockHealthService := setupTestServer(ctrl)

	mockHealthService.EXPECT().CheckHealth(gomock.Any()).Return(
		service.HealthStatus{
			Status:    "UP",
			Timestamp: time.Now(),
			Components: map[string]service.HealthComponent{
				"database": {
					Status:  "UP",
					Details: map[string]interface{}{"connection": "mock"},
				},
			},
		},
	)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := server.HealthGetHealth(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var healthResponse generated.HealthStatus

	err = json.Unmarshal(rec.Body.Bytes(), &healthResponse)
	require.NoError(t, err)
	assert.Equal(t, generated.HealthStatusStatusUP, healthResponse.Status)
}

func TestAuthenticationRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, _, _ := setupTestServer(ctrl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := server.TaskGetAllTasks(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, mockRepo, _ := setupTestServer(ctrl)

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
			createdTask := task.NewTask(taskID, "Valid Task", testUserID)
			mockRepo.EXPECT().Create(gomock.Any(), testUserID, "Valid Task").Return(createdTask, nil)
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
			err := server.TaskCreateTask(c)

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

	server, mockRepo, _ := setupTestServer(ctrl)

	testUserID := uuid.New().String()
	nonExistentTaskID := uuid.New().String()

	tests := []struct {
		name      string
		setupMock func()
		operation func() error
	}{
		{
			name: "get non-existent task",
			setupMock: func() {
				mockRepo.EXPECT().FindById(gomock.Any(), testUserID, nonExistentTaskID).Return(nil, task.ErrTaskNotFound)
			},
			operation: func() error {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/tasks/"+nonExistentTaskID, nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user_id", testUserID)

				err := server.TaskGetTask(c, nonExistentTaskID)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, rec.Code)

				return nil
			},
		},
		{
			name: "update non-existent task",
			setupMock: func() {
				mockRepo.EXPECT().FindById(gomock.Any(), testUserID, nonExistentTaskID).Return(nil, task.ErrTaskNotFound)
			},
			operation: func() error {
				e := echo.New()
				requestBody := `{"title": "Updated Task"}`
				req := httptest.NewRequest(http.MethodPut, "/tasks/"+nonExistentTaskID, strings.NewReader(requestBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user_id", testUserID)

				err := server.TaskUpdateTask(c, nonExistentTaskID)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, rec.Code)

				return nil
			},
		},
		{
			name: "delete non-existent task",
			setupMock: func() {
				mockRepo.EXPECT().FindById(gomock.Any(), testUserID, nonExistentTaskID).Return(nil, task.ErrTaskNotFound)
			},
			operation: func() error {
				e := echo.New()
				req := httptest.NewRequest(http.MethodDelete, "/tasks/"+nonExistentTaskID, nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user_id", testUserID)

				err := server.TaskDeleteTask(c, nonExistentTaskID)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, rec.Code)

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

	server, _, _ := setupTestServer(ctrl)

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
				err = server.TaskCreateTask(c)
			case http.MethodPut:
				parts := strings.Split(tt.url, "/")
				taskID := parts[len(parts)-1]
				err = server.TaskUpdateTask(c, taskID)
			}

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}
