package gitverse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nalgeon/be"
	"gh-mirror/pkg/models"
)

func TestGetAuthenticatedUser(t *testing.T) {
	for _, tc := range GVGetUserTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(GVErrorResponse); ok {
						respBody = MockGVErrorResponse(errResp.Message)
					} else if user, ok := tc.ResponseBody.(GVUserResponse); ok {
						respBody = MockGVUserResponse(user.Login)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitverse.ru",
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
	for _, tc := range GVListReposTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(GVErrorResponse); ok {
						respBody = MockGVErrorResponse(errResp.Message)
					} else if repos, ok := tc.ResponseBody.([]GVRepositoryResponse); ok {
						respBody = MockGVRepoListResponse(repos)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitverse.ru",
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
	for _, tc := range GVGetRepoTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
				if tc.ResponseBody != nil {
					var respBody []byte
					if errResp, ok := tc.ResponseBody.(GVErrorResponse); ok {
						respBody = MockGVErrorResponse(errResp.Message)
					} else if repo, ok := tc.ResponseBody.(GVRepositoryResponse); ok {
						respBody = MockGVRepoResponse(repo)
					}
					w.Write(respBody)
				}
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitverse.ru",
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
	for _, tc := range GVRepoExistsTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ResponseCode)
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				webURL:     "https://gitverse.ru",
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
	client := &Client{webURL: "https://gitverse.ru"}
	repo := models.Repository{FullName: "user/repo"}

	url := client.CloneURL(repo, "token")
	be.Equal(t, url, "https://token@gitverse.ru/user/repo.git")
}

func TestCleanPullRefs(t *testing.T) {
	client := &Client{}
	err := client.CleanPullRefs("/tmp/nonexistent")
	be.Equal(t, err, nil)
}

func TestConfigure(t *testing.T) {
	client := &Client{}
	err := client.Configure("token", "https://api.gitverse.ru", "https://gitverse.ru")
	be.Equal(t, err, nil)
	be.Equal(t, client.token, "token")
	be.Equal(t, client.apiURL, "https://api.gitverse.ru")
}

func TestID(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.ID(), models.PlatformID("gitverse"))
}

func TestName(t *testing.T) {
	client := &Client{}
	be.Equal(t, client.Name(), "GitVerse")
}

func BenchmarkGetAuthenticatedUser(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(MockGVUserResponse("testuser"))
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://gitverse.ru",
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
		w.Write(MockGVRepoListResponse([]GVRepositoryResponse{
			{Name: "repo1", FullName: "user/repo1"},
		}))
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		webURL:     "https://gitverse.ru",
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
		webURL:     "https://gitverse.ru",
		httpClient: &http.Client{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.RepositoryExists(context.Background(), "user", "repo")
	}
}