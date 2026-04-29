package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/paragpratim/gcp-task-orchestration/service/docs" // Import generated docs
	"github.com/paragpratim/gcp-task-orchestration/service/internal/api/handlers"
)

// NewRouter sets up the HTTP routes for the service.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Register handlers
	mux.HandleFunc("/health", handlers.HealthHandler)
	
	// Register Swagger UI handler
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	
	return mux
}
