package models

import "time"

type SourceDefinition struct {
	BucketName  string `json:"bucket_name" firestore:"bucket_name"`
	Prefix      string `json:"prefix,omitempty" firestore:"prefix,omitempty"`
	FilePattern string `json:"file_pattern,omitempty" firestore:"file_pattern,omitempty"`
}

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
