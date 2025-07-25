package constants

// Connection status constants
const (
	// StatusConnecting indicates a connection is being established
	StatusConnecting = "Connecting"

	// StatusRefreshing indicates data is being refreshed
	StatusRefreshing = "Refreshing"

	// StatusConnected indicates a successful connection
	StatusConnected = "Connected"

	// StatusFailed indicates a connection failure
	StatusFailed = "Failed"

	// StatusDisconnected indicates no active connection
	StatusDisconnected = "Disconnected"
)

// Pod status constants
const (
	// PodStatusRunning indicates the pod is running
	PodStatusRunning = "Running"

	// PodStatusPending indicates the pod is pending
	PodStatusPending = "Pending"

	// PodStatusFailed indicates the pod has failed
	PodStatusFailed = "Failed"

	// PodStatusSucceeded indicates the pod completed successfully
	PodStatusSucceeded = "Succeeded"

	// PodStatusUnknown indicates the pod status is unknown
	PodStatusUnknown = "Unknown"

	// PodStatusTerminating indicates the pod is terminating
	PodStatusTerminating = "Terminating"

	// PodStatusCompleted indicates the pod has completed
	PodStatusCompleted = "Completed"

	// PodStatusCrashLoopBackOff indicates the pod is in a crash loop
	PodStatusCrashLoopBackOff = "CrashLoopBackOff"

	// PodStatusError indicates the pod has an error
	PodStatusError = "Error"
)
