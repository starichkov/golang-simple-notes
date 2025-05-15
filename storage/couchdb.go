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
	ctx    context.Context
	cancel context.CancelFunc
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

// NewCouchDBStorage creates a new instance of CouchDBStorage using Kivik
func NewCouchDBStorage(url, dbName string) (*CouchDBStorage, error) {
	// Create a context with timeout for database operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Connect to CouchDB using Kivik
	client, err := kivik.New("couch", url)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to CouchDB: %w", err)
	}

	// Check if database exists
	exists, err := client.DBExists(ctx, dbName)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		if err := client.CreateDB(ctx, dbName); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Get a reference to the database
	db := client.DB(dbName)

	return &CouchDBStorage{
		client: client,
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close closes any resources used by the storage
func (s *CouchDBStorage) Close() error {
	s.cancel()
	return s.client.Close()
}

// Create adds a new note to the storage
func (s *CouchDBStorage) Create(note *model.Note) error {
	doc := Document{
		ID:      note.ID,
		Title:   note.Title,
		Content: note.Content,
		Created: note.CreatedAt,
		Updated: note.UpdatedAt,
	}

	_, err := s.db.Put(s.ctx, note.ID, doc)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

// Get retrieves a note by its ID
func (s *CouchDBStorage) Get(id string) (*model.Note, error) {
	row := s.db.Get(s.ctx, id)
	if err := row.Err(); err != nil {
		// Check if the error is a "not found" error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") {
			return nil, ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	var doc Document
	if err := row.ScanDoc(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	note := &model.Note{
		ID:        doc.ID,
		Title:     doc.Title,
		Content:   doc.Content,
		CreatedAt: doc.Created,
		UpdatedAt: doc.Updated,
	}

	return note, nil
}

// GetAll retrieves all notes from the storage
func (s *CouchDBStorage) GetAll() ([]*model.Note, error) {
	// Use AllDocs to get all documents
	rows := s.db.AllDocs(s.ctx)
	defer rows.Close()

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

		// Get the document by ID
		note, err := s.Get(id)
		if err != nil {
			if err == ErrNoteNotFound {
				continue // Skip deleted documents
			}
			return nil, fmt.Errorf("failed to get note: %w", err)
		}

		notes = append(notes, note)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating through documents: %w", rows.Err())
	}

	return notes, nil
}

// Update updates an existing note
func (s *CouchDBStorage) Update(note *model.Note) error {
	// Get the current document to get the _rev field
	row := s.db.Get(s.ctx, note.ID)
	if err := row.Err(); err != nil {
		// Check if the error is a "not found" error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to get note: %w", err)
	}

	var existingDoc Document
	if err := row.ScanDoc(&existingDoc); err != nil {
		return fmt.Errorf("failed to decode document: %w", err)
	}

	// Update the note with the current time
	note.UpdatedAt = time.Now()

	// Create a new document with the _rev field from the existing document
	doc := Document{
		ID:      note.ID,
		Rev:     existingDoc.Rev,
		Title:   note.Title,
		Content: note.Content,
		Created: note.CreatedAt,
		Updated: note.UpdatedAt,
	}

	_, err := s.db.Put(s.ctx, note.ID, doc)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

// Delete removes a note from the storage
func (s *CouchDBStorage) Delete(id string) error {
	// Get the current document to get the _rev field
	row := s.db.Get(s.ctx, id)
	if err := row.Err(); err != nil {
		// Check if the error is a "not found" error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to get note: %w", err)
	}

	var existingDoc Document
	if err := row.ScanDoc(&existingDoc); err != nil {
		return fmt.Errorf("failed to decode document: %w", err)
	}

	// Delete the document
	_, err := s.db.Delete(s.ctx, id, existingDoc.Rev)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
