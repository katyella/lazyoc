package errors

import (
	"fmt"
	"time"
)

// AppError represents an application-specific error with additional context
type AppError struct {
	Type      ErrorType
	Message   string
	Cause     error
	Timestamp time.Time
	Context   map[string]interface{}
}

// ErrorType represents the category of error
type ErrorType int

const (
	ErrorUnknown ErrorType = iota
	ErrorConnection
	ErrorAuthentication
	ErrorPermission
	ErrorNetwork
	ErrorConfiguration
	ErrorUI
	ErrorInternal
)

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause of the error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:      errorType,
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// Wrap creates a new AppError wrapping an existing error
func Wrap(errorType ErrorType, message string, cause error) *AppError {
	return &AppError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to an AppError
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// GetTypeString returns a human-readable string for the error type
func (e *AppError) GetTypeString() string {
	switch e.Type {
	case ErrorConnection:
		return "Connection Error"
	case ErrorAuthentication:
		return "Authentication Error"
	case ErrorPermission:
		return "Permission Error"
	case ErrorNetwork:
		return "Network Error"
	case ErrorConfiguration:
		return "Configuration Error"
	case ErrorUI:
		return "UI Error"
	case ErrorInternal:
		return "Internal Error"
	default:
		return "Unknown Error"
	}
}

// IsRecoverable returns true if the error might be recoverable
func (e *AppError) IsRecoverable() bool {
	switch e.Type {
	case ErrorNetwork, ErrorConnection:
		return true
	case ErrorAuthentication, ErrorPermission, ErrorConfiguration:
		return false
	default:
		return false
	}
}

// Common error constructors
func NewConnectionError(message string, cause error) *AppError {
	return Wrap(ErrorConnection, message, cause)
}

func NewAuthError(message string, cause error) *AppError {
	return Wrap(ErrorAuthentication, message, cause)
}

func NewNetworkError(message string, cause error) *AppError {
	return Wrap(ErrorNetwork, message, cause)
}

func NewConfigError(message string, cause error) *AppError {
	return Wrap(ErrorConfiguration, message, cause)
}

func NewUIError(message string, cause error) *AppError {
	return Wrap(ErrorUI, message, cause)
}
