package projects

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/k8s"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DefaultProjectManagerFactory implements ProjectManagerFactory
type DefaultProjectManagerFactory struct {
	clientset      kubernetes.Interface
	dynamicClient  dynamic.Interface
	config         *rest.Config
	kubeconfigPath string
	detector       *k8s.ClusterTypeDetector
}

// NewProjectManagerFactory creates a new factory for project managers
func NewProjectManagerFactory(clientset kubernetes.Interface, config *rest.Config, kubeconfigPath string) (*DefaultProjectManagerFactory, error) {
	// Create dynamic client for OpenShift project operations
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create cluster type detector
	detector, err := k8s.NewClusterTypeDetector(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster detector: %w", err)
	}

	// Ensure kubeconfig path is set
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
	}

	return &DefaultProjectManagerFactory{
		clientset:      clientset,
		dynamicClient:  dynamicClient,
		config:         config,
		kubeconfigPath: kubeconfigPath,
		detector:       detector,
	}, nil
}

// CreateManager creates a project manager for the specified cluster type
func (f *DefaultProjectManagerFactory) CreateManager(ctx context.Context, clusterType k8s.ClusterType) (ProjectManager, error) {
	switch clusterType {
	case k8s.ClusterTypeOpenShift:
		return NewOpenShiftProjectManager(
			f.clientset,
			f.dynamicClient,
			f.config,
			f.kubeconfigPath,
		), nil

	case k8s.ClusterTypeKubernetes:
		return NewKubernetesNamespaceManager(
			f.clientset,
			f.config,
			f.kubeconfigPath,
		), nil

	default:
		return nil, fmt.Errorf("unsupported cluster type: %s", clusterType)
	}
}

// CreateAutoDetectManager auto-detects cluster type and creates the appropriate manager
func (f *DefaultProjectManagerFactory) CreateAutoDetectManager(ctx context.Context) (ProjectManager, error) {
	// Detect cluster type
	clusterInfo, err := f.detector.DetectClusterType(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect cluster type: %w", err)
	}

	// Create appropriate manager
	manager, err := f.CreateManager(ctx, clusterInfo.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}
	return manager, nil
}

// CreateManagerForCurrentCluster creates a manager for the currently connected cluster
func (f *DefaultProjectManagerFactory) CreateManagerForCurrentCluster(ctx context.Context) (ProjectManager, error) {
	return f.CreateAutoDetectManager(ctx)
}

// GetDetector returns the cluster type detector
func (f *DefaultProjectManagerFactory) GetDetector() *k8s.ClusterTypeDetector {
	return f.detector
}

// SetKubeconfigPath sets the kubeconfig path for the factory
func (f *DefaultProjectManagerFactory) SetKubeconfigPath(path string) {
	f.kubeconfigPath = path
}

// GetKubeconfigPath returns the current kubeconfig path
func (f *DefaultProjectManagerFactory) GetKubeconfigPath() string {
	return f.kubeconfigPath
}

// Helper functions

// getDefaultKubeconfigPath returns the default kubeconfig path
func getDefaultKubeconfigPath() string {
	// Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Default to ~/.kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, constants.KubeConfigDir, constants.KubeConfigFile)
}

// Utility functions for common operations

// ListAllProjects lists projects/namespaces across all available managers
func ListAllProjects(ctx context.Context, factory ProjectManagerFactory) ([]ProjectInfo, error) {
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}

	return manager.List(ctx, ListOptions{
		IncludeQuotas: false, // Don't include quotas for bulk listing
		IncludeLimits: false,
	})
}

// GetCurrentProject gets the current project/namespace
func GetCurrentProject(ctx context.Context, factory ProjectManagerFactory) (*ProjectInfo, error) {
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}

	return manager.GetCurrent(ctx)
}

// SwitchProject switches to a different project/namespace
func SwitchProject(ctx context.Context, factory ProjectManagerFactory, projectName string) (*SwitchResult, error) {
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}

	return manager.SwitchTo(ctx, projectName)
}

// CreateProject creates a new project/namespace
func CreateProject(ctx context.Context, factory ProjectManagerFactory, name string, opts CreateOptions) (*ProjectInfo, error) {
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}

	return manager.Create(ctx, name, opts)
}

// DeleteProject deletes a project/namespace
func DeleteProject(ctx context.Context, factory ProjectManagerFactory, name string) error {
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		return fmt.Errorf("failed to create project manager: %w", err)
	}

	return manager.Delete(ctx, name)
}

// ProjectExists checks if a project/namespace exists
func ProjectExists(ctx context.Context, factory ProjectManagerFactory, name string) (bool, error) {
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create project manager: %w", err)
	}

	return manager.Exists(ctx, name)
}
