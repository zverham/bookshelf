package cmd

import (
	"bookshelf/internal/publish"

	"github.com/spf13/cobra"
)

var outputDir string

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Generate a static website",
	Long:  `Generate a static HTML website from your bookshelf data.`,
	RunE:  runPublish,
}

func init() {
	publishCmd.Flags().StringVarP(&outputDir, "output", "o", "./public", "Output directory for the static site")
}

func runPublish(cmd *cobra.Command, args []string) error {
	return publish.Generate(outputDir)
}
