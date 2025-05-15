package rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang-simple-notes/model"
	"golang-simple-notes/storage"
)

// Handler handles HTTP requests for the notes API
type Handler struct {
	storage storage.NoteStorage
}

// NewHandler creates a new instance of Handler
func NewHandler(storage storage.NoteStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// RegisterRoutes registers all the routes for the REST API
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/notes", h.GetAllNotes)
	mux.HandleFunc("GET /api/notes/", h.GetNote)
	mux.HandleFunc("POST /api/notes", h.CreateNote)
	mux.HandleFunc("PUT /api/notes/", h.UpdateNote)
	mux.HandleFunc("DELETE /api/notes/", h.DeleteNote)
}

// CreateNote handles the creation of a new note
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var noteRequest struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&noteRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if noteRequest.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	note := model.NewNote(noteRequest.Title, noteRequest.Content)
	if err := h.storage.Create(note); err != nil {
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// GetAllNotes handles retrieving all notes
func (h *Handler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.storage.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// GetNote handles retrieving a single note by ID
func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if id == "" {
		http.Error(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	note, err := h.storage.Get(id)
	if err != nil {
		if err == storage.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve note", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

// UpdateNote handles updating an existing note
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if id == "" {
		http.Error(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	existingNote, err := h.storage.Get(id)
	if err != nil {
		if err == storage.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve note", http.StatusInternalServerError)
		}
		return
	}

	var noteRequest struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&noteRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if noteRequest.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	existingNote.Title = noteRequest.Title
	existingNote.Content = noteRequest.Content
	existingNote.UpdatedAt = time.Now()

	if err := h.storage.Update(existingNote); err != nil {
		http.Error(w, "Failed to update note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingNote)
}

// DeleteNote handles deleting a note
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if id == "" {
		http.Error(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	if err := h.storage.Delete(id); err != nil {
		if err == storage.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
