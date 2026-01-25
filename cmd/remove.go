package cmd

import (
	"bookshelf/internal/db"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var forceRemove bool

var removeCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove a book from your shelf",
	Long:  `Remove a book and its reading entry from your bookshelf.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Skip confirmation prompt")
}

func runRemove(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", args[0])
	}

	book, err := db.GetBook(id)
	if err != nil {
		return fmt.Errorf("book not found: %w", err)
	}

	if !forceRemove {
		fmt.Printf("Remove \"%s\" by %s? [y/N] ", book.Book.Title, book.Book.Author)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := db.DeleteBook(id); err != nil {
		return fmt.Errorf("failed to remove book: %w", err)
	}

	fmt.Printf("Removed \"%s\"\n", book.Book.Title)
	return nil
}
