package constants

import "time"

// Timeout constants define various operation timeouts throughout the application
const (
	// AuthenticationTimeout is the maximum time allowed for authentication operations
	AuthenticationTimeout = 5 * time.Second

	// ConnectionTestTimeout is the maximum time allowed for testing connections
	ConnectionTestTimeout = 3 * time.Second

	// DefaultOperationTimeout is the standard timeout for general operations
	DefaultOperationTimeout = 10 * time.Second

	// ValidationTimeout is the maximum time allowed for validation operations
	ValidationTimeout = 10 * time.Second

	// ClusterDetectionTimeout is the maximum time allowed for cluster type detection
	ClusterDetectionTimeout = 15 * time.Second

	// DefaultRequestTimeout is the standard timeout for API requests
	DefaultRequestTimeout = 10 * time.Second
)

// Interval constants define refresh and check intervals
const (
	// PodRefreshInterval is the time between automatic pod list refreshes
	PodRefreshInterval = 30 * time.Second

	// PodLogRefreshInterval is the time between automatic pod log refreshes
	PodLogRefreshInterval = 500 * time.Millisecond

	// DefaultHealthCheckInterval is the time between health check operations
	DefaultHealthCheckInterval = 30 * time.Second

	// DefaultRetryDelay is the standard delay between retry attempts
	DefaultRetryDelay = 5 * time.Second
)

// Cache duration constants
const (
	// DefaultClusterCacheTime is how long cluster detection results are cached
	DefaultClusterCacheTime = 10 * time.Minute
)

// Backoff configuration constants
const (
	// DefaultInitialDelay is the initial delay for exponential backoff
	DefaultInitialDelay = 1 * time.Second

	// DefaultMaxDelay is the maximum delay for exponential backoff
	DefaultMaxDelay = 30 * time.Second

	// ConnectionInitialDelay is the initial delay for connection retries
	ConnectionInitialDelay = 2 * time.Second

	// ConnectionMaxDelay is the maximum delay for connection retries
	ConnectionMaxDelay = 60 * time.Second

	// DefaultBackoffFactor is the multiplier for exponential backoff
	DefaultBackoffFactor = 2.0
)
