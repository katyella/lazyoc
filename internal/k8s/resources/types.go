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
	Phase         string          `json:"phase"`
	Ready         string          `json:"ready"` // "1/1", "0/1", etc.
	Restarts      int32           `json:"restarts"`
	Age           string          `json:"age"`
	Node          string          `json:"node"`
	IP            string          `json:"ip"`
	ContainerInfo []ContainerInfo `json:"containers"`
}

// ContainerInfo represents container information within a pod
type ContainerInfo struct {
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	Ready        bool            `json:"ready"`
	State        string          `json:"state"` // Running, Waiting, Terminated
	Reason       string          `json:"reason,omitempty"`
	RestartCount int32           `json:"restartCount"`
	Ports        []ContainerPort `json:"ports,omitempty"`
	Env          []EnvVar        `json:"env,omitempty"`
}

// ContainerPort represents a port in a container
type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

// EnvVar represents an environment variable in a container
type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// EnvVarSource represents the source of an environment variable value
type EnvVarSource struct {
	ConfigMapKeyRef  *KeySelector           `json:"configMapKeyRef,omitempty"`
	SecretKeyRef     *KeySelector           `json:"secretKeyRef,omitempty"`
	FieldRef         *FieldSelector         `json:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty"`
}

// KeySelector selects a key from a ConfigMap or Secret
type KeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// FieldSelector selects a field from the pod spec
type FieldSelector struct {
	FieldPath string `json:"fieldPath"`
}

// ResourceFieldSelector selects a resource field
type ResourceFieldSelector struct {
	Resource string `json:"resource"`
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
	Phase string `json:"phase"`
	Age   string `json:"age"`
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
	Type      string `json:"type"`
	DataCount int    `json:"dataCount"`
	Age       string `json:"age"`
}

// ResourceList contains a list of resources with metadata
type ResourceList[T any] struct {
	Items     []T    `json:"items"`
	Total     int    `json:"total"`
	Namespace string `json:"namespace"`
	Continue  string `json:"continue,omitempty"`  // For pagination
	Remaining int64  `json:"remaining,omitempty"` // Estimated remaining items
}

// ListOptions contains options for listing resources
type ListOptions struct {
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"labelSelector,omitempty"`
	FieldSelector string `json:"fieldSelector,omitempty"`
	Limit         int64  `json:"limit,omitempty"`
	Continue      string `json:"continue,omitempty"`
	Watch         bool   `json:"watch,omitempty"`
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
	DisplayName string        `json:"displayName,omitempty"` // OpenShift display name
	Description string        `json:"description,omitempty"` // OpenShift description
	Phase       string        `json:"phase"`
	Age         string        `json:"age"`
	Requester   string        `json:"requester,omitempty"` // OpenShift requester
	IsOpenShift bool          `json:"isOpenShift"`         // Whether this is an OpenShift project or K8s namespace
	Quota       *ProjectQuota `json:"quota,omitempty"`     // Resource quotas if any
}

// ProjectQuota represents resource quota information
type ProjectQuota struct {
	Hard map[string]string `json:"hard,omitempty"`
	Used map[string]string `json:"used,omitempty"`
}

// ProjectContext represents current project context
type ProjectContext struct {
	Current     string        `json:"current"`
	Available   []ProjectInfo `json:"available"`
	Context     string        `json:"context"`
	IsOpenShift bool          `json:"isOpenShift"`
	ClusterInfo string        `json:"clusterInfo,omitempty"`
}

// LogOptions represents options for pod log retrieval
type LogOptions struct {
	TailLines    *int64 `json:"tailLines,omitempty"`    // Number of lines from the end of logs to show
	Follow       bool   `json:"follow,omitempty"`       // Follow log output (streaming)
	Previous     bool   `json:"previous,omitempty"`     // Return previous terminated container logs
	SinceSeconds *int64 `json:"sinceSeconds,omitempty"` // Show logs since this many seconds ago
	Timestamps   bool   `json:"timestamps,omitempty"`   // Include timestamps in log lines
}

// OpenShift-specific resource types

// BuildConfigInfo represents simplified BuildConfig information
type BuildConfigInfo struct {
	ResourceInfo
	Strategy      string               `json:"strategy"`      // Docker, Source, Custom
	Source        BuildSource          `json:"source"`        // Source configuration
	Output        BuildOutput          `json:"output"`        // Output configuration
	Triggers      []BuildTrigger       `json:"triggers"`      // Build triggers
	LastBuild     *BuildInfo           `json:"lastBuild,omitempty"` // Most recent build
	SuccessBuilds int                  `json:"successBuilds"` // Number of successful builds
	FailedBuilds  int                  `json:"failedBuilds"`  // Number of failed builds
	Age           string               `json:"age"`
}

// BuildSource represents build source configuration
type BuildSource struct {
	Type       string     `json:"type"`       // Git, Binary, Dockerfile, etc.
	Git        *GitSource `json:"git,omitempty"`
	Dockerfile string     `json:"dockerfile,omitempty"`
	ContextDir string     `json:"contextDir,omitempty"`
}

// GitSource represents git source configuration
type GitSource struct {
	URI string `json:"uri"`
	Ref string `json:"ref,omitempty"`
}

// BuildOutput represents build output configuration
type BuildOutput struct {
	To       *BuildOutputTo `json:"to,omitempty"`
	PushSecret string       `json:"pushSecret,omitempty"`
}

// BuildOutputTo represents build output destination
type BuildOutputTo struct {
	Kind      string `json:"kind"`      // ImageStreamTag, DockerImage
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// BuildTrigger represents a build trigger
type BuildTrigger struct {
	Type           string              `json:"type"`           // GitHub, Generic, ConfigChange, ImageChange
	GitHub         *GitHubWebHook      `json:"github,omitempty"`
	Generic        *GenericWebHook     `json:"generic,omitempty"`
	ImageChange    *ImageChangeTrigger `json:"imageChange,omitempty"`
}

// GitHubWebHook represents GitHub webhook trigger
type GitHubWebHook struct {
	Secret string `json:"secret,omitempty"`
}

// GenericWebHook represents generic webhook trigger
type GenericWebHook struct {
	Secret string `json:"secret,omitempty"`
}

// ImageChangeTrigger represents image change trigger
type ImageChangeTrigger struct {
	LastTriggeredImageID string `json:"lastTriggeredImageID,omitempty"`
}

// BuildInfo represents simplified Build information
type BuildInfo struct {
	ResourceInfo
	Phase         string    `json:"phase"`         // New, Pending, Running, Complete, Failed, Error, Cancelled
	Message       string    `json:"message,omitempty"`
	Duration      string    `json:"duration"`      // Build duration
	StartTime     time.Time `json:"startTime"`
	CompletionTime *time.Time `json:"completionTime,omitempty"`
	BuildConfig   string    `json:"buildConfig"`   // Parent BuildConfig name
	Strategy      string    `json:"strategy"`      // Build strategy used
	OutputImage   string    `json:"outputImage,omitempty"` // Resulting image
	Age           string    `json:"age"`
}

// ImageStreamInfo represents simplified ImageStream information
type ImageStreamInfo struct {
	ResourceInfo
	DockerImageRepository string                    `json:"dockerImageRepository"`
	Tags                  []ImageStreamTag          `json:"tags"`
	PublicDockerImageRepository string            `json:"publicDockerImageRepository,omitempty"`
	Age                   string                    `json:"age"`
}

// ImageStreamTag represents a tag within an ImageStream
type ImageStreamTag struct {
	Name         string              `json:"name"`
	Items        []ImageStreamImage  `json:"items"`
	Conditions   []TagEventCondition `json:"conditions,omitempty"`
}

// ImageStreamImage represents an image within a tag
type ImageStreamImage struct {
	Created        time.Time `json:"created"`
	DockerImageRef string    `json:"dockerImageRef"`
	Image          string    `json:"image"`      // SHA256 digest
	Generation     int64     `json:"generation"`
}

// TagEventCondition represents the condition of a tag event
type TagEventCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// DeploymentConfigInfo represents simplified DeploymentConfig information
type DeploymentConfigInfo struct {
	ResourceInfo
	Replicas          int32                      `json:"replicas"`
	ReadyReplicas     int32                      `json:"readyReplicas"`
	UpdatedReplicas   int32                      `json:"updatedReplicas"`
	AvailableReplicas int32                      `json:"availableReplicas"`
	LatestVersion     int64                      `json:"latestVersion"`
	Strategy          DeploymentStrategy         `json:"strategy"`
	Triggers          []DeploymentTrigger        `json:"triggers"`
	Conditions        []DeploymentCondition      `json:"conditions"`
	Age               string                     `json:"age"`
}

// DeploymentStrategy represents deployment strategy
type DeploymentStrategy struct {
	Type           string                    `json:"type"`             // Recreate, Rolling
	RollingParams  *RollingDeploymentParams  `json:"rollingParams,omitempty"`
	RecreateParams *RecreateDeploymentParams `json:"recreateParams,omitempty"`
}

// RollingDeploymentParams represents rolling deployment parameters
type RollingDeploymentParams struct {
	UpdatePeriodSeconds     *int64 `json:"updatePeriodSeconds,omitempty"`
	IntervalSeconds         *int64 `json:"intervalSeconds,omitempty"`
	TimeoutSeconds          *int64 `json:"timeoutSeconds,omitempty"`
	MaxUnavailable          string `json:"maxUnavailable,omitempty"`
	MaxSurge                string `json:"maxSurge,omitempty"`
}

// RecreateDeploymentParams represents recreate deployment parameters
type RecreateDeploymentParams struct {
	TimeoutSeconds *int64 `json:"timeoutSeconds,omitempty"`
}

// DeploymentTrigger represents a deployment trigger
type DeploymentTrigger struct {
	Type           string                      `json:"type"`           // ConfigChange, ImageChange
	ImageChange    *DeploymentTriggerImageChange `json:"imageChange,omitempty"`
}

// DeploymentTriggerImageChange represents image change trigger
type DeploymentTriggerImageChange struct {
	From                *ImageStreamReference `json:"from,omitempty"`
	LastTriggeredImage  string               `json:"lastTriggeredImage,omitempty"`
	ContainerNames      []string             `json:"containerNames,omitempty"`
}

// ImageStreamReference represents a reference to an ImageStream
type ImageStreamReference struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

// DeploymentCondition represents a deployment condition
type DeploymentCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastUpdateTime     time.Time `json:"lastUpdateTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// RouteInfo represents simplified Route information
type RouteInfo struct {
	ResourceInfo
	Host               string            `json:"host"`
	Path               string            `json:"path,omitempty"`
	Service            RouteTargetRef    `json:"service"`
	Port               *RoutePort        `json:"port,omitempty"`
	TLS                *TLSConfig        `json:"tls,omitempty"`
	WildcardPolicy     string            `json:"wildcardPolicy,omitempty"`
	AdmittedConditions []RouteCondition  `json:"admittedConditions"`
	Age                string            `json:"age"`
}

// RouteTargetRef represents a route target reference
type RouteTargetRef struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Weight *int32 `json:"weight,omitempty"`
}

// RoutePort represents a route port
type RoutePort struct {
	TargetPort string `json:"targetPort"`
}

// TLSConfig represents TLS configuration for a route
type TLSConfig struct {
	Termination                   string `json:"termination"`                   // edge, passthrough, reencrypt
	Certificate                   string `json:"certificate,omitempty"`
	Key                          string `json:"key,omitempty"`
	CACertificate                string `json:"caCertificate,omitempty"`
	DestinationCACertificate     string `json:"destinationCACertificate,omitempty"`
	InsecureEdgeTerminationPolicy string `json:"insecureEdgeTerminationPolicy,omitempty"`
}

// RouteCondition represents a route condition
type RouteCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// OperatorInfo represents simplified Operator information (ClusterServiceVersion)
type OperatorInfo struct {
	ResourceInfo
	Phase              string                     `json:"phase"`              // Pending, Installing, Succeeded, Failed, Unknown
	Version            string                     `json:"version"`
	DisplayName        string                     `json:"displayName"`
	Description        string                     `json:"description"`
	Provider           OperatorProvider           `json:"provider"`
	InstallModes       []OperatorInstallMode      `json:"installModes"`
	Requirements       []OperatorRequirement      `json:"requirements"`
	Conditions         []OperatorCondition        `json:"conditions"`
	OwnedResources     []OwnedResource           `json:"ownedResources"`     // CRDs owned by this operator
	RequiredResources  []RequiredResource        `json:"requiredResources"`  // Resources required by this operator
	Age                string                     `json:"age"`
}

// OperatorProvider represents operator provider information
type OperatorProvider struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// OperatorInstallMode represents an install mode
type OperatorInstallMode struct {
	Type      string `json:"type"`      // OwnNamespace, SingleNamespace, MultiNamespace, AllNamespaces
	Supported bool   `json:"supported"`
}

// OperatorRequirement represents an operator requirement
type OperatorRequirement struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// OperatorCondition represents an operator condition
type OperatorCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastUpdateTime     time.Time `json:"lastUpdateTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// OwnedResource represents a resource owned by an operator
type OwnedResource struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Kind        string `json:"kind"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// RequiredResource represents a resource required by an operator
type RequiredResource struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Kind        string `json:"kind"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// SubscriptionInfo represents simplified Subscription information
type SubscriptionInfo struct {
	ResourceInfo
	Channel                string               `json:"channel"`
	StartingCSV            string               `json:"startingCSV"`
	CurrentCSV             string               `json:"currentCSV"`
	InstalledCSV           string               `json:"installedCSV"`
	InstallPlanGeneration  int64                `json:"installPlanGeneration"`
	InstallPlanRef         *InstallPlanRef      `json:"installPlanRef,omitempty"`
	State                  string               `json:"state"`                  // UpgradePending, AtLatestKnown, etc.
	Conditions             []SubscriptionCondition `json:"conditions"`
	Age                    string               `json:"age"`
}

// InstallPlanRef represents a reference to an InstallPlan
type InstallPlanRef struct {
	APIVersion      string `json:"apiVersion"`
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	ResourceVersion string `json:"resourceVersion"`
	UID             string `json:"uid"`
}

// SubscriptionCondition represents a subscription condition
type SubscriptionCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}
