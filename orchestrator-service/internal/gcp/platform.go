package gcp

import (
	"context"
	"fmt"

	"orchestrator/internal/models"
)

type Platform struct {
	Firestore  *FirestoreRepository[any]
	IntakeJobs *FirestoreRepository[models.IntakeJobDefinition]
	JobStatus  *FirestoreRepository[models.JobStatus]
	CloudTasks *CloudTasksRepository
	Storage    *StorageRepository
	BigQuery   *BigQueryRepository
}

func NewPlatform(ctx context.Context, env, projectID, baseURL, saEmail string) (*Platform, error) {
	fsRepo, err := NewFirestoreRepository[any](ctx, env, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed firestore intake: %w", err)
	}

	tasksRepo, err := NewCloudTasksRepository(ctx, env, baseURL, saEmail)
	if err != nil {
		_ = fsRepo.Close()
		return nil, fmt.Errorf("failed tasks intake: %w", err)
	}

	gcsRepo, err := NewStorageRepository(ctx, env)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		return nil, fmt.Errorf("failed storage intake: %w", err)
	}

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
