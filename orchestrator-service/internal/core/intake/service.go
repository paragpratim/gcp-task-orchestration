package intake

import (
	"encoding/json"
	"fmt"
	"orchestrator/internal/gcp"
	"orchestrator/internal/logger"
	"orchestrator/internal/models"
	"time"

	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Config struct {
	JobsCollection   string
	StatusCollection string
	QueueName        string
}

type Service struct {
	firestoreRepo gcp.DocumentRepository[any]
	taskQueueRepo gcp.TaskRepository
	cfg           Config
}

func NewService(fs gcp.DocumentRepository[any], tasks gcp.TaskRepository, cfg Config) *Service {
	return &Service{
		firestoreRepo: fs,
		taskQueueRepo: tasks,
		cfg:           cfg,
	}
}

func (s *Service) HealthCheck(ctx *gin.Context) error {
	logger.Debug("Service is Up and Running")
	return nil
}

func (s *Service) CreateIntakeJob(ctx *gin.Context, jobDefinition models.IntakeJobDefinition) (*models.IntakeJobDefinition, error) {
	if jobDefinition.ID == "" {
		jobDefinition.ID = uuid.New().String()
	}

	jobDefinition.CreatedAt = time.Now().UTC()
	jobDefinition.UpdatedAt = time.Now().UTC()

	err := s.firestoreRepo.Put(ctx, s.cfg.JobsCollection, jobDefinition.ID, jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to persist initial intake state: %w", err)
	}
	return &jobDefinition, nil
}

func (s *Service) UpdateIntakeJob(ctx *gin.Context, jobDefinition models.IntakeJobDefinition) (*models.IntakeJobDefinition, error) {

	if jobDefinition.ID == "" {
		return nil, fmt.Errorf("cannot update a job without a valid transaction identifier 'id'")
	}

	existingJob, err := s.firestoreRepo.Get(ctx, s.cfg.JobsCollection, jobDefinition.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve job context before mutation: %w", err)
	}
	if existingJob == nil {
		return nil, fmt.Errorf("job target with id %s does not exist", jobDefinition.ID)
	}

	jobDefinition.UpdatedAt = time.Now().UTC()

	err = s.firestoreRepo.Put(ctx, s.cfg.JobsCollection, jobDefinition.ID, jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to overwrite persistent intake state: %w", err)
	}

	return &jobDefinition, nil
}

func (s *Service) DeleteIntakeJob(ctx *gin.Context, id string) error {

	existingJob, err := s.firestoreRepo.Get(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return fmt.Errorf("failed to locate target job before removal tracking: %w", err)
	}
	if existingJob == nil {
		return fmt.Errorf("cannot delete job: tracking target with id %s was not found", id)
	}

	err = s.firestoreRepo.Delete(ctx, s.cfg.JobsCollection, id)
	if err != nil {
		return fmt.Errorf("failed to completely purge intake job record: %w", err)
	}

	return nil
}

func (s *Service) QueueActiveJobs(ctx *gin.Context, jobDefinition models.IntakeJobDefinition) (*[]models.JobStatus, error) {
	var jobsToQueue []models.IntakeJobDefinition
	var jobStatuses []models.JobStatus

	if jobDefinition.ID != "" {
		retrieved, err := s.firestoreRepo.Get(ctx, s.cfg.JobsCollection, jobDefinition.ID)
		if err != nil {
			return nil, fmt.Errorf("failed verifying dispatch target: %w", err)
		}

		if retrieved != nil {
			if targetJob, ok := any(retrieved).(models.IntakeJobDefinition); ok {
				jobsToQueue = append(jobsToQueue, targetJob)
			} else if targetJob, ok := any(retrieved).(*models.IntakeJobDefinition); ok && targetJob != nil {
				jobsToQueue = append(jobsToQueue, *targetJob)
			} else {
				return nil, fmt.Errorf("repository returned an unexpected type structure for job %s", jobDefinition.ID)
			}
		}
	} else {
		allJobs, err := s.firestoreRepo.GetAll(ctx, s.cfg.JobsCollection)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate jobs for queue migration: %w", err)
		}

		if allJobs != nil {
			for _, item := range *allJobs {
				if targetJob, ok := item.(models.IntakeJobDefinition); ok {
					jobsToQueue = append(jobsToQueue, targetJob)
				} else if targetJob, ok := item.(*models.IntakeJobDefinition); ok && targetJob != nil {
					jobsToQueue = append(jobsToQueue, *targetJob)
				}
			}
		}
	}

	for _, job := range jobsToQueue {
		taskPayload, err := json.Marshal(job)
		if err != nil {
			return nil, fmt.Errorf("failed task data serialization for job %s: %w", job.ID, err)
		}

		err = s.taskQueueRepo.Put(ctx, s.cfg.QueueName, taskspb.HttpMethod_POST, "", taskPayload, 0)
		if err != nil {
			return nil, fmt.Errorf("queue execution aborted at job %s: %w", job.ID, err)
		}

		jobStatus := models.JobStatus{
			JobID:     job.ID,
			Status:    models.StatusQueued,
			UpdatedAt: time.Now().UTC(),
			Metadata:  map[string]any{},
		}

		jobStatuses = append(jobStatuses, jobStatus)

		err = s.firestoreRepo.Put(ctx, s.cfg.StatusCollection, job.ID, jobStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to persist job status for job %s: %w", job.ID, err)
		}
	}

	return &jobStatuses, nil
}
