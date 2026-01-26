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
	commands := []string{"list", "show", "start", "finish", "rate", "review", "stats", "publish", "remove", "search", "add"}

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
