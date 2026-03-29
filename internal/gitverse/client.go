package gitverse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"gh-mirror/pkg/models"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	debug      bool
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.debug {
		dump, _ := httputil.DumpRequestOut(req, body != nil)
		fmt.Printf("GitVerse Request:\n%s\n", string(dump))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if c.debug {
		dump, _ := httputil.DumpResponse(resp, true)
		fmt.Printf("GitVerse Response:\n%s\n", string(dump))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("gitverse API error: status=%d, body=%s", e.StatusCode, e.Message)
}

func (c *Client) GetAuthenticatedUser() (string, error) {
	resp, err := c.doRequest("GET", "/user", nil)
	if err != nil {
		return "", err
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(resp, &user); err != nil {
		return "", fmt.Errorf("parse user response: %w", err)
	}

	return user.Login, nil
}

func (c *Client) ListRepositories() ([]models.Repository, error) {
	var allRepos []models.Repository
	page := 1
	perPage := 100

	for {
		path := fmt.Sprintf("/user/repos?page=%d&per_page=%d", page, perPage)
		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("list repositories: %w", err)
		}

		var repos []struct {
			Name          string `json:"name"`
			FullName      string `json:"full_name"`
			Description   string `json:"description"`
			Private       bool   `json:"private"`
			HTMLURL       string `json:"html_url"`
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.Unmarshal(resp, &repos); err != nil {
			return nil, fmt.Errorf("parse repos response: %w", err)
		}

		for _, r := range repos {
			allRepos = append(allRepos, models.Repository{
				Name:          r.Name,
				FullName:      r.FullName,
				Description:   r.Description,
				Private:       r.Private,
				HTMLURL:       r.HTMLURL,
				DefaultBranch: r.DefaultBranch,
			})
		}

		if len(repos) < perPage {
			break
		}
		page++
	}

	return allRepos, nil
}

func (c *Client) RepositoryExists(owner, repo string) (bool, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	_, err := c.doRequest("GET", path, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) GetRepository(owner, repo string) (*models.Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var r struct {
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		Private       bool   `json:"private"`
		HTMLURL       string `json:"html_url"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("parse repo response: %w", err)
	}

	return &models.Repository{
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		Private:       r.Private,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
	}, nil
}

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Private     bool   `json:"private"`
	Description string `json:"description,omitempty"`
}

type UpdateRepoRequest struct {
	Private bool `json:"private"`
}

func (c *Client) CreateRepository(name string, private bool, description string) (*models.Repository, error) {
	reqBody := CreateRepoRequest{
		Name:        name,
		Private:     private,
		Description: description,
	}

	resp, err := c.doRequest("POST", "/user/repos", reqBody)
	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}

	var r struct {
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		Private       bool   `json:"private"`
		HTMLURL       string `json:"html_url"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return &models.Repository{
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		Private:       r.Private,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
	}, nil
}

func (c *Client) UpdateRepositoryVisibility(owner, repo string, private bool) error {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	reqBody := UpdateRepoRequest{Private: private}

	_, err := c.doRequest("PATCH", path, reqBody)
	if err != nil {
		return fmt.Errorf("update repository visibility: %w", err)
	}

	return nil
}

func (c *Client) DeleteRepository(owner, repo string) error {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	_, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("delete repository: %w", err)
	}
	return nil
}

func (c *Client) CloneURL(repo models.Repository, token string) string {
	return fmt.Sprintf("https://%s@gitverse.ru/%s.git", token, repo.FullName)
}

type BranchListResponse []struct {
	Name string `json:"name"`
}

func (c *Client) ListBranches(owner, repo string) ([]models.Branch, error) {
	path := fmt.Sprintf("/repos/%s/%s/branches", owner, repo)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}

	var branches []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(resp, &branches); err != nil {
		return nil, fmt.Errorf("parse branches response: %w", err)
	}

	var result []models.Branch
	for _, b := range branches {
		result = append(result, models.Branch{
			Name: b.Name,
			SHA:  b.Commit.SHA,
		})
	}

	return result, nil
}

type TagListResponse []struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func (c *Client) ListTags(owner, repo string) ([]models.Tag, error) {
	path := fmt.Sprintf("/repos/%s/%s/tags", owner, repo)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(resp, &tags); err != nil {
		return nil, fmt.Errorf("parse tags response: %w", err)
	}

	var result []models.Tag
	for _, t := range tags {
		result = append(result, models.Tag{
			Name: t.Name,
			SHA:  t.Commit.SHA,
		})
	}

	return result, nil
}
