package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

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

// All other tests are disabled as they require real cluster connections:
// - TestClientFactory_Initialize
// - TestClientFactory_TestConnection
// - TestClientFactory_GetCurrentContext
// - TestClientFactory_GetCurrentNamespace
//
// These tests should be rewritten as proper unit tests with mocked
// Kubernetes clients if needed, but are currently disabled to prevent
// network dependencies.