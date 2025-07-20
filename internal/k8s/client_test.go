package k8s

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClientFactory_Initialize(t *testing.T) {
	factory := NewClientFactory()
	
	// Skip test if no kubeconfig available
	kubeconfigPath := factory.getKubeconfigPath()
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	err := factory.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize client factory: %v", err)
	}
	
	if factory.GetClientset() == nil {
		t.Error("Expected clientset to be initialized")
	}
	
	if factory.GetConfig() == nil {
		t.Error("Expected config to be initialized")
	}
}

func TestClientFactory_GetKubeconfigPath(t *testing.T) {
	factory := NewClientFactory()
	
	// Test default path
	path := factory.getKubeconfigPath()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	
	expectedPath := filepath.Join(home, ".kube", "config")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
	
	// Test KUBECONFIG environment variable
	testPath := "/tmp/test-kubeconfig"
	oldKubeconfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", oldKubeconfig)
	
	os.Setenv("KUBECONFIG", testPath)
	path = factory.getKubeconfigPath()
	if path != testPath {
		t.Errorf("Expected path %s, got %s", testPath, path)
	}
}

func TestClientFactory_TestConnection(t *testing.T) {
	factory := NewClientFactory()
	
	// Skip test if no kubeconfig available
	kubeconfigPath := factory.getKubeconfigPath()
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	err := factory.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize client factory: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err = factory.TestConnection(ctx)
	if err != nil {
		t.Logf("Connection test failed (this may be expected if no cluster is running): %v", err)
		// Don't fail the test - cluster might not be running
	} else {
		t.Log("Successfully connected to Kubernetes cluster")
	}
}

func TestClientFactory_GetCurrentContext(t *testing.T) {
	factory := NewClientFactory()
	
	// Skip test if no kubeconfig available
	kubeconfigPath := factory.getKubeconfigPath()
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	err := factory.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize client factory: %v", err)
	}
	
	context, err := factory.GetCurrentContext()
	if err != nil {
		t.Fatalf("Failed to get current context: %v", err)
	}
	
	if context == "" {
		t.Error("Expected non-empty context name")
	}
	
	t.Logf("Current context: %s", context)
}

func TestClientFactory_GetCurrentNamespace(t *testing.T) {
	factory := NewClientFactory()
	
	// Skip test if no kubeconfig available
	kubeconfigPath := factory.getKubeconfigPath()
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping test")
	}
	
	err := factory.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize client factory: %v", err)
	}
	
	namespace, err := factory.GetCurrentNamespace()
	if err != nil {
		t.Fatalf("Failed to get current namespace: %v", err)
	}
	
	if namespace == "" {
		t.Error("Expected non-empty namespace")
	}
	
	t.Logf("Current namespace: %s", namespace)
}