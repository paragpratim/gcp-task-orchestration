package dataflow

import "orchestrator/internal/gcp"

type Service struct {
	firestoreRepo gcp.DocumentRepository[any]
	tasksRepo     gcp.TaskRepository
}

func NewService(fs gcp.DocumentRepository[any], tasks gcp.TaskRepository) *Service {
	return &Service{
		firestoreRepo: fs,
		tasksRepo:     tasks,
	}
}
