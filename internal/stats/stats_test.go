package stats

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"bookshelf/internal/testutil"
	"strings"
	"testing"
)

func TestPrintStatsEmpty(t *testing.T) {
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	output := testutil.CaptureOutput(t, func() {
		err := PrintStats()
		if err != nil {
			t.Fatalf("PrintStats failed: %v", err)
		}
	})

	expectedStrings := []string{
		"Reading Statistics",
		"Total books:    0",
		"Want to read:   0",
		"Reading:        0",
		"Finished:       0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("output missing: %s", expected)
		}
	}
}

func TestPrintStatsWithBooks(t *testing.T) {
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	// Add books
	id1, _ := db.AddBook("Book 1", "Author 1", nil, nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id1, models.StatusWantToRead)

	id2, _ := db.AddBook("Book 2", "Author 2", nil, nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id2, models.StatusReading)

	id3, _ := db.AddBook("Book 3", "Author 3", nil, nil, nil, nil, nil, nil)
	db.CreateReadingEntry(id3, models.StatusFinished)
	db.UpdateRating(id3, 4)

	output := testutil.CaptureOutput(t, func() {
		err := PrintStats()
		if err != nil {
			t.Fatalf("PrintStats failed: %v", err)
		}
	})

	expectedStrings := []string{
		"Total books:    3",
		"Want to read:   1",
		"Reading:        1",
		"Finished:       1",
		"Average rating: 4.0/5",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("output missing: %s", expected)
		}
	}
}

func TestRenderStars(t *testing.T) {
	tests := []struct {
		rating   float64
		expected string
	}{
		{0.0, "....."},
		{1.0, "*...."},
		{2.5, "**~.."},
		{3.0, "***.."},
		{4.5, "****~"},
		{5.0, "*****"},
	}

	for _, tt := range tests {
		result := renderStars(tt.rating)
		if result != tt.expected {
			t.Errorf("renderStars(%.1f) = %s, expected %s", tt.rating, result, tt.expected)
		}
	}
}

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		current  int
		total    int
		width    int
		expected string
	}{
		{0, 10, 10, "[----------]"},
		{5, 10, 10, "[#####-----]"},
		{10, 10, 10, "[##########]"},
		{15, 10, 10, "[##########]"}, // Over 100%
		{0, 0, 10, "[----------]"},  // Edge case: zero total
		{8, 24, 20, "[######--------------]"},
	}

	for _, tt := range tests {
		result := RenderProgressBar(tt.current, tt.total, tt.width)
		if result != tt.expected {
			t.Errorf("RenderProgressBar(%d, %d, %d) = %s, expected %s", tt.current, tt.total, tt.width, result, tt.expected)
		}
	}
}

func TestPrintStatsWithGoal(t *testing.T) {
	cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	// Set a goal for the current year
	// Note: we can't easily test current year goal display without mocking time
	// So we just test that the function doesn't error with a goal set
	db.SetGoal(2026, 24)

	output := testutil.CaptureOutput(t, func() {
		err := PrintStats()
		if err != nil {
			t.Fatalf("PrintStats failed: %v", err)
		}
	})

	// Stats should still display (goal display depends on current year matching)
	if !strings.Contains(output, "Reading Statistics") {
		t.Error("expected 'Reading Statistics' in output")
	}
}
