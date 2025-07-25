package projects

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/katyella/lazyoc/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestProjectType_String(t *testing.T) {
	tests := []struct {
		projectType ProjectType
		expected    string
	}{
		{ProjectTypeUnknown, "Unknown"},
		{ProjectTypeKubernetesNamespace, "Namespace"},
		{ProjectTypeOpenShiftProject, "Project"},
	}

	for _, test := range tests {
		if result := test.projectType.String(); result != test.expected {
			t.Errorf("ProjectType.String() = %s, expected %s", result, test.expected)
		}
	}
}

func TestKubernetesNamespaceManager_List(t *testing.T) {
	// Create fake clientset with test namespaces
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
				Labels: map[string]string{
					"env": "system",
				},
				Annotations: map[string]string{
					"description": "Default namespace",
				},
				CreationTimestamp: metav1.NewTime(time.Now().Add(-24 * time.Hour)),
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
				Labels: map[string]string{
					"env": "testing",
				},
				CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Hour)),
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
	)

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()
	projects, err := manager.List(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	// Check first project (should be sorted by name: "default")
	defaultProject := projects[0]
	if defaultProject.Name != "default" {
		t.Errorf("Expected first project name 'default', got '%s'", defaultProject.Name)
	}
	if defaultProject.Type != ProjectTypeKubernetesNamespace {
		t.Errorf("Expected project type 'Namespace', got '%s'", defaultProject.Type)
	}
	if defaultProject.ClusterType != k8s.ClusterTypeKubernetes {
		t.Errorf("Expected cluster type 'Kubernetes', got '%s'", defaultProject.ClusterType)
	}
	if defaultProject.Description != "Default namespace" {
		t.Errorf("Expected description 'Default namespace', got '%s'", defaultProject.Description)
	}

	// Check labels
	if env, ok := defaultProject.Labels["env"]; !ok || env != "system" {
		t.Errorf("Expected label env=system, got %v", defaultProject.Labels)
	}
}

func TestKubernetesNamespaceManager_Get(t *testing.T) {
	// Create fake clientset
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
				Labels: map[string]string{
					"env": "testing",
				},
				Annotations: map[string]string{
					"description": "Test namespace",
				},
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
	)

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()
	project, err := manager.Get(ctx, "test-namespace")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if project.Name != "test-namespace" {
		t.Errorf("Expected name 'test-namespace', got '%s'", project.Name)
	}
	if project.Type != ProjectTypeKubernetesNamespace {
		t.Errorf("Expected type 'Namespace', got '%s'", project.Type)
	}
	if project.Description != "Test namespace" {
		t.Errorf("Expected description 'Test namespace', got '%s'", project.Description)
	}
}

func TestKubernetesNamespaceManager_Create(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()
	opts := CreateOptions{
		Description: "Created namespace",
		Labels: map[string]string{
			"env": "test",
		},
		Annotations: map[string]string{
			"created-by": "test",
		},
	}

	project, err := manager.Create(ctx, "new-namespace", opts)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if project.Name != "new-namespace" {
		t.Errorf("Expected name 'new-namespace', got '%s'", project.Name)
	}
	if project.Description != "Created namespace" {
		t.Errorf("Expected description 'Created namespace', got '%s'", project.Description)
	}

	// Verify the namespace was actually created
	ns, err := fakeClientset.CoreV1().Namespaces().Get(ctx, "new-namespace", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get created namespace: %v", err)
	}
	if ns.Name != "new-namespace" {
		t.Errorf("Created namespace has wrong name: %s", ns.Name)
	}
	if ns.Annotations["description"] != "Created namespace" {
		t.Errorf("Created namespace missing description annotation")
	}
	if ns.Labels["env"] != "test" {
		t.Errorf("Created namespace missing env label")
	}
}

func TestKubernetesNamespaceManager_Delete(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "to-delete",
			},
		},
	)

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()
	err := manager.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify the namespace was deleted
	_, err = fakeClientset.CoreV1().Namespaces().Get(ctx, "to-delete", metav1.GetOptions{})
	if err == nil {
		t.Errorf("Expected namespace to be deleted, but it still exists")
	}
}

func TestKubernetesNamespaceManager_Exists(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "existing-namespace",
			},
		},
	)

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()

	// Test existing namespace
	exists, err := manager.Exists(ctx, "existing-namespace")
	if err != nil {
		t.Fatalf("Exists() failed: %v", err)
	}
	if !exists {
		t.Errorf("Expected namespace to exist")
	}

	// Test non-existing namespace
	exists, err = manager.Exists(ctx, "non-existing")
	if err != nil {
		t.Fatalf("Exists() failed for non-existing namespace: %v", err)
	}
	if exists {
		t.Errorf("Expected namespace to not exist")
	}
}

func TestKubernetesNamespaceManager_GetProjectType(t *testing.T) {
	manager := NewKubernetesNamespaceManager(nil, nil, "")

	projectType := manager.GetProjectType()
	if projectType != ProjectTypeKubernetesNamespace {
		t.Errorf("Expected ProjectTypeKubernetesNamespace, got %s", projectType)
	}
}

func TestKubernetesNamespaceManager_GetClusterType(t *testing.T) {
	manager := NewKubernetesNamespaceManager(nil, nil, "")

	clusterType := manager.GetClusterType()
	if clusterType != k8s.ClusterTypeKubernetes {
		t.Errorf("Expected ClusterTypeKubernetes, got %s", clusterType)
	}
}

func TestKubernetesNamespaceManager_ListWithFilters(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prod-namespace",
				Labels: map[string]string{
					"env": "production",
				},
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dev-namespace",
				Labels: map[string]string{
					"env": "development",
				},
			},
		},
	)

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()

	// Test with label selector
	projects, err := manager.List(ctx, ListOptions{
		LabelSelector: "env=production",
	})
	if err != nil {
		t.Fatalf("List() with label selector failed: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project with env=production, got %d", len(projects))
	}

	if len(projects) > 0 && projects[0].Name != "prod-namespace" {
		t.Errorf("Expected prod-namespace, got %s", projects[0].Name)
	}
}

func TestKubernetesNamespaceManager_ResourceQuotas(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		},
		&corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-quota",
				Namespace: "test-namespace",
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{
					corev1.ResourcePods: resource.MustParse("10"),
				},
			},
			Status: corev1.ResourceQuotaStatus{
				Used: corev1.ResourceList{
					corev1.ResourcePods: resource.MustParse("5"),
				},
			},
		},
	)

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()
	project, err := manager.Get(ctx, "test-namespace")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if len(project.ResourceQuotas) != 1 {
		t.Errorf("Expected 1 resource quota, got %d", len(project.ResourceQuotas))
	}

	quota := project.ResourceQuotas[0]
	if quota.Name != "test-quota" {
		t.Errorf("Expected quota name 'test-quota', got '%s'", quota.Name)
	}
	if quota.Hard["pods"] != "10" {
		t.Errorf("Expected hard limit pods=10, got %s", quota.Hard["pods"])
	}
	if quota.Used["pods"] != "5" {
		t.Errorf("Expected used pods=5, got %s", quota.Used["pods"])
	}
}

// Test error conditions
func TestKubernetesNamespaceManager_ErrorHandling(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Add reactor to simulate API errors
	fakeClientset.PrependReactor("get", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("simulated API error")
	})

	manager := NewKubernetesNamespaceManager(fakeClientset, nil, "")

	ctx := context.Background()

	// Test Get with error
	_, err := manager.Get(ctx, "test")
	if err == nil {
		t.Errorf("Expected Get() to fail with simulated error")
	}

	// Test Exists with error (should return error, not just false)
	_, err = manager.Exists(ctx, "test")
	if err == nil {
		t.Errorf("Expected Exists() to return error")
	}
}
