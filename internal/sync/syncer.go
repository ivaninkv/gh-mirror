package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gh-mirror/internal/config"
	"gh-mirror/internal/git"
	"gh-mirror/pkg/models"
	"gh-mirror/pkg/platform"
	"gh-mirror/pkg/platforms/github"
)

type Syncer struct {
	source       platform.Platform
	destinations []platform.Platform
	destUsers    map[models.PlatformID]string
	cfg          *config.Config
	logger       *slog.Logger
	tempDir      string
	sourceUser   string
}

func NewSyncer(source platform.Platform, destinations []platform.Platform, cfg *config.Config, logger *slog.Logger) (*Syncer, error) {
	tempDir, err := os.MkdirTemp("", "gh-mirror-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	return &Syncer{
		source:       source,
		destinations: destinations,
		cfg:          cfg,
		logger:       logger,
		tempDir:      tempDir,
		destUsers:    make(map[models.PlatformID]string),
	}, nil
}

func (s *Syncer) Close() error {
	return os.RemoveAll(s.tempDir)
}

func (s *Syncer) Init(ctx context.Context) error {
	var err error
	s.sourceUser, err = s.source.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("get source username: %w", err)
	}

	for _, dest := range s.destinations {
		destUser, err := dest.GetAuthenticatedUser(ctx)
		if err != nil {
			return fmt.Errorf("get %s username: %w", dest.ID(), err)
		}
		s.destUsers[dest.ID()] = destUser
	}

	s.logger.Info("initialized",
		"source", s.source.ID(),
		"source_user", s.sourceUser,
		"destinations", s.destinationIDs(),
	)

	return nil
}

func (s *Syncer) destinationIDs() []models.PlatformID {
	ids := make([]models.PlatformID, len(s.destinations))
	for i, d := range s.destinations {
		ids[i] = d.ID()
	}
	return ids
}

func (s *Syncer) SyncAll(ctx context.Context) ([]models.SyncResult, error) {
	s.logger.Info("starting full sync")

	sourceRepos, err := s.source.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list source repositories: %w", err)
	}

	destRepoMaps := make(map[models.PlatformID]map[string]models.Repository)
	for _, dest := range s.destinations {
		destRepos, err := dest.ListRepositories(ctx)
		if err != nil {
			return nil, fmt.Errorf("list %s repositories: %w", dest.ID(), err)
		}
		repoMap := make(map[string]models.Repository)
		for _, r := range destRepos {
			repoMap[r.Name] = r
		}
		destRepoMaps[dest.ID()] = repoMap
	}

	var results []models.SyncResult

	for _, srcRepo := range sourceRepos {
		for _, dest := range s.destinations {
			destRepoMap := destRepoMaps[dest.ID()]
			var destRepo models.Repository
			if existing, exists := destRepoMap[srcRepo.Name]; exists {
				destRepo = existing
			}
			result := s.syncRepository(ctx, srcRepo, dest, destRepo)
			results = append(results, result)
		}
	}

	s.logger.Info("sync completed",
		"total", len(results),
		"created", countActions(results, models.ActionCreate),
		"updated", countActions(results, models.ActionUpdate),
		"skipped", countActions(results, models.ActionSkip),
	)

	return results, nil
}

func (s *Syncer) SyncOne(ctx context.Context, repoName string) ([]models.SyncResult, error) {
	srcRepo, err := s.source.GetRepository(ctx, s.sourceUser, repoName)
	if err != nil {
		return nil, fmt.Errorf("get source repository: %w", err)
	}

	var results []models.SyncResult

	for _, dest := range s.destinations {
		destUser := s.destUsers[dest.ID()]

		exists, err := dest.RepositoryExists(ctx, destUser, repoName)
		if err != nil {
			results = append(results, models.SyncResult{
				RepoName:    repoName,
				Destination: dest.ID(),
				Action:      models.ActionSkip,
				Error:       err,
				Message:     "failed to check repository",
			})
			continue
		}

		var destRepo models.Repository
		if exists {
			repo, err := dest.GetRepository(ctx, destUser, repoName)
			if err != nil {
				results = append(results, models.SyncResult{
					RepoName:    repoName,
					Destination: dest.ID(),
					Action:      models.ActionSkip,
					Error:       err,
					Message:     "failed to get repository",
				})
				continue
			}
			destRepo = *repo
		}

		result := s.syncRepository(ctx, *srcRepo, dest, destRepo)
		results = append(results, result)
	}

	return results, nil
}

func (s *Syncer) syncRepository(ctx context.Context, srcRepo models.Repository, dest platform.Platform, destRepo models.Repository) models.SyncResult {
	s.logger.Info("syncing repository",
		"name", srcRepo.Name,
		"source", s.source.ID(),
		"destination", dest.ID(),
		"private", srcRepo.Private,
	)

	action := models.ActionUpdate
	if destRepo.Name == "" {
		action = models.ActionCreate
	}

	if action == models.ActionUpdate {
		s.logger.Info("checking refs for changes", "name", srcRepo.Name)

		sourceRefs, err := s.getRemoteRefs(srcRepo, s.cfg.Platforms[string(s.source.ID())].Token, s.source)
		if err != nil {
			return models.SyncResult{
				RepoName:    srcRepo.Name,
				Destination: dest.ID(),
				Action:      action,
				Error:       err,
				Message:     "failed to get source refs",
			}
		}

		destRefs, err := s.getRemoteRefs(srcRepo, s.cfg.Platforms[string(dest.ID())].Token, dest)
		if err != nil {
			return models.SyncResult{
				RepoName:    srcRepo.Name,
				Destination: dest.ID(),
				Action:      action,
				Error:       err,
				Message:     "failed to get destination refs",
			}
		}

		inSync, reason := compareRefs(sourceRefs, destRefs)
		if inSync {
			s.logger.Info("repository already in sync", "name", srcRepo.Name)
			return models.SyncResult{
				RepoName:    srcRepo.Name,
				Destination: dest.ID(),
				Action:      models.ActionSkip,
				Message:     "already in sync",
			}
		}

		s.logger.Info("refs differ, will sync", "name", srcRepo.Name, "reason", reason)
	}

	if action == models.ActionCreate {
		_, err := dest.CreateRepository(ctx, srcRepo.Name, srcRepo.Private, srcRepo.Description)
		if err != nil {
			return models.SyncResult{
				RepoName:    srcRepo.Name,
				Destination: dest.ID(),
				Action:      action,
				Error:       err,
				Message:     "failed to create",
			}
		}
		s.logger.Info("created repository", "name", srcRepo.Name, "destination", dest.ID())
	} else {
		destUser := s.destUsers[dest.ID()]
		if destRepo.Private != srcRepo.Private || destRepo.Description != srcRepo.Description {
			if err := dest.UpdateRepository(ctx, destUser, srcRepo.Name, srcRepo.Private, srcRepo.Description); err != nil {
				return models.SyncResult{
					RepoName:    srcRepo.Name,
					Destination: dest.ID(),
					Action:      action,
					Error:       err,
					Message:     "failed to update",
				}
			}
			s.logger.Info("updated repository",
				"name", srcRepo.Name,
				"destination", dest.ID(),
				"private", srcRepo.Private,
				"description", srcRepo.Description,
			)
		}
	}

	if err := s.pushMirror(srcRepo, dest); err != nil {
		return models.SyncResult{
			RepoName:    srcRepo.Name,
			Destination: dest.ID(),
			Action:      action,
			Error:       err,
			Message:     "failed to push mirror",
		}
	}

	return models.SyncResult{
		RepoName:    srcRepo.Name,
		Destination: dest.ID(),
		Action:      action,
		Message:     "synced successfully",
	}
}

func (s *Syncer) pushMirror(repo models.Repository, dest platform.Platform) error {
	repoPath := git.GetRepoPath(s.tempDir, repo.Name)

	cloneURL := s.source.CloneURL(repo, s.cfg.Platforms[string(s.source.ID())].Token)

	repoHandle, err := git.Clone(cloneURL, repoPath, s.cfg.Platforms[string(s.source.ID())].Token)
	if err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	if err := github.DeletePullRefs(repoPath); err != nil {
		return fmt.Errorf("delete pull refs: %w", err)
	}

	pushURL := dest.CloneURL(repo, s.cfg.Platforms[string(dest.ID())].Token)

	if err := git.Push(repoHandle, "origin", pushURL, s.cfg.Platforms[string(dest.ID())].Token, true); err != nil {
		return fmt.Errorf("git push mirror: %w", err)
	}

	return nil
}

func (s *Syncer) ListRepositories(ctx context.Context) ([]models.Repository, error) {
	return s.source.ListRepositories(ctx)
}

func (s *Syncer) ListDiff(ctx context.Context) ([]models.DiffItem, error) {
	sourceRepos, err := s.source.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list source repositories: %w", err)
	}

	sourceMap := make(map[string]models.Repository)
	for _, r := range sourceRepos {
		sourceMap[r.Name] = r
	}

	var diff []models.DiffItem

	if len(s.destinations) == 0 {
		return diff, nil
	}

	dest := s.destinations[0]

	destRepos, err := dest.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list %s repositories: %w", dest.ID(), err)
	}

	destMap := make(map[string]models.Repository)
	for _, r := range destRepos {
		destMap[r.Name] = r
	}

	for name, srcRepo := range sourceMap {
		destRepo, exists := destMap[name]
		if !exists {
			diff = append(diff, models.DiffItem{
				Name:        name,
				Source:      &srcRepo,
				Destination: nil,
				Description: fmt.Sprintf("missing on %s", dest.ID()),
			})
		} else if srcRepo.Private != destRepo.Private {
			diff = append(diff, models.DiffItem{
				Name:        name,
				Source:      &srcRepo,
				Destination: &destRepo,
				Description: fmt.Sprintf("visibility mismatch: %s=%v, %s=%v", s.source.ID(), srcRepo.Private, dest.ID(), destRepo.Private),
			})
		}
	}

	for name, destRepo := range destMap {
		if _, exists := sourceMap[name]; !exists {
			diff = append(diff, models.DiffItem{
				Name:        name,
				Source:      nil,
				Destination: &destRepo,
				Description: fmt.Sprintf("only on %s", dest.ID()),
			})
		}
	}

	return diff, nil
}

func (s *Syncer) getRemoteRefs(repo models.Repository, token string, p platform.Platform) (map[string]string, error) {
	cloneURL := p.CloneURL(repo, token)

	refs, err := git.ListRemote(cloneURL, token)
	if err != nil {
		return nil, fmt.Errorf("git ls-remote %s: %w", p.ID(), err)
	}

	return refs, nil
}

func compareRefs(sourceRefs, destRefs map[string]string) (bool, string) {
	if len(sourceRefs) != len(destRefs) {
		return false, fmt.Sprintf("ref count mismatch: source=%d, dest=%d", len(sourceRefs), len(destRefs))
	}

	for ref, sourceSHA := range sourceRefs {
		destSHA, exists := destRefs[ref]
		if !exists {
			return false, fmt.Sprintf("ref %s missing on destination", ref)
		}
		if sourceSHA != destSHA {
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
