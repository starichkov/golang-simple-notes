package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"
)

// Server implements the Notes gRPC service
type Server struct {
	storage storage.NoteStorage
	port    int
}

// NewServer creates a new instance of the gRPC server
func NewServer(storage storage.NoteStorage, port int) *Server {
	return &Server{
		storage: storage,
		port:    port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// In a real implementation, this would start a gRPC server
	// For demonstration purposes, we'll just print a message
	fmt.Printf("Starting gRPC server on port %d\n", s.port)

	// This is a mock implementation that would normally listen for gRPC requests
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	fmt.Printf("gRPC server listening on %s\n", listener.Addr())

	// In a real implementation, we would create a gRPC server and register our service
	// server := grpc.NewServer()
	// proto.RegisterNotesServer(server, s)
	// return server.Serve(listener)

	// For demonstration purposes, we'll just close the listener
	return listener.Close()
}

// The following methods would normally implement the gRPC service interface
// In a real implementation, these would have the correct signatures based on the generated protobuf code

// CreateNote creates a new note
func (s *Server) CreateNote(ctx context.Context, title, content string) (*model.Note, error) {
	note := model.NewNote(title, content)
	if err := s.storage.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create note: %v", err)
	}
	return note, nil
}

// GetNote retrieves a note by ID
func (s *Server) GetNote(ctx context.Context, id string) (*model.Note, error) {
	note, err := s.storage.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNoteNotFound {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to retrieve note: %v", err)
	}
	return note, nil
}

// GetAllNotes retrieves all notes
func (s *Server) GetAllNotes(ctx context.Context) ([]*model.Note, error) {
	notes, err := s.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notes: %v", err)
	}
	return notes, nil
}

// UpdateNote updates an existing note
func (s *Server) UpdateNote(ctx context.Context, id, title, content string) (*model.Note, error) {
	existingNote, err := s.storage.Get(ctx, id)
	if err != nil {
		if err == storage.ErrNoteNotFound {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to retrieve note: %v", err)
	}

	existingNote.Title = title
	existingNote.Content = content
	existingNote.UpdatedAt = time.Now()

	if err := s.storage.Update(ctx, existingNote); err != nil {
		return nil, fmt.Errorf("failed to update note: %v", err)
	}

	return existingNote, nil
}

// DeleteNote deletes a note
func (s *Server) DeleteNote(ctx context.Context, id string) error {
	if err := s.storage.Delete(ctx, id); err != nil {
		if err == storage.ErrNoteNotFound {
			return fmt.Errorf("note not found")
		}
		return fmt.Errorf("failed to delete note: %v", err)
	}

	return nil
}
