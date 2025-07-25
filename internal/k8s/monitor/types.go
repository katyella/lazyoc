package monitor

import (
	"context"
	"time"
)

// ConnectionStatus represents the current connection status
type ConnectionStatus int

const (
	// StatusUnknown indicates the connection status is unknown or uninitialized
	StatusUnknown ConnectionStatus = iota
	
	// StatusConnected indicates an active, healthy connection to the cluster
	StatusConnected
	
	// StatusDisconnected indicates no connection to the cluster
	StatusDisconnected
	
	// StatusConnecting indicates a connection attempt is in progress
	StatusConnecting
	
	// StatusReconnecting indicates an attempt to restore a lost connection
	StatusReconnecting
	
	// StatusError indicates a connection error has occurred
	StatusError
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusConnected:
		return "Connected"
	case StatusDisconnected:
		return "Disconnected"
	case StatusConnecting:
		return "Connecting"
	case StatusReconnecting:
		return "Reconnecting"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ConnectionInfo contains information about the current connection
type ConnectionInfo struct {
	Status        ConnectionStatus `json:"status"`
	ClusterName   string           `json:"clusterName"`
	ServerVersion string           `json:"serverVersion"`
	Context       string           `json:"context"`
	Namespace     string           `json:"namespace"`
	Host          string           `json:"host"`
	IsOpenShift   bool             `json:"isOpenShift"`
	ConnectedAt   time.Time        `json:"connectedAt"`
	LastChecked   time.Time        `json:"lastChecked"`
	Error         string           `json:"error,omitempty"`
}

// Metrics contains connection and API metrics
type Metrics struct {
	RequestCount     int64         `json:"requestCount"`
	ErrorCount       int64         `json:"errorCount"`
	AverageLatency   time.Duration `json:"averageLatency"`
	LastRequestTime  time.Time     `json:"lastRequestTime"`
	TotalConnections int64         `json:"totalConnections"`
	Uptime           time.Duration `json:"uptime"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ConnectionEventType represents the type of connection event
type ConnectionEventType int

const (
	EventConnected ConnectionEventType = iota
	EventDisconnected
	EventReconnecting
	EventError
	EventHealthCheck
)

func (e ConnectionEventType) String() string {
	switch e {
	case EventConnected:
		return "Connected"
	case EventDisconnected:
		return "Disconnected"
	case EventReconnecting:
		return "Reconnecting"
	case EventError:
		return "Error"
	case EventHealthCheck:
		return "HealthCheck"
	default:
		return "Unknown"
	}
}

// ConnectionEvent represents a connection-related event
type ConnectionEvent struct {
	Type      ConnectionEventType    `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ConnectionMonitor defines the interface for connection monitoring
type ConnectionMonitor interface {
	// Start begins monitoring the connection
	Start(ctx context.Context) error

	// Stop stops the monitoring
	Stop()

	// GetStatus returns the current connection status
	GetStatus() *ConnectionInfo

	// GetMetrics returns current metrics
	GetMetrics() *Metrics

	// GetEvents returns recent connection events
	GetEvents(limit int) []ConnectionEvent

	// ForceHealthCheck triggers an immediate health check
	ForceHealthCheck(ctx context.Context) *HealthCheck

	// Reconnect attempts to reconnect
	Reconnect(ctx context.Context) error

	// AddEventListener adds a listener for connection events
	AddEventListener(listener func(ConnectionEvent))

	// IsHealthy returns true if the connection is healthy
	IsHealthy() bool
}

// MonitorConfig contains configuration for the connection monitor
type MonitorConfig struct {
	HealthCheckInterval time.Duration `json:"healthCheckInterval"`
	RequestTimeout      time.Duration `json:"requestTimeout"`
	MaxEvents           int           `json:"maxEvents"`
	RetryAttempts       int           `json:"retryAttempts"`
	RetryDelay          time.Duration `json:"retryDelay"`
}
