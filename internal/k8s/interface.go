package k8s

import (
	"context"

	"github.com/openshift/client-go/apps/clientset/versioned"
	buildclientset "github.com/openshift/client-go/build/clientset/versioned"
	imageclientset "github.com/openshift/client-go/image/clientset/versioned"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned"
	"k8s.io/client-go/dynamic"
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

// OpenShiftClient extends Client interface with OpenShift-specific operations
type OpenShiftClient interface {
	Client

	// OpenShift detection
	IsOpenShift() bool

	// OpenShift clientsets
	GetAppsClient() versioned.Interface
	GetBuildClient() buildclientset.Interface
	GetImageClient() imageclientset.Interface
	GetRouteClient() routeclientset.Interface
	GetDynamicClient() dynamic.Interface
}

// Ensure ClientFactory implements Client interface
var _ Client = (*ClientFactory)(nil)

// Ensure ClientFactory implements OpenShiftClient interface
var _ OpenShiftClient = (*ClientFactory)(nil)
