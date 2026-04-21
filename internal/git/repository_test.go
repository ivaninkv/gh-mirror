package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nalgeon/be"
)

func TestCloneNonExistentRemote(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-clone")

	_, err := Clone("https://nonexistent.invalid/repo.git", repoPath, "fake-token")
	be.True(t, err != nil)
}

func TestPushToNonExistentRemote(t *testing.T) {
	repo, err := initTestRepo(t)
	be.True(t, err == nil)

	err = Push(repo, "origin", "https://nonexistent.invalid/repo.git", "fake-token", true)
	be.True(t, err != nil)
}

func TestSetRemoteURLCreateNew(t *testing.T) {
	repo, err := initTestRepo(t)
	be.True(t, err == nil)

	err = SetRemoteURL(repo, "upstream", "https://example.com/repo.git", "token")
	be.True(t, err == nil)

	remote, rErr := repo.Remote("upstream")
	be.True(t, rErr == nil)
	be.Equal(t, remote.Config().URLs[0], "https://example.com/repo.git")
}

func TestSetRemoteURLUpdateExisting(t *testing.T) {
	repo, err := initTestRepo(t)
	be.True(t, err == nil)

	err = SetRemoteURL(repo, "origin", "https://new-url.example.com/repo.git", "token")
	be.True(t, err == nil)

	remote, rErr := repo.Remote("origin")
	be.True(t, rErr == nil)
	be.Equal(t, remote.Config().URLs[0], "https://new-url.example.com/repo.git")
}

func TestListRemoteNonExistentURL(t *testing.T) {
	_, err := ListRemote("https://nonexistent.invalid/repo.git", "fake-token")
	be.True(t, err != nil)
}

func TestCleanupRepoPathNonExistentDir(t *testing.T) {
	err := CleanupRepoPath("/tmp/gh-mirror-test-nonexistent-dir-12345")
	be.Equal(t, err, nil)
}

func TestCleanupRepoPathExistingDir(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	err := os.MkdirAll(repoPath, 0755)
	be.Equal(t, err, nil)

	err = os.WriteFile(filepath.Join(repoPath, "file.txt"), []byte("test"), 0644)
	be.Equal(t, err, nil)

	err = CleanupRepoPath(repoPath)
	be.Equal(t, err, nil)

	_, statErr := os.Stat(repoPath)
	be.True(t, os.IsNotExist(statErr))
}

func TestCleanupRepoPathFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-file.txt")

	err := os.WriteFile(filePath, []byte("test"), 0644)
	be.Equal(t, err, nil)

	err = CleanupRepoPath(filePath)
	be.Equal(t, err, nil)

	_, statErr := os.Stat(filePath)
	be.True(t, os.IsNotExist(statErr))
}

func TestGetRepoPathSimple(t *testing.T) {
	result := GetRepoPath("/tmp/gh-mirror", "my-repo")
	be.Equal(t, result, "/tmp/gh-mirror/my-repo")
}

func TestGetRepoPathNested(t *testing.T) {
	result := GetRepoPath("/tmp/gh-mirror/nested", "another-repo")
	be.Equal(t, result, "/tmp/gh-mirror/nested/another-repo")
}

func TestGetRepoPathEmptyName(t *testing.T) {
	result := GetRepoPath("/tmp/gh-mirror", "")
	be.Equal(t, result, "/tmp/gh-mirror")
}

func TestGetRepoPathWithPathSeparator(t *testing.T) {
	result := GetRepoPath("/tmp", "user/repo")
	be.Equal(t, result, "/tmp/user/repo")
}

func initTestRepo(t *testing.T) (*git.Repository, error) {
	t.Helper()
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		return nil, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(filePath, []byte("# Test Repo\n"), 0644); err != nil {
		return nil, err
	}

	_, err = wt.Add("README.md")
	if err != nil {
		return nil, err
	}

	commit, err := wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		return nil, err
	}

	branchRef := plumbing.NewHashReference("refs/heads/main", commit)
	if err := repo.Storer.SetReference(branchRef); err != nil {
		return nil, err
	}

	if err := repo.Storer.SetReference(branchRef); err != nil {
		return nil, err
	}

	wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/main"),
	})

	return repo, nil
}