package platform

import (
	"fmt"
	"gh-mirror/pkg/models"
)

var (
	ErrPlatformNotFound   = &PlatformError{Code: "PLATFORM_NOT_FOUND", Message: "platform is not registered"}
	ErrNotAuthenticated   = &PlatformError{Code: "NOT_AUTHENTICATED", Message: "authentication failed"}
	ErrRepositoryNotFound = &PlatformError{Code: "REPOSITORY_NOT_FOUND", Message: "repository not found"}
)

type PlatformError struct {
	Code    string
	Message string
}

func (e *PlatformError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *PlatformError) WithPlatform(id models.PlatformID) *PlatformError {
	return &PlatformError{
		Code:    e.Code,
		Message: fmt.Sprintf("%s (platform: %s)", e.Message, id),
	}
}
