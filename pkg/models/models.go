package models

type PlatformID string

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

type SyncAction string

const (
	ActionCreate SyncAction = "create"
	ActionUpdate SyncAction = "update"
	ActionSkip   SyncAction = "skip"
)

type SyncResult struct {
	RepoName    string
	Destination PlatformID
	Action      SyncAction
	Error       error
	Message     string
}

type DiffItem struct {
	Name                 string
	Source               *Repository
	Destination          *Repository
	DestinationPlatform  PlatformID
	Description          string
}
