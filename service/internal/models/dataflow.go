package models

// DataflowRequest represents the payload for submitting a Dataflow Flex job.
type DataflowRequest struct {
	Action               string                 `json:"action"`                  // e.g., "start", "stop"
	StopMode             string                 `json:"stop_mode"`               // e.g., "drain", "cancel"
	ProjectID            string                 `json:"project_id"`
	Region               string                 `json:"region"`
	JobName              string                 `json:"job_name"`
	ContainerImage       string                 `json:"container_image"`
	Subnetwork           string                 `json:"subnetwork"`
	ServiceAccountEmail  string                 `json:"service_account_email"`
	TempLocation         string                 `json:"temp_location"`
	StagingLocation      string                 `json:"staging_location"`
	NumWorkers           int                    `json:"num_workers"`
	MaxWorkers           int                    `json:"max_workers"`
	IPConfiguration      string                 `json:"ip_configuration"`
	EnableStreamingEngine bool                  `json:"enable_streaming_engine"`
	IsStreaming          bool                   `json:"is_streaming"`
	Zone                 string                 `json:"zone,omitempty"`
	Parameters           map[string]interface{} `json:"parameters"`              // Additional parameters for the pipeline
}

// DataflowResponse represents the result of the submission.
type DataflowResponse struct {
	Message string `json:"message"`
	JobID   string `json:"job_id,omitempty"`
}
