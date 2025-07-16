package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/task"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// MockTaskRepository implements repository.TaskRepository for testing
type MockTaskRepository struct{}

// MockHealthService implements service.HealthService for testing
type MockHealthService struct{}

func (m *MockHealthService) CheckHealth(ctx context.Context) service.HealthStatus {
	return service.HealthStatus{
		Status:    "UP",
		Timestamp: time.Now(),
		Components: map[string]service.HealthComponent{
			"database": {
				Status: "UP",
				Details: map[string]interface{}{
					"connection":   "PostgreSQL",
					"responseTime": "5ms",
				},
			},
		},
	}
}

func (m *MockTaskRepository) FindAll(ctx context.Context) ([]*task.Task, error) {
	// Return mock tasks
	task1 := task.NewTask("1", "Test Task 1")
	task2 := task.NewTask("2", "Test Task 2")
	return []*task.Task{task1, task2}, nil
}

func (m *MockTaskRepository) FindById(ctx context.Context, id string) (*task.Task, error) {
	if id == "1" {
		return task.NewTask("1", "Test Task 1"), nil
	}
	return nil, nil
}

func (m *MockTaskRepository) Create(ctx context.Context, title string) (*task.Task, error) {
	return task.NewTask("new-id", title), nil
}

func (m *MockTaskRepository) Update(ctx context.Context, id, title string) (*task.Task, error) {
	return task.NewTask(id, title), nil
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

func setupTestRouter() *echo.Echo {
	router := echo.New()

	cfg := &config.Config{
		JWTSecret:    "secret-key-for-testing",
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
	}

	// Add CORS middleware globally
	router.Use(handler.CORSMiddleware(*cfg))

	mockRepo := &MockTaskRepository{}
	taskController := controller.NewTask(mockRepo)
	mockHealthService := &MockHealthService{}
	taskServer := handler.NewTaskServer(*taskController, mockHealthService)

	// Setup JWT middleware for protected endpoints
	jwtMiddleware := handler.JWTMiddleware(cfg.JWTSecret)

	// Create wrapper for generated handlers
	wrapper := generated.ServerInterfaceWrapper{
		Handler: taskServer,
	}

	// Register health endpoint without authentication
	router.GET("/health", wrapper.HealthGetHealth)

	// Create a group for protected task endpoints
	taskGroup := router.Group("/tasks")
	taskGroup.Use(jwtMiddleware)

	// Register task endpoints with JWT middleware
	taskGroup.GET("", wrapper.TaskGetAllTasks)
	taskGroup.POST("", wrapper.TaskCreateTask)
	taskGroup.GET("/:taskId", wrapper.TaskGetTask)
	taskGroup.PUT("/:taskId", wrapper.TaskUpdateTask)
	taskGroup.DELETE("/:taskId", wrapper.TaskDeleteTask)

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

func TestCORSOptionsEndpoint(t *testing.T) {
	router := setupTestRouter()

	t.Run("OPTIONS /tasks should return CORS headers without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/tasks", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Echo returns 204 No Content for OPTIONS requests
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		expectedOrigin := "http://localhost:5173"
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != expectedOrigin {
			t.Errorf("Expected Access-Control-Allow-Origin %s, got %s", expectedOrigin, origin)
		}

		// Check if Authorization header is allowed
		allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
		if !strings.Contains(allowedHeaders, "Authorization") {
			t.Errorf("Expected Access-Control-Allow-Headers to contain Authorization, got %s", allowedHeaders)
		}

		// Check if required methods are allowed
		allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
		requiredMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		for _, method := range requiredMethods {
			if !strings.Contains(allowedMethods, method) {
				t.Errorf("Expected Access-Control-Allow-Methods to contain %s, got %s", method, allowedMethods)
			}
		}
	})
}
