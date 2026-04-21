package app

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/nalgeon/be"

	"gh-mirror/pkg/models"
)

func TestGetEnvOrDefaultWithValue(t *testing.T) {
	getenv := func(key string) string {
		if key == "MY_VAR" {
			return "my-value"
		}
		return ""
	}
	result := GetEnvOrDefault("MY_VAR", "default", getenv)
	be.Equal(t, result, "my-value")
}

func TestGetEnvOrDefaultWithDefault(t *testing.T) {
	getenv := func(key string) string { return "" }
	result := GetEnvOrDefault("MISSING_VAR", "default-value", getenv)
	be.Equal(t, result, "default-value")
}

func TestGetEnvOrDefaultEmptyString(t *testing.T) {
	getenv := func(key string) string { return "" }
	result := GetEnvOrDefault("EMPTY_VAR", "fallback", getenv)
	be.Equal(t, result, "fallback")
}

func TestPrintVersion(t *testing.T) {
	old := Version
	Version = "1.2.3"
	defer func() { Version = old }()

	result := captureOutput(PrintVersion)
	be.True(t, strings.Contains(result, "1.2.3"))
}

func TestPrintUsage(t *testing.T) {
	result := captureOutput(PrintUsage)
	be.True(t, strings.Contains(result, "mirror <command>"))
	be.True(t, strings.Contains(result, "sync"))
	be.True(t, strings.Contains(result, "list"))
	be.True(t, strings.Contains(result, "diff"))
	be.True(t, strings.Contains(result, "help"))
}

func TestPrintSyncResultSuccess(t *testing.T) {
	r := &models.SyncResult{
		RepoName:    "test-repo",
		Destination: "gitlab",
		Action:      models.ActionCreate,
		Message:     "created successfully",
	}
	result := captureOutput(func() { PrintSyncResult(r) })
	be.True(t, strings.Contains(result, "✓"))
	be.True(t, strings.Contains(result, "test-repo"))
	be.True(t, strings.Contains(result, "gitlab"))
	be.True(t, strings.Contains(result, "created successfully"))
}

func TestPrintSyncResultError(t *testing.T) {
	r := &models.SyncResult{
		RepoName:    "failed-repo",
		Destination: "codeberg",
		Action:      models.ActionUpdate,
		Error:       fmt.Errorf("connection refused"),
		Message:     "failed to push",
	}
	result := captureOutput(func() { PrintSyncResult(r) })
	be.True(t, strings.Contains(result, "✗"))
	be.True(t, strings.Contains(result, "failed-repo"))
	be.True(t, strings.Contains(result, "connection refused"))
}

func TestPrintSyncResultSkip(t *testing.T) {
	r := &models.SyncResult{
		RepoName:    "synced-repo",
		Destination: "gitverse",
		Action:      models.ActionSkip,
		Message:     "already in sync",
	}
	result := captureOutput(func() { PrintSyncResult(r) })
	be.True(t, strings.Contains(result, "✓"))
	be.True(t, strings.Contains(result, "skip"))
	be.True(t, strings.Contains(result, "already in sync"))
}

func TestRunSyncConfigNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := RunSync([]string{}, "/nonexistent/config.yaml", logger)
	be.True(t, err != nil)
	be.True(t, strings.Contains(err.Error(), "load config"))
}

func TestRunListConfigNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := RunList([]string{}, "/nonexistent/config.yaml", logger)
	be.True(t, err != nil)
	be.True(t, strings.Contains(err.Error(), "load config"))
}

func TestRunDiffConfigNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := RunDiff([]string{}, "/nonexistent/config.yaml", logger)
	be.True(t, err != nil)
	be.True(t, strings.Contains(err.Error(), "load config"))
}

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}