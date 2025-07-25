package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKubeconfigProvider_GetDefaultPath(t *testing.T) {
	// Test default path resolution
	provider := NewKubeconfigProvider("")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedPath := filepath.Join(home, ".kube", "config")
	if provider.GetKubeconfigPath() != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, provider.GetKubeconfigPath())
	}
}

func TestKubeconfigProvider_EnvironmentVariable(t *testing.T) {
	// Test KUBECONFIG environment variable
	testPath := filepath.Join(os.TempDir(), "test-kubeconfig")
	oldKubeconfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", oldKubeconfig)

	os.Setenv("KUBECONFIG", testPath)
	provider := NewKubeconfigProvider("")

	if provider.GetKubeconfigPath() != testPath {
		t.Errorf("Expected path %s, got %s", testPath, provider.GetKubeconfigPath())
	}
}

func TestKubeconfigProvider_FileNotFound(t *testing.T) {
	provider := NewKubeconfigProvider("/nonexistent/kubeconfig")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := provider.Authenticate(ctx)
	if err == nil {
		t.Error("Expected authentication to fail for nonexistent file")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Errorf("Expected AuthError, got %T", err)
	}

	if authErr.Type != "kubeconfig_not_found" {
		t.Errorf("Expected error type 'kubeconfig_not_found', got '%s'", authErr.Type)
	}
}

func TestKubeconfigProvider_MockConfig(t *testing.T) {
	// Create a temporary kubeconfig file for testing
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	// Create a mock kubeconfig file
	mockKubeconfig := `
apiVersion: v1
kind: Config
current-context: test-context
contexts:
- context:
    cluster: test-cluster
    namespace: test-namespace
    user: test-user
  name: test-context
clusters:
- cluster:
    server: https://test-server:6443
    insecure-skip-tls-verify: true
  name: test-cluster
users:
- name: test-user
  user:
    token: test-token
`

	err := os.WriteFile(kubeconfigPath, []byte(mockKubeconfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create mock kubeconfig: %v", err)
	}

	provider := NewKubeconfigProvider(kubeconfigPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := provider.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if config == nil {
		t.Error("Expected non-nil config")
		return
	}

	if config.Host != "https://test-server:6443" {
		t.Errorf("Expected host 'https://test-server:6443', got '%s'", config.Host)
	}

	// Test context and namespace retrieval
	context := provider.GetContext()
	if context != "test-context" {
		t.Errorf("Expected context 'test-context', got '%s'", context)
	}

	namespace := provider.GetNamespace()
	if namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", namespace)
	}
}

func TestKubeconfigProvider_IsValid(t *testing.T) {
	// Test with nonexistent config
	provider := NewKubeconfigProvider("/nonexistent/config")

	ctx := context.Background()

	// Should fail before authentication
	err := provider.IsValid(ctx)
	if err == nil {
		t.Error("Expected IsValid to fail before authentication")
	}

	// Test with mock config
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	mockKubeconfig := `
apiVersion: v1
kind: Config
current-context: test-context
contexts:
- context:
    cluster: test-cluster
    namespace: test-namespace
    user: test-user
  name: test-context
clusters:
- cluster:
    server: https://test-server:6443
    insecure-skip-tls-verify: true
  name: test-cluster
users:
- name: test-user
  user:
    token: test-token
`

	err = os.WriteFile(kubeconfigPath, []byte(mockKubeconfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create mock kubeconfig: %v", err)
	}

	provider = NewKubeconfigProvider(kubeconfigPath)

	// Authenticate first
	_, err = provider.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Should pass after authentication (even though server is fake)
	// IsValid only checks if config was loaded successfully
	err = provider.IsValid(ctx)
	if err != nil {
		t.Errorf("Expected IsValid to pass after authentication: %v", err)
	}
}

func TestKubeconfigProvider_GetAvailableContexts(t *testing.T) {
	// Create a temporary kubeconfig file with multiple contexts
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	mockKubeconfig := `
apiVersion: v1
kind: Config
current-context: test-context
contexts:
- context:
    cluster: test-cluster
    namespace: test-namespace
    user: test-user
  name: test-context
- context:
    cluster: prod-cluster  
    namespace: prod-namespace
    user: prod-user
  name: prod-context
clusters:
- cluster:
    server: https://test-server:6443
    insecure-skip-tls-verify: true
  name: test-cluster
- cluster:
    server: https://prod-server:6443
    insecure-skip-tls-verify: true
  name: prod-cluster
users:
- name: test-user
  user:
    token: test-token
- name: prod-user
  user:
    token: prod-token
`

	err := os.WriteFile(kubeconfigPath, []byte(mockKubeconfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create mock kubeconfig: %v", err)
	}

	provider := NewKubeconfigProvider(kubeconfigPath)

	contexts, err := provider.GetAvailableContexts()
	if err != nil {
		t.Fatalf("Failed to get available contexts: %v", err)
	}

	expectedContexts := []string{"test-context", "prod-context"}
	if len(contexts) != len(expectedContexts) {
		t.Errorf("Expected %d contexts, got %d", len(expectedContexts), len(contexts))
	}

	for _, expected := range expectedContexts {
		found := false
		for _, actual := range contexts {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected context %s not found in %v", expected, contexts)
		}
	}
}

func TestAuthManager_NoProviders(t *testing.T) {
	manager := NewAuthManager()

	ctx := context.Background()
	_, err := manager.Authenticate(ctx)
	if err == nil {
		t.Error("Expected authentication to fail with no providers")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Errorf("Expected AuthError, got %T", err)
	}

	if authErr.Type != "authentication_failed" {
		t.Errorf("Expected error type 'authentication_failed', got '%s'", authErr.Type)
	}
}

func TestAuthManager_WithProvider(t *testing.T) {
	// Create a temporary kubeconfig file for testing
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	mockKubeconfig := `
apiVersion: v1
kind: Config
current-context: test-context
contexts:
- context:
    cluster: test-cluster
    namespace: test-namespace
    user: test-user
  name: test-context
clusters:
- cluster:
    server: https://test-server:6443
    insecure-skip-tls-verify: true
  name: test-cluster
users:
- name: test-user
  user:
    token: test-token
`

	err := os.WriteFile(kubeconfigPath, []byte(mockKubeconfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create mock kubeconfig: %v", err)
	}

	manager := NewAuthManager()
	provider := NewKubeconfigProvider(kubeconfigPath)
	manager.AddProvider(provider)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := manager.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if config == nil {
		t.Error("Expected non-nil config")
	}

	// Check that the provider is set as active
	if manager.GetActiveProvider() != provider {
		t.Error("Expected provider to be set as active")
	}

	// Test IsValid
	err = manager.IsValid(ctx)
	if err != nil {
		t.Errorf("Expected IsValid to pass: %v", err)
	}
}
