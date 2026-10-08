package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"orchestrator/internal/gcp"
	"orchestrator/internal/logger"
	"orchestrator/internal/models"
	"orchestrator/internal/routes"
	"time"

	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/gin-gonic/gin"
)

// Config holds the configuration for the admin service.
type Config struct {
	JobsCollection   string
	StatusCollection string
	LogCollection    string
	AdminQueueName   string
}

// Service provides methods to manage admin jobs and their statuses.
type Service struct {
	jobsRepo      gcp.DocumentRepository[models.JobDefinition]
	statusRepo    gcp.DocumentRepository[models.JobStatus]
	logRepo       gcp.DocumentRepository[models.JobStatusLog]
	taskQueueRepo gcp.TaskRepository
	cfg           Config
}

// NewService creates a new instance of the admin Service with the provided repositories and configuration.
func NewService(jobsRepo gcp.DocumentRepository[models.JobDefinition], statusRepo gcp.DocumentRepository[models.JobStatus], logRepo gcp.DocumentRepository[models.JobStatusLog], tasks gcp.TaskRepository, cfg Config) *Service {
	return &Service{
		jobsRepo:      jobsRepo,
		statusRepo:    statusRepo,
		logRepo:       logRepo,
		taskQueueRepo: tasks,
		cfg:           cfg,
	}
}

// HealthCheck logs a message indicating that the service is operational.
func (s *Service) HealthCheck(ctx *gin.Context) error {
	logger.Debug("Service is Up and Running")
	return nil
}

// CreateJob validates and persists a new admin job definition.
func (s *Service) CreateJob(ctx *gin.Context, jobDefinition models.JobDefinition) (*models.JobDefinition, error) {
	if err := jobDefinition.Source.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	jobDefinition.CreatedAt = &now
	jobDefinition.UpdatedAt = &now

	result, err := s.jobsRepo.Put(ctx, s.cfg.JobsCollection, "", jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to persist initial admin state: %w", err)
	}
	return result, nil
}

// GetJob retrieves an admin job definition by its ID from the repository.
func (s *Service) GetJob(ctx *gin.Context, id string) (*models.JobDefinition, error) {
	job, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve admin job: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("admin job with id %s not found", id)
	}
	return job, nil
}

// GetAllJobs retrieves all admin job definitions from the repository.
func (s *Service) GetAllJobs(ctx *gin.Context) (*[]models.JobDefinition, error) {
	allJobs, err := s.jobsRepo.GetAll(ctx, s.cfg.JobsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all admin jobs: %w", err)
	}
	return allJobs, nil
}

// GetAllJobStatuses retrieves all admin job statuses from the repository.
func (s *Service) GetAllJobStatuses(ctx *gin.Context) (*[]models.JobStatus, error) {
	allJobStatuses, err := s.statusRepo.GetAll(ctx, s.cfg.StatusCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all admin job statuses: %w", err)
	}
	return allJobStatuses, nil
}

// GetAllJobStatusLogs retrieves all admin job status logs from the repository.
func (s *Service) GetAllJobStatusLogs(ctx *gin.Context) (*[]models.JobStatusLog, error) {
	allJobStatusLogs, err := s.logRepo.GetAll(ctx, s.cfg.LogCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all admin job status logs: %w", err)
	}
	return allJobStatusLogs, nil
}

// UpdateJob validates and updates an existing admin job definition.
func (s *Service) UpdateJob(ctx *gin.Context, jobDefinition models.JobDefinition) (*models.JobDefinition, error) {
	if err := jobDefinition.Source.Validate(); err != nil {
		return nil, err
	}

	if jobDefinition.ID == "" {
		return nil, fmt.Errorf("cannot update a job without a valid transaction identifier 'id'")
	}

	existingJob, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobDefinition.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve job context before mutation: %w", err)
	}
	if existingJob == nil {
		return nil, fmt.Errorf("job target with id %s does not exist", jobDefinition.ID)
	}

	now := time.Now().UTC()
	jobDefinition.UpdatedAt = &now

	result, err := s.jobsRepo.Put(ctx, s.cfg.JobsCollection, jobDefinition.ID, jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to overwrite persistent admin state: %w", err)
	}

	return result, nil
}

// DeleteJob removes an existing admin job by its ID.
func (s *Service) DeleteJob(ctx *gin.Context, id string) error {

	existingJob, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return fmt.Errorf("failed to locate target job before removal tracking: %w", err)
	}
	if existingJob == nil {
		return fmt.Errorf("cannot delete job: tracking target with id %s was not found", id)
	}

	err = s.jobsRepo.Delete(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return fmt.Errorf("failed to completely purge admin job record: %w", err)
	}

	return nil
}

// QueueActiveJobs retrieves and queues active admin jobs for processing, returning their statuses.
func (s *Service) QueueActiveJobs(ctx *gin.Context, jobDefinition models.JobDefinition) (*[]models.JobStatus, error) {
	var jobsToQueue []models.JobDefinition
	var jobStatuses []models.JobStatus

	if jobDefinition.ID != "" {
		retrieved, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, jobDefinition.ID)
		if err != nil {
			return nil, fmt.Errorf("failed verifying dispatch target: %w", err)
		}

		if retrieved != nil {
			jobsToQueue = append(jobsToQueue, *retrieved)
		}
	} else {
		allJobs, err := s.jobsRepo.GetAll(ctx, s.cfg.JobsCollection)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate jobs for queue migration: %w", err)
		}

		if allJobs != nil {
			jobsToQueue = append(jobsToQueue, *allJobs...)
		}
	}
	// Iterate over the jobs to queue and handle each one
	for _, job := range jobsToQueue {

		existingJob, err := s.statusRepo.Get(ctx, s.cfg.StatusCollection, job.ID)
		if err != nil {
			return nil, fmt.Errorf("failed checkpoint tracking verification for job %s: %w", job.ID, err)
		}

		if existingJob != nil {
			logger.Warn("Concurrency Guard: Job is already registered in the status store. Skipping queue dispatch.", "JOB_ID", job.ID)
			jobStatuses = append(jobStatuses, *existingJob)
			continue // Skip submitting this job to Cloud Tasks completely
		}

		// Generate a unique task ID based on the current timestamp in milliseconds
		taskID := fmt.Sprintf("%d", time.Now().UTC().UnixMilli())

		// Prepare the payload for the Cloud Task, including the job ID and task ID
		taskPayload := models.PipelineTaskPayload{
			JobID:  job.ID,
			TaskID: taskID,
		}
		payloadBytes, err := json.Marshal(taskPayload)
		if err != nil {
			return nil, fmt.Errorf("failed lean task network serialization for job %s: %w", job.ID, err)
		}

		// Queue the task in Cloud Tasks with the specified method, route, and payload
		err = s.taskQueueRepo.Put(ctx, s.cfg.AdminQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSListFiles), payloadBytes, 0)
		if err != nil {
			return nil, fmt.Errorf("queue execution aborted at task dispatch phase for job %s: %w", job.ID, err)
		}

		// Create a new job status with the generated task ID and persist it
		jobStatus := models.JobStatus{
			JobID:     job.ID,
			Status:    models.StatusQueued,
			UpdatedAt: time.Now().UTC(),
			Metadata:  map[string]any{models.MetadataKeyTaskID: taskID},
		}

		// Persist the job status to the status repository
		_, err = s.statusRepo.Put(ctx, s.cfg.StatusCollection, job.ID, jobStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to persist job status for job %s: %w", job.ID, err)
		}

		// Write the initial log entry
		s.writeLog(ctx, job.ID, taskID, models.StatusQueued, "Job queued for processing")

		jobStatuses = append(jobStatuses, jobStatus)
	}

	return &jobStatuses, nil
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
