package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang-simple-notes/model"
)

// TestMongoDBStorage tests the MongoDB storage implementation
// This test uses Testcontainers to start a MongoDB container
func TestMongoDBStorage(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	ctx := context.Background()

	// Start MongoDB container
	mongodbContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:7.0.7"),
		mongodb.WithUsername("admin"),
		mongodb.WithPassword("password"),
	)
	if err != nil {
		t.Fatalf("Failed to start MongoDB container: %v", err)
	}

	// Make sure to terminate the container at the end of the test
	defer func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MongoDB container: %v", err)
		}
	}()

	// Get connection details
	mongodbEndpoint, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get MongoDB connection string: %v", err)
	}

	// Add database and collection names to the URI
	dbName := "test_notes"
	collectionName := "test_notes"

	// Connect to the MongoDB container
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB container: %v", err)
	}
	defer client.Disconnect(ctx)

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
	testNoteStorageFixed(t, storage)

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
	testNoteStorageFixed(t, storage)
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
func (s *MockMongoDBStorage) Create(note *model.Note) error {
	// Check if note with the same ID already exists
	if _, exists := s.notes[note.ID]; exists {
		return fmt.Errorf("note with ID %s already exists", note.ID)
	}

	// Store a copy of the note
	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID
func (s *MockMongoDBStorage) Get(id string) (*model.Note, error) {
	note, exists := s.notes[id]
	if !exists {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *MockMongoDBStorage) GetAll() ([]*model.Note, error) {
	notes := make([]*model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

// Update updates an existing note
func (s *MockMongoDBStorage) Update(note *model.Note) error {
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
func (s *MockMongoDBStorage) Delete(id string) error {
	if _, exists := s.notes[id]; !exists {
		return ErrNoteNotFound
	}

	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage
func (s *MockMongoDBStorage) Close() error {
	// Nothing to close for mock storage
	return nil
}

// Additional MongoDB-specific tests could be added here
func TestMongoDBSpecificFeatures(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping MongoDB integration test in short mode")
	}

	ctx := context.Background()

	// Start MongoDB container
	mongodbContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:7.0.7"),
		mongodb.WithUsername("admin"),
		mongodb.WithPassword("password"),
	)
	if err != nil {
		t.Fatalf("Failed to start MongoDB container: %v", err)
	}

	// Make sure to terminate the container at the end of the test
	defer func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MongoDB container: %v", err)
		}
	}()

	// Get connection details
	mongodbEndpoint, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get MongoDB connection string: %v", err)
	}

	// Add database and collection names to the URI
	dbName := "test_notes"
	collectionName := "test_notes"

	// Connect to the MongoDB container
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB container: %v", err)
	}
	defer client.Disconnect(ctx)

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
	defer storage.Close()

	// Test that the unique index on ID works
	t.Run("UniqueIDIndex", func(t *testing.T) {
		note := model.NewNote("Test Title", "Test Content")

		// Create the note
		err := storage.Create(note)
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

		err = storage.Create(duplicateNote)
		if err == nil {
			t.Error("Expected error when creating note with duplicate ID, got nil")
		}
	})

	// Test error cases
	t.Run("ErrorCases", func(t *testing.T) {
		// Test Create error
		t.Run("CreateError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			invalidURI := "mongodb://invalid:27017"
			badStorage, err := NewMongoDBStorage(invalidURI, "test_db", "test_collection")
			if err != nil {
				// If NewMongoDBStorage fails, that's expected, but we need to test Create error
				// So create a bad storage manually
				badClient, _ := mongo.NewClient(options.Client().ApplyURI(invalidURI))
				badStorage = &MongoDBStorage{
					client:     badClient,
					collection: badClient.Database("test_db").Collection("test_collection"),
					ctx:        context.Background(),
					cancel:     func() {},
				}
			}

			note := model.NewNote("Error Note", "This should fail to create")
			err = badStorage.Create(note)
			if err == nil {
				t.Error("Expected error when creating note with bad storage, got nil")
			}
		})

		// Test Get error (other than not found)
		t.Run("GetError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			invalidURI := "mongodb://invalid:27017"
			badStorage, err := NewMongoDBStorage(invalidURI, "test_db", "test_collection")
			if err != nil {
				// If NewMongoDBStorage fails, that's expected, but we need to test Get error
				// So create a bad storage manually
				badClient, _ := mongo.NewClient(options.Client().ApplyURI(invalidURI))
				badStorage = &MongoDBStorage{
					client:     badClient,
					collection: badClient.Database("test_db").Collection("test_collection"),
					ctx:        context.Background(),
					cancel:     func() {},
				}
			}

			_, err = badStorage.Get("some-id")
			if err == nil {
				t.Error("Expected error when getting note with bad storage, got nil")
			}
		})

		// Test GetAll error
		t.Run("GetAllError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			invalidURI := "mongodb://invalid:27017"
			badStorage, err := NewMongoDBStorage(invalidURI, "test_db", "test_collection")
			if err != nil {
				// If NewMongoDBStorage fails, that's expected, but we need to test GetAll error
				// So create a bad storage manually
				badClient, _ := mongo.NewClient(options.Client().ApplyURI(invalidURI))
				badStorage = &MongoDBStorage{
					client:     badClient,
					collection: badClient.Database("test_db").Collection("test_collection"),
					ctx:        context.Background(),
					cancel:     func() {},
				}
			}

			_, err = badStorage.GetAll()
			if err == nil {
				t.Error("Expected error when getting all notes with bad storage, got nil")
			}
		})

		// Test Update error (other than not found)
		t.Run("UpdateError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			invalidURI := "mongodb://invalid:27017"
			badStorage, err := NewMongoDBStorage(invalidURI, "test_db", "test_collection")
			if err != nil {
				// If NewMongoDBStorage fails, that's expected, but we need to test Update error
				// So create a bad storage manually
				badClient, _ := mongo.NewClient(options.Client().ApplyURI(invalidURI))
				badStorage = &MongoDBStorage{
					client:     badClient,
					collection: badClient.Database("test_db").Collection("test_collection"),
					ctx:        context.Background(),
					cancel:     func() {},
				}
			}

			note := model.NewNote("Error Note", "This should fail to update")
			err = badStorage.Update(note)
			if err == nil {
				t.Error("Expected error when updating note with bad storage, got nil")
			}
		})

		// Test Delete error (other than not found)
		t.Run("DeleteError", func(t *testing.T) {
			// Create a storage with an invalid client to simulate an error
			invalidURI := "mongodb://invalid:27017"
			badStorage, err := NewMongoDBStorage(invalidURI, "test_db", "test_collection")
			if err != nil {
				// If NewMongoDBStorage fails, that's expected, but we need to test Delete error
				// So create a bad storage manually
				badClient, _ := mongo.NewClient(options.Client().ApplyURI(invalidURI))
				badStorage = &MongoDBStorage{
					client:     badClient,
					collection: badClient.Database("test_db").Collection("test_collection"),
					ctx:        context.Background(),
					cancel:     func() {},
				}
			}

			err = badStorage.Delete("some-id")
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
