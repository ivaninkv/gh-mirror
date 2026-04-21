package codeberg

import (
	"encoding/json"

	"gh-mirror/pkg/models"
)

type CBUserResponse struct {
	Login string `json:"login"`
}

type CBRepositoryResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

type CBErrorResponse struct {
	Message string `json:"message"`
}

type CBCreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
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
	RequestBody  CBCreateRepoRequest
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

func GetUserTestCases() []GetUserTestCase {
	return []GetUserTestCase{
		{
			Name:         "successful get user",
			ResponseCode: 200,
			ResponseBody: CBUserResponse{Login: "testuser"},
			WantUser:     "testuser",
			WantErr:      false,
		},
		{
			Name:         "unauthorized",
			ResponseCode: 401,
			ResponseBody: CBErrorResponse{Message: "Unauthorized"},
			WantUser:     "",
			WantErr:      true,
		},
	}
}

func ListReposTestCases() []ListReposTestCase {
	return []ListReposTestCase{
		{
			Name:         "single page",
			ResponseCode: 200,
			ResponseBody: []CBRepositoryResponse{
				{ID: 1, Name: "repo1", FullName: "user/repo1", Description: "Test", Private: false, HTMLURL: "https://codeberg.org/user/repo1", DefaultBranch: "main"},
				{ID: 2, Name: "repo2", FullName: "user/repo2", Description: "Private", Private: true, HTMLURL: "https://codeberg.org/user/repo2", DefaultBranch: "master"},
			},
			WantCount: 2,
			WantErr:   false,
		},
		{
			Name:         "empty list",
			ResponseCode: 200,
			ResponseBody: []CBRepositoryResponse{},
			WantCount:    0,
			WantErr:      false,
		},
		{
			Name:         "error response",
			ResponseCode: 500,
			ResponseBody: CBErrorResponse{Message: "Internal Server Error"},
			WantCount:    0,
			WantErr:      true,
		},
	}
}

func GetRepoTestCases() []GetRepoTestCase {
	return []GetRepoTestCase{
		{
			Name:         "repository found",
			Owner:        "user",
			Repo:         "myrepo",
			ResponseCode: 200,
			ResponseBody: CBRepositoryResponse{
				ID: 1, Name: "myrepo", FullName: "user/myrepo",
				Description: "Test repo", Private: false,
				HTMLURL: "https://codeberg.org/user/myrepo", DefaultBranch: "main",
			},
			WantRepoName: "myrepo",
			WantErr:      false,
		},
		{
			Name:         "repository not found",
			Owner:        "user",
			Repo:         "nonexistent",
			ResponseCode: 404,
			ResponseBody: CBErrorResponse{Message: "Not Found"},
			WantRepoName: "",
			WantErr:      true,
		},
	}
}

func CreateRepoTestCases() []CreateRepoTestCase {
	return []CreateRepoTestCase{
		{
			Name:         "create public repo",
			RequestBody:  CBCreateRepoRequest{Name: "new-repo", Description: "New repo", Private: false},
			ResponseCode: 201,
			ResponseBody: CBRepositoryResponse{ID: 1, Name: "new-repo", FullName: "user/new-repo", Private: false},
			WantRepoName: "new-repo",
			WantErr:      false,
		},
		{
			Name:         "create private repo",
			RequestBody:  CBCreateRepoRequest{Name: "private-repo", Description: "Private", Private: true},
			ResponseCode: 201,
			ResponseBody: CBRepositoryResponse{ID: 2, Name: "private-repo", FullName: "user/private-repo", Private: true},
			WantRepoName: "private-repo",
			WantErr:      false,
		},
		{
			Name:         "create conflict",
			RequestBody:  CBCreateRepoRequest{Name: "existing-repo", Private: false},
			ResponseCode: 400,
			ResponseBody: CBErrorResponse{Message: "Name already exists"},
			WantRepoName: "",
			WantErr:      true,
		},
	}
}

func RepoExistsTestCases() []RepoExistsTestCase {
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

func MockCBUserResponse(username string) []byte {
	resp := CBUserResponse{Login: username}
	data, _ := json.Marshal(resp)
	return data
}

func MockCBRepoListResponse(repos []CBRepositoryResponse) []byte {
	data, _ := json.Marshal(repos)
	return data
}

func MockCBRepoResponse(repo CBRepositoryResponse) []byte {
	data, _ := json.Marshal(repo)
	return data
}

func MockCBErrorResponse(message string) []byte {
	resp := CBErrorResponse{Message: message}
	data, _ := json.Marshal(resp)
	return data
}

func CBRepoToModel(repo CBRepositoryResponse) models.Repository {
	return models.Repository{
		PlatformID:    "codeberg",
		Name:          repo.Name,
		FullName:      repo.FullName,
		Description:   repo.Description,
		Private:       repo.Private,
		HTMLURL:       repo.HTMLURL,
		DefaultBranch: repo.DefaultBranch,
	}
}