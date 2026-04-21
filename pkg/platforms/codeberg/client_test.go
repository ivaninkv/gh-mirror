package codeberg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nalgeon/be"
	"gh-mirror/pkg/models"
)

func TestGetAuthenticatedUser(t *testing.T) {
	for _, tc := range GetUserTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(CBErrorResponse); ok {
						respBody = MockCBErrorResponse(errResp.Message)
					} else if user, ok := tc.ResponseBody.(CBUserResponse); ok {
						respBody = MockCBUserResponse(user.Login)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://codeberg.org",
				httpClient: &http.Client{},
			}

			user, err := client.GetAuthenticatedUser(context.Background())

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
	for _, tc := range ListReposTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(CBErrorResponse); ok {
						respBody = MockCBErrorResponse(errResp.Message)
					} else if repos, ok := tc.ResponseBody.([]CBRepositoryResponse); ok {
						respBody = MockCBRepoListResponse(repos)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://codeberg.org",
				httpClient: &http.Client{},
			}

			repos, err := client.ListRepositories(context.Background())

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
	for _, tc := range GetRepoTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(CBErrorResponse); ok {
						respBody = MockCBErrorResponse(errResp.Message)
					} else if repo, ok := tc.ResponseBody.(CBRepositoryResponse); ok {
						respBody = MockCBRepoResponse(repo)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://codeberg.org",
				httpClient: &http.Client{},
			}

			repo, err := client.GetRepository(context.Background(), tc.Owner, tc.Repo)

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
	for _, tc := range RepoExistsTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://codeberg.org",
				httpClient: &http.Client{},
			}

			exists, err := client.RepositoryExists(context.Background(), tc.Owner, tc.Repo)

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
	client := &Client{webURL: "https://codeberg.org"}
	repo := models.Repository{FullName: "user/repo"}

	url := client.CloneURL(repo, "token")
	be.Equal(t, url, "https://codeberg.org/user/repo.git")
}

func TestCleanPullRefs(t *testing.T) {
	client := &Client{}
	err := client.CleanPullRefs("/tmp/nonexistent")
	be.Equal(t, err, nil)
}

func TestConfigure(t *testing.T) {
	client := &Client{}
	err := client.Configure("token", "https://codeberg.org/api/v1", "https://codeberg.org")
	be.Equal(t, err, nil)
	be.Equal(t, client.token, "token")
}

func TestID(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.ID(), models.PlatformID("codeberg"))
}

func TestName(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.Name(), "Codeberg")
}

func BenchmarkGetAuthenticatedUser(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(MockCBUserResponse("testuser"))
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://codeberg.org",
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
		w.Write(MockCBRepoListResponse([]CBRepositoryResponse{
			{ID: 1, Name: "repo1", FullName: "user/repo1"},
		}))
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://codeberg.org",
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
		webURL:     "https://codeberg.org",
		httpClient: &http.Client{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.RepositoryExists(context.Background(), "user", "repo")
	}
}