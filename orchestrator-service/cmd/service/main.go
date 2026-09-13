package main

import (
	"orchestrator/internal/api"
	"orchestrator/internal/logger"
	"os"
)

// @title           GCP Task Orchestration API
// @version         1.0
// @description     This is a server for the GCP Task Orchestration service.

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
	if err := router.Run(":" + port); err != nil {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
