package storage

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"golang-simple-notes/model"
)

// TestInMemoryStorage tests the in-memory storage implementation
func TestInMemoryStorage(t *testing.T) {
	// Create a new in-memory storage
	storage := NewInMemoryStorage()

	// Run the fixed storage tests
	testNoteStorage(t, storage)
}

// TestInMemoryStorageConcurrency tests the thread safety of the in-memory storage
func TestInMemoryStorageConcurrency(t *testing.T) {
	storage := NewInMemoryStorage()

	// Create a note to work with
	note := model.NewNote("Concurrency Test", "Testing concurrent access")
	err := storage.Create(note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Number of concurrent operations
	concurrentOps := 100

	// Test concurrent reads
	t.Run("ConcurrentReads", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < concurrentOps; i++ {
			go func() {
				_, err := storage.Get(note.ID)
				if err != nil {
					t.Errorf("Failed to get note: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrentOps; i++ {
			<-done
		}
	})

	// Test concurrent writes
	t.Run("ConcurrentWrites", func(t *testing.T) {
		// Clean up storage before test
		notes, _ := storage.GetAll()
		for _, n := range notes {
			storage.Delete(n.ID)
		}

		// Create the initial note again
		note = model.NewNote("Concurrency Test", "Testing concurrent access")
		err = storage.Create(note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Use a channel to collect created note IDs
		noteIDs := make(chan string, concurrentOps)
		done := make(chan bool, concurrentOps)

		for i := 0; i < concurrentOps; i++ {
			go func(i int) {
				// Create a unique note for each goroutine with a guaranteed unique ID
				uniqueID := fmt.Sprintf("concurrent-note-%d", i)
				uniqueNote := &model.Note{
					ID:        uniqueID,
					Title:     fmt.Sprintf("Concurrent Note %d", i),
					Content:   fmt.Sprintf("Created by goroutine %d", i),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				err := storage.Create(uniqueNote)
				if err != nil {
					t.Errorf("Failed to create note: %v", err)
				} else {
					noteIDs <- uniqueID
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrentOps; i++ {
			<-done
		}

		// Close the noteIDs channel
		close(noteIDs)

		// Collect all created note IDs
		createdIDs := make(map[string]bool)
		for id := range noteIDs {
			createdIDs[id] = true
		}

		// Verify that all notes were created
		notes, err := storage.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all notes: %v", err)
		}

		// We expect concurrentOps + 1 notes (the original note + the new ones)
		expectedCount := len(createdIDs) + 1 // +1 for the original note
		if len(notes) != expectedCount {
			t.Errorf("Expected %d notes, got %d", expectedCount, len(notes))
		}

		// Verify that each created note can be retrieved
		for id := range createdIDs {
			_, err := storage.Get(id)
			if err != nil {
				t.Errorf("Failed to get created note with ID %s: %v", id, err)
			}
		}
	})

	// Test concurrent updates
	t.Run("ConcurrentUpdates", func(t *testing.T) {
		// Create a note to update
		updateNote := model.NewNote("Update Test", "Testing concurrent updates")
		err := storage.Create(updateNote)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		// Use a mutex to protect access to the note ID
		var noteMutex sync.Mutex
		noteID := updateNote.ID

		done := make(chan bool)
		for i := 0; i < concurrentOps; i++ {
			go func(i int) {
				// Create a new note object for each update to avoid race conditions
				// All notes have the same ID but different content
				updatedContent := fmt.Sprintf("Updated content %d", i)

				noteMutex.Lock()
				currentID := noteID
				noteMutex.Unlock()

				updatedNote := &model.Note{
					ID:        currentID,
					Title:     "Update Test",
					Content:   updatedContent,
					CreatedAt: updateNote.CreatedAt,
					UpdatedAt: time.Now(),
				}

				err := storage.Update(updatedNote)
				if err != nil {
					t.Errorf("Failed to update note: %v", err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrentOps; i++ {
			<-done
		}

		// Verify that the note was updated (we don't check specific content
		// since we don't know which goroutine's update was the last one)
		retrieved, err := storage.Get(updateNote.ID)
		if err != nil {
			t.Fatalf("Failed to get updated note: %v", err)
		}

		if !strings.HasPrefix(retrieved.Content, "Updated content ") {
			t.Errorf("Expected content to start with 'Updated content ', got '%s'", retrieved.Content)
		}
	})

	// Test concurrent deletes
	t.Run("ConcurrentDeletes", func(t *testing.T) {
		// Clean up storage before test
		notes, _ := storage.GetAll()
		for _, n := range notes {
			storage.Delete(n.ID)
		}

		// Create notes to delete with guaranteed unique IDs
		deleteNotes := make([]*model.Note, concurrentOps)
		for i := 0; i < concurrentOps; i++ {
			uniqueID := fmt.Sprintf("delete-note-%d", i)
			deleteNotes[i] = &model.Note{
				ID:        uniqueID,
				Title:     fmt.Sprintf("Delete Test %d", i),
				Content:   fmt.Sprintf("Testing concurrent deletes %d", i),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := storage.Create(deleteNotes[i])
			if err != nil {
				t.Fatalf("Failed to create note: %v", err)
			}
		}

		// Use a mutex to protect access to the deleted IDs map
		var deletedMutex sync.Mutex
		deletedIDs := make(map[string]bool)

		done := make(chan bool)
		for i := 0; i < concurrentOps; i++ {
			go func(i int) {
				// Delete the note
				noteID := deleteNotes[i].ID
				err := storage.Delete(noteID)
				if err != nil {
					t.Errorf("Failed to delete note %s: %v", noteID, err)
				} else {
					// Record successful deletion
					deletedMutex.Lock()
					deletedIDs[noteID] = true
					deletedMutex.Unlock()
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < concurrentOps; i++ {
			<-done
		}

		// Verify that all notes were deleted
		for i := 0; i < concurrentOps; i++ {
			noteID := deleteNotes[i].ID

			// Only check notes that were successfully deleted
			deletedMutex.Lock()
			wasDeleted := deletedIDs[noteID]
			deletedMutex.Unlock()

			if wasDeleted {
				_, err := storage.Get(noteID)
				if err != ErrNoteNotFound {
					t.Errorf("Expected ErrNoteNotFound for note %s, got %v", noteID, err)
				}
			}
		}
	})
}

// TestInMemoryStorageEdgeCases tests edge cases for the in-memory storage
func TestInMemoryStorageEdgeCases(t *testing.T) {
	storage := NewInMemoryStorage()

	// Test creating a note with an empty ID
	t.Run("CreateEmptyID", func(t *testing.T) {
		note := &model.Note{
			ID:        "", // Empty ID
			Title:     "Empty ID",
			Content:   "Note with empty ID",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := storage.Create(note)
		if err != nil {
			t.Errorf("Failed to create note with empty ID: %v", err)
		}

		// Try to retrieve the note
		_, err = storage.Get("")
		if err != nil {
			t.Errorf("Failed to get note with empty ID: %v", err)
		}
	})

	// Test updating a note with an empty ID
	t.Run("UpdateEmptyID", func(t *testing.T) {
		note := &model.Note{
			ID:        "", // Empty ID
			Title:     "Updated Empty ID",
			Content:   "Updated note with empty ID",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := storage.Update(note)
		if err != nil {
			t.Errorf("Failed to update note with empty ID: %v", err)
		}

		// Retrieve the note to verify the update
		retrieved, err := storage.Get("")
		if err != nil {
			t.Errorf("Failed to get updated note with empty ID: %v", err)
		}

		if retrieved.Title != "Updated Empty ID" {
			t.Errorf("Expected title 'Updated Empty ID', got '%s'", retrieved.Title)
		}
	})

	// Test deleting a note with an empty ID
	t.Run("DeleteEmptyID", func(t *testing.T) {
		err := storage.Delete("")
		if err != nil {
			t.Errorf("Failed to delete note with empty ID: %v", err)
		}

		// Try to retrieve the deleted note
		_, err = storage.Get("")
		if err != ErrNoteNotFound {
			t.Errorf("Expected ErrNoteNotFound, got %v", err)
		}
	})

	// Test creating a note with nil fields
	t.Run("CreateNilFields", func(t *testing.T) {
		note := &model.Note{
			ID:      "nil-fields",
			Title:   "", // Empty title
			Content: "", // Empty content
			// CreatedAt and UpdatedAt are zero values
		}

		err := storage.Create(note)
		if err != nil {
			t.Errorf("Failed to create note with nil fields: %v", err)
		}

		// Try to retrieve the note
		retrieved, err := storage.Get("nil-fields")
		if err != nil {
			t.Errorf("Failed to get note with nil fields: %v", err)
		}

		if retrieved.Title != "" {
			t.Errorf("Expected empty title, got '%s'", retrieved.Title)
		}

		if retrieved.Content != "" {
			t.Errorf("Expected empty content, got '%s'", retrieved.Content)
		}
	})
}
