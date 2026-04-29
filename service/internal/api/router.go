package api

import (
	"net/http"

	"github.com/paragpratim/gcp-task-orchestration/service/internal/api/handlers"
)

// NewRouter sets up the HTTP routes for the service.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Register handlers
	mux.HandleFunc("/health", handlers.HealthHandler)
	
	return mux
}
