package publish

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDB(t *testing.T) func() {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "bookshelf-publish-test-*")
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

	// Run migrations manually
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

func TestGenerateEmptySite(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	outputDir, err := os.MkdirTemp("", "bookshelf-output-*")
	if err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	err = Generate(outputDir)
	if err != nil {
		t.Fatalf("failed to generate site: %v", err)
	}

	// Check files exist
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("index.html not created")
	}

	cssPath := filepath.Join(outputDir, "style.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Error("style.css not created")
	}

	booksDir := filepath.Join(outputDir, "books")
	if _, err := os.Stat(booksDir); os.IsNotExist(err) {
		t.Error("books directory not created")
	}
}

func TestGenerateWithBooks(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Add a book
	id, _ := db.AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id, models.StatusFinished)
	db.UpdateRating(id, 5)
	db.UpdateReview(id, "Great book!")

	outputDir, err := os.MkdirTemp("", "bookshelf-output-*")
	if err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	err = Generate(outputDir)
	if err != nil {
		t.Fatalf("failed to generate site: %v", err)
	}

	// Check book page exists
	bookPath := filepath.Join(outputDir, "books", "1.html")
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Error("book page not created")
	}

	// Check index contains book
	indexContent, _ := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if !strings.Contains(string(indexContent), "Test Book") {
		t.Error("index.html does not contain book title")
	}

	// Check book page contains review
	bookContent, _ := os.ReadFile(bookPath)
	if !strings.Contains(string(bookContent), "Great book!") {
		t.Error("book page does not contain review")
	}
}

func TestTemplateFuncsStatusClass(t *testing.T) {
	funcs := templateFuncs()
	statusClassFn := funcs["statusClass"].(func(models.BookStatus) string)

	tests := []struct {
		status   models.BookStatus
		expected string
	}{
		{models.StatusWantToRead, "wanttoread"},
		{models.StatusReading, "reading"},
		{models.StatusFinished, "finished"},
	}

	for _, tt := range tests {
		result := statusClassFn(tt.status)
		if result != tt.expected {
			t.Errorf("statusClass(%s) = %s, expected %s", tt.status, result, tt.expected)
		}
	}
}

func TestTemplateFuncsStars(t *testing.T) {
	funcs := templateFuncs()
	starsFn := funcs["stars"].(func(int64) string)

	tests := []struct {
		rating   int64
		expected string
	}{
		{0, "☆☆☆☆☆"},
		{1, "★☆☆☆☆"},
		{3, "★★★☆☆"},
		{5, "★★★★★"},
	}

	for _, tt := range tests {
		result := starsFn(tt.rating)
		if result != tt.expected {
			t.Errorf("stars(%d) = %s, expected %s", tt.rating, result, tt.expected)
		}
	}
}

func TestTemplateFuncsTruncate(t *testing.T) {
	funcs := templateFuncs()
	truncateFn := funcs["truncate"].(func(string, int) string)

	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exactly10!", 10, "exactly10!"},
	}

	for _, tt := range tests {
		result := truncateFn(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncate(%s, %d) = %s, expected %s", tt.input, tt.max, result, tt.expected)
		}
	}
}

func TestGenerateInvalidOutputDir(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Try to generate to an invalid path
	err := Generate("/nonexistent/path/that/cannot/be/created/\x00invalid")
	if err == nil {
		t.Error("expected error for invalid output directory")
	}
}
