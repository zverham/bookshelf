package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string
var testDir string

func TestMain(m *testing.M) {
	// Create temp directory for binary
	var err error
	testDir, err = os.MkdirTemp("", "bookshelf-test-bin-*")
	if err != nil {
		os.Exit(1)
	}

	binaryPath = filepath.Join(testDir, "bookshelf")

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "..")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		os.Stderr.Write(output)
		os.RemoveAll(testDir)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(testDir)
	os.Exit(code)
}

// runCLI executes the CLI with a test database
func runCLI(t *testing.T, dbPath string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "BOOKSHELF_DB_PATH="+dbPath)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// createTestDB creates a temporary database path
func createTestDB(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "bookshelf-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	return dbPath, func() { os.RemoveAll(tmpDir) }
}

func TestListEmpty(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	output, err := runCLI(t, dbPath, "list")
	if err != nil {
		t.Fatalf("list command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "No books found") {
		t.Errorf("expected 'No books found', got: %s", output)
	}
}

func TestListWithStatusFilter(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Test invalid status
	output, err := runCLI(t, dbPath, "list", "--status", "invalid")
	if err == nil {
		t.Error("expected error for invalid status")
	}
	if !strings.Contains(output, "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %s", output)
	}

	// Test valid statuses (should work even with empty DB)
	for _, status := range []string{"want-to-read", "reading", "finished"} {
		output, err := runCLI(t, dbPath, "list", "--status", status)
		if err != nil {
			t.Errorf("list --status %s failed: %v\nOutput: %s", status, err, output)
		}
	}
}

func TestShowInvalidID(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{"non-numeric id", []string{"show", "abc"}, "invalid book ID"},
		{"not found", []string{"show", "999"}, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, dbPath, tt.args...)
			if err == nil {
				t.Errorf("expected error, got success")
			}
			if !strings.Contains(output, tt.errMsg) {
				t.Errorf("expected '%s' in output, got: %s", tt.errMsg, output)
			}
		})
	}
}

func TestStartInvalidID(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{"non-numeric id", []string{"start", "abc"}, "invalid book ID"},
		{"not found", []string{"start", "999"}, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, dbPath, tt.args...)
			if err == nil {
				t.Errorf("expected error, got success")
			}
			if !strings.Contains(output, tt.errMsg) {
				t.Errorf("expected '%s' in output, got: %s", tt.errMsg, output)
			}
		})
	}
}

func TestFinishInvalidID(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{"non-numeric id", []string{"finish", "abc"}, "invalid book ID"},
		{"not found", []string{"finish", "999"}, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, dbPath, tt.args...)
			if err == nil {
				t.Errorf("expected error, got success")
			}
			if !strings.Contains(output, tt.errMsg) {
				t.Errorf("expected '%s' in output, got: %s", tt.errMsg, output)
			}
		})
	}
}

func TestRateValidation(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{"invalid id", []string{"rate", "abc", "5"}, "invalid book ID"},
		{"rating too low", []string{"rate", "1", "0"}, "between 1 and 5"},
		{"rating too high", []string{"rate", "1", "6"}, "between 1 and 5"},
		{"non-numeric rating", []string{"rate", "1", "abc"}, "between 1 and 5"},
		{"not found", []string{"rate", "999", "5"}, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, dbPath, tt.args...)
			if err == nil {
				t.Errorf("expected error, got success")
			}
			if !strings.Contains(output, tt.errMsg) {
				t.Errorf("expected '%s' in output, got: %s", tt.errMsg, output)
			}
		})
	}
}

func TestRemoveInvalidID(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{"invalid id", []string{"remove", "abc", "-f"}, "invalid book ID"},
		{"not found", []string{"remove", "999", "-f"}, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, dbPath, tt.args...)
			if err == nil {
				t.Errorf("expected error, got success")
			}
			if !strings.Contains(output, tt.errMsg) {
				t.Errorf("expected '%s' in output, got: %s", tt.errMsg, output)
			}
		})
	}
}

func TestStatsEmpty(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	output, err := runCLI(t, dbPath, "stats")
	if err != nil {
		t.Fatalf("stats command failed: %v\nOutput: %s", err, output)
	}

	expectedStrings := []string{
		"Reading Statistics",
		"Total books:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected '%s' in output, got: %s", expected, output)
		}
	}
}

func TestPublish(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	outputDir, err := os.MkdirTemp("", "bookshelf-publish-test-*")
	if err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	output, err := runCLI(t, dbPath, "publish", "--output", outputDir)
	if err != nil {
		t.Fatalf("publish command failed: %v\nOutput: %s", err, output)
	}

	// Verify files were created
	expectedFiles := []string{"index.html", "style.css"}
	for _, file := range expectedFiles {
		path := filepath.Join(outputDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to be created", file)
		}
	}
}

func TestHelpCommands(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Test that help works for various commands
	commands := []string{"list", "show", "start", "finish", "rate", "review", "stats", "publish", "remove", "search", "add", "goal", "config"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			output, err := runCLI(t, dbPath, cmd, "--help")
			if err != nil {
				t.Errorf("%s --help failed: %v\nOutput: %s", cmd, err, output)
			}
			if !strings.Contains(output, "Usage:") {
				t.Errorf("expected help output for %s, got: %s", cmd, output)
			}
		})
	}
}

func TestGoalCommands(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Set a goal
	output, err := runCLI(t, dbPath, "goal", "set", "2026", "24")
	if err != nil {
		t.Fatalf("goal set failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Set reading goal for 2026: 24 books") {
		t.Errorf("unexpected output: %s", output)
	}

	// Show specific year
	output, err = runCLI(t, dbPath, "goal", "show", "2026")
	if err != nil {
		t.Fatalf("goal show failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "2026") && !strings.Contains(output, "24") {
		t.Errorf("expected year and target in output, got: %s", output)
	}

	// Show all goals
	output, err = runCLI(t, dbPath, "goal", "show")
	if err != nil {
		t.Fatalf("goal show all failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "2026") {
		t.Errorf("expected 2026 in output, got: %s", output)
	}

	// Update existing goal
	output, err = runCLI(t, dbPath, "goal", "set", "2026", "30")
	if err != nil {
		t.Fatalf("goal update failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "30 books") {
		t.Errorf("expected updated target in output, got: %s", output)
	}

	// Clear goal
	output, err = runCLI(t, dbPath, "goal", "clear", "2026")
	if err != nil {
		t.Fatalf("goal clear failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Cleared") {
		t.Errorf("expected 'Cleared' in output, got: %s", output)
	}

	// Show non-existent goal
	output, err = runCLI(t, dbPath, "goal", "show", "2026")
	if err != nil {
		t.Fatalf("goal show failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "No goal set") {
		t.Errorf("expected 'No goal set' in output, got: %s", output)
	}
}

func TestGoalValidation(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{"invalid year", []string{"goal", "set", "abc", "24"}, "invalid year"},
		{"year too low", []string{"goal", "set", "1800", "24"}, "invalid year"},
		{"year too high", []string{"goal", "set", "2200", "24"}, "invalid year"},
		{"invalid target", []string{"goal", "set", "2026", "abc"}, "invalid target"},
		{"target zero", []string{"goal", "set", "2026", "0"}, "invalid target"},
		{"clear non-existent", []string{"goal", "clear", "2099"}, "no goal found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(t, dbPath, tt.args...)
			if err == nil {
				t.Errorf("expected error, got success")
			}
			if !strings.Contains(output, tt.errMsg) {
				t.Errorf("expected '%s' in output, got: %s", tt.errMsg, output)
			}
		})
	}
}

func TestListWithSearch(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// First, we need to add some books (can't use add command because it requires interactive search)
	// Instead, test that the search flag is accepted
	output, err := runCLI(t, dbPath, "list", "--search", "test")
	if err != nil {
		t.Fatalf("list with search failed: %v\nOutput: %s", err, output)
	}
	// With empty DB, should show "No books found matching"
	if !strings.Contains(output, "No books found") {
		t.Errorf("expected 'No books found' message, got: %s", output)
	}

	// Test short flag
	output, err = runCLI(t, dbPath, "list", "-q", "author name")
	if err != nil {
		t.Fatalf("list with -q failed: %v\nOutput: %s", err, output)
	}
}

func TestListWithSort(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Test all sort options are accepted
	sortOptions := []string{"added", "title", "author", "rating"}
	for _, opt := range sortOptions {
		output, err := runCLI(t, dbPath, "list", "--sort", opt)
		if err != nil {
			t.Errorf("list with --sort %s failed: %v\nOutput: %s", opt, err, output)
		}
	}

	// Test short flag
	output, err := runCLI(t, dbPath, "list", "-o", "title")
	if err != nil {
		t.Errorf("list with -o failed: %v\nOutput: %s", err, output)
	}

	// Test invalid sort option
	output, err = runCLI(t, dbPath, "list", "--sort", "invalid")
	if err == nil {
		t.Error("expected error for invalid sort option")
	}
	if !strings.Contains(output, "invalid sort option") {
		t.Errorf("expected 'invalid sort option' in error, got: %s", output)
	}
}

func TestListCombinedFilters(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Test combining search, status, and sort
	output, err := runCLI(t, dbPath, "list", "-q", "test", "-s", "finished", "-o", "rating")
	if err != nil {
		t.Fatalf("list with combined filters failed: %v\nOutput: %s", err, output)
	}
}

func TestConfigCommands(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Test config keys
	output, err := runCLI(t, dbPath, "config", "keys")
	if err != nil {
		t.Fatalf("config keys failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "site.title") {
		t.Errorf("expected 'site.title' in keys output, got: %s", output)
	}

	// Test config set
	output, err = runCLI(t, dbPath, "config", "set", "site.title", "Test Bookshelf")
	if err != nil {
		t.Fatalf("config set failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Set site.title") {
		t.Errorf("expected confirmation message, got: %s", output)
	}

	// Test config get
	output, err = runCLI(t, dbPath, "config", "get", "site.title")
	if err != nil {
		t.Fatalf("config get failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Test Bookshelf") {
		t.Errorf("expected 'Test Bookshelf' in output, got: %s", output)
	}

	// Test config list
	output, err = runCLI(t, dbPath, "config", "list")
	if err != nil {
		t.Fatalf("config list failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "site.title") {
		t.Errorf("expected 'site.title' in list output, got: %s", output)
	}

	// Test config unset
	output, err = runCLI(t, dbPath, "config", "unset", "site.title")
	if err != nil {
		t.Fatalf("config unset failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Unset") {
		t.Errorf("expected 'Unset' in output, got: %s", output)
	}
}

func TestConfigValidation(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	// Test invalid key
	output, err := runCLI(t, dbPath, "config", "set", "invalid.key", "value")
	if err == nil {
		t.Error("expected error for invalid config key")
	}
	if !strings.Contains(output, "unknown config key") {
		t.Errorf("expected 'unknown config key' error, got: %s", output)
	}

	// Test unset non-existent key
	output, err = runCLI(t, dbPath, "config", "unset", "site.title")
	if err == nil {
		t.Error("expected error for unset non-existent key")
	}
}
