package routes

// APIPrefix is the prefix for all API endpoints.
const APIPrefix = "/api/v1"

const (
	// --- Admin endpoints ---

	AdminHealth              = "/admin/health"
	AdminJobCreate           = "/admin/job"
	AdminJobGet              = "/admin/job/:id"
	AdminJobGetAll           = "/admin/jobs"
	AdminJobUpdate           = "/admin/job"
	AdminJobDelete           = "/admin/job/:id"
	AdminJobsQueue           = "/admin/jobs/queue"
	AdminJobsStatus          = "/admin/jobs/status"
	AdminJobsStatusLogGetAll = "/admin/jobs/status/logs"

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
