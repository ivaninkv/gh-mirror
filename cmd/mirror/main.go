package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"gh-mirror/internal/config"
	"gh-mirror/internal/sync"
	"gh-mirror/pkg/models"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	configPath := getEnvOrDefault("CONFIG_PATH", "config.yaml")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "sync":
		if err := runSync(os.Args[2:], configPath, logger); err != nil {
			logger.Error("sync failed", "error", err)
			os.Exit(1)
		}
	case "list":
		if err := runList(os.Args[2:], configPath, logger); err != nil {
			logger.Error("list failed", "error", err)
			os.Exit(1)
		}
	case "diff":
		if err := runDiff(os.Args[2:], configPath, logger); err != nil {
			logger.Error("diff failed", "error", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func printUsage() {
	fmt.Printf(`GitHub to GitVerse Mirror Tool

Usage:
  mirror <command> [options]

Commands:
  sync [repo-name]    Sync all repositories or a specific one
  list                 List all GitHub repositories
  diff                 Show differences between GitHub and GitVerse
  help                 Show this help message

Environment variables:
  CONFIG_PATH          Path to config file (default: config.yaml)
  GITHUB_TOKEN         GitHub personal access token
  GITVERSE_TOKEN       GitVerse API token

Examples:
  # Sync all repositories
  mirror sync

  # Sync a specific repository
  mirror sync my-repo

  # List GitHub repositories
  mirror list

  # Show differences
  mirror diff

  # With custom config
  CONFIG_PATH=/path/to/config.yaml mirror sync
`)
}

func runSync(args []string, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	syncer, err := sync.NewSyncer(cfg, logger)
	if err != nil {
		return fmt.Errorf("create syncer: %w", err)
	}
	defer syncer.Close()

	if err := syncer.Init(ctx); err != nil {
		return fmt.Errorf("init syncer: %w", err)
	}

	if len(args) > 0 && args[0] != "" {
		repoName := args[0]
		logger.Info("syncing single repository", "name", repoName)

		result, err := syncer.SyncOne(ctx, repoName)
		if err != nil {
			return fmt.Errorf("sync repo: %w", err)
		}

		printSyncResult(result)
	} else {
		results, err := syncer.SyncAll(ctx)
		if err != nil {
			return fmt.Errorf("sync all: %w", err)
		}

		fmt.Println("\nSync Results:")
		fmt.Println(strings.Repeat("─", 60))
		for _, r := range results {
			printSyncResult(&r)
		}
	}

	return nil
}

func printSyncResult(r *models.SyncResult) {
	status := "✓"
	if r.Error != nil {
		status = "✗"
		fmt.Printf("[%s] %s %s: %v - %s\n", status, r.Action, r.RepoName, r.Error, r.Message)
	} else {
		fmt.Printf("[%s] %s %s: %s\n", status, r.Action, r.RepoName, r.Message)
	}
}

func runList(args []string, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	syncer, err := sync.NewSyncer(cfg, logger)
	if err != nil {
		return fmt.Errorf("create syncer: %w", err)
	}
	defer syncer.Close()

	if err := syncer.Init(ctx); err != nil {
		return fmt.Errorf("init syncer: %w", err)
	}

	repos, err := syncer.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("list repositories: %w", err)
	}

	fmt.Printf("GitHub Repositories (%d total):\n", len(repos))
	fmt.Println(strings.Repeat("─", 80))
	for _, r := range repos {
		visibility := "public"
		if r.Private {
			visibility = "private"
		}
		fmt.Printf("%-40s [%s] %s\n", r.Name, visibility, r.Description)
	}

	return nil
}

func runDiff(args []string, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	syncer, err := sync.NewSyncer(cfg, logger)
	if err != nil {
		return fmt.Errorf("create syncer: %w", err)
	}
	defer syncer.Close()

	if err := syncer.Init(ctx); err != nil {
		return fmt.Errorf("init syncer: %w", err)
	}

	diff, err := syncer.ListDiff(ctx)
	if err != nil {
		return fmt.Errorf("list diff: %w", err)
	}

	if len(diff) == 0 {
		fmt.Println("No differences found - repositories are in sync")
		return nil
	}

	fmt.Printf("Differences (%d items):\n", len(diff))
	fmt.Println(strings.Repeat("─", 80))

	for _, d := range diff {
		if d.GitHub != nil && d.GitVerse == nil {
			fmt.Printf("[+] GitHub only: %s (private=%v)\n", d.Name, d.GitHub.Private)
		} else if d.GitHub == nil && d.GitVerse != nil {
			fmt.Printf("[-] GitVerse only: %s (private=%v) - %s\n", d.Name, d.GitVerse.Private, d.Description)
		} else {
			fmt.Printf("[~] Mismatch: %s - %s\n", d.Name, d.Description)
		}
	}

	return nil
}
