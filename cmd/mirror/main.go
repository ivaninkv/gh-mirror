package main

import (
	"fmt"
	"log/slog"
	"os"

	"gh-mirror/internal/app"
	"gh-mirror/pkg/platforms/codeberg"
	"gh-mirror/pkg/platforms/github"
	"gh-mirror/pkg/platforms/gitlab"
	"gh-mirror/pkg/platforms/gitverse"
)

func main() {
	_ = github.PlatformID
	_ = gitverse.PlatformID
	_ = gitlab.PlatformID
	_ = codeberg.PlatformID

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	configPath := app.GetEnvOrDefault("CONFIG_PATH", "config.yaml", os.Getenv)

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