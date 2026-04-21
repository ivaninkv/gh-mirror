package github

import (
	"gh-mirror/pkg/models"
)

type GHUserResponse struct {
	Login string
}

type GHRepositoryResponse struct {
	Name          string
	FullName      string
	Description   string
	Private       bool
	HTMLURL       string
	DefaultBranch string
}

type GHErrorResponse struct {
	Message string `json:"message"`
}

type GHCreateRepoRequest struct {
	Name        string `json:"name"`
	Private     *bool  `json:"private"`
	Description *string `json:"description,omitempty"`
}

type GHGetUserTestCase struct {
	Name      string
	WantUser  string
	WantErr   bool
}

type GHListReposTestCase struct {
	Name       string
	WantCount  int
	WantErr    bool
}

type GHGetRepoTestCase struct {
	Name         string
	Owner        string
	Repo         string
	WantRepoName string
	WantErr      bool
}

type GHSearchUserTestCase struct {
	Name    string
	WantErr bool
}

func GHGetUserTestCases() []GHGetUserTestCase {
	return []GHGetUserTestCase{
		{Name: "get current user", WantUser: "testuser", WantErr: false},
	}
}

func GHListReposTestCases() []GHListReposTestCase {
	return []GHListReposTestCase{
		{Name: "list user repositories", WantCount: 0, WantErr: false},
	}
}

func GHGetRepoTestCases() []GHGetRepoTestCase {
	return []GHGetRepoTestCase{
		{Name: "get repository", Owner: "user", Repo: "repo", WantRepoName: "repo", WantErr: false},
	}
}

func GHRepositoryToModel(repo GHRepositoryResponse) models.Repository {
	return models.Repository{
		PlatformID:    "github",
		Name:          repo.Name,
		FullName:      repo.FullName,
		Description:   repo.Description,
		Private:       repo.Private,
		HTMLURL:       repo.HTMLURL,
		DefaultBranch: repo.DefaultBranch,
	}
}

type GHCloneURLTestCase struct {
	Name      string
	WebURL    string
	FullName  string
	Token     string
	WantURL   string
}

func GHCloneURLTestCases() []GHCloneURLTestCase {
	return []GHCloneURLTestCase{
		{
			Name:     "github.com clone URL",
			WebURL:   "https://github.com",
			FullName: "user/repo",
			Token:    "ghp_token",
			WantURL:  "https://ghp_token@github.com/user/repo.git",
		},
		{
			Name:     "custom github enterprise clone URL",
			WebURL:   "https://github.mycompany.com",
			FullName: "user/repo",
			Token:    "token",
			WantURL:  "https://token@github.mycompany.com/user/repo.git",
		},
	}
}

type GHConfigureTestCase struct {
	Name    string
	Token   string
	APIURL  string
	WebURL  string
	WantErr bool
}

func GHConfigureTestCases() []GHConfigureTestCase {
	return []GHConfigureTestCase{
		{
			Name:    "valid configuration",
			Token:   "ghp_token",
			APIURL:  "https://api.github.com",
			WebURL:  "https://github.com",
			WantErr: false,
		},
		{
			Name:    "missing web URL",
			Token:   "ghp_token",
			APIURL:  "https://api.github.com",
			WebURL:  "",
			WantErr: true,
		},
	}
}