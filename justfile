# Bookshelf CLI - Development Commands

# Default command - show available recipes
default:
    @just --list

# Build the CLI
build:
    go build -o bookshelf .

# Run go mod tidy
tidy:
    go mod tidy

# Build and install to $GOPATH/bin
install:
    go install .

# Run the CLI with arguments
run *args:
    go run . {{args}}

# Search for a book
search query:
    go run . search "{{query}}"

# Add a book interactively
add query:
    go run . add "{{query}}"

# List all books
list:
    go run . list

# List books by status (want-to-read, reading, finished)
list-status status:
    go run . list --status {{status}}

# Show book details
show id:
    go run . show {{id}}

# Start reading a book
start id:
    go run . start {{id}}

# Finish reading a book
finish id:
    go run . finish {{id}}

# Rate a book (1-5)
rate id rating:
    go run . rate {{id}} {{rating}}

# Add/edit a review
review id:
    go run . review {{id}}

# Show reading statistics
stats:
    go run . stats

# Run tests
test:
    go test ./...

# Run tests with verbose output
test-v:
    go test -v ./...

# Run tests with coverage
test-cover:
    go test -cover ./...

# Run tests and generate coverage report
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run tests and open coverage in browser
coverage-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"
    open coverage.html || xdg-open coverage.html || echo "Open coverage.html in your browser"

# Run tests with coverage for specific package
coverage-pkg pkg:
    go test -coverprofile=coverage.out ./internal/{{pkg}}/...
    go tool cover -func=coverage.out

# Format code
fmt:
    go fmt ./...

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -f bookshelf
    go clean

# Reset database (delete and reinitialize)
reset-db:
    rm -f ~/.bookshelf/bookshelf.db
    @echo "Database reset. Run any command to reinitialize."

# Show database location
db-path:
    @echo ~/.bookshelf/bookshelf.db

# Open database with sqlite3
db:
    sqlite3 ~/.bookshelf/bookshelf.db

# Show all books in database (raw SQL)
db-books:
    sqlite3 ~/.bookshelf/bookshelf.db "SELECT * FROM books"

# Show all reading entries (raw SQL)
db-entries:
    sqlite3 ~/.bookshelf/bookshelf.db "SELECT * FROM reading_entries"

# Generate static website to ./public
publish:
    go run . publish

# Generate static website to custom directory
publish-to dir:
    go run . publish --output {{dir}}

# Generate and serve locally (requires python3)
serve: publish
    @echo "Serving at http://localhost:8000"
    cd public && python3 -m http.server 8000

# Clean generated site
clean-site:
    rm -rf public

# Export database to repo for GitHub Pages deployment
export-db:
    cp ~/.bookshelf/bookshelf.db ./bookshelf.db
    @echo "Database exported to ./bookshelf.db"
    @echo "Commit this file to deploy your bookshelf to GitHub Pages"

# Full deploy: export db, publish locally, and show git status
deploy-prep: export-db publish
    @echo ""
    @echo "Ready to deploy! Run:"
    @echo "  git add bookshelf.db && git commit -m 'Update bookshelf' && git push"
