package services

import (
	"context"
	"fmt"
	"time"

	"github.com/katyella/lazyoc/internal/k8s"
	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// KubernetesService provides a high-level interface to Kubernetes operations
type KubernetesService struct {
	clientFactory  *k8s.ClientFactory
	resourceClient *resources.K8sResourceClient
	namespace      string
	context        string
}

// NewKubernetesService creates a new Kubernetes service
func NewKubernetesService() *KubernetesService {
	return &KubernetesService{
		clientFactory: k8s.NewClientFactory(),
	}
}

// Connect connects to the Kubernetes cluster
func (k *KubernetesService) Connect(kubeconfig, contextName string) error {
	// Initialize client factory
	if err := k.clientFactory.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := k.clientFactory.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Get clientset
	clientset := k.clientFactory.GetClientset()
	
	// Get current namespace
	ns, err := k.clientFactory.GetCurrentNamespace()
	if err != nil {
		ns = "default"
	}
	k.namespace = ns

	// Get current context
	currentContext, err := k.clientFactory.GetCurrentContext()
	if err == nil {
		k.context = currentContext
	}

	// Create resource client with config for exec operations
	k.resourceClient = resources.NewK8sResourceClientWithConfig(clientset, k.clientFactory.GetConfig(), k.namespace)

	return nil
}

// Disconnect disconnects from the cluster
func (k *KubernetesService) Disconnect() {
	k.clientFactory = k8s.NewClientFactory()
	k.resourceClient = nil
	k.namespace = ""
	k.context = ""
}

// IsConnected checks if connected to a cluster
func (k *KubernetesService) IsConnected() bool {
	return k.resourceClient != nil
}

// GetClusterInfo returns cluster information
func (k *KubernetesService) GetClusterInfo() (*ClusterInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	// Detect cluster type
	detector, err := k8s.NewClusterTypeDetector(k.clientFactory.GetConfig())
	if err != nil {
		return nil, err
	}
	clusterInfo, err := detector.DetectClusterType(context.Background())
	if err != nil {
		return nil, err
	}

	// Get version info
	versionInfo, err := k.clientFactory.GetClientset().Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	return &ClusterInfo{
		Type:      clusterInfo.Type.String(),
		Version:   versionInfo.String(),
		Context:   k.context,
		Namespace: k.namespace,
	}, nil
}

// SetNamespace sets the current namespace
func (k *KubernetesService) SetNamespace(namespace string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to cluster")
	}

	// Verify namespace exists
	ctx := context.Background()
	namespaces, err := k.resourceClient.ListNamespaces(ctx)
	if err != nil {
		return err
	}

	found := false
	for _, ns := range namespaces.Items {
		if ns.Name == namespace {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("namespace %s not found", namespace)
	}

	k.namespace = namespace
	// Recreate resource client with new namespace
	k.resourceClient = resources.NewK8sResourceClientWithConfig(k.clientFactory.GetClientset(), k.clientFactory.GetConfig(), k.namespace)
	return nil
}

// GetCurrentNamespace returns the current namespace
func (k *KubernetesService) GetCurrentNamespace() string {
	return k.namespace
}

// GetNamespaces returns all namespaces
func (k *KubernetesService) GetNamespaces() ([]string, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	namespaces, err := k.resourceClient.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(namespaces.Items))
	for i, ns := range namespaces.Items {
		result[i] = ns.Name
	}

	return result, nil
}

// GetPods returns pods in the current namespace
func (k *KubernetesService) GetPods() ([]resources.PodInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	opts := resources.ListOptions{
		Namespace: k.namespace,
	}

	podList, err := k.resourceClient.ListPods(ctx, opts)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

// GetPod returns a specific pod
func (k *KubernetesService) GetPod(name string) (*resources.PodInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	return k.resourceClient.GetPod(ctx, k.namespace, name)
}

// GetServices returns services in the current namespace
func (k *KubernetesService) GetServices() ([]resources.ServiceInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	opts := resources.ListOptions{
		Namespace: k.namespace,
	}

	svcList, err := k.resourceClient.ListServices(ctx, opts)
	if err != nil {
		return nil, err
	}

	return svcList.Items, nil
}

// GetService returns a specific service
func (k *KubernetesService) GetService(name string) (*resources.ServiceInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	return k.resourceClient.GetService(ctx, k.namespace, name)
}

// GetDeployments returns deployments in the current namespace
func (k *KubernetesService) GetDeployments() ([]resources.DeploymentInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	opts := resources.ListOptions{
		Namespace: k.namespace,
	}

	depList, err := k.resourceClient.ListDeployments(ctx, opts)
	if err != nil {
		return nil, err
	}

	return depList.Items, nil
}

// GetDeployment returns a specific deployment
func (k *KubernetesService) GetDeployment(name string) (*resources.DeploymentInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	return k.resourceClient.GetDeployment(ctx, k.namespace, name)
}

// GetConfigMaps returns config maps in the current namespace
func (k *KubernetesService) GetConfigMaps() ([]resources.ConfigMapInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	opts := resources.ListOptions{
		Namespace: k.namespace,
	}

	cmList, err := k.resourceClient.ListConfigMaps(ctx, opts)
	if err != nil {
		return nil, err
	}

	return cmList.Items, nil
}

// GetSecrets returns secrets in the current namespace
func (k *KubernetesService) GetSecrets() ([]resources.SecretInfo, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	opts := resources.ListOptions{
		Namespace: k.namespace,
	}

	secretList, err := k.resourceClient.ListSecrets(ctx, opts)
	if err != nil {
		return nil, err
	}

	return secretList.Items, nil
}

// DeleteResource deletes a resource
func (k *KubernetesService) DeleteResource(resourceType, name string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	
	switch resourceType {
	case "pod", "pods":
		return k.resourceClient.DeletePod(ctx, k.namespace, name)
	case "service", "services", "svc":
		return k.resourceClient.DeleteService(ctx, k.namespace, name)
	case "deployment", "deployments", "deploy":
		return k.resourceClient.DeleteDeployment(ctx, k.namespace, name)
	case "configmap", "configmaps", "cm":
		return k.resourceClient.DeleteConfigMap(ctx, k.namespace, name)
	case "secret", "secrets":
		return k.resourceClient.DeleteSecret(ctx, k.namespace, name)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// ExecInPod executes a command in a pod
func (k *KubernetesService) ExecInPod(podName, containerName string, command []string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	
	// For now, we'll execute without stdin/stdout/stderr redirection
	// In a real TUI, you'd want to handle these streams properly
	opts := resources.ExecOptions{
		Namespace:     k.namespace,
		PodName:       podName,
		ContainerName: containerName,
		Command:       command,
		TTY:           false,
	}
	
	return k.resourceClient.ExecuteInPod(ctx, opts)
}

// GetPodLogs returns logs for a pod
func (k *KubernetesService) GetPodLogs(podName, containerName string, lines int64, follow bool) (string, error) {
	if !k.IsConnected() {
		return "", fmt.Errorf("not connected to cluster")
	}

	ctx := context.Background()
	opts := resources.LogOptions{
		TailLines: &lines,
		Follow:    follow,
	}

	logs, err := k.resourceClient.GetPodLogs(ctx, k.namespace, podName, containerName, opts)
	if err != nil {
		return "", err
	}

	return logs, nil
}

// StreamPodLogs streams logs from a pod
func (k *KubernetesService) StreamPodLogs(ctx context.Context, podName, containerName string, lines int64) (<-chan string, <-chan error) {
	logChan := make(chan string, 100)
	errChan := make(chan error, 1)

	if !k.IsConnected() {
		errChan <- fmt.Errorf("not connected to cluster")
		close(errChan)
		close(logChan)
		return logChan, errChan
	}

	go func() {
		defer close(logChan)
		defer close(errChan)

		opts := resources.LogOptions{
			TailLines: &lines,
			Follow:    true,
		}

		logStream, err := k.resourceClient.StreamPodLogs(ctx, k.namespace, podName, containerName, opts)
		if err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-logStream:
				if !ok {
					return
				}
				logChan <- line
			}
		}
	}()

	return logChan, errChan
}

// WatchResources watches for resource changes
func (k *KubernetesService) WatchResources(ctx context.Context, resourceType string) (<-chan ResourceEvent, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to cluster")
	}

	// Create event channel
	events := make(chan ResourceEvent, 100)

	// Start watching based on resource type
	switch resourceType {
	case "pods":
		go k.watchPods(ctx, events)
	case "services":
		go k.watchServices(ctx, events)
	case "deployments":
		go k.watchDeployments(ctx, events)
	default:
		close(events)
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return events, nil
}

// watchPods watches for pod changes
func (k *KubernetesService) watchPods(ctx context.Context, events chan<- ResourceEvent) {
	defer close(events)

	// Implementation would use k8s watch API
	// For now, this is a placeholder
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// In real implementation, this would receive actual watch events
			// For now, we just poll
		}
	}
}

// watchServices watches for service changes
func (k *KubernetesService) watchServices(ctx context.Context, events chan<- ResourceEvent) {
	defer close(events)

	// Similar placeholder implementation
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Poll for changes
		}
	}
}

// watchDeployments watches for deployment changes
func (k *KubernetesService) watchDeployments(ctx context.Context, events chan<- ResourceEvent) {
	defer close(events)

	// Similar placeholder implementation
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Poll for changes
		}
	}
}

// Types

// ClusterInfo contains cluster information
type ClusterInfo struct {
	Type      string
	Version   string
	Context   string
	Namespace string
}

// ResourceEvent represents a resource change event
type ResourceEvent struct {
	Type      EventType
	Resource  interface{}
	Timestamp time.Time
}

// EventType represents the type of resource event
type EventType string

const (
	EventAdded    EventType = "ADDED"
	EventModified EventType = "MODIFIED"
	EventDeleted  EventType = "DELETED"
)