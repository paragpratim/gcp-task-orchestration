package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

type DataRepository interface {
	CreateGCSLoadJob(ctx context.Context, datasetID, tableID, gcsURI string, format bigquery.DataFormat) (string, error)
	CheckJobStatus(ctx context.Context, jobID string) (*bigquery.JobStatus, error)
	DatasetRegionAvailable(ctx context.Context, datasetID string) (string, error)
	CreateTable(ctx context.Context, datasetID, tableID string, schema bigquery.Schema) error
	Close() error
}

type BigQueryRepository struct {
	bqClient  *bigquery.Client
	projectID string
}

func NewBigQueryRepository(ctx context.Context, projectID string) (*BigQueryRepository, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize native bigquery client: %w", err)
	}
	return &BigQueryRepository{
		bqClient:  client,
		projectID: projectID,
	}, nil
}

func (r *BigQueryRepository) CreateGCSLoadJob(ctx context.Context, datasetID, tableID, gcsURI string, format bigquery.DataFormat) (string, error) {
	gcsRef := bigquery.NewGCSReference(gcsURI)
	gcsRef.SourceFormat = format
	gcsRef.AutoDetect = true // Dynamically extracts schemas if not explicitly provided in the file

	tableRef := r.bqClient.Dataset(datasetID).Table(tableID)
	loader := tableRef.LoaderFrom(gcsRef)

	// Fire and forget: starts the load job asynchronously on GCP infrastructure
	job, err := loader.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to initiate Bigquery load job from %s: %w", gcsURI, err)
	}

	return job.ID(), nil
}

// CheckJobStatus polls the current execution and error state of an active or finished job.
func (r *BigQueryRepository) CheckJobStatus(ctx context.Context, jobID string) (*bigquery.JobStatus, error) {
	job, err := r.bqClient.JobFromID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Bigquery job instance for ID %s: %w", jobID, err)
	}

	status, err := job.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Bigquery status metrics for job ID %s: %w", jobID, err)
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
		return fmt.Errorf("failed to create Bigquery native table %s.%s: %w", datasetID, tableID, err)
	}

	return nil
}

// Close explicitly releases connections in the BigQuery client transport layers.
func (r *BigQueryRepository) Close() error {
	return r.bqClient.Close()
}
