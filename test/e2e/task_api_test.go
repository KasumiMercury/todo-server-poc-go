package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/controller"
	infraAuth "github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler/generated"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/repository"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestServer struct {
	server   *echo.Echo
	baseURL  string
	jwtToken string
	cleanup  func()
}

func setupTestServer(t *testing.T) *TestServer {
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

	router := echo.New()
	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "test-secret-key-for-e2e-testing",
		},
		AllowOrigins: []string{"http://localhost:3000"},
	}

	router.Use(handler.CORSMiddleware(*cfg))

	taskRepo := repository.NewTaskDB(db)
	taskController := controller.NewTask(taskRepo)
	healthService := service.NewHealthService(db) // Use real implementation for E2E
	apiServer := handler.NewAPIServer(*taskController, healthService)

	authService, err := infraAuth.NewAuthenticationService(*cfg)
	require.NoError(t, err)

	authMiddleware := handler.NewAuthenticationMiddleware(authService)
	authMiddlewareFunc := authMiddleware.MiddlewareFunc()

	wrapper := generated.ServerInterfaceWrapper{
		Handler: apiServer,
	}

	router.GET("/health", wrapper.HealthGetHealth)

	taskGroup := router.Group("/tasks")
	taskGroup.Use(authMiddlewareFunc)

	taskGroup.GET("", wrapper.TaskGetAllTasks)
	taskGroup.POST("", wrapper.TaskCreateTask)
	taskGroup.GET("/:taskId", wrapper.TaskGetTask)
	taskGroup.PUT("/:taskId", wrapper.TaskUpdateTask)
	taskGroup.DELETE("/:taskId", wrapper.TaskDeleteTask)

	userID := uuid.New().String()
	jwtToken := generateTestJWTToken(userID, cfg.Auth.JWTSecret)

	return &TestServer{
		server:   router,
		baseURL:  "",
		jwtToken: jwtToken,
		cleanup: func() {
			postgresContainer.Terminate(ctx)
		},
	}
}

func generateTestJWTToken(userID, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic("Failed to sign test JWT token: " + err.Error())
	}

	return tokenString
}

func (ts *TestServer) makeRequest(method, path string, body interface{}, headers map[string]string) (*httptest.ResponseRecorder, error) {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+ts.jwtToken)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	ts.server.ServeHTTP(rec, req)

	return rec, nil
}

func TestE2E_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Parallel()

	// Arrange
	testServer := setupTestServer(t)
	defer testServer.cleanup()

	// Act
	rec, err := testServer.makeRequest("GET", "/health", nil, map[string]string{
		"Authorization": "", // Health endpoint doesn't require auth
	})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var healthResponse generated.HealthStatus

	err = json.Unmarshal(rec.Body.Bytes(), &healthResponse)
	require.NoError(t, err)
	assert.Equal(t, generated.HealthStatusStatusUP, healthResponse.Status)
}

func TestE2E_TaskCRUDOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Parallel()

	// Arrange
	testServer := setupTestServer(t)
	defer testServer.cleanup()

	var createdTaskID string

	t.Run("1. Get all tasks - initially empty", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks", nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var tasks []generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &tasks)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("2. Create a new task", func(t *testing.T) {
		// Arrange
		createRequest := map[string]string{
			"title": "E2E Test Task",
		}

		// Act
		rec, err := testServer.makeRequest("POST", "/tasks", createRequest, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var createdTask generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &createdTask)
		require.NoError(t, err)
		assert.NotEmpty(t, createdTask.Id)
		assert.Equal(t, "E2E Test Task", createdTask.Title)

		createdTaskID = createdTask.Id
	})

	t.Run("3. Get all tasks - should contain one task", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks", nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var tasks []generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &tasks)
		require.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, createdTaskID, tasks[0].Id)
		assert.Equal(t, "E2E Test Task", tasks[0].Title)
	})

	t.Run("4. Get specific task by ID", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks/"+createdTaskID, nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var task generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &task)
		require.NoError(t, err)
		assert.Equal(t, createdTaskID, task.Id)
		assert.Equal(t, "E2E Test Task", task.Title)
	})

	t.Run("5. Update task", func(t *testing.T) {
		// Arrange
		updateRequest := map[string]string{
			"title": "Updated E2E Test Task",
		}

		// Act
		rec, err := testServer.makeRequest("PUT", "/tasks/"+createdTaskID, updateRequest, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var updatedTask generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &updatedTask)
		require.NoError(t, err)
		assert.Equal(t, createdTaskID, updatedTask.Id)
		assert.Equal(t, "Updated E2E Test Task", updatedTask.Title)
	})

	t.Run("6. Verify update persisted", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks/"+createdTaskID, nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var task generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &task)
		require.NoError(t, err)
		assert.Equal(t, "Updated E2E Test Task", task.Title)
	})

	t.Run("7. Delete task", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("DELETE", "/tasks/"+createdTaskID, nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		responseBody := rec.Body.String()
		assert.True(t, responseBody == "" || responseBody == "null\n", "Expected empty or null response, got: %s", responseBody)
	})

	t.Run("8. Verify task was deleted", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks/"+createdTaskID, nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("9. Get all tasks - should be empty again", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks", nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var tasks []generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &tasks)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
}

func TestE2E_TaskValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Parallel()

	// Arrange
	testServer := setupTestServer(t)
	defer testServer.cleanup()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		expectedTitle  string
	}{
		{
			name:           "empty title should return 400",
			requestBody:    map[string]string{"title": ""},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title field is required",
			expectedTitle:  "",
		},
		{
			name:           "missing title should return 400",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title field is required",
			expectedTitle:  "",
		},
		{
			name:           "whitespace-only title should return 400",
			requestBody:    map[string]string{"title": "   "},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot be empty",
			expectedTitle:  "",
		},
		{
			name:           "title too long should return 400",
			requestBody:    map[string]string{"title": strings.Repeat("a", 256)},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot exceed 255 characters",
			expectedTitle:  "",
		},
		{
			name:           "valid title should return 201",
			requestBody:    map[string]string{"title": "Valid Task Title"},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			expectedTitle:  "Valid Task Title",
		},
		{
			name:           "chinese characters exactly max length should return 201",
			requestBody:    map[string]string{"title": strings.Repeat("你", 255)},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			expectedTitle:  strings.Repeat("你", 255),
		},
		{
			name:           "chinese characters over max length should return 400",
			requestBody:    map[string]string{"title": strings.Repeat("你", 256)},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot exceed 255 characters",
			expectedTitle:  "",
		},
		{
			name:           "korean characters exactly max length should return 201",
			requestBody:    map[string]string{"title": strings.Repeat("안", 255)},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			expectedTitle:  strings.Repeat("안", 255),
		},
		{
			name:           "korean characters over max length should return 400",
			requestBody:    map[string]string{"title": strings.Repeat("안", 256)},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot exceed 255 characters",
			expectedTitle:  "",
		},
		{
			name:           "zero-width joiner emoji within boundary should return 201",
			requestBody:    map[string]string{"title": strings.Repeat("👨‍💻", 85)}, // 85 * 3 = 255 runes
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			expectedTitle:  strings.Repeat("👨‍💻", 85),
		},
		{
			name:           "zero-width joiner emoji over boundary should return 400",
			requestBody:    map[string]string{"title": strings.Repeat("👨‍💻", 86)}, // 86 * 3 = 258 runes
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot exceed 255 characters",
			expectedTitle:  "",
		},
		{
			name:           "family emoji within boundary should return 201",
			requestBody:    map[string]string{"title": strings.Repeat("👨‍👩‍👧‍👦", 36)}, // 36 * 7 = 252 runes
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			expectedTitle:  strings.Repeat("👨‍👩‍👧‍👦", 36),
		},
		{
			name:           "family emoji over boundary should return 400",
			requestBody:    map[string]string{"title": strings.Repeat("👨‍👩‍👧‍👦", 37)}, // 37 * 7 = 259 runes
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot exceed 255 characters",
			expectedTitle:  "",
		},
		{
			name:           "unicode non-breaking space only should return 400",
			requestBody:    map[string]string{"title": "\u00A0\u00A0\u00A0"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot be empty",
			expectedTitle:  "",
		},
		{
			name:           "ideographic space only should return 400",
			requestBody:    map[string]string{"title": "　　　"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot be empty",
			expectedTitle:  "",
		},
		{
			name:           "zero-width space only should return 400",
			requestBody:    map[string]string{"title": "​​​"}, // Zero-width spaces are now trimmed
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot be empty",
			expectedTitle:  "",
		},
		{
			name:           "mixed unicode whitespace only should return 400",
			requestBody:    map[string]string{"title": " ​ ‌ ‍   　"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot be empty",
			expectedTitle:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			rec, err := testServer.makeRequest("POST", "/tasks", tt.requestBody, nil)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				responseBody := strings.ToLower(rec.Body.String())
				assert.Contains(t, responseBody, strings.ToLower(tt.expectedError))
			}

			if tt.expectedStatus == http.StatusCreated {
				var task generated.Task

				err := json.Unmarshal(rec.Body.Bytes(), &task)
				require.NoError(t, err)
				assert.NotEmpty(t, task.Id)
				assert.Equal(t, tt.expectedTitle, task.Title)
			}
		})
	}
}

func TestE2E_AuthenticationRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Parallel()

	// Arrange
	testServer := setupTestServer(t)
	defer testServer.cleanup()

	endpoints := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"GET", "/tasks", nil},
		{"POST", "/tasks", map[string]string{"title": "Test Task"}},
		{"GET", "/tasks/123", nil},
		{"PUT", "/tasks/123", map[string]string{"title": "Updated Task"}},
		{"DELETE", "/tasks/123", nil},
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("%s %s without auth should return 401", endpoint.method, endpoint.path), func(t *testing.T) {
			// Act
			rec, err := testServer.makeRequest(endpoint.method, endpoint.path, endpoint.body, map[string]string{
				"Authorization": "", // Remove auth header
			})

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run(fmt.Sprintf("%s %s with invalid token should return 401", endpoint.method, endpoint.path), func(t *testing.T) {
			// Act
			rec, err := testServer.makeRequest(endpoint.method, endpoint.path, endpoint.body, map[string]string{
				"Authorization": "Bearer invalid-token",
			})

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func TestE2E_TaskNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Parallel()

	// Arrange
	testServer := setupTestServer(t)
	defer testServer.cleanup()

	nonexistentTaskID := uuid.New().String()

	tests := []struct {
		name     string
		method   string
		path     string
		body     interface{}
		expected int
	}{
		{
			name:     "GET nonexistent task should return 404",
			method:   "GET",
			path:     "/tasks/" + nonexistentTaskID,
			body:     nil,
			expected: http.StatusNotFound,
		},
		{
			name:     "PUT nonexistent task should return 404",
			method:   "PUT",
			path:     "/tasks/" + nonexistentTaskID,
			body:     map[string]string{"title": "Updated Title"},
			expected: http.StatusNotFound,
		},
		{
			name:     "DELETE nonexistent task should return 204",
			method:   "DELETE",
			path:     "/tasks/" + nonexistentTaskID,
			body:     nil,
			expected: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			rec, err := testServer.makeRequest(tt.method, tt.path, tt.body, nil)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expected, rec.Code)
		})
	}
}

func TestE2E_MultipleTasksManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Parallel()

	// Arrange
	testServer := setupTestServer(t)
	defer testServer.cleanup()

	taskTitles := []string{"Task 1", "Task 2", "Task 3", "Task 4", "Task 5"}
	createdTasks := make([]generated.Task, 0, len(taskTitles))

	for _, title := range taskTitles {
		createRequest := map[string]string{"title": title}
		rec, err := testServer.makeRequest("POST", "/tasks", createRequest, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var task generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &task)
		require.NoError(t, err)

		createdTasks = append(createdTasks, task)
	}

	t.Run("verify all tasks are created", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks", nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var tasks []generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &tasks)
		require.NoError(t, err)
		assert.Len(t, tasks, len(taskTitles))

		// Check that all created tasks are present
		taskIDs := make(map[string]bool)
		for _, task := range tasks {
			taskIDs[task.Id] = true
		}

		for _, createdTask := range createdTasks {
			assert.True(t, taskIDs[createdTask.Id], "Task %s should be present", createdTask.Id)
		}
	})

	t.Run("update specific task", func(t *testing.T) {
		// Update the middle task
		taskToUpdate := createdTasks[2]
		updateRequest := map[string]string{"title": "Updated " + taskToUpdate.Title}

		// Act
		rec, err := testServer.makeRequest("PUT", "/tasks/"+taskToUpdate.Id, updateRequest, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var updatedTask generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &updatedTask)
		require.NoError(t, err)
		assert.Equal(t, "Updated "+taskToUpdate.Title, updatedTask.Title)
	})

	t.Run("delete specific task", func(t *testing.T) {
		// Delete the first task
		taskToDelete := createdTasks[0]

		// Act
		rec, err := testServer.makeRequest("DELETE", "/tasks/"+taskToDelete.Id, nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify it's gone
		rec, err = testServer.makeRequest("GET", "/tasks/"+taskToDelete.Id, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("verify remaining tasks", func(t *testing.T) {
		// Act
		rec, err := testServer.makeRequest("GET", "/tasks", nil, nil)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var tasks []generated.Task

		err = json.Unmarshal(rec.Body.Bytes(), &tasks)
		require.NoError(t, err)
		assert.Len(t, tasks, len(taskTitles)-1) // One task was deleted
	})
}
