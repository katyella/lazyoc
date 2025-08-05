package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/client-go/apps/clientset/versioned"
	buildclientset "github.com/openshift/client-go/build/clientset/versioned"
	imageclientset "github.com/openshift/client-go/image/clientset/versioned"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/katyella/lazyoc/internal/constants"
)

// ClientFactory creates and manages Kubernetes clients
type ClientFactory struct {
	config      *rest.Config
	clientset   *kubernetes.Clientset
	kubeconfig  string
	isOpenShift bool

	// OpenShift clients
	appsClient    versioned.Interface
	buildClient   buildclientset.Interface
	imageClient   imageclientset.Interface
	routeClient   routeclientset.Interface
	dynamicClient dynamic.Interface
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

	// Detect if this is an OpenShift cluster and initialize OpenShift clients
	if err := cf.initializeOpenShiftClients(); err != nil {
		// Log the error but don't fail - this might just be a regular Kubernetes cluster
		// In a real implementation, you might want to log this
		cf.isOpenShift = false
	}

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

// InitializeOpenShiftAfterSetup initializes OpenShift clients after external setup
func (cf *ClientFactory) InitializeOpenShiftAfterSetup() error {
	if cf.config == nil || cf.clientset == nil {
		return fmt.Errorf("config and clientset must be set before initializing OpenShift clients")
	}

	return cf.initializeOpenShiftClients()
}

// TestConnection tests the connection to the Kubernetes cluster
func (cf *ClientFactory) TestConnection(ctx context.Context) error {
	if cf.clientset == nil {
		return fmt.Errorf(constants.ErrClientNotInitialized)
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
		return "", fmt.Errorf(constants.ErrKubeconfigNotLoaded)
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
		return "", fmt.Errorf(constants.ErrKubeconfigNotLoaded)
	}

	config, err := clientcmd.LoadFromFile(cf.kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	context := config.Contexts[config.CurrentContext]
	if context == nil {
		return constants.DefaultNamespace, nil
	}

	if context.Namespace == "" {
		return constants.DefaultNamespace, nil
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

	return filepath.Join(home, constants.KubeConfigDir, constants.KubeConfigFile)
}

// loadKubeconfig loads the Kubernetes configuration
func (cf *ClientFactory) loadKubeconfig(kubeconfigPath string) (*rest.Config, error) {
	// First try in-cluster config (for running inside k8s)
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	if kubeconfigPath == "" {
		return nil, fmt.Errorf(constants.ErrNoKubeconfigPath)
	}

	// Check if file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s: %w", kubeconfigPath, err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// initializeOpenShiftClients detects OpenShift and initializes OpenShift clients
func (cf *ClientFactory) initializeOpenShiftClients() error {
	if cf.config == nil {
		return fmt.Errorf("kubernetes config not initialized")
	}

	// Try to detect OpenShift by checking for OpenShift API groups
	discovery := cf.clientset.Discovery()
	groups, err := discovery.ServerGroups()
	if err != nil {
		return fmt.Errorf("failed to get server groups: %w", err)
	}

	// Check if OpenShift API groups are present
	hasOpenShiftGroups := false
	for _, group := range groups.Groups {
		if group.Name == "apps.openshift.io" ||
			group.Name == "build.openshift.io" ||
			group.Name == "image.openshift.io" ||
			group.Name == "route.openshift.io" {
			hasOpenShiftGroups = true
			break
		}
	}

	if !hasOpenShiftGroups {
		cf.isOpenShift = false
		return fmt.Errorf("not an OpenShift cluster")
	}

	// Initialize OpenShift clients
	appsClient, err := versioned.NewForConfig(cf.config)
	if err != nil {
		return fmt.Errorf("failed to create OpenShift apps client: %w", err)
	}
	cf.appsClient = appsClient

	buildClient, err := buildclientset.NewForConfig(cf.config)
	if err != nil {
		return fmt.Errorf("failed to create OpenShift build client: %w", err)
	}
	cf.buildClient = buildClient

	imageClient, err := imageclientset.NewForConfig(cf.config)
	if err != nil {
		return fmt.Errorf("failed to create OpenShift image client: %w", err)
	}
	cf.imageClient = imageClient

	routeClient, err := routeclientset.NewForConfig(cf.config)
	if err != nil {
		return fmt.Errorf("failed to create OpenShift route client: %w", err)
	}
	cf.routeClient = routeClient

	dynamicClient, err := dynamic.NewForConfig(cf.config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	cf.dynamicClient = dynamicClient

	cf.isOpenShift = true
	return nil
}

// OpenShift interface implementations

// IsOpenShift returns true if this is an OpenShift cluster
func (cf *ClientFactory) IsOpenShift() bool {
	return cf.isOpenShift
}

// GetAppsClient returns the OpenShift apps client
func (cf *ClientFactory) GetAppsClient() versioned.Interface {
	return cf.appsClient
}

// GetBuildClient returns the OpenShift build client
func (cf *ClientFactory) GetBuildClient() buildclientset.Interface {
	return cf.buildClient
}

// GetImageClient returns the OpenShift image client
func (cf *ClientFactory) GetImageClient() imageclientset.Interface {
	return cf.imageClient
}

// GetRouteClient returns the OpenShift route client
func (cf *ClientFactory) GetRouteClient() routeclientset.Interface {
	return cf.routeClient
}

// GetDynamicClient returns the dynamic client
func (cf *ClientFactory) GetDynamicClient() dynamic.Interface {
	return cf.dynamicClient
}
