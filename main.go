package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create a context that will be canceled on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize configuration
	config := NewConfig()

	// Create and initialize the application
	app := NewApp(config)
	if err := app.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application
	if err := app.Run(ctx); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
