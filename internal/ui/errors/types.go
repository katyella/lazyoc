package errors

import (
	"fmt"
	"time"
)

// ErrorSeverity defines the severity level of errors
type ErrorSeverity int

const (
	ErrorSeverityInfo ErrorSeverity = iota
	ErrorSeverityWarning
	ErrorSeverityError
	ErrorSeverityCritical
)

func (s ErrorSeverity) String() string {
	switch s {
	case ErrorSeverityInfo:
		return "info"
	case ErrorSeverityWarning:
		return "warning"
	case ErrorSeverityError:
		return "error"
	case ErrorSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ErrorCategory defines the category of errors for better handling
type ErrorCategory int

const (
	ErrorCategoryConnection ErrorCategory = iota
	ErrorCategoryAuthentication
	ErrorCategoryResource
	ErrorCategoryConfiguration
	ErrorCategoryNetwork
	ErrorCategoryPermission
	ErrorCategoryProject
	ErrorCategoryGeneral
)

func (c ErrorCategory) String() string {
	switch c {
	case ErrorCategoryConnection:
		return "connection"
	case ErrorCategoryAuthentication:
		return "authentication"
	case ErrorCategoryResource:
		return "resource"
	case ErrorCategoryConfiguration:
		return "configuration"
	case ErrorCategoryNetwork:
		return "network"
	case ErrorCategoryPermission:
		return "permission"
	case ErrorCategoryProject:
		return "project"
	case ErrorCategoryGeneral:
		return "general"
	default:
		return "unknown"
	}
}

// UserFriendlyError provides enhanced error information for users
type UserFriendlyError struct {
	Title           string
	Message         string
	TechnicalDetail string
	Severity        ErrorSeverity
	Category        ErrorCategory
	Timestamp       time.Time
	Retryable       bool
	SuggestedAction string
	OriginalError   error
}

// Error implements the error interface
func (e *UserFriendlyError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Message)
}

// GetDisplayMessage returns a user-friendly error message
func (e *UserFriendlyError) GetDisplayMessage() string {
	return e.Message
}

// GetSuggestedAction returns suggested action for the user
func (e *UserFriendlyError) GetSuggestedAction() string {
	if e.SuggestedAction != "" {
		return e.SuggestedAction
	}

	// Provide default suggestions based on category
	switch e.Category {
	case ErrorCategoryConnection:
		return "Check your cluster connection and try reconnecting"
	case ErrorCategoryAuthentication:
		return "Run 'oc login' or verify your kubeconfig file"
	case ErrorCategoryPermission:
		return "Check if you have sufficient permissions for this operation"
	case ErrorCategoryProject:
		return "Verify the project/namespace exists and you have access"
	case ErrorCategoryNetwork:
		return "Check your network connection and cluster endpoint"
	default:
		return "Try refreshing or restarting the application"
	}
}

// GetIcon returns an appropriate icon for the error severity
func (e *UserFriendlyError) GetIcon() string {
	switch e.Severity {
	case ErrorSeverityInfo:
		return "ℹ️"
	case ErrorSeverityWarning:
		return "⚠️"
	case ErrorSeverityError:
		return "❌"
	case ErrorSeverityCritical:
		return "🚨"
	default:
		return "❓"
	}
}

// NewUserFriendlyError creates a new user-friendly error
func NewUserFriendlyError(title, message string, severity ErrorSeverity, category ErrorCategory, originalErr error) *UserFriendlyError {
	return &UserFriendlyError{
		Title:           title,
		Message:         message,
		Severity:        severity,
		Category:        category,
		Timestamp:       time.Now(),
		OriginalError:   originalErr,
		Retryable:       isRetryableByCategory(category),
		TechnicalDetail: getTechnicalDetail(originalErr),
	}
}

// isRetryableByCategory determines if an error is retryable based on its category
func isRetryableByCategory(category ErrorCategory) bool {
	switch category {
	case ErrorCategoryConnection, ErrorCategoryNetwork, ErrorCategoryResource:
		return true
	case ErrorCategoryAuthentication, ErrorCategoryPermission, ErrorCategoryConfiguration:
		return false
	default:
		return true // Default to retryable for unknown categories
	}
}

// getTechnicalDetail extracts technical details from the original error
func getTechnicalDetail(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
