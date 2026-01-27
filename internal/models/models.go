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

type ReadingGoal struct {
	ID        int64
	Year      int
	Target    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SortField string

const (
	SortByAdded  SortField = "added"
	SortByTitle  SortField = "title"
	SortByAuthor SortField = "author"
	SortByRating SortField = "rating"
)

type ListOptions struct {
	StatusFilter *BookStatus
	SearchQuery  string
	SortBy       SortField
}

// SiteConfig holds configurable website parameters.
type SiteConfig struct {
	Title       string // Site title (default: "My Bookshelf")
	Subtitle    string // Site subtitle (default: "Personal Reading Tracker")
	Author      string // Author name for attribution
	Description string // Meta description
	BaseURL     string // Base URL for the site (for canonical links)
}

// DefaultSiteConfig returns the default configuration.
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Title:       "My Bookshelf",
		Subtitle:    "Personal Reading Tracker",
		Author:      "",
		Description: "",
		BaseURL:     "",
	}
}
