package gcp

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DocumentRepository[T any] interface {
	Put(ctx context.Context, collection, id string, data T) error
	Get(ctx context.Context, collection, id string) (*T, error)
	GetAll(ctx context.Context, collection string) (*[]T, error)
	Delete(ctx context.Context, collection, id string) error
	Close() error
}

type FirestoreRepository[T any] struct {
	firestoreClient *firestore.Client
}

func NewFirestoreRepository[T any](ctx context.Context, projectId string) (*FirestoreRepository[T], error) {
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to Create firestore client: %w", err)
	}
	return &FirestoreRepository[T]{firestoreClient: client}, nil
}

func NewTypedFirestoreRepository[T any](client *firestore.Client) *FirestoreRepository[T] {
	return &FirestoreRepository[T]{firestoreClient: client}
}

func (c *FirestoreRepository[T]) Put(ctx context.Context, collection, id string, data T) error {
	_, err := c.firestoreClient.Collection(collection).Doc(id).Set(ctx, data, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to Put Document to Firestore: %w", err)
	}
	return nil
}

func (c *FirestoreRepository[T]) Get(ctx context.Context, collection, id string) (*T, error) {
	doc, err := c.firestoreClient.Collection(collection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to Get Document from Firestore: %w", err)
	}
	var data T
	if err := doc.DataTo(&data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall Document: %w", err)
	}
	return &data, nil
}

func (c *FirestoreRepository[T]) GetAll(ctx context.Context, collection string) (*[]T, error) {
	iter := c.firestoreClient.Collection(collection).Documents(ctx)
	defer iter.Stop()

	var results []T
	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to iterate Documents from Firestore: %w", err)
		}
		var data T
		if err := doc.DataTo(&data); err != nil {
			return nil, fmt.Errorf("failed to unmarshall Document: %w", err)
		}
		results = append(results, data)
	}
	return &results, nil
}

func (c *FirestoreRepository[T]) Delete(ctx context.Context, collection, id string) error {
	_, err := c.firestoreClient.Collection(collection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to Delete Document from Firestore: %w", err)
	}
	return nil
}

func (c *FirestoreRepository[T]) Close() error {
	return c.firestoreClient.Close()
}
