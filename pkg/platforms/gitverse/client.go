package gitverse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gh-mirror/pkg/models"
	"gh-mirror/pkg/platform"
)

const PlatformID = models.PlatformID("gitverse")

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func init() {
	platform.Register(PlatformID, func() platform.Platform {
		return &Client{}
	})
}

func (c *Client) ID() models.PlatformID {
	return PlatformID
}

func (c *Client) Name() string {
	return "GitVerse"
}

func (c *Client) Configure(token string, baseURL string) error {
	c.token = token
	c.baseURL = strings.TrimSuffix(baseURL, "/")
	c.httpClient = &http.Client{
		Timeout: 60 * time.Second,
	}
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
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
	req.Header.Set("Accept", "application/vnd.gitverse.object+json;version=1")

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

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

func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	resp, err := c.doRequest(ctx, "GET", "/user", nil)
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

func (c *Client) ListRepositories(ctx context.Context) ([]models.Repository, error) {
	var allRepos []models.Repository
	page := 1
	perPage := 50

	for {
		path := fmt.Sprintf("/user/repos?page=%d&per_page=%d", page, perPage)
		resp, err := c.doRequest(ctx, "GET", path, nil)
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
				PlatformID:    PlatformID,
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

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	resp, err := c.doRequest(ctx, "GET", path, nil)
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
		PlatformID:    PlatformID,
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		Private:       r.Private,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
	}, nil
}

func (c *Client) CreateRepository(ctx context.Context, name string, private bool, description string) (*models.Repository, error) {
	reqBody := struct {
		Name        string `json:"name"`
		Private     bool   `json:"private"`
		Description string `json:"description,omitempty"`
	}{
		Name:        name,
		Private:     private,
		Description: description,
	}

	resp, err := c.doRequest(ctx, "POST", "/user/repos", reqBody)
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
		PlatformID:    PlatformID,
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		Private:       r.Private,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
	}, nil
}

func (c *Client) UpdateRepository(ctx context.Context, owner, repo string, private bool, description string) error {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	reqBody := struct {
		Private     bool   `json:"private"`
		Description string `json:"description,omitempty"`
	}{
		Private:     private,
		Description: description,
	}

	_, err := c.doRequest(ctx, "PATCH", path, reqBody)
	if err != nil {
		return fmt.Errorf("update repository: %w", err)
	}

	return nil
}

func (c *Client) RepositoryExists(ctx context.Context, owner, repo string) (bool, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	_, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) CloneURL(repo models.Repository, token string) string {
	return fmt.Sprintf("https://%s@gitverse.ru/%s.git", token, repo.FullName)
}
