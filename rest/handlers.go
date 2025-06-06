package rest

import (
	"encoding/json"
	"golang-simple-notes/model"
	"golang-simple-notes/storage"
	"net/http"

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
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.handleHealth)

	r.Route("/api/notes", func(r chi.Router) {
		r.Get("/", h.getAllNotes)
		r.Post("/", h.createNote)

		r.Route("/{id}", func(r chi.Router) {
			r.Use(ValidateNoteIDMiddleware)
			r.Get("/", h.getNote)
			r.Put("/", h.updateNote)
			r.Delete("/", h.deleteNote)
		})
	})
}

// handleHealth handles the health check endpoint
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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
func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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
func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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
func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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
