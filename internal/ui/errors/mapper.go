package errors

import (
	"fmt"
	"strings"
)

// MapKubernetesError maps common Kubernetes errors to user-friendly errors
func MapKubernetesError(err error) *UserFriendlyError {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	// Connection and network errors
	if strings.Contains(errLower, "connection refused") {
		return NewUserFriendlyError(
			"Connection Failed",
			"Cannot connect to the Kubernetes cluster. The cluster may be down or unreachable.",
			ErrorSeverityError,
			ErrorCategoryConnection,
			err,
		).WithSuggestedAction("Check if the cluster is running and accessible from your network")
	}

	if strings.Contains(errLower, "timeout") || strings.Contains(errLower, "deadline exceeded") {
		return NewUserFriendlyError(
			"Connection Timeout",
			"The request to the cluster timed out. This may be due to network issues or cluster overload.",
			ErrorSeverityWarning,
			ErrorCategoryNetwork,
			err,
		).WithSuggestedAction("Check your network connection and try again")
	}

	if strings.Contains(errLower, "no such host") || strings.Contains(errLower, "name resolution") {
		return NewUserFriendlyError(
			"DNS Resolution Failed",
			"Cannot resolve the cluster hostname. Check your cluster URL in the kubeconfig.",
			ErrorSeverityError,
			ErrorCategoryNetwork,
			err,
		).WithSuggestedAction("Verify the cluster URL in your kubeconfig file")
	}

	// Authentication errors
	if strings.Contains(errLower, "unauthorized") || strings.Contains(errLower, "authentication") {
		return NewUserFriendlyError(
			"Authentication Failed",
			"Your credentials are invalid or have expired.",
			ErrorSeverityError,
			ErrorCategoryAuthentication,
			err,
		).WithSuggestedAction("Run 'oc login' to refresh your authentication")
	}

	if strings.Contains(errLower, "token") && (strings.Contains(errLower, "expired") || strings.Contains(errLower, "invalid")) {
		return NewUserFriendlyError(
			"Token Expired",
			"Your authentication token has expired and needs to be refreshed.",
			ErrorSeverityWarning,
			ErrorCategoryAuthentication,
			err,
		).WithSuggestedAction("Run 'oc login' to get a new authentication token")
	}

	// Permission errors
	if strings.Contains(errLower, "forbidden") || strings.Contains(errLower, "access denied") {
		return NewUserFriendlyError(
			"Access Denied",
			"You don't have permission to perform this operation.",
			ErrorSeverityError,
			ErrorCategoryPermission,
			err,
		).WithSuggestedAction("Contact your cluster administrator to request the necessary permissions")
	}

	// Resource not found errors
	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "404") {
		if strings.Contains(errLower, "namespace") || strings.Contains(errLower, "project") {
			return NewUserFriendlyError(
				"Project Not Found",
				"The requested project or namespace does not exist or you don't have access to it.",
				ErrorSeverityWarning,
				ErrorCategoryProject,
				err,
			).WithSuggestedAction("Check if the project exists and you have access permissions")
		}

		return NewUserFriendlyError(
			"Resource Not Found",
			"The requested resource could not be found.",
			ErrorSeverityWarning,
			ErrorCategoryResource,
			err,
		).WithSuggestedAction("Verify the resource name and try again")
	}

	// Configuration errors
	if strings.Contains(errLower, "kubeconfig") || strings.Contains(errLower, "config") {
		return NewUserFriendlyError(
			"Configuration Error",
			"There's an issue with your Kubernetes configuration file.",
			ErrorSeverityError,
			ErrorCategoryConfiguration,
			err,
		).WithSuggestedAction("Check your kubeconfig file or run 'oc login' to reconfigure")
	}

	// Certificate errors
	if strings.Contains(errLower, "certificate") || strings.Contains(errLower, "x509") || strings.Contains(errLower, "tls") {
		return NewUserFriendlyError(
			"Certificate Error",
			"There's an issue with the cluster's SSL/TLS certificate.",
			ErrorSeverityError,
			ErrorCategoryNetwork,
			err,
		).WithSuggestedAction("Check if the cluster certificate is valid or contact your administrator")
	}

	// API version errors
	if strings.Contains(errLower, "api version") || strings.Contains(errLower, "no matches for kind") {
		return NewUserFriendlyError(
			"API Version Mismatch",
			"The requested API version or resource type is not supported by this cluster.",
			ErrorSeverityWarning,
			ErrorCategoryResource,
			err,
		).WithSuggestedAction("This may be due to cluster version differences - some features may not be available")
	}

	// OpenShift specific errors
	if strings.Contains(errLower, "openshift") {
		if strings.Contains(errLower, "project") {
			return NewUserFriendlyError(
				"OpenShift Project Error",
				"There's an issue with the OpenShift project operation.",
				ErrorSeverityWarning,
				ErrorCategoryProject,
				err,
			).WithSuggestedAction("Verify the project exists and you have the required permissions")
		}
	}

	// Generic fallback
	return NewUserFriendlyError(
		"Unexpected Error",
		fmt.Sprintf("An unexpected error occurred: %s", truncateError(errStr, 100)),
		ErrorSeverityError,
		ErrorCategoryGeneral,
		err,
	).WithSuggestedAction("Try refreshing the application or contact support if the issue persists")
}

// WithSuggestedAction adds a suggested action to the error
func (e *UserFriendlyError) WithSuggestedAction(action string) *UserFriendlyError {
	e.SuggestedAction = action
	return e
}

// truncateError truncates error messages to a reasonable length for display
func truncateError(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}
