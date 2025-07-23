package auth

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/katyella/lazyoc/internal/constants"
)

// CredentialValidator provides methods to validate authentication credentials
type CredentialValidator struct {
	config    *rest.Config
	clientset *kubernetes.Clientset
}

// NewCredentialValidator creates a new credential validator
func NewCredentialValidator(config *rest.Config) (*CredentialValidator, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, NewAuthError(
			"clientset_creation_failed",
			"failed to create kubernetes clientset for validation",
			err,
		)
	}
	
	return &CredentialValidator{
		config:    config,
		clientset: clientset,
	}, nil
}

// ValidateConnection tests if the credentials can successfully connect to the cluster
func (cv *CredentialValidator) ValidateConnection(ctx context.Context) error {
	// Set a reasonable timeout for validation
	ctx, cancel := context.WithTimeout(ctx, constants.ValidationTimeout)
	defer cancel()
	
	// Try to get server version as a lightweight connectivity test
	version, err := cv.clientset.Discovery().ServerVersion()
	if err != nil {
		return NewAuthError(
			"connection_validation_failed",
			"failed to connect to kubernetes cluster",
			err,
		)
	}
	
	// Basic sanity check on the version response
	if version.GitVersion == "" {
		return NewAuthError(
			"invalid_server_response",
			"received invalid server version response",
			nil,
		)
	}
	
	return nil
}

// ValidatePermissions checks if the credentials have basic permissions
func (cv *CredentialValidator) ValidatePermissions(ctx context.Context) error {
	// Set a reasonable timeout
	ctx, cancel := context.WithTimeout(ctx, constants.ValidationTimeout)
	defer cancel()
	
	// Try to list namespaces (a basic permission check)
	_, err := cv.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return NewAuthError(
			"permission_validation_failed",
			"failed to list namespaces - insufficient permissions or authentication failed",
			err,
		)
	}
	
	return nil
}

// GetServerInfo returns basic information about the connected server
func (cv *CredentialValidator) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.ValidationTimeout)
	defer cancel()
	
	version, err := cv.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, NewAuthError(
			"server_info_failed",
			"failed to get server information",
			err,
		)
	}
	
	// Check if this is an OpenShift cluster by looking for OpenShift-specific APIs
	isOpenShift := cv.isOpenShiftCluster(ctx)
	
	return &ServerInfo{
		GitVersion:   version.GitVersion,
		Major:        version.Major,
		Minor:        version.Minor,
		Platform:     version.Platform,
		IsOpenShift:  isOpenShift,
		ServerHost:   cv.config.Host,
	}, nil
}

// isOpenShiftCluster attempts to detect if this is an OpenShift cluster
func (cv *CredentialValidator) isOpenShiftCluster(ctx context.Context) bool {
	// Try to access OpenShift-specific API groups
	apiGroups, err := cv.clientset.Discovery().ServerGroups()
	if err != nil {
		return false
	}
	
	// Look for OpenShift-specific API groups
	for _, group := range apiGroups.Groups {
		if group.Name == "route.openshift.io" ||
		   group.Name == "build.openshift.io" ||
		   group.Name == "image.openshift.io" {
			return true
		}
	}
	
	return false
}

// ServerInfo contains information about the connected Kubernetes/OpenShift server
type ServerInfo struct {
	GitVersion  string
	Major       string
	Minor       string
	Platform    string
	IsOpenShift bool
	ServerHost  string
}

func (si *ServerInfo) String() string {
	clusterType := "Kubernetes"
	if si.IsOpenShift {
		clusterType = "OpenShift"
	}
	
	return fmt.Sprintf("%s %s.%s (%s) at %s", 
		clusterType, si.Major, si.Minor, si.GitVersion, si.ServerHost)
}