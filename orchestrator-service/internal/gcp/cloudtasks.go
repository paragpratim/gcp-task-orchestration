package gcp

import (
	"context"
	"fmt"
	"os"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TaskRepository defines the interface for interacting with a task queue.
type TaskRepository interface {
	Put(ctx context.Context, queuePath string, method taskspb.HttpMethod, path string, payload []byte, delay time.Duration) error
	Close() error
}

// CloudTasksRepository is a concrete implementation of the TaskRepository interface for Google Cloud Tasks.
type CloudTasksRepository struct {
	tasksClient         *cloudtasks.Client
	projectID           string
	region              string
	baseURL             string
	serviceAccountEmail string
}

// NewCloudTasksRepository creates a new instance of CloudTasksRepository.
func NewCloudTasksRepository(ctx context.Context, env string, projectID string, region string, baseURL string, serviceAccountEmail string) (*CloudTasksRepository, error) {
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
			projectID:           projectID,
			region:              region,
			baseURL:             baseURL,
			serviceAccountEmail: serviceAccountEmail,
		}, nil
	}

	// Production Setup
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create task client: %w", err)
	}
	// Resolve the base URL for Cloud Run if not provided
	baseURL, err = ResolveCloudRunURL(ctx, baseURL, projectID, region)
	if err != nil {
		// If resolving the Cloud Run URL fails, close the client and return an error
		_ = client.Close()
		return nil, fmt.Errorf("failed to resolve cloud run URL: %w", err)
	}
	return &CloudTasksRepository{
		tasksClient:         client,
		projectID:           projectID,
		region:              region,
		baseURL:             baseURL,
		serviceAccountEmail: serviceAccountEmail,
	}, nil
}

// Put enqueues a task into the specified Cloud Tasks queue with an optional delay.
func (c *CloudTasksRepository) Put(ctx context.Context, queueName string, method taskspb.HttpMethod, path string, payload []byte, delay time.Duration) error {
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
		scheduleTime := time.Now().UTC().Add(delay)
		task.ScheduleTime = timestamppb.New(scheduleTime)
	}

	fullyQualifiedQueuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", c.projectID, c.region, queueName)

	req := &taskspb.CreateTaskRequest{
		Parent: fullyQualifiedQueuePath,
		Task:   task,
	}

	_, err := c.tasksClient.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to dispatch task to queue: %w", err)
	}
	return nil
}

// Close explicitly releases connections in the Cloud Tasks client transport layers.
func (c *CloudTasksRepository) Close() error {
	return c.tasksClient.Close()
}

// ResolveCloudRunURL resolves the base URL for a Cloud Run service. If BaseURL is provided, it returns that.
// Otherwise, it attempts to resolve the URL from Cloud Run service metadata.
func ResolveCloudRunURL(ctx context.Context, BaseURL string, projectID string, region string) (string, error) {
	if BaseURL != "" {
		return BaseURL, nil
	}
	// If BaseURL is not provided, attempt to resolve the URL from Cloud Run service metadata
	kService := os.Getenv("K_SERVICE")
	if kService == "" {
		// No service name detected; we are running locally or in a standard docker container
		return "", nil
	}
	// Initialize the Cloud Run client
	runService, err := run.NewService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to initialize cloud run v2 client: %w", err)
	}
	// Construct the resource name for the Cloud Run service
	resourceName := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, region, kService)
	// Fetch the Cloud Run service details
	svc, err := runService.Projects.Locations.Services.Get(resourceName).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed fetching self routing configuration via v2 control plane: %w", err)
	}
	// Check if the service has a valid URL
	if svc == nil || svc.Uri == "" {
		return "", fmt.Errorf("gcp v2 control plane returned an empty ingress URI for service %s", kService)
	}
	return svc.Uri, nil
}
