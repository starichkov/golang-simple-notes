package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang-simple-notes/grpc"
	"golang-simple-notes/model"
	"golang-simple-notes/rest"
	"golang-simple-notes/storage"
)

// App represents the main application
type App struct {
	storage    storage.NoteStorage
	restServer *http.Server
	grpcServer *grpc.Server
	config     *Config
}

// NewApp creates a new App instance
func NewApp(config *Config) *App {
	return &App{
		config: config,
	}
}

// Initialize sets up the application components
func (a *App) Initialize(ctx context.Context) error {
	// Initialize storage
	storage, err := a.initializeStorage(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	a.storage = storage

	// Setup servers
	a.restServer = a.setupRESTServer()
	a.grpcServer = a.setupGRPCServer()

	return nil
}

// Run starts the application servers
func (a *App) Run(ctx context.Context) error {
	// Start servers
	if err := a.startServers(ctx); err != nil {
		return fmt.Errorf("failed to start servers: %w", err)
	}

	// Create sample notes
	if err := a.createSampleNotes(ctx); err != nil {
		return fmt.Errorf("failed to create sample notes: %w", err)
	}

	// Wait for shutdown signal
	return a.waitForShutdown(ctx)
}

// initializeStorage initializes the storage based on configuration
func (a *App) initializeStorage(ctx context.Context) (storage.NoteStorage, error) {
	var noteStorage storage.NoteStorage
	var err error

	switch a.config.StorageType {
	case "couchdb":
		log.Printf("Connecting to CouchDB at %s, database: %s", a.config.CouchDBURL, a.config.CouchDBName)
		noteStorage, err = storage.NewCouchDBStorage(a.config.CouchDBURL, a.config.CouchDBName)
		if err != nil {
			log.Printf("Failed to connect to CouchDB: %v, falling back to in-memory storage", err)
			noteStorage = storage.NewInMemoryStorage()
		} else {
			log.Println("Successfully connected to CouchDB")
		}
	case "mongodb":
		log.Printf("Connecting to MongoDB at %s, database: %s, collection: %s",
			a.config.MongoDBURI, a.config.MongoDBName, a.config.MongoDBCollection)
		noteStorage, err = storage.NewMongoDBStorage(a.config.MongoDBURI, a.config.MongoDBName, a.config.MongoDBCollection)
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

// setupRESTServer creates and configures the REST server
func (a *App) setupRESTServer() *http.Server {
	restHandler := rest.NewHandler(a.storage)
	mux := http.NewServeMux()
	restHandler.RegisterRoutes(mux)

	return &http.Server{
		Addr:    a.config.RESTPort,
		Handler: mux,
	}
}

// setupGRPCServer creates and configures the gRPC server
func (a *App) setupGRPCServer() *grpc.Server {
	// Remove the colon prefix and convert to int
	port := 8081 // Default port
	if a.config.GRPCPort != "" {
		portStr := strings.TrimPrefix(a.config.GRPCPort, ":")
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	return grpc.NewServer(a.storage, port)
}

// startServers starts the REST and gRPC servers in separate goroutines
func (a *App) startServers(ctx context.Context) error {
	// Start REST server
	go func() {
		log.Printf("Starting REST server on %s", a.config.RESTPort)
		if err := a.restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("REST server failed: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		log.Printf("Starting gRPC server on %s", a.config.GRPCPort)
		if err := a.grpcServer.Start(); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}
	}()

	return nil
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the servers
func (a *App) waitForShutdown(ctx context.Context) error {
	<-ctx.Done()
	log.Println("Shutting down servers...")

	// Create a new context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the REST server
	if err := a.restServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("REST server shutdown failed: %v", err)
	}

	// Close the storage
	if err := a.storage.Close(shutdownCtx); err != nil {
		log.Printf("Storage shutdown failed: %v", err)
	}

	log.Println("Servers stopped")
	return ctx.Err()
}

// createSampleNotes creates some sample notes in the storage
func (a *App) createSampleNotes(ctx context.Context) error {
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
		err := a.storage.Create(ctx, n)
		if err != nil {
			// Ignore duplicate key errors (MongoDB error code 11000)
			if isDuplicateKeyError(err) {
				continue
			}
			return fmt.Errorf("failed to create sample note: %w", err)
		}
	}

	return nil
}

// isDuplicateKeyError checks if the error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()

	// MongoDB duplicate key error
	if strings.Contains(errStr, "E11000 duplicate key error") {
		return true
	}
	// CouchDB duplicate key error
	if strings.Contains(errStr, "conflict") || strings.Contains(errStr, "Document update conflict") {
		return true
	}
	// In-memory storage duplicate key error
	if strings.Contains(errStr, "note already exists") {
		return true
	}
	return false
}
