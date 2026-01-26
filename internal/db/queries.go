package db

import (
	"bookshelf/internal/models"
	"database/sql"
	"fmt"
	"time"
)

func AddBook(title, author string, isbn, coverURL, description, openLibraryKey *string, pages *int) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO books (title, author, isbn, pages, cover_url, description, open_library_key)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, title, author, isbn, pages, coverURL, description, openLibraryKey)
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
			b.id, b.title, b.author, b.isbn, b.pages, b.cover_url, b.description, b.open_library_key, b.created_at,
			r.id, r.book_id, r.status, r.started_at, r.finished_at, r.rating, r.review, r.updated_at
		FROM books b
		LEFT JOIN reading_entries r ON b.id = r.book_id
		WHERE b.id = ?
	`, id)

	var book models.BookWithEntry
	err := row.Scan(
		&book.Book.ID, &book.Book.Title, &book.Book.Author, &book.Book.ISBN,
		&book.Book.Pages, &book.Book.CoverURL, &book.Book.Description,
		&book.Book.OpenLibraryKey, &book.Book.CreatedAt,
		&book.ReadingEntry.ID, &book.ReadingEntry.BookID, &book.ReadingEntry.Status,
		&book.ReadingEntry.StartedAt, &book.ReadingEntry.FinishedAt,
		&book.ReadingEntry.Rating, &book.ReadingEntry.Review, &book.ReadingEntry.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func ListBooks(statusFilter *models.BookStatus) ([]models.BookWithEntry, error) {
	query := `
		SELECT
			b.id, b.title, b.author, b.isbn, b.pages, b.cover_url, b.description, b.open_library_key, b.created_at,
			r.id, r.book_id, r.status, r.started_at, r.finished_at, r.rating, r.review, r.updated_at
		FROM books b
		LEFT JOIN reading_entries r ON b.id = r.book_id
	`
	var args []interface{}
	if statusFilter != nil {
		query += " WHERE r.status = ?"
		args = append(args, *statusFilter)
	}
	query += " ORDER BY b.created_at DESC"

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
			&book.Book.OpenLibraryKey, &book.Book.CreatedAt,
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
	now := time.Now().Format("2006-01-02 15:04:05")
	var startedAt, finishedAt interface{}

	switch status {
	case models.StatusReading:
		startedAt = now
	case models.StatusFinished:
		finishedAt = now
	}

	_, err := DB.Exec(`
		UPDATE reading_entries
		SET status = ?, started_at = COALESCE(?, started_at), finished_at = ?, updated_at = ?
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
