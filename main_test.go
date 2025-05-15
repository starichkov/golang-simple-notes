package main

import (
	"fmt"
	"os"
	"testing"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"
)

// TestGetEnv tests the getEnv function
func TestGetEnv(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := getEnv("TEST_VAR", "default_value")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test with environment variable not set
	value = getEnv("NON_EXISTENT_VAR", "default_value")
	if value != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", value)
	}
}

// MockStorage is a simple implementation of NoteStorage for testing
type MockStorage struct {
	notes []*model.Note
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		notes: make([]*model.Note, 0),
	}
}

func (s *MockStorage) Create(note *model.Note) error {
	s.notes = append(s.notes, note)
	return nil
}

func (s *MockStorage) Get(id string) (*model.Note, error) {
	for _, note := range s.notes {
		if note.ID == id {
			return note, nil
		}
	}
	return nil, storage.ErrNoteNotFound
}

func (s *MockStorage) GetAll() ([]*model.Note, error) {
	return s.notes, nil
}

func (s *MockStorage) Update(note *model.Note) error {
	for i, n := range s.notes {
		if n.ID == note.ID {
			s.notes[i] = note
			return nil
		}
	}
	return storage.ErrNoteNotFound
}

func (s *MockStorage) Delete(id string) error {
	for i, note := range s.notes {
		if note.ID == id {
			s.notes = append(s.notes[:i], s.notes[i+1:]...)
			return nil
		}
	}
	return storage.ErrNoteNotFound
}

func (s *MockStorage) Close() error {
	return nil
}

// ErrorMockStorage is a mock storage that returns an error on Create
type ErrorMockStorage struct{}

func (s *ErrorMockStorage) Create(note *model.Note) error {
	return fmt.Errorf("mock error")
}

func (s *ErrorMockStorage) Get(id string) (*model.Note, error) {
	return nil, storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) GetAll() ([]*model.Note, error) {
	return nil, nil
}

func (s *ErrorMockStorage) Update(note *model.Note) error {
	return storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) Delete(id string) error {
	return storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) Close() error {
	return nil
}

// TestSetupRESTServer tests the setupRESTServer function
func TestSetupRESTServer(t *testing.T) {
	mockStorage := NewMockStorage()
	server := setupRESTServer(mockStorage)

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	if server.Addr != ":8080" {
		t.Errorf("Expected server address to be :8080, got %s", server.Addr)
	}

	if server.Handler == nil {
		t.Error("Expected server handler to be non-nil")
	}
}

// TestSetupGRPCServer tests the setupGRPCServer function
func TestSetupGRPCServer(t *testing.T) {
	mockStorage := NewMockStorage()
	server := setupGRPCServer(mockStorage)

	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}
}

// TestInitializeStorage tests the initializeStorage function
func TestInitializeStorage(t *testing.T) {
	// Test with default (memory) storage
	os.Unsetenv("STORAGE_TYPE")
	storage := initializeStorage()

	if storage == nil {
		t.Fatal("Expected storage to be non-nil")
	}

	// Clean up
	defer storage.Close()
}

// TestInitializeStorageWithCouchDB tests the initializeStorage function with CouchDB
func TestInitializeStorageWithCouchDB(t *testing.T) {
	// Set environment variables for CouchDB
	os.Setenv("STORAGE_TYPE", "couchdb")
	os.Setenv("COUCHDB_URL", "http://invalid-url:5984") // Invalid URL to force fallback
	defer func() {
		os.Unsetenv("STORAGE_TYPE")
		os.Unsetenv("COUCHDB_URL")
	}()

	// This should fall back to in-memory storage
	storage := initializeStorage()

	if storage == nil {
		t.Fatal("Expected storage to be non-nil")
	}

	// Clean up
	defer storage.Close()
}

// TestInitializeStorageWithMongoDB tests the initializeStorage function with MongoDB
func TestInitializeStorageWithMongoDB(t *testing.T) {
	// Set environment variables for MongoDB
	os.Setenv("STORAGE_TYPE", "mongodb")
	os.Setenv("MONGODB_URI", "mongodb://invalid-url:27017") // Invalid URL to force fallback
	defer func() {
		os.Unsetenv("STORAGE_TYPE")
		os.Unsetenv("MONGODB_URI")
	}()

	// This should fall back to in-memory storage
	storage := initializeStorage()

	if storage == nil {
		t.Fatal("Expected storage to be non-nil")
	}

	// Clean up
	defer storage.Close()
}

// TestRunServersSimple tests a simplified version of runServers
func TestRunServersSimple(t *testing.T) {
	// This test doesn't actually call runServers() because it would start servers
	// and block waiting for shutdown. Instead, we test the individual components
	// that runServers() calls, which we've already tested separately.

	// Create a mock storage
	mockStorage := NewMockStorage()

	// Test that setupRESTServer works with our mock storage
	restServer := setupRESTServer(mockStorage)
	if restServer == nil {
		t.Fatal("setupRESTServer returned nil")
	}

	// Test that setupGRPCServer works with our mock storage
	grpcServer := setupGRPCServer(mockStorage)
	if grpcServer == nil {
		t.Fatal("setupGRPCServer returned nil")
	}

	// Test that createSampleNotes works with our mock storage
	createSampleNotes(mockStorage)
	notes, err := mockStorage.GetAll()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}
	if len(notes) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(notes))
	}
}

// TestCreateSampleNotes tests the createSampleNotes function
func TestCreateSampleNotes(t *testing.T) {
	mockStorage := NewMockStorage()

	// Call the function to create sample notes
	createSampleNotes(mockStorage)

	// Verify that notes were created
	notes, err := mockStorage.GetAll()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	// Check that we have the expected number of notes
	if len(notes) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(notes))
	}

	// Check the content of the notes
	foundWelcome := false
	foundHowTo := false
	foundEndpoints := false

	for _, note := range notes {
		switch note.Title {
		case "Welcome to Notes API":
			foundWelcome = true
			if note.Content != "This is a simple notes management API with REST and gRPC interfaces." {
				t.Errorf("Unexpected content for 'Welcome to Notes API'")
			}
		case "How to use the API":
			foundHowTo = true
			if note.Content != "You can create, read, update, and delete notes using the REST API or gRPC." {
				t.Errorf("Unexpected content for 'How to use the API'")
			}
		case "REST API Endpoints":
			foundEndpoints = true
			expectedContent := "GET /api/notes - List all notes\nGET /api/notes/{id} - Get a note by ID\nPOST /api/notes - Create a new note\nPUT /api/notes/{id} - Update a note\nDELETE /api/notes/{id} - Delete a note"
			if note.Content != expectedContent {
				t.Errorf("Unexpected content for 'REST API Endpoints'")
			}
		}
	}

	if !foundWelcome {
		t.Error("Missing 'Welcome to Notes API' note")
	}
	if !foundHowTo {
		t.Error("Missing 'How to use the API' note")
	}
	if !foundEndpoints {
		t.Error("Missing 'REST API Endpoints' note")
	}

	// Test error handling with a custom mock that always returns an error
	errorMockStorage := &ErrorMockStorage{}

	// This should not panic even though Create returns an error
	createSampleNotes(errorMockStorage)
}
