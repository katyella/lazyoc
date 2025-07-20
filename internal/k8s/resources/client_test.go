package resources

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getTestClient(t *testing.T) *kubernetes.Clientset {
	// Get kubeconfig path
	kubeconfigPath := ""
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		kubeconfigPath = kubeconfig
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot determine home directory")
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Check if kubeconfig exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Skip("No kubeconfig file found, skipping integration tests")
	}

	// Build config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		t.Skipf("Failed to build config: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Skipf("Failed to create clientset: %v", err)
	}

	return clientset
}

func TestK8sResourceClient_ListNamespaces(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		t.Skipf("Failed to list namespaces (cluster may be unreachable): %v", err)
	}

	if len(namespaces.Items) == 0 {
		t.Error("Expected at least one namespace")
	}

	// Should find default namespace
	found := false
	for _, ns := range namespaces.Items {
		if ns.Name == "default" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'default' namespace")
	}

	t.Logf("Found %d namespaces", len(namespaces.Items))
}

func TestK8sResourceClient_ListPods(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := ListOptions{
		Namespace: "kube-system", // Usually has system pods
		Limit:     10,
	}

	pods, err := client.ListPods(ctx, opts)
	if err != nil {
		t.Skipf("Failed to list pods (cluster may be unreachable or no access): %v", err)
	}

	t.Logf("Found %d pods in kube-system namespace", len(pods.Items))

	// Validate pod structure
	for _, pod := range pods.Items {
		if pod.Name == "" {
			t.Error("Pod name should not be empty")
		}
		if pod.Namespace != "kube-system" {
			t.Errorf("Expected namespace 'kube-system', got '%s'", pod.Namespace)
		}
		if pod.Kind != "Pod" {
			t.Errorf("Expected kind 'Pod', got '%s'", pod.Kind)
		}
		if pod.Age == "" {
			t.Error("Pod age should not be empty")
		}
	}
}

func TestK8sResourceClient_ListServices(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := ListOptions{
		Namespace: "default",
		Limit:     10,
	}

	services, err := client.ListServices(ctx, opts)
	if err != nil {
		t.Skipf("Failed to list services (cluster may be unreachable): %v", err)
	}

	t.Logf("Found %d services in default namespace", len(services.Items))

	// Should at least find kubernetes service
	found := false
	for _, svc := range services.Items {
		if svc.Name == "kubernetes" {
			found = true
			if svc.Type != "ClusterIP" {
				t.Errorf("Expected kubernetes service to be ClusterIP, got %s", svc.Type)
			}
			break
		}
	}

	if !found {
		t.Log("Warning: kubernetes service not found in default namespace")
	}
}

func TestK8sResourceClient_NamespaceOperations(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	// Test current namespace
	current := client.GetCurrentNamespace()
	if current != "default" {
		t.Errorf("Expected current namespace to be 'default', got '%s'", current)
	}

	// Test setting namespace
	err := client.SetCurrentNamespace("kube-system")
	if err != nil {
		t.Errorf("Failed to set namespace: %v", err)
	}

	if client.GetCurrentNamespace() != "kube-system" {
		t.Error("Namespace was not updated")
	}

	// Test namespace context
	nsCtx, err := client.GetNamespaceContext()
	if err != nil {
		t.Skipf("Failed to get namespace context: %v", err)
	}

	if nsCtx.Current != "kube-system" {
		t.Errorf("Expected current namespace 'kube-system', got '%s'", nsCtx.Current)
	}

	if len(nsCtx.Available) == 0 {
		t.Error("Expected at least one available namespace")
	}

	t.Logf("Current: %s, Available: %v", nsCtx.Current, nsCtx.Available)
}

func TestK8sResourceClient_TestConnection(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.TestConnection(ctx)
	if err != nil {
		t.Skipf("Connection test failed (cluster may be unreachable): %v", err)
	}

	t.Log("Connection test passed")
}

func TestK8sResourceClient_GetServerInfo(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := client.GetServerInfo(ctx)
	if err != nil {
		t.Skipf("Failed to get server info (cluster may be unreachable): %v", err)
	}

	if version, ok := info["version"]; !ok || version == "" {
		t.Error("Expected non-empty version in server info")
	}

	if major, ok := info["major"]; !ok || major == "" {
		t.Error("Expected non-empty major version in server info")
	}

	t.Logf("Server info: %+v", info)
}

func TestRetryWrapper_Basic(t *testing.T) {
	clientset := getTestClient(t)
	baseClient := NewK8sResourceClient(clientset, "default")
	retryClient := NewRetryWrapper(baseClient, 3)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with retry wrapper
	namespaces, err := retryClient.ListNamespaces(ctx)
	if err != nil {
		t.Skipf("Failed to list namespaces with retry: %v", err)
	}

	if len(namespaces.Items) == 0 {
		t.Error("Expected at least one namespace")
	}

	t.Logf("Retry wrapper test passed, found %d namespaces", len(namespaces.Items))
}

func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		createdAt time.Time
		expected  string
	}{
		{now.Add(-30 * time.Second), "30s"},
		{now.Add(-5 * time.Minute), "5m"},
		{now.Add(-2 * time.Hour), "2h"},
		{now.Add(-3 * 24 * time.Hour), "3d"},
	}

	for _, test := range tests {
		result := formatAge(test.createdAt)
		if result != test.expected {
			t.Errorf("Expected age %s, got %s", test.expected, result)
		}
	}
}

func TestListOptions_Defaults(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "test-namespace")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with empty namespace in options - should use current namespace
	opts := ListOptions{
		Limit: 5,
	}

	// This will use the current namespace (test-namespace) even though we might not have access
	_, err := client.ListPods(ctx, opts)
	// Don't fail on error since the namespace might not exist - just test the logic
	t.Logf("ListPods with default namespace completed (error expected if namespace doesn't exist): %v", err)
}

func TestResourceClient_ErrorHandling(t *testing.T) {
	clientset := getTestClient(t)
	client := NewK8sResourceClient(clientset, "default")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test getting non-existent resource
	_, err := client.GetPod(ctx, "default", "non-existent-pod")
	if err == nil {
		t.Error("Expected error when getting non-existent pod")
	}

	t.Logf("Error handling test passed: %v", err)
}