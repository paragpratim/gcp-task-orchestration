package main

import (
	"log"
	"net/http"
	"os"

	"github.com/paragpratim/gcp-task-orchestration/service/internal/api"
)

// @title           GCP Task Orchestration API
// @version         1.0
// @description     This is a server for the GCP Task Orchestration service.

// @host      localhost:8080
// @BasePath  /
func main() {
	// Initialize the router
	router := api.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
