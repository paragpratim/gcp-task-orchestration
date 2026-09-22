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

type Config struct {
	JobsCollection   string
	StatusCollection string
	GcsQueueName     string
	BqQueueName      string
	IntakeQueueName  string
}

type Service struct {
	jobsRepo    gcp.DocumentRepository[models.IntakeJobDefinition]
	statusRepo  gcp.DocumentRepository[models.JobStatus]
	tasksRepo   gcp.TaskRepository
	storageRepo gcp.ObjectRepository
	cfg         Config
}

func NewService(jobs gcp.DocumentRepository[models.IntakeJobDefinition], status gcp.DocumentRepository[models.JobStatus], tasks gcp.TaskRepository, storage gcp.ObjectRepository, cfg Config) *Service {
	return &Service{
		jobsRepo:    jobs,
		statusRepo:  status,
		tasksRepo:   tasks,
		storageRepo: storage,
		cfg:         cfg,
	}
}

const (
	stageProcessing = "processing"
	stageProcessed  = "processed"
	stageFailed     = "failed"
)

const nextRunInterval = 15 * time.Minute

func (s *Service) ListFiles(ctx context.Context, task models.PipelineTaskPayload) error {
	jobID := task.JobID
	taskID := task.TaskID

	nextRunPayload := models.PipelineTaskPayload{JobID: jobID}
	nextRunBytes, err := json.Marshal(nextRunPayload)
	if err != nil {
		logger.Error(fmt.Sprintf("GCS Service: Failed marshalling next-run payload for Job %s, Task %s: %v", jobID, taskID, err))
		return fmt.Errorf("failed marshalling next-run payload for job %s: %w", jobID, err)
	}
	if err := s.tasksRepo.Put(ctx, s.cfg.IntakeQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSListFiles), nextRunBytes, nextRunInterval); err != nil {
		logger.Error(fmt.Sprintf("GCS Service: Failed scheduling next run for Job %s, Task %s: %v", jobID, taskID, err))
		return fmt.Errorf("failed scheduling next run for job %s: %w", jobID, err)
	}

	logger.Info(fmt.Sprintf("GCS Service: Initiating file discovery scanning block for Job: %s, Task: %s", jobID, taskID))

	// 1. Fetch current runtime execution state trace out of Firestore
	statusTracker, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, jobID)
	if err != nil {
		logger.Error(fmt.Sprintf("GCS Service: Failed verifying trace checkpoint tracking state for Job %s, Task %s: %v", jobID, taskID, err))
		s.failWorkflowStep(ctx, jobID, taskID, &models.JobStatus{JobID: jobID}, fmt.Sprintf("Failed verifying trace checkpoint tracking state: %v", err))
		return nil
	}
	if statusTracker == nil {
		logger.Warn(fmt.Sprintf("GCS Service: No status tracker found for Job %s, Task %s. Skipping file discovery.", jobID, taskID))
		return nil
	}
	if statusTracker.Status != models.StatusQueued && statusTracker.Status != models.StatusFailedGCS && statusTracker.Status != models.StatusSkipped {
		logger.Warn(fmt.Sprintf("GCS Service: Job %s, Task %s is not in QUEUED, FAILED_GCS, or COMPLETED_SKIPPED state (current: %s). Skipping file discovery.", jobID, taskID, statusTracker.Status))
		return nil
	}

	if taskID == "" {
		taskID = fmt.Sprintf("%d", time.Now().UTC().UnixMilli())
	}

	// 2. Advance execution trace status to PROCESSING_GCS
	statusTracker.Status = models.StatusProcessingGCS
	statusTracker.Message = "Scanning storage bucket boundaries against pattern criteria."
	statusTracker.UpdatedAt = time.Now().UTC()
	if statusTracker.Metadata == nil {
		statusTracker.Metadata = make(map[string]any)
	}
	statusTracker.Metadata["task_id"] = taskID
	if err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		logger.Error(fmt.Sprintf("GCS Service: Failed updating execution trace boundary for Job %s, Task %s: %v", jobID, taskID, err))
		return nil
	}

	jobConfig, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobID)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed to load layout specifications: %v", err))
		return nil
	}
	if jobConfig == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Static layout definitions missing for job %s", jobID))
		return nil
	}

	rawObjects, err := s.storageRepo.ListObjects(ctx, jobConfig.Source.BucketName, jobConfig.Source.Prefix)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Infrastructure object discovery failure: %v", err))
		return nil
	}

	var matchedObjects []string
	pattern := jobConfig.Source.FilePattern
	if pattern == "" {
		pattern = "*" // If empty, evaluate as wildcard matching everything
	}

	for _, objectKey := range rawObjects {
		baseFilename := filepath.Base(objectKey)
		isMatched, err := filepath.Match(pattern, baseFilename)
		if err != nil {
			s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Glob validation engine syntax compilation failure: %v", err))
			return nil
		}

		if isMatched {
			matchedObjects = append(matchedObjects, objectKey)
		}
	}

	// 6. Branch based on discovery matching volume metrics
	if len(matchedObjects) == 0 {
		logger.Warn(fmt.Sprintf("GCS Service: Zero objects matched expression pattern '%s' for Job: %s, Task: %s. Terminating workflow step safely.", pattern, jobID, taskID))
		statusTracker.Status = models.StatusSkipped
		statusTracker.Message = fmt.Sprintf("GCS scan finalized. Zero matching files identified for criteria: %s", pattern)
		statusTracker.UpdatedAt = time.Now().UTC()
		_ = s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker)
		return nil // Short-circuit execution loop cleanly; downstream pipeline calls are skipped
	}

	// 7. Inject discovered file names array into transient metadata map contract space
	if statusTracker.Metadata == nil {
		statusTracker.Metadata = make(map[string]any)
	}
	statusTracker.Metadata["discovered_files"] = matchedObjects
	statusTracker.Metadata["task_id"] = taskID
	statusTracker.Message = fmt.Sprintf("Discovered %d target assets matching pattern schemas.", len(matchedObjects))
	statusTracker.UpdatedAt = time.Now().UTC()

	if err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed persisting discovered assets state to tracking document: %v", err))
		return nil
	}

	// 8. Auto-route execution downstream to Step 2: Move Files
	taskPayload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed serializing next workflow transport block: %v", err))
		return nil
	}

	err = s.tasksRepo.Put(ctx, s.cfg.GcsQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSMoveFiles), payloadBytes, 10*time.Second)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed chaining task execution flow down to file processing archiver: %v", err))
		return nil
	}

	logger.Info(fmt.Sprintf("GCS Service: ListFiles successful for Job %s, Task %s. Triggered move transaction payload.", jobID, taskID))
	return nil
}

func (s *Service) MoveFiles(ctx context.Context, task models.PipelineTaskPayload) error {
	jobID := task.JobID
	taskID := task.TaskID
	logger.Info(fmt.Sprintf("GCS Service: Commencing archive relocation lifecycle sequence for Job: %s, Task: %s", jobID, taskID))

	// 1. Fetch runtime context map state out of tracking store
	statusTracker, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, jobID)
	if err != nil || statusTracker == nil {
		logger.Error(fmt.Sprintf("GCS Service: Critical state drift - failed to extract operational run metadata context tracker for Job %s, Task %s: %v", jobID, taskID, err))
		s.failWorkflowStep(ctx, jobID, taskID, &models.JobStatus{JobID: jobID}, fmt.Sprintf("Critical state drift - failed to extract operational run metadata context tracker: %v", err))
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
		logger.Warn(fmt.Sprintf("GCS Service: Job %s, Task %s is not in a state eligible for file relocation (current: %s). Skipping file move.", jobID, taskID, statusTracker.Status))
		return nil
	}
}

func (s *Service) moveToProcessing(ctx context.Context, jobID, taskID string, statusTracker *models.JobStatus) error {
	// 1b. Advance execution trace status to MOVING_GCS while relocation is in flight
	statusTracker.Status = models.StatusMovingGCS
	statusTracker.Message = "Relocating discovered assets to processing storage zone."
	statusTracker.UpdatedAt = time.Now().UTC()
	if err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed updating execution trace boundary: %v", err))
		return nil
	}

	// 2. Extract file listing slice from metadata map box container
	rawFiles, exists := statusTracker.Metadata["discovered_files"]
	if !exists || rawFiles == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Execution flow logic error: metadata list 'discovered_files' is missing.")
		return nil
	}
	discoveredObjects := extractDiscoveredFiles(rawFiles)

	// 3. Fetch original job rules to resolve bucket targets
	jobConfig, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobID)
	if err != nil || jobConfig == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed loading job configurations to map archive layout constraints: %v", err))
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
			s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Bucket copy-delete transaction aborted at file %s: %v", rawKey, err))
			return nil
		}

		// Save fully qualified Google Cloud Storage URI paths to pass to BigQuery load engine
		fullGcsURI := fmt.Sprintf("gs://%s/%s", bucketName, processingKey)
		relocatedURIs = append(relocatedURIs, fullGcsURI)
	}

	// 5. Overwrite Metadata map variables to point exclusively to the clean relocated GCS URI paths
	statusTracker.Metadata["discovered_files"] = relocatedURIs // Update string paths to points directly to safe processing boundaries
	statusTracker.Status = models.StatusCompletedGCS
	statusTracker.Message = fmt.Sprintf("Successfully relocated %d items to processing storage zone.", len(relocatedURIs))
	statusTracker.UpdatedAt = time.Now().UTC()

	if err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed locking updated metadata URI configurations: %v", err))
		return nil
	}

	// 6. Chain payload execution over to the BigQuery load step queue
	taskPayload := models.PipelineTaskPayload{JobID: jobID, TaskID: taskID}
	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed marshalling downstream batch load pipeline block payload: %v", err))
		return nil
	}

	err = s.tasksRepo.Put(ctx, s.cfg.BqQueueName, taskspb.HttpMethod_POST, routes.Full(routes.BigQueryLoadJobCreate), payloadBytes, 10*time.Second)
	if err != nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed routing ingestion block container down to BigQuery task queue: %v", err))
		return nil
	}

	logger.Info(fmt.Sprintf("GCS Service: MoveFiles archive phase concluded for Job %s, Task %s. Control transferred to BigQuery Queue.", jobID, taskID))
	return nil
}

func (s *Service) finalizeRun(ctx context.Context, jobID, taskID string, statusTracker *models.JobStatus, stage string) error {
	rawFiles, exists := statusTracker.Metadata["discovered_files"]
	if !exists || rawFiles == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, "Execution flow logic error: metadata list 'discovered_files' is missing.")
		return nil
	}
	discoveredObjects := extractDiscoveredFiles(rawFiles)

	jobConfig, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobID)
	if err != nil || jobConfig == nil {
		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed loading job configurations to map archive layout constraints: %v", err))
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
			s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Bucket copy-delete transaction aborted at file %s: %v", srcKey, err))
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

	if err := s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *statusTracker); err != nil {

		s.failWorkflowStep(ctx, jobID, taskID, statusTracker, fmt.Sprintf("Failed persisting requeued job status: %v", err))
		return nil
	}

	taskPayload := models.PipelineTaskPayload{JobID: jobID, TaskID: ""}
	payloadBytes, err := json.Marshal(taskPayload)
	if err != nil {
		logger.Error(fmt.Sprintf("GCS Service: Failed marshalling requeue tracking payload for Job %s, Task %s: %v", jobID, taskID, err))
		return fmt.Errorf("failed marshalling requeue tracking payload for job %s: %w", jobID, err)
	}

	if err := s.tasksRepo.Put(ctx, s.cfg.IntakeQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSListFiles), payloadBytes, nextRunInterval); err != nil {

		logger.Error(fmt.Sprintf("GCS Service: Failed enqueuing requeue tracking task for Job %s, Task %s: %v", jobID, taskID, err))
		return fmt.Errorf("failed enqueuing requeue tracking task for job %s: %w", jobID, err)
	}

	logger.Info(fmt.Sprintf("GCS Service: Finalized run for Job %s, Task %s. Relocated %d items to '%s' and reset to QUEUED.", jobID, taskID, len(relocatedURIs), stage))
	return nil
}

func (s *Service) failWorkflowStep(ctx context.Context, jobID, taskID string, tracker *models.JobStatus, cause string) {
	logger.Error(fmt.Sprintf("GCS Service Execution Failure [Job %s] [Task %s]: %s", jobID, taskID, cause))
	tracker.Status = models.StatusFailedGCS
	tracker.Message = cause
	tracker.UpdatedAt = time.Now().UTC()
	_ = s.statusRepo.Put(ctx, s.cfg.StatusCollection, jobID, *tracker)
}

func buildObjectKey(prefix, stage, taskID, filename string) string {
	segments := make([]string, 0, 4)
	if prefix != "" {
		segments = append(segments, strings.Trim(prefix, "/"))
	}
	segments = append(segments, stage, taskID, filename)
	return strings.Join(segments, "/")
}

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
