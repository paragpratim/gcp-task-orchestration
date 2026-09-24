package config

import (
	"fmt"
	"os"
)

type AppConfig struct {
	// Application configuration
	AppName     string `json:"app_name"`
	Environment string `json:"environment"`
	Port        string `json:"port"`
	// GCP Project configuration
	ProjectID           string `json:"project_id"`
	BaseURL             string `json:"base_url"`
	ServiceAccountEmail string `json:"service_account_email"`
	// Firestore configuration
	JobsCollection      string `json:"jobs_collection"`
	JobStatusCollection string `json:"job_status_collection"`
	// Task queue configuration
	IntakeQueueName string `json:"intake_queue_name"`
	GcsQueueName    string `json:"gcs_queue_name"`
	BqQueueName     string `json:"bq_queue_name"`
}

func LoadConfig() *AppConfig {
	env := getEnv("APP_ENV", "local")

	return &AppConfig{
		AppName:             fmt.Sprintf("%s_task_orchestrator", env),
		Environment:         env,
		Port:                getEnv("PORT", "8080"),
		ProjectID:           getEnv("GCP_PROJECT_ID", "local-project"),
		BaseURL:             getEnv("QUEUE_BASE_URL", "localhost:8080"),
		ServiceAccountEmail: getEnv("QUEUE_SERVICE_ACCOUNT", "local-service-account"),
		JobsCollection:      fmt.Sprintf("%s_intake_jobs", env),
		JobStatusCollection: fmt.Sprintf("%s_intake_job_status", env),
		IntakeQueueName:     fmt.Sprintf("orchestrator-intake-queue-%s", env),
		GcsQueueName:        fmt.Sprintf("orchestrator-gcs-queue-%s", env),
		BqQueueName:         fmt.Sprintf("orchestrator-bq-queue-%s", env),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
