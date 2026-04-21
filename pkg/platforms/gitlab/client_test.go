package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nalgeon/be"
	"gh-mirror/pkg/models"
)

func TestGetAuthenticatedUser(t *testing.T) {
	for _, tc := range GetAuthenticatedUserTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if str, ok := tc.ResponseBody.(APIErrorResponse); ok {
						respBody = MockErrorResponse(str.Message)
					} else if user, ok := tc.ResponseBody.(UserResponse); ok {
						respBody = MockUserResponse(user.Username)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitlab.com",
				httpClient: &http.Client{},
			}

			user, err := client.GetAuthenticatedUser(context.Background())
			be.Equal(t, receivedPath, "/user")

			if tc.WantErr {
				be.True(t, err != nil)
			} else {
				be.Equal(t, err, nil)
				be.Equal(t, user, tc.WantUser)
			}
		})
	}
}

func TestListRepositories(t *testing.T) {
	for _, tc := range ListRepositoriesTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(APIErrorResponse); ok {
						respBody = MockErrorResponse(errResp.Message)
					} else if repos, ok := tc.ResponseBody.([]RepositoryResponse); ok {
						respBody = MockRepositoryListResponse(repos)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitlab.com",
				httpClient: &http.Client{},
			}

			repos, err := client.ListRepositories(context.Background())
			be.True(t, len(receivedPath) > 0)

			if tc.WantErr {
				be.True(t, err != nil)
			} else {
				be.Equal(t, err, nil)
				be.Equal(t, len(repos), tc.WantCount)
			}
		})
	}
}

func TestGetRepository(t *testing.T) {
	for _, tc := range GetRepositoryTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(APIErrorResponse); ok {
						respBody = MockErrorResponse(errResp.Message)
					} else if repo, ok := tc.ResponseBody.(RepositoryResponse); ok {
						respBody = MockRepositoryResponse(repo)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitlab.com",
				httpClient: &http.Client{},
			}

			repo, err := client.GetRepository(context.Background(), tc.Owner, tc.Repo)
			be.True(t, len(receivedPath) > 0)

			if tc.WantErr {
				be.True(t, err != nil)
			} else {
				be.Equal(t, err, nil)
				be.Equal(t, repo.Name, tc.WantRepoName)
			}
		})
	}
}

func TestRepositoryExists(t *testing.T) {
	for _, tc := range RepositoryExistsTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(tc.ResponseCode)
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitlab.com",
				httpClient: &http.Client{},
			}

			exists, err := client.RepositoryExists(context.Background(), tc.Owner, tc.Repo)
			be.True(t, len(receivedPath) > 0)

			if tc.WantErr {
				be.True(t, err != nil)
			} else {
				be.Equal(t, err, nil)
				be.Equal(t, exists, tc.WantExists)
			}
		})
	}
}

func TestCloneURL(t *testing.T) {
	client := &Client{webURL: "https://gitlab.com"}
	repo := models.Repository{FullName: "user/repo"}

	url := client.CloneURL(repo, "token")
	be.Equal(t, url, "https://gitlab.com/user/repo.git")
}

func TestCleanPullRefs(t *testing.T) {
	client := &Client{}
	err := client.CleanPullRefs("/tmp/nonexistent")
	be.Equal(t, err, nil)
}

func TestConfigure(t *testing.T) {
	client := &Client{}
	err := client.Configure("token", "https://gitlab.com/api/v4", "https://gitlab.com")
	be.Equal(t, err, nil)
	be.Equal(t, client.token, "token")
	be.Equal(t, client.apiURL, "https://gitlab.com/api/v4")
}

func TestID(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.ID(), models.PlatformID("gitlab"))
}

func TestName(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.Name(), "GitLab")
}

func BenchmarkGetAuthenticatedUser(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(MockUserResponse("testuser"))
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://gitlab.com",
		httpClient: &http.Client{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetAuthenticatedUser(context.Background())
	}
}

func BenchmarkListRepositories(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(MockRepositoryListResponse([]RepositoryResponse{
			{ID: 1, Name: "repo1", PathWithNameSpace: "user/repo1", Visibility: "public"},
		}))
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://gitlab.com",
		httpClient: &http.Client{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ListRepositories(context.Background())
	}
}

func BenchmarkRepositoryExists(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://gitlab.com",
		httpClient: &http.Client{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.RepositoryExists(context.Background(), "user", "repo")
	}
}