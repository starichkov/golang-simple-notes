package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-simple-notes/grpc"
	"golang-simple-notes/model"
	"golang-simple-notes/rest"
	"golang-simple-notes/storage"
)

func main() {
	// Create a context that will be canceled on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize the application
	noteStorage, err := initializeStorage(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Start the servers
	if err := runServers(ctx, noteStorage); err != nil {
		log.Fatalf("Failed to run servers: %v", err)
	}
}

// initializeStorage initializes the storage based on environment variables
func initializeStorage(ctx context.Context) (storage.NoteStorage, error) {
	// Get storage configuration from environment variables
	storageType := getEnv("STORAGE_TYPE", "memory")

	// Get CouchDB configuration from environment variables
	couchDBURL := getEnv("COUCHDB_URL", "http://localhost:5984")
	couchDBName := getEnv("COUCHDB_DB", "notes")

	// Get MongoDB configuration from environment variables
	mongoDBURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	mongoDBName := getEnv("MONGODB_DB", "notes")
	mongoDBCollection := getEnv("MONGODB_COLLECTION", "notes")

	// Create a new storage instance
	var noteStorage storage.NoteStorage
	var err error

	// Initialize the appropriate storage based on configuration
	switch storageType {
	case "couchdb":
		log.Printf("Connecting to CouchDB at %s, database: %s", couchDBURL, couchDBName)
		noteStorage, err = storage.NewCouchDBStorage(couchDBURL, couchDBName)
		if err != nil {
			log.Printf("Failed to connect to CouchDB: %v, falling back to in-memory storage", err)
			noteStorage = storage.NewInMemoryStorage()
		} else {
			log.Println("Successfully connected to CouchDB")
		}
	case "mongodb":
		log.Printf("Connecting to MongoDB at %s, database: %s, collection: %s", mongoDBURI, mongoDBName, mongoDBCollection)
		noteStorage, err = storage.NewMongoDBStorage(mongoDBURI, mongoDBName, mongoDBCollection)
		if err != nil {
			log.Printf("Failed to connect to MongoDB: %v, falling back to in-memory storage", err)
			noteStorage = storage.NewInMemoryStorage()
		} else {
			log.Println("Successfully connected to MongoDB")
		}
	default:
		log.Println("Using in-memory storage")
		noteStorage = storage.NewInMemoryStorage()
	}

	return noteStorage, nil
}

// runServers starts the REST and gRPC servers
func runServers(ctx context.Context, noteStorage storage.NoteStorage) error {
	// Setup servers
	restServer := setupRESTServer(noteStorage)
	grpcServer := setupGRPCServer(noteStorage)

	// Start servers
	if err := startServers(ctx, restServer, grpcServer); err != nil {
		return fmt.Errorf("failed to start servers: %w", err)
	}

	// Create some sample notes
	if err := createSampleNotes(ctx, noteStorage); err != nil {
		return fmt.Errorf("failed to create sample notes: %w", err)
	}

	// Wait for shutdown signal
	return waitForShutdown(ctx, restServer, noteStorage)
}

// setupRESTServer creates and configures the REST server
func setupRESTServer(noteStorage storage.NoteStorage) *http.Server {
	restHandler := rest.NewHandler(noteStorage)
	mux := http.NewServeMux()
	restHandler.RegisterRoutes(mux)

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}

// setupGRPCServer creates and configures the gRPC server
func setupGRPCServer(noteStorage storage.NoteStorage) *grpc.Server {
	return grpc.NewServer(noteStorage, 8081)
}

// startServers starts the REST and gRPC servers in separate goroutines
func startServers(ctx context.Context, restServer *http.Server, grpcServer *grpc.Server) error {
	// Start REST server
	go func() {
		log.Println("Starting REST server on :8080")
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("REST server failed: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		log.Println("Starting gRPC server on :8081")
		if err := grpcServer.Start(); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}
	}()

	return nil
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the servers
func waitForShutdown(ctx context.Context, restServer *http.Server, noteStorage storage.NoteStorage) error {
	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Shutting down servers...")

	// Create a new context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the REST server
	if err := restServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("REST server shutdown failed: %v", err)
	}

	// Close the storage
	if err := noteStorage.Close(shutdownCtx); err != nil {
		log.Printf("Storage shutdown failed: %v", err)
	}

	log.Println("Servers stopped")
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// createSampleNotes creates some sample notes in the storage
func createSampleNotes(ctx context.Context, storage storage.NoteStorage) error {
	notes := []struct {
		title   string
		content string
	}{
		{
			title:   "Welcome to Notes API",
			content: "This is a simple notes management API with REST and gRPC interfaces.",
		},
		{
			title:   "How to use the API",
			content: "You can create, read, update, and delete notes using the REST API or gRPC.",
		},
		{
			title:   "REST API Endpoints",
			content: "GET /api/notes - List all notes\nGET /api/notes/{id} - Get a note by ID\nPOST /api/notes - Create a new note\nPUT /api/notes/{id} - Update a note\nDELETE /api/notes/{id} - Delete a note",
		},
	}

	for _, note := range notes {
		n := model.NewNote(note.title, note.content)
		if err := storage.Create(ctx, n); err != nil {
			return fmt.Errorf("failed to create sample note: %w", err)
		}
	}

	return nil
}
