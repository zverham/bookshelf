package stats

import (
	"bookshelf/internal/db"
	"fmt"
	"strings"
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
