package apiclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nalgeon/be"
)

func TestTruncateBody(t *testing.T) {
	for _, tc := range TruncateBodyTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			got := TruncateBody([]byte(tc.Input))
			be.Equal(t, got, tc.Want)
		})
	}
}

func TestTruncateBodyExactBoundary(t *testing.T) {
	input := strings.Repeat("a", 1024)
	got := TruncateBody([]byte(input))
	be.Equal(t, len(got), 1024)
	be.Equal(t, got, input)
}

func TestTruncateBodyOverBoundary(t *testing.T) {
	input := strings.Repeat("a", 2048)
	got := TruncateBody([]byte(input))
	be.Equal(t, len(got), 1024)
}

func TestAPIError(t *testing.T) {
	for _, tc := range APIErrorTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			got := tc.Error.Error()
			be.Equal(t, got, tc.WantMsg)
		})
	}
}

func TestAPIErrorFields(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "Not Found"}
	be.Equal(t, err.StatusCode, 404)
	be.Equal(t, err.Message, "Not Found")
}

func TestDoRequestSuccess(t *testing.T) {
	for _, tc := range DoRequestSuccessTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			var receivedMethod string
			var receivedPath string
			var receivedAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				receivedPath = r.URL.Path
				if tc.AuthHeader != "" {
					receivedAuth = r.Header.Get(tc.AuthHeader)
				}
				w.WriteHeader(200)
				w.Write([]byte(tc.Response))
			}))
			defer server.Close()

			client := New(server.URL, tc.Token, Config{
				AuthHeader: tc.AuthHeader,
				AuthPrefix: tc.AuthPrefix,
			})

			resp, err := client.DoRequest(context.Background(), tc.Method, tc.Path, tc.Body)
			be.Equal(t, err, nil)
			be.Equal(t, receivedMethod, tc.Method)
			be.Equal(t, receivedPath, tc.Path)
			be.Equal(t, string(resp), tc.WantBody)

			if tc.AuthHeader != "" {
				be.Equal(t, receivedAuth, tc.AuthPrefix+tc.Token)
			}
		})
	}
}

func TestDoRequestAuthHeaders(t *testing.T) {
	for _, tc := range AuthHeaderTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			var receivedAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedAuth = r.Header.Get(tc.AuthHeader)
				w.WriteHeader(200)
				w.Write([]byte(`{}`))
			}))
			defer server.Close()

			client := New(server.URL, tc.Token, Config{
				AuthHeader: tc.AuthHeader,
				AuthPrefix: tc.AuthPrefix,
			})

			_, err := client.DoRequest(context.Background(), "GET", "/test", nil)
			be.Equal(t, err, nil)
			be.Equal(t, receivedAuth, tc.WantHeader)
		})
	}
}

func TestDoRequestNoAuthWhenEmptyHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		privateToken := r.Header.Get("PRIVATE-TOKEN")
		be.Equal(t, auth, "")
		be.Equal(t, privateToken, "")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "", AuthPrefix: ""})
	_, err := client.DoRequest(context.Background(), "GET", "/test", nil)
	be.Equal(t, err, nil)
}

func TestDoRequestContentTypeJSON(t *testing.T) {
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})
	_, err := client.DoRequest(context.Background(), "GET", "/test", nil)
	be.Equal(t, err, nil)
	be.Equal(t, contentType, "application/json")
}

func TestDoRequestClientErrorNoRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})
	_, err := client.DoRequest(context.Background(), "GET", "/test", nil)
	be.True(t, err != nil)

	var apiErr *APIError
	be.True(t, errors.As(err, &apiErr))
	be.Equal(t, apiErr.StatusCode, 404)
	be.Equal(t, apiErr.Message, `{"message":"Not Found"}`)
}

func TestDoRequestUnauthorizedNoRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})
	_, err := client.DoRequest(context.Background(), "GET", "/test", nil)
	be.True(t, err != nil)

	var apiErr *APIError
	be.True(t, errors.As(err, &apiErr))
	be.Equal(t, apiErr.StatusCode, 401)
}

func TestDoRequestServerErrorsWithRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 4 {
			w.WriteHeader(500)
			w.Write([]byte("internal server error"))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewWithHTTPClient(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "}, &http.Client{})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.DoRequest(ctx, "GET", "/test", nil)
	be.Equal(t, err, nil)
	be.Equal(t, callCount, 4)
	be.True(t, strings.Contains(string(resp), `"ok":true`))
}

func TestDoRequest429Retries(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			w.WriteHeader(429)
			w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewWithHTTPClient(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "}, &http.Client{})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.DoRequest(ctx, "GET", "/test", nil)
	be.Equal(t, err, nil)
	be.Equal(t, callCount, 2)
	be.True(t, strings.Contains(string(resp), `"ok":true`))
}

func TestDoRequestAllRetriesExhausted(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client := NewWithHTTPClient(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "}, &http.Client{})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.DoRequest(ctx, "GET", "/test", nil)
	be.True(t, err != nil)
	be.Equal(t, callCount, 4)
	be.True(t, strings.Contains(err.Error(), "request failed after 3 retries"))
}

func TestDoRequestContextCancellation(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately after first attempt
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := client.DoRequest(ctx, "GET", "/test", nil)
	be.True(t, err != nil)
	be.True(t, ctx.Err() != nil)
}

func TestDoRequestPOSTWithBody(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		be.Equal(t, r.Method, "POST")
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})

	body := map[string]interface{}{"name": "test-repo", "private": true}
	resp, err := client.DoRequest(context.Background(), "POST", "/repos", body)
	be.Equal(t, err, nil)
	be.Equal(t, receivedBody["name"], "test-repo")
	be.Equal(t, receivedBody["private"], true)
	be.True(t, strings.Contains(string(resp), `"id":1`))
}

func TestDoRequestNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		be.Equal(t, r.ContentLength, int64(0))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New(server.URL, "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})
	_, err := client.DoRequest(context.Background(), "GET", "/test", nil)
	be.Equal(t, err, nil)
}

func TestDoRequestConnectionError(t *testing.T) {
	client := NewWithHTTPClient("http://nonexistent.invalid", "token", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "}, &http.Client{})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.DoRequest(ctx, "GET", "/test", nil)
	be.True(t, err != nil)
}

func TestNewClientConfig(t *testing.T) {
	client := New("http://localhost", "mytoken", Config{AuthHeader: "Authorization", AuthPrefix: "Bearer "})
	be.Equal(t, client.apiURL, "http://localhost")
	be.Equal(t, client.token, "mytoken")
	be.True(t, client.httpClient != nil)
	be.Equal(t, client.cfg.AuthHeader, "Authorization")
	be.Equal(t, client.cfg.AuthPrefix, "Bearer ")
}

func BenchmarkTruncateBody(b *testing.B) {
	body := []byte(strings.Repeat("a", 2048))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TruncateBody(body)
	}
}

func BenchmarkAPIErrorError(b *testing.B) {
	err := &APIError{StatusCode: 500, Message: "internal server error"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}