package storage

import (
	"errors"
	"sync"

	"golang-simple-notes/model"
)

// Common errors
var (
	ErrNoteNotFound = errors.New("note not found")
)

// NoteStorage defines the interface for note storage operations
type NoteStorage interface {
	// Create adds a new note to the storage
	Create(note *model.Note) error

	// Get retrieves a note by its ID
	Get(id string) (*model.Note, error)

	// GetAll retrieves all notes from the storage
	GetAll() ([]*model.Note, error)

	// Update updates an existing note
	Update(note *model.Note) error

	// Delete removes a note from the storage
	Delete(id string) error

	// Close closes any resources used by the storage
	Close() error
}

// InMemoryStorage implements NoteStorage using an in-memory map
type InMemoryStorage struct {
	notes map[string]*model.Note
	mutex sync.RWMutex
}

// NewInMemoryStorage creates a new instance of InMemoryStorage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		notes: make(map[string]*model.Note),
	}
}

// Create adds a new note to the storage
func (s *InMemoryStorage) Create(note *model.Note) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID
func (s *InMemoryStorage) Get(id string) (*model.Note, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	note, exists := s.notes[id]
	if !exists {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *InMemoryStorage) GetAll() ([]*model.Note, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	notes := make([]*model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

// Update updates an existing note
func (s *InMemoryStorage) Update(note *model.Note) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.notes[note.ID]; !exists {
		return ErrNoteNotFound
	}

	s.notes[note.ID] = note
	return nil
}

// Delete removes a note from the storage
func (s *InMemoryStorage) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.notes[id]; !exists {
		return ErrNoteNotFound
	}

	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage
func (s *InMemoryStorage) Close() error {
	// Nothing to close for in-memory storage
	return nil
}
