# Bookshelf CLI - Development Guide

## Project Overview

A personal reading tracker CLI written in Go. Users can search Open Library, add books to their shelf, track reading progress, rate/review books, and publish a static website.

## Quick Commands

```bash
# Build and run
go build -o bookshelf .
go test ./...

# Or use just (recommended)
just build        # Build binary
just test         # Run tests
just test-v       # Verbose tests
just lint         # Run golangci-lint
just fmt          # Format code
```

## Project Structure

```
bookshelf/
├── main.go                 # Entry point - calls cmd.Execute()
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command, DB init/close hooks
│   ├── add.go             # Search + add book
│   ├── list.go            # List books with filters
│   ├── show.go            # Show book details
│   ├── start.go           # Mark book as reading
│   ├── finish.go          # Mark book as finished
│   ├── rate.go            # Rate 1-5 stars
│   ├── review.go          # Edit review in $EDITOR
│   ├── remove.go          # Remove book
│   ├── refresh.go         # Back-fill metadata from API
│   ├── search.go          # Search without adding
│   ├── stats.go           # Reading statistics
│   ├── publish.go         # Generate static site
│   └── integration_test.go # CLI integration tests
├── internal/
│   ├── api/               # Open Library API client
│   │   └── openlibrary.go
│   ├── db/                # SQLite database
│   │   ├── db.go          # Init, migrate, connection
│   │   └── queries.go     # All SQL operations
│   ├── models/            # Data structures
│   │   └── models.go      # Book, ReadingEntry structs
│   ├── publish/           # Static site generator
│   │   └── publish.go     # HTML templates
│   ├── stats/             # Statistics calculations
│   │   └── stats.go
│   └── testutil/          # Shared test helpers
│       └── testutil.go
└── public/                # Generated static site output
```

## Key Patterns

### Database

- SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- Database location: `~/.bookshelf/bookshelf.db`
- Override with `BOOKSHELF_DB_PATH` env var (used in tests)
- Schema migrations in `internal/db/db.go` `Migrate()`
- All queries in `internal/db/queries.go`

### CLI Commands

- Uses Cobra framework
- Root command handles DB init/close via `PersistentPreRunE`/`PersistentPostRun`
- Each command in separate file under `cmd/`
- Register new commands in `cmd/root.go` `init()`

### Models

- `Book` - Core book data with `sql.NullString` for optional fields
- `ReadingEntry` - Tracks status changes (want-to-read, reading, finished)
- `BookWithEntry` - Joined view for display
- Genres stored as JSON array string in `genres` column

### API Integration

- Open Library API (`openlibrary.org`)
- Search: `/search.json?q=...`
- Work details: `/works/{key}.json` (for descriptions, subjects)
- Cover images: `covers.openlibrary.org/b/id/{id}-M.jpg`

### Testing

- Unit tests alongside source files (`*_test.go`)
- Integration tests in `cmd/integration_test.go` (build + run binary)
- Test helpers in `internal/testutil/testutil.go`:
  - `SetupTestDB(t)` - Creates temp DB, returns cleanup func
  - `CaptureOutput(t, f)` - Captures stdout
  - `StrPtr(s)`, `IntPtr(i)` - Pointer helpers

### Static Site Generation

- Templates embedded in `internal/publish/publish.go`
- Output to `./public/` by default
- Dark mode support via CSS `prefers-color-scheme`
- Filter tabs (All/Reading/Finished/Want to Read)

## Database Schema

```sql
books (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    isbn TEXT,
    cover_url TEXT,
    description TEXT,
    genres TEXT,           -- JSON array of strings
    pages INTEGER,
    open_library_key TEXT,
    rating INTEGER,        -- 1-5
    review TEXT,
    created_at DATETIME,
    updated_at DATETIME
)

reading_entries (
    id INTEGER PRIMARY KEY,
    book_id INTEGER REFERENCES books(id),
    status TEXT,           -- want-to-read, reading, finished
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME
)
```

## Common Development Tasks

### Adding a new command

1. Create `cmd/newcmd.go` with Cobra command
2. Register in `cmd/root.go` `init()`: `rootCmd.AddCommand(newCmd)`
3. Add to justfile if useful

### Adding a database column

1. Add column to schema in `internal/db/db.go` `Migrate()`
2. Use `ALTER TABLE ... ADD COLUMN` with error handling (column may exist)
3. Update model in `internal/models/models.go`
4. Update queries in `internal/db/queries.go` (SELECT, INSERT, Scan)

### Running specific tests

```bash
go test ./internal/db/...           # Just db package
go test -run TestAddBook ./...      # Just one test
go test -v -race ./...              # Verbose with race detection
```

## CI Pipeline

- Runs on push/PR to main
- Jobs: lint (golangci-lint), test (with coverage), build
- Config: `.github/workflows/ci.yml`
- Linter config: `.golangci.yml`

## GitHub Pages Deployment

1. Export local DB: `just export-db` (copies to `./bookshelf.db`)
2. Commit and push
3. Workflow builds site from committed DB and deploys
