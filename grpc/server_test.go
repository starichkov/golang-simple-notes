package grpc

import (
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
func (s *MockStorage) Create(note *model.Note) error {
	s.notes[note.ID] = note
	return nil
}

// Get retrieves a note by its ID
func (s *MockStorage) Get(id string) (*model.Note, error) {
	note, exists := s.notes[id]
	if !exists {
		return nil, storage.ErrNoteNotFound
	}
	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *MockStorage) GetAll() ([]*model.Note, error) {
	notes := make([]*model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

// Update updates an existing note
func (s *MockStorage) Update(note *model.Note) error {
	if _, exists := s.notes[note.ID]; !exists {
		return storage.ErrNoteNotFound
	}
	s.notes[note.ID] = note
	return nil
}

// Delete removes a note from the storage
func (s *MockStorage) Delete(id string) error {
	if _, exists := s.notes[id]; !exists {
		return storage.ErrNoteNotFound
	}
	delete(s.notes, id)
	return nil
}

// Close closes any resources used by the storage
func (s *MockStorage) Close() error {
	return nil
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

// TestCreateNote tests the CreateNote method
func TestCreateNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)

	note, err := server.CreateNote("Test Title", "Test Content")
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
	retrieved, err := mockStorage.Get(note.ID)
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

	// Create a note
	originalNote := model.NewNote("Test Title", "Test Content")
	mockStorage.Create(originalNote)

	// Test getting an existing note
	note, err := server.GetNote(originalNote.ID)
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
	_, err = server.GetNote("non-existent-id")
	if err == nil {
		t.Error("Expected error when getting non-existent note, got nil")
	}
}

// TestGetAllNotes tests the GetAllNotes method
func TestGetAllNotes(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)

	// Create some notes
	note1 := model.NewNote("Title 1", "Content 1")
	note2 := model.NewNote("Title 2", "Content 2")
	mockStorage.Create(note1)
	mockStorage.Create(note2)

	// Get all notes
	notes, err := server.GetAllNotes()
	if err != nil {
		t.Fatalf("Failed to get all notes: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}

	// Check that both notes are in the result
	found1, found2 := false, false
	for _, n := range notes {
		if n.ID == note1.ID {
			found1 = true
		}
		if n.ID == note2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Note 1 not found in GetAllNotes results")
	}

	if !found2 {
		t.Error("Note 2 not found in GetAllNotes results")
	}
}

// TestUpdateNote tests the UpdateNote method
func TestUpdateNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)

	// Create a note
	originalNote := model.NewNote("Original Title", "Original Content")
	mockStorage.Create(originalNote)

	// Update the note
	updatedNote, err := server.UpdateNote(originalNote.ID, "Updated Title", "Updated Content")
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	if updatedNote == nil {
		t.Fatal("Expected note to be updated, got nil")
	}

	if updatedNote.ID != originalNote.ID {
		t.Errorf("Expected ID to remain %s, got %s", originalNote.ID, updatedNote.ID)
	}

	if updatedNote.Title != "Updated Title" {
		t.Errorf("Expected title to be 'Updated Title', got '%s'", updatedNote.Title)
	}

	if updatedNote.Content != "Updated Content" {
		t.Errorf("Expected content to be 'Updated Content', got '%s'", updatedNote.Content)
	}

	// Verify the note was updated in storage
	retrieved, err := mockStorage.Get(originalNote.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve note from storage: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected title in storage to be 'Updated Title', got '%s'", retrieved.Title)
	}

	// Test updating a non-existent note
	_, err = server.UpdateNote("non-existent-id", "New Title", "New Content")
	if err == nil {
		t.Error("Expected error when updating non-existent note, got nil")
	}
}

// TestDeleteNote tests the DeleteNote method
func TestDeleteNote(t *testing.T) {
	mockStorage := NewMockStorage()
	server := NewServer(mockStorage, 8081)

	// Create a note
	note := model.NewNote("To Delete", "This note will be deleted")
	mockStorage.Create(note)

	// Delete the note
	err := server.DeleteNote(note.ID)
	if err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	// Verify the note was deleted from storage
	_, err = mockStorage.Get(note.ID)
	if err != storage.ErrNoteNotFound {
		t.Errorf("Expected note to be deleted, but it still exists")
	}

	// Test deleting a non-existent note
	err = server.DeleteNote("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent note, got nil")
	}
}
