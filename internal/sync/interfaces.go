package sync

import (
	"context"
	"gh-mirror/pkg/models"
)

type GitHubClient interface {
	ListRepositories(ctx context.Context) ([]models.Repository, error)
	GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error)
	CloneURLWithToken(repo models.Repository, token string) string
}

type GitVerseClient interface {
	GetAuthenticatedUser(ctx context.Context) (string, error)
	ListRepositories(ctx context.Context) ([]models.Repository, error)
	RepositoryExists(ctx context.Context, owner, repo string) (bool, error)
	GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error)
	CreateRepository(ctx context.Context, name string, private bool, description string) (*models.Repository, error)
	UpdateRepository(ctx context.Context, owner, repo string, private bool, description string) error
	CloneURL(repo models.Repository, token string) string
}