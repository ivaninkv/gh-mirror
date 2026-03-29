package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gh-mirror/internal/config"
	"gh-mirror/internal/github"
	"gh-mirror/internal/gitverse"
	"gh-mirror/pkg/models"
)

type Syncer struct {
	ghClient     *github.Client
	gvClient     *gitverse.Client
	cfg          *config.Config
	logger       *slog.Logger
	tempDir      string
	githubUser   string
	gitverseUser string
}

func NewSyncer(cfg *config.Config, logger *slog.Logger) (*Syncer, error) {
	ghClient := github.NewClient(cfg.GitHub.Token)
	gvClient := gitverse.NewClient(cfg.GitVerse.BaseURL, cfg.GitVerse.Token)

	tempDir, err := os.MkdirTemp("", "gh-mirror-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	return &Syncer{
		ghClient: ghClient,
		gvClient: gvClient,
		cfg:      cfg,
		logger:   logger,
		tempDir:  tempDir,
	}, nil
}

func (s *Syncer) Close() error {
	return os.RemoveAll(s.tempDir)
}

func (s *Syncer) Init() error {
	var err error
	s.githubUser, err = s.getGitHubUsername()
	if err != nil {
		return fmt.Errorf("get GitHub username: %w", err)
	}

	s.gitverseUser, err = s.gvClient.GetAuthenticatedUser()
	if err != nil {
		return fmt.Errorf("get GitVerse username: %w", err)
	}

	s.logger.Info("initialized",
		"github_user", s.githubUser,
		"gitverse_user", s.gitverseUser,
	)

	return nil
}

func (s *Syncer) getGitHubUsername() (string, error) {
	repos, err := s.ghClient.ListRepositories()
	if err != nil {
		return "", err
	}
	if len(repos) == 0 {
		return "", fmt.Errorf("no repositories found")
	}
	parts := strings.Split(repos[0].FullName, "/")
	return parts[0], nil
}

func (s *Syncer) SyncAll() ([]models.SyncResult, error) {
	s.logger.Info("starting full sync")

	githubRepos, err := s.ghClient.ListRepositories()
	if err != nil {
		return nil, fmt.Errorf("list GitHub repositories: %w", err)
	}

	gitverseRepos, err := s.gvClient.ListRepositories()
	if err != nil {
		return nil, fmt.Errorf("list GitVerse repositories: %w", err)
	}

	gitverseRepoMap := make(map[string]models.Repository)
	for _, r := range gitverseRepos {
		gitverseRepoMap[r.Name] = r
	}

	var results []models.SyncResult

	for _, ghRepo := range githubRepos {
		result := s.syncRepository(ghRepo, gitverseRepoMap[ghRepo.Name])
		results = append(results, result)
	}

	if s.cfg.Sync.OnDelete == "delete" {
		githubRepoNames := make(map[string]bool)
		for _, r := range githubRepos {
			githubRepoNames[r.Name] = true
		}

		for _, gvRepo := range gitverseRepos {
			if !githubRepoNames[gvRepo.Name] {
				s.logger.Info("deleting repository not found on GitHub",
					"repo", gvRepo.Name)
				if err := s.gvClient.DeleteRepository(s.gitverseUser, gvRepo.Name); err != nil {
					results = append(results, models.SyncResult{
						RepoName: gvRepo.Name,
						Action:   models.ActionDelete,
						Error:    err,
						Message:  "failed to delete",
					})
				} else {
					results = append(results, models.SyncResult{
						RepoName: gvRepo.Name,
						Action:   models.ActionDelete,
						Message:  "deleted successfully",
					})
				}
			}
		}
	}

	s.logger.Info("sync completed",
		"total", len(results),
		"created", countActions(results, models.ActionCreate),
		"updated", countActions(results, models.ActionUpdate),
		"deleted", countActions(results, models.ActionDelete),
		"skipped", countActions(results, models.ActionSkip),
	)

	return results, nil
}

func (s *Syncer) SyncOne(repoName string) (*models.SyncResult, error) {
	ghRepo, err := s.ghClient.GetRepository(s.githubUser, repoName)
	if err != nil {
		return nil, fmt.Errorf("get GitHub repository: %w", err)
	}

	exists, err := s.gvClient.RepositoryExists(s.gitverseUser, repoName)
	if err != nil {
		return nil, fmt.Errorf("check GitVerse repository: %w", err)
	}

	var gvRepo models.Repository
	if exists {
		repo, _ := s.gvClient.GetRepository(s.gitverseUser, repoName)
		if repo != nil {
			gvRepo = *repo
		}
	}

	result := s.syncRepository(*ghRepo, gvRepo)
	return &result, nil
}

func (s *Syncer) syncRepository(ghRepo, gvRepo models.Repository) models.SyncResult {
	s.logger.Info("syncing repository",
		"name", ghRepo.Name,
		"private", ghRepo.Private,
	)

	action := models.ActionUpdate
	if gvRepo.Name == "" {
		action = models.ActionCreate
	}

	if action == models.ActionCreate {
		_, err := s.gvClient.CreateRepository(ghRepo.Name, ghRepo.Private, ghRepo.Description)
		if err != nil {
			return models.SyncResult{
				RepoName: ghRepo.Name,
				Action:   action,
				Error:    err,
				Message:  "failed to create",
			}
		}
		s.logger.Info("created repository", "name", ghRepo.Name)
	} else {
		if gvRepo.Private != ghRepo.Private {
			if err := s.gvClient.UpdateRepositoryVisibility(s.gitverseUser, ghRepo.Name, ghRepo.Private); err != nil {
				return models.SyncResult{
					RepoName: ghRepo.Name,
					Action:   action,
					Error:    err,
					Message:  "failed to update visibility",
				}
			}
			s.logger.Info("updated repository visibility",
				"name", ghRepo.Name,
				"private", ghRepo.Private,
			)
		}
	}

	if err := s.pushMirror(ghRepo); err != nil {
		return models.SyncResult{
			RepoName: ghRepo.Name,
			Action:   action,
			Error:    err,
			Message:  "failed to push mirror",
		}
	}

	return models.SyncResult{
		RepoName: ghRepo.Name,
		Action:   action,
		Message:  "synced successfully",
	}
}

func (s *Syncer) pushMirror(repo models.Repository) error {
	repoPath := filepath.Join(s.tempDir, repo.Name)

	if _, err := os.Stat(repoPath); err == nil {
		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("clean existing repo dir: %w", err)
		}
	}

	cloneURL := s.ghClient.CloneURLWithToken(repo, s.cfg.GitHub.Token)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.Sync.TimeoutMinutes)*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", cloneURL, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	pushURL := s.gvClient.CloneURL(repo, s.cfg.GitVerse.Token)

	cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "set-url", "origin", pushURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set push remote: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "push", "--mirror", "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push mirror: %w", err)
	}

	return nil
}

func (s *Syncer) ListRepositories() ([]models.Repository, error) {
	return s.ghClient.ListRepositories()
}

func (s *Syncer) ListDiff() ([]DiffItem, error) {
	githubRepos, err := s.ghClient.ListRepositories()
	if err != nil {
		return nil, fmt.Errorf("list GitHub repositories: %w", err)
	}

	gitverseRepos, err := s.gvClient.ListRepositories()
	if err != nil {
		return nil, fmt.Errorf("list GitVerse repositories: %w", err)
	}

	githubMap := make(map[string]models.Repository)
	for _, r := range githubRepos {
		githubMap[r.Name] = r
	}

	gitverseMap := make(map[string]models.Repository)
	for _, r := range gitverseRepos {
		gitverseMap[r.Name] = r
	}

	var diff []DiffItem

	for name, ghRepo := range githubMap {
		gvRepo, exists := gitverseMap[name]
		if !exists {
			diff = append(diff, DiffItem{
				Name:        name,
				GitHub:      &ghRepo,
				GitVerse:    nil,
				Description: "missing on GitVerse",
			})
		} else if ghRepo.Private != gvRepo.Private {
			diff = append(diff, DiffItem{
				Name:        name,
				GitHub:      &ghRepo,
				GitVerse:    &gvRepo,
				Description: fmt.Sprintf("visibility mismatch: GitHub=%v, GitVerse=%v", ghRepo.Private, gvRepo.Private),
			})
		}
	}

	for name, gvRepo := range gitverseMap {
		if _, exists := githubMap[name]; !exists {
			diff = append(diff, DiffItem{
				Name:        name,
				GitHub:      nil,
				GitVerse:    &gvRepo,
				Description: "missing on GitHub (will be deleted if on_delete=delete)",
			})
		}
	}

	return diff, nil
}

type DiffItem struct {
	Name        string
	GitHub      *models.Repository
	GitVerse    *models.Repository
	Description string
}

func countActions(results []models.SyncResult, action models.SyncAction) int {
	count := 0
	for _, r := range results {
		if r.Action == action {
			count++
		}
	}
	return count
}
