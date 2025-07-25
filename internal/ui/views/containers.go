package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// ContainersView handles container list display and management
type ContainersView struct {
	selectedContainer int
	podIndex          int
}

// NewContainersView creates a new containers view
func NewContainersView() *ContainersView {
	return &ContainersView{
		selectedContainer: 0,
		podIndex:          0,
	}
}

// GetType returns the view type
func (v *ContainersView) GetType() ViewType {
	return ViewTypeContainers
}

// CanHandle returns true if this view can handle the given message
func (v *ContainersView) CanHandle(msg tea.Msg) bool {
	switch msg.(type) {
	case tea.KeyMsg:
		keyMsg := msg.(tea.KeyMsg)
		// Handle container navigation keys
		switch keyMsg.String() {
		case "j", "k", "up", "down":
			return true
		}
	}
	return false
}

// Update handles messages for the containers view
func (v *ContainersView) Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle container navigation when focused on main panel
		if ctx.FocusedPanel == 0 {
			pod := v.getCurrentPod(ctx)
			if pod != nil && len(pod.ContainerInfo) > 0 {
				switch msg.String() {
				case "j", "down":
					v.selectedContainer = (v.selectedContainer + 1) % len(pod.ContainerInfo)
				case "k", "up":
					v.selectedContainer = v.selectedContainer - 1
					if v.selectedContainer < 0 {
						v.selectedContainer = len(pod.ContainerInfo) - 1
					}
				}
			}
		}
	}

	return v, nil
}

// Render renders the containers view
func (v *ContainersView) Render(ctx ViewContext) string {
	if !ctx.Connected {
		return v.renderDisconnected()
	}

	pod := v.getCurrentPod(ctx)
	if pod == nil {
		return v.renderNoPod()
	}

	if len(pod.ContainerInfo) == 0 {
		return v.renderNoContainers(pod)
	}

	return v.renderContainerList(ctx, pod)
}

// renderDisconnected renders the disconnected state
func (v *ContainersView) renderDisconnected() string {
	return `🐳 Containers

❌ Not connected to any cluster

To connect to a cluster:
1. Run 'oc login <cluster-url>' in your terminal
2. Or start LazyOC with: lazyoc --kubeconfig /path/to/config

Press 'q' to quit`
}

// renderNoPod renders when no pod is selected
func (v *ContainersView) renderNoPod() string {
	return `🐳 Containers

No pod selected

Select a pod from the Pods tab to view its containers.`
}

// renderNoContainers renders when pod has no containers
func (v *ContainersView) renderNoContainers(pod *resources.PodInfo) string {
	return fmt.Sprintf(`🐳 Containers in %s

No containers found in this pod.`, pod.Name)
}

// renderContainerList renders the container list
func (v *ContainersView) renderContainerList(ctx ViewContext, pod *resources.PodInfo) string {
	var content strings.Builder

	// Header with pod name
	content.WriteString(fmt.Sprintf("🐳 Containers in %s\n\n", pod.Name))

	// Table header
	content.WriteString("NAME                          STATE      READY    RESTARTS\n")
	content.WriteString("────────────────────────────  ─────────  ─────    ────────\n")

	// Container rows
	for i, container := range pod.ContainerInfo {
		// Highlight selected container
		prefix := "  "
		if i == v.selectedContainer && ctx.FocusedPanel == 0 {
			prefix = "▶ "
		}

		// Truncate name if too long
		name := container.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Add status indicator with emoji
		statusIndicator := v.getContainerStatusIndicator(container.Ready, container.State)

		content.WriteString(fmt.Sprintf("%s%-28s  %s%-8s  %-7v  %d\n",
			prefix, name, statusIndicator, container.State, container.Ready, container.RestartCount))
	}

	return content.String()
}

// getContainerStatusIndicator returns an emoji indicator for container status
func (v *ContainersView) getContainerStatusIndicator(ready bool, state string) string {
	if ready {
		return "✅"
	}

	switch state {
	case "Running":
		return "🟡" // Running but not ready
	case "Waiting":
		return "⏳"
	case "Terminated":
		return "❌"
	case "CrashLoopBackOff":
		return "🔄"
	default:
		return "❓"
	}
}

// getCurrentPod returns the currently selected pod
func (v *ContainersView) getCurrentPod(ctx ViewContext) *resources.PodInfo {
	if ctx.SelectedPod >= len(ctx.Pods) || ctx.SelectedPod < 0 {
		return nil
	}
	return &ctx.Pods[ctx.SelectedPod]
}

// GetSelectedContainer returns the currently selected container index
func (v *ContainersView) GetSelectedContainer() int {
	return v.selectedContainer
}

// SetSelectedContainer sets the selected container index
func (v *ContainersView) SetSelectedContainer(index int) {
	v.selectedContainer = index
}

// GetContainerDetails returns formatted details for the selected container
func (v *ContainersView) GetContainerDetails(ctx ViewContext) string {
	pod := v.getCurrentPod(ctx)
	if pod == nil || v.selectedContainer >= len(pod.ContainerInfo) || v.selectedContainer < 0 {
		return "No container selected"
	}

	container := pod.ContainerInfo[v.selectedContainer]
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🐳 Container Details: %s\n\n", container.Name))

	details.WriteString(fmt.Sprintf("Pod:         %s\n", pod.Name))
	details.WriteString(fmt.Sprintf("State:       %s\n", container.State))
	details.WriteString(fmt.Sprintf("Ready:       %t\n", container.Ready))
	details.WriteString(fmt.Sprintf("Restarts:    %d\n", container.RestartCount))

	if container.Image != "" {
		details.WriteString(fmt.Sprintf("Image:       %s\n", container.Image))
	}

	// Add port information if available
	if len(container.Ports) > 0 {
		details.WriteString("\nPorts:\n")
		for _, port := range container.Ports {
			protocol := port.Protocol
			if protocol == "" {
				protocol = "TCP"
			}
			details.WriteString(fmt.Sprintf("  %d/%s", port.ContainerPort, protocol))
			if port.Name != "" {
				details.WriteString(fmt.Sprintf(" (%s)", port.Name))
			}
			details.WriteString("\n")
		}
	}

	// Add environment variables if available
	if len(container.Env) > 0 {
		details.WriteString("\nEnvironment Variables:\n")
		for _, env := range container.Env {
			if env.ValueFrom != nil {
				details.WriteString(fmt.Sprintf("  %s: <from %s>\n", env.Name, v.getEnvSourceDescription(env.ValueFrom)))
			} else {
				// Truncate long values
				value := env.Value
				if len(value) > 50 {
					value = value[:47] + "..."
				}
				details.WriteString(fmt.Sprintf("  %s: %s\n", env.Name, value))
			}
		}
	}

	return details.String()
}

// getEnvSourceDescription returns a description of the environment variable source
func (v *ContainersView) getEnvSourceDescription(source *resources.EnvVarSource) string {
	if source.ConfigMapKeyRef != nil {
		return fmt.Sprintf("ConfigMap %s.%s", source.ConfigMapKeyRef.Name, source.ConfigMapKeyRef.Key)
	}
	if source.SecretKeyRef != nil {
		return fmt.Sprintf("Secret %s.%s", source.SecretKeyRef.Name, source.SecretKeyRef.Key)
	}
	if source.FieldRef != nil {
		return fmt.Sprintf("Field %s", source.FieldRef.FieldPath)
	}
	if source.ResourceFieldRef != nil {
		return fmt.Sprintf("Resource %s", source.ResourceFieldRef.Resource)
	}
	return "Unknown source"
}
