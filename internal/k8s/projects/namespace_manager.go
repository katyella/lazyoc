package projects

import (
	"context"
	"fmt"
	"sort"

	"github.com/katyella/lazyoc/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesNamespaceManager implements ProjectManager for vanilla Kubernetes namespaces
type KubernetesNamespaceManager struct {
	clientset      kubernetes.Interface
	config         *rest.Config
	kubeconfigPath string
	clusterType    k8s.ClusterType
}

// NewKubernetesNamespaceManager creates a new namespace manager for Kubernetes
func NewKubernetesNamespaceManager(clientset kubernetes.Interface, config *rest.Config, kubeconfigPath string) *KubernetesNamespaceManager {
	return &KubernetesNamespaceManager{
		clientset:      clientset,
		config:         config,
		kubeconfigPath: kubeconfigPath,
		clusterType:    k8s.ClusterTypeKubernetes,
	}
}

// List all accessible namespaces
func (m *KubernetesNamespaceManager) List(ctx context.Context, opts ListOptions) ([]ProjectInfo, error) {
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

		// Optionally include quotas and limits
		if opts.IncludeQuotas {
			quotas, _ := m.getResourceQuotas(ctx, ns.Name)
			project.ResourceQuotas = quotas
		}
		if opts.IncludeLimits {
			limits, _ := m.getLimitRanges(ctx, ns.Name)
			project.LimitRanges = limits
		}

		projects = append(projects, project)
	}

	// Sort by name for consistent ordering
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// Get detailed information about a specific namespace
func (m *KubernetesNamespaceManager) Get(ctx context.Context, name string) (*ProjectInfo, error) {
	namespace, err := m.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %s: %w", name, err)
	}

	project := m.convertNamespaceToProject(namespace)

	// Always include quotas and limits for detailed view
	quotas, _ := m.getResourceQuotas(ctx, name)
	project.ResourceQuotas = quotas

	limits, _ := m.getLimitRanges(ctx, name)
	project.LimitRanges = limits

	return &project, nil
}

// Create a new namespace
func (m *KubernetesNamespaceManager) Create(ctx context.Context, name string, opts CreateOptions) (*ProjectInfo, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
	}

	// Add description as annotation if provided
	if opts.Description != "" {
		if namespace.Annotations == nil {
			namespace.Annotations = make(map[string]string)
		}
		namespace.Annotations["description"] = opts.Description
	}

	createdNS, err := m.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace %s: %w", name, err)
	}

	project := m.convertNamespaceToProject(createdNS)

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

	return &project, nil
}

// Delete a namespace
func (m *KubernetesNamespaceManager) Delete(ctx context.Context, name string) error {
	err := m.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", name, err)
	}
	return nil
}

// SwitchTo switches to a different namespace by updating kubeconfig
func (m *KubernetesNamespaceManager) SwitchTo(ctx context.Context, name string) (*SwitchResult, error) {
	if m.kubeconfigPath == "" {
		return &SwitchResult{
			Success: false,
			Message: "No kubeconfig path available",
		}, fmt.Errorf("kubeconfig path not set")
	}

	// Check if namespace exists first
	exists, err := m.Exists(ctx, name)
	if err != nil {
		return &SwitchResult{
			Success: false,
			Message: fmt.Sprintf("Failed to check if namespace exists: %v", err),
		}, err
	}
	if !exists {
		return &SwitchResult{
			Success: false,
			Message: fmt.Sprintf("Namespace '%s' does not exist", name),
		}, fmt.Errorf("namespace %s does not exist", name)
	}

	// Get current namespace for comparison
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

	// Update the namespace
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
		Message:     fmt.Sprintf("Switched to namespace '%s'", name),
		ProjectInfo: projectInfo,
	}, nil
}

// GetCurrent returns the current namespace from kubeconfig
func (m *KubernetesNamespaceManager) GetCurrent(ctx context.Context) (*ProjectInfo, error) {
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

// Exists checks if a namespace exists and is accessible
func (m *KubernetesNamespaceManager) Exists(ctx context.Context, name string) (bool, error) {
	_, err := m.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetProjectType returns the project type this manager handles
func (m *KubernetesNamespaceManager) GetProjectType() ProjectType {
	return ProjectTypeKubernetesNamespace
}

// GetClusterType returns the cluster type
func (m *KubernetesNamespaceManager) GetClusterType() k8s.ClusterType {
	return m.clusterType
}

// RefreshCache refreshes any cached information (no-op for this implementation)
func (m *KubernetesNamespaceManager) RefreshCache(ctx context.Context) error {
	// No caching in this basic implementation
	return nil
}

// Helper methods

// convertNamespaceToProject converts a Kubernetes namespace to ProjectInfo
func (m *KubernetesNamespaceManager) convertNamespaceToProject(ns *corev1.Namespace) ProjectInfo {
	project := ProjectInfo{
		Name:        ns.Name,
		DisplayName: ns.Name, // Kubernetes namespaces don't have separate display names
		Labels:      ns.Labels,
		Annotations: ns.Annotations,
		CreatedAt:   ns.CreationTimestamp.Time,
		Status:      string(ns.Status.Phase),
		Type:        ProjectTypeKubernetesNamespace,
		ClusterType: k8s.ClusterTypeKubernetes,
	}

	// Extract description from annotation if present
	if ns.Annotations != nil {
		if desc, ok := ns.Annotations["description"]; ok {
			project.Description = desc
		}
	}

	return project
}

// getResourceQuotas retrieves resource quotas for a namespace
func (m *KubernetesNamespaceManager) getResourceQuotas(ctx context.Context, namespace string) ([]ResourceQuota, error) {
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

		// Convert hard limits
		for k, v := range quota.Spec.Hard {
			rq.Hard[string(k)] = v.String()
		}

		// Convert used resources
		for k, v := range quota.Status.Used {
			rq.Used[string(k)] = v.String()
		}

		result = append(result, rq)
	}

	return result, nil
}

// getLimitRanges retrieves limit ranges for a namespace
func (m *KubernetesNamespaceManager) getLimitRanges(ctx context.Context, namespace string) ([]LimitRange, error) {
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

			// Convert limits
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

// createResourceQuota creates a resource quota in the namespace
func (m *KubernetesNamespaceManager) createResourceQuota(ctx context.Context, namespace string, quota ResourceQuota) error {
	// Convert to Kubernetes resource quota
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

	// Convert hard limits
	for k, v := range quota.Hard {
		if quantity, err := resource.ParseQuantity(v); err == nil {
			rq.Spec.Hard[corev1.ResourceName(k)] = quantity
		}
	}

	_, err := m.clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, rq, metav1.CreateOptions{})
	return err
}

// createLimitRange creates a limit range in the namespace
func (m *KubernetesNamespaceManager) createLimitRange(ctx context.Context, namespace string, limitRange LimitRange) error {
	// Convert to Kubernetes limit range
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      limitRange.Name,
			Namespace: namespace,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: make([]corev1.LimitRangeItem, len(limitRange.Limits)),
		},
	}

	// Convert limit items
	for i, item := range limitRange.Limits {
		lri := corev1.LimitRangeItem{
			Type:           corev1.LimitType(item.Type),
			Max:            make(corev1.ResourceList),
			Min:            make(corev1.ResourceList),
			Default:        make(corev1.ResourceList),
			DefaultRequest: make(corev1.ResourceList),
		}

		// Convert resource limits
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
