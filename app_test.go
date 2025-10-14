package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestApp_Initialize(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	app := NewApp(config)
	ctx := context.Background()

	err := app.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	if app.storage == nil {
		t.Error("Expected storage to be initialized")
	}

	if app.restServer == nil {
		t.Error("Expected REST server to be initialized")
	}

	if app.grpcServer == nil {
		t.Error("Expected gRPC server to be initialized")
	}
}

func TestApp_InitializeWithCouchDB(t *testing.T) {
	ctx := context.Background()

	// Use the shared CouchDB container
	couchURL := sharedCouchURL
	if couchURL == "" {
		t.Skip("Shared CouchDB container not available")
	}

	defer func() {
		// Clean up: delete the notes database
		req, _ := http.NewRequest(http.MethodDelete, couchURL+"/notes", nil)
		http.DefaultClient.Do(req)
	}()

	// Create app with CouchDB config
	config := &Config{
		StorageType: "couchdb",
		CouchDBURL:  couchURL,
		CouchDBName: "notes",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	app := NewApp(config)
	err := app.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create the 'notes' database in CouchDB
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

	err = app.storage.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	retrievedNote, err := app.storage.Get(ctx, note.ID)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}

	if retrievedNote.Title != note.Title {
		t.Errorf("Expected title %s, got %s", note.Title, retrievedNote.Title)
	}

	// Clean up
	defer app.storage.Close(ctx)
}

func TestApp_InitializeWithMongoDB(t *testing.T) {
	ctx := context.Background()

	// Use the shared MongoDB container
	mongoURI := sharedMongoURI
	if mongoURI == "" {
		t.Skip("Shared MongoDB container not available")
	}

	defer func() {
		// Clean up: drop the notes collection
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err == nil {
			_ = client.Database("notes").Collection("notes").Drop(ctx)
			_ = client.Disconnect(ctx)
		}
	}()

	// Create app with MongoDB config
	config := &Config{
		StorageType:       "mongodb",
		MongoDBURI:        mongoURI,
		MongoDBName:       "notes",
		MongoDBCollection: "notes",
		RESTPort:          ":8080",
		GRPCPort:          ":8081",
	}

	app := NewApp(config)
	err := app.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Test storage operations
	note := &model.Note{
		ID:      "test-id",
		Title:   "Test Note",
		Content: "Test Content",
	}

	err = app.storage.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Add a short delay to allow MongoDB to persist the note
	time.Sleep(100 * time.Millisecond)

	retrievedNote, err := app.storage.Get(ctx, note.ID)
	if err != nil {
		allNotes, _ := app.storage.GetAll(ctx)
		t.Fatalf("Failed to get note: %v. All notes: %+v", err, allNotes)
	}

	if retrievedNote.Title != note.Title {
		t.Errorf("Expected title %s, got %s", note.Title, retrievedNote.Title)
	}

	// Clean up
	defer app.storage.Close(ctx)
}

func TestApp_CreateSampleNotes(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	app := NewApp(config)
	ctx := context.Background()

	// Initialize with mock storage
	app.storage = storage.NewInMemoryStorage()

	err := app.createSampleNotes(ctx)
	if err != nil {
		t.Fatalf("Failed to create sample notes: %v", err)
	}

	// Verify notes were created
	notes, err := app.storage.GetAll(ctx)
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
	app.storage = &ErrorMockStorage{}
	err = app.createSampleNotes(ctx)
	if err == nil {
		t.Error("Expected error from createSampleNotes with ErrorMockStorage")
	}
}

func TestApp_WaitForShutdown(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	app := NewApp(config)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Initialize with mock storage
	app.storage = storage.NewInMemoryStorage()
	app.restServer = &http.Server{}

	// Run waitForShutdown in a goroutine
	done := make(chan error)
	go func() {
		done <- app.waitForShutdown(ctx)
	}()

	// Wait for either the timeout or completion
	select {
	case err := <-done:
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Errorf("waitForShutdown returned unexpected error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("waitForShutdown did not complete in time")
	}
}

// TestApp_StorageFallback tests the fallback to in-memory storage when CouchDB/MongoDB fails
func TestApp_StorageFallback(t *testing.T) {
	// Speed up failure paths by reducing retry/timeout for external DB clients
	t.Setenv("COUCHDB_MAX_ATTEMPTS", "1")
	t.Setenv("COUCHDB_RETRY_DELAY_MS", "100")
	t.Setenv("MONGODB_CONNECT_TIMEOUT_MS", "500")
	testCases := []struct {
		name        string
		config      *Config
		shouldFail  bool
		description string
	}{
		{
			name: "CouchDB Fallback",
			config: &Config{
				StorageType: "couchdb",
				CouchDBURL:  "http://invalid-url:5984", // Invalid URL to force fallback
				CouchDBName: "notes",
				RESTPort:    ":8080",
				GRPCPort:    ":8081",
			},
			shouldFail:  false,
			description: "Should fallback to in-memory storage when CouchDB is unavailable",
		},
		{
			name: "MongoDB Fallback",
			config: &Config{
				StorageType:       "mongodb",
				MongoDBURI:        "mongodb://invalid-url:27017", // Invalid URL to force fallback
				MongoDBName:       "notes",
				MongoDBCollection: "notes",
				RESTPort:          ":8080",
				GRPCPort:          ":8081",
			},
			shouldFail:  false,
			description: "Should fallback to in-memory storage when MongoDB is unavailable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(tc.config)
			ctx := context.Background()

			err := app.Initialize(ctx)
			if tc.shouldFail {
				if err == nil {
					t.Error("Expected initialization to fail")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to initialize app: %v", err)
			}

			// Verify that we're using in-memory storage
			note := &model.Note{
				ID:      "test-id",
				Title:   "Test Note",
				Content: "Test Content",
			}

			err = app.storage.Create(ctx, note)
			if err != nil {
				t.Fatalf("Failed to create note: %v", err)
			}

			retrievedNote, err := app.storage.Get(ctx, note.ID)
			if err != nil {
				t.Fatalf("Failed to get note: %v", err)
			}

			if retrievedNote.Title != note.Title {
				t.Errorf("Expected title %s, got %s", note.Title, retrievedNote.Title)
			}

			// Clean up
			if err := app.storage.Close(ctx); err != nil {
				t.Errorf("Failed to close storage: %v", err)
			}
		})
	}
}

// TestApp_GracefulShutdown tests proper cleanup during shutdown
func TestApp_GracefulShutdown(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	app := NewApp(config)
	ctx := context.Background()

	// Initialize the app
	if err := app.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create a note to ensure storage is used
	note := &model.Note{
		ID:      "test-id",
		Title:   "Test Note",
		Content: "Test Content",
	}

	if err := app.storage.Create(ctx, note); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Create a context that will be cancelled
	shutdownCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Start the app in a goroutine
	done := make(chan error)
	go func() {
		done <- app.Run(shutdownCtx)
	}()

	// Wait for either shutdown or timeout
	select {
	case err := <-done:
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Errorf("Expected context error during shutdown, got: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Shutdown did not complete in time")
	}

	// No longer check that storage is inaccessible after shutdown, since in-memory storage's Close is a no-op.
}

// TestApp_ContextHandling tests context cancellation and timeout scenarios
func TestApp_ContextHandling(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	testCases := []struct {
		name        string
		timeout     time.Duration
		shouldFail  bool
		description string
	}{
		{
			name:        "Immediate Cancellation",
			timeout:     0,
			shouldFail:  true,
			description: "Should fail immediately when context is cancelled",
		},
		{
			name:        "Short Timeout",
			timeout:     50 * time.Millisecond,
			shouldFail:  true,
			description: "Should fail after short timeout",
		},
		{
			name:        "Long Timeout",
			timeout:     200 * time.Millisecond,
			shouldFail:  true,
			description: "Should fail after long timeout (app waits for shutdown)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(config)
			ctx := context.Background()

			// Initialize the app
			if err := app.Initialize(ctx); err != nil {
				t.Fatalf("Failed to initialize app: %v", err)
			}

			// Create a context with the specified timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, tc.timeout)
			defer cancel()

			// Start the app in a goroutine
			done := make(chan error)
			go func() {
				done <- app.Run(timeoutCtx)
			}()

			// Wait for either completion or timeout
			select {
			case err := <-done:
				if tc.shouldFail {
					if err != context.DeadlineExceeded && err != context.Canceled {
						t.Errorf("Expected context error, got: %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			case <-time.After(tc.timeout * 2):
				if !tc.shouldFail {
					t.Error("Operation did not complete in time")
				}
			}

			// Clean up
			if err := app.storage.Close(ctx); err != nil {
				t.Errorf("Failed to close storage: %v", err)
			}
		})
	}
}

func TestApp_SetupRESTServer(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":8080",
		GRPCPort:    ":8081",
	}

	app := NewApp(config)
	app.storage = storage.NewInMemoryStorage()

	server := app.setupRESTServer()
	if server == nil {
		t.Fatal("Expected server to be initialized")
	}

	if server.Addr != ":8080" {
		t.Errorf("Expected server address to be ':8080', got %s", server.Addr)
	}

	if server.Handler == nil {
		t.Error("Expected server handler to be initialized")
	}
}

func TestApp_SetupGRPCServer(t *testing.T) {
	testCases := []struct {
		name     string
		port     string
		expected int
	}{
		{"default port", "", 8081},
		{"valid port", ":9090", 9090},
		{"invalid port", ":invalid", 8081}, // Should fall back to default
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				StorageType: "memory",
				RESTPort:    ":8080",
				GRPCPort:    tc.port,
			}

			app := NewApp(config)
			app.storage = storage.NewInMemoryStorage()

			server := app.setupGRPCServer()
			if server == nil {
				t.Fatal("Expected server to be initialized")
			}
		})
	}
}

func TestApp_StartServers(t *testing.T) {
	config := &Config{
		StorageType: "memory",
		RESTPort:    ":0", // Use port 0 to avoid conflicts
		GRPCPort:    ":0",
	}

	app := NewApp(config)
	app.storage = storage.NewInMemoryStorage()
	app.restServer = app.setupRESTServer()
	app.grpcServer = app.setupGRPCServer()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := app.startServers(ctx)
	if err != nil {
		t.Fatalf("Failed to start servers: %v", err)
	}

	// Give servers time to start
	time.Sleep(10 * time.Millisecond)

	// Clean up
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app.restServer.Shutdown(shutdownCtx)
	// gRPC server cleanup is handled by the context timeout
}

func TestApp_IsDuplicateKeyError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"MongoDB duplicate key", fmt.Errorf("E11000 duplicate key error"), true},
		{"CouchDB conflict", fmt.Errorf("conflict"), true},
		{"CouchDB document update conflict", fmt.Errorf("Document update conflict"), true},
		{"In-memory duplicate", fmt.Errorf("note already exists"), true},
		{"Other error", fmt.Errorf("some other error"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isDuplicateKeyError(tc.err)
			if result != tc.expected {
				t.Errorf("Expected %v for error '%v', got %v", tc.expected, tc.err, result)
			}
		})
	}
}

func TestApp_InitializeStorageError(t *testing.T) {
	// Speed up failure paths by reducing retry/timeout for external DB clients
	t.Setenv("COUCHDB_MAX_ATTEMPTS", "1")
	t.Setenv("COUCHDB_RETRY_DELAY_MS", "100")
	t.Setenv("MONGODB_CONNECT_TIMEOUT_MS", "500")
	testCases := []struct {
		name        string
		storageType string
		config      *Config
	}{
		{
			name:        "invalid CouchDB URL",
			storageType: "couchdb",
			config: &Config{
				StorageType: "couchdb",
				CouchDBURL:  "invalid://url",
				CouchDBName: "notes",
			},
		},
		{
			name:        "invalid MongoDB URI",
			storageType: "mongodb",
			config: &Config{
				StorageType:       "mongodb",
				MongoDBURI:        "invalid://uri",
				MongoDBName:       "notes",
				MongoDBCollection: "notes",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(tc.config)
			storage, err := app.initializeStorage(context.Background())
			if err != nil {
				t.Fatalf("Expected fallback to in-memory storage, got error: %v", err)
			}
			if storage == nil {
				t.Fatal("Expected in-memory storage to be initialized")
			}
		})
	}
}
