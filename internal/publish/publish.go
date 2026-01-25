package publish

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SiteData struct {
	Books       []models.BookWithEntry
	Stats       *db.Stats
	GeneratedAt string
}

type BookPageData struct {
	Book        models.BookWithEntry
	GeneratedAt string
}

func Generate(outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Fetch all data
	books, err := db.ListBooks(nil)
	if err != nil {
		return fmt.Errorf("failed to fetch books: %w", err)
	}

	stats, err := db.GetStats()
	if err != nil {
		return fmt.Errorf("failed to fetch stats: %w", err)
	}

	generatedAt := time.Now().Format("January 2, 2006")

	// Generate index page
	siteData := SiteData{
		Books:       books,
		Stats:       stats,
		GeneratedAt: generatedAt,
	}

	if err := generateIndex(outputDir, siteData); err != nil {
		return err
	}

	// Generate individual book pages
	booksDir := filepath.Join(outputDir, "books")
	if err := os.MkdirAll(booksDir, 0755); err != nil {
		return fmt.Errorf("failed to create books directory: %w", err)
	}

	for _, book := range books {
		if err := generateBookPage(booksDir, book, generatedAt); err != nil {
			return err
		}
	}

	// Generate CSS
	if err := generateCSS(outputDir); err != nil {
		return err
	}

	fmt.Printf("Generated static site in %s/\n", outputDir)
	fmt.Printf("  - index.html\n")
	fmt.Printf("  - style.css\n")
	fmt.Printf("  - books/ (%d book pages)\n", len(books))

	return nil
}

func generateIndex(outputDir string, data SiteData) error {
	tmpl, err := template.New("index").Funcs(templateFuncs()).Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse index template: %w", err)
	}

	f, err := os.Create(filepath.Join(outputDir, "index.html"))
	if err != nil {
		return fmt.Errorf("failed to create index.html: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func generateBookPage(booksDir string, book models.BookWithEntry, generatedAt string) error {
	tmpl, err := template.New("book").Funcs(templateFuncs()).Parse(bookTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse book template: %w", err)
	}

	filename := filepath.Join(booksDir, fmt.Sprintf("%d.html", book.Book.ID))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create book page: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, BookPageData{Book: book, GeneratedAt: generatedAt})
}

func generateCSS(outputDir string) error {
	return os.WriteFile(filepath.Join(outputDir, "style.css"), []byte(cssStyles), 0644)
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"statusClass": func(status models.BookStatus) string {
			return strings.ReplaceAll(string(status), "-", "")
		},
		"stars": func(rating int64) string {
			var s strings.Builder
			for i := int64(0); i < rating; i++ {
				s.WriteString("★")
			}
			for i := rating; i < 5; i++ {
				s.WriteString("☆")
			}
			return s.String()
		},
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
		},
		"truncate": func(s string, max int) string {
			if len(s) <= max {
				return s
			}
			return s[:max-3] + "..."
		},
	}
}

const indexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Bookshelf</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <header>
        <h1>My Bookshelf</h1>
        <p class="subtitle">Personal Reading Tracker</p>
    </header>

    <main>
        <section class="stats">
            <div class="stat-card">
                <span class="stat-number">{{.Stats.TotalBooks}}</span>
                <span class="stat-label">Total Books</span>
            </div>
            <div class="stat-card">
                <span class="stat-number">{{.Stats.Finished}}</span>
                <span class="stat-label">Finished</span>
            </div>
            <div class="stat-card">
                <span class="stat-number">{{.Stats.Reading}}</span>
                <span class="stat-label">Reading</span>
            </div>
            <div class="stat-card">
                <span class="stat-number">{{.Stats.WantToRead}}</span>
                <span class="stat-label">Want to Read</span>
            </div>
            {{if gt .Stats.RatedBooksCount 0}}
            <div class="stat-card">
                <span class="stat-number">{{printf "%.1f" .Stats.AverageRating}}</span>
                <span class="stat-label">Avg Rating</span>
            </div>
            {{end}}
        </section>

        {{if .Books}}
        <section class="books">
            <h2>All Books</h2>
            <div class="book-grid">
                {{range .Books}}
                <article class="book-card">
                    <a href="books/{{.Book.ID}}.html">
                        {{if .Book.CoverURL.Valid}}
                        <img src="{{.Book.CoverURL.String}}" alt="{{.Book.Title}}" class="book-cover">
                        {{else}}
                        <div class="book-cover placeholder">
                            <span>{{truncate .Book.Title 30}}</span>
                        </div>
                        {{end}}
                    </a>
                    <div class="book-info">
                        <h3><a href="books/{{.Book.ID}}.html">{{.Book.Title}}</a></h3>
                        <p class="author">{{.Book.Author}}</p>
                        <span class="status {{statusClass .ReadingEntry.Status}}">{{.ReadingEntry.Status}}</span>
                        {{if .ReadingEntry.Rating.Valid}}
                        <span class="rating">{{stars .ReadingEntry.Rating.Int64}}</span>
                        {{end}}
                    </div>
                </article>
                {{end}}
            </div>
        </section>
        {{else}}
        <section class="empty">
            <p>No books yet. Start adding books with <code>bookshelf add</code></p>
        </section>
        {{end}}
    </main>

    <footer>
        <p>Generated on {{.GeneratedAt}} with <a href="https://github.com/anthropics/claude-code">Bookshelf CLI</a></p>
    </footer>
</body>
</html>`

const bookTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Book.Book.Title}} - My Bookshelf</title>
    <link rel="stylesheet" href="../style.css">
</head>
<body>
    <header>
        <h1><a href="../index.html">My Bookshelf</a></h1>
    </header>

    <main>
        <article class="book-detail">
            <div class="book-header">
                {{if .Book.Book.CoverURL.Valid}}
                <img src="{{.Book.Book.CoverURL.String}}" alt="{{.Book.Book.Title}}" class="book-cover-large">
                {{else}}
                <div class="book-cover-large placeholder">
                    <span>No Cover</span>
                </div>
                {{end}}
                <div class="book-meta">
                    <h2>{{.Book.Book.Title}}</h2>
                    <p class="author">by {{.Book.Book.Author}}</p>

                    <dl class="details">
                        <dt>Status</dt>
                        <dd><span class="status {{statusClass .Book.ReadingEntry.Status}}">{{.Book.ReadingEntry.Status}}</span></dd>

                        {{if .Book.ReadingEntry.Rating.Valid}}
                        <dt>Rating</dt>
                        <dd class="rating">{{stars .Book.ReadingEntry.Rating.Int64}}</dd>
                        {{end}}

                        {{if .Book.Book.Pages.Valid}}
                        <dt>Pages</dt>
                        <dd>{{.Book.Book.Pages.Int64}}</dd>
                        {{end}}

                        {{if .Book.Book.ISBN.Valid}}
                        <dt>ISBN</dt>
                        <dd>{{.Book.Book.ISBN.String}}</dd>
                        {{end}}

                        {{if .Book.ReadingEntry.StartedAt.Valid}}
                        <dt>Started</dt>
                        <dd>{{formatDate .Book.ReadingEntry.StartedAt.Time}}</dd>
                        {{end}}

                        {{if .Book.ReadingEntry.FinishedAt.Valid}}
                        <dt>Finished</dt>
                        <dd>{{formatDate .Book.ReadingEntry.FinishedAt.Time}}</dd>
                        {{end}}
                    </dl>
                </div>
            </div>

            {{if .Book.Book.Description.Valid}}
            <section class="description">
                <h3>Description</h3>
                <p>{{.Book.Book.Description.String}}</p>
            </section>
            {{end}}

            {{if .Book.ReadingEntry.Review.Valid}}
            <section class="review">
                <h3>My Review</h3>
                <div class="review-text">{{nl2br .Book.ReadingEntry.Review.String}}</div>
            </section>
            {{end}}

            <a href="../index.html" class="back-link">← Back to all books</a>
        </article>
    </main>

    <footer>
        <p>Generated on {{.GeneratedAt}} with <a href="https://github.com/anthropics/claude-code">Bookshelf CLI</a></p>
    </footer>
</body>
</html>`

const cssStyles = `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    line-height: 1.6;
    color: #333;
    background: #f5f5f5;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

header {
    background: #2c3e50;
    color: white;
    padding: 2rem;
    text-align: center;
}

header h1 {
    font-size: 2.5rem;
    margin-bottom: 0.5rem;
}

header h1 a {
    color: white;
    text-decoration: none;
}

header .subtitle {
    opacity: 0.8;
}

main {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    flex: 1;
    width: 100%;
}

.stats {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
    margin-bottom: 2rem;
}

.stat-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    text-align: center;
    flex: 1;
    min-width: 120px;
}

.stat-number {
    display: block;
    font-size: 2rem;
    font-weight: bold;
    color: #2c3e50;
}

.stat-label {
    font-size: 0.9rem;
    color: #666;
}

.books h2 {
    margin-bottom: 1.5rem;
    color: #2c3e50;
}

.book-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1.5rem;
}

.book-card {
    background: white;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    transition: transform 0.2s, box-shadow 0.2s;
}

.book-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.book-cover {
    width: 100%;
    height: 280px;
    object-fit: cover;
    display: block;
}

.book-cover.placeholder {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 0.9rem;
    padding: 1rem;
    text-align: center;
}

.book-info {
    padding: 1rem;
}

.book-info h3 {
    font-size: 1rem;
    margin-bottom: 0.25rem;
}

.book-info h3 a {
    color: #2c3e50;
    text-decoration: none;
}

.book-info h3 a:hover {
    color: #3498db;
}

.author {
    color: #666;
    font-size: 0.9rem;
    margin-bottom: 0.5rem;
}

.status {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
}

.status.wanttoread {
    background: #e8f4fd;
    color: #2980b9;
}

.status.reading {
    background: #fef3e2;
    color: #e67e22;
}

.status.finished {
    background: #e8f8f0;
    color: #27ae60;
}

.rating {
    color: #f39c12;
    font-size: 0.9rem;
    margin-left: 0.5rem;
}

/* Book detail page */
.book-detail {
    background: white;
    border-radius: 8px;
    padding: 2rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.book-header {
    display: flex;
    gap: 2rem;
    margin-bottom: 2rem;
}

.book-cover-large {
    width: 200px;
    height: 300px;
    object-fit: cover;
    border-radius: 4px;
    flex-shrink: 0;
}

.book-cover-large.placeholder {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
}

.book-meta h2 {
    font-size: 1.75rem;
    margin-bottom: 0.5rem;
    color: #2c3e50;
}

.book-meta .author {
    font-size: 1.1rem;
    margin-bottom: 1.5rem;
}

.details {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.5rem 1rem;
}

.details dt {
    font-weight: 600;
    color: #666;
}

.details dd {
    color: #333;
}

.details .rating {
    font-size: 1.2rem;
    margin-left: 0;
}

.description, .review {
    margin-top: 2rem;
    padding-top: 2rem;
    border-top: 1px solid #eee;
}

.description h3, .review h3 {
    margin-bottom: 1rem;
    color: #2c3e50;
}

.review-text {
    background: #f9f9f9;
    padding: 1rem;
    border-radius: 4px;
    border-left: 4px solid #3498db;
}

.back-link {
    display: inline-block;
    margin-top: 2rem;
    color: #3498db;
    text-decoration: none;
}

.back-link:hover {
    text-decoration: underline;
}

.empty {
    text-align: center;
    padding: 4rem 2rem;
    background: white;
    border-radius: 8px;
}

.empty code {
    background: #f0f0f0;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
}

footer {
    text-align: center;
    padding: 2rem;
    color: #666;
    font-size: 0.9rem;
}

footer a {
    color: #3498db;
}

@media (max-width: 600px) {
    .book-header {
        flex-direction: column;
        align-items: center;
        text-align: center;
    }

    .details {
        text-align: left;
    }

    .stats {
        justify-content: center;
    }
}
`
