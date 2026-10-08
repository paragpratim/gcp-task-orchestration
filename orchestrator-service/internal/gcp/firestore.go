package gcp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// DocumentRepository is a generic interface for interacting with a Firestore database.
type DocumentRepository[T any] interface {
	Put(ctx context.Context, collection, id string, data T) (*T, error)
	Get(ctx context.Context, collection, id string) (*T, error)
	GetAll(ctx context.Context, collection string) (*[]T, error)
	Delete(ctx context.Context, collection, id string) error
	Close() error
}

// FirestoreRepository is a generic implementation of the DocumentRepository interface for Firestore.
type FirestoreRepository[T any] struct {
	firestoreClient *firestore.Client
}

// NewFirestoreRepository creates a new instance of FirestoreRepository.
func NewFirestoreRepository[T any](ctx context.Context, env string, projectId string, firestoreDB string) (*FirestoreRepository[T], error) {
	//Local Emulator Setup
	if env == "local" {
		// Set the environment variable for the Cloud Tasks emulator
		client, err := firestore.NewClient(ctx, projectId,
			option.WithEndpoint("firestore-emulator:8081"),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			option.WithoutAuthentication(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create firestore client for emulator: %w", err)
		}
		return &FirestoreRepository[T]{
			firestoreClient: client,
		}, nil
	}

	client, err := firestore.NewClientWithDatabase(ctx, projectId, firestoreDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return &FirestoreRepository[T]{firestoreClient: client}, nil
}

// NewTypedFirestoreRepository creates a new instance of FirestoreRepository with an existing Firestore client.
// This is useful when you already have a Firestore client and want to create a repository for a specific type.
func NewTypedFirestoreRepository[T any](client *firestore.Client) *FirestoreRepository[T] {
	return &FirestoreRepository[T]{firestoreClient: client}
}

// Put adds or updates a document in the specified collection with the given ID and data.
func (c *FirestoreRepository[T]) Put(ctx context.Context, collection, id string, data T) (*T, error) {
	// If id is empty, generate a new document reference with a random ID
	var docRef *firestore.DocumentRef
	if id == "" {
		docRef = c.firestoreClient.Collection(collection).NewDoc()
		id = docRef.ID
	} else {
		docRef = c.firestoreClient.Collection(collection).Doc(id)
	}

	// Convert the struct to a map using the native Firestore mapping function.
	mapData, err := structToFirestoreMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to map generic struct natively: %w", err)
	}
	// Ensure the ID is included in the map data.
	mapData["id"] = id

	// Use MergeAll to merge the new data with existing data in Firestore.
	_, err = docRef.Set(ctx, mapData, firestore.MergeAll)
	if err != nil {
		return nil, fmt.Errorf("failed to Put Document to Firestore: %w", err)
	}

	// Retrieve the document after setting it to ensure we return the latest data.
	doc, err := docRef.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Document after Put: %w", err)
	}
	var result T
	if err := doc.DataTo(&result); err != nil {
		return nil, fmt.Errorf("failed to unmarshall Document after Put: %w", err)
	}

	return &result, nil
}

// Get retrieves a document from the specified collection with the given ID and unmarshals it into the specified type.
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

// GetAll retrieves all documents from the specified collection and unmarshals them into the specified type.
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

// Delete removes a document from the specified collection with the given ID.
func (c *FirestoreRepository[T]) Delete(ctx context.Context, collection, id string) error {
	_, err := c.firestoreClient.Collection(collection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete Document from Firestore: %w", err)
	}
	return nil
}

// Close explicitly releases connections in the Firestore client transport layers.
func (c *FirestoreRepository[T]) Close() error {
	return c.firestoreClient.Close()
}

// structToFirestoreMap converts a struct to a map[string]any, using the "firestore" struct tags to determine the field names.
func structToFirestoreMap(obj any) (map[string]any, error) {
	v := reflect.Indirect(reflect.ValueOf(obj))
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("provided data is not a struct")
	}

	t := v.Type()
	res := make(map[string]any)

	for i := 0; i < v.NumField(); i++ {
		tag := t.Field(i).Tag.Get("firestore")
		if tag == "" || tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		field := v.Field(i)

		// Check omitempty short-circuit
		if len(parts) > 1 && parts[1] == "omitempty" && field.IsZero() {
			continue
		}

		res[parts[0]] = field.Interface()
	}
	return res, nil
}
