package models

import "time"

type ExecutionStatus string

const (
	StatusQueued             ExecutionStatus = "QUEUED"
	StatusProcessingGCS      ExecutionStatus = "PROCESSING_GCS"
	StatusCompletedGCS       ExecutionStatus = "COMPLETED_GCS"
	StatusProcessingBigQuery ExecutionStatus = "PROCESSING_BIGQUERY"
	StatusCompletedBigQuery  ExecutionStatus = "COMPLETED_BIGQUERY"
	StatusProcessingDataflow ExecutionStatus = "PROCESSING_DATAFLOW"
	StatusSuccess            ExecutionStatus = "SUCCESS"
	StatusFailed             ExecutionStatus = "FAILED"
	StatusSkipped            ExecutionStatus = "COMPLETED_SKIPPED"
)

func (e ExecutionStatus) String() string {
	return string(e)
}

type JobStatus struct {
	ExecutionID string          `json:"execution_id" firestore:"execution_id"` // Matches your service architecture
	JobID       string          `json:"job_id" firestore:"job_id"`
	Status      ExecutionStatus `json:"status" firestore:"status"` // Replaced raw string with Enum type
	Message     string          `json:"message" firestore:"message"`
	CreatedAt   time.Time       `json:"created_at" firestore:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" firestore:"updated_at"`
	Metadata    map[string]any  `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}
