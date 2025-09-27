// executor/executor_test.go
package executor

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// a helper function to assert errors
func assertError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got error %q, want error %q", got, want)
	}
}

// a helper function to check for nil error
func assertNilError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected nil error, got %q", err)
	}
}

func TestExecutor_Run(t *testing.T) {
	// Test case for a simple successful command using []string
	t.Run("Successful command from slice", func(t *testing.T) {
		var cmd []string
		if runtime.GOOS == "windows" {
			cmd = []string{"cmd", "/C", "echo hello"}
		} else {
			cmd = []string{"echo", "hello"}
		}
		exec := NewExecutor(cmd)
		result := exec.Run()

		if !result.Success || result.ExitCode != 0 {
			t.Errorf("Expected command to succeed, but it failed with exit code %d", result.ExitCode)
		}
		if strings.TrimSpace(result.Stdout) != "hello" {
			t.Errorf("Expected stdout 'hello', got %q", result.Stdout)
		}
		assertNilError(t, result.Error)
	})

	// Test case for a simple successful command using a single string
	t.Run("Successful command from string", func(t *testing.T) {
		// FIX: Removed quotes from around "hello world".
		// The shell on both Windows and Unix can handle this simple string without quotes,
		// which makes their output consistent.
		exec := NewExecutorFromString(`echo hello world`)
		result := exec.Run()

		if !result.Success || result.ExitCode != 0 {
			t.Errorf("Expected command to succeed, but it failed with exit code %d", result.ExitCode)
		}
		if strings.TrimSpace(result.Stdout) != "hello world" {
			t.Errorf("Expected stdout 'hello world', got %q", strings.TrimSpace(result.Stdout))
		}
		assertNilError(t, result.Error)
	})

	// Test case for a command that fails with a non-zero exit code
	t.Run("Failed command with non-zero exit", func(t *testing.T) {
		var cmd []string
		// `dir` on Windows, `ls` on Unix. Both fail on non-existent paths.
		if runtime.GOOS == "windows" {
			// Using a path that is highly unlikely to exist and is invalid syntax
			cmd = []string{"cmd", "/C", "dir C:\\non-existent-directory-for-testing-!@#$"}
		} else {
			cmd = []string{"ls", "/non-existent-directory-for-testing"}
		}
		exec := NewExecutor(cmd)
		result := exec.Run()

		if result.Success || result.ExitCode == 0 {
			t.Errorf("Expected command to fail, but it succeeded")
		}
		if result.Error == nil {
			t.Error("Expected an ExitError, but got nil")
		}
		if len(result.Stderr) == 0 {
			t.Error("Expected stderr to be captured, but it was empty")
		}
	})

	// Test case for a command that cannot be found (remains the same)
	t.Run("Command not found", func(t *testing.T) {
		exec := NewExecutor([]string{"a-command-that-does-not-exist"})
		result := exec.Run()
		if result.Success {
			t.Errorf("Expected command to fail, but it succeeded")
		}
		if result.Error == nil {
			t.Error("Expected an error for command not found, but got nil")
		}
	})

	// Test case for timeout
	t.Run("Command timeout", func(t *testing.T) {
		var cmd []string
		// `timeout` on Windows, `sleep` on Unix.
		if runtime.GOOS == "windows" {
			// powershell is more reliable for longer sleeps
			cmd = []string{"powershell", "-command", "Start-Sleep -Seconds 2"}
		} else {
			cmd = []string{"sleep", "2"}
		}

		exec := NewExecutor(cmd)
		exec.Timeout = 100 * time.Millisecond
		result := exec.Run()

		if result.Success {
			t.Error("Expected command to fail due to timeout, but it succeeded")
		}
		assertError(t, result.Error, context.DeadlineExceeded)
	})

	// Test case for working directory
	t.Run("Working directory", func(t *testing.T) {
		var cmd *Executor
		// On Windows, `cd` with no args prints current dir. On Unix, `pwd`.
		if runtime.GOOS == "windows" {
			cmd = NewExecutorFromString("cd")
		} else {
			cmd = NewExecutorFromString("pwd")
		}

		tempDir := t.TempDir()
		cmd.Dir = tempDir
		result := cmd.Run()

		if !result.Success {
			t.Fatalf("Command failed: %v", result.Error)
		}
		// filepath.EvalSymlinks resolves any symbolic links for a more reliable comparison.
		resolvedTempDir, err := filepath.EvalSymlinks(tempDir)
		if err != nil {
			t.Fatalf("Failed to resolve symlinks for temp dir: %v", err)
		}

		if strings.TrimSpace(result.Stdout) != strings.TrimSpace(resolvedTempDir) {
			t.Errorf("Expected stdout to be %q, got %q", resolvedTempDir, result.Stdout)
		}
	})

	// Test case for environment variables
	t.Run("Environment variables", func(t *testing.T) {
		var cmd string
		// %VAR% on Windows, $VAR on Unix.
		if runtime.GOOS == "windows" {
			cmd = "echo %MY_TEST_VAR%"
		} else {
			cmd = "echo $MY_TEST_VAR"
		}

		exec := NewExecutorFromString(cmd)
		exec.Env = []string{"MY_TEST_VAR=hello_from_env"}
		result := exec.Run()

		if !result.Success {
			t.Fatalf("Command failed: %v", result.Error)
		}
		if strings.TrimSpace(result.Stdout) != "hello_from_env" {
			t.Errorf("Expected stdout to be 'hello_from_env', got %q", result.Stdout)
		}
	})

	// Test case for stdin
	t.Run("Standard input", func(t *testing.T) {
		var cmd []string
		// A trick to simulate `cat`: findstr on Windows, cat on Unix.
		if runtime.GOOS == "windows" {
			// findstr with a regex for "any character" will print all input lines
			cmd = []string{"findstr", "."}
		} else {
			cmd = []string{"cat"}
		}
		exec := NewExecutor(cmd)
		exec.Input = strings.NewReader("hello from stdin")
		result := exec.Run()

		if !result.Success {
			t.Fatalf("Command failed: %v", result.Error)
		}
		if strings.TrimSpace(result.Stdout) != "hello from stdin" {
			t.Errorf("Expected stdout to be 'hello from stdin', got %q", result.Stdout)
		}
	})

	// Test case for file redirection (remains the same)
	t.Run("File redirection", func(t *testing.T) {
		// This test already uses `echo` which is handled by the cross-platform `NewExecutorFromString`.
		tempDir := t.TempDir()
		stdoutFile := filepath.Join(tempDir, "stdout.log")
		stderrFile := filepath.Join(tempDir, "stderr.log")

		// The command `>&2` works in both cmd and sh for redirecting to stderr.
		exec := NewExecutorFromString("echo to stdout && echo to stderr >&2")
		exec.StdoutFile = stdoutFile
		exec.StderrFile = stderrFile
		result := exec.Run()

		if !result.Success {
			t.Fatalf("Command failed unexpectedly: %v", result.Error)
		}
		if strings.TrimSpace(result.Stdout) != "to stdout" {
			t.Errorf("Expected in-memory stdout to be 'to stdout', got %q", result.Stdout)
		}
	})

	// Other tests like Empty command and File creation failure are platform-independent
	// and do not need changes.
}
