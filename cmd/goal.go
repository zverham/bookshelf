package cmd

import (
	"bookshelf/internal/db"
	"bookshelf/internal/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage reading goals",
	Long:  `Set, view, and manage yearly reading goals.`,
}

var goalSetCmd = &cobra.Command{
	Use:   "set <year> <target>",
	Short: "Set a reading goal for a year",
	Long:  `Set a target number of books to read in a given year. Updates existing goal if one exists.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runGoalSet,
}

var goalShowCmd = &cobra.Command{
	Use:   "show [year]",
	Short: "Show reading goal(s) with progress",
	Long:  `Show reading goal progress. If year is specified, shows that year; otherwise shows all goals.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGoalShow,
}

var goalClearCmd = &cobra.Command{
	Use:   "clear <year>",
	Short: "Remove a reading goal",
	Long:  `Delete the reading goal for a specific year.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGoalClear,
}

func init() {
	goalCmd.AddCommand(goalSetCmd)
	goalCmd.AddCommand(goalShowCmd)
	goalCmd.AddCommand(goalClearCmd)
}

func runGoalSet(cmd *cobra.Command, args []string) error {
	year, err := strconv.Atoi(args[0])
	if err != nil || year < 1900 || year > 2100 {
		return fmt.Errorf("invalid year: %s (must be between 1900 and 2100)", args[0])
	}

	target, err := strconv.Atoi(args[1])
	if err != nil || target <= 0 {
		return fmt.Errorf("invalid target: %s (must be a positive number)", args[1])
	}

	if err := db.SetGoal(year, target); err != nil {
		return fmt.Errorf("failed to set goal: %w", err)
	}

	fmt.Printf("Set reading goal for %d: %d books\n", year, target)
	return nil
}

func runGoalShow(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		// Show specific year
		year, err := strconv.Atoi(args[0])
		if err != nil || year < 1900 || year > 2100 {
			return fmt.Errorf("invalid year: %s (must be between 1900 and 2100)", args[0])
		}

		goal, err := db.GetGoal(year)
		if err != nil {
			return fmt.Errorf("failed to get goal: %w", err)
		}
		if goal == nil {
			fmt.Printf("No goal set for %d.\n", year)
			return nil
		}

		return printGoalProgress(goal)
	}

	// Show all goals
	goals, err := db.GetAllGoals()
	if err != nil {
		return fmt.Errorf("failed to get goals: %w", err)
	}

	if len(goals) == 0 {
		fmt.Println("No reading goals set. Use 'bookshelf goal set <year> <target>' to create one.")
		return nil
	}

	for i, goal := range goals {
		if i > 0 {
			fmt.Println()
		}
		if err := printGoalProgress(&goal); err != nil {
			return err
		}
	}

	return nil
}

func runGoalClear(cmd *cobra.Command, args []string) error {
	year, err := strconv.Atoi(args[0])
	if err != nil || year < 1900 || year > 2100 {
		return fmt.Errorf("invalid year: %s (must be between 1900 and 2100)", args[0])
	}

	if err := db.ClearGoal(year); err != nil {
		return fmt.Errorf("failed to clear goal: %w", err)
	}

	fmt.Printf("Cleared reading goal for %d.\n", year)
	return nil
}

func printGoalProgress(goal *models.ReadingGoal) error {
	finished, err := db.GetBooksFinishedInYear(goal.Year)
	if err != nil {
		return fmt.Errorf("failed to get books finished: %w", err)
	}

	percentage := float64(finished) / float64(goal.Target) * 100
	if percentage > 100 {
		percentage = 100
	}

	currentYear := time.Now().Year()
	yearLabel := fmt.Sprintf("%d", goal.Year)
	if goal.Year == currentYear {
		yearLabel += " (current)"
	}

	fmt.Printf("Reading Goal %s\n", yearLabel)
	fmt.Printf("  Progress: %d/%d books (%.0f%%)\n", finished, goal.Target, percentage)
	fmt.Printf("  %s\n", renderProgressBar(finished, goal.Target, 20))

	if finished >= goal.Target {
		fmt.Printf("  Goal complete!\n")
	} else {
		remaining := goal.Target - finished
		fmt.Printf("  %d books to go\n", remaining)
	}

	return nil
}

func renderProgressBar(current, total, width int) string {
	if total <= 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}

	filled := (current * width) / total
	if filled > width {
		filled = width
	}

	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}
