package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // CouchDB driver

	"golang-simple-notes/model"
)

// CouchDBStorage implements NoteStorage using CouchDB with Kivik
type CouchDBStorage struct {
	client *kivik.Client
	db     *kivik.DB
}

// Document represents a CouchDB document with revision
type Document struct {
	ID      string    `json:"_id"`
	Rev     string    `json:"_rev,omitempty"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
}

// NewCouchDBStorage creates a new CouchDB storage instance
func NewCouchDBStorage(url, dbName string) (*CouchDBStorage, error) {
	var client *kivik.Client
	var err error
	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		client, err = kivik.New("couch", url)
		if err == nil {
			// Try to get server version as a readiness check
			if _, err = client.Version(context.Background()); err == nil {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CouchDB after retries: %w", err)
	}

	// Create the database if it doesn't exist
	exists, err := client.DBExists(context.Background(), dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	if !exists {
		if err := client.CreateDB(context.Background(), dbName); err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	db := client.DB(dbName)
	if db.Err() != nil {
		return nil, fmt.Errorf("failed to get database: %w", db.Err())
	}

	return &CouchDBStorage{
		db: db,
	}, nil
}

// Create adds a new note to CouchDB
func (s *CouchDBStorage) Create(ctx context.Context, note *model.Note) error {
	_, err := s.db.Put(ctx, note.ID, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// Get retrieves a note by its ID from CouchDB
func (s *CouchDBStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	var note model.Note
	err := s.db.Get(ctx, id).ScanDoc(&note)
	if err != nil {
		if err.Error() == "Not Found: missing" || err.Error() == "Not Found: deleted" {
			return nil, ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	return &note, nil
}

// GetAll retrieves all notes from CouchDB
func (s *CouchDBStorage) GetAll(ctx context.Context) ([]*model.Note, error) {
	rows := s.db.AllDocs(ctx, kivik.Param("include_docs", true))
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to get all notes: %w", rows.Err())
	}

	var notes []*model.Note
	for rows.Next() {
		var id string
		if err := rows.ScanKey(&id); err != nil {
			return nil, fmt.Errorf("failed to scan document key: %w", err)
		}

		// Skip design documents
		if strings.HasPrefix(id, "_design/") || strings.HasPrefix(id, "_") {
			continue
		}

		var note model.Note
		if err := rows.ScanDoc(&note); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, &note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notes: %w", err)
	}

	return notes, nil
}

// Update updates an existing note in CouchDB
func (s *CouchDBStorage) Update(ctx context.Context, note *model.Note) error {
	row := s.db.Get(ctx, note.ID)
	if row.Err() != nil {
		if row.Err().Error() == "Not Found: missing" {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to get note for update: %w", row.Err())
	}

	rev, err := row.Rev()
	if err != nil {
		return fmt.Errorf("failed to get revision for update: %w", err)
	}

	note.Rev = rev
	_, err = s.db.Put(ctx, note.ID, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}
	return nil
}

// Delete removes a note from CouchDB
func (s *CouchDBStorage) Delete(ctx context.Context, id string) error {
	// First check if the document exists
	row := s.db.Get(ctx, id)
	if row.Err() != nil {
		if row.Err().Error() == "Not Found: missing" || row.Err().Error() == "Not Found: deleted" {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to get note for deletion: %w", row.Err())
	}

	// Get the revision
	rev, err := row.Rev()
	if err != nil {
		return fmt.Errorf("failed to get revision for deletion: %w", err)
	}

	// Delete the document
	_, err = s.db.Delete(ctx, id, rev)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	// Verify the document is deleted
	row = s.db.Get(ctx, id)
	if row.Err() == nil {
		return fmt.Errorf("document still exists after deletion")
	}
	if row.Err().Error() != "Not Found: deleted" {
		return fmt.Errorf("unexpected error after deletion: %w", row.Err())
	}

	return nil
}

// Close closes the CouchDB connection
func (s *CouchDBStorage) Close(context.Context) error {
	// CouchDB client doesn't need explicit closing
	return nil
}
