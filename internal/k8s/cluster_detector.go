package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClusterType represents the type of Kubernetes cluster
type ClusterType int

const (
	ClusterTypeUnknown ClusterType = iota
	ClusterTypeKubernetes
	ClusterTypeOpenShift
)

func (ct ClusterType) String() string {
	switch ct {
	case ClusterTypeKubernetes:
		return "Kubernetes"
	case ClusterTypeOpenShift:
		return "OpenShift"
	default:
		return "Unknown"
	}
}

// ClusterInfo contains information about the detected cluster
type ClusterInfo struct {
	Type           ClusterType
	Version        string
	ServerVersion  string
	DetectionTime  time.Time
	APIGroups      []string
	OpenShiftAPIs  []string
}

// ClusterTypeDetector provides methods to detect cluster type
type ClusterTypeDetector struct {
	config     *rest.Config
	clientset  kubernetes.Interface
	discovery  discovery.DiscoveryInterface
	
	// Caching
	mu          sync.RWMutex
	cached      bool
	cachedInfo  *ClusterInfo
	cacheTime   time.Duration
}

// NewClusterTypeDetector creates a new cluster type detector
func NewClusterTypeDetector(config *rest.Config) (*ClusterTypeDetector, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	
	return &ClusterTypeDetector{
		config:    config,
		clientset: clientset,
		discovery: clientset.Discovery(),
		cacheTime: 10 * time.Minute, // Cache results for 10 minutes
	}, nil
}

// DetectClusterType detects whether this is an OpenShift or vanilla Kubernetes cluster
func (d *ClusterTypeDetector) DetectClusterType(ctx context.Context) (*ClusterInfo, error) {
	// Check cache first
	d.mu.RLock()
	if d.cached && d.cachedInfo != nil && time.Since(d.cachedInfo.DetectionTime) < d.cacheTime {
		info := *d.cachedInfo // Copy to avoid race conditions
		d.mu.RUnlock()
		return &info, nil
	}
	d.mu.RUnlock()

	// Set timeout for detection
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Start with unknown cluster info
	info := &ClusterInfo{
		Type:          ClusterTypeUnknown,
		DetectionTime: time.Now(),
		APIGroups:     []string{},
		OpenShiftAPIs: []string{},
	}

	// Get server version
	version, err := d.discovery.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}
	info.ServerVersion = version.GitVersion
	info.Version = fmt.Sprintf("%s.%s", version.Major, version.Minor)

	// Get API groups to check for OpenShift-specific APIs
	apiGroups, err := d.discovery.ServerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to get server groups: %w", err)
	}

	// Extract API group names
	for _, group := range apiGroups.Groups {
		info.APIGroups = append(info.APIGroups, group.Name)
	}

	// Check for OpenShift-specific API groups
	openShiftAPIs := []string{
		"route.openshift.io",
		"build.openshift.io", 
		"image.openshift.io",
		"project.openshift.io",
		"apps.openshift.io",
		"template.openshift.io",
		"security.openshift.io",
		"user.openshift.io",
		"quota.openshift.io",
		"network.openshift.io",
		"authorization.openshift.io",
	}

	foundOpenShiftAPIs := 0
	for _, osAPI := range openShiftAPIs {
		for _, groupAPI := range info.APIGroups {
			if groupAPI == osAPI {
				info.OpenShiftAPIs = append(info.OpenShiftAPIs, osAPI)
				foundOpenShiftAPIs++
				break
			}
		}
	}

	// Determine cluster type based on found APIs
	if foundOpenShiftAPIs >= 3 { // Need at least 3 OpenShift APIs to be confident
		info.Type = ClusterTypeOpenShift
	} else {
		info.Type = ClusterTypeKubernetes
	}

	// Additional verification for OpenShift - try to access /oapi endpoint
	if info.Type == ClusterTypeOpenShift {
		if err := d.verifyOpenShiftOAPI(ctx); err != nil {
			// If /oapi fails but we found OpenShift APIs, it's still likely OpenShift
			// but maybe a newer version that doesn't use /oapi
			info.Type = ClusterTypeOpenShift
		}
	}

	// Cache the result
	d.mu.Lock()
	d.cachedInfo = info
	d.cached = true
	d.mu.Unlock()

	return info, nil
}

// verifyOpenShiftOAPI attempts to verify OpenShift by checking the /oapi endpoint
func (d *ClusterTypeDetector) verifyOpenShiftOAPI(ctx context.Context) error {
	// This is a legacy OpenShift endpoint check
	// Newer OpenShift versions might not have this, so failure is not definitive
	// For now, we'll skip this check as the API group detection is more reliable
	return nil
}

// IsOpenShift returns true if the cluster is detected as OpenShift
func (d *ClusterTypeDetector) IsOpenShift(ctx context.Context) (bool, error) {
	info, err := d.DetectClusterType(ctx)
	if err != nil {
		return false, err
	}
	return info.Type == ClusterTypeOpenShift, nil
}

// IsKubernetes returns true if the cluster is detected as vanilla Kubernetes
func (d *ClusterTypeDetector) IsKubernetes(ctx context.Context) (bool, error) {
	info, err := d.DetectClusterType(ctx)
	if err != nil {
		return false, err
	}
	return info.Type == ClusterTypeKubernetes, nil
}

// GetClusterType returns the cached cluster type or detects it
func (d *ClusterTypeDetector) GetClusterType(ctx context.Context) (ClusterType, error) {
	info, err := d.DetectClusterType(ctx)
	if err != nil {
		return ClusterTypeUnknown, err
	}
	return info.Type, nil
}

// GetClusterInfo returns full cluster information
func (d *ClusterTypeDetector) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	return d.DetectClusterType(ctx)
}

// ClearCache clears the cached cluster detection result
func (d *ClusterTypeDetector) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cached = false
	d.cachedInfo = nil
}

// SetCacheTime sets the cache duration for cluster detection results
func (d *ClusterTypeDetector) SetCacheTime(duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cacheTime = duration
}

// HasAPIGroup checks if a specific API group is available in the cluster
func (d *ClusterTypeDetector) HasAPIGroup(ctx context.Context, apiGroup string) (bool, error) {
	info, err := d.DetectClusterType(ctx)
	if err != nil {
		return false, err
	}
	
	for _, group := range info.APIGroups {
		if group == apiGroup {
			return true, nil
		}
	}
	return false, nil
}

// SupportsProjects returns true if the cluster supports OpenShift projects
func (d *ClusterTypeDetector) SupportsProjects(ctx context.Context) (bool, error) {
	return d.HasAPIGroup(ctx, "project.openshift.io")
}