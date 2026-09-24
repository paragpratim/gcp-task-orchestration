package models

import (
	"fmt"
	"time"
)

// BigQueryFileType represents the supported file types for BigQuery ingestion.
type BigQueryFileType string

const (
	FileTypeCSV             BigQueryFileType = "CSV"
	FileTypeAvro            BigQueryFileType = "AVRO"
	FileTypeJSON            BigQueryFileType = "NEWLINE_DELIMITED_JSON"
	FileTypeDatastoreBackup BigQueryFileType = "DATASTORE_BACKUP"
	FileTypeGoogleSheets    BigQueryFileType = "GOOGLE_SHEETS"
	FileTypeBigtable        BigQueryFileType = "BIGTABLE"
	FileTypeParquet         BigQueryFileType = "PARQUET"
	FileTypeORC             BigQueryFileType = "ORC"
	FileTypeTFSavedModel    BigQueryFileType = "ML_TF_SAVED_MODEL"
	FileTypeXGBoostBooster  BigQueryFileType = "ML_XGBOOST_BOOSTER"
	FileTypeIceberg         BigQueryFileType = "ICEBERG"
)

// IsValid checks if the BigQueryFileType is one of the supported types.
func (t BigQueryFileType) IsValid() bool {
	switch t {
	case FileTypeCSV, FileTypeAvro, FileTypeJSON, FileTypeDatastoreBackup,
		FileTypeGoogleSheets, FileTypeBigtable, FileTypeParquet, FileTypeORC,
		FileTypeTFSavedModel, FileTypeXGBoostBooster, FileTypeIceberg:
		return true
	default:
		return false
	}
}

func (t BigQueryFileType) String() string {
	return string(t)
}

// Validate checks if the SourceDefinition has valid fields.
func (s SourceDefinition) Validate() error {
	if s.FileType == "" {
		return fmt.Errorf("source.file_type is required")
	}
	if !s.FileType.IsValid() {
		return fmt.Errorf("source.file_type %q is invalid: supported values are CSV, AVRO, NEWLINE_DELIMITED_JSON, DATASTORE_BACKUP, GOOGLE_SHEETS, BIGTABLE, PARQUET, ORC, ML_TF_SAVED_MODEL, ML_XGBOOST_BOOSTER, ICEBERG", s.FileType)
	}
	return nil
}

// SourceDefinition represents the source configuration for a BigQuery ingestion job.
type SourceDefinition struct {
	BucketName  string           `json:"bucket_name" firestore:"bucket_name"`
	Prefix      string           `json:"prefix,omitempty" firestore:"prefix,omitempty"`
	FilePattern string           `json:"file_pattern,omitempty" firestore:"file_pattern,omitempty"`
	FileType    BigQueryFileType `json:"file_type,omitempty" firestore:"file_type,omitempty"`
}

// TargetDefinition represents the target configuration for a BigQuery ingestion job.
type TargetDefinition struct {
	ProjectID string `json:"project_id" firestore:"project_id"`
	DatasetID string `json:"dataset_id" firestore:"dataset_id"`
	TableName string `json:"table_name" firestore:"table_name"`
}

// IntakeJobDefinition is the Firestore-backed model used to persist intake job metadata.
type IntakeJobDefinition struct {
	ID          string           `json:"id" firestore:"id"`
	Name        string           `json:"name" firestore:"name"`
	Description string           `json:"description,omitempty" firestore:"description,omitempty"`
	Source      SourceDefinition `json:"source" firestore:"source"`
	Target      TargetDefinition `json:"target" firestore:"target"`
	CreatedAt   time.Time        `json:"created_at" firestore:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" firestore:"updated_at"`
	Metadata    map[string]any   `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}
