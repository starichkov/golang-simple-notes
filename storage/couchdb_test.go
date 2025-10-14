package storage

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-kivik/kivik/v4"

	"golang-simple-notes/model"
)

// TestCouchDBStorage tests the CouchDB storage implementation
// This test uses the shared CouchDB container from TestMain
func TestCouchDBStorage(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping CouchDB integration test in short mode")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the shared CouchDB container
	url := getSharedCouchURL()
	if url == "" {
		t.Skip("Shared CouchDB container not available")
	}
	dbName := "test_notes"

	// Connect to the CouchDB container
	client, err := kivik.New("couch", url)
	if err != nil {
		t.Fatalf("Failed to connect to CouchDB container: %v", err)
	}
	defer client.Close()

	// Check if the server is available
	_, err = client.AllDBs(ctx)
	if err != nil {
		t.Fatalf("Failed to list databases in CouchDB container: %v", err)
	}

	// Clean up any existing test database
	if exists, _ := client.DBExists(ctx, dbName); exists {
		if err := client.DestroyDB(ctx, dbName); err != nil {
			t.Logf("Warning: Failed to destroy test database: %v", err)
		}
	}

	// Create the database
	if err := client.CreateDB(ctx, dbName); err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create a new CouchDB storage
	storage, err := NewCouchDBStorage(url, dbName)
	if err != nil {
		t.Fatalf("Failed to create CouchDB storage: %v", err)
	}

	// Run the fixed storage tests
	testNoteStorage(t, storage, ctx)

	// Clean up after the test
	if err := client.DestroyDB(ctx, dbName); err != nil {
		t.Logf("Warning: Failed to destroy test database: %v", err)
	}
}

// TestCouchDBStorageUnit tests the CouchDB storage implementation with unit tests
func TestCouchDBStorageUnit(t *testing.T) {
	// Create a mock implementation of NoteStorage that behaves like CouchDB
	storage := NewMockCouchDBStorage()

	// Run the fixed storage tests
	testNoteStorage(t, storage, context.Background())
}

// MockCouchDBStorage is a mock implementation of NoteStorage that behaves like CouchDB
type MockCouchDBStorage struct {
	notes map[string]*model.Note
}

// NewMockCouchDBStorage creates a new instance of MockCouchDBStorage
func NewMockCouchDBStorage() *MockCouchDBStorage {
	return &MockCouchDBStorage{
		notes: make(map[string]*model.Note),
	}
}

// Create adds a new note to the storage
func (s *MockCouchDBStorage) Create(_ context.Context, note *model.Note) error {
	// Check if note with the same ID already exists
	if _, exists := s.notes[note.ID]; exists {
		return fmt.Errorf("note with ID %s already exists", note.ID)
	}

	// Store a copy of the note
	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID
func (s *MockCouchDBStorage) Get(_ context.Context, id string) (*model.Note, error) {
	note, exists := s.notes[id]
	if !exists {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *MockCouchDBStorage) GetAll(_ context.Context) ([]*model.Note, error) {
	notes := make([]*model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		// Skip design documents (in CouchDB, these have IDs starting with "_design/")
		if strings.HasPrefix(note.ID, "_design/") || strings.HasPrefix(note.ID, "_") {
			continue
		}
		notes = append(notes, note)
	}
	return notes, nil
}

// Update updates an existing note
func (s *MockCouchDBStorage) Update(_ context.Context, note *model.Note) error {
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
func (s *MockCouchDBStorage) Delete(_ context.Context, id string) error {
	if _, exists := s.notes[id]; !exists {
		return ErrNoteNotFound
	}

	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage
func (s *MockCouchDBStorage) Close(_ context.Context) error {
	// Nothing to close for mock storage
	return nil
}

// Additional CouchDB-specific tests could be added here
func TestCouchDBSpecificFeatures(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping CouchDB integration test in short mode")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the shared CouchDB container
	url := getSharedCouchURL()
	if url == "" {
		t.Skip("Shared CouchDB container not available")
	}
	dbName := "test_notes_specific"

	// Connect to the CouchDB container
	client, err := kivik.New("couch", url)
	if err != nil {
		t.Fatalf("Failed to connect to CouchDB container: %v", err)
	}
	defer client.Close()

	// Check if the server is available
	_, err = client.AllDBs(ctx)
	if err != nil {
		t.Fatalf("Failed to list databases in CouchDB container: %v", err)
	}

	// Clean up any existing test database
	if exists, _ := client.DBExists(ctx, dbName); exists {
		if err := client.DestroyDB(ctx, dbName); err != nil {
			t.Logf("Warning: Failed to destroy test database: %v", err)
		}
	}

	// Create the database
	if err := client.CreateDB(ctx, dbName); err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create a new CouchDB storage
	storage, err := NewCouchDBStorage(url, dbName)
	if err != nil {
		t.Fatalf("Failed to create CouchDB storage: %v", err)
	}
	defer storage.Close(ctx)

	// Test document revision handling
	t.Run("DocumentRevision", func(t *testing.T) {
		note := model.NewNote("Test Title", "Test Content")

		// Create the note
		err := storage.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Update the note
		note.Title = "Updated Title"
		err = storage.Update(ctx, note)
		if err != nil {
			t.Fatalf("Failed to update note: %v", err)
		}

		// Get the note to verify the update
		retrieved, err := storage.Get(ctx, note.ID)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if retrieved.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", retrieved.Title)
		}
	})

	// Test handling of design documents in GetAll
	t.Run("SkipDesignDocuments", func(t *testing.T) {
		// Clean up any existing test database
		if exists, _ := client.DBExists(ctx, dbName); exists {
			if err := client.DestroyDB(ctx, dbName); err != nil {
				t.Logf("Warning: Failed to destroy test database: %v", err)
			}
		}

		// Create the database again
		if err := client.CreateDB(ctx, dbName); err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}

		// Create a new CouchDB storage
		storage, err := NewCouchDBStorage(url, dbName)
		if err != nil {
			t.Fatalf("Failed to create CouchDB storage: %v", err)
		}
		defer storage.Close(ctx)

		// Create a design document directly using the CouchDB client
		designDoc := map[string]interface{}{
			"_id": "_design/test",
			"views": map[string]interface{}{
				"test": map[string]interface{}{
					"map": "function(doc) { emit(doc._id, null); }",
				},
			},
		}

		db := client.DB(dbName)
		_, err = db.Put(ctx, "_design/test", designDoc)
		if err != nil {
			t.Fatalf("Failed to create design document: %v", err)
		}

		// Create a regular note
		note := model.NewNote("Regular Note", "This is a regular note")
		err = storage.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Print all documents in the database for debugging
		rows := db.AllDocs(ctx)
		defer rows.Close()
		t.Log("All documents in the database:")
		for rows.Next() {
			var id string
			if err := rows.ScanKey(&id); err != nil {
				t.Logf("Failed to scan document key: %v", err)
				continue
			}
			t.Logf("Document ID: %s", id)
		}

		// Get all notes
		notes, err := storage.GetAll(ctx)
		if err != nil {
			t.Fatalf("Failed to get all notes: %v", err)
		}

		// Print all notes for debugging
		t.Log("All notes returned by GetAll:")
		for i, n := range notes {
			t.Logf("Note %d: ID=%s, Title=%s", i, n.ID, n.Title)
		}

		// Verify that only the regular note is returned
		if len(notes) != 1 {
			t.Errorf("Expected 1 note, got %d", len(notes))
		}

		if len(notes) > 0 && notes[0].ID != note.ID {
			t.Errorf("Expected note ID %s, got %s", note.ID, notes[0].ID)
		}
	})

	// Test error cases
	t.Run("ErrorCases", func(t *testing.T) {
		// Test Create error
		t.Run("CreateError", func(t *testing.T) {
			// Create a storage with a closed client to simulate an error
			badStorage := &CouchDBStorage{
				client: client,
				db:     client.DB("non_existent_db"), // Use a non-existent database
			}
			defer badStorage.Close(ctx)

			note := model.NewNote("Error Note", "This should fail to create")
			err := badStorage.Create(ctx, note)
			if err == nil {
				t.Error("Expected error when creating note with bad storage, got nil")
			}
		})

		// Test Get error (other than not found)
		t.Run("GetError", func(t *testing.T) {
			// Create a storage with a closed context to simulate an error
			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel the context immediately
			defer cancel()

			badStorage := &CouchDBStorage{
				client: client,
				db:     client.DB(dbName),
			}
			defer badStorage.Close(ctx)

			_, err := badStorage.Get(canceledCtx, "some-id")
			if err == nil {
				t.Error("Expected error when getting note with canceled context, got nil")
			}
		})

		// Test GetAll error
		t.Run("GetAllError", func(t *testing.T) {
			// Create a storage with a closed context to simulate an error
			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel the context immediately
			defer cancel()

			badStorage := &CouchDBStorage{
				client: client,
				db:     client.DB(dbName),
			}
			defer badStorage.Close(ctx)

			_, err := badStorage.GetAll(canceledCtx)
			if err == nil {
				t.Error("Expected error when getting all notes with canceled context, got nil")
			}
		})

		// Test Update error (other than not found)
		t.Run("UpdateError", func(t *testing.T) {
			// Create a storage with a closed context to simulate an error
			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel the context immediately
			defer cancel()

			badStorage := &CouchDBStorage{
				client: client,
				db:     client.DB(dbName),
			}
			defer badStorage.Close(ctx)

			note := model.NewNote("Error Note", "This should fail to update")
			err := badStorage.Update(canceledCtx, note)
			if err == nil {
				t.Error("Expected error when updating note with canceled context, got nil")
			}
		})

		// Test Delete error (other than not found)
		t.Run("DeleteError", func(t *testing.T) {
			// Create a storage with a closed context to simulate an error
			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel the context immediately
			defer cancel()

			badStorage := &CouchDBStorage{
				client: client,
				db:     client.DB(dbName),
			}
			defer badStorage.Close(ctx)

			err := badStorage.Delete(canceledCtx, "some-id")
			if err == nil {
				t.Error("Expected error when deleting note with canceled context, got nil")
			}
		})
	})

	// Clean up after the test
	if err := client.DestroyDB(ctx, dbName); err != nil {
		t.Logf("Warning: Failed to destroy test database: %v", err)
	}
}
