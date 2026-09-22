package routes

const APIPrefix = "/api/v1"

const (
	// Intake endpoints

	IntakeHealth    = "/intake/health"
	IntakeJobCreate = "/intake/job"
	IntakeJobUpdate = "/intake/job"
	IntakeJobDelete = "/intake/job/:id"
	IntakeJobsQueue = "/intake/jobs/queue"

	// GCS endpoints

	GCSListFiles = "/gcs/files/list"
	GCSMoveFiles = "/gcs/files/move"

	// BigQuery endpoints

	BigQueryLoadJob = "/bigquery/job"
)

func Full(path string) string {
	return APIPrefix + path
}
