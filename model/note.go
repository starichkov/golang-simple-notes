package model

import (
	"time"
)

// Note represents a single note in the system
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewNote creates a new note with the given title and content
func NewNote(title, content string) *Note {
	now := time.Now()
	return &Note{
		ID:        generateID(),
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// generateID creates a simple unique ID for a note
// In a production environment, you might want to use UUID or another robust ID generation method
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
