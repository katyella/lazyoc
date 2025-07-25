package constants

// Error message constants
const (
	// ErrClientNotInitialized is returned when the Kubernetes client is not initialized
	ErrClientNotInitialized = "clientset not initialized"

	// ErrKubeconfigNotLoaded is returned when kubeconfig is not loaded
	ErrKubeconfigNotLoaded = "kubeconfig not loaded"

	// ErrNoKubeconfigPath is returned when no kubeconfig path is found
	ErrNoKubeconfigPath = "no kubeconfig path found"

	// ErrClusterUnreachable is returned when the cluster cannot be reached
	ErrClusterUnreachable = "unable to connect to cluster"

	// ErrInvalidKubeconfig is returned when the kubeconfig is invalid
	ErrInvalidKubeconfig = "invalid kubeconfig file"
)

// Error detection keywords for error mapping
const (
	// Timeout related keywords
	ErrKeywordTimeout  = "timeout"
	ErrKeywordDeadline = "deadline exceeded"

	// Authentication related keywords
	ErrKeywordUnauthorized   = "unauthorized"
	ErrKeywordAuthentication = "authentication"
	ErrKeywordToken          = "token"
	ErrKeywordExpired        = "expired"
	ErrKeywordInvalid        = "invalid"

	// Permission related keywords
	ErrKeywordForbidden    = "forbidden"
	ErrKeywordAccessDenied = "access denied"

	// Resource related keywords
	ErrKeywordNamespace = "namespace"
	ErrKeywordProject   = "project"

	// Configuration related keywords
	ErrKeywordKubeconfig = "kubeconfig"
	ErrKeywordConfig     = "config"

	// Certificate related keywords
	ErrKeywordCertificate = "certificate"
	ErrKeywordX509        = "x509"
	ErrKeywordTLS         = "tls"

	// Platform specific keywords
	ErrKeywordOpenShift = "openshift"
)
