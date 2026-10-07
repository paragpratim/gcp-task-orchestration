package bigquery

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/internal/gcp"
	"orchestrator/internal/logger"
	"orchestrator/internal/models"
	"orchestrator/internal/routes"
	"time"

	"cloud.google.com/go/bigquery"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
)

// Config holds configuration values for the BigQuery service.
type Config struct {
	JobsCollection   string
	StatusCollection string
	LogCollection    string
	GcsQueueName     string
	BqQueueName      string
}

// Service encapsulates the logic for managing BigQuery load jobs within the orchestrator pipeline.
type Service struct {
	jobsRepo     gcp.DocumentRepository[models.JobDefinition]
	statusRepo   gcp.DocumentRepository[models.JobStatus]
	logRepo      gcp.DocumentRepository[models.JobStatusLog]
	tasksRepo    gcp.TaskRepository
	bigQueryRepo gcp.DataRepository
	cfg          Config
}

// NewService creates a new instance of the BigQuery service with the provided repositories and configuration.
func NewService(jobs gcp.DocumentRepository[models.JobDefinition], status gcp.DocumentRepository[models.JobStatus], logs gcp.DocumentRepository[models.JobStatusLog], tasks gcp.TaskRepository, bigQueryRepo gcp.DataRepository, cfg Config) *Service {
	return &Service{
		jobsRepo:     jobs,
		statusRepo:   status,
		logRepo:      logs,
		tasksRepo:    tasks,
		bigQueryRepo: bigQueryRepo,
		cfg:          cfg,
	}
}

// nextBQCheckInterval defines the interval at which the orchestrator will poll the status of an in-flight BigQuery load job.
const nextBQCheckInterval = 1 * time.Minute

// nextTaskInterval defines the interval for scheduling the next task run.
const nextTaskInterval = 10 * time.Second

// CreateLoadJob kicks off a BigQuery load job for the files relocated by the
// GCS stage (expected to run once the tracker is in COMPLETED_GCS).
func (s *Service) CreateLoadJob(ctx context.Context, task models.PipelineTaskPayload) error {
	jobID := task.JobID
	taskID := task.TaskID

	logger.Info("Starting BigQuery load job creation", "JOB_ID", jobID, "TASK_ID", taskID)

	statusTracker, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, jobID)
	if err != nil {
		return fmt.Errorf("failed loading lifecycle tracker before BigQuery load creation for JOB_ID %s: %w", jobID, err)
	}
	if statusTracker == nil {
		logger.Warn("No status tracker found before BigQuery load creation; skipping job", "JOB_ID", jobID, "TASK_ID", taskID)
		return nil
	}
	if statusTracker.Status != models.StatusCompletedGCS {
		logger.Warn("Skipping BigQuery creation because job is not in COMPLETED_GCS", "JOB_ID", jobID, "TASK_ID", taskID, "STATUS", statusTracker.Status)
		return nil
	}

	if statusTracker.Metadata == nil {
		return fmt.Errorf("no metadata found for job ID %s", jobID)
	}

	statusTracker.Status = models.StatusProcessingBigQuery
	statusTracker.Message = "Submitting BigQuery load job for relocated GCS files."
	statusTracker.UpdatedAt = time.Now().UTC()
	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		return fmt.Errorf("failed persisting BigQuery load job state transition for JOB_ID %s: %w", jobID, err)
	}
	s.writeLog(ctx, jobID, taskID, models.StatusProcessingBigQuery, statusTracker.Message)

	jobConfig, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobID)
	if err != nil || jobConfig == nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed loading job configuration for BigQuery load", err)
	}

	rawFiles, exists := statusTracker.Metadata[models.MetadataKeyDiscovered]
	if !exists || rawFiles == nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Metadata list 'discovered_files' is missing for BigQuery load", nil)
	}

	fileURIs := extractDiscoveredFiles(rawFiles)
	if len(fileURIs) == 0 {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "No GCS files were available to start the BigQuery load job", nil)
	}

	datasetID := jobConfig.Target.DatasetID
	tableName := jobConfig.Target.TableName
	sourceType := jobConfig.Source.FileType
	if datasetID == "" || tableName == "" || sourceType == "" {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Target dataset or table information is missing for BigQuery load", nil)
	}

	bqJobID, err := s.bigQueryRepo.CreateGCSLoadJob(ctx, datasetID, tableName, fileURIs, bigquery.DataFormat(sourceType))
	if err != nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "BigQuery load job creation failed", err)
	}

	statusTracker.Metadata[models.MetadataKeyBigQueryJobID] = bqJobID
	statusTracker.Message = fmt.Sprintf("BigQuery load job %s started successfully for %d GCS files.", bqJobID, len(fileURIs))
	statusTracker.UpdatedAt = time.Now().UTC()
	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed persisting BigQuery job ID back into tracker", err)
	}
	s.writeLog(ctx, jobID, taskID, statusTracker.Status, statusTracker.Message)

	payload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed marshalling BigQuery status check payload", err)
	}

	if err := s.tasksRepo.Put(ctx, s.cfg.BqQueueName, taskspb.HttpMethod_POST, routes.Full(routes.BigQueryLoadJobCheck), payloadBytes, nextBQCheckInterval); err != nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed enqueueing BigQuery load status check task", err)
	}

	logger.Info("BigQuery load job created and status polling task scheduled", "JOB_ID", jobID, "TASK_ID", taskID, "BQ_JOB_ID", bqJobID)
	return nil
}

// CheckLoadJobStatus polls an in-flight BigQuery load job and finalizes the
// pipeline status (COMPLETED_BIGQUERY / FAILED_BIGQUERY) based on the outcome,
// which then allows gcs.MoveFiles to relocate files to their final home.
func (s *Service) CheckLoadJobStatus(ctx context.Context, task models.PipelineTaskPayload) error {
	jobID := task.JobID
	taskID := task.TaskID

	logger.Info("Polling BigQuery load job status", "JOB_ID", jobID, "TASK_ID", taskID)

	statusTracker, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, jobID)
	if err != nil {
		return fmt.Errorf("failed loading job status for JOB_ID %s: %w", jobID, err)
	}
	if statusTracker == nil {
		logger.Warn("No status tracker found while polling BigQuery load job; skipping", "JOB_ID", jobID, "TASK_ID", taskID)
		return nil
	}
	if statusTracker.Status != models.StatusProcessingBigQuery {
		logger.Warn("Job is not in PROCESSING_BIGQUERY state; skipping stale/duplicate status check", "JOB_ID", jobID, "TASK_ID", taskID, "STATUS", statusTracker.Status)
		return nil
	}

	bqJobID, ok := statusTracker.Metadata[models.MetadataKeyBigQueryJobID].(string)
	if !ok || bqJobID == "" {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Missing BigQuery job ID in tracker metadata", nil)
	}

	jobStatus, err := s.bigQueryRepo.CheckJobStatus(ctx, bqJobID)
	if err != nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed checking BigQuery load job status", err)
	}

	if jobStatus == nil {
		logger.Warn("BigQuery job status lookup returned nil; retrying later", "JOB_ID", jobID, "TASK_ID", taskID, "BQ_JOB_ID", bqJobID)
		return s.rescheduleBQCheck(ctx, jobID, taskID)
	}

	if jobStatus.Err() != nil {
		return s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "BigQuery load job failed", jobStatus.Err())
	}

	if jobStatus.State == bigquery.Done {
		statusTracker.Status = models.StatusCompletedBigQuery
		statusTracker.Message = "BigQuery load job completed successfully. Finalizing GCS run."
		statusTracker.UpdatedAt = time.Now().UTC()
		if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
			return fmt.Errorf("failed persisting successful BigQuery job state for job %s: %w", jobID, err)
		}
		s.writeLog(ctx, jobID, taskID, models.StatusCompletedBigQuery, statusTracker.Message)

		payload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed marshalling GCS finalization payload for job %s: %w", jobID, err)
		}

		if err := s.tasksRepo.Put(ctx, s.cfg.GcsQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSMoveFiles), payloadBytes, nextTaskInterval); err != nil {
			return fmt.Errorf("failed enqueuing downstream GCS finalization task for job %s: %w", jobID, err)
		}

		logger.Info("BigQuery load job complete; routing to GCS finalization", "JOB_ID", jobID, "TASK_ID", taskID, "BQ_JOB_ID", bqJobID)
		return nil
	}

	logger.Info("BigQuery load job is still running; rechecking later", "JOB_ID", jobID, "TASK_ID", taskID, "BQ_JOB_ID", bqJobID, "STATE", jobStatus.State)
	return s.rescheduleBQCheck(ctx, jobID, taskID)
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

// failWorkflowStep updates the job status to FAILED_BIGQUERY and enqueues a GCS finalization task.
func (s *Service) failWorkflowStep(ctx context.Context, jobID, taskID string, tracker *models.JobStatus, message string, err error) error {
	logger.Error("BigQuery Load Job failed", "JOB_ID", jobID, "TASK_ID", taskID, "MESSAGE", message, "ERROR", err)
	if tracker == nil {
		tracker = &models.JobStatus{JobID: jobID}
	}
	tracker.Status = models.StatusFailedBQ
	tracker.Message = message
	tracker.UpdatedAt = time.Now().UTC()
	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *tracker); err != nil {
		return fmt.Errorf("failed persisting failed BigQuery state before finalization route for job %s: %w", jobID, err)
	}
	s.writeLog(ctx, jobID, taskID, models.StatusFailedBQ, message)

	payload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return fmt.Errorf("failed marshalling GCS failed-finalization payload for job %s: %w", jobID, marshalErr)
	}

	if err := s.tasksRepo.Put(ctx, s.cfg.GcsQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSMoveFiles), payloadBytes, nextTaskInterval); err != nil {
		return fmt.Errorf("failed enqueueing GCS failed-finalization task for job %s: %w", jobID, err)
	}
	return nil
}

// rescheduleBQCheck enqueues a new task to check the status of an in-flight BigQuery load job after a defined interval.
func (s *Service) rescheduleBQCheck(ctx context.Context, jobID, taskID string) error {
	payload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed marshalling retry payload for job %s: %w", jobID, err)
	}

	if err := s.tasksRepo.Put(ctx, s.cfg.BqQueueName, taskspb.HttpMethod_POST, routes.Full(routes.BigQueryLoadJobCheck), payloadBytes, nextBQCheckInterval); err != nil {
		return fmt.Errorf("failed enqueuing BigQuery retry task for job %s: %w", jobID, err)
	}
	return nil
}

// extractDiscoveredFiles converts the raw metadata value for discovered files into a slice of strings.
func extractDiscoveredFiles(rawFiles any) []string {
	var discovered []string
	if slice, ok := rawFiles.([]any); ok {
		for _, item := range slice {
			if str, ok := item.(string); ok {
				discovered = append(discovered, str)
			}
		}
		return discovered
	}
	if slice, ok := rawFiles.([]string); ok {
		return slice
	}
	return discovered
}

// writeLog writes a JobStatusLog entry to the log repository, keyed by taskID.
func (s *Service) writeLog(ctx context.Context, jobID, taskID string, status models.ExecutionStatus, message string) {
	if taskID == "" {
		return
	}
	logEntry := models.JobStatusLog{
		JobID:     jobID,
		TaskID:    taskID,
		Status:    status,
		Message:   message,
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = s.logRepo.Put(ctx, s.cfg.LogCollection, taskID, logEntry)
}
