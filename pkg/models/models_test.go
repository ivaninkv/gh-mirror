package models

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestPlatformID(t *testing.T) {
	for _, tc := range PlatformIDTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			got := string(tc.ID)
			be.Equal(t, got, tc.WantStr)
		})
	}
}

func TestRepository(t *testing.T) {
	for _, tc := range RepositoryTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			be.Equal(t, tc.Repo.Name, tc.WantName)
			be.Equal(t, tc.Repo.Private, tc.WantPrivate)
		})
	}
}

func TestSyncAction(t *testing.T) {
	for _, tc := range SyncActionTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			got := string(tc.Action)
			be.Equal(t, got, tc.Want)
		})
	}
}

func TestSyncActionConstants(t *testing.T) {
	be.Equal(t, string(ActionCreate), "create")
	be.Equal(t, string(ActionUpdate), "update")
	be.Equal(t, string(ActionSkip), "skip")
}

func TestSyncResult(t *testing.T) {
	for _, tc := range SyncResultTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			be.Equal(t, tc.Result.Action, tc.WantAction)
		})
	}
}

func TestDiffItem(t *testing.T) {
	for _, tc := range DiffItemTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.HasSource {
				be.True(t, tc.Item.Source != nil)
			} else {
				be.True(t, tc.Item.Source == nil)
			}

			if tc.HasDest {
				be.True(t, tc.Item.Destination != nil)
			} else {
				be.True(t, tc.Item.Destination == nil)
			}
		})
	}
}

func BenchmarkPlatformID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range PlatformIDTestCases() {
			_ = string(tc.ID)
		}
	}
}

func BenchmarkRepository(b *testing.B) {
	cases := RepositoryTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			_ = tc.Repo.Name
			_ = tc.Repo.Private
		}
	}
}

func BenchmarkSyncAction(b *testing.B) {
	cases := SyncActionTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			_ = string(tc.Action)
		}
	}
}

func BenchmarkSyncResult(b *testing.B) {
	cases := SyncResultTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			_ = tc.Result.Action
			_ = tc.Result.RepoName
		}
	}
}

func BenchmarkDiffItem(b *testing.B) {
	cases := DiffItemTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			_ = tc.HasSource
			_ = tc.HasDest
			_ = tc.Item.Name
		}
	}
}