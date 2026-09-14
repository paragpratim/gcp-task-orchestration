package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type Repository[T any] interface {
	Put(ctx context.Context, collection, id string, data T) error
	Get(ctx context.Context, collection, id string) (*T, error)
	Close() error
}

type Client[T any] struct {
	firestoreClient *firestore.Client
}

func NewClient[T any](ctx context.Context, projectId string) (*Client[T], error) {
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to Create firestore client: %w", err)
	}
	return &Client[T]{firestoreClient: client}, nil
}

func (c *Client[T]) Put(ctx context.Context, collection, id string, data T) error {
	_, err := c.firestoreClient.Collection(collection).Doc(id).Set(ctx, data, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to Put Document to Firestore: %w", err)
	}
	return nil
}

func (c *Client[T]) Get(ctx context.Context, collection, id string) (*T, error) {
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

func (c *Client[T]) Close() error {
	return c.firestoreClient.Close()
}
