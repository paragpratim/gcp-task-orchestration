package gcp

import (
	"context"
	"fmt"
)

type Platform struct {
	Firestore  *FirestoreRepository[any]
	CloudTasks *CloudTasksRepository
	Storage    *StorageRepository
	BigQuery   *BigQueryRepository
}

func NewPlatform(ctx context.Context, projectID, baseURL, saEmail string) (*Platform, error) {
	fsRepo, err := NewFirestoreRepository[any](ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed firestore init: %w", err)
	}

	tasksRepo, err := NewCloudTasksRepository(ctx, baseURL, saEmail)
	if err != nil {
		_ = fsRepo.Close()
		return nil, fmt.Errorf("failed tasks init: %w", err)
	}

	gcsRepo, err := NewStorageRepository(ctx)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		return nil, fmt.Errorf("failed storage init: %w", err)
	}

	bqRepo, err := NewBigQueryRepository(ctx, projectID)
	if err != nil {
		_ = fsRepo.Close()
		_ = tasksRepo.Close()
		_ = gcsRepo.Close()
		return nil, fmt.Errorf("failed BigQuery init: %w", err)
	}

	return &Platform{
		Firestore:  fsRepo,
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
