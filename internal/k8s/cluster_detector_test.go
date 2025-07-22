package k8s

import (
	"testing"
)

func TestClusterType_String(t *testing.T) {
	tests := []struct {
		clusterType ClusterType
		expected    string
	}{
		{ClusterTypeUnknown, "Unknown"},
		{ClusterTypeKubernetes, "Kubernetes"},
		{ClusterTypeOpenShift, "OpenShift"},
	}
	
	for _, test := range tests {
		if result := test.clusterType.String(); result != test.expected {
			t.Errorf("ClusterType.String() = %s, expected %s", result, test.expected)
		}
	}
}

func TestNewClusterTypeDetector(t *testing.T) {
	// Test with nil config should fail
	detector, err := NewClusterTypeDetector(nil)
	if err == nil {
		t.Error("Expected error when creating detector with nil config")
	}
	if detector != nil {
		t.Error("Expected nil detector when config is nil")
	}
}

func TestClusterTypeDetector_CacheManagement(t *testing.T) {
	// We can't easily test the full detector without a real cluster,
	// but we can test cache management methods
	detector := &ClusterTypeDetector{
		cached: true,
		cachedInfo: &ClusterInfo{
			Type: ClusterTypeKubernetes,
		},
	}
	
	// Test ClearCache
	detector.ClearCache()
	if detector.cached {
		t.Error("Expected cached to be false after ClearCache")
	}
	if detector.cachedInfo != nil {
		t.Error("Expected cachedInfo to be nil after ClearCache")
	}
	
	// Test SetCacheTime
	detector.SetCacheTime(5 * 60) // 5 minutes
	if detector.cacheTime != 5*60 {
		t.Error("Expected cacheTime to be set correctly")
	}
}