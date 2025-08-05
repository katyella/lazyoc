// Package messages defines the message types used for communication in the LazyOC TUI.
// It provides structured message types for UI events, Kubernetes operations,
// and application state changes using the Bubble Tea architecture.
package messages

import "github.com/katyella/lazyoc/internal/k8s/resources"

// ConnectionError is sent when K8s connection fails
type ConnectionError struct {
	Err error
}

// ConnectionSuccess is sent when K8s connection succeeds
type ConnectionSuccess struct {
	Context   string
	Namespace string
}

// LoadPodsError is sent when pod loading fails
type LoadPodsError struct {
	Err error
}

// PodsLoaded is sent when pods are successfully loaded
type PodsLoaded struct {
	Pods []resources.PodInfo
}

// RefreshPods is sent to trigger pod list refresh
type RefreshPods struct{}

// RefreshPodLogs is sent to trigger pod logs refresh
type RefreshPodLogs struct{}

// PodLogStreamUpdate is sent when new log lines are received in real-time
type PodLogStreamUpdate struct{
	PodName   string
	Container string
	LogLine   string
}

// PodLogStreamError is sent when log streaming encounters an error
type PodLogStreamError struct{
	PodName   string
	Container string
	Err       error
}

// NamespaceChanged is sent when namespace is changed
type NamespaceChanged struct {
	Namespace string
}

// NoKubeconfigMsg is sent when no kubeconfig is found
type NoKubeconfigMsg struct {
	Message string
}

// ConnectingMsg is sent when starting connection
type ConnectingMsg struct {
	KubeconfigPath string
}

// ClusterInfoLoaded is sent when cluster information is successfully loaded
type ClusterInfoLoaded struct {
	Version    string
	ServerInfo map[string]interface{}
}

// ClusterInfoError is sent when cluster information loading fails
type ClusterInfoError struct {
	Err error
}

// Kubernetes resource messages

// ServicesLoaded is sent when Services are successfully loaded
type ServicesLoaded struct {
	Services []resources.ServiceInfo
}

// ServicesLoadError is sent when Service loading fails
type ServicesLoadError struct {
	Err error
}

// DeploymentsLoaded is sent when Deployments are successfully loaded
type DeploymentsLoaded struct {
	Deployments []resources.DeploymentInfo
}

// DeploymentsLoadError is sent when Deployment loading fails
type DeploymentsLoadError struct {
	Err error
}

// ConfigMapsLoaded is sent when ConfigMaps are successfully loaded
type ConfigMapsLoaded struct {
	ConfigMaps []resources.ConfigMapInfo
}

// ConfigMapsLoadError is sent when ConfigMap loading fails
type ConfigMapsLoadError struct {
	Err error
}

// SecretsLoaded is sent when Secrets are successfully loaded
type SecretsLoaded struct {
	Secrets []resources.SecretInfo
}

// SecretsLoadError is sent when Secret loading fails
type SecretsLoadError struct {
	Err error
}

// ServiceLogsLoaded is sent when service logs are successfully loaded
type ServiceLogsLoaded struct {
	ServiceName string
	Pods        []resources.PodInfo
	Logs        []string
}

// ServiceLogsLoadError is sent when service log loading fails
type ServiceLogsLoadError struct {
	Err error
}

// SecretDataLoaded is sent when secret data is successfully loaded
type SecretDataLoaded struct {
	SecretName string
	Data       map[string]string
	Keys       []string
}

// SecretDataLoadError is sent when secret data loading fails
type SecretDataLoadError struct {
	Err error
}

// OpenShift-specific messages

// BuildConfigsLoaded is sent when BuildConfigs are successfully loaded
type BuildConfigsLoaded struct {
	BuildConfigs []resources.BuildConfigInfo
}

// BuildConfigsLoadError is sent when BuildConfig loading fails
type BuildConfigsLoadError struct {
	Err error
}

// ImageStreamsLoaded is sent when ImageStreams are successfully loaded
type ImageStreamsLoaded struct {
	ImageStreams []resources.ImageStreamInfo
}

// ImageStreamsLoadError is sent when ImageStream loading fails
type ImageStreamsLoadError struct {
	Err error
}

// RoutesLoaded is sent when Routes are successfully loaded
type RoutesLoaded struct {
	Routes []resources.RouteInfo
}

// RoutesLoadError is sent when Route loading fails
type RoutesLoadError struct {
	Err error
}
