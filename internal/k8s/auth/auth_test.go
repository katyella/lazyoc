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
	testPath := "/tmp/test-kubeconfig"
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

func TestKubeconfigProvider_RealConfig(t *testing.T) {
	provider := NewKubeconfigProvider("")
	kubeconfigPath := provider.GetKubeconfigPath()
	
	// Skip test if no kubeconfig available
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	config, err := provider.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	
	if config == nil {
		t.Error("Expected non-nil config")
	}
	
	if config.Host == "" {
		t.Error("Expected non-empty host in config")
	}
	
	// Test context and namespace retrieval
	context := provider.GetContext()
	if context == "" {
		t.Error("Expected non-empty context")
	}
	
	namespace := provider.GetNamespace()
	if namespace == "" {
		t.Error("Expected non-empty namespace")
	}
	
	t.Logf("Context: %s, Namespace: %s", context, namespace)
}

func TestKubeconfigProvider_IsValid(t *testing.T) {
	provider := NewKubeconfigProvider("")
	
	ctx := context.Background()
	
	// Should fail before authentication
	err := provider.IsValid(ctx)
	if err == nil {
		t.Error("Expected IsValid to fail before authentication")
	}
	
	// Skip further test if no kubeconfig available
	if _, err := os.Stat(provider.GetKubeconfigPath()); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping validation test")
	}
	
	// Authenticate first
	_, err = provider.Authenticate(ctx)
	if err != nil {
		t.Skipf("Authentication failed, skipping validation test: %v", err)
	}
	
	// Should pass after authentication
	err = provider.IsValid(ctx)
	if err != nil {
		t.Errorf("Expected IsValid to pass after authentication: %v", err)
	}
}

func TestKubeconfigProvider_GetAvailableContexts(t *testing.T) {
	provider := NewKubeconfigProvider("")
	
	// Skip test if no kubeconfig available
	if _, err := os.Stat(provider.GetKubeconfigPath()); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	contexts, err := provider.GetAvailableContexts()
	if err != nil {
		t.Fatalf("Failed to get available contexts: %v", err)
	}
	
	if len(contexts) == 0 {
		t.Error("Expected at least one context")
	}
	
	t.Logf("Available contexts: %v", contexts)
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
	manager := NewAuthManager()
	provider := NewKubeconfigProvider("")
	manager.AddProvider(provider)
	
	// Skip test if no kubeconfig available
	if _, err := os.Stat(provider.GetKubeconfigPath()); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func TestCredentialValidator_RealCluster(t *testing.T) {
	provider := NewKubeconfigProvider("")
	
	// Skip test if no kubeconfig available
	if _, err := os.Stat(provider.GetKubeconfigPath()); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	config, err := provider.Authenticate(ctx)
	if err != nil {
		t.Skipf("Authentication failed, skipping validator test: %v", err)
	}
	
	validator, err := NewCredentialValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	
	// Test connection validation
	err = validator.ValidateConnection(ctx)
	if err != nil {
		t.Logf("Connection validation failed (may be expected if cluster is unreachable): %v", err)
		// Don't fail the test - cluster might not be running
		return
	}
	
	t.Log("Connection validation passed")
	
	// Test server info retrieval
	info, err := validator.GetServerInfo(ctx)
	if err != nil {
		t.Errorf("Failed to get server info: %v", err)
		return
	}
	
	t.Logf("Server info: %s", info.String())
	
	if info.GitVersion == "" {
		t.Error("Expected non-empty git version")
	}
	
	// Test permission validation (might fail depending on cluster permissions)
	err = validator.ValidatePermissions(ctx)
	if err != nil {
		t.Logf("Permission validation failed (may be expected with limited permissions): %v", err)
	} else {
		t.Log("Permission validation passed")
	}
}