package projects

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/katyella/lazyoc/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// OpenShiftProjectManager implements ProjectManager for OpenShift projects
type OpenShiftProjectManager struct {
	clientset      kubernetes.Interface
	dynamicClient  dynamic.Interface
	config         *rest.Config
	kubeconfigPath string
	clusterType    k8s.ClusterType

	// OpenShift API resources
	projectResource        schema.GroupVersionResource
	projectRequestResource schema.GroupVersionResource
}

// NewOpenShiftProjectManager creates a new project manager for OpenShift
func NewOpenShiftProjectManager(clientset kubernetes.Interface, dynamicClient dynamic.Interface, config *rest.Config, kubeconfigPath string) *OpenShiftProjectManager {
	return &OpenShiftProjectManager{
		clientset:      clientset,
		dynamicClient:  dynamicClient,
		config:         config,
		kubeconfigPath: kubeconfigPath,
		clusterType:    k8s.ClusterTypeOpenShift,

		// OpenShift API resources
		projectResource: schema.GroupVersionResource{
			Group:    "project.openshift.io",
			Version:  "v1",
			Resource: "projects",
		},
		projectRequestResource: schema.GroupVersionResource{
			Group:    "project.openshift.io",
			Version:  "v1",
			Resource: "projectrequests",
		},
	}
}

// List all accessible projects
func (m *OpenShiftProjectManager) List(ctx context.Context, opts ListOptions) ([]ProjectInfo, error) {
	// Use the dynamic client to query OpenShift projects
	listOpts := metav1.ListOptions{}
	if opts.LabelSelector != "" {
		listOpts.LabelSelector = opts.LabelSelector
	}
	if opts.FieldSelector != "" {
		listOpts.FieldSelector = opts.FieldSelector
	}

	projectList, err := m.dynamicClient.Resource(m.projectResource).List(ctx, listOpts)
	if err != nil {
		// Fallback to namespaces if projects API is not available
		return m.listNamespacesAsFallback(ctx, opts)
	}

	var projects []ProjectInfo
	for _, item := range projectList.Items {
		project, err := m.convertUnstructuredToProject(&item)
		if err != nil {
			continue // Skip invalid projects
		}

		// Optionally include quotas and limits
		if opts.IncludeQuotas {
			quotas, _ := m.getResourceQuotas(ctx, project.Name)
			project.ResourceQuotas = quotas
		}
		if opts.IncludeLimits {
			limits, _ := m.getLimitRanges(ctx, project.Name)
			project.LimitRanges = limits
		}

		projects = append(projects, *project)
	}

	// Sort by name for consistent ordering
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// Get detailed information about a specific project
func (m *OpenShiftProjectManager) Get(ctx context.Context, name string) (*ProjectInfo, error) {
	project, err := m.dynamicClient.Resource(m.projectResource).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// Fallback to namespace
		return m.getNamespaceAsFallback(ctx, name)
	}

	projectInfo, err := m.convertUnstructuredToProject(project)
	if err != nil {
		return nil, fmt.Errorf("failed to convert project: %w", err)
	}

	// Always include quotas and limits for detailed view
	quotas, _ := m.getResourceQuotas(ctx, name)
	projectInfo.ResourceQuotas = quotas

	limits, _ := m.getLimitRanges(ctx, name)
	projectInfo.LimitRanges = limits

	return projectInfo, nil
}

// Create a new project
func (m *OpenShiftProjectManager) Create(ctx context.Context, name string, opts CreateOptions) (*ProjectInfo, error) {
	// Create project request
	projectRequest := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "project.openshift.io/v1",
			"kind":       "ProjectRequest",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}

	// Add display name if provided
	if opts.DisplayName != "" {
		metadata, ok := projectRequest.Object["metadata"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("metadata is not a map")
		}
		annotations := make(map[string]interface{})
		if existingAnnotations, ok := metadata["annotations"]; ok {
			if annotationsMap, ok := existingAnnotations.(map[string]interface{}); ok {
				annotations = annotationsMap
			}
		}
		annotations["openshift.io/display-name"] = opts.DisplayName
		metadata["annotations"] = annotations
	}

	// Add description if provided
	if opts.Description != "" {
		metadata, ok := projectRequest.Object["metadata"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("metadata is not a map")
		}
		annotations := make(map[string]interface{})
		if existingAnnotations, ok := metadata["annotations"]; ok {
			if annotationsMap, ok := existingAnnotations.(map[string]interface{}); ok {
				annotations = annotationsMap
			}
		}
		annotations["openshift.io/description"] = opts.Description
		metadata["annotations"] = annotations
	}

	// Add labels if provided
	if len(opts.Labels) > 0 {
		metadata, ok := projectRequest.Object["metadata"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("metadata is not a map")
		}
		labels := make(map[string]interface{})
		for k, v := range opts.Labels {
			labels[k] = v
		}
		metadata["labels"] = labels
	}

	// Add additional annotations if provided
	if len(opts.Annotations) > 0 {
		metadata, ok := projectRequest.Object["metadata"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("metadata is not a map")
		}
		annotations := make(map[string]interface{})
		if existingAnnotations, ok := metadata["annotations"]; ok {
			if annotationsMap, ok := existingAnnotations.(map[string]interface{}); ok {
				annotations = annotationsMap
			}
		}
		for k, v := range opts.Annotations {
			annotations[k] = v
		}
		metadata["annotations"] = annotations
	}

	// Create the project
	createdProject, err := m.dynamicClient.Resource(m.projectRequestResource).Create(ctx, projectRequest, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create project %s: %w", name, err)
	}

	// Convert to ProjectInfo
	project, err := m.convertUnstructuredToProject(createdProject)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created project: %w", err)
	}

	// Create resource quotas if specified
	for _, quota := range opts.ResourceQuotas {
		err := m.createResourceQuota(ctx, name, quota)
		if err != nil {
			// Log warning silently, don't interfere with TUI
			continue
		}
	}

	// Create limit ranges if specified
	for _, limitRange := range opts.LimitRanges {
		err := m.createLimitRange(ctx, name, limitRange)
		if err != nil {
			// Log warning silently, don't interfere with TUI
			continue
		}
	}

	return project, nil
}

// Delete a project
func (m *OpenShiftProjectManager) Delete(ctx context.Context, name string) error {
	err := m.dynamicClient.Resource(m.projectResource).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete project %s: %w", name, err)
	}
	return nil
}

// SwitchTo switches to a different project by updating kubeconfig
func (m *OpenShiftProjectManager) SwitchTo(ctx context.Context, name string) (*SwitchResult, error) {
	if m.kubeconfigPath == "" {
		return &SwitchResult{
			Success: false,
			Message: "No kubeconfig path available",
		}, fmt.Errorf("kubeconfig path not set")
	}

	// Check if project exists first
	exists, err := m.Exists(ctx, name)
	if err != nil {
		return &SwitchResult{
			Success: false,
			Message: fmt.Sprintf("Failed to check if project exists: %v", err),
		}, err
	}
	if !exists {
		return &SwitchResult{
			Success: false,
			Message: fmt.Sprintf("Project '%s' does not exist", name),
		}, fmt.Errorf("project %s does not exist", name)
	}

	// Get current project for comparison
	currentProject, _ := m.GetCurrent(ctx)
	currentName := ""
	if currentProject != nil {
		currentName = currentProject.Name
	}

	// Load and modify kubeconfig
	config, err := clientcmd.LoadFromFile(m.kubeconfigPath)
	if err != nil {
		return &SwitchResult{
			From:    currentName,
			To:      name,
			Success: false,
			Message: fmt.Sprintf("Failed to load kubeconfig: %v", err),
		}, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Update the current context's namespace
	if config.CurrentContext == "" {
		return &SwitchResult{
			From:    currentName,
			To:      name,
			Success: false,
			Message: "No current context set in kubeconfig",
		}, fmt.Errorf("no current context set")
	}

	currentContext := config.Contexts[config.CurrentContext]
	if currentContext == nil {
		return &SwitchResult{
			From:    currentName,
			To:      name,
			Success: false,
			Message: fmt.Sprintf("Current context '%s' not found in kubeconfig", config.CurrentContext),
		}, fmt.Errorf("current context not found")
	}

	// Update the namespace (OpenShift projects are backed by namespaces)
	currentContext.Namespace = name

	// Write back to kubeconfig
	err = clientcmd.WriteToFile(*config, m.kubeconfigPath)
	if err != nil {
		return &SwitchResult{
			From:    currentName,
			To:      name,
			Success: false,
			Message: fmt.Sprintf("Failed to write kubeconfig: %v", err),
		}, fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	// Get project info for the result
	projectInfo, _ := m.Get(ctx, name)

	return &SwitchResult{
		From:        currentName,
		To:          name,
		Success:     true,
		Message:     fmt.Sprintf("Switched to project '%s'", name),
		ProjectInfo: projectInfo,
	}, nil
}

// GetCurrent returns the current project from kubeconfig
func (m *OpenShiftProjectManager) GetCurrent(ctx context.Context) (*ProjectInfo, error) {
	if m.kubeconfigPath == "" {
		return nil, fmt.Errorf("no kubeconfig path available")
	}

	config, err := clientcmd.LoadFromFile(m.kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if config.CurrentContext == "" {
		return nil, fmt.Errorf("no current context set")
	}

	currentContext := config.Contexts[config.CurrentContext]
	if currentContext == nil {
		return nil, fmt.Errorf("current context not found")
	}

	namespace := currentContext.Namespace
	if namespace == "" {
		namespace = "default"
	}

	return m.Get(ctx, namespace)
}

// Exists checks if a project exists and is accessible
func (m *OpenShiftProjectManager) Exists(ctx context.Context, name string) (bool, error) {
	_, err := m.dynamicClient.Resource(m.projectResource).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetProjectType returns the project type this manager handles
func (m *OpenShiftProjectManager) GetProjectType() ProjectType {
	return ProjectTypeOpenShiftProject
}

// GetClusterType returns the cluster type
func (m *OpenShiftProjectManager) GetClusterType() k8s.ClusterType {
	return m.clusterType
}

// RefreshCache refreshes any cached information (no-op for this implementation)
func (m *OpenShiftProjectManager) RefreshCache(ctx context.Context) error {
	// No caching in this basic implementation
	return nil
}

// Helper methods

// convertUnstructuredToProject converts an unstructured OpenShift project to ProjectInfo
func (m *OpenShiftProjectManager) convertUnstructuredToProject(obj *unstructured.Unstructured) (*ProjectInfo, error) {
	name, found, err := unstructured.NestedString(obj.Object, "metadata", "name")
	if err != nil || !found {
		return nil, fmt.Errorf("project name not found")
	}

	project := &ProjectInfo{
		Name:        name,
		DisplayName: name, // Default to name if display name not found
		Type:        ProjectTypeOpenShiftProject,
		ClusterType: k8s.ClusterTypeOpenShift,
	}

	// Extract metadata
	if metadata, found, _ := unstructured.NestedMap(obj.Object, "metadata"); found {
		// Labels
		if labels, found, _ := unstructured.NestedStringMap(metadata, "labels"); found {
			project.Labels = labels
		}

		// Annotations
		if annotations, found, _ := unstructured.NestedStringMap(metadata, "annotations"); found {
			project.Annotations = annotations

			// Extract OpenShift-specific fields from annotations
			if displayName, ok := annotations["openshift.io/display-name"]; ok {
				project.DisplayName = displayName
			}
			if description, ok := annotations["openshift.io/description"]; ok {
				project.Description = description
			}
			if requester, ok := annotations["openshift.io/requester"]; ok {
				project.Requester = requester
			}
		}

		// Creation timestamp
		if creationTimestamp, found, _ := unstructured.NestedString(metadata, "creationTimestamp"); found {
			if t, err := time.Parse(time.RFC3339, creationTimestamp); err == nil {
				project.CreatedAt = t
			}
		}
	}

	// Extract status
	if status, found, _ := unstructured.NestedMap(obj.Object, "status"); found {
		if phase, found, _ := unstructured.NestedString(status, "phase"); found {
			project.Status = phase
		}
	}

	if project.Status == "" {
		project.Status = "Active" // Default for OpenShift projects
	}

	return project, nil
}

// listNamespacesAsFallback lists namespaces when projects API is not available
func (m *OpenShiftProjectManager) listNamespacesAsFallback(ctx context.Context, opts ListOptions) ([]ProjectInfo, error) {
	listOpts := metav1.ListOptions{}
	if opts.LabelSelector != "" {
		listOpts.LabelSelector = opts.LabelSelector
	}
	if opts.FieldSelector != "" {
		listOpts.FieldSelector = opts.FieldSelector
	}

	namespaces, err := m.clientset.CoreV1().Namespaces().List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var projects []ProjectInfo
	for _, ns := range namespaces.Items {
		project := m.convertNamespaceToProject(&ns)
		projects = append(projects, project)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// getNamespaceAsFallback gets namespace information when project API is not available
func (m *OpenShiftProjectManager) getNamespaceAsFallback(ctx context.Context, name string) (*ProjectInfo, error) {
	namespace, err := m.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %s: %w", name, err)
	}

	project := m.convertNamespaceToProject(namespace)
	return &project, nil
}

// convertNamespaceToProject converts a Kubernetes namespace to ProjectInfo (for fallback)
func (m *OpenShiftProjectManager) convertNamespaceToProject(ns *corev1.Namespace) ProjectInfo {
	project := ProjectInfo{
		Name:        ns.Name,
		DisplayName: ns.Name,
		Labels:      ns.Labels,
		Annotations: ns.Annotations,
		CreatedAt:   ns.CreationTimestamp.Time,
		Status:      string(ns.Status.Phase),
		Type:        ProjectTypeOpenShiftProject, // Treat as OpenShift project
		ClusterType: k8s.ClusterTypeOpenShift,
	}

	// Extract OpenShift-specific fields from annotations if present
	if ns.Annotations != nil {
		if displayName, ok := ns.Annotations["openshift.io/display-name"]; ok {
			project.DisplayName = displayName
		}
		if description, ok := ns.Annotations["openshift.io/description"]; ok {
			project.Description = description
		}
		if requester, ok := ns.Annotations["openshift.io/requester"]; ok {
			project.Requester = requester
		}
	}

	return project
}

// getResourceQuotas retrieves resource quotas for a project/namespace
func (m *OpenShiftProjectManager) getResourceQuotas(ctx context.Context, namespace string) ([]ResourceQuota, error) {
	quotas, err := m.clientset.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ResourceQuota
	for _, quota := range quotas.Items {
		rq := ResourceQuota{
			Name:   quota.Name,
			Hard:   make(map[string]string),
			Used:   make(map[string]string),
			Scopes: make([]string, len(quota.Spec.Scopes)),
		}

		// Convert scopes
		for i, scope := range quota.Spec.Scopes {
			rq.Scopes[i] = string(scope)
		}

		for k, v := range quota.Spec.Hard {
			rq.Hard[string(k)] = v.String()
		}
		for k, v := range quota.Status.Used {
			rq.Used[string(k)] = v.String()
		}

		result = append(result, rq)
	}

	return result, nil
}

// getLimitRanges retrieves limit ranges for a project/namespace
func (m *OpenShiftProjectManager) getLimitRanges(ctx context.Context, namespace string) ([]LimitRange, error) {
	limits, err := m.clientset.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []LimitRange
	for _, limit := range limits.Items {
		lr := LimitRange{
			Name:   limit.Name,
			Limits: make([]LimitItem, len(limit.Spec.Limits)),
		}

		for i, item := range limit.Spec.Limits {
			li := LimitItem{
				Type:           string(item.Type),
				Max:            make(map[string]string),
				Min:            make(map[string]string),
				Default:        make(map[string]string),
				DefaultRequest: make(map[string]string),
			}

			for k, v := range item.Max {
				li.Max[string(k)] = v.String()
			}
			for k, v := range item.Min {
				li.Min[string(k)] = v.String()
			}
			for k, v := range item.Default {
				li.Default[string(k)] = v.String()
			}
			for k, v := range item.DefaultRequest {
				li.DefaultRequest[string(k)] = v.String()
			}

			lr.Limits[i] = li
		}

		result = append(result, lr)
	}

	return result, nil
}

// createResourceQuota creates a resource quota in the project/namespace
func (m *OpenShiftProjectManager) createResourceQuota(ctx context.Context, namespace string, quota ResourceQuota) error {
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      quota.Name,
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard:   make(corev1.ResourceList),
			Scopes: make([]corev1.ResourceQuotaScope, len(quota.Scopes)),
		},
	}

	// Convert scopes
	for i, scope := range quota.Scopes {
		rq.Spec.Scopes[i] = corev1.ResourceQuotaScope(scope)
	}

	for k, v := range quota.Hard {
		if quantity, err := resource.ParseQuantity(v); err == nil {
			rq.Spec.Hard[corev1.ResourceName(k)] = quantity
		}
	}

	_, err := m.clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, rq, metav1.CreateOptions{})
	return err
}

// createLimitRange creates a limit range in the project/namespace
func (m *OpenShiftProjectManager) createLimitRange(ctx context.Context, namespace string, limitRange LimitRange) error {
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      limitRange.Name,
			Namespace: namespace,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: make([]corev1.LimitRangeItem, len(limitRange.Limits)),
		},
	}

	for i, item := range limitRange.Limits {
		lri := corev1.LimitRangeItem{
			Type:           corev1.LimitType(item.Type),
			Max:            make(corev1.ResourceList),
			Min:            make(corev1.ResourceList),
			Default:        make(corev1.ResourceList),
			DefaultRequest: make(corev1.ResourceList),
		}

		for k, v := range item.Max {
			if quantity, err := resource.ParseQuantity(v); err == nil {
				lri.Max[corev1.ResourceName(k)] = quantity
			}
		}
		for k, v := range item.Min {
			if quantity, err := resource.ParseQuantity(v); err == nil {
				lri.Min[corev1.ResourceName(k)] = quantity
			}
		}
		for k, v := range item.Default {
			if quantity, err := resource.ParseQuantity(v); err == nil {
				lri.Default[corev1.ResourceName(k)] = quantity
			}
		}
		for k, v := range item.DefaultRequest {
			if quantity, err := resource.ParseQuantity(v); err == nil {
				lri.DefaultRequest[corev1.ResourceName(k)] = quantity
			}
		}

		lr.Spec.Limits[i] = lri
	}

	_, err := m.clientset.CoreV1().LimitRanges(namespace).Create(ctx, lr, metav1.CreateOptions{})
	return err
}
