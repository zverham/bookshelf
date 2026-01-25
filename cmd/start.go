package cmd

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [id]",
	Short: "Mark a book as currently reading",
	Long:  `Start reading a book. This will set its status to "reading" and record the start date.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", args[0])
	}

	exists, err := db.BookExists(id)
	if err != nil {
		return fmt.Errorf("failed to check book: %w", err)
	}
	if !exists {
		return fmt.Errorf("book with ID %d not found", id)
	}

	if err := db.UpdateStatus(id, models.StatusReading); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	book, _ := db.GetBook(id)
	fmt.Printf("Started reading \"%s\"\n", book.Book.Title)
	return nil
}
