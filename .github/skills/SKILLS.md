# AI Assistant Skills and Coding Standards

This file serves as a guide for any AI assistant or developer working on the `gcp-task-orchestration` repository. When adding new features or modifying existing code, adhere strictly to these architectural rules and patterns.

## 1. Golang Project Layout & Feature Packages

The Golang service is located in the `orchestrator-service/` directory and follows a strict **Package-by-Feature (Domain-Driven) Clean Architecture**, powered by the **Gin HTTP Framework**.

- **`cmd/api/main.go`**: This is the application entry point. It is strictly limited to reading configuration environment variables, initializing the flat infrastructure platform wrapper exactly once, passing it to the router, and handling graceful OS termination signals (`SIGINT`/`SIGTERM`) within a concurrent background listener thread. No business logic or manual route wiring belongs here.
- **`internal/gcp/`**: A single, flat infrastructure package containing our raw cloud provider SDK drivers (`firestore.go`, `cloudtasks.go`, `storage.go`, `bigquery.go`). 
  - All structs must use explicit resource naming instead of generic terms (e.g., `FirestoreRepository[T]`, `CloudTasksRepository`).
  - To prevent memory and connection pool leaks, these MUST be instantiated once at boot time as singletons via `gcp.NewPlatform` and closed cleanly via `infra.Close()` during teardown.
- **`internal/core/`**: Contains our business domain modules (`intake`, `gcs`, `bigquery`, `dataflow`). Each feature is completely encapsulated within its own dedicated directory.
  - **`service.go`**: Contains pure business logic. It interacts *only* with infrastructure repositories passed into its constructor via dependency injection. It must remain completely decoupled from HTTP/Gin context signatures, instead accepting standard `context.Context` parameters.
  - **`handlers.go`**: Contains the HTTP handler structs and methods. Handlers are strictly focused on request validation (`c.ShouldBindJSON`), executing the underlying business service, and returning structured serialization frames (`c.JSON`). Each handler package must expose a `RegisterRoutes(rg *gin.RouterGroup)` method to self-register its endpoint tree paths.
- **`internal/api/router.go`**: The central factory gateway. It receives the `*gcp.Platform` infrastructure singleton, instantiates the business domain services and handlers sequentially in one location, and mounts them to their respective API version groups.

## 2. Dependency Injection & Symmetrical Naming

To ensure uniform code and high testability, all packages must follow a symmetrical interface abstraction pattern:

1. **Define Behavioral Interfaces:** Services must define the infrastructure behaviors they need using localized interfaces (e.g., `ObjectRepository` inside `internal/core/gcs/service.go`) rather than importing raw clients.
2. **Constructor Injection:** Always pass concrete dependency structs or interfaces straight into constructor functions (`NewService`, `NewHandler`). Avoid using global states, hidden initialization steps, or implicit service factories.

## 3. API Documentation Standard (Swagger via Gin)

We use `swaggo/swag` with Gin middleware wrappers to automatically generate OpenAPI/Swagger documentation from declarative comments in the code.

### Writing Handlers
When creating or updating a route handler method inside a feature's `handlers.go`, you MUST write declarative Swagger comments directly above the method block. Do not write manual HTTP verb verification (e.g., checking for `POST`), as Gin's router group configuration enforces routing paths natively.

**Example Pattern:**
```go
// HandleStart handles the initial workflow ingestion.
// @Summary      Start Ingestion Workflow
// @Description  Accepts the onboarding request metadata and schedules async task pipelines.
// @Tags         intake
// @Accept       json
// @Produce      json
// @Param        request  body      intake.IntakeRequest  true  "Intake Request Payload"
// @Success      202      {object}  map[string]string "Status Accepted"
// @Failure      400      {object}  map[string]string "Bad Request"
// @Router       /api/v1/intake/start [post]
func (h *Handler) HandleStart(c *gin.Context) {
    var req IntakeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Execute business service logic...
    
    c.JSON(http.StatusAccepted, gin.H{"status": "INTAKE_ACCEPTED"})
}
```

### Regenerating Documentation
Whenever an endpoint path is added or a Swagger comment is modified, you MUST regenerate the documentation tracking files. Navigate to the root directory and execute:

```bash
swag init -g cmd/api/main.go -o internal/api
```
This forces updates to compile into your internal `internal/api/` layout, updating the live interactive UI (`http://localhost:8080/swagger/index.html`).

## 4. Concurrent Lifecycles & Graceful Shutdowns

Every developer or AI agent must preserve the non-blocking execution lifecycle inside `main.go`:
- The HTTP server listener loop must execute inside a concurrent background goroutine via `go func() { srv.ListenAndServe() }()`.
- Intentional server closures must safely bypass error alerts by verifying errors via `!errors.Is(err, http.ErrServerClosed)`.
- System termination intercepts (`SIGINT`/`SIGTERM`) must block on a channel signal to capture scaling and rollout warnings from Cloud Run/Kubernetes, allowing a standard `context.WithTimeout` buffer window to process leftover in-flight API tasks safely.

## 5. Dependency Management

- After adding any new cloud imports or modifying external files, always execute `go mod tidy` in the service directory.
- Critical dependencies include:
  - Framework: `github.com/gin-gonic/gin`
  - Documentation Engine: `github.com/swaggo/gin-swagger` and `github.com/swaggo/files`
  - Generation Core: `github.com/swaggo/swag`
- The platform target operates on Go 1.25+. Ensure your multi-stage Dockerfile and localized runtimes match the Go compiler version declared inside `go.mod`.
