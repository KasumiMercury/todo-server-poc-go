package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// MockTaskRepository implements repository.TaskRepository for testing
type MockTaskRepository struct{}

func (m *MockTaskRepository) FindAll(ctx context.Context) ([]*domain.Task, error) {
	// Return mock tasks
	task1 := domain.NewTask("1", "Test Task 1")
	task2 := domain.NewTask("2", "Test Task 2")
	return []*domain.Task{task1, task2}, nil
}

func (m *MockTaskRepository) FindById(ctx context.Context, id string) (*domain.Task, error) {
	if id == "1" {
		return domain.NewTask("1", "Test Task 1"), nil
	}
	return nil, nil
}

func (m *MockTaskRepository) Create(ctx context.Context, title string) (*domain.Task, error) {
	return domain.NewTask("new-id", title), nil
}

func (m *MockTaskRepository) Update(ctx context.Context, id, title string) (*domain.Task, error) {
	return domain.NewTask(id, title), nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func generateTestJWT() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, _ := token.SignedString([]byte("secret-key-for-testing"))
	return tokenString
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	cfg := &config.Config{
		JWTSecret: "secret-key-for-testing",
	}

	mockRepo := &MockTaskRepository{}
	taskController := controller.NewTask(mockRepo)
	taskServer := handler.NewTaskServer(*taskController)

	jwtMiddleware := handler.JWTMiddleware(cfg.JWTSecret)
	handler.RegisterHandlersWithOptions(router, taskServer, handler.GinServerOptions{
		Middlewares: []handler.MiddlewareFunc{
			handler.MiddlewareFunc(jwtMiddleware),
		},
	})

	return router
}

func TestJWTAuthenticationUnauthorized(t *testing.T) {
	router := setupTestRouter()

	t.Run("GET /tasks without JWT token should return 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tasks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("GET /tasks with invalid JWT token should return 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tasks", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestJWTAuthenticationAuthorized(t *testing.T) {
	router := setupTestRouter()
	testJWT := generateTestJWT()

	t.Run("GET /tasks with valid JWT token should return 200", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+testJWT)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("POST /tasks with valid JWT token should return 201", func(t *testing.T) {
		requestBody := map[string]string{"name": "New Test Task"}
		jsonData, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("GET /tasks/1 with valid JWT token should return 200", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tasks/1", nil)
		req.Header.Set("Authorization", "Bearer "+testJWT)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("PUT /tasks/1 with valid JWT token should return 200", func(t *testing.T) {
		updateBody := map[string]string{"name": "Updated Test Task"}
		jsonData, _ := json.Marshal(updateBody)
		req, _ := http.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("DELETE /tasks/1 with valid JWT token should return 204", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/tasks/1", nil)
		req.Header.Set("Authorization", "Bearer "+testJWT)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}
