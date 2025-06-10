// Package grpc implements the gRPC server for the Notes API.
// It provides a gRPC interface for creating, reading, updating, and deleting notes.
// This is a simplified implementation for demonstration purposes.
package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"
)

// Server implements the Notes gRPC service.
// It uses a storage implementation to persist and retrieve notes.
// This follows the dependency injection pattern, allowing the server
// to work with any storage implementation that satisfies the NoteStorage interface.
type Server struct {
	storage storage.NoteStorage // Storage backend for notes
	port    int                 // Port to listen on
}

// NewServer creates a new instance of the gRPC server with the provided storage and port.
// This follows the factory pattern for creating servers.
//
// Parameters:
//   - storage: An implementation of the NoteStorage interface
//   - port: The port number to listen on
//
// Returns:
//   - A pointer to a new Server instance
func NewServer(storage storage.NoteStorage, port int) *Server {
	return &Server{
		storage: storage,
		port:    port,
	}
}

// Start starts the gRPC server on the configured port.
// This is a simplified implementation for demonstration purposes.
// In a real-world application, this would set up a full gRPC server
// with the generated protobuf code.
//
// Returns:
//   - An error if the server fails to start
func (s *Server) Start() error {
	// In a real implementation, this would start a gRPC server
	// For demonstration purposes, we'll just print a message and set up a basic listener
	fmt.Printf("Starting gRPC server on port %d\n", s.port)

	// Create a TCP listener on the configured port
	// This is a mock implementation that would normally listen for gRPC requests
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	fmt.Printf("gRPC server listening on %s\n", listener.Addr())

	// In a real implementation, we would create a gRPC server and register our service
	// using the generated protobuf code, like this:
	//
	// server := grpc.NewServer()
	// proto.RegisterNotesServer(server, s)
	// return server.Serve(listener)

	// For demonstration purposes, we'll just close the listener
	return listener.Close()
}

// The following methods would normally implement the gRPC service interface
// In a real implementation, these would have the correct signatures based on the generated protobuf code
// from the proto/notes.proto file. For demonstration purposes, we're using simplified signatures.

// CreateNote creates a new note with the given title and content.
// This method would normally be called by the gRPC framework in response to a client request.
//
// Parameters:
//   - ctx: The context for the operation, which can include deadlines, cancellation signals, etc.
//   - title: The title of the new note
//   - content: The content of the new note
//
// Returns:
//   - The created note, including its generated ID and timestamps
//   - An error if the creation fails
func (s *Server) CreateNote(ctx context.Context, title, content string) (*model.Note, error) {
	// Create a new note with the provided title and content
	// This will generate a unique ID and set the creation/update timestamps
	note := model.NewNote(title, content)

	// Save the note to the storage
	if err := s.storage.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create note: %v", err)
	}

	return note, nil
}

// GetNote retrieves a note by its ID.
// This method would normally be called by the gRPC framework in response to a client request.
//
// Parameters:
//   - ctx: The context for the operation
//   - id: The ID of the note to retrieve
//
// Returns:
//   - The requested note if found
//   - An error if the note doesn't exist or if retrieval fails
func (s *Server) GetNote(ctx context.Context, id string) (*model.Note, error) {
	// Get the note from the storage
	note, err := s.storage.Get(ctx, id)
	if err != nil {
		// Handle specific error cases
		if err == storage.ErrNoteNotFound {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to retrieve note: %v", err)
	}

	return note, nil
}

// GetAllNotes retrieves all notes from the storage.
// This method would normally be called by the gRPC framework in response to a client request.
//
// Parameters:
//   - ctx: The context for the operation
//
// Returns:
//   - A slice of all notes, which may be empty if there are no notes
//   - An error if retrieval fails
func (s *Server) GetAllNotes(ctx context.Context) ([]*model.Note, error) {
	// Get all notes from the storage
	notes, err := s.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notes: %v", err)
	}

	return notes, nil
}

// UpdateNote updates an existing note with the given title and content.
// This method would normally be called by the gRPC framework in response to a client request.
//
// Parameters:
//   - ctx: The context for the operation
//   - id: The ID of the note to update
//   - title: The new title for the note
//   - content: The new content for the note
//
// Returns:
//   - The updated note
//   - An error if the note doesn't exist or if the update fails
func (s *Server) UpdateNote(ctx context.Context, id, title, content string) (*model.Note, error) {
	// First, get the existing note to make sure it exists
	existingNote, err := s.storage.Get(ctx, id)
	if err != nil {
		// Handle specific error cases
		if err == storage.ErrNoteNotFound {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to retrieve note: %v", err)
	}

	// Update the note's fields
	existingNote.Title = title
	existingNote.Content = content
	existingNote.UpdatedAt = time.Now() // Update the "last updated" timestamp

	// Save the updated note to the storage
	if err := s.storage.Update(ctx, existingNote); err != nil {
		return nil, fmt.Errorf("failed to update note: %v", err)
	}

	return existingNote, nil
}

// DeleteNote deletes a note by its ID.
// This method would normally be called by the gRPC framework in response to a client request.
//
// Parameters:
//   - ctx: The context for the operation
//   - id: The ID of the note to delete
//
// Returns:
//   - An error if the note doesn't exist or if deletion fails
//   - nil if deletion is successful
func (s *Server) DeleteNote(ctx context.Context, id string) error {
	// Delete the note from the storage
	if err := s.storage.Delete(ctx, id); err != nil {
		// Handle specific error cases
		if err == storage.ErrNoteNotFound {
			return fmt.Errorf("note not found")
		}
		return fmt.Errorf("failed to delete note: %v", err)
	}

	return nil
}
