package cmd

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listStatus string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all books",
	Long:  `List all books in your collection. Use --status to filter by reading status.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (want-to-read, reading, finished)")
}

func runList(cmd *cobra.Command, args []string) error {
	var statusFilter *models.BookStatus
	if listStatus != "" {
		status := models.BookStatus(listStatus)
		if status != models.StatusWantToRead && status != models.StatusReading && status != models.StatusFinished {
			return fmt.Errorf("invalid status: %s (use: want-to-read, reading, finished)", listStatus)
		}
		statusFilter = &status
	}

	books, err := db.ListBooks(statusFilter)
	if err != nil {
		return fmt.Errorf("failed to list books: %w", err)
	}

	if len(books) == 0 {
		fmt.Println("No books found. Use 'bookshelf add' to add some books.")
		return nil
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header("ID", "Title", "Author", "Status", "Rating")

	for _, book := range books {
		rating := "-"
		if book.ReadingEntry.Rating.Valid {
			rating = fmt.Sprintf("%d/5", book.ReadingEntry.Rating.Int64)
		}

		title := book.Book.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		table.Append(
			strconv.FormatInt(book.Book.ID, 10),
			title,
			book.Book.Author,
			string(book.ReadingEntry.Status),
			rating,
		)
	}

	table.Render()
	return nil
}
