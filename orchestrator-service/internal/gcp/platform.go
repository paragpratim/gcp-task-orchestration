package gcp

import (
	"context"
	"fmt"

	"orchestrator/internal/models"
)

// Platform is a struct that holds references to various GCP repositories used in the orchestrator.
type Platform struct {
	Firestore    *FirestoreRepository[any]
	Jobs         *FirestoreRepository[models.JobDefinition]
	JobStatus    *FirestoreRepository[models.JobStatus]
	JobStatusLog *FirestoreRepository[models.JobStatusLog]
	CloudTasks   *CloudTasksRepository
	Storage      *StorageRepository
	BigQuery     *BigQueryRepository
}

// NewPlatform initializes and returns a new Platform instance with the provided configuration.
func NewPlatform(ctx context.Context, env, projectID, region, firestoreDB, baseURL, saEmail string) (*Platform, error) {
	// Initialize Firestore repository
	fsRepo, err := NewFirestoreRepository[any](ctx, env, projectID, firestoreDB)
	if err != nil {
		return nil, fmt.Errorf("failed firestore admin: %w", err)
	}

	// Initialize Cloud Tasks repository
	tasksRepo, err := NewCloudTasksRepository(ctx, env, projectID, region, baseURL, saEmail)
	if err != nil {
		_ = fsRepo.Close()
		return nil, fmt.Errorf("failed tasks admin: %w", err)
	}

	// Initialize Storage repository
	gcsRepo, err := NewStorageRepository(ctx, env)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		return nil, fmt.Errorf("failed storage admin: %w", err)
	}

	// Initialize BigQuery repository
	bqRepo, err := NewBigQueryRepository(ctx, env, projectID, region)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		_ = gcsRepo.Close()
		return nil, fmt.Errorf("failed BigQuery admin: %w", err)
	}

	return &Platform{
		Firestore:    fsRepo,
		Jobs:         NewTypedFirestoreRepository[models.JobDefinition](fsRepo.firestoreClient),
		JobStatus:    NewTypedFirestoreRepository[models.JobStatus](fsRepo.firestoreClient),
		JobStatusLog: NewTypedFirestoreRepository[models.JobStatusLog](fsRepo.firestoreClient),
		CloudTasks:   tasksRepo,
		Storage:      gcsRepo,
		BigQuery:     bqRepo,
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
