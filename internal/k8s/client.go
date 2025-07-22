package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientFactory creates and manages Kubernetes clients
type ClientFactory struct {
	config     *rest.Config
	clientset  *kubernetes.Clientset
	kubeconfig string
}

// NewClientFactory creates a new client factory
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// Initialize sets up the Kubernetes client configuration
func (cf *ClientFactory) Initialize() error {
	// Try to get kubeconfig path
	kubeconfigPath := cf.getKubeconfigPath()
	
	// Load the kubeconfig
	config, err := cf.loadKubeconfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	
	cf.config = config
	cf.kubeconfig = kubeconfigPath
	
	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	
	cf.clientset = clientset
	
	return nil
}

// GetClientset returns the Kubernetes clientset
func (cf *ClientFactory) GetClientset() *kubernetes.Clientset {
	return cf.clientset
}

// GetConfig returns the Kubernetes rest config
func (cf *ClientFactory) GetConfig() *rest.Config {
	return cf.config
}

// SetClientset sets the clientset directly (for external initialization)
func (cf *ClientFactory) SetClientset(clientset *kubernetes.Clientset) {
	cf.clientset = clientset
}

// SetConfig sets the rest config directly (for external initialization)
func (cf *ClientFactory) SetConfig(config *rest.Config) {
	cf.config = config
}

// TestConnection tests the connection to the Kubernetes cluster
func (cf *ClientFactory) TestConnection(ctx context.Context) error {
	if cf.clientset == nil {
		return fmt.Errorf("clientset not initialized")
	}
	
	// Try to get server version as a connectivity test
	_, err := cf.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes cluster: %w", err)
	}
	
	return nil
}

// GetCurrentContext returns the current context name
func (cf *ClientFactory) GetCurrentContext() (string, error) {
	if cf.kubeconfig == "" {
		return "", fmt.Errorf("kubeconfig not loaded")
	}
	
	config, err := clientcmd.LoadFromFile(cf.kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	
	return config.CurrentContext, nil
}

// GetCurrentNamespace returns the current namespace
func (cf *ClientFactory) GetCurrentNamespace() (string, error) {
	if cf.kubeconfig == "" {
		return "", fmt.Errorf("kubeconfig not loaded")
	}
	
	config, err := clientcmd.LoadFromFile(cf.kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	
	context := config.Contexts[config.CurrentContext]
	if context == nil {
		return "default", nil
	}
	
	if context.Namespace == "" {
		return "default", nil
	}
	
	return context.Namespace, nil
}

// getKubeconfigPath determines the kubeconfig file path
func (cf *ClientFactory) getKubeconfigPath() string {
	// Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}
	
	// Default to ~/.kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	return filepath.Join(home, ".kube", "config")
}

// loadKubeconfig loads the Kubernetes configuration
func (cf *ClientFactory) loadKubeconfig(kubeconfigPath string) (*rest.Config, error) {
	// First try in-cluster config (for running inside k8s)
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}
	
	// Fall back to kubeconfig file
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("no kubeconfig path found")
	}
	
	// Check if file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
	}
	
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}
	
	return config, nil
}