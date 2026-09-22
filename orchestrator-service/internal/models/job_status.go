package models

import "time"

type ExecutionStatus string

const (
	StatusQueued             ExecutionStatus = "QUEUED"
	StatusProcessingGCS      ExecutionStatus = "PROCESSING_GCS"
	StatusMovingGCS          ExecutionStatus = "MOVING_GCS"
	StatusCompletedGCS       ExecutionStatus = "COMPLETED_GCS"
	StatusProcessingBigQuery ExecutionStatus = "PROCESSING_BIGQUERY"
	StatusCompletedBigQuery  ExecutionStatus = "COMPLETED_BIGQUERY"
	StatusProcessingDataflow ExecutionStatus = "PROCESSING_DATAFLOW"
	StatusSuccess            ExecutionStatus = "SUCCESS"
	StatusFailedGCS          ExecutionStatus = "FAILED_GCS"
	StatusFailedBQ           ExecutionStatus = "FAILED_BIGQUERY"
	StatusSkipped            ExecutionStatus = "COMPLETED_SKIPPED"
)

func (e ExecutionStatus) String() string {
	return string(e)
}

type JobStatus struct {
	JobID     string          `json:"job_id" firestore:"job_id"`
	Status    ExecutionStatus `json:"status" firestore:"status"` // Holds the ExecutionStatus string
	Message   string          `json:"message" firestore:"message"`
	UpdatedAt time.Time       `json:"updated_at" firestore:"updated_at"`
	Metadata  map[string]any  `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

type PipelineTaskPayload struct {
	JobID  string `json:"job_id" binding:"required"`
	TaskID string `json:"task_id,omitempty"`
}
