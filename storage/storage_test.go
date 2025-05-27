package storage

import (
	"context"
	"testing"
	"time"

	"golang-simple-notes/model"
)

// testNoteStorage is a helper function that tests any implementation of NoteStorage
func testNoteStorage(t *testing.T, storage NoteStorage, ctx context.Context) {
	// Test Create and Get
	t.Run("Create and Get", func(t *testing.T) {
		// Clean up any existing notes
		cleanupStorage(t, storage, ctx)

		note := model.NewNote("Test Title", "Test Content")

		err := storage.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		retrieved, err := storage.Get(ctx, note.ID)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if retrieved.ID != note.ID {
			t.Errorf("Expected ID %s, got %s", note.ID, retrieved.ID)
		}

		if retrieved.Title != note.Title {
			t.Errorf("Expected title %s, got %s", note.Title, retrieved.Title)
		}

		if retrieved.Content != note.Content {
			t.Errorf("Expected content %s, got %s", note.Content, retrieved.Content)
		}
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		// Clean up any existing notes
		cleanupStorage(t, storage, ctx)

		// Create some notes
		note1 := model.NewNote("Title 1", "Content 1")
		note2 := model.NewNote("Title 2", "Content 2")

		err := storage.Create(ctx, note1)
		if err != nil {
			t.Fatalf("Failed to create note1: %v", err)
		}

		err = storage.Create(ctx, note2)
		if err != nil {
			t.Fatalf("Failed to create note2: %v", err)
		}

		notes, err := storage.GetAll(ctx)
		if err != nil {
			t.Fatalf("Failed to get all notes: %v", err)
		}

		if len(notes) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(notes))
		}

		// Check that both notes are in the result
		found1, found2 := false, false
		for _, n := range notes {
			if n.ID == note1.ID {
				found1 = true
			}
			if n.ID == note2.ID {
				found2 = true
			}
		}

		if !found1 {
			t.Error("Note 1 not found in GetAll results")
		}

		if !found2 {
			t.Error("Note 2 not found in GetAll results")
		}
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		// Clean up any existing notes
		cleanupStorage(t, storage, ctx)

		note := model.NewNote("Original Title", "Original Content")

		err := storage.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Update the note
		note.Title = "Updated Title"
		note.Content = "Updated Content"
		note.UpdatedAt = time.Now()

		err = storage.Update(ctx, note)
		if err != nil {
			t.Fatalf("Failed to update note: %v", err)
		}

		// Retrieve the updated note
		retrieved, err := storage.Get(ctx, note.ID)
		if err != nil {
			t.Fatalf("Failed to get updated note: %v", err)
		}

		if retrieved.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", retrieved.Title)
		}

		if retrieved.Content != "Updated Content" {
			t.Errorf("Expected content 'Updated Content', got '%s'", retrieved.Content)
		}
	})

	// Test Update with non-existent note
	t.Run("Update Non-existent", func(t *testing.T) {
		note := model.NewNote("Non-existent", "This note doesn't exist in storage")

		err := storage.Update(ctx, note)
		if err != ErrNoteNotFound {
			t.Errorf("Expected ErrNoteNotFound, got %v", err)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		// Clean up any existing notes
		cleanupStorage(t, storage, ctx)

		note := model.NewNote("To Delete", "This note will be deleted")

		err := storage.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Delete the note
		err = storage.Delete(ctx, note.ID)
		if err != nil {
			t.Fatalf("Failed to delete note: %v", err)
		}

		// Try to retrieve the deleted note
		_, err = storage.Get(ctx, note.ID)
		if err != ErrNoteNotFound {
			t.Errorf("Expected ErrNoteNotFound, got %v", err)
		}
	})

	// Test Delete with non-existent note
	t.Run("Delete Non-existent", func(t *testing.T) {
		err := storage.Delete(ctx, "non-existent-id")
		if err != ErrNoteNotFound {
			t.Errorf("Expected ErrNoteNotFound, got %v", err)
		}
	})

	// Test Close
	t.Run("Close", func(t *testing.T) {
		err := storage.Close(ctx)
		if err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	})
}

// cleanupStorage is a helper function to clean up any existing notes in the storage
func cleanupStorage(t *testing.T, storage NoteStorage, ctx context.Context) {
	notes, err := storage.GetAll(ctx)
	if err != nil {
		t.Logf("Warning: Failed to get all notes for cleanup: %v", err)
		return
	}

	for _, note := range notes {
		err := storage.Delete(ctx, note.ID)
		if err != nil {
			t.Logf("Warning: Failed to delete note %s during cleanup: %v", note.ID, err)
		}
	}
}
