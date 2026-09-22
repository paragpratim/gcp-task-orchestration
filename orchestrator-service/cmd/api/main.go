package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"orchestrator/internal/api"
	"orchestrator/internal/config"
	"orchestrator/internal/gcp"
	"orchestrator/internal/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title           GCP Task Orchestration API
// @version         1.0
// @description     This is a server for the GCP Task Orchestration api.

// @BasePath 		/api/v1
func main() {
	// Initialize the structured logger
	logger.Init()

	// Load application configuration from environment variables
	appCfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if appCfg.ProjectID == "" || appCfg.BaseURL == "" || appCfg.ServiceAccountEmail == "" {
		logger.Warn("WARNING: Essential environment variables (GCP_PROJECT_ID, QUEUE_BASE_URL, QUEUE_SERVICE_ACCOUNT) are missing. Infrastructure may fail to authenticate.")
	}

	// 1. Initialize complete infrastructure layer in one step
	infra, err := gcp.NewPlatform(ctx, appCfg.ProjectID, appCfg.BaseURL, appCfg.ServiceAccountEmail)
	if err != nil {
		logger.Fatal("Critical platform initialization failure: %v", err)
	}
	defer infra.Close() // Automatically clean up everything on exit

	// 2. Set up and spin up HTTP network server
	srv := &http.Server{
		Addr:         ":" + appCfg.Port,
		Handler:      api.SetupRouter(infra, appCfg),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Orchestrator API spinning up on port %s", appCfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server crash: %v", err)
		}
	}()

	// 3. Clean, blocking OS signal termination handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down engine gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Forced shutdown executed: %v", err)
	}
}
