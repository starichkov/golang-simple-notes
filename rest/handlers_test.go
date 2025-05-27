package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
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
func (s *ErrorMockStorage) Create(ctx context.Context, note *model.Note) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// Get returns an error if shouldError is true
func (s *ErrorMockStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	if s.shouldError {
		return nil, errors.New("storage error")
	}
	return &model.Note{ID: id, Title: "Test Title", Content: "Test Content"}, nil
}

// GetAll returns an error if shouldError is true
func (s *ErrorMockStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	if s.shouldError {
		return nil, errors.New("storage error")
	}
	return []*model.Note{}, nil
}

// Update returns an error if shouldError is true
func (s *ErrorMockStorage) Update(ctx context.Context, note *model.Note) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// Delete returns an error if shouldError is true
func (s *ErrorMockStorage) Delete(ctx context.Context, id string) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// Close returns an error if shouldError is true
func (s *ErrorMockStorage) Close(ctx context.Context) error {
	if s.shouldError {
		return errors.New("storage error")
	}
	return nil
}

// setupTestRequest creates a test request with the given method, path, and body
func setupTestRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	chiCtx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	return req
}

// TestCreateNote tests the createNote handler
func TestCreateNote(t *testing.T) {
	// Test valid request
	t.Run("Valid Request", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Test Title","content":"Test Content"}`
		req := setupTestRequest("POST", "/api/notes", reqBody)
		w := httptest.NewRecorder()

		handler.handleNotes(w, req)

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

	// Test invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Test Title","content":"Test Content"`
		req := setupTestRequest("POST", "/api/notes", reqBody)
		w := httptest.NewRecorder()

		handler.handleNotes(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Invalid request body") {
			t.Errorf("Expected error message to contain 'Invalid request body', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		reqBody := `{"title":"Test Title","content":"Test Content"}`
		req := setupTestRequest("POST", "/api/notes", reqBody)
		w := httptest.NewRecorder()

		handler.handleNotes(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Failed to create note") {
			t.Errorf("Expected error message to contain 'Failed to create note', got: %s", w.Body.String())
		}
	})
}

// TestGetAllNotes tests the getAllNotes handler
func TestGetAllNotes(t *testing.T) {
	// Test getting all notes successfully
	t.Run("Success", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add some notes to the storage
		note1 := model.NewNote("Title 1", "Content 1")
		note2 := model.NewNote("Title 2", "Content 2")
		mockStorage.Create(context.Background(), note1)
		mockStorage.Create(context.Background(), note2)

		req := setupTestRequest("GET", "/api/notes", "")
		w := httptest.NewRecorder()

		handler.handleNotes(w, req)

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

		req := setupTestRequest("GET", "/api/notes", "")
		w := httptest.NewRecorder()

		handler.handleNotes(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Failed to get notes") {
			t.Errorf("Expected error message to contain 'Failed to get notes', got: %s", w.Body.String())
		}
	})
}

// TestGetNote tests the getNote handler
func TestGetNote(t *testing.T) {
	// Test getting a note successfully
	t.Run("Success", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("Test Title", "Test Content")
		mockStorage.Create(context.Background(), note)

		req := setupTestRequest("GET", "/api/notes/"+note.ID, "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", note.ID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response model.Note
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != note.ID {
			t.Errorf("Expected note ID %s, got %s", note.ID, response.ID)
		}
	})

	// Test note not found
	t.Run("Note Not Found", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		req := setupTestRequest("GET", "/api/notes/nonexistent", "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Note not found") {
			t.Errorf("Expected error message to contain 'Note not found', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		req := setupTestRequest("GET", "/api/notes/test", "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Failed to get note") {
			t.Errorf("Expected error message to contain 'Failed to get note', got: %s", w.Body.String())
		}
	})
}

// TestUpdateNote tests the updateNote handler
func TestUpdateNote(t *testing.T) {
	// Test updating a note successfully
	t.Run("Success", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("Original Title", "Original Content")
		mockStorage.Create(context.Background(), note)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := setupTestRequest("PUT", "/api/notes/"+note.ID, reqBody)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", note.ID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

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

	// Test note not found
	t.Run("Note Not Found", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := setupTestRequest("PUT", "/api/notes/nonexistent", reqBody)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Note not found") {
			t.Errorf("Expected error message to contain 'Note not found', got: %s", w.Body.String())
		}
	})

	// Test invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		reqBody := `{"title":"Updated Title","content":"Updated Content"`
		req := setupTestRequest("PUT", "/api/notes/test", reqBody)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Invalid request body") {
			t.Errorf("Expected error message to contain 'Invalid request body', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := setupTestRequest("PUT", "/api/notes/test", reqBody)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Failed to update note") {
			t.Errorf("Expected error message to contain 'Failed to update note', got: %s", w.Body.String())
		}
	})
}

// TestDeleteNote tests the deleteNote handler
func TestDeleteNote(t *testing.T) {
	// Test deleting a note successfully
	t.Run("Success", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		// Add a note to the storage
		note := model.NewNote("Test Title", "Test Content")
		mockStorage.Create(context.Background(), note)

		req := setupTestRequest("DELETE", "/api/notes/"+note.ID, "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", note.ID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
		}
	})

	// Test note not found
	t.Run("Note Not Found", func(t *testing.T) {
		mockStorage := NewMockStorage()
		handler := NewHandler(mockStorage)

		req := setupTestRequest("DELETE", "/api/notes/nonexistent", "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Note not found") {
			t.Errorf("Expected error message to contain 'Note not found', got: %s", w.Body.String())
		}
	})

	// Test storage error
	t.Run("Storage Error", func(t *testing.T) {
		errorStorage := NewErrorMockStorage(true)
		handler := NewHandler(errorStorage)

		req := setupTestRequest("DELETE", "/api/notes/test", "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.handleNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Failed to delete note") {
			t.Errorf("Expected error message to contain 'Failed to delete note', got: %s", w.Body.String())
		}
	})
}
