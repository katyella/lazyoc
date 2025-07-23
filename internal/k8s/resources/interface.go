package resources

import (
	"context"
)

// ResourceClient defines the interface for resource operations
type ResourceClient interface {
	// Pod operations
	ListPods(ctx context.Context, opts ListOptions) (*ResourceList[PodInfo], error)
	GetPod(ctx context.Context, namespace, name string) (*PodInfo, error)

	// Service operations
	ListServices(ctx context.Context, opts ListOptions) (*ResourceList[ServiceInfo], error)
	GetService(ctx context.Context, namespace, name string) (*ServiceInfo, error)

	// Deployment operations
	ListDeployments(ctx context.Context, opts ListOptions) (*ResourceList[DeploymentInfo], error)
	GetDeployment(ctx context.Context, namespace, name string) (*DeploymentInfo, error)

	// ConfigMap operations
	ListConfigMaps(ctx context.Context, opts ListOptions) (*ResourceList[ConfigMapInfo], error)
	GetConfigMap(ctx context.Context, namespace, name string) (*ConfigMapInfo, error)

	// Secret operations
	ListSecrets(ctx context.Context, opts ListOptions) (*ResourceList[SecretInfo], error)
	GetSecret(ctx context.Context, namespace, name string) (*SecretInfo, error)

	// Project/Namespace operations (unified interface)
	ListProjects(ctx context.Context) (*ResourceList[ProjectInfo], error)
	GetCurrentProject() string
	SetCurrentProject(project string) error
	GetProjectContext() (*ProjectContext, error)
	SwitchToProject(ctx context.Context, project string) error

	// Legacy namespace operations (for backward compatibility)
	ListNamespaces(ctx context.Context) (*ResourceList[NamespaceInfo], error)
	GetCurrentNamespace() string
	SetCurrentNamespace(namespace string) error
	GetNamespaceContext() (*NamespaceContext, error)

	// Connection management
	TestConnection(ctx context.Context) error
	GetServerInfo(ctx context.Context) (map[string]interface{}, error)
}

// ResourceManager manages resource operations with error handling and retry logic
type ResourceManager struct {
	client     ResourceClient
	retryCount int
}

// NewResourceManager creates a new resource manager
func NewResourceManager(client ResourceClient) *ResourceManager {
	return &ResourceManager{
		client:     client,
		retryCount: 3,
	}
}

// SetRetryCount sets the number of retries for failed operations
func (rm *ResourceManager) SetRetryCount(count int) {
	rm.retryCount = count
}

// GetClient returns the underlying resource client
func (rm *ResourceManager) GetClient() ResourceClient {
	return rm.client
}