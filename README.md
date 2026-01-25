# Bookshelf

A personal reading tracker CLI, similar to Goodreads. Track books you want to read, are currently reading, and have finished. Rate and review your reads, view statistics, and publish your bookshelf as a static website.

## Requirements

- Go 1.24+
- [just](https://github.com/casey/just) (optional, for convenience commands)

## Installation

```bash
# Clone and build
git clone <repo-url>
cd bookshelf
go build -o bookshelf .

# Or install to $GOPATH/bin
go install .
```

## Usage

### Adding Books

Search Open Library and add books to your shelf:

```bash
bookshelf add "The Great Gatsby"
```

This searches for the book, displays results, and prompts you to select one. The book is added with status "want-to-read".

### Searching Without Adding

Browse Open Library without adding to your shelf:

```bash
bookshelf search "Hemingway"
bookshelf search "1984" --limit 20
```

### Listing Books

```bash
bookshelf list                        # All books
bookshelf list --status want-to-read  # Filter by status
bookshelf list --status reading
bookshelf list --status finished
```

### Viewing Book Details

```bash
bookshelf show <id>
```

### Tracking Reading Progress

```bash
bookshelf start <id>   # Mark as currently reading (records start date)
bookshelf finish <id>  # Mark as finished (records finish date)
```

### Rating and Reviewing

```bash
bookshelf rate <id> <1-5>  # Rate from 1 to 5 stars
bookshelf review <id>      # Opens $EDITOR to write/edit your review
```

### Statistics

View your reading stats:

```bash
bookshelf stats
```

### Publishing to the Web

Generate a static website from your bookshelf:

```bash
bookshelf publish                  # Output to ./public
bookshelf publish --output ./site  # Custom output directory
```

## Development

If you have `just` installed, run `just` to see available commands:

```bash
just build       # Build the CLI
just test        # Run tests
just coverage    # Run tests with coverage report
just fmt         # Format code
just lint        # Run linter (requires golangci-lint)
```

### Database

The database is stored at `~/.bookshelf/bookshelf.db`. Useful commands:

```bash
just db          # Open database with sqlite3
just reset-db    # Delete and reinitialize database
just db-path     # Show database location
```

## Deploying to GitHub Pages

This project includes a GitHub Actions workflow that automatically deploys your bookshelf to GitHub Pages.

### Setup

1. Enable GitHub Pages in your repository:
   - Go to **Settings** > **Pages**
   - Set **Source** to **GitHub Actions**

2. Export your local database to the repo:
   ```bash
   just export-db
   # or manually:
   cp ~/.bookshelf/bookshelf.db ./bookshelf.db
   ```

3. Commit and push:
   ```bash
   git add bookshelf.db
   git commit -m "Update bookshelf"
   git push
   ```

The workflow will automatically build and deploy your bookshelf website.

### Manual Deployment

You can also trigger deployment manually from the Actions tab on GitHub.

### Local Preview

Preview the generated site locally:

```bash
just serve  # Generates site and serves at http://localhost:8000
```
