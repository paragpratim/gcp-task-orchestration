# AI Assistant Skills and Coding Standards

This file serves as a guide for any AI assistant or developer working on the `gcp-task-orchestration` repository. When adding new features or modifying existing code, adhere to the following standards.

## 1. Golang Project Layout

The Golang service is located in the `orchestrator-service/` directory and strictly follows the Standard Go Project Layout, powered by the **Gin HTTP Framework**.

- **`cmd/server/main.go`**: This is the entrypoint. It should ONLY be responsible for wiring up dependencies, reading configuration/environment variables, and starting the HTTP server using `router.Run()`. Do not put business logic here.
- **`internal/`**: All core application logic MUST reside within the `internal/` directory. This ensures that the code cannot be imported by external applications.
  - **`internal/api/router.go`**: Central location for registering HTTP routes. Returns a `*gin.Engine`.
  - **`internal/api/handlers/`**: Contains the HTTP handler functions. Handlers MUST use Gin's context signature (`func MyHandler(c *gin.Context)`). Each handler should be focused on request parsing (`c.ShouldBindJSON`), calling business logic, and returning responses (`c.JSON` / `c.String`).
  - **`internal/core/`**: Use this for business logic, services, and models completely agnostic to HTTP (e.g., interacting with GCP Dataflow APIs). Pass `c.Request.Context()` down to these services if a standard library `context.Context` context window is required.

## 2. API Documentation Standard (Swagger via Gin)

We use `swaggo/swag` with Gin middleware wrappers to automatically generate OpenAPI/Swagger documentation from declarative comments in the code.

### Writing Handlers
When creating a new HTTP handler function in `internal/api/handlers/`, you MUST write declarative Swagger comments directly above the function signature. Do not write manual HTTP verb verification (e.g. `r.Method != http.MethodPost`) inside handlers, as Gin's router enforces routing rules automatically.

**Example:**
```go
// CreateTaskHandler creates a new orchestration task.
// @Summary      Create a task
// @Description  Accepts a task payload and schedules it for execution.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body      models.TaskRequest  true  "Task Payload"
// @Success      201   {object}  models.TaskResponse
// @Failure      400   {string}  string "Bad Request"
// @Router       /tasks [post]
func CreateTaskHandler(c *gin.Context) {
    // 1. Parse and bind payload
    var req models.TaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.String(http.StatusBadRequest, "Invalid JSON payload")
        return
    }
    
    // 2. Business logic execution happens here...
    
    // 3. Return JSON response safely
    c.JSON(http.StatusCreated, models.TaskResponse{Message: "Success"})
}
```

### Regenerating Documentation
Whenever you add a new endpoint or modify an existing Swagger comment, you MUST regenerate the documentation files.

Navigate to the `orchestrator-service/` directory and run:
```bash
swag init -g cmd/server/main.go
```
This updates the `orchestrator-service/docs/` directory. If the generated files are not updated, the Swagger UI (`http://localhost:8080/swagger/index.html`) will not reflect the latest changes.

## 3. Dependency Management

- After adding any new imports, always run `go mod tidy` in the `orchestrator-service/` directory.
- Major dependencies include:
  - Framework: `github.com/gin-gonic/gin v1.12.0`
  - Documentation UI: `github.com/swaggo/gin-swagger v1.6.1` and `github.com/swaggo/files v1.0.1`
  - Code Generation Tool: `github.com/swaggo/swag v1.16.6`
- The project runs on Go 1.25+. Ensure the Dockerfile builder stage matches the Go version used in `go.mod`.

## 4. Docker & Local Testing

- The project provides a `docker-compose.yml` in the root. 
- To test changes locally, run `docker compose up --build`. This rebuilds the Go service and serves it on port `8080`.
