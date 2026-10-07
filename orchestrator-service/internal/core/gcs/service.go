package gcs

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/internal/gcp"
	"orchestrator/internal/logger"
	"orchestrator/internal/models"
	"orchestrator/internal/routes"
	"path/filepath"
	"strings"
	"time"

	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
)

// Config holds configuration parameters for the GCS service.
type Config struct {
	JobsCollection   string
	StatusCollection string
	GcsQueueName     string
	BqQueueName      string
	AdminQueueName   string
}

// Service provides methods to manage GCS file operations and task scheduling.
type Service struct {
	jobsRepo    gcp.DocumentRepository[models.JobDefinition]
	statusRepo  gcp.DocumentRepository[models.JobStatus]
	tasksRepo   gcp.TaskRepository
	storageRepo gcp.ObjectRepository
	cfg         Config
}

// NewService creates a new instance of the GCS Service with the provided repositories and configuration.
func NewService(jobs gcp.DocumentRepository[models.JobDefinition], status gcp.DocumentRepository[models.JobStatus], tasks gcp.TaskRepository, storage gcp.ObjectRepository, cfg Config) *Service {
	return &Service{
		jobsRepo:    jobs,
		statusRepo:  status,
		tasksRepo:   tasks,
		storageRepo: storage,
		cfg:         cfg,
	}
}

// Constants representing different subdirectories of file processing in GCS.
const (
	stageProcessing = "processing"
	stageProcessed  = "processed"
	stageFailed     = "failed"
)

// nextRunInterval defines the interval for scheduling the next Job run.
const nextRunInterval = 15 * time.Minute

// nextTaskInterval defines the interval for scheduling the next task run.
const nextTaskInterval = 10 * time.Second

// ListFiles scans the GCS bucket for files matching the job's pattern and schedules the next run.
func (s *Service) ListFiles(ctx context.Context, task models.PipelineTaskPayload) error {
	jobID := task.JobID
	taskID := task.TaskID

	nextTaskID := fmt.Sprintf("%d", time.Now().UTC().UnixMilli())
	nextRunPayload := models.PipelineTaskPayload{JobID: jobID, TaskID: nextTaskID}
	nextRunBytes, err := json.Marshal(nextRunPayload)
	if err != nil {
		return fmt.Errorf("failed marshalling next-run payload for job %s: %w", jobID, err)
	}
	if err := s.tasksRepo.Put(ctx, s.cfg.AdminQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSListFiles), nextRunBytes, nextRunInterval); err != nil {
		return fmt.Errorf("failed scheduling next run for job %s: %w", jobID, err)
	}

	logger.Info("Initiating file discovery scanning", "JOB_ID", jobID, "TASK_ID", taskID)

	// 1. Fetch current runtime execution state trace out of Firestore
	statusTracker, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, jobID)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, &models.JobStatus{JobID: jobID}, "Failed verifying trace checkpoint tracking state", err)
		return nil
	}
	if statusTracker == nil {
		logger.Warn("No status tracker found. Skipping file discovery.", "JOB_ID", jobID, "TASK_ID", taskID)
		return nil
	}
	if statusTracker.Status != models.StatusQueued && statusTracker.Status != models.StatusFailedGCS && statusTracker.Status != models.StatusSkipped {
		logger.Warn("Job is not in QUEUED, FAILED_GCS, or COMPLETED_SKIPPED state. Skipping file discovery.", "JOB_ID", jobID, "TASK_ID", taskID, "CURRENT_STATUS", statusTracker.Status)
		return nil
	}

	// 2. Advance execution trace status to PROCESSING_GCS
	statusTracker.Status = models.StatusProcessingGCS
	statusTracker.Message = "Scanning storage bucket boundaries against pattern criteria."
	statusTracker.UpdatedAt = time.Now().UTC()
	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		logger.Error("Failed updating execution trace boundary ", "JOB_ID", jobID, "TASK_ID", taskID, "ERROR", err)
		return nil
	}

	jobConfig, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobID)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed to load layout specifications", err)
		return nil
	}
	if jobConfig == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Static layout definitions missing", nil)
		return nil
	}

	matchedObjects, err := s.storageRepo.ListObjects(ctx, jobConfig.Source.BucketName, jobConfig.Source.Prefix, jobConfig.Source.FilePattern)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "GCS object discovery failure", err)
		return nil
	}

	fullFilePattern := fmt.Sprintf("gs://%s/%s/%s", jobConfig.Source.BucketName, jobConfig.Source.Prefix, jobConfig.Source.FilePattern)
	// Branch based on discovery matching volume metrics
	if len(matchedObjects) == 0 {
		logger.Warn("Zero objects matched expression pattern. Terminating workflow step safely.", "JOB_ID", jobID, "TASK_ID", taskID, "PATTERN", fullFilePattern)
		statusTracker.Status = models.StatusSkipped
		statusTracker.Message = fmt.Sprintf("GCS scan finalized. Zero matching files identified for criteria: %s", fullFilePattern)
		statusTracker.UpdatedAt = time.Now().UTC()
		_, _ = s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker)
		return nil // Short-circuit execution loop cleanly; downstream pipeline calls are skipped
	}

	// Inject discovered file names array into transient metadata map contract space
	statusTracker.Metadata[models.MetadataKeyDiscovered] = matchedObjects
	statusTracker.Message = fmt.Sprintf("Discovered %d target assets matching pattern schemas.", len(matchedObjects))
	statusTracker.UpdatedAt = time.Now().UTC()

	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed persisting discovered assets state to tracking document", err)
		return nil
	}

	// Auto-route execution downstream to Step 2: Move Files
	taskPayload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed serializing next workflow transport block", err)
		return nil
	}

	err = s.tasksRepo.Put(ctx, s.cfg.GcsQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSMoveFiles), payloadBytes, nextTaskInterval)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed chaining task execution flow down to file processing archiver", err)
		return nil
	}

	logger.Info("ListFiles successful. Triggered move transaction payload.", "JOB_ID", jobID, "TASK_ID", taskID)
	return nil
}

// MoveFiles relocates files from the discovery phase to the processing phase in GCS, and handles finalization for completed or failed runs.
func (s *Service) MoveFiles(ctx context.Context, task models.PipelineTaskPayload) error {
	jobID := task.JobID
	taskID := task.TaskID
	logger.Info("Commencing archive relocation lifecycle sequence", "JOB_ID", jobID, "TASK_ID", taskID)

	// 1. Fetch runtime context map state out of tracking store
	statusTracker, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, jobID)
	if err != nil || statusTracker == nil {
		s.failWorkflowStep(ctx, jobID, taskID, &models.JobStatus{JobID: jobID}, "Failed to extract operational run metadata context tracker", err)
		return nil
	}

	// 2. Branch relocation behavior based on which stage of the pipeline requested it.
	switch statusTracker.Status {
	case models.StatusProcessingGCS:
		return s.moveToProcessing(ctx, jobID, taskID, statusTracker)
	case models.StatusCompletedBigQuery:
		return s.finalizeRun(ctx, jobID, taskID, statusTracker, stageProcessed)
	case models.StatusFailedBQ:
		return s.finalizeRun(ctx, jobID, taskID, statusTracker, stageFailed)
	default:
		logger.Warn("Job is not in a state eligible for file relocation. Skipping file move.", "JOB_ID", jobID, "TASK_ID", taskID, "STATUS", statusTracker.Status)
		return nil
	}
}

// moveToProcessing handles the relocation of discovered files to the processing phase in GCS.
func (s *Service) moveToProcessing(ctx context.Context, jobID, taskID string, statusTracker *models.JobStatus) error {
	// 1b. Advance execution trace status to MOVING_GCS while relocation is in flight
	statusTracker.Status = models.StatusMovingGCS
	statusTracker.Message = "Relocating discovered assets to processing storage zone."
	statusTracker.UpdatedAt = time.Now().UTC()
	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed updating execution trace boundary", err)
		return nil
	}

	discoveredObjects, jobConfig, err := s.loadDiscoveredFilesContext(ctx, jobID, taskID, statusTracker)
	if err != nil {
		return nil
	}

	bucketName := jobConfig.Source.BucketName
	prefix := jobConfig.Source.Prefix
	var relocatedURIs []string

	for _, rawKey := range discoveredObjects {

		cleanFilename := filepath.Base(rawKey)

		processingKey := buildObjectKey(prefix, stageProcessing, taskID, cleanFilename)

		err := s.storageRepo.MoveFile(ctx, bucketName, rawKey, processingKey)
		if err != nil {
			s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Bucket copy-delete transaction aborted", err)
			return nil
		}

		// Save fully qualified Google Cloud Storage URI paths to pass to BigQuery load engine
		fullGcsURI := fmt.Sprintf("gs://%s/%s", bucketName, processingKey)
		relocatedURIs = append(relocatedURIs, fullGcsURI)
	}

	// 5. Overwrite Metadata map variables to point exclusively to the clean relocated GCS URI paths
	statusTracker.Metadata[models.MetadataKeyDiscovered] = relocatedURIs // Update string paths to points directly to safe processing boundaries
	statusTracker.Status = models.StatusCompletedGCS
	statusTracker.Message = fmt.Sprintf("Successfully relocated %d items to processing storage zone.", len(relocatedURIs))
	statusTracker.UpdatedAt = time.Now().UTC()

	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed locking updated metadata URI configurations", err)
		return nil
	}

	// 6. Chain payload execution over to the BigQuery load step queue
	taskPayload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed marshalling downstream batch load pipeline block payload", err)
		return nil
	}

	err = s.tasksRepo.Put(ctx, s.cfg.BqQueueName, taskspb.HttpMethod_POST, routes.Full(routes.BigQueryLoadJobCreate), payloadBytes, nextTaskInterval)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed routing ingestion block container down to BigQuery task queue", err)
		return nil
	}

	logger.Info("MoveFiles archive phase concluded. Control transferred to BigQuery Queue.", "JOB_ID", jobID, "TASK_ID", taskID)
	return nil
}

// finalizeRun handles the finalization of a run by relocating files to the specified stage and updating the job status.
func (s *Service) finalizeRun(ctx context.Context, jobID, taskID string, statusTracker *models.JobStatus, stage string) error {
	discoveredObjects, jobConfig, err := s.loadDiscoveredFilesContext(ctx, jobID, taskID, statusTracker)
	if err != nil {
		return nil
	}

	bucketName := jobConfig.Source.BucketName
	prefix := jobConfig.Source.Prefix
	var relocatedURIs []string

	for _, rawKey := range discoveredObjects {
		cleanFilename := filepath.Base(rawKey)

		srcKey := buildObjectKey(prefix, stageProcessing, taskID, cleanFilename)
		destKey := buildObjectKey(prefix, stage, taskID, cleanFilename)

		if err := s.storageRepo.MoveFile(ctx, bucketName, srcKey, destKey); err != nil {
			s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Bucket copy-delete transaction aborted", err)
			return nil
		}

		relocatedURIs = append(relocatedURIs, fmt.Sprintf("gs://%s/%s", bucketName, destKey))
	}

	// Reset the tracker to QUEUED, clearing per-run metadata and the TaskID so the
	// next ListFiles invocation starts a fresh run for this job.
	statusTracker.Status = models.StatusQueued
	statusTracker.Message = fmt.Sprintf("Run finalized: %d items relocated to '%s' storage zone. Job requeued for next run.", len(relocatedURIs), stage)
	statusTracker.UpdatedAt = time.Now().UTC()
	statusTracker.Metadata = map[string]any{}

	if _, err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed persisting requeued job status", err)
		return nil
	}

	logger.Info("Finalized run for Job and status reset to QUEUED.", "JOB_ID", jobID, "TASK_ID", taskID, "RELOCATED_COUNT", len(relocatedURIs), "STAGE", stage)
	return nil
}

// loadDiscoveredFilesContext retrieves the list of discovered files and the job configuration for the given job ID.
func (s *Service) loadDiscoveredFilesContext(ctx context.Context, jobID, taskID string, statusTracker *models.JobStatus) ([]string, *models.JobDefinition, error) {
	rawFiles, exists := statusTracker.Metadata[models.MetadataKeyDiscovered]
	if !exists || rawFiles == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Metadata list 'discovered_files' is missing.", nil)
		return nil, nil, fmt.Errorf("missing discovered files metadata for job %s", jobID)
	}

	discoveredObjects := extractDiscoveredFiles(rawFiles)
	jobConfig, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobID)
	if err != nil || jobConfig == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Failed loading job configurations", err)
		return nil, nil, fmt.Errorf("job config unavailable for job %s: %w", jobID, err)
	}

	return discoveredObjects, jobConfig, nil
}

// failWorkflowStep logs an error, updates the job status to failed, and persists the updated status to the repository.
func (s *Service) failWorkflowStep(ctx context.Context, jobID, taskID string, tracker *models.JobStatus, message string, err error) {
	logger.Error("GCS Service Execution Failure. It will be retried...", "JOB_ID", jobID, "TASK_ID", taskID, "MESSAGE", message, "ERROR", err)
	tracker.Status = models.StatusFailedGCS
	tracker.Message = message
	tracker.UpdatedAt = time.Now().UTC()
	_, _ = s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *tracker)
}

// buildObjectKey constructs a GCS object key by combining the prefix, stage, task ID, and filename.
func buildObjectKey(prefix, stage, taskID, filename string) string {
	segments := make([]string, 0, 4)
	if prefix != "" {
		segments = append(segments, strings.Trim(prefix, "/"))
	}
	segments = append(segments, stage, taskID, filename)
	return strings.Join(segments, "/")
}

// extractDiscoveredFiles converts the raw metadata value into a slice of strings representing discovered file paths.
func extractDiscoveredFiles(rawFiles any) []string {
	var discoveredObjects []string
	if slice, ok := rawFiles.([]any); ok {
		for _, item := range slice {
			if str, ok := item.(string); ok {
				discoveredObjects = append(discoveredObjects, str)
			}
		}
	} else if slice, ok := rawFiles.([]string); ok {
		discoveredObjects = slice
	}
	return discoveredObjects
}
