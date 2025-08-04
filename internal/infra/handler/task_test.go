package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"strings"
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSimpleTaskRepository implements task.TaskRepository for simple testing
type MockSimpleTaskRepository struct {
	tasks map[string]*task.Task
}

func NewMockSimpleTaskRepository() *MockSimpleTaskRepository {
	return &MockSimpleTaskRepository{
		tasks: make(map[string]*task.Task),
	}
}

func (m *MockSimpleTaskRepository) FindAllByUserID(ctx context.Context, userID string) ([]*task.Task, error) {
	var result []*task.Task

	for _, t := range m.tasks {
		if t.UserID() == userID {
			result = append(result, t)
		}
	}

	if len(result) == 0 {
		return []*task.Task{}, nil
	}

	return result, nil
}

func (m *MockSimpleTaskRepository) FindById(ctx context.Context, userID, id string) (*task.Task, error) {
	if t, exists := m.tasks[id]; exists && t.UserID() == userID {
		return t, nil
	}

	return nil, task.ErrTaskNotFound
}

func (m *MockSimpleTaskRepository) Create(ctx context.Context, userID, title string) (*task.Task, error) {
	id := uuid.New().String()
	t := task.NewTask(id, title, userID)
	m.tasks[id] = t

	return t, nil
}

func (m *MockSimpleTaskRepository) Update(ctx context.Context, userID, id, title string) (*task.Task, error) {
	if t, exists := m.tasks[id]; exists && t.UserID() == userID {
		err := t.UpdateTitle(title)
		if err != nil {
			return nil, err
		}

		return t, nil
	}

	return nil, task.ErrTaskNotFound
}

func (m *MockSimpleTaskRepository) Delete(ctx context.Context, userID, id string) error {
	if t, exists := m.tasks[id]; exists && t.UserID() == userID {
		delete(m.tasks, id)

		return nil
	}

	return task.ErrTaskNotFound
}

// MockSimpleHealthService implements service.HealthService for simple testing
type MockSimpleHealthService struct{}

func (m *MockSimpleHealthService) CheckHealth(ctx context.Context) service.HealthStatus {
	return service.HealthStatus{
		Status:    "UP",
		Timestamp: time.Now(),
		Components: map[string]service.HealthComponent{
			"database": {
				Status: "UP",
				Details: map[string]interface{}{
					"connection": "mock",
				},
			},
		},
	}
}

func setupSimpleTestServer() (*TaskServer, *MockSimpleTaskRepository) {
	mockRepo := NewMockSimpleTaskRepository()
	mockHealthService := &MockSimpleHealthService{}
	taskController := controller.NewTask(mockRepo)
	server := NewTaskServer(*taskController, mockHealthService)

	return server, mockRepo
}

func TestSimple_TaskGetAllTasks(t *testing.T) {
	t.Parallel()

	server, mockRepo := setupSimpleTestServer()

	// Generate UUID for test user
	testUserID := uuid.New().String()

	// Pre-create some tasks
	ctx := context.Background()
	mockRepo.Create(ctx, testUserID, "Task 1")
	mockRepo.Create(ctx, testUserID, "Task 2")

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

func TestSimple_TaskCreateTask(t *testing.T) {
	t.Parallel()

	server, _ := setupSimpleTestServer()

	// Generate UUID for test user
	testUserID := uuid.New().String()

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

	var task generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.NotEmpty(t, task.Id)
	assert.Equal(t, "New Task", task.Title)
}

func TestSimple_TaskGetTask(t *testing.T) {
	t.Parallel()

	server, mockRepo := setupSimpleTestServer()

	// Generate UUID for test user
	testUserID := uuid.New().String()

	// Pre-create a task
	ctx := context.Background()
	createdTask, err := mockRepo.Create(ctx, testUserID, "Test Task")
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks/"+createdTask.ID(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err = server.TaskGetTask(c, createdTask.ID())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var task generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, createdTask.ID(), task.Id)
	assert.Equal(t, "Test Task", task.Title)
}

func TestSimple_TaskUpdateTask(t *testing.T) {
	t.Parallel()

	server, mockRepo := setupSimpleTestServer()

	// Generate UUID for test user
	testUserID := uuid.New().String()

	// Pre-create a task
	ctx := context.Background()
	createdTask, err := mockRepo.Create(ctx, testUserID, "Original Task")
	require.NoError(t, err)

	e := echo.New()
	requestBody := `{"title": "Updated Task"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/"+createdTask.ID(), strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err = server.TaskUpdateTask(c, createdTask.ID())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var task generated.Task

	err = json.Unmarshal(rec.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, createdTask.ID(), task.Id)
	assert.Equal(t, "Updated Task", task.Title)
}

func TestSimple_TaskDeleteTask(t *testing.T) {
	t.Parallel()

	server, mockRepo := setupSimpleTestServer()

	// Generate UUID for test user
	testUserID := uuid.New().String()

	// Pre-create a task
	ctx := context.Background()
	createdTask, err := mockRepo.Create(ctx, testUserID, "Task to Delete")
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+createdTask.ID(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	// Act
	err = server.TaskDeleteTask(c, createdTask.ID())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify task was deleted
	_, err = mockRepo.FindById(ctx, testUserID, createdTask.ID())
	assert.ErrorIs(t, err, task.ErrTaskNotFound)
}

func TestSimple_HealthGetHealth(t *testing.T) {
	t.Parallel()

	server, _ := setupSimpleTestServer()

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

func TestSimple_AuthenticationRequired(t *testing.T) {
	t.Parallel()

	server, _ := setupSimpleTestServer()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Don't set user_id to simulate missing authentication

	// Act
	err := server.TaskGetAllTasks(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSimple_ValidationErrors(t *testing.T) {
	t.Parallel()

	server, _ := setupSimpleTestServer()

	// Generate UUID for test user
	testUserID := uuid.New().String()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{"empty title", `{"title": ""}`, http.StatusBadRequest},
		{"whitespace title", `{"title": "   "}`, http.StatusBadRequest},
		{"title too long", `{"title": "` + strings.Repeat("a", 256) + `"}`, http.StatusBadRequest},
		{"valid title", `{"title": "Valid Task"}`, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
