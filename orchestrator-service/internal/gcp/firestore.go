package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type DocumentRepository[T any] interface {
	Put(ctx context.Context, collection, id string, data T) error
	Get(ctx context.Context, collection, id string) (*T, error)
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
		return nil, fmt.Errorf("failed to Get Document from Firestore: %w", err)
	}
	var data T
	if err := doc.DataTo(&data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall Document: %w", err)
	}
	return &data, nil
}

func (c *FirestoreRepository[T]) Close() error {
	return c.firestoreClient.Close()
}
