package intake

import (
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

type Config struct {
	JobsCollection   string
	StatusCollection string
	IntakeQueueName  string
}

type Service struct {
	jobsRepo      gcp.DocumentRepository[models.IntakeJobDefinition]
	statusRepo    gcp.DocumentRepository[models.JobStatus]
	taskQueueRepo gcp.TaskRepository
	cfg           Config
}

func NewService(jobsRepo gcp.DocumentRepository[models.IntakeJobDefinition], statusRepo gcp.DocumentRepository[models.JobStatus], tasks gcp.TaskRepository, cfg Config) *Service {
	return &Service{
		jobsRepo:      jobsRepo,
		statusRepo:    statusRepo,
		taskQueueRepo: tasks,
		cfg:           cfg,
	}
}

func (s *Service) HealthCheck(ctx *gin.Context) error {
	logger.Debug("Service is Up and Running")
	return nil
}

func (s *Service) CreateIntakeJob(ctx *gin.Context, jobDefinition models.IntakeJobDefinition) (*models.IntakeJobDefinition, error) {
	if err := jobDefinition.Source.Validate(); err != nil {
		return nil, err
	}

	jobDefinition.CreatedAt = time.Now().UTC()
	jobDefinition.UpdatedAt = time.Now().UTC()

	err := s.jobsRepo.Put(ctx, s.cfg.JobsCollection, "", jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to persist initial intake state: %w", err)
	}
	return &jobDefinition, nil
}

func (s *Service) UpdateIntakeJob(ctx *gin.Context, jobDefinition models.IntakeJobDefinition) (*models.IntakeJobDefinition, error) {
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

	jobDefinition.UpdatedAt = time.Now().UTC()

	err = s.jobsRepo.Put(ctx, s.cfg.JobsCollection, jobDefinition.ID, jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to overwrite persistent intake state: %w", err)
	}

	return &jobDefinition, nil
}

func (s *Service) DeleteIntakeJob(ctx *gin.Context, id string) error {

	existingJob, err := s.jobsRepo.Get(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return fmt.Errorf("failed to locate target job before removal tracking: %w", err)
	}
	if existingJob == nil {
		return fmt.Errorf("cannot delete job: tracking target with id %s was not found", id)
	}

	err = s.jobsRepo.Delete(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return fmt.Errorf("failed to completely purge intake job record: %w", err)
	}

	return nil
}

func (s *Service) QueueActiveJobs(ctx *gin.Context, jobDefinition models.IntakeJobDefinition) (*[]models.JobStatus, error) {
	var jobsToQueue []models.IntakeJobDefinition
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

		taskID := fmt.Sprintf("%d", time.Now().UTC().UnixMilli())

		jobStatus := models.JobStatus{
			JobID:     job.ID,
			Status:    models.StatusQueued,
			UpdatedAt: time.Now().UTC(),
			Metadata:  map[string]any{models.MetadataKeyTaskID: taskID},
		}

		err = s.statusRepo.Put(ctx, s.cfg.StatusCollection, job.ID, jobStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to persist job status for job %s: %w", job.ID, err)
		}

		taskPayload := models.PipelineTaskPayload{
			JobID:  job.ID,
			TaskID: taskID,
		}
		payloadBytes, err := json.Marshal(taskPayload)
		if err != nil {
			return nil, fmt.Errorf("failed lean task network serialization for job %s: %w", job.ID, err)
		}

		err = s.taskQueueRepo.Put(ctx, s.cfg.IntakeQueueName, taskspb.HttpMethod_POST, routes.Full(routes.GCSListFiles), payloadBytes, 0)
		if err != nil {
			return nil, fmt.Errorf("queue execution aborted at task dispatch phase for job %s: %w", job.ID, err)
		}

		jobStatuses = append(jobStatuses, jobStatus)
	}

	return &jobStatuses, nil
}
