// Package models defines the core domain types shared across the application.
package models

// PlatformID uniquely identifies a Git hosting platform (e.g. "github", "gitlab").
type PlatformID string

// Repository holds metadata about a Git repository on a hosting platform.
type Repository struct {
	PlatformID    PlatformID
	Name          string
	FullName      string
	Description   string
	Private       bool
	HTMLURL       string
	DefaultBranch string
	UpdatedAt     string
}

// SyncAction represents the type of operation performed during a sync.
type SyncAction string

const (
	ActionCreate SyncAction = "create"
	ActionUpdate SyncAction = "update"
	ActionSkip   SyncAction = "skip"
)

// SyncResult captures the outcome of syncing a single repository to a single destination.
type SyncResult struct {
	RepoName    string
	Destination PlatformID
	Action      SyncAction
	Error       error
	Message     string
}

// DiffItem describes a discrepancy between source and destination platforms.
type DiffItem struct {
	Name                string
	Source              *Repository
	Destination         *Repository
	DestinationPlatform PlatformID
	Description         string
}
