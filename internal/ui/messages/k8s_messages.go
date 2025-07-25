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
