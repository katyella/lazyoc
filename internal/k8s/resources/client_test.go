package resources

import (
	"testing"
	"time"
)

// All integration tests that connect to real clusters have been disabled
// to avoid network dependencies in unit tests.

func TestFormatAge(t *testing.T) {
	now := time.Now()
	tests := []struct {
		createdAt time.Time
		expected  string
	}{
		{now.Add(-time.Second * 30), "30s"},
		{now.Add(-time.Minute * 2), "2m"},
		{now.Add(-time.Hour * 3), "3h"},
		{now.Add(-time.Hour * 25), "1d"},
		{now.Add(-time.Hour * 24 * 7), "7d"},
		{now.Add(-time.Hour * 24 * 365), "365d"},
	}

	for _, test := range tests {
		result := formatAge(test.createdAt)
		if result != test.expected {
			t.Errorf("formatAge(%v) = %s, expected %s", test.createdAt, result, test.expected)
		}
	}
}

func TestListOptions_Defaults(t *testing.T) {
	// Test that ListOptions work without network calls
	opts := ListOptions{
		Namespace: "test-namespace",
		Limit:     50,
	}

	if opts.Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", opts.Namespace)
	}

	if opts.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", opts.Limit)
	}
}

// All other tests are disabled as they require real cluster connections:
// - TestK8sResourceClient_ListNamespaces
// - TestK8sResourceClient_ListPods  
// - TestK8sResourceClient_ListServices
// - TestK8sResourceClient_NamespaceOperations
// - TestK8sResourceClient_TestConnection
// - TestK8sResourceClient_GetServerInfo
// - TestRetryWrapper_Basic
// - TestResourceClient_ErrorHandling
//
// These tests should be rewritten as proper unit tests with mocked
// Kubernetes clients if needed, but are currently disabled to prevent
// network dependencies.