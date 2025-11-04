package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang-simple-notes/model"
)

// TestMongoDBStorage tests the MongoDB storage implementation
// This test uses the shared MongoDB container from TestMain
func TestMongoDBStorage(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	ctx := context.Background()

	// Use the shared MongoDB container
	mongodbEndpoint := getSharedMongoURI()
	if mongodbEndpoint == "" {
		t.Skip("Shared MongoDB container not available")
	}

	// Add database and collection names to the URI
	dbName := "test_notes"
	collectionName := "test_notes"

	// Connect to the MongoDB container
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB container: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Logf("Warning: Failed to disconnect from MongoDB container: %v", err)
		}
	}(client, ctx)

	// Ping the MongoDB server to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB container: %v", err)
	}

	// Create a new MongoDB storage
	storage, err := NewMongoDBStorage(mongodbEndpoint, dbName, collectionName)
	if err != nil {
		t.Fatalf("Failed to create MongoDB storage: %v", err)
	}

	// Clean up the test collection before starting
	err = client.Database(dbName).Collection(collectionName).Drop(ctx)
	if err != nil {
		t.Logf("Warning: Failed to drop test collection: %v", err)
	}

	// Run the fixed storage tests
	testNoteStorage(t, storage, ctx)

	// Clean up after the test
	err = client.Database(dbName).Collection(collectionName).Drop(ctx)
	if err != nil {
		t.Logf("Warning: Failed to drop test collection: %v", err)
	}
}

// TestMongoDBStorageUnit tests the MongoDB storage implementation with unit tests
func TestMongoDBStorageUnit(t *testing.T) {
	// Create a mock implementation of NoteStorage that behaves like MongoDB
	storage := NewMockMongoDBStorage()

	// Run the fixed storage tests
	testNoteStorage(t, storage, context.Background())
}

// MockMongoDBStorage is a mock implementation of NoteStorage that behaves like MongoDB
type MockMongoDBStorage struct {
	notes map[string]*model.Note
}

// NewMockMongoDBStorage creates a new instance of MockMongoDBStorage
func NewMockMongoDBStorage() *MockMongoDBStorage {
	return &MockMongoDBStorage{
		notes: make(map[string]*model.Note),
	}
}

// Create adds a new note to the storage
func (s *MockMongoDBStorage) Create(ctx context.Context, note *model.Note) error {
	// Check if note with the same ID already exists
	if _, exists := s.notes[note.ID]; exists {
		return fmt.Errorf("note with ID %s already exists", note.ID)
	}

	// Store a copy of the note
	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID
func (s *MockMongoDBStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	note, exists := s.notes[id]
	if !exists {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *MockMongoDBStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	notes := make([]*model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

// Update updates an existing note
func (s *MockMongoDBStorage) Update(ctx context.Context, note *model.Note) error {
	if _, exists := s.notes[note.ID]; !exists {
		return ErrNoteNotFound
	}

	// Update the note with the current time
	note.UpdatedAt = time.Now()

	// Store a copy of the note
	s.notes[note.ID] = note
	return nil
}

// Delete removes a note from the storage
func (s *MockMongoDBStorage) Delete(ctx context.Context, id string) error {
	if _, exists := s.notes[id]; !exists {
		return ErrNoteNotFound
	}

	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage
func (s *MockMongoDBStorage) Close(ctx context.Context) error {
	// Nothing to close for mock storage
	return nil
}

// Additional MongoDB-specific tests could be added here
func TestMongoDBSpecificFeatures(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the shared MongoDB container
	mongodbEndpoint := getSharedMongoURI()
	if mongodbEndpoint == "" {
		t.Skip("Shared MongoDB container not available")
	}

	// Add database and collection names to the URI
	dbName := "test_notes"
	collectionName := "test_notes"

	// Connect to the MongoDB container
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB container: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Logf("Warning: Failed to disconnect from MongoDB container: %v", err)
		}
	}(client, ctx)

	// Ping the MongoDB server to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB container: %v", err)
	}

	// Clean up the test collection before starting
	err = client.Database(dbName).Collection(collectionName).Drop(ctx)
	if err != nil {
		t.Logf("Warning: Failed to drop test collection: %v", err)
	}

	// Create a new MongoDB storage
	storage, err := NewMongoDBStorage(mongodbEndpoint, dbName, collectionName)
	if err != nil {
		t.Fatalf("Failed to create MongoDB storage: %v", err)
	}
	CleanupCloseWithContext(t, ctx, storage)

	// Test that the unique index on ID works
	t.Run("UniqueIDIndex", func(t *testing.T) {
		note := model.NewNote("Test Title", "Test Content")

		// Create the note
		err := storage.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Try to create another note with the same ID
		duplicateNote := &model.Note{
			ID:        note.ID, // Same ID
			Title:     "Duplicate Title",
			Content:   "Duplicate Content",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = storage.Create(ctx, duplicateNote)
		if err == nil {
			t.Error("Expected error when creating note with duplicate ID, got nil")
		}
	})

	// Test error cases
	t.Run("ErrorCases", func(t *testing.T) {
		// Create a context with a shorter timeout for error cases
		errorCtx, errorCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer errorCancel()

		// Test Create error
		t.Run("CreateError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			invalidURI := "mongodb://invalid:27017"
			badClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(invalidURI))
			badStorage := &MongoDBStorage{
				client:     badClient,
				collection: badClient.Database("test_db").Collection("test_collection"),
			}
			CleanupCloseWithContext(t, errorCtx, badStorage)

			note := model.NewNote("Error Note", "This should fail to create")
			err := badStorage.Create(errorCtx, note)
			if err == nil {
				t.Error("Expected error when creating note with bad storage, got nil")
			}
		})

		// Test Get error (other than not found)
		t.Run("GetError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			invalidURI := "mongodb://invalid:27017"
			badClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(invalidURI))
			badStorage := &MongoDBStorage{
				client:     badClient,
				collection: badClient.Database("test_db").Collection("test_collection"),
			}
			CleanupCloseWithContext(t, errorCtx, badStorage)

			_, err := badStorage.Get(errorCtx, "some-id")
			if err == nil {
				t.Error("Expected error when getting note with bad storage, got nil")
			}
		})

		// Test GetAll error
		t.Run("GetAllError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			invalidURI := "mongodb://invalid:27017"
			badClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(invalidURI))
			badStorage := &MongoDBStorage{
				client:     badClient,
				collection: badClient.Database("test_db").Collection("test_collection"),
			}
			CleanupCloseWithContext(t, errorCtx, badStorage)

			_, err := badStorage.GetAll(errorCtx)
			if err == nil {
				t.Error("Expected error when getting all notes with bad storage, got nil")
			}
		})

		// Test Update error (other than not found)
		t.Run("UpdateError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			invalidURI := "mongodb://invalid:27017"
			badClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(invalidURI))
			badStorage := &MongoDBStorage{
				client:     badClient,
				collection: badClient.Database("test_db").Collection("test_collection"),
			}
			CleanupCloseWithContext(t, errorCtx, badStorage)

			note := model.NewNote("Error Note", "This should fail to update")
			err := badStorage.Update(errorCtx, note)
			if err == nil {
				t.Error("Expected error when updating note with bad storage, got nil")
			}
		})

		// Test Delete error (other than not found)
		t.Run("DeleteError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			invalidURI := "mongodb://invalid:27017"
			badClient, _ := mongo.Connect(ctx, options.Client().ApplyURI(invalidURI))
			badStorage := &MongoDBStorage{
				client:     badClient,
				collection: badClient.Database("test_db").Collection("test_collection"),
			}
			CleanupCloseWithContext(t, errorCtx, badStorage)

			err := badStorage.Delete(errorCtx, "some-id")
			if err == nil {
				t.Error("Expected error when deleting note with bad storage, got nil")
			}
		})
	})

	// Clean up after the test
	err = client.Database(dbName).Collection(collectionName).Drop(ctx)
	if err != nil {
		t.Logf("Warning: Failed to drop test collection: %v", err)
	}
}
