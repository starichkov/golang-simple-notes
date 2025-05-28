package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"

	"github.com/go-chi/chi/v5"
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

		handler.createNote(w, req)

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

		handler.createNote(w, req)

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

		handler.createNote(w, req)

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

		handler.getAllNotes(w, req)

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

		handler.getAllNotes(w, req)

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

		// Add a note to the storage with a valid ID
		note := &model.Note{ID: "testid123", Title: "Test Title", Content: "Test Content"}
		mockStorage.Create(context.Background(), note)

		req := setupTestRequest("GET", "/api/notes/"+note.ID, "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", note.ID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.getNote(w, req)

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

		handler.getNote(w, req)

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

		handler.getNote(w, req)

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

		// Add a note to the storage with a valid ID
		note := &model.Note{ID: "testid123", Title: "Original Title", Content: "Original Content"}
		mockStorage.Create(context.Background(), note)

		reqBody := `{"title":"Updated Title","content":"Updated Content"}`
		req := setupTestRequest("PUT", "/api/notes/"+note.ID, reqBody)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", note.ID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.updateNote(w, req)

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

		handler.updateNote(w, req)

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

		handler.updateNote(w, req)

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

		handler.updateNote(w, req)

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

		// Add a note to the storage with a valid ID
		note := &model.Note{ID: "testid123", Title: "Test Title", Content: "Test Content"}
		mockStorage.Create(context.Background(), note)

		req := setupTestRequest("DELETE", "/api/notes/"+note.ID, "")
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", note.ID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		w := httptest.NewRecorder()

		handler.deleteNote(w, req)

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

		handler.deleteNote(w, req)

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

		handler.deleteNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !strings.Contains(w.Body.String(), "Failed to delete note") {
			t.Errorf("Expected error message to contain 'Failed to delete note', got: %s", w.Body.String())
		}
	})
}

// TestHealthEndpoint tests the /health endpoint
func TestHealthEndpoint(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	t.Run("GET /health", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code 405, got %d", w.Code)
		}
	})
}

// TestEmptyNoteID tests that getNote returns 400 Bad Request when the ID is empty
func TestEmptyNoteID(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)

	req := setupTestRequest("GET", "/api/notes/", "")
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	w := httptest.NewRecorder()

	handler.getNote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Note ID is required") {
		t.Errorf("Expected error message to contain 'Note ID is required', got: %s", w.Body.String())
	}
}

// TestUnsupportedMethods tests that unsupported methods return 405 Method Not Allowed
func TestUnsupportedMethods(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Test unsupported methods on /api/notes
	unsupportedMethods := []string{"PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, method := range unsupportedMethods {
		t.Run("Notes Endpoint - "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/notes", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status code 405 for %s, got %d", method, w.Code)
			}
		})
	}

	// Test unsupported methods on /api/notes/{id}
	unsupportedNoteIDMethods := []string{"PATCH", "OPTIONS", "HEAD"}
	for _, method := range unsupportedNoteIDMethods {
		t.Run("Note Endpoint - "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/notes/test-id", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status code 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestIsValidNoteID covers all branches of isValidNoteID
func TestIsValidNoteID(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
		name  string
	}{
		{"", false, "Empty"},
		{"valid_id-123", true, "Valid"},
		{"has space", false, "Space"},
		{strings.Repeat("a", 256), false, "TooLong"},
		{"invalid@id", false, "InvalidChar"},
		{"UPPER_and-lower123", true, "MixedCase"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isValidNoteID(c.id); got != c.valid {
				t.Errorf("isValidNoteID(%q) = %v, want %v", c.id, got, c.valid)
			}
		})
	}
}

// TestEmptyAndInvalidID tests empty and invalid IDs for PUT and DELETE methods
func TestEmptyAndInvalidID(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewHandler(mockStorage)

	invalidIDs := []struct {
		id   string
		msg  string
		code int
	}{
		{"", "Note ID is required", http.StatusBadRequest},
		{"invalid@id", "Invalid note ID format", http.StatusBadRequest},
		{"has space", "Invalid note ID format", http.StatusBadRequest},
		{strings.Repeat("a", 256), "Invalid note ID format", http.StatusBadRequest},
	}

	// Test PUT method with updateNote
	for _, tc := range invalidIDs {
		t.Run("PUT/"+tc.id, func(t *testing.T) {
			encodedID := tc.id
			if encodedID != "" {
				encodedID = url.PathEscape(tc.id)
			}
			req := setupTestRequest("PUT", "/api/notes/"+encodedID, `{"title":"Test","content":"Test"}`)
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tc.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			w := httptest.NewRecorder()

			handler.updateNote(w, req)

			if w.Code != tc.code {
				t.Errorf("Expected status code %d, got %d", tc.code, w.Code)
			}
			if !strings.Contains(w.Body.String(), tc.msg) {
				t.Errorf("Expected error message to contain '%s', got: %s", tc.msg, w.Body.String())
			}
		})
	}

	// Test DELETE method with deleteNote
	for _, tc := range invalidIDs {
		t.Run("DELETE/"+tc.id, func(t *testing.T) {
			encodedID := tc.id
			if encodedID != "" {
				encodedID = url.PathEscape(tc.id)
			}
			req := setupTestRequest("DELETE", "/api/notes/"+encodedID, "")
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tc.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			w := httptest.NewRecorder()

			handler.deleteNote(w, req)

			if w.Code != tc.code {
				t.Errorf("Expected status code %d, got %d", tc.code, w.Code)
			}
			if !strings.Contains(w.Body.String(), tc.msg) {
				t.Errorf("Expected error message to contain '%s', got: %s", tc.msg, w.Body.String())
			}
		})
	}
}
