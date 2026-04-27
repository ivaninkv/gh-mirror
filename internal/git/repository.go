// Package git provides low-level Git operations (clone, push, ls-remote)
// using the pure-Go go-git library.
package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Clone performs a bare-like clone (NoCheckout) of a remote repository.
// If the target path already exists, it opens the existing repository instead.
func Clone(url, path string, token string) (*git.Repository, error) {
	if err := CleanupRepoPath(path); err != nil {
		return nil, err
	}

	auth := &http.BasicAuth{
		Username: "x-access-token",
		Password: token,
	}

	r, err := git.PlainClone(path, true, &git.CloneOptions{
		URL:        url,
		Auth:       auth,
		NoCheckout: true,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return git.PlainOpen(path)
		}
		return nil, fmt.Errorf("git clone: %w", err)
	}

	return r, nil
}

// Push force-pushes all branches and tags to the specified remote.
// It creates or updates the remote URL before pushing.
func Push(repo *git.Repository, remoteName string, pushURL string, token string, force bool) error {
	if err := SetRemoteURL(repo, remoteName, pushURL, token); err != nil {
		return fmt.Errorf("set remote URL: %w", err)
	}

	refSpecs := []config.RefSpec{
		"+refs/heads/*:refs/heads/*",
		"+refs/tags/*:refs/tags/*",
	}

	auth := &http.BasicAuth{
		Username: "x-access-token",
		Password: token,
	}

	opts := &git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   refSpecs,
		Auth:       auth,
		Force:      force,
	}

	if err := repo.Push(opts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("git push: %w", err)
	}

	return nil
}

// SetRemoteURL creates or updates a remote's URL in the repository config.
func SetRemoteURL(repo *git.Repository, remoteName, url, token string) error {
	_, err := repo.Remote(remoteName)
	if err != nil {
		if err == git.ErrRemoteNotFound {
			_, err = repo.CreateRemote(&config.RemoteConfig{
				Name: remoteName,
				URLs: []string{url},
			})
			if err != nil {
				return fmt.Errorf("create remote: %w", err)
			}
			return nil
		}
		return fmt.Errorf("get remote: %w", err)
	}

	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	for i, r := range cfg.Remotes {
		if r.Name == remoteName {
			cfg.Remotes[i].URLs = []string{url}
			break
		}
	}

	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("set config: %w", err)
	}

	return nil
}

// ListRemote performs a git ls-remote equivalent, returning a map of ref names to commit SHAs.
func ListRemote(url string, token string) (map[string]string, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})

	auth := &http.BasicAuth{
		Username: "x-access-token",
		Password: token,
	}

	refs, err := remote.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return nil, fmt.Errorf("git ls-remote: %w", err)
	}

	result := make(map[string]string)
	for _, ref := range refs {
		result[ref.Name().String()] = ref.Hash().String()
	}

	return result, nil
}

// CleanupRepoPath removes any existing directory at repoPath to prepare for a fresh clone.
func CleanupRepoPath(repoPath string) error {
	if _, err := os.Stat(repoPath); err == nil {
		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("clean existing repo dir: %w", err)
		}
	}
	return nil
}

// GetRepoPath joins a temporary directory path with a repository name.
func GetRepoPath(tempDir, repoName string) string {
	return filepath.Join(tempDir, repoName)
}
