package gitverse

import (
	"encoding/json"

	"gh-mirror/pkg/models"
)

type GVUserResponse struct {
	Login string `json:"login"`
}

type GVRepositoryResponse struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
}

type GVErrorResponse struct {
	Message string `json:"message"`
}

type GVCreateRepoRequest struct {
	Name        string `json:"name"`
	Private     bool   `json:"private"`
	Description string `json:"description,omitempty"`
}

type GVUpdateRepoRequest struct {
	Private     bool   `json:"private"`
	Description string `json:"description,omitempty"`
}

type GetUserTestCase struct {
	Name         string
	ResponseCode int
	ResponseBody interface{}
	WantUser     string
	WantErr      bool
}

type ListReposTestCase struct {
	Name         string
	ResponseCode int
	ResponseBody interface{}
	WantCount    int
	WantErr      bool
}

type GetRepoTestCase struct {
	Name          string
	Owner         string
	Repo          string
	ResponseCode  int
	ResponseBody  interface{}
	WantRepoName  string
	WantErr       bool
}

type CreateRepoTestCase struct {
	Name         string
	RequestBody  GVCreateRepoRequest
	ResponseCode int
	ResponseBody interface{}
	WantRepoName string
	WantErr      bool
}

type RepoExistsTestCase struct {
	Name         string
	Owner        string
	Repo         string
	ResponseCode int
	WantExists   bool
	WantErr      bool
}

func GVGetUserTestCases() []GetUserTestCase {
	return []GetUserTestCase{
		{
			Name:         "successful get user",
			ResponseCode: 200,
			ResponseBody: GVUserResponse{Login: "testuser"},
			WantUser:     "testuser",
			WantErr:      false,
		},
		{
			Name:         "unauthorized",
			ResponseCode: 401,
			ResponseBody: GVErrorResponse{Message: "Unauthorized"},
			WantUser:     "",
			WantErr:      true,
		},
	}
}

func GVListReposTestCases() []ListReposTestCase {
	return []ListReposTestCase{
		{
			Name:         "single page",
			ResponseCode: 200,
			ResponseBody: []GVRepositoryResponse{
				{Name: "repo1", FullName: "user/repo1", Description: "Test", Private: false, HTMLURL: "https://gitverse.ru/user/repo1", DefaultBranch: "main"},
				{Name: "repo2", FullName: "user/repo2", Description: "Private", Private: true, HTMLURL: "https://gitverse.ru/user/repo2", DefaultBranch: "master"},
			},
			WantCount: 2,
			WantErr:   false,
		},
		{
			Name:         "empty list",
			ResponseCode: 200,
			ResponseBody: []GVRepositoryResponse{},
			WantCount:    0,
			WantErr:      false,
		},
		{
			Name:         "error response",
			ResponseCode: 500,
			ResponseBody: GVErrorResponse{Message: "Internal Server Error"},
			WantCount:    0,
			WantErr:      true,
		},
	}
}

func GVGetRepoTestCases() []GetRepoTestCase {
	return []GetRepoTestCase{
		{
			Name:         "repository found",
			Owner:        "user",
			Repo:         "myrepo",
			ResponseCode: 200,
			ResponseBody: GVRepositoryResponse{
				Name: "myrepo", FullName: "user/myrepo",
				Description: "Test repo", Private: false,
				HTMLURL: "https://gitverse.ru/user/myrepo", DefaultBranch: "main",
			},
			WantRepoName: "myrepo",
			WantErr:      false,
		},
		{
			Name:         "repository not found",
			Owner:        "user",
			Repo:         "nonexistent",
			ResponseCode: 404,
			ResponseBody: GVErrorResponse{Message: "Not Found"},
			WantRepoName: "",
			WantErr:      true,
		},
	}
}

func GVCreateRepoTestCases() []CreateRepoTestCase {
	return []CreateRepoTestCase{
		{
			Name:         "create public repo",
			RequestBody:  GVCreateRepoRequest{Name: "new-repo", Description: "New repo", Private: false},
			ResponseCode: 201,
			ResponseBody: GVRepositoryResponse{Name: "new-repo", FullName: "user/new-repo", Private: false},
			WantRepoName: "new-repo",
			WantErr:      false,
		},
		{
			Name:         "create private repo",
			RequestBody:  GVCreateRepoRequest{Name: "private-repo", Description: "Private", Private: true},
			ResponseCode: 201,
			ResponseBody: GVRepositoryResponse{Name: "private-repo", FullName: "user/private-repo", Private: true},
			WantRepoName: "private-repo",
			WantErr:      false,
		},
		{
			Name:         "create conflict",
			RequestBody:  GVCreateRepoRequest{Name: "existing-repo", Private: false},
			ResponseCode: 400,
			ResponseBody: GVErrorResponse{Message: "Name already exists"},
			WantRepoName: "",
			WantErr:      true,
		},
	}
}

func GVRepoExistsTestCases() []RepoExistsTestCase {
	return []RepoExistsTestCase{
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

func MockGVUserResponse(username string) []byte {
	resp := GVUserResponse{Login: username}
	data, _ := json.Marshal(resp)
	return data
}

func MockGVRepoListResponse(repos []GVRepositoryResponse) []byte {
	data, _ := json.Marshal(repos)
	return data
}

func MockGVRepoResponse(repo GVRepositoryResponse) []byte {
	data, _ := json.Marshal(repo)
	return data
}

func MockGVErrorResponse(message string) []byte {
	resp := GVErrorResponse{Message: message}
	data, _ := json.Marshal(resp)
	return data
}

func GVRepoToModel(repo GVRepositoryResponse) models.Repository {
	return models.Repository{
		PlatformID:    "gitverse",
		Name:          repo.Name,
		FullName:      repo.FullName,
		Description:   repo.Description,
		Private:       repo.Private,
		HTMLURL:       repo.HTMLURL,
		DefaultBranch: repo.DefaultBranch,
	}
}