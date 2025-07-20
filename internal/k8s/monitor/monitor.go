package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/katyella/lazyoc/internal/k8s/auth"
	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// K8sConnectionMonitor implements ConnectionMonitor for Kubernetes
type K8sConnectionMonitor struct {
	authProvider   auth.AuthProvider
	resourceClient resources.ResourceClient
	config         MonitorConfig
	
	// State
	mu          sync.RWMutex
	status      ConnectionInfo
	metrics     Metrics
	events      []ConnectionEvent
	listeners   []func(ConnectionEvent)
	
	// Control
	ctx        context.Context
	cancel     context.CancelFunc
	started    bool
	healthTicker *time.Ticker
	startTime    time.Time
}

// NewK8sConnectionMonitor creates a new Kubernetes connection monitor
func NewK8sConnectionMonitor(authProvider auth.AuthProvider, resourceClient resources.ResourceClient) *K8sConnectionMonitor {
	config := MonitorConfig{
		HealthCheckInterval: 30 * time.Second,
		RequestTimeout:      10 * time.Second,
		MaxEvents:          100,
		RetryAttempts:      3,
		RetryDelay:         5 * time.Second,
	}
	
	return &K8sConnectionMonitor{
		authProvider:   authProvider,
		resourceClient: resourceClient,
		config:         config,
		status: ConnectionInfo{
			Status: StatusUnknown,
		},
		events:    make([]ConnectionEvent, 0, config.MaxEvents),
		listeners: make([]func(ConnectionEvent), 0),
	}
}

// Start begins monitoring the connection
func (m *K8sConnectionMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.started {
		return fmt.Errorf("monitor already started")
	}
	
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.started = true
	m.startTime = time.Now()
	
	// Initial connection attempt
	go m.initialConnect()
	
	// Start health check ticker
	m.healthTicker = time.NewTicker(m.config.HealthCheckInterval)
	go m.healthCheckLoop()
	
	m.addEvent(ConnectionEvent{
		Type:      EventConnected,
		Timestamp: time.Now(),
		Message:   "Connection monitor started",
	})
	
	return nil
}

// Stop stops the monitoring
func (m *K8sConnectionMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.started {
		return
	}
	
	m.started = false
	if m.cancel != nil {
		m.cancel()
	}
	
	if m.healthTicker != nil {
		m.healthTicker.Stop()
	}
	
	m.status.Status = StatusDisconnected
	m.addEvent(ConnectionEvent{
		Type:      EventDisconnected,
		Timestamp: time.Now(),
		Message:   "Connection monitor stopped",
	})
}

// GetStatus returns the current connection status
func (m *K8sConnectionMonitor) GetStatus() *ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Update uptime
	if m.status.Status == StatusConnected {
		m.status.LastChecked = time.Now()
	}
	
	statusCopy := m.status
	return &statusCopy
}

// GetMetrics returns current metrics
func (m *K8sConnectionMonitor) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metricsCopy := m.metrics
	if m.started {
		metricsCopy.Uptime = time.Since(m.startTime)
	}
	
	return &metricsCopy
}

// GetEvents returns recent connection events
func (m *K8sConnectionMonitor) GetEvents(limit int) []ConnectionEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if limit <= 0 || limit > len(m.events) {
		limit = len(m.events)
	}
	
	// Return the most recent events
	start := len(m.events) - limit
	if start < 0 {
		start = 0
	}
	
	events := make([]ConnectionEvent, limit)
	copy(events, m.events[start:])
	
	return events
}

// ForceHealthCheck triggers an immediate health check
func (m *K8sConnectionMonitor) ForceHealthCheck(ctx context.Context) *HealthCheck {
	return m.performHealthCheck(ctx)
}

// Reconnect attempts to reconnect
func (m *K8sConnectionMonitor) Reconnect(ctx context.Context) error {
	m.mu.Lock()
	m.status.Status = StatusReconnecting
	m.mu.Unlock()
	
	m.addEvent(ConnectionEvent{
		Type:      EventReconnecting,
		Timestamp: time.Now(),
		Message:   "Manual reconnection triggered",
	})
	
	return m.attemptConnection(ctx)
}

// AddEventListener adds a listener for connection events
func (m *K8sConnectionMonitor) AddEventListener(listener func(ConnectionEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.listeners = append(m.listeners, listener)
}

// IsHealthy returns true if the connection is healthy
func (m *K8sConnectionMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.status.Status == StatusConnected
}

// Internal methods

func (m *K8sConnectionMonitor) initialConnect() {
	ctx, cancel := context.WithTimeout(m.ctx, m.config.RequestTimeout)
	defer cancel()
	
	m.mu.Lock()
	m.status.Status = StatusConnecting
	m.mu.Unlock()
	
	err := m.attemptConnection(ctx)
	if err != nil {
		m.addEvent(ConnectionEvent{
			Type:      EventError,
			Timestamp: time.Now(),
			Message:   "Initial connection failed",
			Error:     err.Error(),
		})
	}
}

func (m *K8sConnectionMonitor) healthCheckLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.healthTicker.C:
			ctx, cancel := context.WithTimeout(m.ctx, m.config.RequestTimeout)
			check := m.performHealthCheck(ctx)
			cancel()
			
			if !check.Success {
				m.addEvent(ConnectionEvent{
					Type:      EventError,
					Timestamp: time.Now(),
					Message:   "Health check failed",
					Error:     check.Error,
				})
				
				// Attempt reconnection
				go m.attemptReconnection()
			}
		}
	}
}

func (m *K8sConnectionMonitor) attemptConnection(ctx context.Context) error {
	start := time.Now()
	
	// Test authentication
	err := m.authProvider.IsValid(ctx)
	if err != nil {
		m.mu.Lock()
		m.status.Status = StatusError
		m.status.Error = err.Error()
		m.mu.Unlock()
		return fmt.Errorf("authentication failed: %w", err)
	}
	
	// Test connection
	err = m.resourceClient.TestConnection(ctx)
	if err != nil {
		m.mu.Lock()
		m.status.Status = StatusError
		m.status.Error = err.Error()
		m.mu.Unlock()
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	// Get server info
	serverInfo, err := m.resourceClient.GetServerInfo(ctx)
	if err != nil {
		m.mu.Lock()
		m.status.Status = StatusError
		m.status.Error = err.Error()
		m.mu.Unlock()
		return fmt.Errorf("failed to get server info: %w", err)
	}
	
	// Update status
	m.mu.Lock()
	m.status.Status = StatusConnected
	m.status.ConnectedAt = time.Now()
	m.status.LastChecked = time.Now()
	m.status.Context = m.authProvider.GetContext()
	m.status.Namespace = m.authProvider.GetNamespace()
	m.status.Error = ""
	
	if version, ok := serverInfo["version"].(string); ok {
		m.status.ServerVersion = version
	}
	
	// Simple OpenShift detection (would need more sophisticated logic)
	m.status.IsOpenShift = false
	
	// Update metrics
	m.metrics.TotalConnections++
	m.metrics.LastRequestTime = time.Now()
	latency := time.Since(start)
	
	// Simple moving average for latency
	if m.metrics.AverageLatency == 0 {
		m.metrics.AverageLatency = latency
	} else {
		m.metrics.AverageLatency = (m.metrics.AverageLatency + latency) / 2
	}
	
	m.mu.Unlock()
	
	m.addEvent(ConnectionEvent{
		Type:      EventConnected,
		Timestamp: time.Now(),
		Message:   "Successfully connected to cluster",
		Details: map[string]interface{}{
			"version": m.status.ServerVersion,
			"context": m.status.Context,
			"latency": latency.String(),
		},
	})
	
	return nil
}

func (m *K8sConnectionMonitor) performHealthCheck(ctx context.Context) *HealthCheck {
	start := time.Now()
	
	err := m.resourceClient.TestConnection(ctx)
	duration := time.Since(start)
	
	check := &HealthCheck{
		Timestamp: start,
		Duration:  duration,
		Success:   err == nil,
	}
	
	if err != nil {
		check.Error = err.Error()
		
		m.mu.Lock()
		m.status.Status = StatusError
		m.status.Error = err.Error()
		m.metrics.ErrorCount++
		m.mu.Unlock()
	} else {
		m.mu.Lock()
		m.status.Status = StatusConnected
		m.status.LastChecked = time.Now()
		m.status.Error = ""
		m.metrics.RequestCount++
		m.metrics.LastRequestTime = time.Now()
		
		// Update average latency
		if m.metrics.AverageLatency == 0 {
			m.metrics.AverageLatency = duration
		} else {
			m.metrics.AverageLatency = (m.metrics.AverageLatency + duration) / 2
		}
		m.mu.Unlock()
		
		check.Details = map[string]interface{}{
			"latency": duration.String(),
		}
	}
	
	return check
}

func (m *K8sConnectionMonitor) attemptReconnection() {
	m.mu.Lock()
	m.status.Status = StatusReconnecting
	m.mu.Unlock()
	
	for attempt := 1; attempt <= m.config.RetryAttempts; attempt++ {
		select {
		case <-m.ctx.Done():
			return
		default:
		}
		
		ctx, cancel := context.WithTimeout(m.ctx, m.config.RequestTimeout)
		err := m.attemptConnection(ctx)
		cancel()
		
		if err == nil {
			return // Successfully reconnected
		}
		
		if attempt < m.config.RetryAttempts {
			select {
			case <-m.ctx.Done():
				return
			case <-time.After(m.config.RetryDelay * time.Duration(attempt)):
				// Exponential backoff
			}
		}
	}
	
	// All attempts failed
	m.mu.Lock()
	m.status.Status = StatusError
	m.mu.Unlock()
	
	m.addEvent(ConnectionEvent{
		Type:      EventError,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Reconnection failed after %d attempts", m.config.RetryAttempts),
	})
}

func (m *K8sConnectionMonitor) addEvent(event ConnectionEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Add event to list
	m.events = append(m.events, event)
	
	// Trim events if we exceed max
	if len(m.events) > m.config.MaxEvents {
		m.events = m.events[len(m.events)-m.config.MaxEvents:]
	}
	
	// Notify listeners
	for _, listener := range m.listeners {
		go listener(event)
	}
}