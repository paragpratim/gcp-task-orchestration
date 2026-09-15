package gcp

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type ObjectRepository interface {
	ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error)
	MoveFile(ctx context.Context, bucketName, srcObject, dstObject string) error
	Close() error
}

type StorageRepository struct {
	storageClient *storage.Client
}

func NewStorageRepository(ctx context.Context) (*StorageRepository, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize native storage client: %w", err)
	}
	return &StorageRepository{
		storageClient: client,
	}, nil
}

func (r *StorageRepository) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {
	var objects []string

	// Create an object query filter targeting the folder/prefix path
	query := &storage.Query{Prefix: prefix}
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

func (r *StorageRepository) Close() error {
	return r.storageClient.Close()
}
