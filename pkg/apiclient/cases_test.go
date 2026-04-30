package apiclient

import (
	"encoding/json"
	"fmt"
)

type TruncateBodyTestCase struct {
	Name   string
	Input  string
	MaxLen int
	Want   string
}

func TruncateBodyTestCases() []TruncateBodyTestCase {
	return []TruncateBodyTestCase{
		{
			Name:   "short body unchanged",
			Input:  "hello",
			MaxLen: 1024,
			Want:   "hello",
		},
		{
			Name:   "body exactly at limit",
			Input:  string(make([]byte, 1024)),
			MaxLen: 1024,
			Want:   string(make([]byte, 1024)),
		},
		{
			Name:   "body truncated at limit",
			Input:  string(make([]byte, 2048)),
			MaxLen: 1024,
			Want:   string(make([]byte, 1024)),
		},
	}
}

type APIErrorTestCase struct {
	Name    string
	Error   *APIError
	WantMsg string
}

func APIErrorTestCases() []APIErrorTestCase {
	return []APIErrorTestCase{
		{
			Name:    "error with status and message",
			Error:   &APIError{StatusCode: 404, Message: "Not Found"},
			WantMsg: "API error: status=404, body=Not Found",
		},
		{
			Name:    "error with empty message",
			Error:   &APIError{StatusCode: 500, Message: ""},
			WantMsg: "API error: status=500, body=",
		},
	}
}

type DoRequestSuccessTestCase struct {
	Name       string
	Method     string
	Path       string
	Body       interface{}
	AuthHeader string
	AuthPrefix string
	Token      string
	Response   string
	WantBody   string
}

func DoRequestSuccessTestCases() []DoRequestSuccessTestCase {
	return []DoRequestSuccessTestCase{
		{
			Name:       "GET request",
			Method:     "GET",
			Path:       "/user",
			AuthHeader: "Authorization",
			AuthPrefix: "Bearer ",
			Token:      "test-token",
			Response:   `{"login":"testuser"}`,
			WantBody:   `{"login":"testuser"}`,
		},
		{
			Name:       "POST request with body",
			Method:     "POST",
			Path:       "/projects",
			Body:        map[string]string{"name": "test"},
			AuthHeader: "PRIVATE-TOKEN",
			AuthPrefix: "",
			Token:      "glpat-test",
			Response:   `{"id":1}`,
			WantBody:   `{"id":1}`,
		},
		{
			Name:       "no auth header",
			Method:     "GET",
			Path:       "/health",
			AuthHeader: "",
			AuthPrefix: "",
			Token:      "",
			Response:   `{"status":"ok"}`,
			WantBody:   `{"status":"ok"}`,
		},
	}
}

type ClientErrorTestCase struct {
	Name       string
	StatusCode int
	WantErr    bool
	WantRetry  bool
}

func ClientErrorTestCases() []ClientErrorTestCase {
	return []ClientErrorTestCase{
		{
			Name:       "404 not found no retry",
			StatusCode: 404,
			WantErr:    true,
			WantRetry:  false,
		},
		{
			Name:       "401 unauthorized no retry",
			StatusCode: 401,
			WantErr:    true,
			WantRetry:  false,
		},
		{
			Name:       "403 forbidden no retry",
			StatusCode: 403,
			WantErr:    true,
			WantRetry:  false,
		},
		{
			Name:       "500 internal server error retry",
			StatusCode: 500,
			WantErr:    true,
			WantRetry:  true,
		},
		{
			Name:       "502 bad gateway retry",
			StatusCode: 502,
			WantErr:    true,
			WantRetry:  true,
		},
		{
			Name:       "429 too many requests retry",
			StatusCode: 429,
			WantErr:    true,
			WantRetry:  true,
		},
	}
}

type AuthHeaderTestCase struct {
	Name       string
	AuthHeader string
	AuthPrefix string
	Token      string
	WantHeader string
}

func AuthHeaderTestCases() []AuthHeaderTestCase {
	return []AuthHeaderTestCase{
		{
			Name:       "bearer token",
			AuthHeader: "Authorization",
			AuthPrefix: "Bearer ",
			Token:      "mytoken",
			WantHeader: "Bearer mytoken",
		},
		{
			Name:       "private token",
			AuthHeader: "PRIVATE-TOKEN",
			AuthPrefix: "",
			Token:      "glpat-xyz",
			WantHeader: "glpat-xyz",
		},
		{
			Name:       "token header",
			AuthHeader: "Authorization",
			AuthPrefix: "token ",
			Token:      "abc123",
			WantHeader: "token abc123",
		},
	}
}

func MakeRequestBody(body interface{}) []byte {
	if body == nil {
		return nil
	}
	data, _ := json.Marshal(body)
	return data
}

func MakeLargeJSON(keys int) string {
	result := "{"
	for i := 0; i < keys; i++ {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"key%d":"%s"`, i, string(make([]byte, 50)))
	}
	result += "}"
	return result
}