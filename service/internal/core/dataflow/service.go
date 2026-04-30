package dataflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/paragpratim/gcp-task-orchestration/service/internal/models"
	df "google.golang.org/api/dataflow/v1b3"
	"google.golang.org/api/option"
)

// Service encapsulates the Dataflow API client.
type Service struct {
	client *df.Service
}

// NewService creates a new Dataflow service client.
func NewService(ctx context.Context) (*Service, error) {
	// By default, df.NewService uses Application Default Credentials.
	client, err := df.NewService(ctx, option.WithScopes(df.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create dataflow service: %w", err)
	}
	return &Service{client: client}, nil
}

// CheckJobActive checks if a job with the given name is currently active.
// It returns the job ID if found, a boolean indicating if it's active, and any error.
func (s *Service) CheckJobActive(projectID, region, jobName string) (string, bool, error) {
	// Filter by ACTIVE state to reduce payload size and focus on running jobs
	call := s.client.Projects.Locations.Jobs.List(projectID, region).Filter("ACTIVE")
	resp, err := call.Do()
	if err != nil {
		return "", false, fmt.Errorf("failed to list jobs: %w", err)
	}

	for _, job := range resp.Jobs {
		if job.Name == jobName {
			// Job is active
			return job.Id, true, nil
		}
	}
	return "", false, nil
}

// LaunchFlexJob submits a new Dataflow Flex Template job.
func (s *Service) LaunchFlexJob(req models.DataflowRequest) (*df.LaunchFlexTemplateResponse, error) {
	// Convert parameters from map[string]interface{} to map[string]string
	params := make(map[string]string)
	for k, v := range req.Parameters {
		params[k] = fmt.Sprintf("%v", v)
	}

	launchReq := &df.LaunchFlexTemplateRequest{
		LaunchParameter: &df.LaunchFlexTemplateParameter{
			JobName:    req.JobName,
			Parameters: params,
			Environment: &df.FlexTemplateRuntimeEnvironment{
				ServiceAccountEmail:   req.ServiceAccountEmail,
				TempLocation:          req.TempLocation,
				StagingLocation:       req.StagingLocation,
				NumWorkers:            int64(req.NumWorkers),
				MaxWorkers:            int64(req.MaxWorkers),
				IpConfiguration:       req.IPConfiguration,
				EnableStreamingEngine: req.EnableStreamingEngine,
				WorkerRegion:          req.Region,
				Zone:                  req.Zone,
			},
		},
	}

	// Set container image directly on ContainerSpec
	if req.ContainerImage != "" {
		launchReq.LaunchParameter.ContainerSpec = &df.ContainerSpec{
			Image: req.ContainerImage,
		}
	}

	call := s.client.Projects.Locations.FlexTemplates.Launch(req.ProjectID, req.Region, launchReq)
	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to launch flex template: %w", err)
	}

	return resp, nil
}

// StopJob attempts to stop a running job by draining or cancelling it.
func (s *Service) StopJob(projectID, region, jobID, stopMode string) (*df.Job, error) {
	var requestedState string
	switch strings.ToLower(stopMode) {
	case "drain":
		requestedState = "JOB_STATE_DRAINED"
	case "cancel":
		requestedState = "JOB_STATE_CANCELLED"
	default:
		// Default to drain if not specified
		requestedState = "JOB_STATE_DRAINED"
	}

	job := &df.Job{
		RequestedState: requestedState,
	}

	call := s.client.Projects.Locations.Jobs.Update(projectID, region, jobID, job)
	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update job state: %w", err)
	}

	return resp, nil
}
