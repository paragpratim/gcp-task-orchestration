package gcp

import (
	"context"
	"fmt"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskRepository interface {
	Put(ctx context.Context, queuePath string, method taskspb.HttpMethod, path string, payload []byte, delay time.Duration) error
	Close() error
}
type CloudTasksRepository struct {
	tasksClient         *cloudtasks.Client
	baseURL             string
	serviceAccountEmail string
}

func NewCloudTasksRepository(ctx context.Context, env string, baseURL string, serviceAccountEmail string) (*CloudTasksRepository, error) {
	//Local Emulator Setup
	if env == "local" {
		// Set the environment variable for the Cloud Tasks emulator
		client, err := cloudtasks.NewClient(ctx,
			option.WithEndpoint("cloud-tasks-emulator:8123"),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			option.WithoutAuthentication(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create task client for emulator: %w", err)
		}
		return &CloudTasksRepository{
			tasksClient:         client,
			baseURL:             baseURL,
			serviceAccountEmail: serviceAccountEmail,
		}, nil
	}

	// Production Setup
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create task client: %w", err)
	}
	return &CloudTasksRepository{
		tasksClient:         client,
		baseURL:             baseURL,
		serviceAccountEmail: serviceAccountEmail,
	}, nil
}

func (c *CloudTasksRepository) Put(ctx context.Context, queuePath string, method taskspb.HttpMethod, path string, payload []byte, delay time.Duration) error {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)

	task := &taskspb.Task{
		MessageType: &taskspb.Task_HttpRequest{
			HttpRequest: &taskspb.HttpRequest{
				HttpMethod: method,
				Url:        fullURL,
				Body:       payload,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
					OidcToken: &taskspb.OidcToken{
						ServiceAccountEmail: c.serviceAccountEmail,
					},
				},
			},
		},
	}

	if delay > 0 {
		scheduleTime := time.Now().Add(delay)
		task.ScheduleTime = timestamppb.New(scheduleTime)
	}

	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task:   task,
	}

	_, err := c.tasksClient.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to dispatch task to queue: %w", err)
	}
	return nil
}

func (c *CloudTasksRepository) Close() error {
	return c.tasksClient.Close()
}
