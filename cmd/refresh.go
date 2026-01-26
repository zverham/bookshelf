package cmd

import (
	"bookshelf/internal/api"
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh [id]",
	Short: "Refresh book metadata from Open Library",
	Long: `Fetch updated metadata (description, genres) from Open Library for books.

If an ID is provided, refreshes only that book.
If no ID is provided, refreshes all books that have an Open Library key.`,
	RunE: runRefresh,
}

func runRefresh(cmd *cobra.Command, args []string) error {
	client := api.NewClient()

	if len(args) > 0 {
		// Refresh single book
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid book ID: %s", args[0])
		}
		return refreshBook(client, id)
	}

	// Refresh all books
	return refreshAllBooks(client)
}

func refreshBook(client *api.Client, id int64) error {
	book, err := db.GetBook(id)
	if err != nil {
		return fmt.Errorf("book not found: %w", err)
	}

	if !book.Book.OpenLibraryKey.Valid || book.Book.OpenLibraryKey.String == "" {
		return fmt.Errorf("book %d has no Open Library key, cannot refresh", id)
	}

	updated, err := fetchAndUpdateMetadata(client, book)
	if err != nil {
		return err
	}

	if updated {
		fmt.Printf("Refreshed \"%s\"\n", book.Book.Title)
	} else {
		fmt.Printf("No new metadata for \"%s\"\n", book.Book.Title)
	}
	return nil
}

func refreshAllBooks(client *api.Client) error {
	books, err := db.GetBooksWithOpenLibraryKey()
	if err != nil {
		return fmt.Errorf("failed to get books: %w", err)
	}

	if len(books) == 0 {
		fmt.Println("No books with Open Library keys found.")
		return nil
	}

	fmt.Printf("Refreshing %d books...\n\n", len(books))

	var refreshed, skipped int
	for _, book := range books {
		updated, err := fetchAndUpdateMetadata(client, &book)
		if err != nil {
			fmt.Printf("  [!] %s: %v\n", book.Book.Title, err)
			skipped++
			continue
		}

		if updated {
			fmt.Printf("  [+] %s\n", book.Book.Title)
			refreshed++
		} else {
			fmt.Printf("  [-] %s (no new data)\n", book.Book.Title)
			skipped++
		}
	}

	fmt.Printf("\nDone: %d refreshed, %d unchanged\n", refreshed, skipped)
	return nil
}

func fetchAndUpdateMetadata(client *api.Client, book *models.BookWithEntry) (bool, error) {
	work, err := client.GetWorkDetails(book.Book.OpenLibraryKey.String)
	if err != nil {
		return false, fmt.Errorf("failed to fetch from Open Library: %w", err)
	}

	var description, genres *string
	var updated bool

	// Update description if missing
	if !book.Book.Description.Valid || book.Book.Description.String == "" {
		if desc := work.DescriptionText(); desc != nil {
			description = desc
			updated = true
		}
	}

	// Update genres if missing
	if !book.Book.Genres.Valid || book.Book.Genres.String == "" {
		if subjects := work.TopSubjects(5); len(subjects) > 0 {
			if jsonBytes, err := json.Marshal(subjects); err == nil {
				jsonStr := string(jsonBytes)
				genres = &jsonStr
				updated = true
			}
		}
	}

	if updated {
		if err := db.UpdateBookMetadata(book.Book.ID, description, genres); err != nil {
			return false, fmt.Errorf("failed to update database: %w", err)
		}
	}

	return updated, nil
}
