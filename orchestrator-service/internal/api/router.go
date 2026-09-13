package api

import (
	"net/http"

	_ "github.com/paragpratim/gcp-task-orchestration/orchestrator-service/docs" // Import generated docs
	"github.com/paragpratim/gcp-task-orchestration/orchestrator-service/internal/api/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter sets up the HTTP routes for the service.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/dataflow/flex/submit", handlers.SubmitDataflowJobHandler)

	// Register Swagger UI handler
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	return mux
}
