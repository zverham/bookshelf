// Package testutil provides shared test helpers for the bookshelf project.
package testutil

import (
	"bookshelf/internal/db"
	"bytes"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// SetupTestDB creates a temporary SQLite database for testing.
// Returns a cleanup function that must be deferred.
func SetupTestDB(t *testing.T) func() {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "bookshelf-test-*")
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

	if err := db.Migrate(); err != nil {
		db.DB.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	return func() {
		db.DB.Close()
		os.RemoveAll(tmpDir)
	}
}

// CaptureOutput captures stdout during function execution and returns it as a string.
func CaptureOutput(t *testing.T, f func()) string {
	t.Helper()

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

// StrPtr returns a pointer to the given string.
func StrPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int {
	return &i
}
