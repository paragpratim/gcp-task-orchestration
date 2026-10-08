package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AppConfig holds the configuration for the application.
type AppConfig struct {
	// Application configuration
	AppName     string `json:"app_name"`
	Environment string `json:"environment"`
	Port        string `json:"port"`
	// GCP Project configuration
	ProjectID           string `json:"project_id"`
	Region              string `json:"region"`
	FirestoreDB         string `json:"firestore_db"`
	BaseURL             string `json:"base_url"`
	ServiceAccountEmail string `json:"service_account_email"`
	// Firestore configuration
	JobsCollection         string `json:"jobs_collection"`
	JobStatusCollection    string `json:"job_status_collection"`
	JobStatusLogCollection string `json:"job_status_log_collection"`
	// Task queue configuration
	AdminQueueName string `json:"admin_queue_name"`
	GcsQueueName   string `json:"gcs_queue_name"`
	BqQueueName    string `json:"bq_queue_name"`
	// Time intervals configuration
	TaskFrequency           time.Duration `json:"task_frequency"`
	StatusCheckerFrequency  time.Duration `json:"status_checker_frequency"`
	BQLoadJobCheckFrequency time.Duration `json:"bq_load_job_check_frequency"`
}

// LoadConfig loads the application configuration from environment variables.
func LoadConfig() *AppConfig {
	env := getEnv("APP_ENV", "local")

	return &AppConfig{
		AppName:                fmt.Sprintf("%s_task_orchestrator", env),
		Environment:            env,
		Port:                   getEnv("PORT", "8080"),
		ProjectID:              getEnv("GCP_PROJECT_ID", "local-project"),
		Region:                 getEnv("GCP_REGION", "europe-west1"),
		FirestoreDB:            getEnv("FIRESTORE_DB", "(default)"),
		BaseURL:                getEnv("QUEUE_BASE_URL", "http://orchestrator:8080"),
		ServiceAccountEmail:    getEnv("QUEUE_SERVICE_ACCOUNT", "local-service-account"),
		JobsCollection:         fmt.Sprintf("%s_jobs", env),
		JobStatusCollection:    fmt.Sprintf("%s_job_status", env),
		JobStatusLogCollection: fmt.Sprintf("%s_job_status_log", env),
		AdminQueueName:         fmt.Sprintf("orchestrator-admin-queue-%s", env),
		GcsQueueName:           fmt.Sprintf("orchestrator-gcs-queue-%s", env),
		BqQueueName:            fmt.Sprintf("orchestrator-bq-queue-%s", env),
		// Time intervals - configurable via environment variables, with sensible defaults
		TaskFrequency:           getDurationEnv("TASK_FREQUENCY", 15*time.Minute),
		StatusCheckerFrequency:  getDurationEnv("STATUS_CHECKER_FREQUENCY", 10*time.Second),
		BQLoadJobCheckFrequency: getDurationEnv("BQ_LOAD_JOB_CHECK_FREQUENCY", 1*time.Minute),
	}
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, it returns the fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getDurationEnv retrieves a time.Duration from an environment variable.
// The value can be a string like "15m", "10s", "1h" (Go duration format),
// or a plain integer representing seconds. If the variable is not present
// or invalid, it returns the fallback value.
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		// Try parsing as a Go duration string first (e.g., "15m", "10s", "1h")
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		// Fall back to parsing as an integer representing seconds
		if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return fallback
}
