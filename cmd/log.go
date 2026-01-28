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

var logRating int

var logCmd = &cobra.Command{
	Use:   "log [title]",
	Short: "Log a book you've already read",
	Long: `Search for a book and add it as already finished. Use this for books
you read in the past where you don't know the exact start/finish dates.

Optionally provide a rating with --rating.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLog,
}

func init() {
	logCmd.Flags().IntVarP(&logRating, "rating", "r", 0, "Rating from 1-5 stars")
}

func runLog(cmd *cobra.Command, args []string) error {
	if logRating != 0 && (logRating < 1 || logRating > 5) {
		return fmt.Errorf("rating must be between 1 and 5")
	}

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

	// Create entry as finished with no date
	if err := db.CreateReadingEntry(bookID, models.StatusFinished); err != nil {
		return fmt.Errorf("failed to create reading entry: %w", err)
	}

	// Apply rating if provided
	if logRating > 0 {
		if err := db.UpdateRating(bookID, logRating); err != nil {
			return fmt.Errorf("failed to set rating: %w", err)
		}
	}

	fmt.Printf("\nLogged \"%s\" by %s as read (ID: %d)\n", selected.Title, selected.Author(), bookID)
	if logRating > 0 {
		fmt.Printf("Rated: %s\n", strings.Repeat("*", logRating))
	} else {
		fmt.Printf("Don't forget to rate it with: bookshelf rate %d <1-5>\n", bookID)
	}
	return nil
}
