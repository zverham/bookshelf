package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init() error {
	dbPath := os.Getenv("BOOKSHELF_DB_PATH")
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		dbDir := filepath.Join(homeDir, ".bookshelf")
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}

		dbPath = filepath.Join(dbDir, "bookshelf.db")
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	return Migrate()
}

// Migrate creates the database schema. Exported for use by test helpers.
func Migrate() error {
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
		genres TEXT,
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

	CREATE INDEX IF NOT EXISTS idx_reading_entries_book_id ON reading_entries(book_id);
	CREATE INDEX IF NOT EXISTS idx_reading_entries_status ON reading_entries(status);

	CREATE TABLE IF NOT EXISTS reading_goals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL UNIQUE,
		target INTEGER NOT NULL CHECK(target > 0),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_reading_goals_year ON reading_goals(year);

	CREATE TABLE IF NOT EXISTS site_config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := DB.Exec(schema); err != nil {
		return err
	}

	// Add genres column to existing databases (ignore error if column already exists)
	DB.Exec("ALTER TABLE books ADD COLUMN genres TEXT")

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
