package stats

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"bytes"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDB(t *testing.T) func() {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "bookshelf-stats-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	var dbErr error
	db.DB, dbErr = sql.Open("sqlite", dbPath)
	if dbErr != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open database: %v", dbErr)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		isbn TEXT,
		pages INTEGER,
		cover_url TEXT,
		description TEXT,
		open_library_key TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS reading_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'want-to-read',
		started_at DATETIME,
		finished_at DATETIME,
		rating INTEGER CHECK(rating >= 1 AND rating <= 5),
		review TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
	);
	`
	if _, err := db.DB.Exec(schema); err != nil {
		db.DB.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	return func() {
		db.DB.Close()
		os.RemoveAll(tmpDir)
	}
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintStatsEmpty(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	output := captureOutput(func() {
		err := PrintStats()
		if err != nil {
			t.Fatalf("PrintStats failed: %v", err)
		}
	})

	expectedStrings := []string{
		"Reading Statistics",
		"Total books:    0",
		"Want to read:   0",
		"Reading:        0",
		"Finished:       0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("output missing: %s", expected)
		}
	}
}

func TestPrintStatsWithBooks(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Add books
	id1, _ := db.AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := db.AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id2, models.StatusReading)

	id3, _ := db.AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id3, models.StatusFinished)
	db.UpdateRating(id3, 4)

	output := captureOutput(func() {
		err := PrintStats()
		if err != nil {
			t.Fatalf("PrintStats failed: %v", err)
		}
	})

	expectedStrings := []string{
		"Total books:    3",
		"Want to read:   1",
		"Reading:        1",
		"Finished:       1",
		"Average rating: 4.0/5",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("output missing: %s", expected)
		}
	}
}

func TestRenderStars(t *testing.T) {
	tests := []struct {
		rating   float64
		expected string
	}{
		{0.0, "....."},
		{1.0, "*...."},
		{2.5, "**~.."},
		{3.0, "***.."},
		{4.5, "****~"},
		{5.0, "*****"},
	}

	for _, tt := range tests {
		result := renderStars(tt.rating)
		if result != tt.expected {
			t.Errorf("renderStars(%.1f) = %s, expected %s", tt.rating, result, tt.expected)
		}
	}
}
