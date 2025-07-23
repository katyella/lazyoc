package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/katyella/lazyoc/internal/constants"
)

// KubeconfigProvider implements authentication using kubeconfig files
type KubeconfigProvider struct {
	kubeconfigPath string
	context        string
	namespace      string
	config         *rest.Config
	rawConfig      *api.Config
}

// NewKubeconfigProvider creates a new kubeconfig authentication provider
func NewKubeconfigProvider(kubeconfigPath string) *KubeconfigProvider {
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
	}
	
	return &KubeconfigProvider{
		kubeconfigPath: kubeconfigPath,
	}
}

// NewKubeconfigProviderWithContext creates a provider with a specific context
func NewKubeconfigProviderWithContext(kubeconfigPath, context string) *KubeconfigProvider {
	provider := NewKubeconfigProvider(kubeconfigPath)
	provider.context = context
	return provider
}

// Authenticate loads and validates the kubeconfig
func (kp *KubeconfigProvider) Authenticate(ctx context.Context) (*rest.Config, error) {
	
	// Check if kubeconfig file exists
	if _, err := os.Stat(kp.kubeconfigPath); os.IsNotExist(err) {
		return nil, NewAuthError(
			"kubeconfig_not_found",
			fmt.Sprintf("kubeconfig file not found at %s", kp.kubeconfigPath),
			err,
		)
	}
	
	// Load the raw kubeconfig for context/namespace info
	rawConfig, err := clientcmd.LoadFromFile(kp.kubeconfigPath)
	if err != nil {
		return nil, NewAuthError(
			"kubeconfig_load_failed",
			"failed to load kubeconfig file",
			err,
		)
	}
	
	kp.rawConfig = rawConfig
	
	// Determine which context to use
	contextName := kp.context
	if contextName == "" {
		contextName = rawConfig.CurrentContext
	}
	
	if contextName == "" {
		return nil, NewAuthError(
			"no_current_context",
			"no current context set in kubeconfig and none specified",
			nil,
		)
	}
	
	// Validate the context exists
	if _, exists := rawConfig.Contexts[contextName]; !exists {
		return nil, NewAuthError(
			"context_not_found",
			fmt.Sprintf("context '%s' not found in kubeconfig", contextName),
			nil,
		)
	}
	
	kp.context = contextName
	
	// Extract namespace from context
	if context := rawConfig.Contexts[contextName]; context != nil {
		kp.namespace = context.Namespace
		if kp.namespace == "" {
			kp.namespace = "default"
		}
	} else {
		kp.namespace = "default"
	}
	
	// Build the rest.Config
	config, err := clientcmd.BuildConfigFromFlags("", kp.kubeconfigPath)
	if err != nil {
		return nil, NewAuthError(
			"config_build_failed",
			"failed to build rest config from kubeconfig",
			err,
		)
	}
	
	// Override context if specified
	if kp.context != rawConfig.CurrentContext {
		contextConfig, err := clientcmd.NewNonInteractiveClientConfig(
			*rawConfig,
			kp.context,
			&clientcmd.ConfigOverrides{},
			nil,
		).ClientConfig()
		if err != nil {
			return nil, NewAuthError(
				"context_config_failed",
				fmt.Sprintf("failed to build config for context '%s'", kp.context),
				err,
			)
		}
		config = contextConfig
	} else {
	}
	
	kp.config = config
	return config, nil
}

// IsValid checks if the current configuration is still valid
func (kp *KubeconfigProvider) IsValid(ctx context.Context) error {
	if kp.config == nil {
		return NewAuthError(
			"not_authenticated",
			"authentication has not been performed",
			nil,
		)
	}
	
	// Check if kubeconfig file still exists
	if _, err := os.Stat(kp.kubeconfigPath); os.IsNotExist(err) {
		return NewAuthError(
			"kubeconfig_missing",
			"kubeconfig file no longer exists",
			err,
		)
	}
	
	// For basic validation, we assume kubeconfig is valid if file exists
	// More sophisticated validation could check token expiry, etc.
	return nil
}

// Refresh reloads the kubeconfig file
func (kp *KubeconfigProvider) Refresh(ctx context.Context) error {
	_, err := kp.Authenticate(ctx)
	return err
}

// GetContext returns the current context name
func (kp *KubeconfigProvider) GetContext() string {
	return kp.context
}

// GetNamespace returns the default namespace
func (kp *KubeconfigProvider) GetNamespace() string {
	return kp.namespace
}

// GetKubeconfigPath returns the path to the kubeconfig file
func (kp *KubeconfigProvider) GetKubeconfigPath() string {
	return kp.kubeconfigPath
}

// GetAvailableContexts returns all available contexts in the kubeconfig
func (kp *KubeconfigProvider) GetAvailableContexts() ([]string, error) {
	if kp.rawConfig == nil {
		// Load the config if not already loaded
		rawConfig, err := clientcmd.LoadFromFile(kp.kubeconfigPath)
		if err != nil {
			return nil, NewAuthError(
				"kubeconfig_load_failed",
				"failed to load kubeconfig file",
				err,
			)
		}
		kp.rawConfig = rawConfig
	}
	
	var contexts []string
	for name := range kp.rawConfig.Contexts {
		contexts = append(contexts, name)
	}
	
	return contexts, nil
}

// SwitchContext switches to a different context
func (kp *KubeconfigProvider) SwitchContext(ctx context.Context, contextName string) error {
	if kp.rawConfig == nil {
		return NewAuthError(
			"not_initialized",
			"kubeconfig not loaded",
			nil,
		)
	}
	
	if _, exists := kp.rawConfig.Contexts[contextName]; !exists {
		return NewAuthError(
			"context_not_found",
			fmt.Sprintf("context '%s' not found in kubeconfig", contextName),
			nil,
		)
	}
	
	kp.context = contextName
	_, err := kp.Authenticate(ctx)
	return err
}

// getDefaultKubeconfigPath returns the default kubeconfig path
func getDefaultKubeconfigPath() string {
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