package models

import "errors"

var errSyncFailed = errors.New("sync failed")

type RepositoryTestCase struct {
	Name        string
	Repo        Repository
	WantName    string
	WantPrivate bool
}

type SyncResultTestCase struct {
	Name       string
	Result     SyncResult
	WantAction SyncAction
}

type DiffItemTestCase struct {
	Name        string
	Item        DiffItem
	HasSource   bool
	HasDest     bool
}

func RepositoryTestCases() []RepositoryTestCase {
	return []RepositoryTestCase{
		{
			Name: "public repository",
			Repo: Repository{
				PlatformID:    "github",
				Name:          "test-repo",
				FullName:      "user/test-repo",
				Description:   "Test description",
				Private:       false,
				HTMLURL:       "https://github.com/user/test-repo",
				DefaultBranch: "main",
				UpdatedAt:     "2024-01-01 00:00:00 +0000 UTC",
			},
			WantName:    "test-repo",
			WantPrivate: false,
		},
		{
			Name: "private repository",
			Repo: Repository{
				PlatformID:    "gitlab",
				Name:          "private-repo",
				FullName:      "user/private-repo",
				Description:   "Private repo",
				Private:       true,
				HTMLURL:       "https://gitlab.com/user/private-repo",
				DefaultBranch: "master",
			},
			WantName:    "private-repo",
			WantPrivate: true,
		},
		{
			Name: "repository with empty description",
			Repo: Repository{
				PlatformID:    "codeberg",
				Name:          "minimal",
				FullName:      "user/minimal",
				Description:   "",
				Private:       false,
				HTMLURL:       "https://codeberg.org/user/minimal",
				DefaultBranch: "main",
			},
			WantName:    "minimal",
			WantPrivate: false,
		},
	}
}

func SyncResultTestCases() []SyncResultTestCase {
	return []SyncResultTestCase{
		{
			Name: "create action",
			Result: SyncResult{
				RepoName:    "test-repo",
				Destination: "gitlab",
				Action:      ActionCreate,
				Message:     "created successfully",
			},
			WantAction: ActionCreate,
		},
		{
			Name: "update action",
			Result: SyncResult{
				RepoName:    "test-repo",
				Destination: "gitverse",
				Action:      ActionUpdate,
				Message:     "updated successfully",
			},
			WantAction: ActionUpdate,
		},
		{
			Name: "skip action",
			Result: SyncResult{
				RepoName:    "test-repo",
				Destination: "codeberg",
				Action:      ActionSkip,
				Message:     "already in sync",
			},
			WantAction: ActionSkip,
		},
		{
			Name: "action with error",
			Result: SyncResult{
				RepoName:    "test-repo",
				Destination: "github",
				Action:      ActionCreate,
				Error:       errSyncFailed,
				Message:     "failed to create",
			},
			WantAction: ActionCreate,
		},
	}
}

func DiffItemTestCases() []DiffItemTestCase {
	return []DiffItemTestCase{
		{
			Name: "missing on destination",
			Item: DiffItem{
				Name:                "repo1",
				Source:              &Repository{Name: "repo1", Private: false},
				Destination:         nil,
				DestinationPlatform: "gitlab",
				Description:         "missing on gitlab",
			},
			HasSource: true,
			HasDest:   false,
		},
		{
			Name: "only on destination",
			Item: DiffItem{
				Name:                "repo2",
				Source:              nil,
				Destination:         &Repository{Name: "repo2", Private: true},
				DestinationPlatform: "gitverse",
				Description:         "only on gitverse",
			},
			HasSource: false,
			HasDest:   true,
		},
		{
			Name: "visibility mismatch",
			Item: DiffItem{
				Name:                "repo3",
				Source:              &Repository{Name: "repo3", Private: true},
				Destination:         &Repository{Name: "repo3", Private: false},
				DestinationPlatform: "codeberg",
				Description:         "visibility mismatch",
			},
			HasSource: true,
			HasDest:   true,
		},
	}
}

type PlatformIDTestCase struct {
	Name    string
	ID      PlatformID
	WantStr string
}

func PlatformIDTestCases() []PlatformIDTestCase {
	return []PlatformIDTestCase{
		{Name: "github", ID: "github", WantStr: "github"},
		{Name: "gitlab", ID: "gitlab", WantStr: "gitlab"},
		{Name: "gitverse", ID: "gitverse", WantStr: "gitverse"},
		{Name: "codeberg", ID: "codeberg", WantStr: "codeberg"},
	}
}

type SyncActionTestCase struct {
	Name   string
	Action SyncAction
	Want   string
}

func SyncActionTestCases() []SyncActionTestCase {
	return []SyncActionTestCase{
		{Name: "create", Action: ActionCreate, Want: "create"},
		{Name: "update", Action: ActionUpdate, Want: "update"},
		{Name: "skip", Action: ActionSkip, Want: "skip"},
	}
}