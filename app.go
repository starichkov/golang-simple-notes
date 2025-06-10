// Package main contains the core application logic for the Notes API.
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// App represents the main application that coordinates all components:
// - Storage backend (in-memory, CouchDB, or MongoDB)
// - REST API server
// - gRPC API server
// It handles initialization, running, and graceful shutdown of these components.
type App struct {
	storage    storage.NoteStorage // Interface for storing and retrieving notes
	restServer *http.Server        // HTTP server for REST API
	grpcServer *grpc.Server        // gRPC server for gRPC API
	config     *Config             // Application configuration
}

// NewApp creates a new App instance with the provided configuration.
// It only initializes the App struct with the configuration; actual component
// initialization happens in the Initialize method.
func NewApp(config *Config) *App {
	return &App{
		config: config,
	}
}

// Initialize sets up the application components in the following order:
// 1. Initializes the appropriate storage backend based on configuration
// 2. Sets up the REST server with routes
// 3. Sets up the gRPC server
// This method must be called before Run.
func (a *App) Initialize(ctx context.Context) error {
	// Initialize storage backend (in-memory, CouchDB, or MongoDB)
	// based on the configuration
	storage, err := a.initializeStorage(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	a.storage = storage

	// Setup the REST and gRPC servers with the initialized storage
	a.restServer = a.setupRESTServer()
	a.grpcServer = a.setupGRPCServer()

	return nil
}

// Run starts the application servers and performs the following steps:
// 1. Starts the REST and gRPC servers in separate goroutines
// 2. Creates sample notes in the storage
// 3. Waits for a shutdown signal (e.g., Ctrl+C)
// This method blocks until the application is shut down.
func (a *App) Run(ctx context.Context) error {
	// Start the REST and gRPC servers in separate goroutines
	if err := a.startServers(ctx); err != nil {
		return fmt.Errorf("failed to start servers: %w", err)
	}

	// Create sample notes in the storage for demonstration purposes
	if err := a.createSampleNotes(ctx); err != nil {
		return fmt.Errorf("failed to create sample notes: %w", err)
	}

	// Wait for shutdown signal (context cancellation)
	// This blocks until the context is canceled (e.g., by Ctrl+C)
	return a.waitForShutdown(ctx)
}

// initializeStorage initializes the storage backend based on the configuration.
// It supports three types of storage:
// - "couchdb": Uses CouchDB as the storage backend
// - "mongodb": Uses MongoDB as the storage backend
// - Any other value (default): Uses in-memory storage
//
// If connecting to CouchDB or MongoDB fails, it falls back to in-memory storage
// to ensure the application can still run.
func (a *App) initializeStorage(ctx context.Context) (storage.NoteStorage, error) {
	var noteStorage storage.NoteStorage
	var err error

	// Choose the storage backend based on the configuration
	switch a.config.StorageType {
	case "couchdb":
		// Try to connect to CouchDB
		log.Printf("Connecting to CouchDB at %s, database: %s", a.config.CouchDBURL, a.config.CouchDBName)
		noteStorage, err = storage.NewCouchDBStorage(a.config.CouchDBURL, a.config.CouchDBName)
		if err != nil {
			// If connection fails, log the error and fall back to in-memory storage
			log.Printf("Failed to connect to CouchDB: %v, falling back to in-memory storage", err)
			noteStorage = storage.NewInMemoryStorage()
		} else {
			log.Println("Successfully connected to CouchDB")
		}
	case "mongodb":
		// Try to connect to MongoDB
		log.Printf("Connecting to MongoDB at %s, database: %s, collection: %s",
			a.config.MongoDBURI, a.config.MongoDBName, a.config.MongoDBCollection)
		noteStorage, err = storage.NewMongoDBStorage(a.config.MongoDBURI, a.config.MongoDBName, a.config.MongoDBCollection)
		if err != nil {
			// If connection fails, log the error and fall back to in-memory storage
			log.Printf("Failed to connect to MongoDB: %v, falling back to in-memory storage", err)
			noteStorage = storage.NewInMemoryStorage()
		} else {
			log.Println("Successfully connected to MongoDB")
		}
	default:
		// Use in-memory storage by default
		log.Println("Using in-memory storage")
		noteStorage = storage.NewInMemoryStorage()
	}

	return noteStorage, nil
}

// setupRESTServer creates and configures the REST API server.
// It sets up:
// 1. A new REST handler with the storage backend
// 2. A Chi router with middleware for logging and panic recovery
// 3. Routes for the REST API endpoints
// 4. An HTTP server with the configured port
func (a *App) setupRESTServer() *http.Server {
	// Create a new REST handler with the storage backend
	restHandler := rest.NewHandler(a.storage)

	// Create a new Chi router
	// Chi is a lightweight, idiomatic and composable router for Go HTTP services
	r := chi.NewRouter()

	// Add middleware to the router
	r.Use(middleware.Logger)    // Log all HTTP requests
	r.Use(middleware.Recoverer) // Recover from panics without crashing the server

	// Register the API routes with the router
	// This sets up endpoints like GET /api/notes, POST /api/notes, etc.
	restHandler.RegisterRoutes(r)

	// Create and return an HTTP server with the configured port and router
	return &http.Server{
		Addr:    a.config.RESTPort, // Port to listen on (e.g., ":8080")
		Handler: r,                 // The router that handles requests
	}
}

// setupGRPCServer creates and configures the gRPC server.
// It extracts the port number from the configuration and creates a new gRPC server
// with the storage backend and port.
func (a *App) setupGRPCServer() *grpc.Server {
	// Extract the port number from the configuration
	// The port might be in the format ":8081", so we need to remove the colon prefix
	port := 8081 // Default port if parsing fails
	if a.config.GRPCPort != "" {
		portStr := strings.TrimPrefix(a.config.GRPCPort, ":")
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Create and return a new gRPC server with the storage backend and port
	return grpc.NewServer(a.storage, port)
}

// startServers starts the REST and gRPC servers in separate goroutines.
// This method doesn't block; it returns immediately after starting the servers.
// Each server runs in its own goroutine (a lightweight thread) to allow them to run concurrently.
func (a *App) startServers(ctx context.Context) error {
	// Start REST server in a separate goroutine
	go func() {
		log.Printf("Starting REST server on %s", a.config.RESTPort)
		// ListenAndServe blocks until the server is stopped or encounters an error
		if err := a.restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log any error that isn't just the server being closed normally
			log.Printf("REST server failed: %v", err)
		}
	}()

	// Start gRPC server in a separate goroutine
	go func() {
		log.Printf("Starting gRPC server on %s", a.config.GRPCPort)
		// Start blocks until the server is stopped or encounters an error
		if err := a.grpcServer.Start(); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}
	}()

	return nil
}

// waitForShutdown waits for the context to be canceled (e.g., by an interrupt signal)
// and then gracefully shuts down the servers.
// This method blocks until the context is canceled and the servers are shut down.
func (a *App) waitForShutdown(ctx context.Context) error {
	// Block until the context is canceled (e.g., by Ctrl+C)
	<-ctx.Done()
	log.Println("Shutting down servers...")

	// Create a new context with a 5-second timeout for the shutdown process
	// This ensures that shutdown doesn't hang indefinitely
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the context is canceled when the function returns

	// Gracefully shut down the REST server
	// This allows in-flight requests to complete before shutting down
	if err := a.restServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("REST server shutdown failed: %v", err)
	}

	// Close the storage connection
	// This ensures any database connections are properly closed
	if err := a.storage.Close(shutdownCtx); err != nil {
		log.Printf("Storage shutdown failed: %v", err)
	}

	log.Println("Servers stopped")
	// Return the original context's error (typically context.Canceled)
	return ctx.Err()
}

// createSampleNotes creates some sample notes in the storage for demonstration purposes.
// This provides initial data for users to see when they first access the API.
func (a *App) createSampleNotes(ctx context.Context) error {
	// Define a list of sample notes to create
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

	// Create each sample note in the storage
	for _, note := range notes {
		// Create a new Note object with the title and content
		n := model.NewNote(note.title, note.content)

		// Try to save the note to the storage
		err := a.storage.Create(ctx, n)
		if err != nil {
			// If the note already exists (duplicate key error), skip it and continue
			if isDuplicateKeyError(err) {
				continue
			}
			// For any other error, return it
			return fmt.Errorf("failed to create sample note: %w", err)
		}

		// Add a small delay between creating notes
		// This ensures unique IDs when using timestamp-based ID generation
		// (since our ID generation uses the current timestamp)
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// isDuplicateKeyError checks if the error is a duplicate key error from any of the
// supported storage backends (MongoDB, CouchDB, or in-memory).
//
// Different databases return different error messages for duplicate key errors,
// so this function normalizes them to a single boolean result.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()

	// Check for MongoDB duplicate key error
	// MongoDB returns error code E11000 for duplicate key errors
	if strings.Contains(errStr, "E11000 duplicate key error") {
		return true
	}

	// Check for CouchDB duplicate key error
	// CouchDB returns "conflict" or "Document update conflict" for duplicate key errors
	if strings.Contains(errStr, "conflict") || strings.Contains(errStr, "Document update conflict") {
		return true
	}

	// Check for in-memory storage duplicate key error
	// Our in-memory implementation returns "note already exists" for duplicate key errors
	if strings.Contains(errStr, "note already exists") {
		return true
	}

	// If none of the above patterns match, it's not a duplicate key error
	return false
}
