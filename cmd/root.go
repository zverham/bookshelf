package cmd

import (
	"bookshelf/internal/db"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bookshelf",
	Short: "A personal reading tracker CLI",
	Long:  `Bookshelf is a CLI tool for tracking your personal reading history, similar to Goodreads.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return db.Init()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		db.Close()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(finishCmd)
	rootCmd.AddCommand(rateCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(goalCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(logCmd)
}
