package main

import (
	"context"
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
	// Initialize the application
	noteStorage := initializeStorage()

	// Start the servers
	runServers(noteStorage)
}

// initializeStorage initializes the storage based on environment variables
func initializeStorage() storage.NoteStorage {
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

	return noteStorage
}

// runServers starts the REST and gRPC servers
func runServers(noteStorage storage.NoteStorage) {
	// Setup servers
	restServer := setupRESTServer(noteStorage)
	grpcServer := setupGRPCServer(noteStorage)

	// Start servers
	startServers(restServer, grpcServer)

	// Create some sample notes
	createSampleNotes(noteStorage)

	// Wait for shutdown signal
	waitForShutdown(restServer, noteStorage)
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
func startServers(restServer *http.Server, grpcServer *grpc.Server) {
	go func() {
		log.Println("Starting REST server on :8080")
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST server failed: %v", err)
		}
	}()

	go func() {
		log.Println("Starting gRPC server on :8081")
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the servers
func waitForShutdown(restServer *http.Server, noteStorage storage.NoteStorage) {
	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	// Shutdown the REST server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := restServer.Shutdown(ctx); err != nil {
		log.Printf("REST server shutdown failed: %v", err)
	}

	// Close the storage
	if err := noteStorage.Close(); err != nil {
		log.Printf("Storage shutdown failed: %v", err)
	}

	log.Println("Servers stopped")
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
func createSampleNotes(storage storage.NoteStorage) {
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
		if err := storage.Create(n); err != nil {
			log.Printf("Failed to create sample note: %v", err)
		}
	}
}
