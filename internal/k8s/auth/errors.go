package auth

import "fmt"

// AuthError represents authentication-related errors
type AuthError struct {
	Type    string
	Message string
	Cause   error
}

func (e *AuthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AuthError) Unwrap() error {
	return e.Cause
}

// NewAuthError creates a new authentication error
func NewAuthError(authType, message string, cause error) *AuthError {
	return &AuthError{
		Type:    authType,
		Message: message,
		Cause:   cause,
	}
}