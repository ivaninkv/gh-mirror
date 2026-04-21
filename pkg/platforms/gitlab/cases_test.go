package gitlab

import (
	"encoding/json"

	"gh-mirror/pkg/models"
)

type UserResponse struct {
	Username string `json:"username"`
}

type RepositoryResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	PathWithNameSpace string `json:"path_with_namespace"`
	Description     string `json:"description"`
	Visibility      string `json:"visibility"`
	WebURL          string `json:"web_url"`
	DefaultBranch   string `json:"default_branch"`
}

type CreateRepositoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}

type GetAuthenticatedUserTestCase struct {
	Name         string
	ResponseCode int
	ResponseBody interface{}
	WantUser     string
	WantErr      bool
}

type ListRepositoriesTestCase struct {
	Name          string
	ResponseCode  int
	ResponseBody  interface{}
	Page          int
	PerPage       int
	WantCount     int
	WantErr       bool
}

type GetRepositoryTestCase struct {
	Name          string
	Owner         string
	Repo          string
	ResponseCode  int
	ResponseBody  interface{}
	WantRepoName  string
	WantErr       bool
}

type CreateRepositoryTestCase struct {
	Name          string
	RequestBody   CreateRepositoryRequest
	ResponseCode  int
	ResponseBody  interface{}
	WantRepoName  string
	WantErr       bool
}

type RepositoryExistsTestCase struct {
	Name         string
	Owner        string
	Repo         string
	ResponseCode int
	WantExists   bool
	WantErr      bool
}

func GetAuthenticatedUserTestCases() []GetAuthenticatedUserTestCase {
	return []GetAuthenticatedUserTestCase{
		{
			Name:         "successful get user",
			ResponseCode: 200,
			ResponseBody: UserResponse{Username: "testuser"},
			WantUser:     "testuser",
			WantErr:      false,
		},
		{
			Name:         "unauthorized",
			ResponseCode: 401,
			ResponseBody: APIErrorResponse{Message: "Unauthorized"},
			WantUser:     "",
			WantErr:      true,
		},
	}
}

func ListRepositoriesTestCases() []ListRepositoriesTestCase {
	return []ListRepositoriesTestCase{
		{
			Name:         "single page",
			ResponseCode: 200,
			ResponseBody: []RepositoryResponse{
				{ID: 1, Name: "repo1", PathWithNameSpace: "user/repo1", Visibility: "public", WebURL: "https://gitlab.com/user/repo1", DefaultBranch: "main"},
				{ID: 2, Name: "repo2", PathWithNameSpace: "user/repo2", Visibility: "private", WebURL: "https://gitlab.com/user/repo2", DefaultBranch: "master"},
			},
			WantCount: 2,
			WantErr:   false,
		},
		{
			Name:         "empty list",
			ResponseCode: 200,
			ResponseBody: []RepositoryResponse{},
			WantCount:    0,
			WantErr:      false,
		},
		{
			Name:         "error response",
			ResponseCode: 500,
			ResponseBody: APIErrorResponse{Message: "Internal Server Error"},
			WantCount:    0,
			WantErr:      true,
		},
	}
}

func GetRepositoryTestCases() []GetRepositoryTestCase {
	return []GetRepositoryTestCase{
		{
			Name:         "repository found",
			Owner:        "user",
			Repo:         "myrepo",
			ResponseCode: 200,
			ResponseBody: RepositoryResponse{
				ID: 1, Name: "myrepo", PathWithNameSpace: "user/myrepo",
				Description: "Test repo", Visibility: "public",
				WebURL: "https://gitlab.com/user/myrepo", DefaultBranch: "main",
			},
			WantRepoName: "myrepo",
			WantErr:      false,
		},
		{
			Name:         "repository not found",
			Owner:        "user",
			Repo:         "nonexistent",
			ResponseCode: 404,
			ResponseBody: APIErrorResponse{Message: "404 Project Not Found"},
			WantRepoName: "",
			WantErr:      true,
		},
	}
}

func CreateRepositoryTestCases() []CreateRepositoryTestCase {
	return []CreateRepositoryTestCase{
		{
			Name:          "create public repo",
			RequestBody:   CreateRepositoryRequest{Name: "new-repo", Description: "New repo", Visibility: "public"},
			ResponseCode:  201,
			ResponseBody:  RepositoryResponse{ID: 1, Name: "new-repo", PathWithNameSpace: "user/new-repo", Visibility: "public"},
			WantRepoName:  "new-repo",
			WantErr:       false,
		},
		{
			Name:          "create private repo",
			RequestBody:   CreateRepositoryRequest{Name: "private-repo", Description: "Private", Visibility: "private"},
			ResponseCode:  201,
			ResponseBody:  RepositoryResponse{ID: 2, Name: "private-repo", PathWithNameSpace: "user/private-repo", Visibility: "private"},
			WantRepoName:  "private-repo",
			WantErr:       false,
		},
		{
			Name:          "create conflict",
			RequestBody:   CreateRepositoryRequest{Name: "existing-repo", Visibility: "public"},
			ResponseCode:  400,
			ResponseBody:  APIErrorResponse{Message: "Name already exists"},
			WantRepoName:  "",
			WantErr:       true,
		},
	}
}

func RepositoryExistsTestCases() []RepositoryExistsTestCase {
	return []RepositoryExistsTestCase{
		{
			Name:         "exists",
			Owner:        "user",
			Repo:         "myrepo",
			ResponseCode: 200,
			WantExists:   true,
			WantErr:      false,
		},
		{
			Name:         "not found",
			Owner:        "user",
			Repo:         "nonexistent",
			ResponseCode: 404,
			WantExists:   false,
			WantErr:      false,
		},
		{
			Name:         "server error",
			Owner:        "user",
			Repo:         "myrepo",
			ResponseCode: 500,
			WantExists:   false,
			WantErr:      true,
		},
	}
}

func MockUserResponse(username string) []byte {
	resp := UserResponse{Username: username}
	data, _ := json.Marshal(resp)
	return data
}

func MockRepositoryListResponse(repos []RepositoryResponse) []byte {
	data, _ := json.Marshal(repos)
	return data
}

func MockRepositoryResponse(repo RepositoryResponse) []byte {
	data, _ := json.Marshal(repo)
	return data
}

func MockErrorResponse(message string) []byte {
	resp := APIErrorResponse{Message: message}
	data, _ := json.Marshal(resp)
	return data
}

func RepositoryToModels(repo RepositoryResponse) models.Repository {
	return models.Repository{
		PlatformID:    "gitlab",
		Name:          repo.Name,
		FullName:      repo.PathWithNameSpace,
		Description:   repo.Description,
		Private:       repo.Visibility == "private",
		HTMLURL:       repo.WebURL,
		DefaultBranch: repo.DefaultBranch,
	}
}