package k8s

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client defines the interface for Kubernetes client operations
type Client interface {
	// Initialization
	Initialize() error
	TestConnection(ctx context.Context) error
	
	// Client access
	GetClientset() *kubernetes.Clientset
	GetConfig() *rest.Config
	
	// Context and namespace information
	GetCurrentContext() (string, error)
	GetCurrentNamespace() (string, error)
}

// Ensure ClientFactory implements Client interface
var _ Client = (*ClientFactory)(nil)