package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// TestCreateNote tests the CreateNote handler
func TestCreateNote(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)

	// Test valid request
	t.Run("Valid Request", func(t *testing.T) {
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
		reqBody := `{"content":"Test Content"}`
		req := httptest.NewRequest("POST", "/api/notes", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	// Test invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		reqBody := `{"title":"Test Title","content":"Test Content"`
		req := httptest.NewRequest("POST", "/api/notes", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestGetAllNotes tests the GetAllNotes handler
func TestGetAllNotes(t *testing.T) {
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
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)

	// Add a note to the storage
	note := model.NewNote("Original Title", "Original Content")
	mockStorage.Create(note)

	// Test updating an existing note
	t.Run("Existing Note", func(t *testing.T) {
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
		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/non-existent", bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/non-existent"
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	// Test with missing title
	t.Run("Missing Title", func(t *testing.T) {
		reqBody := `{"content":"Updated Content"}`
		req := httptest.NewRequest("PUT", "/api/notes/"+note.ID, bytes.NewBufferString(reqBody))
		req.URL.Path = "/api/notes/" + note.ID
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestDeleteNote tests the DeleteNote handler
func TestDeleteNote(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)

	// Add a note to the storage
	note := model.NewNote("To Delete", "This note will be deleted")
	mockStorage.Create(note)

	// Test deleting an existing note
	t.Run("Existing Note", func(t *testing.T) {
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
		req := httptest.NewRequest("DELETE", "/api/notes/non-existent", nil)
		req.URL.Path = "/api/notes/non-existent"
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}
