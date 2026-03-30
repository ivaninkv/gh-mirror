package github

import (
	"context"
	"fmt"

	"gh-mirror/pkg/models"
	"gh-mirror/pkg/platform"
	"github.com/google/go-github/v67/github"
)

const PlatformID = models.PlatformID("github")

type Client struct {
	token  string
	client *github.Client
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
	return "GitHub"
}

func (c *Client) Configure(token string, baseURL string) error {
	c.token = token
	c.client = github.NewTokenClient(context.Background(), token)
	return nil
}

func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("get authenticated user: %w", err)
	}
	return user.GetLogin(), nil
}

func (c *Client) ListRepositories(ctx context.Context) ([]models.Repository, error) {
	var allRepos []models.Repository
	page := 1
	perPage := 100

	for {
		repos, resp, err := c.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
			Type:      "owner",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list repositories: %w", err)
		}

		for _, r := range repos {
			allRepos = append(allRepos, models.Repository{
				PlatformID:    PlatformID,
				Name:          r.GetName(),
				FullName:      r.GetFullName(),
				Description:   r.GetDescription(),
				Private:       r.GetPrivate(),
				HTMLURL:       r.GetHTMLURL(),
				DefaultBranch: r.GetDefaultBranch(),
				UpdatedAt:     r.GetUpdatedAt().String(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allRepos, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error) {
	r, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get repository %s/%s: %w", owner, repo, err)
	}

	return &models.Repository{
		PlatformID:    PlatformID,
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		Description:   r.GetDescription(),
		Private:       r.GetPrivate(),
		HTMLURL:       r.GetHTMLURL(),
		DefaultBranch: r.GetDefaultBranch(),
		UpdatedAt:     r.GetUpdatedAt().String(),
	}, nil
}

func (c *Client) CreateRepository(ctx context.Context, name string, private bool, description string) (*models.Repository, error) {
	r, _, err := c.client.Repositories.Create(ctx, "", &github.Repository{
		Name:        &name,
		Private:     &private,
		Description: &description,
	})
	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}

	return &models.Repository{
		PlatformID:    PlatformID,
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		Description:   r.GetDescription(),
		Private:       r.GetPrivate(),
		HTMLURL:       r.GetHTMLURL(),
		DefaultBranch: r.GetDefaultBranch(),
		UpdatedAt:     r.GetUpdatedAt().String(),
	}, nil
}

func (c *Client) UpdateRepository(ctx context.Context, owner, repo string, private bool, description string) error {
	_, _, err := c.client.Repositories.Edit(ctx, owner, repo, &github.Repository{
		Private:     &private,
		Description: &description,
	})
	if err != nil {
		return fmt.Errorf("update repository: %w", err)
	}
	return nil
}

func (c *Client) RepositoryExists(ctx context.Context, owner, repo string) (bool, error) {
	_, resp, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) CloneURL(repo models.Repository, token string) string {
	return fmt.Sprintf("https://%s@github.com/%s.git", token, repo.FullName)
}
