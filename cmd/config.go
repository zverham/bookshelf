package cmd

import (
	"bookshelf/internal/db"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Valid configuration keys
var validConfigKeys = map[string]string{
	"site.title":       "Website title (default: 'My Bookshelf')",
	"site.subtitle":    "Website subtitle (default: 'Personal Reading Tracker')",
	"site.author":      "Your name for attribution in footer",
	"site.description": "Meta description for SEO",
	"site.base_url":    "Base URL for canonical links (e.g., https://example.com)",
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage website configuration",
	Long:  `View and modify website configuration settings used when publishing.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. Use 'bookshelf config list' to see available keys.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE:  runConfigList,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a configuration value (revert to default)",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigUnset,
}

var configKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Show available configuration keys",
	RunE:  runConfigKeys,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configKeysCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := strings.Join(args[1:], " ")

	if _, valid := validConfigKeys[key]; !valid {
		return fmt.Errorf("unknown config key: %s\nUse 'bookshelf config keys' to see available keys", key)
	}

	if err := db.SetConfig(key, value); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	value, err := db.GetConfig(key)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if value == "" {
		fmt.Printf("%s: (not set)\n", key)
	} else {
		fmt.Printf("%s: %s\n", key, value)
	}
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	config, err := db.GetAllConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if len(config) == 0 {
		fmt.Println("No configuration values set. Using defaults.")
		fmt.Println("Use 'bookshelf config keys' to see available options.")
		return nil
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header("Key", "Value")

	// Sort keys for consistent output
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		table.Append(k, config[k])
	}

	table.Render()
	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	key := args[0]

	if err := db.DeleteConfig(key); err != nil {
		return fmt.Errorf("failed to unset config: %w", err)
	}

	fmt.Printf("Unset %s (reverted to default)\n", key)
	return nil
}

func runConfigKeys(cmd *cobra.Command, args []string) error {
	fmt.Println("Available configuration keys:")
	fmt.Println()

	// Sort keys for consistent output
	keys := make([]string, 0, len(validConfigKeys))
	for k := range validConfigKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("  %-20s %s\n", k, validConfigKeys[k])
	}
	return nil
}
