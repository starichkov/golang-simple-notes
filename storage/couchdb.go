// This file contains the CouchDB implementation of the NoteStorage interface.
// It uses the Kivik library to interact with CouchDB.
package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // Import the CouchDB driver for Kivik

	"golang-simple-notes/model"
)

// CouchDBStorage implements NoteStorage using CouchDB with the Kivik library.
// CouchDB is a document-oriented NoSQL database that stores data as JSON documents.
// It provides features like document revisions, which are used to handle concurrent updates.
type CouchDBStorage struct {
	client *kivik.Client // Kivik client for connecting to CouchDB
	db     *kivik.DB     // Database handle for the notes database
}

// Document represents a CouchDB document with revision.
// This struct is used internally for mapping between the application's Note model
// and CouchDB's document format. CouchDB requires documents to have _id and _rev fields.
type Document struct {
	ID      string    `json:"_id"`            // Document ID (same as Note.ID)
	Rev     string    `json:"_rev,omitempty"` // Document revision (required by CouchDB for updates)
	Title   string    `json:"title"`          // Note title
	Content string    `json:"content"`        // Note content
	Created time.Time `json:"created_at"`     // Creation timestamp
	Updated time.Time `json:"updated_at"`     // Last update timestamp
}

// NewCouchDBStorage creates a new CouchDB storage instance.
// It connects to the CouchDB server at the specified URL, creates the database if it doesn't exist,
// and returns a CouchDBStorage instance ready to use.
//
// Parameters:
//   - url: The URL of the CouchDB server, including credentials if needed (e.g., "http://admin:password@localhost:5984")
//   - dbName: The name of the database to use for storing notes
//
// Returns:
//   - A pointer to a new CouchDBStorage instance
//   - An error if the connection or database creation fails
func NewCouchDBStorage(url, dbName string) (*CouchDBStorage, error) {
	var client *kivik.Client
	var err error

	// Try to connect to CouchDB with retries
	// This is useful when starting the application with Docker Compose,
	// as CouchDB might not be immediately available
	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create a new Kivik client for CouchDB
		client, err = kivik.New("couch", url)
		if err == nil {
			// Try to get server version as a readiness check
			// This verifies that the server is not only reachable but also ready to accept commands
			if _, err = client.Version(context.Background()); err == nil {
				break // Connection successful, exit the retry loop
			}
		}
		// Wait before retrying
		time.Sleep(2 * time.Second)
	}

	// If we still have an error after all retries, return it
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CouchDB after retries: %w", err)
	}

	// Create the database if it doesn't exist
	exists, err := client.DBExists(context.Background(), dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	if !exists {
		// Database doesn't exist, create it
		if err := client.CreateDB(context.Background(), dbName); err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Get a handle to the database
	db := client.DB(dbName)
	if db.Err() != nil {
		return nil, fmt.Errorf("failed to get database: %w", db.Err())
	}

	// Return a new CouchDBStorage instance with the database handle
	return &CouchDBStorage{
		client: client,
		db:     db,
	}, nil
}

// Create adds a new note to CouchDB.
// It uses the Kivik library's Put method to store the note as a JSON document.
// The note's ID is used as the document ID in CouchDB.
func (s *CouchDBStorage) Create(ctx context.Context, note *model.Note) error {
	// Put the note into CouchDB
	// This creates a new document with the note's ID
	_, err := s.db.Put(ctx, note.ID, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// Get retrieves a note by its ID from CouchDB.
// It returns the note if found, or ErrNoteNotFound if no note with the specified ID exists.
func (s *CouchDBStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	var note model.Note

	// Get the document from CouchDB and scan it into the note struct
	err := s.db.Get(ctx, id).ScanDoc(&note)
	if err != nil {
		// Check for "not found" errors from CouchDB
		// CouchDB returns specific error messages for missing or deleted documents
		if err.Error() == "Not Found: missing" || err.Error() == "Not Found: deleted" {
			return nil, ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return &note, nil
}

// GetAll retrieves all notes from CouchDB.
// It returns a slice of all notes in the database, which may be empty if there are no notes.
func (s *CouchDBStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	// Query all documents in the database
	// The "include_docs" parameter tells CouchDB to include the full document content
	rows := s.db.AllDocs(ctx, kivik.Param("include_docs", true))
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to get all notes: %w", rows.Err())
	}

	// Create a slice to hold the notes
	var notes []*model.Note

	// Iterate through all documents
	for rows.Next() {
		var id string
		// Get the document ID
		if err := rows.ScanKey(&id); err != nil {
			return nil, fmt.Errorf("failed to scan document key: %w", err)
		}

		// Skip design documents and other special documents
		// CouchDB uses documents with IDs starting with "_" for special purposes
		if strings.HasPrefix(id, "_design/") || strings.HasPrefix(id, "_") {
			continue
		}

		// Scan the document into a Note struct
		var note model.Note
		if err := rows.ScanDoc(&note); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		// Add the note to the slice
		notes = append(notes, &note)
	}

	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notes: %w", err)
	}

	return notes, nil
}

// Update updates an existing note in CouchDB.
// It returns ErrNoteNotFound if no note with the specified ID exists.
//
// CouchDB requires the current revision of a document to update it.
// This prevents conflicts when multiple clients try to update the same document.
func (s *CouchDBStorage) Update(ctx context.Context, note *model.Note) error {
	// First, get the current document to check if it exists and get its revision
	row := s.db.Get(ctx, note.ID)
	if row.Err() != nil {
		// If the document doesn't exist, return ErrNoteNotFound
		if row.Err().Error() == "Not Found: missing" {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to get note for update: %w", row.Err())
	}

	// Get the current revision of the document
	// CouchDB requires this for updates to prevent conflicts
	rev, err := row.Rev()
	if err != nil {
		return fmt.Errorf("failed to get revision for update: %w", err)
	}

	// Set the revision in the note
	note.Rev = rev

	// Update the document in CouchDB
	_, err = s.db.Put(ctx, note.ID, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

// Delete removes a note from CouchDB.
// It returns ErrNoteNotFound if no note with the specified ID exists.
//
// CouchDB requires the current revision of a document to delete it.
// This prevents conflicts when multiple clients try to delete the same document.
func (s *CouchDBStorage) Delete(ctx context.Context, id string) error {
	// First check if the document exists
	row := s.db.Get(ctx, id)
	if row.Err() != nil {
		// If the document doesn't exist or is already deleted, return ErrNoteNotFound
		if row.Err().Error() == "Not Found: missing" || row.Err().Error() == "Not Found: deleted" {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to get note for deletion: %w", row.Err())
	}

	// Get the current revision of the document
	// CouchDB requires this for deletion to prevent conflicts
	rev, err := row.Rev()
	if err != nil {
		return fmt.Errorf("failed to get revision for deletion: %w", err)
	}

	// Delete the document from CouchDB
	_, err = s.db.Delete(ctx, id, rev)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	// Verify the document is deleted
	// This is an extra check to ensure the deletion was successful
	row = s.db.Get(ctx, id)
	if row.Err() == nil {
		return fmt.Errorf("document still exists after deletion")
	}
	// CouchDB marks documents as deleted rather than removing them completely
	// So we expect to see a "Not Found: deleted" error
	if row.Err().Error() != "Not Found: deleted" {
		return fmt.Errorf("unexpected error after deletion: %w", row.Err())
	}

	return nil
}

// Close closes the CouchDB connection.
// For the CouchDB implementation, there are no resources to close,
// as the Kivik library doesn't require explicit closing.
func (s *CouchDBStorage) Close(context.Context) error {
	// CouchDB client doesn't need explicit closing
	return nil
}
