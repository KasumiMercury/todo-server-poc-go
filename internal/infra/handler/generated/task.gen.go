// Package generated provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.5.0 DO NOT EDIT.
package generated

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Defines values for HealthComponentStatus.
const (
	HealthComponentStatusDOWN HealthComponentStatus = "DOWN"
	HealthComponentStatusUP   HealthComponentStatus = "UP"
)

// Defines values for HealthStatusStatus.
const (
	HealthStatusStatusDOWN HealthStatusStatus = "DOWN"
	HealthStatusStatusUP   HealthStatusStatus = "UP"
)

// Error defines model for error.
type Error struct {
	// Code Error code
	Code int `json:"code"`

	// Details Additional error details
	Details *string `json:"details,omitempty"`

	// Message Error message
	Message string `json:"message"`
}

// HealthComponent defines model for healthComponent.
type HealthComponent struct {
	// Details Additional component details
	Details *map[string]interface{} `json:"details,omitempty"`

	// Status Component health status
	Status HealthComponentStatus `json:"status"`
}

// HealthComponentStatus Component health status
type HealthComponentStatus string

// HealthStatus defines model for healthStatus.
type HealthStatus struct {
	// Components Health status of individual components
	Components struct {
		Database *HealthComponent `json:"database,omitempty"`
	} `json:"components"`

	// Status Overall application health status
	Status HealthStatusStatus `json:"status"`

	// Timestamp Timestamp of the health check
	Timestamp time.Time `json:"timestamp"`
}

// HealthStatusStatus Overall application health status
type HealthStatusStatus string

// Task defines model for task.
type Task struct {
	// Id The unique identifier for the task
	Id string `json:"id"`

	// Title The title of the task
	Title string `json:"title"`
}

// TaskCreate defines model for taskCreate.
type TaskCreate struct {
	// Title The title of the task
	Title string `json:"title"`
}

// TaskUpdate defines model for taskUpdate.
type TaskUpdate struct {
	// Title The title of the task
	Title *string `json:"title,omitempty"`
}

// TaskCreateTaskJSONRequestBody defines body for TaskCreateTask for application/json ContentType.
type TaskCreateTaskJSONRequestBody = TaskCreate

// TaskUpdateTaskJSONRequestBody defines body for TaskUpdateTask for application/json ContentType.
type TaskUpdateTaskJSONRequestBody = TaskUpdate

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get application health status
	// (GET /health)
	HealthGetHealth(ctx echo.Context) error
	// Get all tasks
	// (GET /tasks)
	TaskGetAllTasks(ctx echo.Context) error
	// Create a new task
	// (POST /tasks)
	TaskCreateTask(ctx echo.Context) error
	// Delete a task
	// (DELETE /tasks/{taskId})
	TaskDeleteTask(ctx echo.Context, taskId string) error
	// Get a task
	// (GET /tasks/{taskId})
	TaskGetTask(ctx echo.Context, taskId string) error
	// Update a task
	// (PUT /tasks/{taskId})
	TaskUpdateTask(ctx echo.Context, taskId string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// HealthGetHealth converts echo context to params.
func (w *ServerInterfaceWrapper) HealthGetHealth(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.HealthGetHealth(ctx)
	return err
}

// TaskGetAllTasks converts echo context to params.
func (w *ServerInterfaceWrapper) TaskGetAllTasks(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{})

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.TaskGetAllTasks(ctx)
	return err
}

// TaskCreateTask converts echo context to params.
func (w *ServerInterfaceWrapper) TaskCreateTask(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{})

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.TaskCreateTask(ctx)
	return err
}

// TaskDeleteTask converts echo context to params.
func (w *ServerInterfaceWrapper) TaskDeleteTask(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "taskId" -------------
	var taskId string

	err = runtime.BindStyledParameterWithOptions("simple", "taskId", ctx.Param("taskId"), &taskId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter taskId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{})

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.TaskDeleteTask(ctx, taskId)
	return err
}

// TaskGetTask converts echo context to params.
func (w *ServerInterfaceWrapper) TaskGetTask(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "taskId" -------------
	var taskId string

	err = runtime.BindStyledParameterWithOptions("simple", "taskId", ctx.Param("taskId"), &taskId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter taskId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{})

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.TaskGetTask(ctx, taskId)
	return err
}

// TaskUpdateTask converts echo context to params.
func (w *ServerInterfaceWrapper) TaskUpdateTask(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "taskId" -------------
	var taskId string

	err = runtime.BindStyledParameterWithOptions("simple", "taskId", ctx.Param("taskId"), &taskId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter taskId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{})

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.TaskUpdateTask(ctx, taskId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/health", wrapper.HealthGetHealth)
	router.GET(baseURL+"/tasks", wrapper.TaskGetAllTasks)
	router.POST(baseURL+"/tasks", wrapper.TaskCreateTask)
	router.DELETE(baseURL+"/tasks/:taskId", wrapper.TaskDeleteTask)
	router.GET(baseURL+"/tasks/:taskId", wrapper.TaskGetTask)
	router.PUT(baseURL+"/tasks/:taskId", wrapper.TaskUpdateTask)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xYUVPjNhD+K5ptH33EgWTmxm8ctJQO06MlzM2UyYOwN4kOWzaSnF7K+L93VpITJ3Yg",
	"MOFopzzFsZTVp28/fbvKA8R5VuQSpdEQPYCOZ5hx+4hK5YoeCpUXqIxA+zrOE6TPBHWsRGFELiGCn2gy",
	"s2MB4DeeFSlCNAjDAMyiQIhASINTVFAFkKDhItXtKMdJIuiRp8yuzuqZjZhwLuc8FQkTsigNK7jiGRpU",
	"NMkvpY0SckorZag1n27FWw83w3/iCfsD70vUph2xCkDhfSkUJhDdgN9vHWa8nJ/ffsXYEIIZ8tTMTmqS",
	"23zuQsYyR12EUEqkxNj/7DLXZqrw6vcLILC6yKXGkcgI1jDThKkFUhtuyg4IS9jMbYP5iQGgLDNi4PoS",
	"Ajj9/OU32vyKRPv6ce58qO2cXS1BbQqwKdh1wL80YbJ8woRMxFwkZZNEwr+RA274LdeWzB8VTiCCH3qr",
	"+T1/KnqbyayeQ+bnOSqepowXRSpiTm/3QWsARmSoDc+K9pqjeoi4MDOsF4xnGN+tyf4wPBx8CPsf+sNR",
	"P4yOwigM/4QAJrnKuIGIKMIPtNKuiW3iCppJ68q44fqunWmRdOxohqyU4r5EJhKURkwEKjbJld2ejdPc",
	"Vv/waDDsJs2k2B3eDtWEtSJe2Qc2cu8z/u0C5dTMIDocDp+iRiRQr7yNhROF3GCbizfB+zjU6yLZL1QX",
	"MNkd6wYsOnwYl0qYxRWdWAfnFrlCdVxSnPrbz7Wsf/0ygsBVPYrkRld6mRlTQEWBhZzkHXvi+o5doZqj",
	"YseX58vsdo3MUWn3q/5BeBASi3mBkhcCIjg6CA+OyJe4mVnU3mvocYq2bBDF1jXOk6XTnaFxDw2vtz8/",
	"DENfGYyvOg3b6X3VhONhvYasuWrDEBsF6rmFprZCMq5q4/uab3XbT+Uzw3fzZV8wbLo2SmnDcoX2Jrig",
	"+MPn80Ttz5AamyUxddmxrso0qrmIyaX4nIuU36bNHoH6F4OKqrpXh21Fdt+ra8s6NrmMq11crOMOw6NX",
	"0IJvD+FkqQmmcFJqTNYyb4tY1Xrzhtkv5TL/DcOA6GYcgC6zjKsFRHCG5tFKbfhUk0X6YzqmWD1yNL31",
	"yJIlnKE5TtORnfeiI7uiRBjM9FPcWI9dOSVXii+6OLoQ2lhftsiqAAZh/1loXiTZa8lLM8uV+NupZvhM",
	"CvZ5Tta1sF42bsZVWxxp6tlaicGyPa4CKHK9RQCuvvsKp9w941OeLPa270YTUa3Xc6NKrFqi6+915S7C",
	"bSmMLaSE6TKOUetJmaYLp7PvkHK603mu37X9hLaddhhnEv+qW7QNfS+trvdAH+dJ5WpDiq4dbKv+1I55",
	"1Teu7ASm3SWenzZbRGZy5mNTGwaRbZIgAMltr+EgwKbSm9Vjs3Mct07BYEtv5xbuku0bSGjgYL7uonbb",
	"MjdskpfyP6VcJzLGt6g2eLQqv1ybCo0SOH9NdYav7tHHbjurC967wP+lbcdWdRflFnW7a/XLBV66e/5e",
	"5f06fY//R2Knvif8Pn1P6f/T+N/2Pe9n+tEz7SS79Vi7WBS867xe5DFPWYJzTPMiQ2k8EAigVKn//yrq",
	"9VKaN8u1iT6GH0OoxtU/AQAA//+yzzY3+BkAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
