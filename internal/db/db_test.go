package db

import (
	"bookshelf/internal/models"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) func() {
	t.Helper()

	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "bookshelf-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	var dbErr error
	DB, dbErr = sql.Open("sqlite", dbPath)
	if dbErr != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open database: %v", dbErr)
	}

	if err := migrate(); err != nil {
		DB.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	return func() {
		DB.Close()
		os.RemoveAll(tmpDir)
	}
}

func TestAddBook(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	isbn := "978-0135957059"
	coverURL := "https://example.com/cover.jpg"
	description := "A great book"
	olKey := "/works/OL123"
	pages := 352

	id, err := AddBook("The Pragmatic Programmer", "Andy Hunt", &isbn, &coverURL, &description, &olKey, &pages)
	if err != nil {
		t.Fatalf("failed to add book: %v", err)
	}

	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}

func TestAddBookMinimalFields(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, err := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to add book: %v", err)
	}

	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}

func TestCreateReadingEntry(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)

	err := CreateReadingEntry(id, models.StatusWantToRead)
	if err != nil {
		t.Fatalf("failed to create reading entry: %v", err)
	}
}

func TestGetBook(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	CreateReadingEntry(id, models.StatusWantToRead)

	book, err := GetBook(id)
	if err != nil {
		t.Fatalf("failed to get book: %v", err)
	}

	if book.Book.Title != "Test Book" {
		t.Errorf("expected title 'Test Book', got %s", book.Book.Title)
	}

	if book.Book.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got %s", book.Book.Author)
	}

	if book.ReadingEntry.Status != models.StatusWantToRead {
		t.Errorf("expected status 'want-to-read', got %s", book.ReadingEntry.Status)
	}
}

func TestGetBookNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	_, err := GetBook(999)
	if err == nil {
		t.Error("expected error for non-existent book")
	}
}

func TestListBooks(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Add multiple books
	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	id3, _ := AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusFinished)

	// List all books
	books, err := ListBooks(nil)
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if len(books) != 3 {
		t.Errorf("expected 3 books, got %d", len(books))
	}
}

func TestListBooksWithFilter(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	id3, _ := AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusFinished)

	// Filter by status
	status := models.StatusReading
	books, err := ListBooks(&status)
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("expected 1 book, got %d", len(books))
	}

	if books[0].Book.Title != "Book 2" {
		t.Errorf("expected 'Book 2', got %s", books[0].Book.Title)
	}
}

func TestUpdateStatus(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	CreateReadingEntry(id, models.StatusWantToRead)

	// Update to reading
	err := UpdateStatus(id, models.StatusReading)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	book, _ := GetBook(id)
	if book.ReadingEntry.Status != models.StatusReading {
		t.Errorf("expected status 'reading', got %s", book.ReadingEntry.Status)
	}

	if !book.ReadingEntry.StartedAt.Valid {
		t.Error("expected started_at to be set")
	}

	// Update to finished
	err = UpdateStatus(id, models.StatusFinished)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	book, _ = GetBook(id)
	if book.ReadingEntry.Status != models.StatusFinished {
		t.Errorf("expected status 'finished', got %s", book.ReadingEntry.Status)
	}

	if !book.ReadingEntry.FinishedAt.Valid {
		t.Error("expected finished_at to be set")
	}
}

func TestUpdateRating(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	CreateReadingEntry(id, models.StatusFinished)

	err := UpdateRating(id, 5)
	if err != nil {
		t.Fatalf("failed to update rating: %v", err)
	}

	book, _ := GetBook(id)
	if !book.ReadingEntry.Rating.Valid || book.ReadingEntry.Rating.Int64 != 5 {
		t.Errorf("expected rating 5, got %v", book.ReadingEntry.Rating)
	}
}

func TestUpdateReview(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	CreateReadingEntry(id, models.StatusFinished)

	review := "This was a great book!"
	err := UpdateReview(id, review)
	if err != nil {
		t.Fatalf("failed to update review: %v", err)
	}

	book, _ := GetBook(id)
	if !book.ReadingEntry.Review.Valid || book.ReadingEntry.Review.String != review {
		t.Errorf("expected review '%s', got %v", review, book.ReadingEntry.Review)
	}
}

func TestGetReview(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)
	CreateReadingEntry(id, models.StatusFinished)

	expectedReview := "Great book!"
	UpdateReview(id, expectedReview)

	review, err := GetReview(id)
	if err != nil {
		t.Fatalf("failed to get review: %v", err)
	}

	if review != expectedReview {
		t.Errorf("expected '%s', got '%s'", expectedReview, review)
	}
}

func TestBookExists(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil)

	exists, err := BookExists(id)
	if err != nil {
		t.Fatalf("failed to check book exists: %v", err)
	}

	if !exists {
		t.Error("expected book to exist")
	}

	exists, err = BookExists(999)
	if err != nil {
		t.Fatalf("failed to check book exists: %v", err)
	}

	if exists {
		t.Error("expected book to not exist")
	}
}

func TestGetStats(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Add books with different statuses
	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	pages := 300
	id3, _ := AddBook("Book 3", "Author 3", nil, nil, nil, nil, &pages)
	CreateReadingEntry(id3, models.StatusFinished)
	UpdateRating(id3, 4)

	id4, _ := AddBook("Book 4", "Author 4", nil, nil, nil, nil, nil)
	CreateReadingEntry(id4, models.StatusFinished)
	UpdateRating(id4, 5)

	stats, err := GetStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalBooks != 4 {
		t.Errorf("expected 4 total books, got %d", stats.TotalBooks)
	}

	if stats.WantToRead != 1 {
		t.Errorf("expected 1 want-to-read, got %d", stats.WantToRead)
	}

	if stats.Reading != 1 {
		t.Errorf("expected 1 reading, got %d", stats.Reading)
	}

	if stats.Finished != 2 {
		t.Errorf("expected 2 finished, got %d", stats.Finished)
	}

	if stats.RatedBooksCount != 2 {
		t.Errorf("expected 2 rated books, got %d", stats.RatedBooksCount)
	}

	expectedAvg := 4.5
	if stats.AverageRating != expectedAvg {
		t.Errorf("expected average rating %.1f, got %.1f", expectedAvg, stats.AverageRating)
	}
}
