package main

import (
	"net/http"
	"os"

	"github.com/paragpratim/gcp-task-orchestration/service/internal/api"
	"github.com/paragpratim/gcp-task-orchestration/service/internal/logger"
)

// @title           GCP Task Orchestration API
// @version         1.0
// @description     This is a server for the GCP Task Orchestration service.

// @host      localhost:8080
// @BasePath  /
func main() {
	// Initialize the structured logger
	logger.Init()

	// Initialize the router
	router := api.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		logger.Info("defaulting to port", "port", port)
	}

	logger.Info("starting server", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
