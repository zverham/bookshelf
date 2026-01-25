package cmd

import (
	"bookshelf/internal/db"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var rateCmd = &cobra.Command{
	Use:   "rate [id] [1-5]",
	Short: "Rate a book",
	Long:  `Give a book a rating from 1 to 5 stars.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runRate,
}

func runRate(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", args[0])
	}

	rating, err := strconv.Atoi(args[1])
	if err != nil || rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	exists, err := db.BookExists(id)
	if err != nil {
		return fmt.Errorf("failed to check book: %w", err)
	}
	if !exists {
		return fmt.Errorf("book with ID %d not found", id)
	}

	if err := db.UpdateRating(id, rating); err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	book, _ := db.GetBook(id)
	stars := ""
	for i := 0; i < rating; i++ {
		stars += "*"
	}
	fmt.Printf("Rated \"%s\" %s (%d/5)\n", book.Book.Title, stars, rating)
	return nil
}
