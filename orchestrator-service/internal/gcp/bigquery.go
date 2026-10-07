package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
)

// DataRepository defines the interface for interacting with Google BigQuery.
type DataRepository interface {
	CreateGCSLoadJob(ctx context.Context, datasetID, tableID string, gcsURIs []string, format bigquery.DataFormat) (string, error)
	CheckJobStatus(ctx context.Context, jobID string) (*bigquery.JobStatus, error)
	DatasetRegionAvailable(ctx context.Context, datasetID string) (string, error)
	CreateTable(ctx context.Context, datasetID, tableID string, schema bigquery.Schema) error
	Close() error
}

// BigQueryRepository is a concrete implementation of the DataRepository interface for Google BigQuery.
type BigQueryRepository struct {
	bqClient  *bigquery.Client
	projectID string
	region    string
}

// NewBigQueryRepository creates a new instance of BigQueryRepository.
func NewBigQueryRepository(ctx context.Context, env string, projectID string, region string) (*BigQueryRepository, error) {
	// Local Client without ADC
	if env == "local" {
		client, err := bigquery.NewClient(ctx, projectID, option.WithoutAuthentication())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize native BigQuery client: %w", err)
		}
		return &BigQueryRepository{
			bqClient:  client,
			projectID: projectID,
			region:    region,
		}, nil
	}

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize native BigQuery client: %w", err)
	}
	return &BigQueryRepository{
		bqClient:  client,
		projectID: projectID,
		region:    region,
	}, nil
}

// CreateGCSLoadJob creates a BigQuery load job to load data from GCS into a specified table.
func (r *BigQueryRepository) CreateGCSLoadJob(ctx context.Context, datasetID, tableID string, gcsURIs []string, format bigquery.DataFormat) (string, error) {
	gcsRef := bigquery.NewGCSReference(gcsURIs...)
	gcsRef.SourceFormat = format
	gcsRef.AutoDetect = true // Dynamically extracts schemas if not explicitly provided in the file

	tableRef := r.bqClient.Dataset(datasetID).Table(tableID)
	loader := tableRef.LoaderFrom(gcsRef)

	// Fire and forget: starts the load job asynchronously on GCP infrastructure
	job, err := loader.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to initiate BigQuery load job from %v: %w", gcsURIs, err)
	}

	return job.ID(), nil
}

// CheckJobStatus polls the current execution and error state of an active or finished job.
func (r *BigQueryRepository) CheckJobStatus(ctx context.Context, jobID string) (*bigquery.JobStatus, error) {
	job, err := r.bqClient.JobFromIDLocation(ctx, jobID, r.region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BigQuery job instance for ID %s: %w", jobID, err)
	}

	status, err := job.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BigQuery status metrics for job ID %s: %w", jobID, err)
	}

	return status, nil
}

// DatasetRegionAvailable verifies a target dataset exists and extracts its geographic region location (e.g. "US", "EU", "asia-northeast1").
func (r *BigQueryRepository) DatasetRegionAvailable(ctx context.Context, datasetID string) (string, error) {
	meta, err := r.bqClient.Dataset(datasetID).Metadata(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to query dataset location metadata for %s: %w", datasetID, err)
	}

	return meta.Location, nil
}

// CreateTable provisions an empty native structural table inside an existing dataset using a specified schema definition.
func (r *BigQueryRepository) CreateTable(ctx context.Context, datasetID, tableID string, schema bigquery.Schema) error {
	tableRef := r.bqClient.Dataset(datasetID).Table(tableID)

	err := tableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: schema,
	})
	if err != nil {
		return fmt.Errorf("failed to create BigQuery native table %s.%s: %w", datasetID, tableID, err)
	}

	return nil
}

// Close explicitly releases connections in the BigQuery client transport layers.
func (r *BigQueryRepository) Close() error {
	return r.bqClient.Close()
}
