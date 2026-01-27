package stats

import (
	"bookshelf/internal/db"
	"fmt"
	"strings"
	"time"
)

func PrintStats() error {
	stats, err := db.GetStats()
	if err != nil {
		return err
	}

	fmt.Println("=== Reading Statistics ===")
	fmt.Println()

	fmt.Println("Library Overview:")
	fmt.Printf("  Total books:    %d\n", stats.TotalBooks)
	fmt.Printf("  Want to read:   %d\n", stats.WantToRead)
	fmt.Printf("  Reading:        %d\n", stats.Reading)
	fmt.Printf("  Finished:       %d\n", stats.Finished)
	fmt.Println()

	fmt.Println("This Year:")
	fmt.Printf("  Books finished: %d\n", stats.BooksThisYear)

	// Show goal progress if a goal is set for current year
	currentYear := time.Now().Year()
	goal, err := db.GetGoal(currentYear)
	if err == nil && goal != nil {
		percentage := float64(stats.BooksThisYear) / float64(goal.Target) * 100
		if percentage > 100 {
			percentage = 100
		}
		fmt.Printf("  Goal progress:  %d/%d (%.0f%%)\n", stats.BooksThisYear, goal.Target, percentage)
		fmt.Printf("  %s\n", RenderProgressBar(stats.BooksThisYear, goal.Target, 20))
	}

	fmt.Printf("  Pages read:     %d\n", stats.PagesThisYear)
	fmt.Println()

	if stats.RatedBooksCount > 0 {
		fmt.Println("Ratings:")
		fmt.Printf("  Average rating: %.1f/5 (%d books rated)\n", stats.AverageRating, stats.RatedBooksCount)
		fmt.Printf("  Stars:          %s\n", renderStars(stats.AverageRating))
	}

	return nil
}

func renderStars(rating float64) string {
	fullStars := int(rating)
	halfStar := rating-float64(fullStars) >= 0.5

	var stars strings.Builder
	for i := 0; i < fullStars; i++ {
		stars.WriteString("*")
	}
	if halfStar {
		stars.WriteString("~")
	}
	for i := stars.Len(); i < 5; i++ {
		stars.WriteString(".")
	}
	return stars.String()
}

// RenderProgressBar renders a text-based progress bar.
func RenderProgressBar(current, total, width int) string {
	if total <= 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}

	filled := (current * width) / total
	if filled > width {
		filled = width
	}

	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}
