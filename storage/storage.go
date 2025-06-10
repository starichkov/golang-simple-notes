// Package storage provides interfaces and implementations for storing and retrieving notes.
// It follows the repository pattern, which abstracts the data access layer from the rest of the application.
// This allows the application to work with different storage backends (in-memory, CouchDB, MongoDB)
// without changing the core business logic.
package storage

import (
	"context"
	"errors"
	"sync"

	"golang-simple-notes/model"
)

// Common errors that can be returned by any storage implementation.
// Using predefined errors allows the application to handle specific error cases
// consistently across different storage backends.
var (
	// ErrNoteNotFound is returned when a note with the specified ID doesn't exist.
	ErrNoteNotFound = errors.New("note not found")
)

// NoteStorage defines the interface for note storage operations.
// Any storage implementation (in-memory, CouchDB, MongoDB) must implement this interface.
// This allows the application to switch between different storage backends without
// changing the core business logic.
type NoteStorage interface {
	// Create adds a new note to the storage.
	// It returns an error if the operation fails (e.g., if a note with the same ID already exists).
	Create(ctx context.Context, note *model.Note) error

	// Get retrieves a note by its ID.
	// It returns the note if found, or ErrNoteNotFound if no note with the specified ID exists.
	Get(ctx context.Context, id string) (*model.Note, error)

	// GetAll retrieves all notes from the storage.
	// It returns a slice of notes, which may be empty if there are no notes.
	GetAll(ctx context.Context) ([]*model.Note, error)

	// Update updates an existing note.
	// It returns ErrNoteNotFound if no note with the specified ID exists.
	Update(ctx context.Context, note *model.Note) error

	// Delete removes a note from the storage.
	// It returns ErrNoteNotFound if no note with the specified ID exists.
	Delete(ctx context.Context, id string) error

	// Close closes any resources used by the storage (e.g., database connections).
	// It should be called when the application is shutting down.
	Close(ctx context.Context) error
}

// InMemoryStorage implements NoteStorage using an in-memory map.
// This is the simplest storage implementation, useful for development and testing.
// It stores notes in memory, so they are lost when the application restarts.
type InMemoryStorage struct {
	notes map[string]*model.Note // Map of note ID to note
	mutex sync.RWMutex           // Mutex to protect concurrent access to the map
}

// NewInMemoryStorage creates a new instance of InMemoryStorage.
// It initializes the notes map and returns a ready-to-use storage instance.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		notes: make(map[string]*model.Note), // Initialize an empty map
	}
}

// Create adds a new note to the storage.
// In this implementation, it simply adds the note to the map using its ID as the key.
// This method is thread-safe due to the use of a mutex.
func (s *InMemoryStorage) Create(ctx context.Context, note *model.Note) error {
	s.mutex.Lock()         // Lock for writing
	defer s.mutex.Unlock() // Ensure the lock is released when the function returns

	// Store the note in the map using its ID as the key
	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID.
// It returns the note if found, or ErrNoteNotFound if no note with the specified ID exists.
// This method is thread-safe due to the use of a mutex.
func (s *InMemoryStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	s.mutex.RLock()         // Lock for reading (allows concurrent reads)
	defer s.mutex.RUnlock() // Ensure the lock is released when the function returns

	// Look up the note in the map
	note, exists := s.notes[id]
	if !exists {
		return nil, ErrNoteNotFound // Return error if note doesn't exist
	}
	return note, nil
}

// GetAll retrieves all notes from the storage.
// It returns a slice of all notes in the storage, which may be empty if there are no notes.
// This method is thread-safe due to the use of a mutex.
func (s *InMemoryStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	s.mutex.RLock()         // Lock for reading (allows concurrent reads)
	defer s.mutex.RUnlock() // Ensure the lock is released when the function returns

	// Create a slice with capacity equal to the number of notes
	notes := make([]*model.Note, 0, len(s.notes))

	// Add each note from the map to the slice
	for _, note := range s.notes {
		notes = append(notes, note)
	}

	return notes, nil
}

// Update updates an existing note.
// It returns ErrNoteNotFound if no note with the specified ID exists.
// This method is thread-safe due to the use of a mutex.
func (s *InMemoryStorage) Update(ctx context.Context, note *model.Note) error {
	s.mutex.Lock()         // Lock for writing
	defer s.mutex.Unlock() // Ensure the lock is released when the function returns

	// Check if the note exists
	if _, exists := s.notes[note.ID]; !exists {
		return ErrNoteNotFound // Return error if note doesn't exist
	}

	// Update the note in the map
	s.notes[note.ID] = note
	return nil
}

// Delete removes a note from the storage.
// It returns ErrNoteNotFound if no note with the specified ID exists.
// This method is thread-safe due to the use of a mutex.
func (s *InMemoryStorage) Delete(ctx context.Context, id string) error {
	s.mutex.Lock()         // Lock for writing
	defer s.mutex.Unlock() // Ensure the lock is released when the function returns

	// Check if the note exists
	if _, exists := s.notes[id]; !exists {
		return ErrNoteNotFound // Return error if note doesn't exist
	}

	// Remove the note from the map
	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage.
// For the in-memory implementation, there are no resources to close,
// so this method does nothing and always returns nil.
func (s *InMemoryStorage) Close(ctx context.Context) error {
	// Nothing to close for in-memory storage
	return nil
}
