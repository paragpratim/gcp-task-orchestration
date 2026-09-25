package gcp

import (
	"context"
	"fmt"

	"orchestrator/internal/models"
)

// Platform is a struct that holds references to various GCP repositories used in the orchestrator.
type Platform struct {
	Firestore  *FirestoreRepository[any]
	IntakeJobs *FirestoreRepository[models.IntakeJobDefinition]
	JobStatus  *FirestoreRepository[models.JobStatus]
	CloudTasks *CloudTasksRepository
	Storage    *StorageRepository
	BigQuery   *BigQueryRepository
}

// NewPlatform initializes and returns a new Platform instance with the provided configuration.
func NewPlatform(ctx context.Context, env, projectID, baseURL, saEmail string) (*Platform, error) {
	// Initialize Firestore repository
	fsRepo, err := NewFirestoreRepository[any](ctx, env, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed firestore intake: %w", err)
	}

	// Initialize Cloud Tasks repository
	tasksRepo, err := NewCloudTasksRepository(ctx, env, baseURL, saEmail)
	if err != nil {
		_ = fsRepo.Close()
		return nil, fmt.Errorf("failed tasks intake: %w", err)
	}

	// Initialize Storage repository
	gcsRepo, err := NewStorageRepository(ctx, env)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		return nil, fmt.Errorf("failed storage intake: %w", err)
	}

	// Initialize BigQuery repository
	bqRepo, err := NewBigQueryRepository(ctx, env, projectID)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		_ = gcsRepo.Close()
		return nil, fmt.Errorf("failed BigQuery intake: %w", err)
	}

	return &Platform{
		Firestore:  fsRepo,
		IntakeJobs: NewTypedFirestoreRepository[models.IntakeJobDefinition](fsRepo.firestoreClient),
		JobStatus:  NewTypedFirestoreRepository[models.JobStatus](fsRepo.firestoreClient),
		CloudTasks: tasksRepo,
		Storage:    gcsRepo,
		BigQuery:   bqRepo,
	}, nil
}

// Close releases any resources held by the Platform, such as Firestore, Cloud Tasks, Storage, and BigQuery clients.
func (p *Platform) Close() {
	if p.Firestore != nil {
		_ = p.Firestore.Close()
	}
	if p.CloudTasks != nil {
		_ = p.CloudTasks.Close()
	}
	if p.Storage != nil {
		_ = p.Storage.Close()
	}
	if p.BigQuery != nil {
		_ = p.BigQuery.Close()
	}
}
