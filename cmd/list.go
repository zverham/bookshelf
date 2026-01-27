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
var listSearch string
var listSort string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all books",
	Long:  `List all books in your collection. Use --status to filter by reading status, --search to find books, and --sort to change ordering.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (want-to-read, reading, finished)")
	listCmd.Flags().StringVarP(&listSearch, "search", "q", "", "Search by title or author")
	listCmd.Flags().StringVarP(&listSort, "sort", "o", "added", "Sort by: added, title, author, rating")
}

func runList(cmd *cobra.Command, args []string) error {
	opts := models.ListOptions{}

	if listStatus != "" {
		status := models.BookStatus(listStatus)
		if status != models.StatusWantToRead && status != models.StatusReading && status != models.StatusFinished {
			return fmt.Errorf("invalid status: %s (use: want-to-read, reading, finished)", listStatus)
		}
		opts.StatusFilter = &status
	}

	opts.SearchQuery = listSearch

	switch listSort {
	case "added", "":
		opts.SortBy = models.SortByAdded
	case "title":
		opts.SortBy = models.SortByTitle
	case "author":
		opts.SortBy = models.SortByAuthor
	case "rating":
		opts.SortBy = models.SortByRating
	default:
		return fmt.Errorf("invalid sort option: %s (use: added, title, author, rating)", listSort)
	}

	books, err := db.ListBooks(opts)
	if err != nil {
		return fmt.Errorf("failed to list books: %w", err)
	}

	if len(books) == 0 {
		if listSearch != "" {
			fmt.Printf("No books found matching '%s'.\n", listSearch)
		} else {
			fmt.Println("No books found. Use 'bookshelf add' to add some books.")
		}
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
