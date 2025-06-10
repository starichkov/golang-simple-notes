// Package main is the entry point for the Notes API application.
// This application provides a simple microservice for notes management with both REST and gRPC APIs,
// supporting multiple storage backends (in-memory, CouchDB, and MongoDB).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// main is the entry point of the application.
// It initializes the application, sets up signal handling for graceful shutdown,
// and starts the servers.
func main() {
	// Create a context that will be canceled on interrupt signal (Ctrl+C)
	// This allows for graceful shutdown when the application is terminated
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // Ensure the signal handler is removed when the function exits

	// Initialize configuration from environment variables
	// See config.go for details on available configuration options
	config := NewConfig()

	// Create and initialize the application with the configuration
	app := NewApp(config)
	if err := app.Initialize(ctx); err != nil {
		// If initialization fails, log the error and exit
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application, which starts the REST and gRPC servers
	if err := app.Run(ctx); err != nil {
		// If running fails, log the error and exit
		log.Fatalf("Failed to run application: %v", err)
	}
}
