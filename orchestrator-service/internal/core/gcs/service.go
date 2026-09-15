package gcs

import "orchestrator/internal/gcp"

type Service struct {
	firestoreRepo gcp.DocumentRepository[any]
	tasksRepo     gcp.TaskRepository
	storageRepo   gcp.ObjectRepository
}

func NewService(fs gcp.DocumentRepository[any], tasks gcp.TaskRepository, storage gcp.ObjectRepository) *Service {
	return &Service{
		firestoreRepo: fs,
		tasksRepo:     tasks,
		storageRepo:   storage,
	}
}
