package rest

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ValidateNoteIDMiddleware is a middleware that validates the note ID in the request URL
func ValidateNoteIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Note ID is required", http.StatusBadRequest)
			return
		}
		if !isValidNoteID(id) {
			http.Error(w, "Invalid note ID format", http.StatusBadRequest)
			return
		}

		// ID is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
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
