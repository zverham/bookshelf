package cmd

import (
	"bookshelf/internal/api"
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Search and add a book to your shelf",
	Long:  `Search for a book by title and add it to your reading list.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	client := api.NewClient()

	fmt.Printf("Searching for \"%s\"...\n\n", query)

	docs, err := client.Search(query, 5)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(docs) == 0 {
		fmt.Println("No books found.")
		return nil
	}

	fmt.Println("Search results:")
	for i, doc := range docs {
		year := ""
		if doc.FirstPublishYear > 0 {
			year = fmt.Sprintf(" (%d)", doc.FirstPublishYear)
		}
		fmt.Printf("  %d. %s by %s%s\n", i+1, doc.Title, doc.Author(), year)
	}

	fmt.Print("\nSelect a book (1-5) or 0 to cancel: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 0 || choice > len(docs) {
		fmt.Println("Invalid selection.")
		return nil
	}

	if choice == 0 {
		fmt.Println("Cancelled.")
		return nil
	}

	selected := docs[choice-1]

	// Fetch additional details if available
	var description *string
	var genres *string
	if selected.Key != "" {
		work, err := client.GetWorkDetails(selected.Key)
		if err == nil {
			description = work.DescriptionText()
			// Get top 5 subjects as genres
			if subjects := work.TopSubjects(5); len(subjects) > 0 {
				if jsonBytes, err := json.Marshal(subjects); err == nil {
					jsonStr := string(jsonBytes)
					genres = &jsonStr
				}
			}
		}
	}

	bookID, err := db.AddBook(
		selected.Title,
		selected.Author(),
		selected.FirstISBN(),
		selected.CoverURL(),
		description,
		&selected.Key,
		genres,
		selected.Pages(),
	)
	if err != nil {
		return fmt.Errorf("failed to add book: %w", err)
	}

	if err := db.CreateReadingEntry(bookID, models.StatusWantToRead); err != nil {
		return fmt.Errorf("failed to create reading entry: %w", err)
	}

	fmt.Printf("\nAdded \"%s\" by %s (ID: %d)\n", selected.Title, selected.Author(), bookID)
	return nil
}
