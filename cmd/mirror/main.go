package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gh-mirror/internal/app"
	"gh-mirror/pkg/platforms/codeberg"
	"gh-mirror/pkg/platforms/github"
	"gh-mirror/pkg/platforms/gitlab"
	"gh-mirror/pkg/platforms/gitverse"
)

func main() {
	// Trigger init() functions in platform packages to self-register
	// via the platform.Register factory — required for platform.Create().
	_ = github.PlatformID
	_ = gitverse.PlatformID
	_ = gitlab.PlatformID
	_ = codeberg.PlatformID

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	configPath := resolveConfigPath()

	if len(os.Args) < 2 {
		app.PrintUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "sync":
		if err := app.RunSync(os.Args[2:], configPath, logger); err != nil {
			logger.Error("sync failed", "error", err)
			os.Exit(1)
		}
	case "list":
		if err := app.RunList(os.Args[2:], configPath, logger); err != nil {
			logger.Error("list failed", "error", err)
			os.Exit(1)
		}
	case "diff":
		if err := app.RunDiff(os.Args[2:], configPath, logger); err != nil {
			logger.Error("diff failed", "error", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		app.PrintUsage()
	case "--version", "-v":
		app.PrintVersion()
	default:
		fmt.Printf("unknown command: %s\n\n", os.Args[1])
		app.PrintUsage()
		os.Exit(1)
	}
}

// resolveConfigPath returns the path to config.yaml, checking in order:
// 1. CONFIG_PATH env variable (explicit)
// 2. ./config.yaml (current directory)
// 3. ~/.config/gh-mirror/config.yaml (XDG user config)
// 4. /etc/gh-mirror/config.yaml (system-wide)
func resolveConfigPath() string {
	if cp := os.Getenv("CONFIG_PATH"); cp != "" {
		return cp
	}

	paths := []string{
		"config.yaml",
		filepath.Join(homeDir(), ".config", "gh-mirror", "config.yaml"),
		"/etc/gh-mirror/config.yaml",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "config.yaml"
}

func homeDir() string {
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return os.Getenv("HOME")
}