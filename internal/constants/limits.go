package constants

// List and pagination limits
const (
	// DefaultListLimit is the default number of items to retrieve in list operations
	DefaultListLimit = 100

	// DefaultPageSize is the default number of items per page in paginated views
	DefaultPageSize = 20
)

// Resource limits
const (
	// MaxPods is the maximum number of pods the UI can handle efficiently
	MaxPods = 1000

	// MaxContainers is the maximum number of containers per pod
	MaxContainers = 50

	// DefaultMaxEvents is the maximum number of events to retain
	DefaultMaxEvents = 100
)

// Buffer and channel sizes
const (
	// LogChannelBufferSize is the buffer size for log streaming channels
	LogChannelBufferSize = 100

	// MaxLogLines is the maximum number of log lines to keep in memory per pod
	MaxLogLines = 1000

	// MaxAppLogEntries is the maximum number of application log entries to keep
	MaxAppLogEntries = 500
)

// Retry configuration
const (
	// DefaultRetryAttempts is the standard number of retry attempts
	DefaultRetryAttempts = 3

	// MinOpenShiftAPIsThreshold is the minimum number of OpenShift APIs required for detection
	MinOpenShiftAPIsThreshold = 3
)
