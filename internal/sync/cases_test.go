package sync

import "gh-mirror/pkg/models"

type CompareRefsTestCase struct {
	Name           string
	SourceRefs     map[string]string
	DestRefs       map[string]string
	WantInSync     bool
	WantReason     string
	WantReasonHas  string
}

type CountActionsTestCase struct {
	Name     string
	Results  []models.SyncResult
	Action   models.SyncAction
	WantCount int
}



func CompareRefsTestCases() []CompareRefsTestCase {
	return []CompareRefsTestCase{
		{
			Name:       "identical refs",
			SourceRefs: map[string]string{"refs/heads/main": "abc123", "refs/tags/v1.0": "def456"},
			DestRefs:   map[string]string{"refs/heads/main": "abc123", "refs/tags/v1.0": "def456"},
			WantInSync: true,
		},
		{
			Name:       "different ref count",
			SourceRefs: map[string]string{"refs/heads/main": "abc123", "refs/heads/dev": "xyz789"},
			DestRefs:   map[string]string{"refs/heads/main": "abc123"},
			WantInSync: false,
			WantReasonHas: "ref count mismatch",
		},
		{
			Name:       "missing ref on destination",
			SourceRefs: map[string]string{"refs/heads/main": "abc123"},
			DestRefs:   map[string]string{"refs/heads/other": "xyz789"},
			WantInSync: false,
			WantReasonHas: "missing on destination",
		},
		{
			Name:       "SHA mismatch",
			SourceRefs: map[string]string{"refs/heads/main": "abc123"},
			DestRefs:   map[string]string{"refs/heads/main": "different"},
			WantInSync: false,
			WantReasonHas: "SHA mismatch",
		},
		{
			Name:       "empty refs",
			SourceRefs: map[string]string{},
			DestRefs:   map[string]string{},
			WantInSync: true,
		},
		{
			Name:       "multiple branches in sync",
			SourceRefs: map[string]string{
				"refs/heads/main":    "aaa",
				"refs/heads/feature": "bbb",
				"refs/tags/v1.0":     "ccc",
				"refs/tags/v2.0":     "ddd",
			},
			DestRefs: map[string]string{
				"refs/heads/main":    "aaa",
				"refs/heads/feature": "bbb",
				"refs/tags/v1.0":     "ccc",
				"refs/tags/v2.0":     "ddd",
			},
			WantInSync: true,
		},
	}
}

func CountActionsTestCases() []CountActionsTestCase {
	return []CountActionsTestCase{
		{
			Name: "count create actions",
			Results: []models.SyncResult{
				{Action: models.ActionCreate},
				{Action: models.ActionUpdate},
				{Action: models.ActionCreate},
				{Action: models.ActionSkip},
			},
			Action:    models.ActionCreate,
			WantCount: 2,
		},
		{
			Name: "count update actions",
			Results: []models.SyncResult{
				{Action: models.ActionCreate},
				{Action: models.ActionUpdate},
				{Action: models.ActionUpdate},
			},
			Action:    models.ActionUpdate,
			WantCount: 2,
		},
		{
			Name: "count skip actions",
			Results: []models.SyncResult{
				{Action: models.ActionSkip},
				{Action: models.ActionSkip},
				{Action: models.ActionSkip},
			},
			Action:    models.ActionSkip,
			WantCount: 3,
		},
		{
			Name:     "empty results",
			Results:  []models.SyncResult{},
			Action:   models.ActionCreate,
			WantCount: 0,
		},
		{
			Name: "no matching action",
			Results: []models.SyncResult{
				{Action: models.ActionUpdate},
				{Action: models.ActionSkip},
			},
			Action:    models.ActionCreate,
			WantCount: 0,
		},
	}
}

type SyncResultActionTestCase struct {
	Name   string
	Result models.SyncResult
	Want   models.SyncAction
}

func SyncResultActionTestCases() []SyncResultActionTestCase {
	return []SyncResultActionTestCase{
		{Name: "create result", Result: models.SyncResult{Action: models.ActionCreate}, Want: models.ActionCreate},
		{Name: "update result", Result: models.SyncResult{Action: models.ActionUpdate}, Want: models.ActionUpdate},
		{Name: "skip result", Result: models.SyncResult{Action: models.ActionSkip}, Want: models.ActionSkip},
	}
}