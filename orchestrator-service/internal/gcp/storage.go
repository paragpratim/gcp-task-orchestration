package gcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ObjectRepository defines the interface for interacting with object storage.
type ObjectRepository interface {
	ListObjects(ctx context.Context, bucketName, prefix, filePattern string) ([]string, error)
	MoveFile(ctx context.Context, bucketName, srcObject, dstObject string) error
	Close() error
}

// StorageRepository is a concrete implementation of ObjectRepository for Google Cloud Storage.
type StorageRepository struct {
	storageClient *storage.Client
}

// NewStorageRepository creates a new instance of StorageRepository.
func NewStorageRepository(ctx context.Context, env string) (*StorageRepository, error) {
	// Local Client without ADC
	if env == "local" {
		client, err := storage.NewClient(ctx, option.WithoutAuthentication())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize native storage client: %w", err)
		}
		return &StorageRepository{
			storageClient: client,
		}, nil
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize native storage client: %w", err)
	}
	return &StorageRepository{
		storageClient: client,
	}, nil
}

// ListObjects retrieves a list of object names from the specified bucket that match the given prefix and file pattern.
// Only returns objects in the immediate directory, not subdirectories.
// Prefix and filePattern can be blank.
// Prefix may have leading/trailing slashes; they are stripped automatically.
func (r *StorageRepository) ListObjects(ctx context.Context, bucketName, prefix string, filePattern string) ([]string, error) {
	var objects []string

	// Normalize prefix: strip leading and trailing slashes
	prefix = strings.Trim(prefix, "/")

	// Build the MatchGlob pattern
	// GCS MatchGlob must include the full path from bucket root, not just the filename
	// Wildcard * does not match / characters, so it only matches within the immediate directory
	matchGlob := ""
	if filePattern != "" {
		if prefix != "" {
			matchGlob = prefix + "/" + filePattern
		} else {
			matchGlob = filePattern
		}
	}

	// Create an object query filter targeting the folder/prefix path and applying the file pattern for matching
	query := &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
		MatchGlob: matchGlob,
	}
	it := r.storageClient.Bucket(bucketName).Objects(ctx, query)

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects under bucket %s with prefix %s: %w", bucketName, prefix, err)
		}
		objects = append(objects, attrs.Name)
	}

	return objects, nil
}

// MoveFile moves an object from the source path to the destination path within the same bucket.
func (r *StorageRepository) MoveFile(ctx context.Context, bucketName, srcObject, dstObject string) error {
	bucketRef := r.storageClient.Bucket(bucketName)
	srcRef := bucketRef.Object(srcObject)

	// Configure the destination move target inside the exact same bucket context
	destination := storage.MoveObjectDestination{
		Object: dstObject,
	}

	// Trigger high-efficiency intra-bucket metadata and byte movement natively on GCP infrastructure
	if _, err := srcRef.Move(ctx, destination); err != nil {
		return fmt.Errorf("failed to move object from source [%s] to destination [%s] in bucket %s: %w", srcObject, dstObject, bucketName, err)
	}

	return nil
}

// Close releases any resources held by the StorageRepository.
func (r *StorageRepository) Close() error {
	return r.storageClient.Close()
}
