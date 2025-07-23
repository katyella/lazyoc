package resources

import (
	"time"
)

// ResourceInfo contains basic information about any Kubernetes resource
type ResourceInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Kind        string            `json:"kind"`
	APIVersion  string            `json:"apiVersion"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"createdAt"`
	Status      string            `json:"status"`
}

// PodInfo represents simplified Pod information
type PodInfo struct {
	ResourceInfo
	Phase         string            `json:"phase"`
	Ready         string            `json:"ready"` // "1/1", "0/1", etc.
	Restarts      int32             `json:"restarts"`
	Age           string            `json:"age"`
	Node          string            `json:"node"`
	IP            string            `json:"ip"`
	ContainerInfo []ContainerInfo   `json:"containers"`
}

// ContainerInfo represents container information within a pod
type ContainerInfo struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Ready   bool   `json:"ready"`
	State   string `json:"state"` // Running, Waiting, Terminated
	Reason  string `json:"reason,omitempty"`
}

// ServiceInfo represents simplified Service information
type ServiceInfo struct {
	ResourceInfo
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []string `json:"ports"`
	Selector    string   `json:"selector"`
	Age         string   `json:"age"`
}

// DeploymentInfo represents simplified Deployment information
type DeploymentInfo struct {
	ResourceInfo
	Replicas          int32  `json:"replicas"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	UpdatedReplicas   int32  `json:"updatedReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Age               string `json:"age"`
	Strategy          string `json:"strategy"`
	Condition         string `json:"condition"`
}

// NamespaceInfo represents simplified Namespace information
type NamespaceInfo struct {
	ResourceInfo
	Phase  string `json:"phase"`
	Age    string `json:"age"`
}

// ConfigMapInfo represents simplified ConfigMap information
type ConfigMapInfo struct {
	ResourceInfo
	DataCount int    `json:"dataCount"`
	Age       string `json:"age"`
}

// SecretInfo represents simplified Secret information
type SecretInfo struct {
	ResourceInfo
	Type     string `json:"type"`
	DataCount int    `json:"dataCount"`
	Age       string `json:"age"`
}

// ResourceList contains a list of resources with metadata
type ResourceList[T any] struct {
	Items      []T    `json:"items"`
	Total      int    `json:"total"`
	Namespace  string `json:"namespace"`
	Continue   string `json:"continue,omitempty"` // For pagination
	Remaining  int64  `json:"remaining,omitempty"` // Estimated remaining items
}

// ListOptions contains options for listing resources
type ListOptions struct {
	Namespace     string            `json:"namespace"`
	LabelSelector string            `json:"labelSelector,omitempty"`
	FieldSelector string            `json:"fieldSelector,omitempty"`
	Limit         int64             `json:"limit,omitempty"`
	Continue      string            `json:"continue,omitempty"`
	Watch         bool              `json:"watch,omitempty"`
}

// NamespaceContext represents current namespace context
type NamespaceContext struct {
	Current   string   `json:"current"`
	Available []string `json:"available"`
	Context   string   `json:"context"`
}

// ProjectInfo represents simplified Project/Namespace information
type ProjectInfo struct {
	ResourceInfo
	DisplayName string            `json:"displayName,omitempty"` // OpenShift display name
	Description string            `json:"description,omitempty"` // OpenShift description
	Phase       string            `json:"phase"`
	Age         string            `json:"age"`
	Requester   string            `json:"requester,omitempty"`   // OpenShift requester
	IsOpenShift bool              `json:"isOpenShift"`           // Whether this is an OpenShift project or K8s namespace
	Quota       *ProjectQuota     `json:"quota,omitempty"`       // Resource quotas if any
}

// ProjectQuota represents resource quota information
type ProjectQuota struct {
	Hard map[string]string `json:"hard,omitempty"`
	Used map[string]string `json:"used,omitempty"`
}

// ProjectContext represents current project context
type ProjectContext struct {
	Current      string        `json:"current"`
	Available    []ProjectInfo `json:"available"`
	Context      string        `json:"context"`
	IsOpenShift  bool          `json:"isOpenShift"`
	ClusterInfo  string        `json:"clusterInfo,omitempty"`
}

// LogOptions represents options for pod log retrieval
type LogOptions struct {
	TailLines    *int64 `json:"tailLines,omitempty"`    // Number of lines from the end of logs to show
	Follow       bool   `json:"follow,omitempty"`       // Follow log output (streaming)
	Previous     bool   `json:"previous,omitempty"`     // Return previous terminated container logs
	SinceSeconds *int64 `json:"sinceSeconds,omitempty"` // Show logs since this many seconds ago
	Timestamps   bool   `json:"timestamps,omitempty"`   // Include timestamps in log lines
}