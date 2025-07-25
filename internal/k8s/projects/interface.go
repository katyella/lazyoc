package projects

import (
	"context"
	"time"

	"github.com/katyella/lazyoc/internal/k8s"
)

// ProjectInfo represents unified project/namespace information
type ProjectInfo struct {
	// Common fields
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	Status      string            `json:"status"`

	// Type information
	Type        ProjectType     `json:"type"`        // OpenShift project or Kubernetes namespace
	ClusterType k8s.ClusterType `json:"clusterType"` // OpenShift or Kubernetes cluster

	// OpenShift-specific fields (empty for Kubernetes namespaces)
	Requester string `json:"requester,omitempty"`

	// Resource limits and quotas
	ResourceQuotas []ResourceQuota `json:"resourceQuotas,omitempty"`
	LimitRanges    []LimitRange    `json:"limitRanges,omitempty"`
}

// ProjectType represents whether this is an OpenShift project or Kubernetes namespace
type ProjectType int

const (
	ProjectTypeUnknown ProjectType = iota
	ProjectTypeKubernetesNamespace
	ProjectTypeOpenShiftProject
)

func (pt ProjectType) String() string {
	switch pt {
	case ProjectTypeKubernetesNamespace:
		return "Namespace"
	case ProjectTypeOpenShiftProject:
		return "Project"
	default:
		return "Unknown"
	}
}

// ResourceQuota represents resource quota information
type ResourceQuota struct {
	Name   string            `json:"name"`
	Hard   map[string]string `json:"hard,omitempty"`
	Used   map[string]string `json:"used,omitempty"`
	Scopes []string          `json:"scopes,omitempty"`
}

// LimitRange represents limit range information
type LimitRange struct {
	Name   string      `json:"name"`
	Limits []LimitItem `json:"limits,omitempty"`
}

// LimitItem represents a single limit item
type LimitItem struct {
	Type           string            `json:"type"`
	Max            map[string]string `json:"max,omitempty"`
	Min            map[string]string `json:"min,omitempty"`
	Default        map[string]string `json:"default,omitempty"`
	DefaultRequest map[string]string `json:"defaultRequest,omitempty"`
}

// ListOptions provides filtering options for listing projects/namespaces
type ListOptions struct {
	// Label selector
	LabelSelector string

	// Field selector
	FieldSelector string

	// Include projects/namespaces user doesn't have access to (returns limited info)
	IncludeInaccessible bool

	// Include resource quotas and limits
	IncludeQuotas bool
	IncludeLimits bool
}

// CreateOptions provides options for creating projects/namespaces
type CreateOptions struct {
	// Display name (OpenShift projects only)
	DisplayName string

	// Description
	Description string

	// Labels to apply
	Labels map[string]string

	// Annotations to apply
	Annotations map[string]string

	// Resource quotas to create
	ResourceQuotas []ResourceQuota

	// Limit ranges to create
	LimitRanges []LimitRange
}

// SwitchResult contains information about a project/namespace switch
type SwitchResult struct {
	From        string       `json:"from"`
	To          string       `json:"to"`
	Success     bool         `json:"success"`
	Message     string       `json:"message,omitempty"`
	ProjectInfo *ProjectInfo `json:"projectInfo,omitempty"`
}

// ProjectManager provides a unified interface for managing OpenShift projects and Kubernetes namespaces
type ProjectManager interface {
	// List all accessible projects/namespaces
	List(ctx context.Context, opts ListOptions) ([]ProjectInfo, error)

	// Get detailed information about a specific project/namespace
	Get(ctx context.Context, name string) (*ProjectInfo, error)

	// Create a new project/namespace
	Create(ctx context.Context, name string, opts CreateOptions) (*ProjectInfo, error)

	// Delete a project/namespace
	Delete(ctx context.Context, name string) error

	// Switch to a different project/namespace (updates kubeconfig current context)
	SwitchTo(ctx context.Context, name string) (*SwitchResult, error)

	// Get current project/namespace
	GetCurrent(ctx context.Context) (*ProjectInfo, error)

	// Check if a project/namespace exists and is accessible
	Exists(ctx context.Context, name string) (bool, error)

	// Get the type of projects this manager handles
	GetProjectType() ProjectType

	// Get the cluster type this manager is connected to
	GetClusterType() k8s.ClusterType

	// Refresh cached information
	RefreshCache(ctx context.Context) error
}

// ProjectManagerFactory creates the appropriate ProjectManager based on cluster type
type ProjectManagerFactory interface {
	// Create a project manager for the given cluster type
	CreateManager(ctx context.Context, clusterType k8s.ClusterType) (ProjectManager, error)

	// Auto-detect cluster type and create appropriate manager
	CreateAutoDetectManager(ctx context.Context) (ProjectManager, error)
}
