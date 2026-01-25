package cmd

import (
	"bookshelf/internal/api"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search Open Library without adding",
	Long:  `Search for books on Open Library without adding them to your shelf.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	client := api.NewClient()

	fmt.Printf("Searching for \"%s\"...\n\n", query)

	docs, err := client.Search(query, searchLimit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(docs) == 0 {
		fmt.Println("No books found.")
		return nil
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header("Title", "Author", "Year", "Pages")

	for _, doc := range docs {
		year := "-"
		if doc.FirstPublishYear > 0 {
			year = fmt.Sprintf("%d", doc.FirstPublishYear)
		}
		pages := "-"
		if doc.NumberOfPages > 0 {
			pages = fmt.Sprintf("%d", doc.NumberOfPages)
		}

		title := doc.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}

		author := doc.Author()
		if len(author) > 25 {
			author = author[:22] + "..."
		}

		table.Append(title, author, year, pages)
	}

	table.Render()
	fmt.Printf("\nFound %d results. Use 'bookshelf add \"%s\"' to add a book.\n", len(docs), query)
	return nil
}
