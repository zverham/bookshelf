package cmd

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
	Use:   "finish [id]",
	Short: "Mark a book as finished",
	Long:  `Finish reading a book. This will set its status to "finished" and record the finish date.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFinish,
}

func runFinish(cmd *cobra.Command, args []string) error {
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

	if err := db.UpdateStatus(id, models.StatusFinished); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	book, _ := db.GetBook(id)
	fmt.Printf("Finished reading \"%s\"\n", book.Book.Title)
	fmt.Println("Don't forget to rate it with: bookshelf rate", id, "<1-5>")
	return nil
}
