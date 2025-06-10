// Package model contains the data models used throughout the application.
// This package defines the structure of the data that is stored and manipulated.
package model

import (
	"time"
)

// Note represents a single note in the system.
// It contains all the information about a note, including its content and metadata.
// The struct tags (`json:"..."` and `bson:"..."`) are used for JSON serialization
// and MongoDB document mapping, respectively.
type Note struct {
	ID        string    `json:"_id" bson:"_id"`                       // Unique identifier for the note
	Rev       string    `json:"_rev,omitempty" bson:"_rev,omitempty"` // Revision ID (used by CouchDB)
	Title     string    `json:"title" bson:"title"`                   // Title of the note
	Content   string    `json:"content" bson:"content"`               // Content/body of the note
	CreatedAt time.Time `json:"created_at" bson:"created_at"`         // When the note was created
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`         // When the note was last updated
}

// NewNote creates a new note with the given title and content.
// It automatically generates a unique ID and sets the creation and update timestamps.
// Example usage:
//
//	note := model.NewNote("Shopping List", "Milk, Eggs, Bread")
func NewNote(title, content string) *Note {
	now := time.Now() // Get the current time for timestamps
	return &Note{
		ID:        generateID(), // Generate a unique ID
		Title:     title,
		Content:   content,
		CreatedAt: now, // Set creation time
		UpdatedAt: now, // Set initial update time (same as creation time)
	}
}

// generateID creates a simple unique ID for a note based on the current timestamp.
// The format used (year, month, day, hour, minute, second, microsecond) ensures
// uniqueness as long as two notes aren't created in the exact same microsecond.
//
// In a production environment, you might want to use UUID or another robust ID
// generation method to ensure global uniqueness across distributed systems.
func generateID() string {
	// Format the current time as a string in the format "YYYYMMDDhhmmss.microseconds"
	// For example: "20230415123045.123456"
	return time.Now().Format("20060102150405.000000")
}
