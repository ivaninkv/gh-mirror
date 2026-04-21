package sync

import (
	"testing"

	"github.com/nalgeon/be"
	"gh-mirror/pkg/models"
)

func TestCompareRefs(t *testing.T) {
	for _, tc := range CompareRefsTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			inSync, reason := compareRefs(tc.SourceRefs, tc.DestRefs)
			be.Equal(t, inSync, tc.WantInSync)
			if tc.WantReasonHas != "" {
				be.True(t, len(reason) > 0 && contains(reason, tc.WantReasonHas))
			}
		})
	}
}

func TestCompareRefsEmpty(t *testing.T) {
	inSync, _ := compareRefs(map[string]string{}, map[string]string{})
	be.True(t, inSync)
}

func TestCompareRefsNilSource(t *testing.T) {
	inSync, reason := compareRefs(nil, map[string]string{"refs/heads/main": "abc"})
	be.True(t, !inSync)
	be.True(t, len(reason) > 0)
}

func TestCompareRefsNilDest(t *testing.T) {
	inSync, reason := compareRefs(map[string]string{"refs/heads/main": "abc"}, nil)
	be.True(t, !inSync)
	be.True(t, len(reason) > 0)
}

func TestCountActions(t *testing.T) {
	for _, tc := range CountActionsTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			count := countActions(tc.Results, tc.Action)
			be.Equal(t, count, tc.WantCount)
		})
	}
}

func TestCountActionsAllTypes(t *testing.T) {
	results := []models.SyncResult{
		{Action: models.ActionCreate},
		{Action: models.ActionUpdate},
		{Action: models.ActionSkip},
		{Action: models.ActionCreate},
		{Action: models.ActionUpdate},
	}

	be.Equal(t, countActions(results, models.ActionCreate), 2)
	be.Equal(t, countActions(results, models.ActionUpdate), 2)
	be.Equal(t, countActions(results, models.ActionSkip), 1)
}

func TestSyncResultAction(t *testing.T) {
	for _, tc := range SyncResultActionTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			be.Equal(t, tc.Result.Action, tc.Want)
		})
	}
}

func TestCompareRefsWithSameRefsDifferentOrder(t *testing.T) {
	source := map[string]string{
		"refs/heads/a": "1",
		"refs/heads/b": "2",
	}
	dest := map[string]string{
		"refs/heads/b": "2",
		"refs/heads/a": "1",
	}
	inSync, _ := compareRefs(source, dest)
	be.True(t, inSync)
}

func BenchmarkCompareRefs(b *testing.B) {
	cases := CompareRefsTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			compareRefs(tc.SourceRefs, tc.DestRefs)
		}
	}
}

func BenchmarkCountActions(b *testing.B) {
	cases := CountActionsTestCases()
	results := make([]models.SyncResult, 100)
	for i := range results {
		results[i] = models.SyncResult{Action: models.ActionCreate}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			_ = countActions(results, tc.Action)
		}
	}
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}