package routes

// APIPrefix is the prefix for all API endpoints.
const APIPrefix = "/api/v1"

const (
	// --- Intake endpoints ---

	IntakeHealth     = "/intake/health"
	IntakeJobCreate  = "/intake/job"
	IntakeJobGet     = "/intake/job/:id"
	IntakeJobGetAll  = "/intake/jobs"
	IntakeJobUpdate  = "/intake/job"
	IntakeJobDelete  = "/intake/job/:id"
	IntakeJobsQueue  = "/intake/jobs/queue"
	IntakeJobsStatus = "/intake/jobs/status"

	// --- GCS endpoints ---

	GCSListFiles = "/gcs/files/list"
	GCSMoveFiles = "/gcs/files/move"

	// --- BigQuery endpoints ---

	BigQueryLoadJobCreate = "/bigquery/job/create"
	BigQueryLoadJobCheck  = "/bigquery/job/check"
	BigQueryRegionCheck   = "/bigquery/region/check"
)

// Full returns the full API path by prepending the APIPrefix to the given path.
func Full(path string) string {
	return APIPrefix + path
}
