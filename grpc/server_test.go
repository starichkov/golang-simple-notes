package grpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"
)

// MockStorage is a mock implementation of the NoteStorage interface for testing
type MockStorage struct {
	notes map[string]*model.Note
}

// NewMockStorage creates a new instance of MockStorage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		notes: make(map[string]*model.Note),
	}
}

// Create adds a new note to the storage
func (s *MockStorage) Create(ctx context.Context, note *model.Note) error {
	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID
func (s *MockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	note, exists := s.notes[id]
	if !exists {
		return nil, storage.ErrNoteNotFound
	}
	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *MockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	notes := make([]*model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

// Update updates an existing note
func (s *MockStorage) Update(ctx context.Context, note *model.Note) error {
	if _, exists := s.notes[note.ID]; !exists {
		return storage.ErrNoteNotFound
	}
	s.notes[note.ID] = note
	return nil
}

// Delete removes a note from the storage
func (s *MockStorage) Delete(ctx context.Context, id string) error {
	if _, exists := s.notes[id]; !exists {
		return storage.ErrNoteNotFound
	}
	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage
func (s *MockStorage) Close(ctx context.Context) error {
	return nil
}

// FailingMockStorage is a mock implementation that always returns errors
type FailingMockStorage struct{}

// NewFailingMockStorage creates a new instance of FailingMockStorage
func NewFailingMockStorage() *FailingMockStorage {
	return &FailingMockStorage{}
}

// Create always returns an error
func (s *FailingMockStorage) Create(ctx context.Context, note *model.Note) error {
	return errors.New("mock storage create error")
}

// Get always returns an error
func (s *FailingMockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	return nil, errors.New("mock storage get error")
}

// GetAll always returns an error
func (s *FailingMockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	return nil, errors.New("mock storage getall error")
}

// Update always returns an error
func (s *FailingMockStorage) Update(ctx context.Context, note *model.Note) error {
	return errors.New("mock storage update error")
}

// Delete always returns an error
func (s *FailingMockStorage) Delete(ctx context.Context, id string) error {
	return errors.New("mock storage delete error")
}

// Close always returns an error
func (s *FailingMockStorage) Close(ctx context.Context) error {
	return errors.New("mock storage close error")
}

// TestNewServer tests the creation of a new server
func TestNewServer(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.storage != mockStorage {
		t.Error("Expected server to use the provided storage")
	}

	if server.port != 8081 {
		t.Errorf("Expected port to be 8081, got %d", server.port)
	}
}

// TestStart tests the Start method
func TestStart(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)

	// Since the Start method is a mock implementation that just returns nil,
	// we just verify that it doesn't return an error
	err := server.Start()
	if err != nil {
		t.Errorf("Expected Start to return nil, got %v", err)
	}
}

// TestStartError tests error handling in the Start method
func TestStartError(t *testing.T) {
	mockStorage := NewMockStorage()

	// First, create a listener on the port we want to use
	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		t.Fatalf("Failed to create listener for test: %v", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			t.Errorf("Failed to close listener: %v", err)
		}
	}(listener)

	// Now try to start a server on the same port, which should fail
	server := NewServer(mockStorage, 8082)
	err = server.Start()

	// We expect an error because the port is already in use
	if err == nil {
		t.Fatal("Expected error when starting server on already used port, got nil")
	}
}

// TestCreateNote tests the CreateNote method
func TestCreateNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)
	ctx := context.Background()

	note, err := server.CreateNote(ctx, "Test Title", "Test Content")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if note == nil {
		t.Fatal("Expected note to be created, got nil")
	}

	if note.Title != "Test Title" {
		t.Errorf("Expected title to be 'Test Title', got '%s'", note.Title)
	}

	if note.Content != "Test Content" {
		t.Errorf("Expected content to be 'Test Content', got '%s'", note.Content)
	}

	// Verify the note was added to storage
	retrieved, err := mockStorage.Get(ctx, note.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve note from storage: %v", err)
	}

	if retrieved.ID != note.ID {
		t.Errorf("Expected ID %s, got %s", note.ID, retrieved.ID)
	}
}

// TestGetNote tests the GetNote method
func TestGetNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)
	ctx := context.Background()

	// Create a note
	originalNote := model.NewNote("Test Title", "Test Content")
	err := mockStorage.Create(ctx, originalNote)
	if err != nil {
		return
	}

	// Test getting an existing note
	note, err := server.GetNote(ctx, originalNote.ID)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}

	if note == nil {
		t.Fatal("Expected note to be retrieved, got nil")
	}

	if note.ID != originalNote.ID {
		t.Errorf("Expected ID %s, got %s", originalNote.ID, note.ID)
	}

	if note.Title != originalNote.Title {
		t.Errorf("Expected title to be '%s', got '%s'", originalNote.Title, note.Title)
	}

	// Test getting a non-existent note
	_, err = server.GetNote(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error when getting non-existent note, got nil")
	}
}

// TestGetAllNotes tests the GetAllNotes method
func TestGetAllNotes(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)
	ctx := context.Background()

	// Create some notes
	note1 := model.NewNote("Title 1", "Content 1")
	note2 := model.NewNote("Title 2", "Content 2")
	err1 := mockStorage.Create(ctx, note1)
	if err1 != nil {
		return
	}
	err2 := mockStorage.Create(ctx, note2)
	if err2 != nil {
		return
	}

	// Get all notes
	notes, err := server.GetAllNotes(ctx)
	if err != nil {
		t.Fatalf("Failed to get all notes: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}

	// Check that both notes are in the result
	found1, found2 := false, false
	for _, note := range notes {
		if note.ID == note1.ID {
			found1 = true
		}
		if note.ID == note2.ID {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Not all notes were found in the result")
	}
}

// TestUpdateNote tests the UpdateNote method
func TestUpdateNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)
	ctx := context.Background()

	// Create a note
	originalNote := model.NewNote("Original Title", "Original Content")
	err := mockStorage.Create(ctx, originalNote)
	if err != nil {
		return
	}

	// Update the note
	updatedNote, err := server.UpdateNote(ctx, originalNote.ID, "Updated Title", "Updated Content")
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	if updatedNote == nil {
		t.Fatal("Expected updated note to be returned, got nil")
	}

	if updatedNote.Title != "Updated Title" {
		t.Errorf("Expected title to be 'Updated Title', got '%s'", updatedNote.Title)
	}

	if updatedNote.Content != "Updated Content" {
		t.Errorf("Expected content to be 'Updated Content', got '%s'", updatedNote.Content)
	}

	// Verify the note was updated in storage
	retrieved, err := mockStorage.Get(ctx, originalNote.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated note: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected title to be 'Updated Title', got '%s'", retrieved.Title)
	}
}

// TestDeleteNote tests the DeleteNote method
func TestDeleteNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)
	ctx := context.Background()

	// Create a note
	note := model.NewNote("Test Title", "Test Content")
	errc := mockStorage.Create(ctx, note)
	if errc != nil {
		return
	}

	// Delete the note
	err := server.DeleteNote(ctx, note.ID)
	if err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	// Verify the note was deleted
	_, err = mockStorage.Get(ctx, note.ID)
	if err != storage.ErrNoteNotFound {
		t.Error("Expected note to be deleted")
	}
}

// TestCreateNoteError tests error handling in CreateNote
func TestCreateNoteError(t *testing.T) {
	failingStorage := NewFailingMockStorage()
	server := NewServer(failingStorage, 8081)
	ctx := context.Background()

	_, err := server.CreateNote(ctx, "Test Title", "Test Content")
	if err == nil {
		t.Error("Expected error when creating note with failing storage")
	}
}

// TestGetNoteError tests error handling in GetNote
func TestGetNoteError(t *testing.T) {
	failingStorage := NewFailingMockStorage()
	server := NewServer(failingStorage, 8081)
	ctx := context.Background()

	_, err := server.GetNote(ctx, "test-id")
	if err == nil {
		t.Error("Expected error when getting note with failing storage")
	}
}

// TestGetAllNotesError tests error handling in GetAllNotes
func TestGetAllNotesError(t *testing.T) {
	failingStorage := NewFailingMockStorage()
	server := NewServer(failingStorage, 8081)
	ctx := context.Background()

	_, err := server.GetAllNotes(ctx)
	if err == nil {
		t.Error("Expected error when getting all notes with failing storage")
	}
}

// TestUpdateNoteError tests error handling in UpdateNote
func TestUpdateNoteError(t *testing.T) {
	failingStorage := NewFailingMockStorage()
	server := NewServer(failingStorage, 8081)
	ctx := context.Background()

	_, err := server.UpdateNote(ctx, "test-id", "New Title", "New Content")
	if err == nil {
		t.Error("Expected error when updating note with failing storage")
	}
}

// TestDeleteNoteError tests error handling in DeleteNote
func TestDeleteNoteError(t *testing.T) {
	failingStorage := NewFailingMockStorage()
	server := NewServer(failingStorage, 8081)
	ctx := context.Background()

	err := server.DeleteNote(ctx, "test-id")
	if err == nil {
		t.Error("Expected error when deleting note with failing storage")
	}
}
