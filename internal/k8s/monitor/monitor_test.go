package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/katyella/lazyoc/internal/k8s/auth"
	"github.com/katyella/lazyoc/internal/k8s/resources"
	"k8s.io/client-go/rest"
)

// mockAuthProvider is a mock implementation of AuthProvider for testing
type mockAuthProvider struct {
	namespace string
	context   string
}

func (m *mockAuthProvider) Authenticate(ctx context.Context) (*rest.Config, error) {
	// Return a fake config that won't be used for real connections
	return &rest.Config{
		Host:            "https://mock-server:6443",
		BearerToken:     "mock-token",
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}, nil
}

func (m *mockAuthProvider) IsValid(ctx context.Context) error {
	return nil // Always valid in mock
}

func (m *mockAuthProvider) GetContext() string {
	return m.context
}

func (m *mockAuthProvider) GetNamespace() string {
	return m.namespace
}

func (m *mockAuthProvider) GetAvailableContexts() ([]string, error) {
	return []string{m.context}, nil
}

func (m *mockAuthProvider) Refresh(ctx context.Context) error {
	return nil // Always successful in mock
}

// mockResourceClient is a mock implementation of ResourceClient for testing
type mockResourceClient struct {
	namespace string
}

func (m *mockResourceClient) TestConnection(ctx context.Context) error {
	return nil // Always successful in mock
}

func (m *mockResourceClient) GetServerInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"version":    "v1.31.0",
		"gitVersion": "v1.31.0",
	}, nil
}

// Pod operations
func (m *mockResourceClient) ListPods(ctx context.Context, opts resources.ListOptions) (*resources.ResourceList[resources.PodInfo], error) {
	return &resources.ResourceList[resources.PodInfo]{}, nil
}

func (m *mockResourceClient) GetPod(ctx context.Context, namespace, name string) (*resources.PodInfo, error) {
	return &resources.PodInfo{}, nil
}

// Service operations
func (m *mockResourceClient) ListServices(ctx context.Context, opts resources.ListOptions) (*resources.ResourceList[resources.ServiceInfo], error) {
	return &resources.ResourceList[resources.ServiceInfo]{}, nil
}

func (m *mockResourceClient) GetService(ctx context.Context, namespace, name string) (*resources.ServiceInfo, error) {
	return &resources.ServiceInfo{}, nil
}

// Deployment operations
func (m *mockResourceClient) ListDeployments(ctx context.Context, opts resources.ListOptions) (*resources.ResourceList[resources.DeploymentInfo], error) {
	return &resources.ResourceList[resources.DeploymentInfo]{}, nil
}

func (m *mockResourceClient) GetDeployment(ctx context.Context, namespace, name string) (*resources.DeploymentInfo, error) {
	return &resources.DeploymentInfo{}, nil
}

// ConfigMap operations
func (m *mockResourceClient) ListConfigMaps(ctx context.Context, opts resources.ListOptions) (*resources.ResourceList[resources.ConfigMapInfo], error) {
	return &resources.ResourceList[resources.ConfigMapInfo]{}, nil
}

func (m *mockResourceClient) GetConfigMap(ctx context.Context, namespace, name string) (*resources.ConfigMapInfo, error) {
	return &resources.ConfigMapInfo{}, nil
}

// Secret operations
func (m *mockResourceClient) ListSecrets(ctx context.Context, opts resources.ListOptions) (*resources.ResourceList[resources.SecretInfo], error) {
	return &resources.ResourceList[resources.SecretInfo]{}, nil
}

func (m *mockResourceClient) GetSecret(ctx context.Context, namespace, name string) (*resources.SecretInfo, error) {
	return &resources.SecretInfo{}, nil
}

// Project operations
func (m *mockResourceClient) ListProjects(ctx context.Context) (*resources.ResourceList[resources.ProjectInfo], error) {
	return &resources.ResourceList[resources.ProjectInfo]{}, nil
}

func (m *mockResourceClient) GetCurrentProject() string {
	return m.namespace
}

func (m *mockResourceClient) SetCurrentProject(project string) error {
	m.namespace = project
	return nil
}

func (m *mockResourceClient) GetProjectContext() (*resources.ProjectContext, error) {
	return &resources.ProjectContext{}, nil
}

func (m *mockResourceClient) SwitchToProject(ctx context.Context, project string) error {
	m.namespace = project
	return nil
}

// Legacy namespace operations
func (m *mockResourceClient) ListNamespaces(ctx context.Context) (*resources.ResourceList[resources.NamespaceInfo], error) {
	return &resources.ResourceList[resources.NamespaceInfo]{}, nil
}

func (m *mockResourceClient) GetCurrentNamespace() string {
	return m.namespace
}

func (m *mockResourceClient) SetCurrentNamespace(namespace string) error {
	m.namespace = namespace
	return nil
}

func (m *mockResourceClient) GetNamespaceContext() (*resources.NamespaceContext, error) {
	return &resources.NamespaceContext{}, nil
}

// Pod log operations
func (m *mockResourceClient) GetPodLogs(ctx context.Context, namespace, podName, containerName string, opts resources.LogOptions) (string, error) {
	return "mock logs", nil
}

func (m *mockResourceClient) StreamPodLogs(ctx context.Context, namespace, podName, containerName string, opts resources.LogOptions) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "mock log stream"
	close(ch)
	return ch, nil
}

func getTestComponents(t *testing.T) (auth.AuthProvider, resources.ResourceClient) {
	// Return mock components instead of real ones
	authProvider := &mockAuthProvider{
		namespace: "test-namespace",
		context:   "test-context",
	}

	resourceClient := &mockResourceClient{
		namespace: "test-namespace",
	}

	return authProvider, resourceClient
}

func TestConnectionStatus_String(t *testing.T) {
	tests := []struct {
		status   ConnectionStatus
		expected string
	}{
		{StatusConnected, "Connected"},
		{StatusDisconnected, "Disconnected"},
		{StatusConnecting, "Connecting"},
		{StatusReconnecting, "Reconnecting"},
		{StatusError, "Error"},
		{StatusUnknown, "Unknown"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestConnectionEventType_String(t *testing.T) {
	tests := []struct {
		eventType ConnectionEventType
		expected  string
	}{
		{EventConnected, "Connected"},
		{EventDisconnected, "Disconnected"},
		{EventReconnecting, "Reconnecting"},
		{EventError, "Error"},
		{EventHealthCheck, "HealthCheck"},
	}

	for _, test := range tests {
		result := test.eventType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestK8sConnectionMonitor_Creation(t *testing.T) {
	authProvider, resourceClient := getTestComponents(t)

	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	if monitor == nil {
		t.Fatal("Expected non-nil monitor")
	}

	status := monitor.GetStatus()
	if status.Status != StatusUnknown {
		t.Errorf("Expected initial status to be Unknown, got %s", status.Status)
	}

	if monitor.IsHealthy() {
		t.Error("Expected monitor to not be healthy initially")
	}
}

func TestK8sConnectionMonitor_StartStop(t *testing.T) {
	t.Skip("Disabled: This test deadlocks when connecting to real cluster")

	authProvider, resourceClient := getTestComponents(t)
	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start monitoring
	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	// Give it a moment to attempt connection
	time.Sleep(2 * time.Second)

	// Check that it started
	status := monitor.GetStatus()
	if status.Status == StatusUnknown {
		t.Error("Expected status to change from Unknown after start")
	}

	// Get events
	events := monitor.GetEvents(10)
	if len(events) == 0 {
		t.Error("Expected at least one event after start")
	}

	// Stop monitoring
	monitor.Stop()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	// Check final status
	finalStatus := monitor.GetStatus()
	if finalStatus.Status != StatusDisconnected {
		t.Errorf("Expected final status to be Disconnected, got %s", finalStatus.Status)
	}
}

func TestK8sConnectionMonitor_HealthCheck(t *testing.T) {
	t.Skip("Disabled: Health check may attempt to connect to cluster")
}

func TestK8sConnectionMonitor_Metrics(t *testing.T) {
	authProvider, resourceClient := getTestComponents(t)
	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	metrics := monitor.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	// Initial metrics should be zero
	if metrics.RequestCount != 0 {
		t.Errorf("Expected initial request count to be 0, got %d", metrics.RequestCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected initial error count to be 0, got %d", metrics.ErrorCount)
	}
}

func TestK8sConnectionMonitor_Events(t *testing.T) {
	t.Skip("Disabled: This test may hang when connecting to real cluster")

	authProvider, resourceClient := getTestComponents(t)
	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	// Initially no events
	events := monitor.GetEvents(10)
	if len(events) != 0 {
		t.Errorf("Expected no initial events, got %d", len(events))
	}

	// Start monitor to generate events
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()

	// Give it time to generate events
	time.Sleep(1 * time.Second)

	events = monitor.GetEvents(10)
	if len(events) == 0 {
		t.Error("Expected at least one event after starting")
	}

	// Test event listener
	var receivedEvent *ConnectionEvent
	monitor.AddEventListener(func(event ConnectionEvent) {
		receivedEvent = &event
	})

	// Force a health check to generate an event
	monitor.ForceHealthCheck(ctx)

	// Give listener time to be called
	time.Sleep(100 * time.Millisecond)

	// Note: receivedEvent might be nil if health check doesn't generate an event
	// This is acceptable as it depends on cluster connectivity
	t.Logf("Event listener test completed (received event: %v)", receivedEvent != nil)
}

func TestK8sConnectionMonitor_Reconnect(t *testing.T) {
	t.Skip("Disabled: Reconnect may attempt to connect to cluster")
}

func TestK8sConnectionMonitor_EventLimiting(t *testing.T) {
	authProvider, resourceClient := getTestComponents(t)
	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	// Test getting more events than exist
	events := monitor.GetEvents(1000)
	if len(events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events))
	}

	// Test getting negative number of events
	events = monitor.GetEvents(-1)
	if len(events) != 0 {
		t.Errorf("Expected 0 events for negative limit, got %d", len(events))
	}
}

func TestK8sConnectionMonitor_DoubleStart(t *testing.T) {
	t.Skip("Disabled: This test may hang when connecting to real cluster")

	authProvider, resourceClient := getTestComponents(t)
	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start monitor
	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()

	// Try to start again
	err = monitor.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting monitor twice")
	}
}

func TestK8sConnectionMonitor_StopBeforeStart(t *testing.T) {
	authProvider, resourceClient := getTestComponents(t)
	monitor := NewK8sConnectionMonitor(authProvider, resourceClient)

	// Stop without starting (should not panic)
	monitor.Stop()

	status := monitor.GetStatus()
	if status.Status != StatusUnknown {
		t.Errorf("Expected status to remain Unknown, got %s", status.Status)
	}
}
