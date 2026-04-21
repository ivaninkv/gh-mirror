package git

import "testing"

type GetRepoPathTestCase struct {
	Name     string
	TempDir  string
	RepoName string
	Want     string
}

type CleanupRepoPathTestCase struct {
	Name      string
	Setup    func(t *testing.T, tmpDir string) string
	WantErr  bool
	WantNil  bool
}

func GetRepoPathTestCases() []GetRepoPathTestCase {
	return []GetRepoPathTestCase{
		{
			Name:     "simple path",
			TempDir:  "/tmp/gh-mirror",
			RepoName: "my-repo",
			Want:     "/tmp/gh-mirror/my-repo",
		},
		{
			Name:     "nested path",
			TempDir:  "/tmp/gh-mirror/nested",
			RepoName: "another-repo",
			Want:     "/tmp/gh-mirror/nested/another-repo",
		},
		{
			Name:     "empty repo name",
			TempDir:  "/tmp/gh-mirror",
			RepoName: "",
			Want:     "/tmp/gh-mirror",
		},
		{
			Name:     "repo name with path separators",
			TempDir:  "/tmp",
			RepoName: "user/repo",
			Want:     "/tmp/user/repo",
		},
	}
}

type PathJoinTestCase struct {
	Name    string
	Parts   []string
	WantLen int
}

func PathJoinTestCases() []PathJoinTestCase {
	return []PathJoinTestCase{
		{Name: "two parts", Parts: []string{"/tmp", "repo"}, WantLen: 9},
		{Name: "three parts", Parts: []string{"/tmp", "gh-mirror", "repo"}, WantLen: 19},
	}
}