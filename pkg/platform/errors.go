package platform

import "fmt"

// Sentinel errors returned by platform operations.
var (
	ErrPlatformNotFound   = &PlatformError{Code: "PLATFORM_NOT_FOUND", Message: "platform is not registered"}
	ErrNotAuthenticated   = &PlatformError{Code: "NOT_AUTHENTICATED", Message: "authentication failed"}
	ErrRepositoryNotFound = &PlatformError{Code: "REPOSITORY_NOT_FOUND", Message: "repository not found"}
)

// PlatformError is a structured error with a machine-readable code and a human-readable message.
type PlatformError struct {
	Code    string
	Message string
}

func (e *PlatformError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
