package models

type Repository struct {
	Name          string
	FullName      string
	Description   string
	Private       bool
	HTMLURL       string
	DefaultBranch string
	UpdatedAt     string
}

type Branch struct {
	Name      string
	SHA       string
	Protected bool
}

type Tag struct {
	Name string
	SHA  string
}

type SyncAction string

const (
	ActionCreate SyncAction = "create"
	ActionUpdate SyncAction = "update"
	ActionDelete SyncAction = "delete"
	ActionSkip   SyncAction = "skip"
)

type SyncResult struct {
	RepoName string
	Action   SyncAction
	Error    error
	Message  string
}

type DiffItem struct {
	Name        string
	GitHub      *Repository
	GitVerse    *Repository
	Description string
}
