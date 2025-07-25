package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// ResourcesView handles generic resource views for services, deployments, etc.
type ResourcesView struct {
	resourceType string
	selectedItem int
	loadingItems bool
	services     []resources.ServiceInfo
	deployments  []resources.DeploymentInfo
	configMaps   []resources.ConfigMapInfo
	secrets      []resources.SecretInfo
}

// NewResourcesView creates a new resources view
func NewResourcesView() *ResourcesView {
	return &ResourcesView{
		resourceType: "services",
		selectedItem: 0,
		loadingItems: false,
		services:     []resources.ServiceInfo{},
		deployments:  []resources.DeploymentInfo{},
		configMaps:   []resources.ConfigMapInfo{},
		secrets:      []resources.SecretInfo{},
	}
}

// GetType returns the view type
func (v *ResourcesView) GetType() ViewType {
	return ViewTypeResources
}

// CanHandle returns true if this view can handle the given message
func (v *ResourcesView) CanHandle(msg tea.Msg) bool {
	switch msg.(type) {
	case ServicesLoaded, DeploymentsLoaded, ConfigMapsLoaded, SecretsLoaded:
		return true
	case ResourcesLoadError:
		return true
	case tea.KeyMsg:
		keyMsg := msg.(tea.KeyMsg)
		// Handle resource navigation keys
		switch keyMsg.String() {
		case "j", "k", "up", "down":
			return true
		}
	}
	return false
}

// Update handles messages for the resources view
func (v *ResourcesView) Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case ServicesLoaded:
		v.loadingItems = false
		v.services = msg.Services
		if v.resourceType == "services" {
			v.selectedItem = 0
		}

	case DeploymentsLoaded:
		v.loadingItems = false
		v.deployments = msg.Deployments
		if v.resourceType == "deployments" {
			v.selectedItem = 0
		}

	case ConfigMapsLoaded:
		v.loadingItems = false
		v.configMaps = msg.ConfigMaps
		if v.resourceType == "configmaps" {
			v.selectedItem = 0
		}

	case SecretsLoaded:
		v.loadingItems = false
		v.secrets = msg.Secrets
		if v.resourceType == "secrets" {
			v.selectedItem = 0
		}

	case ResourcesLoadError:
		v.loadingItems = false

	case tea.KeyMsg:
		// Handle resource navigation when focused on main panel
		if ctx.FocusedPanel == 0 {
			itemCount := v.getItemCount()
			if itemCount > 0 {
				switch msg.String() {
				case "j", "down":
					v.selectedItem = (v.selectedItem + 1) % itemCount
				case "k", "up":
					v.selectedItem = v.selectedItem - 1
					if v.selectedItem < 0 {
						v.selectedItem = itemCount - 1
					}
				}
			}
		}
	}

	return v, nil
}

// Render renders the resources view based on the active tab
func (v *ResourcesView) Render(ctx ViewContext) string {
	// Determine resource type from active tab
	tabName := ctx.App.GetTabName(ctx.App.ActiveTab)
	v.resourceType = strings.ToLower(tabName)

	if !ctx.Connected {
		return v.renderDisconnected(tabName)
	}

	if v.loadingItems {
		return v.renderLoading(tabName)
	}

	switch v.resourceType {
	case "services":
		return v.renderServices(ctx)
	case "deployments":
		return v.renderDeployments(ctx)
	case "configmaps":
		return v.renderConfigMaps(ctx)
	case "secrets":
		return v.renderSecrets(ctx)
	default:
		return v.renderComingSoon(tabName)
	}
}

// renderDisconnected renders the disconnected state
func (v *ResourcesView) renderDisconnected(resourceType string) string {
	return fmt.Sprintf(`📦 %s

❌ Not connected to any cluster

To connect to a cluster:
1. Run 'oc login <cluster-url>' in your terminal
2. Or start LazyOC with: lazyoc --kubeconfig /path/to/config

Press 'q' to quit`, resourceType)
}

// renderLoading renders the loading state
func (v *ResourcesView) renderLoading(resourceType string) string {
	return fmt.Sprintf("📦 %s\n\nLoading %s...", resourceType, strings.ToLower(resourceType))
}

// renderComingSoon renders placeholder for unimplemented resources
func (v *ResourcesView) renderComingSoon(resourceType string) string {
	return fmt.Sprintf("📦 %s Resources\n\n%s\n\nUse h/l or arrow keys to navigate tabs\nPress ? for help", resourceType, constants.ComingSoonMessage)
}

// renderServices renders the services list
func (v *ResourcesView) renderServices(ctx ViewContext) string {
	if len(v.services) == 0 {
		return fmt.Sprintf("📦 Services in %s\n\nNo services found in this namespace.", ctx.Namespace)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("📦 Services in %s\n\n", ctx.Namespace))

	// Table header
	content.WriteString("NAME                          TYPE        CLUSTER-IP      PORTS          AGE\n")
	content.WriteString("────                          ────        ──────────      ─────          ───\n")

	// Service rows
	for i, service := range v.services {
		// Highlight selected service
		prefix := "  "
		if i == v.selectedItem && ctx.FocusedPanel == 0 {
			prefix = "▶ "
		}

		// Truncate name if too long
		name := service.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		ports := strings.Join(service.Ports, ",")
		if len(ports) > 12 {
			ports = ports[:9] + "..."
		}

		content.WriteString(fmt.Sprintf("%s%-28s  %-10s  %-12s  %-12s  %s\n",
			prefix, name, service.Type, service.ClusterIP, ports, service.Age))
	}

	return content.String()
}

// renderDeployments renders the deployments list
func (v *ResourcesView) renderDeployments(ctx ViewContext) string {
	if len(v.deployments) == 0 {
		return fmt.Sprintf("📦 Deployments in %s\n\nNo deployments found in this namespace.", ctx.Namespace)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("📦 Deployments in %s\n\n", ctx.Namespace))

	// Table header
	content.WriteString("NAME                          READY    UP-TO-DATE   AVAILABLE   AGE\n")
	content.WriteString("────                          ─────    ──────────   ─────────   ───\n")

	// Deployment rows
	for i, deployment := range v.deployments {
		// Highlight selected deployment
		prefix := "  "
		if i == v.selectedItem && ctx.FocusedPanel == 0 {
			prefix = "▶ "
		}

		// Truncate name if too long
		name := deployment.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		ready := fmt.Sprintf("%d/%d", deployment.ReadyReplicas, deployment.Replicas)

		content.WriteString(fmt.Sprintf("%s%-28s  %-7s  %-10d   %-9d   %s\n",
			prefix, name, ready, deployment.UpdatedReplicas, deployment.AvailableReplicas, deployment.Age))
	}

	return content.String()
}

// renderConfigMaps renders the configmaps list
func (v *ResourcesView) renderConfigMaps(ctx ViewContext) string {
	if len(v.configMaps) == 0 {
		return fmt.Sprintf("📦 ConfigMaps in %s\n\nNo configmaps found in this namespace.", ctx.Namespace)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("📦 ConfigMaps in %s\n\n", ctx.Namespace))

	// Table header
	content.WriteString("NAME                          DATA   AGE\n")
	content.WriteString("────                          ────   ───\n")

	// ConfigMap rows
	for i, configMap := range v.configMaps {
		// Highlight selected configmap
		prefix := "  "
		if i == v.selectedItem && ctx.FocusedPanel == 0 {
			prefix = "▶ "
		}

		// Truncate name if too long
		name := configMap.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		content.WriteString(fmt.Sprintf("%s%-28s  %-5d  %s\n",
			prefix, name, configMap.DataCount, configMap.Age))
	}

	return content.String()
}

// renderSecrets renders the secrets list
func (v *ResourcesView) renderSecrets(ctx ViewContext) string {
	if len(v.secrets) == 0 {
		return fmt.Sprintf("📦 Secrets in %s\n\nNo secrets found in this namespace.", ctx.Namespace)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("📦 Secrets in %s\n\n", ctx.Namespace))

	// Table header
	content.WriteString("NAME                          TYPE                             DATA   AGE\n")
	content.WriteString("────                          ────                             ────   ───\n")

	// Secret rows
	for i, secret := range v.secrets {
		// Highlight selected secret
		prefix := "  "
		if i == v.selectedItem && ctx.FocusedPanel == 0 {
			prefix = "▶ "
		}

		// Truncate name if too long
		name := secret.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Truncate type if too long
		secretType := secret.Type
		if len(secretType) > 30 {
			secretType = secretType[:27] + "..."
		}

		content.WriteString(fmt.Sprintf("%s%-28s  %-30s   %-5d  %s\n",
			prefix, name, secretType, secret.DataCount, secret.Age))
	}

	return content.String()
}

// getItemCount returns the number of items for the current resource type
func (v *ResourcesView) getItemCount() int {
	switch v.resourceType {
	case "services":
		return len(v.services)
	case "deployments":
		return len(v.deployments)
	case "configmaps":
		return len(v.configMaps)
	case "secrets":
		return len(v.secrets)
	default:
		return 0
	}
}

// GetSelectedItem returns the currently selected item index
func (v *ResourcesView) GetSelectedItem() int {
	return v.selectedItem
}

// SetSelectedItem sets the selected item index
func (v *ResourcesView) SetSelectedItem(index int) {
	v.selectedItem = index
}

// GetResourceDetails returns formatted details for the selected resource
func (v *ResourcesView) GetResourceDetails(ctx ViewContext) string {
	switch v.resourceType {
	case "services":
		return v.getServiceDetails()
	case "deployments":
		return v.getDeploymentDetails()
	case "configmaps":
		return v.getConfigMapDetails()
	case "secrets":
		return v.getSecretDetails()
	default:
		return "No resource selected"
	}
}

// getServiceDetails returns details for the selected service
func (v *ResourcesView) getServiceDetails() string {
	if v.selectedItem >= len(v.services) || v.selectedItem < 0 {
		return "No service selected"
	}

	service := v.services[v.selectedItem]
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🌐 Service Details: %s\n\n", service.Name))

	details.WriteString(fmt.Sprintf("Namespace:     %s\n", service.Namespace))
	details.WriteString(fmt.Sprintf("Type:          %s\n", service.Type))
	details.WriteString(fmt.Sprintf("Cluster IP:    %s\n", service.ClusterIP))
	details.WriteString(fmt.Sprintf("Age:           %s\n", service.Age))

	if len(service.ExternalIPs) > 0 {
		details.WriteString(fmt.Sprintf("External IPs:  %s\n", strings.Join(service.ExternalIPs, ", ")))
	}

	if len(service.Ports) > 0 {
		details.WriteString("\nPorts:\n")
		for _, port := range service.Ports {
			details.WriteString(fmt.Sprintf("  %s\n", port))
		}
	}

	if service.Selector != "" {
		details.WriteString(fmt.Sprintf("\nSelector:      %s\n", service.Selector))
	}

	return details.String()
}

// getDeploymentDetails returns details for the selected deployment
func (v *ResourcesView) getDeploymentDetails() string {
	if v.selectedItem >= len(v.deployments) || v.selectedItem < 0 {
		return "No deployment selected"
	}

	deployment := v.deployments[v.selectedItem]
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🚀 Deployment Details: %s\n\n", deployment.Name))

	details.WriteString(fmt.Sprintf("Namespace:         %s\n", deployment.Namespace))
	details.WriteString(fmt.Sprintf("Replicas:          %d\n", deployment.Replicas))
	details.WriteString(fmt.Sprintf("Ready Replicas:    %d\n", deployment.ReadyReplicas))
	details.WriteString(fmt.Sprintf("Updated Replicas:  %d\n", deployment.UpdatedReplicas))
	details.WriteString(fmt.Sprintf("Available Replicas: %d\n", deployment.AvailableReplicas))
	details.WriteString(fmt.Sprintf("Strategy:          %s\n", deployment.Strategy))
	details.WriteString(fmt.Sprintf("Condition:         %s\n", deployment.Condition))
	details.WriteString(fmt.Sprintf("Age:               %s\n", deployment.Age))

	return details.String()
}

// getConfigMapDetails returns details for the selected configmap
func (v *ResourcesView) getConfigMapDetails() string {
	if v.selectedItem >= len(v.configMaps) || v.selectedItem < 0 {
		return "No configmap selected"
	}

	configMap := v.configMaps[v.selectedItem]
	var details strings.Builder
	details.WriteString(fmt.Sprintf("📋 ConfigMap Details: %s\n\n", configMap.Name))

	details.WriteString(fmt.Sprintf("Namespace:   %s\n", configMap.Namespace))
	details.WriteString(fmt.Sprintf("Data Count:  %d\n", configMap.DataCount))
	details.WriteString(fmt.Sprintf("Age:         %s\n", configMap.Age))

	return details.String()
}

// getSecretDetails returns details for the selected secret
func (v *ResourcesView) getSecretDetails() string {
	if v.selectedItem >= len(v.secrets) || v.selectedItem < 0 {
		return "No secret selected"
	}

	secret := v.secrets[v.selectedItem]
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🔐 Secret Details: %s\n\n", secret.Name))

	details.WriteString(fmt.Sprintf("Namespace:   %s\n", secret.Namespace))
	details.WriteString(fmt.Sprintf("Type:        %s\n", secret.Type))
	details.WriteString(fmt.Sprintf("Data Count:  %d\n", secret.DataCount))
	details.WriteString(fmt.Sprintf("Age:         %s\n", secret.Age))

	return details.String()
}

// Message types for resources loading
type ServicesLoaded struct {
	Services []resources.ServiceInfo
}

type DeploymentsLoaded struct {
	Deployments []resources.DeploymentInfo
}

type ConfigMapsLoaded struct {
	ConfigMaps []resources.ConfigMapInfo
}

type SecretsLoaded struct {
	Secrets []resources.SecretInfo
}

type ResourcesLoadError struct {
	Err          error
	ResourceType string
}
