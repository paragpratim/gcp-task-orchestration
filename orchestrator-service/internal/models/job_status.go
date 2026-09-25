package models

import "time"

// ExecutionStatus represents the status of a job execution.
type ExecutionStatus string

const (
	StatusQueued             ExecutionStatus = "QUEUED"
	StatusSkipped            ExecutionStatus = "COMPLETED_SKIPPED"
	StatusSuccess            ExecutionStatus = "SUCCESS"
	StatusProcessingGCS      ExecutionStatus = "PROCESSING_GCS"
	StatusMovingGCS          ExecutionStatus = "MOVING_GCS"
	StatusCompletedGCS       ExecutionStatus = "COMPLETED_GCS"
	StatusFailedGCS          ExecutionStatus = "FAILED_GCS"
	StatusProcessingBigQuery ExecutionStatus = "PROCESSING_BIGQUERY"
	StatusCompletedBigQuery  ExecutionStatus = "COMPLETED_BIGQUERY"
	StatusFailedBQ           ExecutionStatus = "FAILED_BIGQUERY"
	StatusProcessingDataflow ExecutionStatus = "PROCESSING_DATAFLOW"
	StatusCompletedDataflow  ExecutionStatus = "COMPLETED_DATAFLOW"
	StatusFailedDataflow     ExecutionStatus = "FAILED_DATAFLOW"
)

const (
	MetadataKeyJobID         = "job_id"
	MetadataKeyTaskID        = "task_id"
	MetadataKeyDiscovered    = "discovered_files"
	MetadataKeyBigQueryJobID = "bq_job_id"
)

func (e ExecutionStatus) String() string {
	return string(e)
}

// JobStatus represents the status of a job execution.
type JobStatus struct {
	JobID     string          `json:"job_id" firestore:"job_id"`
	Status    ExecutionStatus `json:"status" firestore:"status"` // Holds the ExecutionStatus string
	Message   string          `json:"message" firestore:"message"`
	UpdatedAt time.Time       `json:"updated_at" firestore:"updated_at"`
	Metadata  map[string]any  `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// PipelineTaskPayload represents the payload for a pipeline task.
type PipelineTaskPayload struct {
	JobID  string `json:"job_id" binding:"required"`
	TaskID string `json:"task_id,omitempty"`
}
