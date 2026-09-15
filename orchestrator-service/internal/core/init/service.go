package init

import (
	"context"
	"orchestrator/internal/gcp"
	"orchestrator/internal/logger"
)

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

func (s *Service) HealthCheck(ctx context.Context) {
	logger.Debug("Service is Up and Running")
}
