package bigquery

import (
	"context"
	"orchestrator/internal/gcp"
	"orchestrator/internal/models"
)

type Config struct {
	JobsCollection   string
	StatusCollection string
	BqQueueName      string
}

type Service struct {
	jobsRepo     gcp.DocumentRepository[models.IntakeJobDefinition]
	statusRepo   gcp.DocumentRepository[models.JobStatus]
	tasksRepo    gcp.TaskRepository
	bigQueryRepo gcp.DataRepository
	cfg          Config
}

func NewService(jobs gcp.DocumentRepository[models.IntakeJobDefinition], status gcp.DocumentRepository[models.JobStatus], tasks gcp.TaskRepository, bigQueryRepo gcp.DataRepository, cfg Config) *Service {
	return &Service{
		jobsRepo:     jobs,
		statusRepo:   status,
		tasksRepo:    tasks,
		bigQueryRepo: bigQueryRepo,
		cfg:          cfg,
	}
}

// CreateLoadJob kicks off a BigQuery load job for the files relocated by the
// GCS stage (expected to run once the tracker is in COMPLETED_GCS).
func (s *Service) CreateLoadJob(ctx context.Context, task models.PipelineTaskPayload) error {
	//TODO: Implement BigQuery load job creation:
	// 1. Fetch/validate the JobStatus tracker (expect COMPLETED_GCS; else log+return nil)
	// 2. Fetch job config for datasetID/tableID and the discovered file GCS URIs
	// 3. Call s.bigQueryRepo.CreateGCSLoadJob(...) and persist the returned BQ job ID
	//    onto the tracker's Metadata
	// 4. Advance status to PROCESSING_BIGQUERY and dispatch a follow-up task to
	//    CheckLoadJobStatus (with a short delay) via s.tasksRepo.Put(...)
	// 5. Follow the same fault-tolerant pattern as gcs.Service: log errors and
	//    return nil so Cloud Tasks doesn't fast-retry; only return an error when
	//    losing the dispatch would strand the job with no way back into the flow.
	return nil
}

// CheckLoadJobStatus polls an in-flight BigQuery load job and finalizes the
// pipeline status (COMPLETED_BIGQUERY / FAILED_BIGQUERY) based on the outcome,
// which then allows gcs.MoveFiles to relocate files to their final home.
func (s *Service) CheckLoadJobStatus(ctx context.Context, task models.PipelineTaskPayload) error {
	//TODO: Implement BigQuery load job status polling:
	// 1. Fetch the JobStatus tracker and the stored BQ job ID from Metadata
	// 2. Call s.bigQueryRepo.CheckJobStatus(...)
	// 3. If still running, reschedule this same check onto BqQueueName with a delay
	// 4. If done, set status to StatusCompletedBigQuery or StatusFailedBQ and
	//    dispatch a task to gcs.MoveFiles so the run gets finalized/relocated
	return nil
}

// CheckDatasetRegion verifies a job's destination dataset region is available
// before allowing a load job to be started against it.
func (s *Service) CheckDatasetRegion(ctx context.Context, task models.PipelineTaskPayload) error {
	//TODO: Implement BigQuery dataset region validation:
	// 1. Fetch job config to resolve the target datasetID
	// 2. Call s.bigQueryRepo.DatasetRegionAvailable(...) and compare against the
	//    job's expected region
	// 3. Log/report a mismatch (does not currently mutate JobStatus)
	return nil
}
