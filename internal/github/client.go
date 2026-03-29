package github

import (
	"context"
	"fmt"

	"gh-mirror/pkg/models"
	"github.com/google/go-github/v67/github"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
	client := github.NewTokenClient(context.Background(), token)
	return &Client{
		client: client,
		ctx:    context.Background(),
	}
}

func (c *Client) ListRepositories() ([]models.Repository, error) {
	var allRepos []models.Repository
	page := 1
	perPage := 100

	for {
		repos, resp, err := c.client.Repositories.List(c.ctx, "", &github.RepositoryListOptions{
			Type:      "all",
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

func (c *Client) ListBranches(owner, repo string) ([]models.Branch, error) {
	var allBranches []models.Branch
	page := 1
	perPage := 100

	for {
		branches, resp, err := c.client.Repositories.ListBranches(c.ctx, owner, repo, &github.BranchListOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list branches for %s/%s: %w", owner, repo, err)
		}

		for _, b := range branches {
			allBranches = append(allBranches, models.Branch{
				Name:      b.GetName(),
				SHA:       b.GetCommit().GetSHA(),
				Protected: b.GetProtected(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allBranches, nil
}

func (c *Client) ListTags(owner, repo string) ([]models.Tag, error) {
	var allTags []models.Tag
	page := 1
	perPage := 100

	for {
		tags, resp, err := c.client.Repositories.ListTags(c.ctx, owner, repo, &github.ListOptions{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return nil, fmt.Errorf("list tags for %s/%s: %w", owner, repo, err)
		}

		for _, t := range tags {
			allTags = append(allTags, models.Tag{
				Name: t.GetName(),
				SHA:  t.GetCommit().GetSHA(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allTags, nil
}

func (c *Client) GetRepository(owner, repo string) (*models.Repository, error) {
	r, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get repository %s/%s: %w", owner, repo, err)
	}

	return &models.Repository{
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		Description:   r.GetDescription(),
		Private:       r.GetPrivate(),
		HTMLURL:       r.GetHTMLURL(),
		DefaultBranch: r.GetDefaultBranch(),
		UpdatedAt:     r.GetUpdatedAt().String(),
	}, nil
}

func (c *Client) CloneURL(repo models.Repository) string {
	return fmt.Sprintf("https://github.com/%s.git", repo.FullName)
}

func (c *Client) CloneURLWithToken(repo models.Repository, token string) string {
	return fmt.Sprintf("https://%s@github.com/%s.git", token, repo.FullName)
}
