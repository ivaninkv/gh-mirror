package platform

import (
	"context"
	"gh-mirror/pkg/models"
)

type Platform interface {
	ID() models.PlatformID
	Name() string
	Configure(token string, baseURL string) error

	GetAuthenticatedUser(ctx context.Context) (string, error)
	ListRepositories(ctx context.Context) ([]models.Repository, error)
	GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error)
	CreateRepository(ctx context.Context, name string, private bool, description string) (*models.Repository, error)
	UpdateRepository(ctx context.Context, owner, repo string, private bool, description string) error
	RepositoryExists(ctx context.Context, owner, repo string) (bool, error)
	CloneURL(repo models.Repository, token string) string
}
