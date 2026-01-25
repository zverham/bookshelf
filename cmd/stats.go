package cmd

import (
	"bookshelf/internal/stats"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show reading statistics",
	Long:  `Display your reading statistics including books by status, yearly progress, and average ratings.`,
	RunE:  runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
	return stats.PrintStats()
}
