package models

import (
	"database/sql"
	"time"
)

type BookStatus string

const (
	StatusWantToRead BookStatus = "want-to-read"
	StatusReading    BookStatus = "reading"
	StatusFinished   BookStatus = "finished"
)

type Book struct {
	ID             int64
	Title          string
	Author         string
	ISBN           sql.NullString
	Pages          sql.NullInt64
	CoverURL       sql.NullString
	Description    sql.NullString
	OpenLibraryKey sql.NullString
	Genres         sql.NullString
	CreatedAt      time.Time
}

type ReadingEntry struct {
	ID         int64
	BookID     int64
	Status     BookStatus
	StartedAt  sql.NullTime
	FinishedAt sql.NullTime
	Rating     sql.NullInt64
	Review     sql.NullString
	UpdatedAt  time.Time
}

type BookWithEntry struct {
	Book
	ReadingEntry
}
