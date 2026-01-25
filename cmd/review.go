package cmd

import (
	"bookshelf/internal/db"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review [id]",
	Short: "Add or edit a review",
	Long:  `Add or edit your review for a book. Opens your default editor ($EDITOR).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runReview,
}

func runReview(cmd *cobra.Command, args []string) error {
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

	book, err := db.GetBook(id)
	if err != nil {
		return fmt.Errorf("failed to get book: %w", err)
	}

	// Get existing review
	existingReview, _ := db.GetReview(id)

	// Create temp file with existing review
	tmpFile, err := os.CreateTemp("", "bookshelf-review-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	header := fmt.Sprintf("# Review for: %s by %s\n# Lines starting with # will be ignored\n\n", book.Book.Title, book.Book.Author)
	tmpFile.WriteString(header)
	tmpFile.WriteString(existingReview)
	tmpFile.Close()

	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, tmpFile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read the edited content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read temp file: %w", err)
	}

	// Filter out comment lines
	lines := strings.Split(string(content), "\n")
	var reviewLines []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			reviewLines = append(reviewLines, line)
		}
	}
	review := strings.TrimSpace(strings.Join(reviewLines, "\n"))

	if err := db.UpdateReview(id, review); err != nil {
		return fmt.Errorf("failed to save review: %w", err)
	}

	if review == "" {
		fmt.Println("Review cleared.")
	} else {
		fmt.Println("Review saved.")
	}
	return nil
}
