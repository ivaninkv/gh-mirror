package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gh-mirror/internal/config"
	"gh-mirror/internal/git"
	"gh-mirror/internal/github"
	"gh-mirror/internal/gitverse"
	"gh-mirror/pkg/models"
)

type Syncer struct {
	ghClient     GitHubClient
	gvClient     GitVerseClient
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

func (s *Syncer) Init(ctx context.Context) error {
	var err error
	s.githubUser, err = s.getGitHubUsername(ctx)
	if err != nil {
		return fmt.Errorf("get GitHub username: %w", err)
	}

	s.gitverseUser, err = s.gvClient.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("get GitVerse username: %w", err)
	}

	s.logger.Info("initialized",
		"github_user", s.githubUser,
		"gitverse_user", s.gitverseUser,
	)

	return nil
}

func (s *Syncer) getGitHubUsername(ctx context.Context) (string, error) {
	repos, err := s.ghClient.ListRepositories(ctx)
	if err != nil {
		return "", err
	}
	if len(repos) == 0 {
		return "", fmt.Errorf("no repositories found")
	}
	parts := strings.Split(repos[0].FullName, "/")
	return parts[0], nil
}

func (s *Syncer) SyncAll(ctx context.Context) ([]models.SyncResult, error) {
	s.logger.Info("starting full sync")

	githubRepos, err := s.ghClient.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list GitHub repositories: %w", err)
	}

	gitverseRepos, err := s.gvClient.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list GitVerse repositories: %w", err)
	}

	gitverseRepoMap := make(map[string]models.Repository)
	for _, r := range gitverseRepos {
		gitverseRepoMap[r.Name] = r
	}

	var results []models.SyncResult

	for _, ghRepo := range githubRepos {
		result := s.syncRepository(ctx, ghRepo, gitverseRepoMap[ghRepo.Name])
		results = append(results, result)
	}

	githubRepoNames := make(map[string]bool)
	for _, r := range githubRepos {
		githubRepoNames[r.Name] = true
	}

	var extraOnGitVerse []string
	for _, gvRepo := range gitverseRepos {
		if !githubRepoNames[gvRepo.Name] {
			extraOnGitVerse = append(extraOnGitVerse, gvRepo.Name)
		}
	}

	s.logger.Info("sync completed",
		"total", len(results),
		"created", countActions(results, models.ActionCreate),
		"updated", countActions(results, models.ActionUpdate),
		"skipped", countActions(results, models.ActionSkip),
		"extra_on_gitverse", len(extraOnGitVerse),
	)

	if len(extraOnGitVerse) > 0 {
		s.logger.Info("repositories on GitVerse not found on GitHub",
			"repositories", extraOnGitVerse,
		)
	}

	return results, nil
}

func (s *Syncer) SyncOne(ctx context.Context, repoName string) (*models.SyncResult, error) {
	ghRepo, err := s.ghClient.GetRepository(ctx, s.githubUser, repoName)
	if err != nil {
		return nil, fmt.Errorf("get GitHub repository: %w", err)
	}

	exists, err := s.gvClient.RepositoryExists(ctx, s.gitverseUser, repoName)
	if err != nil {
		return nil, fmt.Errorf("check GitVerse repository: %w", err)
	}

	var gvRepo models.Repository
	if exists {
		repo, err := s.gvClient.GetRepository(ctx, s.gitverseUser, repoName)
		if err != nil {
			return nil, fmt.Errorf("get GitVerse repository: %w", err)
		}
		if repo != nil {
			gvRepo = *repo
		}
	}

	result := s.syncRepository(ctx, *ghRepo, gvRepo)
	return &result, nil
}

func (s *Syncer) syncRepository(ctx context.Context, ghRepo, gvRepo models.Repository) models.SyncResult {
	s.logger.Info("syncing repository",
		"name", ghRepo.Name,
		"private", ghRepo.Private,
	)

	action := models.ActionUpdate
	if gvRepo.Name == "" {
		action = models.ActionCreate
	}

	if action == models.ActionUpdate {
		s.logger.Info("checking refs for changes", "name", ghRepo.Name)

		githubRefs, err := s.getRemoteRefs(ghRepo, s.cfg.GitHub.Token, "github")
		if err != nil {
			return models.SyncResult{
				RepoName: ghRepo.Name,
				Action:   action,
				Error:    err,
				Message:  "failed to get GitHub refs",
			}
		}

		gitverseRefs, err := s.getRemoteRefs(ghRepo, s.cfg.GitVerse.Token, "gitverse")
		if err != nil {
			return models.SyncResult{
				RepoName: ghRepo.Name,
				Action:   action,
				Error:    err,
				Message:  "failed to get GitVerse refs",
			}
		}

		inSync, reason := compareRefs(githubRefs, gitverseRefs)
		if inSync {
			s.logger.Info("repository already in sync", "name", ghRepo.Name)
			return models.SyncResult{
				RepoName: ghRepo.Name,
				Action:   models.ActionSkip,
				Message:  "already in sync",
			}
		}

		s.logger.Info("refs differ, will sync", "name", ghRepo.Name, "reason", reason)
	}

	if action == models.ActionCreate {
		_, err := s.gvClient.CreateRepository(ctx, ghRepo.Name, ghRepo.Private, ghRepo.Description)
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
		if gvRepo.Private != ghRepo.Private || gvRepo.Description != ghRepo.Description {
			if err := s.gvClient.UpdateRepository(ctx, s.gitverseUser, ghRepo.Name, ghRepo.Private, ghRepo.Description); err != nil {
				return models.SyncResult{
					RepoName: ghRepo.Name,
					Action:   action,
					Error:    err,
					Message:  "failed to update",
				}
			}
			s.logger.Info("updated repository",
				"name", ghRepo.Name,
				"private", ghRepo.Private,
				"description", ghRepo.Description,
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
	repoPath := git.GetRepoPath(s.tempDir, repo.Name)

	cloneURL := s.ghClient.CloneURLWithToken(repo, s.cfg.GitHub.Token)

	repoHandle, err := git.Clone(cloneURL, repoPath, s.cfg.GitHub.Token)
	if err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	if err := git.DeletePullRefs(repoPath); err != nil {
		return fmt.Errorf("delete pull refs: %w", err)
	}

	pushURL := s.gvClient.CloneURL(repo, s.cfg.GitVerse.Token)

	if err := git.Push(repoHandle, "origin", pushURL, s.cfg.GitVerse.Token, true); err != nil {
		return fmt.Errorf("git push mirror: %w", err)
	}

	return nil
}

func (s *Syncer) ListRepositories(ctx context.Context) ([]models.Repository, error) {
	return s.ghClient.ListRepositories(ctx)
}

func (s *Syncer) ListDiff(ctx context.Context) ([]models.DiffItem, error) {
	githubRepos, err := s.ghClient.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list GitHub repositories: %w", err)
	}

	gitverseRepos, err := s.gvClient.ListRepositories(ctx)
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

	var diff []models.DiffItem

	for name, ghRepo := range githubMap {
		gvRepo, exists := gitverseMap[name]
		if !exists {
			diff = append(diff, models.DiffItem{
				Name:        name,
				GitHub:      &ghRepo,
				GitVerse:    nil,
				Description: "missing on GitVerse",
			})
		} else if ghRepo.Private != gvRepo.Private {
			diff = append(diff, models.DiffItem{
				Name:        name,
				GitHub:      &ghRepo,
				GitVerse:    &gvRepo,
				Description: fmt.Sprintf("visibility mismatch: GitHub=%v, GitVerse=%v", ghRepo.Private, gvRepo.Private),
			})
		}
	}

	for name, gvRepo := range gitverseMap {
		if _, exists := githubMap[name]; !exists {
			diff = append(diff, models.DiffItem{
				Name:        name,
				GitHub:      nil,
				GitVerse:    &gvRepo,
				Description: "only on GitVerse",
			})
		}
	}

	return diff, nil
}

func (s *Syncer) getRemoteRefs(repo models.Repository, token string, remoteType string) (git.RefMap, error) {
	var cloneURL string

	if remoteType == "github" {
		cloneURL = s.ghClient.CloneURLWithToken(repo, token)
	} else {
		cloneURL = s.gvClient.CloneURL(repo, token)
	}

	refs, err := git.ListRemote(cloneURL, token)
	if err != nil {
		return nil, fmt.Errorf("git ls-remote %s: %w", remoteType, err)
	}

	return refs, nil
}

func compareRefs(githubRefs, gitverseRefs git.RefMap) (bool, string) {
	if len(githubRefs) != len(gitverseRefs) {
		return false, fmt.Sprintf("ref count mismatch: GitHub=%d, GitVerse=%d", len(githubRefs), len(gitverseRefs))
	}

	for ref, githubSHA := range githubRefs {
		gitverseSHA, exists := gitverseRefs[ref]
		if !exists {
			return false, fmt.Sprintf("ref %s missing on GitVerse", ref)
		}
		if githubSHA != gitverseSHA {
			return false, fmt.Sprintf("SHA mismatch for %s", ref)
		}
	}

	return true, ""
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
