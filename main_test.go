package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
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

func (s *MockStorage) Create(ctx context.Context, note *model.Note) error {
	s.notes = append(s.notes, note)
	return nil
}

func (s *MockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	for _, note := range s.notes {
		if note.ID == id {
			return note, nil
		}
	}
	return nil, storage.ErrNoteNotFound
}

func (s *MockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	return s.notes, nil
}

func (s *MockStorage) Update(ctx context.Context, note *model.Note) error {
	for i, n := range s.notes {
		if n.ID == note.ID {
			s.notes[i] = note
			return nil
		}
	}
	return storage.ErrNoteNotFound
}

func (s *MockStorage) Delete(ctx context.Context, id string) error {
	for i, note := range s.notes {
		if note.ID == id {
			s.notes = append(s.notes[:i], s.notes[i+1:]...)
			return nil
		}
	}
	return storage.ErrNoteNotFound
}

func (s *MockStorage) Close(ctx context.Context) error {
	return nil
}

// ErrorMockStorage is a mock storage that returns an error on Create
type ErrorMockStorage struct{}

func (s *ErrorMockStorage) Create(ctx context.Context, note *model.Note) error {
	return fmt.Errorf("mock error")
}

func (s *ErrorMockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	return nil, storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	return nil, nil
}

func (s *ErrorMockStorage) Update(ctx context.Context, note *model.Note) error {
	return storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) Delete(ctx context.Context, id string) error {
	return storage.ErrNoteNotFound
}

func (s *ErrorMockStorage) Close(ctx context.Context) error {
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
	ctx := context.Background()
	// Test with default (memory) storage
	os.Unsetenv("STORAGE_TYPE")
	storage, err := initializeStorage(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Expected storage to be non-nil")
	}

	// Clean up
	defer storage.Close(ctx)
}

// startCouchDBContainer starts a CouchDB container and returns the container and connection URL
func startCouchDBContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "couchdb:3.3.2",
		ExposedPorts: []string{"5984/tcp"},
		WaitingFor:   wait.ForListeningPort("5984/tcp"),
		Env: map[string]string{
			"COUCHDB_USER":     "admin",
			"COUCHDB_PASSWORD": "password",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}
	port, err := container.MappedPort(ctx, "5984")
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}
	url := fmt.Sprintf("http://admin:password@%s:%s", host, port.Port())
	return container, url, nil
}

// TestInitializeStorageWithCouchDB tests the initializeStorage function with CouchDB
func TestInitializeStorageWithCouchDB(t *testing.T) {
	ctx := context.Background()

	// Start CouchDB container using the generic API
	couchContainer, couchURL, err := startCouchDBContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start CouchDB container: %v", err)
	}
	defer func() {
		if err := couchContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate CouchDB container: %v", err)
		}
	}()

	// Set environment variables for CouchDB
	os.Setenv("STORAGE_TYPE", "couchdb")
	os.Setenv("COUCHDB_URL", couchURL)
	defer func() {
		os.Unsetenv("STORAGE_TYPE")
		os.Unsetenv("COUCHDB_URL")
	}()

	// Initialize storage with CouchDB
	storage, err := initializeStorage(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Expected storage to be non-nil")
	}

	// Create the 'notes' database in CouchDB using http.NewRequest
	req, err := http.NewRequest(http.MethodPut, couchURL+"/notes", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create CouchDB database: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusPreconditionFailed {
		t.Fatalf("Failed to create CouchDB database: %s", resp.Status)
	}

	// Test storage operations
	note := &model.Note{
		ID:      "test-id",
		Title:   "Test Note",
		Content: "Test Content",
	}

	err = storage.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	retrievedNote, err := storage.Get(ctx, note.ID)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}

	if retrievedNote.Title != note.Title {
		t.Errorf("Expected title %s, got %s", note.Title, retrievedNote.Title)
	}

	// Clean up
	defer storage.Close(ctx)
}

// TestInitializeStorageWithMongoDB tests the initializeStorage function with MongoDB
func TestInitializeStorageWithMongoDB(t *testing.T) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6.0"))
	if err != nil {
		t.Fatalf("Failed to start MongoDB container: %v", err)
	}
	defer func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate MongoDB container: %v", err)
		}
	}()

	// Get MongoDB connection details
	mongoURI, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get MongoDB connection string: %v", err)
	}

	// Set environment variables for MongoDB
	os.Setenv("STORAGE_TYPE", "mongodb")
	os.Setenv("MONGODB_URI", mongoURI)
	defer func() {
		os.Unsetenv("STORAGE_TYPE")
		os.Unsetenv("MONGODB_URI")
	}()

	// Initialize storage with MongoDB
	storage, err := initializeStorage(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Expected storage to be non-nil")
	}

	// Test storage operations
	note := &model.Note{
		ID:      "test-id",
		Title:   "Test Note",
		Content: "Test Content",
	}

	// Create the note first
	err = storage.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Add a short delay to allow MongoDB to persist the note
	time.Sleep(100 * time.Millisecond)

	// Then get the note
	retrievedNote, err := storage.Get(ctx, note.ID)
	if err != nil {
		allNotes, _ := storage.GetAll(ctx)
		t.Fatalf("Failed to get note: %v. All notes: %+v", err, allNotes)
	}

	if retrievedNote.Title != note.Title {
		t.Errorf("Expected title %s, got %s", note.Title, retrievedNote.Title)
	}

	// Clean up
	defer storage.Close(ctx)
}

// TestRunServersSimple tests a simplified version of runServers
func TestRunServersSimple(t *testing.T) {
	ctx := context.Background()
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
	err := createSampleNotes(ctx, mockStorage)
	if err != nil {
		t.Fatalf("Failed to create sample notes: %v", err)
	}
	notes, err := mockStorage.GetAll(ctx)
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
	ctx := context.Background()

	// Call the function to create sample notes
	err := createSampleNotes(ctx, mockStorage)
	if err != nil {
		t.Fatalf("Failed to create sample notes: %v", err)
	}

	// Verify that notes were created
	notes, err := mockStorage.GetAll(ctx)
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

	// This should return an error
	err = createSampleNotes(ctx, errorMockStorage)
	if err == nil {
		t.Error("Expected error from createSampleNotes with ErrorMockStorage")
	}
}
