package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gh-mirror/pkg/models"
	"gh-mirror/pkg/platform"
)

const PlatformID = models.PlatformID("gitlab")

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
	return "GitLab"
}

func (c *Client) Configure(token string, baseURL string) error {
	c.token = token
	c.baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.Contains(c.baseURL, "/api/") {
		c.baseURL = c.baseURL + "/api/v4"
	}
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

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Content-Type", "application/json")

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
	return fmt.Sprintf("gitlab API error: status=%d, body=%s", e.StatusCode, e.Message)
}

func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	resp, err := c.doRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return "", err
	}

	var user struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(resp, &user); err != nil {
		return "", fmt.Errorf("parse user response: %w", err)
	}

	return user.Username, nil
}

func (c *Client) ListRepositories(ctx context.Context) ([]models.Repository, error) {
	var allRepos []models.Repository
	page := 1
	perPage := 100

	for {
		path := fmt.Sprintf("/projects?membership=true&page=%d&per_page=%d", page, perPage)
		resp, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("list repositories: %w", err)
		}

		var repos []struct {
			ID            int64  `json:"id"`
			Name          string `json:"name"`
			PathWithNameSpace string `json:"path_with_namespace"`
			Description   string `json:"description"`
			Visibility    string `json:"visibility"`
			WebURL        string `json:"web_url"`
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.Unmarshal(resp, &repos); err != nil {
			return nil, fmt.Errorf("parse repos response: %w", err)
		}

		for _, r := range repos {
			allRepos = append(allRepos, models.Repository{
				PlatformID:    PlatformID,
				Name:          r.Name,
				FullName:      r.PathWithNameSpace,
				Description:   r.Description,
				Private:       r.Visibility == "private",
				HTMLURL:       r.WebURL,
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
	path := fmt.Sprintf("/projects/%s%%2F%s", owner, repo)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var r struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		PathWithNameSpace string `json:"path_with_namespace"`
		Description   string `json:"description"`
		Visibility    string `json:"visibility"`
		WebURL        string `json:"web_url"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("parse repo response: %w", err)
	}

	return &models.Repository{
		PlatformID:    PlatformID,
		Name:          r.Name,
		FullName:      r.PathWithNameSpace,
		Description:   r.Description,
		Private:       r.Visibility == "private",
		HTMLURL:       r.WebURL,
		DefaultBranch: r.DefaultBranch,
	}, nil
}

func (c *Client) CreateRepository(ctx context.Context, name string, private bool, description string) (*models.Repository, error) {
	visibility := "public"
	if private {
		visibility = "private"
	}

	reqBody := struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Visibility  string `json:"visibility"`
	}{
		Name:        name,
		Description: description,
		Visibility:  visibility,
	}

	resp, err := c.doRequest(ctx, "POST", "/projects", reqBody)
	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}

	var r struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		PathWithNameSpace string `json:"path_with_namespace"`
		Description   string `json:"description"`
		Visibility    string `json:"visibility"`
		WebURL        string `json:"web_url"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return &models.Repository{
		PlatformID:    PlatformID,
		Name:          r.Name,
		FullName:      r.PathWithNameSpace,
		Description:   r.Description,
		Private:       r.Visibility == "private",
		HTMLURL:       r.WebURL,
		DefaultBranch: r.DefaultBranch,
	}, nil
}

func (c *Client) UpdateRepository(ctx context.Context, owner, repo string, private bool, description string) error {
	path := fmt.Sprintf("/projects/%s%%2F%s", owner, repo)
	visibility := "public"
	if private {
		visibility = "private"
	}
	reqBody := struct {
		Description string `json:"description,omitempty"`
		Visibility  string `json:"visibility"`
	}{
		Description: description,
		Visibility:  visibility,
	}

	_, err := c.doRequest(ctx, "PUT", path, reqBody)
	if err != nil {
		return fmt.Errorf("update repository: %w", err)
	}

	return nil
}

func (c *Client) RepositoryExists(ctx context.Context, owner, repo string) (bool, error) {
	path := fmt.Sprintf("/projects/%s%%2F%s", owner, repo)
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
	return fmt.Sprintf("https://gitlab.com/%s.git", repo.FullName)
}

func ExtractOwnerAndRepo(fullName string) (owner, repoName string, err error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid full name format: %s", fullName)
	}
	return parts[0], parts[1], nil
}

func ProjectID(owner, repo string) string {
	return fmt.Sprintf("%s%%2F%s", owner, repo)
}

func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
