package github

import (
	"testing"

	"github.com/nalgeon/be"
	"gh-mirror/pkg/models"
)

func TestCloneURL(t *testing.T) {
	for _, tc := range GHCloneURLTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			client := &Client{webURL: tc.WebURL}
			repo := models.Repository{FullName: tc.FullName}

			url := client.CloneURL(repo, tc.Token)
			be.Equal(t, url, tc.WantURL)
		})
	}
}

func TestConfigure(t *testing.T) {
	for _, tc := range GHConfigureTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			client := &Client{}
			err := client.Configure(tc.Token, tc.APIURL, tc.WebURL)

			if tc.WantErr {
				be.True(t, err != nil)
			} else {
				be.Equal(t, err, nil)
				be.Equal(t, client.token, tc.Token)
				be.Equal(t, client.webURL, tc.WebURL)
			}
		})
	}
}

func TestConfigureRequiresWebURL(t *testing.T) {
	client := &Client{}
	err := client.Configure("token", "", "")
	be.True(t, err != nil)
}

func TestID(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.ID(), models.PlatformID("github"))
}

func TestName(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.Name(), "GitHub")
}

func TestCleanPullRefsNonExistent(t *testing.T) {
	client := &Client{}
	err := client.CleanPullRefs("/tmp/nonexistent-path-12345")
	be.True(t, err != nil)
}

func TestCleanPullRefsInvalidPath(t *testing.T) {
	client := &Client{}
	err := client.CleanPullRefs("/invalid/path/that/cannot/be/opened")
	be.True(t, err != nil)
}

func BenchmarkCloneURL(b *testing.B) {
	client := &Client{webURL: "https://github.com"}
	repo := models.Repository{FullName: "user/repo"}
	token := "ghp_test_token"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.CloneURL(repo, token)
	}
}

func BenchmarkConfigure(b *testing.B) {
	client := &Client{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Configure("token", "https://api.github.com", "https://github.com")
	}
}

func BenchmarkID(b *testing.B) {
	client := &Client{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.ID()
	}
}

func BenchmarkName(b *testing.B) {
	client := &Client{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Name()
	}
}