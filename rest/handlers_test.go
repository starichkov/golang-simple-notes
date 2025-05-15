package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

// ErrorMockStorage is a mock implementation that returns errors for testing error handling
type ErrorMockStorage struct {
	shouldError bool
}

// NewErrorMockStorage creates a new instance of ErrorMockStorage
func NewErrorMockStorage(shouldError bool) *ErrorMockStorage {
	return &ErrorMockStorage{
		shouldError: shouldError,
	}
}

// Create returns an error if shouldError is true
func (s *ErrorMockStorage) Create(note *model.Note) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// Get returns an error if shouldError is true
func (s *ErrorMockStorage) Get(id string) (*model.Note, error) {
	if s.shouldError {
		return nil, errors.New("storage error")
	}
	return &model.Note{ID: id, Title: "Test Title", Content: "Test Content"}, nil
}

// GetAll returns an error if shouldError is true
func (s *ErrorMockStorage) GetAll() ([]*model.Note, error) {
	if s.shouldError {
		return nil, errors.New("storage error")
	}
	return []*model.Note{}, nil
}

// Update returns an error if shouldError is true
func (s *ErrorMockStorage) Update(note *model.Note) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// Delete returns an error if shouldError is true
func (s *ErrorMockStorage) Delete(id string) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// Close returns an error if shouldError is true
func (s *ErrorMockStorage) Close() error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// TestCreateNote tests the CreateNote handler
func TestCreateNote(t *testing.T) {
	// Test valid request
	t.Run("Valid Request", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Test Title","content":"Test Content"}`
		req := httptest.NewRequest("POST", "/api/notes", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var response model.Note
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Title != "Test Title" {
			t.Errorf("Expected title 'Test Title', got '%s'", response.Title)
		}

		if response.Content != "Test Content" {
			t.Errorf("Expected content 'Test Content', got '%s'", response.Content)
		}
	})

	// Test invalid request (missing title)
	t.Run("Missing Title", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"content":"Test Content"}`
		req := httptest.NewRequest("POST", "/api/notes", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Title is required") {
			t.Errorf("Expected error message to contain 'Title is required', got: %s", w.Body.String())
		}
	})

	// Test invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Test Title","content":"Test Content"`
		req := httptest.NewRequest("POST", "/api/notes", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Invalid request body") {
			t.Errorf("Expected error message to contain 'Invalid request body', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		reqBody := `{"title":"Test Title","content":"Test Content"}`
		req := httptest.NewRequest("POST", "/api/notes", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Failed to create note") {
			t.Errorf("Expected error message to contain 'Failed to create note', got: %s", w.Body.String())
		}
	})
}

// TestGetAllNotes tests the GetAllNotes handler
func TestGetAllNotes(t *testing.T) {
	// Test getting all notes successfully
	t.Run("Success", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add some notes to the storage
		note1 := model.NewNote("Title 1", "Content 1")
		note2 := model.NewNote("Title 2", "Content 2")
		mockStorage.Create(note1)
		mockStorage.Create(note2)

		// Test getting all notes
		req := httptest.NewRequest("GET", "/api/notes", nil)
		w := httptest.NewRecorder()

		handler.GetAllNotes(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response []*model.Note
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(response))
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		req := httptest.NewRequest("GET", "/api/notes", nil)
		w := httptest.NewRecorder()

		handler.GetAllNotes(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Failed to retrieve notes") {
			t.Errorf("Expected error message to contain 'Failed to retrieve notes', got: %s", w.Body.String())
		}
	})
}

// TestGetNote tests the GetNote handler
func TestGetNote(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)

	// Add a note to the storage
	note := model.NewNote("Test Title", "Test Content")
	mockStorage.Create(note)

	// Test getting an existing note
	t.Run("Existing Note", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes/"+note.ID, nil)
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response model.Note
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != note.ID {
			t.Errorf("Expected ID %s, got %s", note.ID, response.ID)
		}

		if response.Title != note.Title {
			t.Errorf("Expected title '%s', got '%s'", note.Title, response.Title)
		}
	})

	// Test getting a non-existent note
	t.Run("Non-existent Note", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes/non-existent", nil)
		req.URL.Path = "/api/notes/non-existent"
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	// Test with missing ID
	t.Run("Missing ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes/", nil)
		req.URL.Path = "/api/notes/"
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestUpdateNote tests the UpdateNote handler
func TestUpdateNote(t *testing.T) {
	// Test updating an existing note
	t.Run("Existing Note", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("Original Title", "Original Content")
		mockStorage.Create(note)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/"+note.ID, bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response model.Note
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", response.Title)
		}

		if response.Content != "Updated Content" {
			t.Errorf("Expected content 'Updated Content', got '%s'", response.Content)
		}
	})

	// Test updating a non-existent note
	t.Run("Non-existent Note", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/non-existent", bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/non-existent"
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Note not found") {
			t.Errorf("Expected error message to contain 'Note not found', got: %s", w.Body.String())
		}
	})

	// Test with missing title
	t.Run("Missing Title", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("Original Title", "Original Content")
		mockStorage.Create(note)

		reqBody := `{"content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/"+note.ID, bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Title is required") {
			t.Errorf("Expected error message to contain 'Title is required', got: %s", w.Body.String())
		}
	})

	// Test invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("Original Title", "Original Content")
		mockStorage.Create(note)

		reqBody := `{"title":"Updated Title","content":"Updated Content"`
		req := httptest.NewRequest("PUT", "/api/notes/"+note.ID, bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Invalid request body") {
			t.Errorf("Expected error message to contain 'Invalid request body', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		// Create a mock storage that returns a note but fails on update
		mockStorage := NewMockStorage()
		note := model.NewNote("Original Title", "Original Content")
		mockStorage.Create(note)

		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/"+note.ID, bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Failed to") {
			t.Errorf("Expected error message to contain 'Failed to', got: %s", w.Body.String())
		}
	})

	// Test missing ID
	t.Run("Missing ID", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/", bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/"
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Note ID is required") {
			t.Errorf("Expected error message to contain 'Note ID is required', got: %s", w.Body.String())
		}
	})
}

// TestDeleteNote tests the DeleteNote handler
func TestDeleteNote(t *testing.T) {
	// Test deleting an existing note
	t.Run("Existing Note", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("To Delete", "This note will be deleted")
		mockStorage.Create(note)

		req := httptest.NewRequest("DELETE", "/api/notes/"+note.ID, nil)
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
		}

		// Verify the note was deleted
		_, err := mockStorage.Get(note.ID)
		if err != storage.ErrNoteNotFound {
			t.Errorf("Expected note to be deleted, but it still exists")
		}
	})

	// Test deleting a non-existent note
	t.Run("Non-existent Note", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		req := httptest.NewRequest("DELETE", "/api/notes/non-existent", nil)
		req.URL.Path = "/api/notes/non-existent"
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Note not found") {
			t.Errorf("Expected error message to contain 'Note not found', got: %s", w.Body.String())
		}
	})

	// Test missing ID
	t.Run("Missing ID", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		req := httptest.NewRequest("DELETE", "/api/notes/", nil)
		req.URL.Path = "/api/notes/"
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Note ID is required") {
			t.Errorf("Expected error message to contain 'Note ID is required', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		req := httptest.NewRequest("DELETE", "/api/notes/test-id", nil)
		req.URL.Path = "/api/notes/test-id"
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		// Check error message
		if !strings.Contains(w.Body.String(), "Failed to delete note") {
			t.Errorf("Expected error message to contain 'Failed to delete note', got: %s", w.Body.String())
		}
	})
}
