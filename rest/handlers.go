// Package rest implements the REST API handlers for the Notes API.
// It provides HTTP handlers for creating, reading, updating, and deleting notes,
// as well as middleware for request validation.
package rest

import (
	"encoding/json"
	"golang-simple-notes/model"
	"golang-simple-notes/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for notes.
// It uses a storage implementation to persist and retrieve notes.
// This follows the dependency injection pattern, allowing the handler
// to work with any storage implementation that satisfies the NoteStorage interface.
type Handler struct {
	storage storage.NoteStorage // Storage backend for notes
}

// NewHandler creates a new Handler instance with the provided storage.
// This follows the factory pattern for creating handlers.
//
// Parameters:
//   - storage: An implementation of the NoteStorage interface
//
// Returns:
//   - A pointer to a new Handler instance
func NewHandler(storage storage.NoteStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// RegisterRoutes registers the handler's routes with the provided router.
// This sets up all the API endpoints for the Notes API.
//
// The routes are:
//   - GET /health - Health check endpoint
//   - GET /api/notes - Get all notes
//   - POST /api/notes - Create a new note
//   - GET /api/notes/{id} - Get a note by ID
//   - PUT /api/notes/{id} - Update a note
//   - DELETE /api/notes/{id} - Delete a note
//
// The {id} routes use the ValidateNoteIDMiddleware to ensure the ID is valid.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Health check endpoint
	r.Get("/health", h.handleHealth)

	// Group all note-related routes under /api/notes
	r.Route("/api/notes", func(r chi.Router) {
		// Routes for operations on all notes
		r.Get("/", h.getAllNotes) // Get all notes
		r.Post("/", h.createNote) // Create a new note

		// Routes for operations on a specific note
		r.Route("/{id}", func(r chi.Router) {
			// Add middleware to validate the note ID
			r.Use(ValidateNoteIDMiddleware)
			r.Get("/", h.getNote)       // Get a note by ID
			r.Put("/", h.updateNote)    // Update a note
			r.Delete("/", h.deleteNote) // Delete a note
		})
	})
}

// handleHealth handles the health check endpoint (GET /health).
// It returns a simple "OK" response with a 200 status code to indicate that the service is running.
// This endpoint can be used by load balancers or monitoring tools to check if the service is healthy.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // Set the status code to 200 OK
	w.Write([]byte("OK"))        // Write a simple "OK" response
}

// getAllNotes handles GET /api/notes.
// It retrieves all notes from the storage and returns them as a JSON array.
// If there are no notes, it returns an empty array.
func (h *Handler) getAllNotes(w http.ResponseWriter, r *http.Request) {
	// Get all notes from the storage
	notes, err := h.storage.GetAll(r.Context())
	if err != nil {
		// If there's an error, return a 500 Internal Server Error
		http.Error(w, "Failed to get notes", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode the notes as JSON and write to the response
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		// If encoding fails, return a 500 Internal Server Error
		http.Error(w, "Failed to encode notes", http.StatusInternalServerError)
		return
	}
}

// getNote handles GET /api/notes/{id}.
// It retrieves a note by its ID from the storage and returns it as JSON.
// If the note doesn't exist, it returns a 404 Not Found.
func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	// Get the note ID from the URL path parameter
	id := chi.URLParam(r, "id")

	// Get the note from the storage
	note, err := h.storage.Get(r.Context(), id)
	if err != nil {
		// Handle specific error cases
		if err == storage.ErrNoteNotFound {
			// If the note doesn't exist, return a 404 Not Found
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		// For any other error, return a 500 Internal Server Error
		http.Error(w, "Failed to get note", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode the note as JSON and write to the response
	if err := json.NewEncoder(w).Encode(note); err != nil {
		// If encoding fails, return a 500 Internal Server Error
		http.Error(w, "Failed to encode note", http.StatusInternalServerError)
		return
	}
}

// createNote handles POST /api/notes.
// It creates a new note from the request body and returns the created note as JSON.
// The note ID is generated automatically.
func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var note model.Note

	// Decode the request body into a Note struct
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		// If decoding fails, return a 400 Bad Request
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create the note in the storage
	if err := h.storage.Create(r.Context(), &note); err != nil {
		// If creation fails, return a 500 Internal Server Error
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the status code to 201 Created
	w.WriteHeader(http.StatusCreated)

	// Encode the created note as JSON and write to the response
	if err := json.NewEncoder(w).Encode(note); err != nil {
		// If encoding fails, return a 500 Internal Server Error
		http.Error(w, "Failed to encode note", http.StatusInternalServerError)
		return
	}
}

// updateNote handles PUT /api/notes/{id}.
// It updates an existing note with the data from the request body and returns the updated note as JSON.
// If the note doesn't exist, it returns a 404 Not Found.
func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	// Get the note ID from the URL path parameter
	id := chi.URLParam(r, "id")

	var note model.Note

	// Decode the request body into a Note struct
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		// If decoding fails, return a 400 Bad Request
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set the note ID to the one from the URL path
	// This ensures the correct note is updated, regardless of any ID in the request body
	note.ID = id

	// Update the note in the storage
	if err := h.storage.Update(r.Context(), &note); err != nil {
		// Handle specific error cases
		if err == storage.ErrNoteNotFound {
			// If the note doesn't exist, return a 404 Not Found
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		// For any other error, return a 500 Internal Server Error
		http.Error(w, "Failed to update note", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode the updated note as JSON and write to the response
	if err := json.NewEncoder(w).Encode(note); err != nil {
		// If encoding fails, return a 500 Internal Server Error
		http.Error(w, "Failed to encode note", http.StatusInternalServerError)
		return
	}
}

// deleteNote handles DELETE /api/notes/{id}.
// It deletes a note by its ID from the storage.
// If the note doesn't exist, it returns a 404 Not Found.
// If the deletion is successful, it returns a 204 No Content.
func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	// Get the note ID from the URL path parameter
	id := chi.URLParam(r, "id")

	// Delete the note from the storage
	if err := h.storage.Delete(r.Context(), id); err != nil {
		// Handle specific error cases
		if err == storage.ErrNoteNotFound {
			// If the note doesn't exist, return a 404 Not Found
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		// For any other error, return a 500 Internal Server Error
		http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	// Set the status code to 204 No Content
	// This indicates that the request was successful but there's no content to return
	w.WriteHeader(http.StatusNoContent)
}
