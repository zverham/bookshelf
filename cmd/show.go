package cmd

import (
	"bookshelf/internal/db"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show book details",
	Long:  `Show detailed information about a book including its review.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", args[0])
	}

	book, err := db.GetBook(id)
	if err != nil {
		return fmt.Errorf("book not found: %w", err)
	}

	fmt.Printf("Title:  %s\n", book.Book.Title)
	fmt.Printf("Author: %s\n", book.Book.Author)
	fmt.Printf("Status: %s\n", book.ReadingEntry.Status)

	if book.Book.ISBN.Valid {
		fmt.Printf("ISBN:   %s\n", book.Book.ISBN.String)
	}

	if book.Book.Pages.Valid {
		fmt.Printf("Pages:  %d\n", book.Book.Pages.Int64)
	}

	if book.ReadingEntry.StartedAt.Valid {
		fmt.Printf("Started:  %s\n", book.ReadingEntry.StartedAt.Time.Format("Jan 02, 2006"))
	}

	if book.ReadingEntry.FinishedAt.Valid {
		fmt.Printf("Finished: %s\n", book.ReadingEntry.FinishedAt.Time.Format("Jan 02, 2006"))
	}

	if book.ReadingEntry.Rating.Valid {
		fmt.Printf("Rating: %d/5\n", book.ReadingEntry.Rating.Int64)
	}

	if book.Book.Description.Valid && book.Book.Description.String != "" {
		fmt.Printf("\nDescription:\n%s\n", truncateString(book.Book.Description.String, 500))
	}

	if book.ReadingEntry.Review.Valid && book.ReadingEntry.Review.String != "" {
		fmt.Printf("\nYour Review:\n%s\n", book.ReadingEntry.Review.String)
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
