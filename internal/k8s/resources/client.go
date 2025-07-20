package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sResourceClient implements ResourceClient using Kubernetes client-go
type K8sResourceClient struct {
	clientset        *kubernetes.Clientset
	currentNamespace string
	defaultLimit     int64
}

// NewK8sResourceClient creates a new Kubernetes resource client
func NewK8sResourceClient(clientset *kubernetes.Clientset, defaultNamespace string) *K8sResourceClient {
	return &K8sResourceClient{
		clientset:        clientset,
		currentNamespace: defaultNamespace,
		defaultLimit:     100, // Default limit for list operations
	}
}

// ListPods lists pods in the specified namespace
func (c *K8sResourceClient) ListPods(ctx context.Context, opts ListOptions) (*ResourceList[PodInfo], error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.currentNamespace
	}

	listOpts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	}

	if listOpts.Limit == 0 {
		listOpts.Limit = c.defaultLimit
	}

	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	pods := make([]PodInfo, len(podList.Items))
	for i, pod := range podList.Items {
		pods[i] = c.convertPod(&pod)
	}

	return &ResourceList[PodInfo]{
		Items:      pods,
		Total:      len(pods),
		Namespace:  namespace,
		Continue:   podList.Continue,
		Remaining:  func() int64 {
			if podList.RemainingItemCount != nil {
				return *podList.RemainingItemCount
			}
			return 0
		}(),
	}, nil
}

// GetPod gets a specific pod
func (c *K8sResourceClient) GetPod(ctx context.Context, namespace, name string) (*PodInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, name, err)
	}

	podInfo := c.convertPod(pod)
	return &podInfo, nil
}

// ListServices lists services in the specified namespace
func (c *K8sResourceClient) ListServices(ctx context.Context, opts ListOptions) (*ResourceList[ServiceInfo], error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.currentNamespace
	}

	listOpts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	}

	if listOpts.Limit == 0 {
		listOpts.Limit = c.defaultLimit
	}

	serviceList, err := c.clientset.CoreV1().Services(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make([]ServiceInfo, len(serviceList.Items))
	for i, svc := range serviceList.Items {
		services[i] = c.convertService(&svc)
	}

	return &ResourceList[ServiceInfo]{
		Items:      services,
		Total:      len(services),
		Namespace:  namespace,
		Continue:   serviceList.Continue,
		Remaining:  func() int64 {
			if serviceList.RemainingItemCount != nil {
				return *serviceList.RemainingItemCount
			}
			return 0
		}(),
	}, nil
}

// GetService gets a specific service
func (c *K8sResourceClient) GetService(ctx context.Context, namespace, name string) (*ServiceInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	svc, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s/%s: %w", namespace, name, err)
	}

	serviceInfo := c.convertService(svc)
	return &serviceInfo, nil
}

// ListDeployments lists deployments in the specified namespace
func (c *K8sResourceClient) ListDeployments(ctx context.Context, opts ListOptions) (*ResourceList[DeploymentInfo], error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.currentNamespace
	}

	listOpts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	}

	if listOpts.Limit == 0 {
		listOpts.Limit = c.defaultLimit
	}

	deploymentList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	deployments := make([]DeploymentInfo, len(deploymentList.Items))
	for i, deploy := range deploymentList.Items {
		deployments[i] = c.convertDeployment(&deploy)
	}

	return &ResourceList[DeploymentInfo]{
		Items:      deployments,
		Total:      len(deployments),
		Namespace:  namespace,
		Continue:   deploymentList.Continue,
		Remaining:  func() int64 {
			if deploymentList.RemainingItemCount != nil {
				return *deploymentList.RemainingItemCount
			}
			return 0
		}(),
	}, nil
}

// GetDeployment gets a specific deployment
func (c *K8sResourceClient) GetDeployment(ctx context.Context, namespace, name string) (*DeploymentInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	deploymentInfo := c.convertDeployment(deploy)
	return &deploymentInfo, nil
}

// ListConfigMaps lists configmaps in the specified namespace
func (c *K8sResourceClient) ListConfigMaps(ctx context.Context, opts ListOptions) (*ResourceList[ConfigMapInfo], error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.currentNamespace
	}

	listOpts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	}

	if listOpts.Limit == 0 {
		listOpts.Limit = c.defaultLimit
	}

	configMapList, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	configMaps := make([]ConfigMapInfo, len(configMapList.Items))
	for i, cm := range configMapList.Items {
		configMaps[i] = c.convertConfigMap(&cm)
	}

	return &ResourceList[ConfigMapInfo]{
		Items:      configMaps,
		Total:      len(configMaps),
		Namespace:  namespace,
		Continue:   configMapList.Continue,
		Remaining:  func() int64 {
			if configMapList.RemainingItemCount != nil {
				return *configMapList.RemainingItemCount
			}
			return 0
		}(),
	}, nil
}

// GetConfigMap gets a specific configmap
func (c *K8sResourceClient) GetConfigMap(ctx context.Context, namespace, name string) (*ConfigMapInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	cm, err := c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %w", namespace, name, err)
	}

	configMapInfo := c.convertConfigMap(cm)
	return &configMapInfo, nil
}

// ListSecrets lists secrets in the specified namespace
func (c *K8sResourceClient) ListSecrets(ctx context.Context, opts ListOptions) (*ResourceList[SecretInfo], error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.currentNamespace
	}

	listOpts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	}

	if listOpts.Limit == 0 {
		listOpts.Limit = c.defaultLimit
	}

	secretList, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secrets := make([]SecretInfo, len(secretList.Items))
	for i, secret := range secretList.Items {
		secrets[i] = c.convertSecret(&secret)
	}

	return &ResourceList[SecretInfo]{
		Items:      secrets,
		Total:      len(secrets),
		Namespace:  namespace,
		Continue:   secretList.Continue,
		Remaining:  func() int64 {
			if secretList.RemainingItemCount != nil {
				return *secretList.RemainingItemCount
			}
			return 0
		}(),
	}, nil
}

// GetSecret gets a specific secret
func (c *K8sResourceClient) GetSecret(ctx context.Context, namespace, name string) (*SecretInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	secret, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", namespace, name, err)
	}

	secretInfo := c.convertSecret(secret)
	return &secretInfo, nil
}

// ListNamespaces lists all namespaces
func (c *K8sResourceClient) ListNamespaces(ctx context.Context) (*ResourceList[NamespaceInfo], error) {
	namespaceList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]NamespaceInfo, len(namespaceList.Items))
	for i, ns := range namespaceList.Items {
		namespaces[i] = c.convertNamespace(&ns)
	}

	return &ResourceList[NamespaceInfo]{
		Items:     namespaces,
		Total:     len(namespaces),
		Namespace: "", // Global resource
	}, nil
}

// GetCurrentNamespace returns the current namespace
func (c *K8sResourceClient) GetCurrentNamespace() string {
	return c.currentNamespace
}

// SetCurrentNamespace sets the current namespace
func (c *K8sResourceClient) SetCurrentNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	c.currentNamespace = namespace
	return nil
}

// GetNamespaceContext returns namespace context information
func (c *K8sResourceClient) GetNamespaceContext() (*NamespaceContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaceList, err := c.ListNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace context: %w", err)
	}

	available := make([]string, len(namespaceList.Items))
	for i, ns := range namespaceList.Items {
		available[i] = ns.Name
	}

	return &NamespaceContext{
		Current:   c.currentNamespace,
		Available: available,
		Context:   "", // Could be populated from auth provider
	}, nil
}

// TestConnection tests the connection to the cluster
func (c *K8sResourceClient) TestConnection(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// GetServerInfo returns server information
func (c *K8sResourceClient) GetServerInfo(ctx context.Context) (map[string]interface{}, error) {
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	return map[string]interface{}{
		"version":      version.GitVersion,
		"major":        version.Major,
		"minor":        version.Minor,
		"platform":     version.Platform,
		"buildDate":    version.BuildDate,
		"goVersion":    version.GoVersion,
		"compiler":     version.Compiler,
	}, nil
}

// Helper methods for conversion

func (c *K8sResourceClient) convertPod(pod *corev1.Pod) PodInfo {
	// Calculate ready containers
	ready := 0
	total := len(pod.Spec.Containers)
	var containers []ContainerInfo

	for _, container := range pod.Spec.Containers {
		containerInfo := ContainerInfo{
			Name:  container.Name,
			Image: container.Image,
			Ready: false,
			State: "Unknown",
		}

		// Find container status
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == container.Name {
				containerInfo.Ready = status.Ready
				if status.Ready {
					ready++
				}

				if status.State.Running != nil {
					containerInfo.State = "Running"
				} else if status.State.Waiting != nil {
					containerInfo.State = "Waiting"
					containerInfo.Reason = status.State.Waiting.Reason
				} else if status.State.Terminated != nil {
					containerInfo.State = "Terminated"
					containerInfo.Reason = status.State.Terminated.Reason
				}
				break
			}
		}

		containers = append(containers, containerInfo)
	}

	// Calculate total restarts
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}

	return PodInfo{
		ResourceInfo: ResourceInfo{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Kind:        "Pod",
			APIVersion:  pod.APIVersion,
			Labels:      pod.Labels,
			Annotations: pod.Annotations,
			CreatedAt:   pod.CreationTimestamp.Time,
			Status:      string(pod.Status.Phase),
		},
		Phase:         string(pod.Status.Phase),
		Ready:         fmt.Sprintf("%d/%d", ready, total),
		Restarts:      restarts,
		Age:           formatAge(pod.CreationTimestamp.Time),
		Node:          pod.Spec.NodeName,
		IP:            pod.Status.PodIP,
		ContainerInfo: containers,
	}
}

func (c *K8sResourceClient) convertService(svc *corev1.Service) ServiceInfo {
	// Format ports
	var ports []string
	for _, port := range svc.Spec.Ports {
		portStr := strconv.Itoa(int(port.Port))
		if port.Protocol != "TCP" {
			portStr += "/" + string(port.Protocol)
		}
		if port.Name != "" {
			portStr = port.Name + ":" + portStr
		}
		ports = append(ports, portStr)
	}

	// Format selector
	var selectorParts []string
	for k, v := range svc.Spec.Selector {
		selectorParts = append(selectorParts, k+"="+v)
	}
	selector := strings.Join(selectorParts, ",")

	return ServiceInfo{
		ResourceInfo: ResourceInfo{
			Name:        svc.Name,
			Namespace:   svc.Namespace,
			Kind:        "Service",
			APIVersion:  svc.APIVersion,
			Labels:      svc.Labels,
			Annotations: svc.Annotations,
			CreatedAt:   svc.CreationTimestamp.Time,
			Status:      "Active", // Services don't have a status phase
		},
		Type:        string(svc.Spec.Type),
		ClusterIP:   svc.Spec.ClusterIP,
		ExternalIPs: svc.Spec.ExternalIPs,
		Ports:       ports,
		Selector:    selector,
		Age:         formatAge(svc.CreationTimestamp.Time),
	}
}

func (c *K8sResourceClient) convertDeployment(deploy *appsv1.Deployment) DeploymentInfo {
	// Determine strategy
	strategy := "RollingUpdate"
	if deploy.Spec.Strategy.Type == appsv1.RecreateDeploymentStrategyType {
		strategy = "Recreate"
	}

	// Determine condition
	condition := "Unknown"
	for _, cond := range deploy.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing {
			if cond.Status == corev1.ConditionTrue {
				condition = "Progressing"
			}
		}
		if cond.Type == appsv1.DeploymentAvailable {
			if cond.Status == corev1.ConditionTrue {
				condition = "Available"
			}
		}
	}

	replicas := int32(0)
	if deploy.Spec.Replicas != nil {
		replicas = *deploy.Spec.Replicas
	}

	return DeploymentInfo{
		ResourceInfo: ResourceInfo{
			Name:        deploy.Name,
			Namespace:   deploy.Namespace,
			Kind:        "Deployment",
			APIVersion:  deploy.APIVersion,
			Labels:      deploy.Labels,
			Annotations: deploy.Annotations,
			CreatedAt:   deploy.CreationTimestamp.Time,
			Status:      condition,
		},
		Replicas:          replicas,
		ReadyReplicas:     deploy.Status.ReadyReplicas,
		UpdatedReplicas:   deploy.Status.UpdatedReplicas,
		AvailableReplicas: deploy.Status.AvailableReplicas,
		Age:               formatAge(deploy.CreationTimestamp.Time),
		Strategy:          strategy,
		Condition:         condition,
	}
}

func (c *K8sResourceClient) convertConfigMap(cm *corev1.ConfigMap) ConfigMapInfo {
	dataCount := len(cm.Data) + len(cm.BinaryData)

	return ConfigMapInfo{
		ResourceInfo: ResourceInfo{
			Name:        cm.Name,
			Namespace:   cm.Namespace,
			Kind:        "ConfigMap",
			APIVersion:  cm.APIVersion,
			Labels:      cm.Labels,
			Annotations: cm.Annotations,
			CreatedAt:   cm.CreationTimestamp.Time,
			Status:      "Active",
		},
		DataCount: dataCount,
		Age:       formatAge(cm.CreationTimestamp.Time),
	}
}

func (c *K8sResourceClient) convertSecret(secret *corev1.Secret) SecretInfo {
	dataCount := len(secret.Data)

	return SecretInfo{
		ResourceInfo: ResourceInfo{
			Name:        secret.Name,
			Namespace:   secret.Namespace,
			Kind:        "Secret",
			APIVersion:  secret.APIVersion,
			Labels:      secret.Labels,
			Annotations: secret.Annotations,
			CreatedAt:   secret.CreationTimestamp.Time,
			Status:      "Active",
		},
		Type:      string(secret.Type),
		DataCount: dataCount,
		Age:       formatAge(secret.CreationTimestamp.Time),
	}
}

func (c *K8sResourceClient) convertNamespace(ns *corev1.Namespace) NamespaceInfo {
	return NamespaceInfo{
		ResourceInfo: ResourceInfo{
			Name:        ns.Name,
			Namespace:   "", // Namespaces are cluster-scoped
			Kind:        "Namespace",
			APIVersion:  ns.APIVersion,
			Labels:      ns.Labels,
			Annotations: ns.Annotations,
			CreatedAt:   ns.CreationTimestamp.Time,
			Status:      string(ns.Status.Phase),
		},
		Phase: string(ns.Status.Phase),
		Age:   formatAge(ns.CreationTimestamp.Time),
	}
}

// formatAge formats a time duration as a human-readable age string
func formatAge(createdAt time.Time) string {
	age := time.Since(createdAt)
	
	days := int(age.Hours()) / 24
	hours := int(age.Hours()) % 24
	minutes := int(age.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	} else {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	}
}