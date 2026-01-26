package publish

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"encoding/json"
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
		"initials": func(title string) string {
			words := strings.Fields(title)
			if len(words) == 0 {
				return "?"
			}
			if len(words) == 1 {
				if len(words[0]) >= 2 {
					return strings.ToUpper(words[0][:2])
				}
				return strings.ToUpper(words[0][:1])
			}
			return strings.ToUpper(string(words[0][0])) + strings.ToUpper(string(words[1][0]))
		},
		"readingDays": func(started, finished time.Time) int {
			return int(finished.Sub(started).Hours()/24) + 1
		},
		"openLibraryURL": func(key string) string {
			return "https://openlibrary.org" + key
		},
		"parseGenres": func(genresJSON string) []string {
			if genresJSON == "" {
				return nil
			}
			var genres []string
			if err := json.Unmarshal([]byte(genresJSON), &genres); err != nil {
				return nil
			}
			return genres
		},
	}
}

const indexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Bookshelf</title>
    <meta name="description" content="My personal reading tracker - {{.Stats.TotalBooks}} books, {{.Stats.Finished}} finished">
    <meta property="og:title" content="My Bookshelf">
    <meta property="og:description" content="{{.Stats.TotalBooks}} books tracked, {{.Stats.Finished}} finished, {{.Stats.Reading}} currently reading">
    <meta property="og:type" content="website">
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>📚</text></svg>">
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
            {{if gt .Stats.BooksThisYear 0}}
            <div class="stat-card highlight">
                <span class="stat-number">{{.Stats.BooksThisYear}}</span>
                <span class="stat-label">This Year</span>
            </div>
            {{end}}
            {{if gt .Stats.RatedBooksCount 0}}
            <div class="stat-card">
                <span class="stat-number">{{printf "%.1f" .Stats.AverageRating}}</span>
                <span class="stat-label">Avg Rating</span>
            </div>
            {{end}}
        </section>

        {{if .Books}}
        <section class="books">
            <div class="books-header">
                <h2>All Books</h2>
                <div class="filter-tabs">
                    <button class="filter-btn active" data-filter="all">All</button>
                    <button class="filter-btn" data-filter="reading">Reading</button>
                    <button class="filter-btn" data-filter="finished">Finished</button>
                    <button class="filter-btn" data-filter="wanttoread">Want to Read</button>
                </div>
            </div>
            <div class="book-grid">
                {{range .Books}}
                <article class="book-card" data-status="{{statusClass .ReadingEntry.Status}}">
                    <a href="books/{{.Book.ID}}.html" class="book-cover-link">
                        {{if .Book.CoverURL.Valid}}
                        <img src="{{.Book.CoverURL.String}}" alt="{{.Book.Title}}" class="book-cover" loading="lazy">
                        {{else}}
                        <div class="book-cover placeholder">
                            <span class="initials">{{initials .Book.Title}}</span>
                            <span class="placeholder-title">{{truncate .Book.Title 40}}</span>
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
                        {{if .Book.Genres.Valid}}
                        <div class="genres">
                            {{range parseGenres .Book.Genres.String}}
                            <span class="genre-tag">{{.}}</span>
                            {{end}}
                        </div>
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

    <script>
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            const filter = btn.dataset.filter;
            document.querySelectorAll('.book-card').forEach(card => {
                if (filter === 'all' || card.dataset.status === filter) {
                    card.style.display = '';
                } else {
                    card.style.display = 'none';
                }
            });
        });
    });
    </script>
</body>
</html>`

const bookTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Book.Book.Title}} - My Bookshelf</title>
    <meta name="description" content="{{.Book.Book.Title}} by {{.Book.Book.Author}} - {{.Book.ReadingEntry.Status}}">
    <meta property="og:title" content="{{.Book.Book.Title}} - My Bookshelf">
    <meta property="og:description" content="{{.Book.Book.Title}} by {{.Book.Book.Author}}{{if .Book.ReadingEntry.Rating.Valid}} - Rated {{.Book.ReadingEntry.Rating.Int64}}/5{{end}}">
    <meta property="og:type" content="book">
    {{if .Book.Book.CoverURL.Valid}}<meta property="og:image" content="{{.Book.Book.CoverURL.String}}">{{end}}
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>📚</text></svg>">
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
                    <span class="initials">{{initials .Book.Book.Title}}</span>
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

                        {{if and .Book.ReadingEntry.StartedAt.Valid .Book.ReadingEntry.FinishedAt.Valid}}
                        <dt>Reading Time</dt>
                        <dd>{{readingDays .Book.ReadingEntry.StartedAt.Time .Book.ReadingEntry.FinishedAt.Time}} days</dd>
                        {{end}}

                        {{if .Book.Book.Genres.Valid}}
                        <dt>Genres</dt>
                        <dd class="genres-list">
                            {{range parseGenres .Book.Book.Genres.String}}
                            <span class="genre-tag">{{.}}</span>
                            {{end}}
                        </dd>
                        {{end}}
                    </dl>

                    {{if .Book.Book.OpenLibraryKey.Valid}}
                    <a href="{{openLibraryURL .Book.Book.OpenLibraryKey.String}}" class="external-link" target="_blank" rel="noopener">View on Open Library →</a>
                    {{end}}
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

const cssStyles = `:root {
    --bg-primary: #f5f5f5;
    --bg-card: white;
    --bg-header: #2c3e50;
    --text-primary: #333;
    --text-secondary: #666;
    --text-header: white;
    --accent: #3498db;
    --accent-hover: #2980b9;
    --border: #eee;
    --shadow: rgba(0,0,0,0.1);
    --shadow-hover: rgba(0,0,0,0.15);
}

@media (prefers-color-scheme: dark) {
    :root {
        --bg-primary: #1a1a2e;
        --bg-card: #16213e;
        --bg-header: #0f3460;
        --text-primary: #eee;
        --text-secondary: #aaa;
        --text-header: #eee;
        --accent: #4dabf7;
        --accent-hover: #74c0fc;
        --border: #2a2a4a;
        --shadow: rgba(0,0,0,0.3);
        --shadow-hover: rgba(0,0,0,0.4);
    }

    .status.wanttoread {
        background: #1a3a5c;
        color: #74c0fc;
    }

    .status.reading {
        background: #3d2a1a;
        color: #ffc078;
    }

    .status.finished {
        background: #1a3d2a;
        color: #69db7c;
    }

    .review-text {
        background: #1a1a2e;
    }

    .empty code {
        background: #2a2a4a;
    }
}

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    line-height: 1.6;
    color: var(--text-primary);
    background: var(--bg-primary);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

header {
    background: var(--bg-header);
    color: var(--text-header);
    padding: 2rem;
    text-align: center;
}

header h1 {
    font-size: 2.5rem;
    margin-bottom: 0.5rem;
}

header h1 a {
    color: var(--text-header);
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
    background: var(--bg-card);
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px var(--shadow);
    text-align: center;
    flex: 1;
    min-width: 120px;
}

.stat-card.highlight {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
}

.stat-card.highlight .stat-number,
.stat-card.highlight .stat-label {
    color: white;
}

.stat-number {
    display: block;
    font-size: 2rem;
    font-weight: bold;
    color: var(--accent);
}

.stat-label {
    font-size: 0.9rem;
    color: var(--text-secondary);
}

.books-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
    gap: 1rem;
}

.books-header h2 {
    color: var(--text-primary);
    margin: 0;
}

.filter-tabs {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
}

.filter-btn {
    padding: 0.5rem 1rem;
    border: 1px solid var(--border);
    background: var(--bg-card);
    color: var(--text-secondary);
    border-radius: 20px;
    cursor: pointer;
    font-size: 0.85rem;
    transition: all 0.2s;
}

.filter-btn:hover {
    border-color: var(--accent);
    color: var(--accent);
}

.filter-btn.active {
    background: var(--accent);
    border-color: var(--accent);
    color: white;
}

.book-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1.5rem;
}

.book-card {
    background: var(--bg-card);
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 4px var(--shadow);
    transition: transform 0.2s, box-shadow 0.2s;
}

.book-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px var(--shadow-hover);
}

.book-cover-link {
    display: block;
    overflow: hidden;
}

.book-cover {
    width: 100%;
    height: 280px;
    object-fit: cover;
    display: block;
    transition: transform 0.3s;
}

.book-card:hover .book-cover {
    transform: scale(1.05);
}

.book-cover.placeholder {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: white;
    padding: 1rem;
    text-align: center;
    height: 280px;
}

.book-cover.placeholder .initials {
    font-size: 3rem;
    font-weight: bold;
    opacity: 0.9;
}

.book-cover.placeholder .placeholder-title {
    font-size: 0.85rem;
    opacity: 0.8;
    margin-top: 0.5rem;
}

.book-info {
    padding: 1rem;
}

.book-info h3 {
    font-size: 1rem;
    margin-bottom: 0.25rem;
}

.book-info h3 a {
    color: var(--text-primary);
    text-decoration: none;
}

.book-info h3 a:hover {
    color: var(--accent);
}

.author {
    color: var(--text-secondary);
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

.genres {
    margin-top: 0.5rem;
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
}

.genre-tag {
    display: inline-block;
    padding: 0.15rem 0.5rem;
    background: var(--border);
    color: var(--text-secondary);
    border-radius: 12px;
    font-size: 0.7rem;
    text-transform: lowercase;
}

.genres-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.genres-list .genre-tag {
    font-size: 0.85rem;
    padding: 0.25rem 0.75rem;
}

/* Book detail page */
.book-detail {
    background: var(--bg-card);
    border-radius: 8px;
    padding: 2rem;
    box-shadow: 0 2px 4px var(--shadow);
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
    box-shadow: 0 4px 12px var(--shadow);
}

.book-cover-large.placeholder {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: white;
}

.book-cover-large.placeholder .initials {
    font-size: 4rem;
    font-weight: bold;
}

.book-meta h2 {
    font-size: 1.75rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);
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
    color: var(--text-secondary);
}

.details dd {
    color: var(--text-primary);
}

.details .rating {
    font-size: 1.2rem;
    margin-left: 0;
}

.external-link {
    display: inline-block;
    margin-top: 1.5rem;
    color: var(--accent);
    text-decoration: none;
    font-size: 0.9rem;
}

.external-link:hover {
    text-decoration: underline;
}

.description, .review {
    margin-top: 2rem;
    padding-top: 2rem;
    border-top: 1px solid var(--border);
}

.description h3, .review h3 {
    margin-bottom: 1rem;
    color: var(--text-primary);
}

.description p {
    color: var(--text-secondary);
    line-height: 1.8;
}

.review-text {
    background: var(--bg-primary);
    padding: 1rem;
    border-radius: 4px;
    border-left: 4px solid var(--accent);
}

.back-link {
    display: inline-block;
    margin-top: 2rem;
    color: var(--accent);
    text-decoration: none;
}

.back-link:hover {
    text-decoration: underline;
}

.empty {
    text-align: center;
    padding: 4rem 2rem;
    background: var(--bg-card);
    border-radius: 8px;
}

.empty code {
    background: var(--bg-primary);
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
}

footer {
    text-align: center;
    padding: 2rem;
    color: var(--text-secondary);
    font-size: 0.9rem;
}

footer a {
    color: var(--accent);
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

    .books-header {
        flex-direction: column;
        align-items: flex-start;
    }

    .filter-tabs {
        width: 100%;
        justify-content: flex-start;
    }
}

@media print {
    header {
        background: none;
        color: black;
        padding: 1rem 0;
    }

    .filter-tabs, .back-link, .external-link, footer {
        display: none;
    }

    .book-card {
        break-inside: avoid;
        box-shadow: none;
        border: 1px solid #ddd;
    }

    .book-detail {
        box-shadow: none;
    }

    .stats {
        border-bottom: 1px solid #ddd;
        padding-bottom: 1rem;
    }

    .stat-card {
        box-shadow: none;
        border: 1px solid #ddd;
    }

    .stat-card.highlight {
        background: #f0f0f0;
        color: black;
    }

    .stat-card.highlight .stat-number,
    .stat-card.highlight .stat-label {
        color: black;
    }
}
`
