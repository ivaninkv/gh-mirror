package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nalgeon/be"
)

func TestGetRepoPath(t *testing.T) {
	for _, tc := range GetRepoPathTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			got := GetRepoPath(tc.TempDir, tc.RepoName)
			be.Equal(t, got, tc.Want)
		})
	}
}

func TestGetRepoPathPreservesPathSeparators(t *testing.T) {
	result := GetRepoPath("/base", "user/repo")
	be.Equal(t, filepath.Base(result), "repo")
	be.Equal(t, filepath.Dir(result), "/base/user")
}

func TestGetRepoPathHandlesSpecialCharacters(t *testing.T) {
	result := GetRepoPath("/tmp", "repo-with-dashes")
	be.Equal(t, result, "/tmp/repo-with-dashes")
}

func TestGetRepoPathHandlesDots(t *testing.T) {
	result := GetRepoPath("/tmp", "...")
	be.Equal(t, result, "/tmp/...")
}

func TestCleanupRepoPathNonExistent(t *testing.T) {
	err := CleanupRepoPath("/nonexistent/path/12345")
	be.Equal(t, err, nil)
}

func TestCleanupRepoPathExisting(t *testing.T) {
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

func TestPathJoinBehavior(t *testing.T) {
	for _, tc := range PathJoinTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			result := filepath.Join(tc.Parts...)
			be.Equal(t, len(result), tc.WantLen)
		})
	}
}

func BenchmarkGetRepoPath(b *testing.B) {
	cases := GetRepoPathTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			_ = GetRepoPath(tc.TempDir, tc.RepoName)
		}
	}
}

func BenchmarkCleanupRepoPath(b *testing.B) {
	tmpDir := b.TempDir()
	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		p := filepath.Join(tmpDir, "repo"+string(rune('0'+i)))
		os.MkdirAll(p, 0755)
		paths[i] = p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range paths {
			CleanupRepoPath(p)
			os.MkdirAll(p, 0755)
		}
	}
}

func BenchmarkPathJoin(b *testing.B) {
	parts := []string{"/tmp", "gh-mirror", "repo"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filepath.Join(parts...)
	}
}