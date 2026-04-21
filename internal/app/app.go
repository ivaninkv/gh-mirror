package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gh-mirror/internal/config"
	"gh-mirror/internal/sync"
	"gh-mirror/pkg/models"
	"gh-mirror/pkg/platform"
)

var Version = "dev"

func GetEnvOrDefault(key, defaultVal string, getenv func(string) string) string {
	if val := getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func PrintUsage() {
	fmt.Printf(`GitHub to GitVerse Mirror Tool

Usage:
  mirror <command> [options]

Commands:
  sync [repo-name]    Sync all repositories or a specific one
  list                 List all repositories from source platform
  diff                 Show differences between source and first destination
  help                 Show this help message

Configuration:
  All settings are managed via config.yaml (see config.yaml.example)
`)
}

func PrintSyncResult(r *models.SyncResult) {
	status := "✓"
	if r.Error != nil {
		status = "✗"
		fmt.Printf("[%s] %s %s -> %s: %v - %s\n", status, r.Action, r.RepoName, r.Destination, r.Error, r.Message)
	} else {
		fmt.Printf("[%s] %s %s -> %s: %s\n", status, r.Action, r.RepoName, r.Destination, r.Message)
	}
}

func PrintVersion() {
	fmt.Printf("mirror version %s\n", Version)
}

func createSourceAndDests(cfg *config.Config) (platform.Platform, []platform.Platform, error) {
	source, err := platform.Create(models.PlatformID(cfg.Source))
	if err != nil {
		return nil, nil, fmt.Errorf("create source platform: %w", err)
	}
	if err := source.Configure(cfg.Platforms[cfg.Source].Token, cfg.Platforms[cfg.Source].APIURL, cfg.Platforms[cfg.Source].URL); err != nil {
		return nil, nil, fmt.Errorf("configure source platform: %w", err)
	}

	var destinations []platform.Platform
	for _, destID := range cfg.Destinations {
		dest, err := platform.Create(models.PlatformID(destID))
		if err != nil {
			return nil, nil, fmt.Errorf("create destination platform %s: %w", destID, err)
		}
		destCfg := cfg.Platforms[destID]
		if err := dest.Configure(destCfg.Token, destCfg.APIURL, destCfg.URL); err != nil {
			return nil, nil, fmt.Errorf("configure destination platform %s: %w", destID, err)
		}
		destinations = append(destinations, dest)
	}

	return source, destinations, nil
}

func RunSync(args []string, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	source, destinations, err := createSourceAndDests(cfg)
	if err != nil {
		return err
	}

	syncer, err := sync.NewSyncer(source, destinations, cfg, logger)
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

		results, err := syncer.SyncOne(ctx, repoName)
		if err != nil {
			return fmt.Errorf("sync repo: %w", err)
		}

		for _, r := range results {
			PrintSyncResult(&r)
		}
	} else {
		results, err := syncer.SyncAll(ctx)
		if err != nil {
			return fmt.Errorf("sync all: %w", err)
		}

		fmt.Println("\nSync Results:")
		fmt.Println(strings.Repeat("─", 60))
		for _, r := range results {
			PrintSyncResult(&r)
		}
	}

	return nil
}

func RunList(args []string, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	source, destinations, err := createSourceAndDests(cfg)
	if err != nil {
		return err
	}

	syncer, err := sync.NewSyncer(source, destinations, cfg, logger)
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

	fmt.Printf("Source: %s (%s)\n", cfg.Source, source.Name())
	fmt.Printf("Repositories (%d total):\n", len(repos))
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

func RunDiff(args []string, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	source, destinations, err := createSourceAndDests(cfg)
	if err != nil {
		return err
	}

	syncer, err := sync.NewSyncer(source, destinations, cfg, logger)
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
		if d.Source != nil && d.Destination == nil {
			fmt.Printf("[+] %s only on source: %s (private=%v)\n", d.Name, cfg.Source, d.Source.Private)
		} else if d.Source == nil && d.Destination != nil {
			fmt.Printf("[-] %s only on destination: %s (private=%v) - %s\n", d.Name, d.DestinationPlatform, d.Destination.Private, d.Description)
		} else {
			fmt.Printf("[~] Mismatch: %s - %s\n", d.Name, d.Description)
		}
	}

	return nil
}