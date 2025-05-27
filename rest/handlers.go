package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for notes
type Handler struct {
	storage storage.NoteStorage
}

// NewHandler creates a new Handler instance
func NewHandler(storage storage.NoteStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// RegisterRoutes registers the handler's routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/api/notes", h.handleNotes)
	mux.HandleFunc("/api/notes/", h.handleNote)
}

// handleHealth handles the health check endpoint
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleNotes handles requests for the /api/notes endpoint
func (h *Handler) handleNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAllNotes(w, r)
	case http.MethodPost:
		h.createNote(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// isValidNoteID checks if a note ID is valid
func isValidNoteID(id string) bool {
	// ID should not be empty
	if id == "" {
		return false
	}

	// ID should not contain spaces
	if strings.Contains(id, " ") {
		return false
	}

	// ID should not be too long (e.g., max 255 characters)
	if len(id) > 255 {
		return false
	}

	// ID should only contain alphanumeric characters, hyphens, and underscores
	for _, c := range id {
		if !isValidIDChar(c) {
			return false
		}
	}

	return true
}

// isValidIDChar checks if a character is valid in a note ID
func isValidIDChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '_'
}

// handleNote handles requests for the /api/notes/{id} endpoint
func (h *Handler) handleNote(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Note ID is required", http.StatusBadRequest)
			return
		}
		if !isValidNoteID(id) {
			http.Error(w, "Invalid note ID format", http.StatusBadRequest)
			return
		}
		h.getNote(w, r, id)
	case http.MethodPut:
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Note ID is required", http.StatusBadRequest)
			return
		}
		if !isValidNoteID(id) {
			http.Error(w, "Invalid note ID format", http.StatusBadRequest)
			return
		}
		h.updateNote(w, r, id)
	case http.MethodDelete:
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Note ID is required", http.StatusBadRequest)
			return
		}
		if !isValidNoteID(id) {
			http.Error(w, "Invalid note ID format", http.StatusBadRequest)
			return
		}
		h.deleteNote(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getAllNotes handles GET /api/notes
func (h *Handler) getAllNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.storage.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to get notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		http.Error(w, "Failed to encode notes", http.StatusInternalServerError)
		return
	}
}

// getNote handles GET /api/notes/{id}
func (h *Handler) getNote(w http.ResponseWriter, r *http.Request, id string) {
	note, err := h.storage.Get(r.Context(), id)
	if err != nil {
		if err == storage.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, "Failed to encode note", http.StatusInternalServerError)
		return
	}
}

// createNote handles POST /api/notes
func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var note model.Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.storage.Create(r.Context(), &note); err != nil {
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, "Failed to encode note", http.StatusInternalServerError)
		return
	}
}

// updateNote handles PUT /api/notes/{id}
func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request, id string) {
	var note model.Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	note.ID = id
	if err := h.storage.Update(r.Context(), &note); err != nil {
		if err == storage.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, "Failed to encode note", http.StatusInternalServerError)
		return
	}
}

// deleteNote handles DELETE /api/notes/{id}
func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.storage.Delete(r.Context(), id); err != nil {
		if err == storage.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
