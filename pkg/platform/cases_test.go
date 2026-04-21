package platform

import "gh-mirror/pkg/models"

type PlatformErrorTestCase struct {
	Name     string
	Error    *PlatformError
	WantMsg  string
	WantCode string
}

type ErrorAsTestCase struct {
	Name     string
	Err      error
	WantType *PlatformError
	WantMsg  string
}

func PlatformErrorTestCases() []PlatformErrorTestCase {
	return []PlatformErrorTestCase{
		{
			Name:     "platform not found error",
			Error:    ErrPlatformNotFound,
			WantMsg:  "PLATFORM_NOT_FOUND: platform is not registered",
			WantCode: "PLATFORM_NOT_FOUND",
		},
		{
			Name:     "not authenticated error",
			Error:    ErrNotAuthenticated,
			WantMsg:  "NOT_AUTHENTICATED: authentication failed",
			WantCode: "NOT_AUTHENTICATED",
		},
		{
			Name:     "repository not found error",
			Error:    ErrRepositoryNotFound,
			WantMsg:  "REPOSITORY_NOT_FOUND: repository not found",
			WantCode: "REPOSITORY_NOT_FOUND",
		},
	}
}

func ErrorAsTestCases() []ErrorAsTestCase {
	return []ErrorAsTestCase{
		{
			Name:     "platform not found is type of platform error",
			Err:      ErrPlatformNotFound,
			WantType: &PlatformError{Code: "PLATFORM_NOT_FOUND"},
			WantMsg:  "platform is not registered",
		},
		{
			Name:     "not authenticated is type of platform error",
			Err:      ErrNotAuthenticated,
			WantType: &PlatformError{Code: "NOT_AUTHENTICATED"},
			WantMsg:  "authentication failed",
		},
	}
}

type CreateTestCase struct {
	Name    string
	ID      models.PlatformID
	WantErr error
	WantNil bool
}

func CreateTestCases() []CreateTestCase {
	return []CreateTestCase{
		{
			Name:    "unregistered platform returns error",
			ID:      "nonexistent",
			WantErr: ErrPlatformNotFound,
			WantNil: true,
		},
	}
}