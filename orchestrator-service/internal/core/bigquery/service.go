package bigquery

import "orchestrator/internal/gcp"

type Service struct {
	firestoreRepo gcp.DocumentRepository[any]
	tasksRepo     gcp.TaskRepository
	bigQueryRepo  gcp.DataRepository
}

func NewService(firestoreRepo gcp.DocumentRepository[any], tasksRepo gcp.TaskRepository, bigQueryRepo gcp.DataRepository) *Service {
	return &Service{
		firestoreRepo: firestoreRepo,
		tasksRepo:     tasksRepo,
		bigQueryRepo:  bigQueryRepo,
	}
}
