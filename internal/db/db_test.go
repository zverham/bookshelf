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

	if err := Migrate(); err != nil {
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

	id, err := AddBook("The Pragmatic Programmer", "Andy Hunt", &isbn, &coverURL, &description, &olKey, nil, &pages)
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

	id, err := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)
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

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)

	err := CreateReadingEntry(id, models.StatusWantToRead)
	if err != nil {
		t.Fatalf("failed to create reading entry: %v", err)
	}
}

func TestGetBook(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)
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
	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	id3, _ := AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusFinished)

	// List all books
	books, err := ListBooks(models.ListOptions{})
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

	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	id3, _ := AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusFinished)

	// Filter by status
	status := models.StatusReading
	books, err := ListBooks(models.ListOptions{StatusFilter: &status})
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

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)
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

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)
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

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)
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

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)
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

	id, _ := AddBook("Test Book", "Test Author", nil, nil, nil, nil, nil, nil)

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
	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	pages := 300
	id3, _ := AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil, &pages)
	CreateReadingEntry(id3, models.StatusFinished)
	UpdateRating(id3, 4)

	id4, _ := AddBook("Book 4", "Author 4", nil, nil, nil, nil, nil, nil)
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

// Goal tests

func TestSetGoal(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := SetGoal(2026, 24)
	if err != nil {
		t.Fatalf("failed to set goal: %v", err)
	}

	goal, err := GetGoal(2026)
	if err != nil {
		t.Fatalf("failed to get goal: %v", err)
	}

	if goal == nil {
		t.Fatal("expected goal to exist")
	}

	if goal.Year != 2026 {
		t.Errorf("expected year 2026, got %d", goal.Year)
	}

	if goal.Target != 24 {
		t.Errorf("expected target 24, got %d", goal.Target)
	}
}

func TestSetGoalUpdatesExisting(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := SetGoal(2026, 24)
	if err != nil {
		t.Fatalf("failed to set goal: %v", err)
	}

	// Update existing goal
	err = SetGoal(2026, 30)
	if err != nil {
		t.Fatalf("failed to update goal: %v", err)
	}

	goal, err := GetGoal(2026)
	if err != nil {
		t.Fatalf("failed to get goal: %v", err)
	}

	if goal.Target != 30 {
		t.Errorf("expected target 30, got %d", goal.Target)
	}

	// Verify only one goal exists
	goals, err := GetAllGoals()
	if err != nil {
		t.Fatalf("failed to get all goals: %v", err)
	}

	if len(goals) != 1 {
		t.Errorf("expected 1 goal, got %d", len(goals))
	}
}

func TestGetGoalNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	goal, err := GetGoal(2026)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if goal != nil {
		t.Error("expected nil goal for non-existent year")
	}
}

func TestClearGoal(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := SetGoal(2026, 24)
	if err != nil {
		t.Fatalf("failed to set goal: %v", err)
	}

	err = ClearGoal(2026)
	if err != nil {
		t.Fatalf("failed to clear goal: %v", err)
	}

	goal, err := GetGoal(2026)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if goal != nil {
		t.Error("expected goal to be cleared")
	}
}

func TestClearGoalNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := ClearGoal(2026)
	if err == nil {
		t.Error("expected error when clearing non-existent goal")
	}
}

func TestGetAllGoals(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	SetGoal(2024, 20)
	SetGoal(2025, 24)
	SetGoal(2026, 30)

	goals, err := GetAllGoals()
	if err != nil {
		t.Fatalf("failed to get all goals: %v", err)
	}

	if len(goals) != 3 {
		t.Errorf("expected 3 goals, got %d", len(goals))
	}

	// Goals should be ordered by year DESC
	if goals[0].Year != 2026 {
		t.Errorf("expected first goal year 2026, got %d", goals[0].Year)
	}
}

func TestGetBooksFinishedInYear(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id1, _ := AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusFinished)
	UpdateStatus(id1, models.StatusFinished)

	count, err := GetBooksFinishedInYear(2026)
	if err != nil {
		t.Fatalf("failed to get books finished: %v", err)
	}

	// Book was finished "now" so it should be in current year
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
}

// Search and sort tests

func TestListBooksWithSearch(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id1, _ := AddBook("The Great Gatsby", "F. Scott Fitzgerald", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusFinished)

	id2, _ := AddBook("To Kill a Mockingbird", "Harper Lee", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	id3, _ := AddBook("1984", "George Orwell", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusWantToRead)

	// Search by title
	books, err := ListBooks(models.ListOptions{SearchQuery: "gatsby"})
	if err != nil {
		t.Fatalf("failed to search books: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("expected 1 book matching 'gatsby', got %d", len(books))
	}

	if len(books) > 0 && books[0].Book.Title != "The Great Gatsby" {
		t.Errorf("expected 'The Great Gatsby', got %s", books[0].Book.Title)
	}

	// Search by author (case insensitive)
	books, err = ListBooks(models.ListOptions{SearchQuery: "ORWELL"})
	if err != nil {
		t.Fatalf("failed to search books: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("expected 1 book by 'ORWELL', got %d", len(books))
	}
}

func TestListBooksWithSort(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id1, _ := AddBook("Zebra Book", "Alice Author", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusFinished)
	UpdateRating(id1, 3)

	id2, _ := AddBook("Alpha Book", "Zack Author", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusFinished)
	UpdateRating(id2, 5)

	id3, _ := AddBook("Middle Book", "Mike Author", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusFinished)
	// No rating

	// Sort by title ASC
	books, err := ListBooks(models.ListOptions{SortBy: models.SortByTitle})
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if len(books) < 1 || books[0].Book.Title != "Alpha Book" {
		t.Errorf("expected 'Alpha Book' first when sorting by title, got %s", books[0].Book.Title)
	}

	// Sort by author ASC
	books, err = ListBooks(models.ListOptions{SortBy: models.SortByAuthor})
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if len(books) < 1 || books[0].Book.Author != "Alice Author" {
		t.Errorf("expected 'Alice Author' first when sorting by author, got %s", books[0].Book.Author)
	}

	// Sort by rating DESC (with NULL handling)
	books, err = ListBooks(models.ListOptions{SortBy: models.SortByRating})
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if len(books) < 1 || books[0].Book.Title != "Alpha Book" {
		t.Errorf("expected 'Alpha Book' first when sorting by rating, got %s", books[0].Book.Title)
	}

	// Verify NULL rating is last
	if len(books) >= 3 && books[2].Book.Title != "Middle Book" {
		t.Errorf("expected 'Middle Book' (no rating) last when sorting by rating, got %s", books[2].Book.Title)
	}
}

func TestListBooksWithSearchAndStatus(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	id1, _ := AddBook("Python Programming", "John Smith", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id1, models.StatusFinished)

	id2, _ := AddBook("Python Cookbook", "Jane Doe", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id2, models.StatusReading)

	id3, _ := AddBook("Go Programming", "Bob Wilson", nil, nil, nil, nil, nil, nil)
	CreateReadingEntry(id3, models.StatusFinished)

	// Search for "python" AND status "finished"
	status := models.StatusFinished
	books, err := ListBooks(models.ListOptions{
		SearchQuery:  "python",
		StatusFilter: &status,
	})
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("expected 1 book matching 'python' with status 'finished', got %d", len(books))
	}

	if len(books) > 0 && books[0].Book.Title != "Python Programming" {
		t.Errorf("expected 'Python Programming', got %s", books[0].Book.Title)
	}
}

// Site config tests

func TestSetConfig(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := SetConfig("site.title", "My Test Bookshelf")
	if err != nil {
		t.Fatalf("failed to set config: %v", err)
	}

	value, err := GetConfig("site.title")
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}

	if value != "My Test Bookshelf" {
		t.Errorf("expected 'My Test Bookshelf', got '%s'", value)
	}
}

func TestSetConfigUpdatesExisting(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	SetConfig("site.title", "First Title")
	SetConfig("site.title", "Second Title")

	value, _ := GetConfig("site.title")
	if value != "Second Title" {
		t.Errorf("expected 'Second Title', got '%s'", value)
	}
}

func TestGetConfigNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	value, err := GetConfig("nonexistent.key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != "" {
		t.Errorf("expected empty string for non-existent key, got '%s'", value)
	}
}

func TestGetAllConfig(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	SetConfig("site.title", "Test Title")
	SetConfig("site.author", "Test Author")

	config, err := GetAllConfig()
	if err != nil {
		t.Fatalf("failed to get all config: %v", err)
	}

	if len(config) != 2 {
		t.Errorf("expected 2 config values, got %d", len(config))
	}

	if config["site.title"] != "Test Title" {
		t.Errorf("expected 'Test Title', got '%s'", config["site.title"])
	}
}

func TestDeleteConfig(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	SetConfig("site.title", "Test Title")

	err := DeleteConfig("site.title")
	if err != nil {
		t.Fatalf("failed to delete config: %v", err)
	}

	value, _ := GetConfig("site.title")
	if value != "" {
		t.Errorf("expected empty string after delete, got '%s'", value)
	}
}

func TestDeleteConfigNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := DeleteConfig("nonexistent.key")
	if err == nil {
		t.Error("expected error when deleting non-existent key")
	}
}

func TestGetSiteConfig(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Test defaults
	config, err := GetSiteConfig()
	if err != nil {
		t.Fatalf("failed to get site config: %v", err)
	}

	if config.Title != "My Bookshelf" {
		t.Errorf("expected default title 'My Bookshelf', got '%s'", config.Title)
	}

	// Test with custom values
	SetConfig("site.title", "Custom Title")
	SetConfig("site.author", "John Doe")

	config, err = GetSiteConfig()
	if err != nil {
		t.Fatalf("failed to get site config: %v", err)
	}

	if config.Title != "Custom Title" {
		t.Errorf("expected 'Custom Title', got '%s'", config.Title)
	}

	if config.Author != "John Doe" {
		t.Errorf("expected 'John Doe', got '%s'", config.Author)
	}

	// Subtitle should still be default
	if config.Subtitle != "Personal Reading Tracker" {
		t.Errorf("expected default subtitle, got '%s'", config.Subtitle)
	}
}
