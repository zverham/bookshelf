package db

import (
	"bookshelf/internal/models"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func AddBook(title, author string, isbn, coverURL, description, openLibraryKey, genres *string, pages *int) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO books (title, author, isbn, pages, cover_url, description, open_library_key, genres)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, title, author, isbn, pages, coverURL, description, openLibraryKey, genres)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func CreateReadingEntry(bookID int64, status models.BookStatus) error {
	_, err := DB.Exec(`
		INSERT INTO reading_entries (book_id, status, updated_at)
		VALUES (?, ?, ?)
	`, bookID, status, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

func GetBook(id int64) (*models.BookWithEntry, error) {
	row := DB.QueryRow(`
		SELECT
			b.id, b.title, b.author, b.isbn, b.pages, b.cover_url, b.description, b.open_library_key, b.genres, b.created_at,
			r.id, r.book_id, r.status, r.started_at, r.finished_at, r.rating, r.review, r.updated_at
		FROM books b
		LEFT JOIN reading_entries r ON b.id = r.book_id
		WHERE b.id = ?
	`, id)

	var book models.BookWithEntry
	err := row.Scan(
		&book.Book.ID, &book.Book.Title, &book.Book.Author, &book.Book.ISBN,
		&book.Book.Pages, &book.Book.CoverURL, &book.Book.Description,
		&book.Book.OpenLibraryKey, &book.Book.Genres, &book.Book.CreatedAt,
		&book.ReadingEntry.ID, &book.ReadingEntry.BookID, &book.ReadingEntry.Status,
		&book.ReadingEntry.StartedAt, &book.ReadingEntry.FinishedAt,
		&book.ReadingEntry.Rating, &book.ReadingEntry.Review, &book.ReadingEntry.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func ListBooks(opts models.ListOptions) ([]models.BookWithEntry, error) {
	query := `
		SELECT
			b.id, b.title, b.author, b.isbn, b.pages, b.cover_url, b.description, b.open_library_key, b.genres, b.created_at,
			r.id, r.book_id, r.status, r.started_at, r.finished_at, r.rating, r.review, r.updated_at
		FROM books b
		LEFT JOIN reading_entries r ON b.id = r.book_id
	`
	var args []any
	var conditions []string

	if opts.StatusFilter != nil {
		conditions = append(conditions, "r.status = ?")
		args = append(args, *opts.StatusFilter)
	}

	if opts.SearchQuery != "" {
		conditions = append(conditions, "(LOWER(b.title) LIKE ? OR LOWER(b.author) LIKE ?)")
		searchPattern := "%" + strings.ToLower(opts.SearchQuery) + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Apply sorting
	switch opts.SortBy {
	case models.SortByTitle:
		query += " ORDER BY LOWER(b.title) ASC"
	case models.SortByAuthor:
		query += " ORDER BY LOWER(b.author) ASC"
	case models.SortByRating:
		query += " ORDER BY r.rating DESC NULLS LAST, b.created_at DESC"
	default: // SortByAdded or empty
		query += " ORDER BY b.created_at DESC"
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.BookWithEntry
	for rows.Next() {
		var book models.BookWithEntry
		err := rows.Scan(
			&book.Book.ID, &book.Book.Title, &book.Book.Author, &book.Book.ISBN,
			&book.Book.Pages, &book.Book.CoverURL, &book.Book.Description,
			&book.Book.OpenLibraryKey, &book.Book.Genres, &book.Book.CreatedAt,
			&book.ReadingEntry.ID, &book.ReadingEntry.BookID, &book.ReadingEntry.Status,
			&book.ReadingEntry.StartedAt, &book.ReadingEntry.FinishedAt,
			&book.ReadingEntry.Rating, &book.ReadingEntry.Review, &book.ReadingEntry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func UpdateStatus(bookID int64, status models.BookStatus) error {
	return UpdateStatusWithDate(bookID, status, true)
}

func UpdateStatusWithDate(bookID int64, status models.BookStatus, setDate bool) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	var startedAt, finishedAt any

	if setDate {
		switch status {
		case models.StatusReading:
			startedAt = now
		case models.StatusFinished:
			finishedAt = now
		}
	}

	_, err := DB.Exec(`
		UPDATE reading_entries
		SET status = ?, started_at = COALESCE(?, started_at), finished_at = COALESCE(?, finished_at), updated_at = ?
		WHERE book_id = ?
	`, status, startedAt, finishedAt, now, bookID)
	return err
}

func UpdateRating(bookID int64, rating int) error {
	_, err := DB.Exec(`
		UPDATE reading_entries
		SET rating = ?, updated_at = ?
		WHERE book_id = ?
	`, rating, time.Now().Format("2006-01-02 15:04:05"), bookID)
	return err
}

func UpdateReview(bookID int64, review string) error {
	_, err := DB.Exec(`
		UPDATE reading_entries
		SET review = ?, updated_at = ?
		WHERE book_id = ?
	`, review, time.Now().Format("2006-01-02 15:04:05"), bookID)
	return err
}

type Stats struct {
	TotalBooks      int
	WantToRead      int
	Reading         int
	Finished        int
	BooksThisYear   int
	PagesThisYear   int
	AverageRating   float64
	RatedBooksCount int
}

func GetStats() (*Stats, error) {
	stats := &Stats{}
	currentYear := fmt.Sprintf("%d", time.Now().Year())

	// Total books by status
	row := DB.QueryRow(`SELECT COUNT(*) FROM books`)
	row.Scan(&stats.TotalBooks)

	row = DB.QueryRow(`SELECT COUNT(*) FROM reading_entries WHERE status = 'want-to-read'`)
	row.Scan(&stats.WantToRead)

	row = DB.QueryRow(`SELECT COUNT(*) FROM reading_entries WHERE status = 'reading'`)
	row.Scan(&stats.Reading)

	row = DB.QueryRow(`SELECT COUNT(*) FROM reading_entries WHERE status = 'finished'`)
	row.Scan(&stats.Finished)

	// Books finished this year
	row = DB.QueryRow(`
		SELECT COUNT(*) FROM reading_entries
		WHERE status = 'finished' AND strftime('%Y', finished_at) = ?
	`, currentYear)
	row.Scan(&stats.BooksThisYear)

	// Pages read this year
	row = DB.QueryRow(`
		SELECT COALESCE(SUM(b.pages), 0) FROM books b
		JOIN reading_entries r ON b.id = r.book_id
		WHERE r.status = 'finished' AND strftime('%Y', r.finished_at) = ?
	`, currentYear)
	row.Scan(&stats.PagesThisYear)

	// Average rating
	row = DB.QueryRow(`
		SELECT COALESCE(AVG(rating), 0), COUNT(rating) FROM reading_entries
		WHERE rating IS NOT NULL
	`)
	row.Scan(&stats.AverageRating, &stats.RatedBooksCount)

	return stats, nil
}

func BookExists(id int64) (bool, error) {
	var exists bool
	err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM books WHERE id = ?)`, id).Scan(&exists)
	return exists, err
}

func GetReview(bookID int64) (string, error) {
	var review sql.NullString
	err := DB.QueryRow(`SELECT review FROM reading_entries WHERE book_id = ?`, bookID).Scan(&review)
	if err != nil {
		return "", err
	}
	return review.String, nil
}

func DeleteBook(bookID int64) error {
	_, err := DB.Exec(`DELETE FROM reading_entries WHERE book_id = ?`, bookID)
	if err != nil {
		return err
	}
	_, err = DB.Exec(`DELETE FROM books WHERE id = ?`, bookID)
	return err
}

// UpdateBookMetadata updates optional fields that may have been missing when the book was added.
func UpdateBookMetadata(bookID int64, description, genres *string) error {
	_, err := DB.Exec(`
		UPDATE books
		SET description = COALESCE(?, description),
		    genres = COALESCE(?, genres)
		WHERE id = ?
	`, description, genres, bookID)
	return err
}

// GetBooksWithOpenLibraryKey returns all books that have an Open Library key for refreshing metadata.
func GetBooksWithOpenLibraryKey() ([]models.BookWithEntry, error) {
	rows, err := DB.Query(`
		SELECT
			b.id, b.title, b.author, b.isbn, b.pages, b.cover_url, b.description, b.open_library_key, b.genres, b.created_at,
			r.id, r.book_id, r.status, r.started_at, r.finished_at, r.rating, r.review, r.updated_at
		FROM books b
		LEFT JOIN reading_entries r ON b.id = r.book_id
		WHERE b.open_library_key IS NOT NULL AND b.open_library_key != ''
		ORDER BY b.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.BookWithEntry
	for rows.Next() {
		var book models.BookWithEntry
		err := rows.Scan(
			&book.Book.ID, &book.Book.Title, &book.Book.Author, &book.Book.ISBN,
			&book.Book.Pages, &book.Book.CoverURL, &book.Book.Description,
			&book.Book.OpenLibraryKey, &book.Book.Genres, &book.Book.CreatedAt,
			&book.ReadingEntry.ID, &book.ReadingEntry.BookID, &book.ReadingEntry.Status,
			&book.ReadingEntry.StartedAt, &book.ReadingEntry.FinishedAt,
			&book.ReadingEntry.Rating, &book.ReadingEntry.Review, &book.ReadingEntry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

// SetGoal creates or updates a reading goal for a given year (UPSERT).
func SetGoal(year, target int) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(`
		INSERT INTO reading_goals (year, target, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(year) DO UPDATE SET target = excluded.target, updated_at = excluded.updated_at
	`, year, target, now, now)
	return err
}

// GetGoal retrieves a reading goal for a specific year. Returns nil if not found.
func GetGoal(year int) (*models.ReadingGoal, error) {
	row := DB.QueryRow(`
		SELECT id, year, target, created_at, updated_at
		FROM reading_goals
		WHERE year = ?
	`, year)

	var goal models.ReadingGoal
	err := row.Scan(&goal.ID, &goal.Year, &goal.Target, &goal.CreatedAt, &goal.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

// GetAllGoals retrieves all reading goals ordered by year descending.
func GetAllGoals() ([]models.ReadingGoal, error) {
	rows, err := DB.Query(`
		SELECT id, year, target, created_at, updated_at
		FROM reading_goals
		ORDER BY year DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.ReadingGoal
	for rows.Next() {
		var goal models.ReadingGoal
		err := rows.Scan(&goal.ID, &goal.Year, &goal.Target, &goal.CreatedAt, &goal.UpdatedAt)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, nil
}

// ClearGoal deletes a reading goal for a specific year.
func ClearGoal(year int) error {
	result, err := DB.Exec(`DELETE FROM reading_goals WHERE year = ?`, year)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no goal found for year %d", year)
	}
	return nil
}

// GetBooksFinishedInYear returns the count of books finished in a given year.
func GetBooksFinishedInYear(year int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM reading_entries
		WHERE status = 'finished' AND strftime('%Y', finished_at) = ?
	`, fmt.Sprintf("%d", year)).Scan(&count)
	return count, err
}

// Site configuration functions

// SetConfig sets a configuration value.
func SetConfig(key, value string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(`
		INSERT INTO site_config (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value, now)
	return err
}

// GetConfig retrieves a configuration value. Returns empty string if not found.
func GetConfig(key string) (string, error) {
	var value string
	err := DB.QueryRow(`SELECT value FROM site_config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetAllConfig retrieves all configuration values as a map.
func GetAllConfig() (map[string]string, error) {
	rows, err := DB.Query(`SELECT key, value FROM site_config ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		config[key] = value
	}
	return config, nil
}

// DeleteConfig removes a configuration value.
func DeleteConfig(key string) error {
	result, err := DB.Exec(`DELETE FROM site_config WHERE key = ?`, key)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("config key '%s' not found", key)
	}
	return nil
}

// GetSiteConfig retrieves the full site configuration with defaults.
func GetSiteConfig() (models.SiteConfig, error) {
	config := models.DefaultSiteConfig()

	allConfig, err := GetAllConfig()
	if err != nil {
		return config, err
	}

	if v, ok := allConfig["site.title"]; ok && v != "" {
		config.Title = v
	}
	if v, ok := allConfig["site.subtitle"]; ok && v != "" {
		config.Subtitle = v
	}
	if v, ok := allConfig["site.author"]; ok && v != "" {
		config.Author = v
	}
	if v, ok := allConfig["site.description"]; ok && v != "" {
		config.Description = v
	}
	if v, ok := allConfig["site.base_url"]; ok && v != "" {
		config.BaseURL = v
	}

	return config, nil
}
