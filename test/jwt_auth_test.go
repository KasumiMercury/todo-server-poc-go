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
	infraAuth "github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth"
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

func (m *MockTaskRepository) FindAllByUserID(ctx context.Context, userID string) ([]*task.Task, error) {
	// Return mock tasks for the specific user
	if userID == "test-user" {
		task1 := task.NewTask("1", "Test Task 1", userID)
		task2 := task.NewTask("2", "Test Task 2", userID)

		return []*task.Task{task1, task2}, nil
	}

	if userID == "other-user" {
		task3 := task.NewTask("3", "Other User Task", userID)

		return []*task.Task{task3}, nil
	}

	return []*task.Task{}, nil
}

func (m *MockTaskRepository) FindById(ctx context.Context, userID, id string) (*task.Task, error) {
	if userID == "test-user" && id == "1" {
		return task.NewTask("1", "Test Task 1", userID), nil
	}

	if userID == "other-user" && id == "3" {
		return task.NewTask("3", "Other User Task", userID), nil
	}

	return nil, task.ErrTaskNotFound
}

func (m *MockTaskRepository) Create(ctx context.Context, userID, title string) (*task.Task, error) {
	return task.NewTask("new-id", title, userID), nil
}

func (m *MockTaskRepository) Update(ctx context.Context, userID, id, title string) (*task.Task, error) {
	if userID == "test-user" && id == "1" {
		return task.NewTask(id, title, userID), nil
	}

	return nil, task.ErrTaskNotFound
}

func (m *MockTaskRepository) Delete(ctx context.Context, userID, id string) error {
	return nil
}

func generateTestJWT() string {
	return generateTestJWTForUser("test-user")
}

func generateTestJWTForUser(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret-key-for-testing"))
	if err != nil {
		panic("Failed to sign test JWT token: " + err.Error())
	}

	return tokenString
}

func setupTestRouter() *echo.Echo {
	router := echo.New()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "test-secret-key-for-testing",
		},
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
	}

	// Add CORS middleware globally
	router.Use(handler.CORSMiddleware(*cfg))

	mockRepo := &MockTaskRepository{}
	taskController := controller.NewTask(mockRepo)
	mockHealthService := &MockHealthService{}
	taskServer := handler.NewTaskServer(*taskController, mockHealthService)

	// Setup authentication service and middleware
	authService, err := infraAuth.NewAuthenticationService(*cfg)
	if err != nil {
		panic(err)
	}

	authMiddleware := handler.NewAuthenticationMiddleware(authService)
	authMiddlewareFunc := authMiddleware.MiddlewareFunc()

	// Create wrapper for generated handlers
	wrapper := generated.ServerInterfaceWrapper{
		Handler: taskServer,
	}

	// Register health endpoint without authentication
	router.GET("/health", wrapper.HealthGetHealth)

	// Create a group for protected task endpoints
	taskGroup := router.Group("/tasks")
	taskGroup.Use(authMiddlewareFunc)

	// Register task endpoints with authentication middleware
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
		req, err := http.NewRequest("GET", "/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("GET /tasks with invalid JWT token should return 401", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

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
		req, err := http.NewRequest("GET", "/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("POST /tasks with valid JWT token should return 201", func(t *testing.T) {
		requestBody := map[string]string{"title": "New Test Task"}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("GET /tasks/1 with valid JWT token should return 200", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/tasks/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("PUT /tasks/1 with valid JWT token should return 200", func(t *testing.T) {
		updateBody := map[string]string{"title": "Updated Test Task"}

		jsonData, err := json.Marshal(updateBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("DELETE /tasks/1 with valid JWT token should return 204", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/tasks/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}

func TestUserTaskSeparation(t *testing.T) {
	router := setupTestRouter()
	testUserJWT := generateTestJWTForUser("test-user")
	otherUserJWT := generateTestJWTForUser("other-user")

	t.Run("Users should only see their own tasks", func(t *testing.T) {
		// Test user should see 2 tasks
		req, err := http.NewRequest("GET", "/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testUserJWT)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var testUserTasks []map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &testUserTasks); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(testUserTasks) != 2 {
			t.Errorf("Expected test-user to have 2 tasks, got %d", len(testUserTasks))
		}

		// Other user should see 1 task
		req, err = http.NewRequest("GET", "/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+otherUserJWT)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var otherUserTasks []map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &otherUserTasks); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(otherUserTasks) != 1 {
			t.Errorf("Expected other-user to have 1 task, got %d", len(otherUserTasks))
		}
	})

	t.Run("Users should not access other users' tasks by ID", func(t *testing.T) {
		// Test user tries to access other user's task (task ID 3)
		req, err := http.NewRequest("GET", "/tasks/3", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testUserJWT)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d when test-user tries to access other user's task, got %d", http.StatusNotFound, w.Code)
		}

		// Other user tries to access test user's task (task ID 1)
		req, err = http.NewRequest("GET", "/tasks/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+otherUserJWT)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d when other-user tries to access test user's task, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Created tasks should belong to the authenticated user", func(t *testing.T) {
		requestBody := map[string]string{"title": "Test User New Task"}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testUserJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		// Verify response doesn't contain userID (as per OpenAPI schema)
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, exists := response["userId"]; exists {
			t.Error("Response should not contain userId field")
		}

		if response["title"] != "Test User New Task" {
			t.Errorf("Expected title 'Test User New Task', got %s", response["title"])
		}
	})
}

func TestCORSOptionsEndpoint(t *testing.T) {
	router := setupTestRouter()

	t.Run("OPTIONS /tasks should return CORS headers without authentication", func(t *testing.T) {
		req, err := http.NewRequest("OPTIONS", "/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

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

func TestTaskTitleValidation(t *testing.T) {
	router := setupTestRouter()
	testJWT := generateTestJWT()

	t.Run("POST /tasks with empty title should return 400", func(t *testing.T) {
		requestBody := map[string]string{"title": ""}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for empty title, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("POST /tasks with whitespace-only title should return 400", func(t *testing.T) {
		requestBody := map[string]string{"title": "   "}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for whitespace-only title, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("POST /tasks with exactly 255 character title should return 201", func(t *testing.T) {
		longTitle := strings.Repeat("a", 255)
		requestBody := map[string]string{"title": longTitle}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d for 255-char title, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("POST /tasks with 256 character title should return 400", func(t *testing.T) {
		longTitle := strings.Repeat("a", 256)
		requestBody := map[string]string{"title": longTitle}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for 256-char title, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("PUT /tasks/1 with empty title should return 400", func(t *testing.T) {
		updateBody := map[string]string{"title": ""}

		jsonData, err := json.Marshal(updateBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for empty title in update, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("PUT /tasks/1 with 256 character title should return 400", func(t *testing.T) {
		longTitle := strings.Repeat("b", 256)
		updateBody := map[string]string{"title": longTitle}

		jsonData, err := json.Marshal(updateBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for 256-char title in update, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("PUT /tasks/1 with exactly 255 character title should return 200", func(t *testing.T) {
		longTitle := strings.Repeat("c", 255)
		updateBody := map[string]string{"title": longTitle}

		jsonData, err := json.Marshal(updateBody)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+testJWT)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d for 255-char title in update, got %d", http.StatusOK, w.Code)
		}
	})
}
