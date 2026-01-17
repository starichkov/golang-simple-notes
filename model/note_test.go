package model

import (
	"strings"
	"testing"
	"time"
)

func TestNewNote(t *testing.T) {
	title := "Test Title"
	content := "Test Content"

	note := NewNote(title, content)

	if note == nil {
		t.Fatal("Expected note to be created, got nil")
	}

	if note.Title != title {
		t.Errorf("Expected title to be %q, got %q", title, note.Title)
	}

	if note.Content != content {
		t.Errorf("Expected content to be %q, got %q", content, note.Content)
	}

	if note.ID == "" {
		t.Error("Expected ID to be generated, got empty string")
	}

	// Check that timestamps are set and close to now
	now := time.Now()
	if note.CreatedAt.After(now) || note.CreatedAt.Before(now.Add(-time.Second)) {
		t.Errorf("Expected CreatedAt to be close to now, got %v", note.CreatedAt)
	}

	if note.UpdatedAt.After(now) || note.UpdatedAt.Before(now.Add(-time.Second)) {
		t.Errorf("Expected UpdatedAt to be close to now, got %v", note.UpdatedAt)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()

	if id1 == "" {
		t.Error("Expected ID to be generated, got empty string")
	}

	// Sleep a tiny bit to ensure different timestamps
	time.Sleep(time.Millisecond)

	id2 := generateID()

	if id1 == id2 {
		t.Errorf("Expected different IDs, got the same ID twice: %s", id1)
	}

	// Check format (should be a timestamp in the format "20060102150405.000000" followed by a random suffix)
	// Example: "20230415123045.123456.a1b2c3d4"
	parts := strings.Split(id1, ".")
	if len(parts) != 3 {
		t.Fatalf("ID %s is not in the expected format: expected 3 parts, got %d", id1, len(parts))
	}

	timestampStr := parts[0] + "." + parts[1]
	_, err := time.Parse("20060102150405.000000", timestampStr)
	if err != nil {
		t.Errorf("ID %s timestamp part is not in the expected format: %v", id1, err)
	}

	if len(parts[2]) != 8 {
		t.Errorf("ID %s random part should be 8 characters, got %d", id1, len(parts[2]))
	}
}
