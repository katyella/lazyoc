package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/ui/messages"
)

// PodsView handles pod list display and navigation
type PodsView struct {
	loadingPods bool
	selectedPod int
}

// NewPodsView creates a new pods view
func NewPodsView() *PodsView {
	return &PodsView{
		loadingPods: false,
		selectedPod: 0,
	}
}

// GetType returns the view type
func (v *PodsView) GetType() ViewType {
	return ViewTypePods
}

// CanHandle returns true if this view can handle the given message
func (v *PodsView) CanHandle(msg tea.Msg) bool {
	switch msg.(type) {
	case messages.PodsLoaded, messages.LoadPodsError:
		return true
	case tea.KeyMsg:
		keyMsg := msg.(tea.KeyMsg)
		// Handle pod navigation keys
		switch keyMsg.String() {
		case "j", "k", "up", "down":
			return true
		}
	}
	return false
}

// Update handles messages for the pods view
func (v *PodsView) Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.PodsLoaded:
		// Store the previously selected pod name to preserve selection during refresh
		var previouslySelectedPodName string
		if len(ctx.Pods) > 0 && v.selectedPod < len(ctx.Pods) {
			previouslySelectedPodName = ctx.Pods[v.selectedPod].Name
		}

		v.loadingPods = false

		// Try to preserve the selected pod after refresh
		newSelectedPod := 0
		if previouslySelectedPodName != "" {
			for i, pod := range msg.Pods {
				if pod.Name == previouslySelectedPodName {
					newSelectedPod = i
					break
				}
			}
		}
		v.selectedPod = newSelectedPod

	case messages.LoadPodsError:
		v.loadingPods = false

	case tea.KeyMsg:
		// Handle pod navigation when focused on main panel
		if ctx.FocusedPanel == 0 && len(ctx.Pods) > 0 {
			switch msg.String() {
			case "j", "down":
				v.selectedPod = (v.selectedPod + 1) % len(ctx.Pods)
				return v, v.loadPodLogs(ctx)
			case "k", "up":
				v.selectedPod = v.selectedPod - 1
				if v.selectedPod < 0 {
					v.selectedPod = len(ctx.Pods) - 1
				}
				return v, v.loadPodLogs(ctx)
			}
		}
	}

	return v, nil
}

// Render renders the pods view
func (v *PodsView) Render(ctx ViewContext) string {
	if !ctx.Connected {
		return v.renderDisconnected()
	}

	if v.loadingPods {
		return constants.LoadingPodsMessage
	}

	if len(ctx.Pods) == 0 {
		return v.renderNoPods(ctx)
	}

	return v.renderPodList(ctx)
}

// renderDisconnected renders the disconnected state
func (v *PodsView) renderDisconnected() string {
	return `📦 Pods

❌ Not connected to any cluster

To connect to a cluster:
1. Run 'oc login <cluster-url>' in your terminal
2. Or start LazyOC with: lazyoc --kubeconfig /path/to/config

Press 'q' to quit`
}

// renderNoPods renders when no pods are found
func (v *PodsView) renderNoPods(ctx ViewContext) string {
	if ctx.Namespace != "" {
		return fmt.Sprintf("📦 Pods in %s\n\nNo pods found in this namespace.", ctx.Namespace)
	}
	return "📦 Pods\n\nNo pods found in this namespace."
}

// renderPodList renders the pod list
func (v *PodsView) renderPodList(ctx ViewContext) string {
	var content strings.Builder

	// Header with namespace
	if ctx.Namespace != "" {
		content.WriteString(fmt.Sprintf("📦 Pods in %s\n\n", ctx.Namespace))
	} else {
		content.WriteString("📦 Pods\n\n")
	}

	// Table header
	content.WriteString("NAME                                    STATUS    READY   AGE\n")
	content.WriteString("────────────────────────────────────    ──────    ─────   ───\n")

	// Pod rows
	for i, pod := range ctx.Pods {
		// Highlight selected pod
		prefix := "  "
		if i == v.selectedPod && ctx.FocusedPanel == 0 {
			prefix = "▶ "
		}

		// Truncate name if too long
		name := pod.Name
		if len(name) > constants.PodNameTruncateLength {
			name = name[:constants.PodNameTruncateLengthCompact] + "..."
		}

		// Add status indicator with emoji
		statusIndicator := v.getPodStatusIndicator(pod.Phase)

		content.WriteString(fmt.Sprintf("%s%-38s  %s%-7s  %-5s   %s\n",
			prefix, name, statusIndicator, pod.Phase, pod.Ready, pod.Age))
	}

	return content.String()
}

// getPodStatusIndicator returns an emoji indicator for pod status
func (v *PodsView) getPodStatusIndicator(phase string) string {
	switch phase {
	case "Running":
		return "✅"
	case "Pending":
		return "⏳"
	case "Failed":
		return "❌"
	case "Succeeded":
		return "✨"
	case "Unknown":
		return "❓"
	default:
		return "⚪"
	}
}

// GetSelectedPod returns the currently selected pod index
func (v *PodsView) GetSelectedPod() int {
	return v.selectedPod
}

// SetSelectedPod sets the selected pod index
func (v *PodsView) SetSelectedPod(index int) {
	v.selectedPod = index
}

// GetPodDetails returns formatted details for the selected pod
func (v *PodsView) GetPodDetails(ctx ViewContext) string {
	if v.selectedPod >= len(ctx.Pods) || v.selectedPod < 0 {
		return "No pod selected"
	}

	pod := ctx.Pods[v.selectedPod]
	var details strings.Builder
	details.WriteString(fmt.Sprintf("📄 Pod Details: %s\n\n", pod.Name))

	details.WriteString(fmt.Sprintf("Namespace:  %s\n", pod.Namespace))
	details.WriteString(fmt.Sprintf("Status:     %s\n", pod.Phase))
	details.WriteString(fmt.Sprintf("Ready:      %s\n", pod.Ready))
	details.WriteString(fmt.Sprintf("Restarts:   %d\n", pod.Restarts))
	details.WriteString(fmt.Sprintf("Age:        %s\n", pod.Age))
	details.WriteString(fmt.Sprintf("Node:       %s\n", pod.Node))
	details.WriteString(fmt.Sprintf("IP:         %s\n", pod.IP))

	if len(pod.ContainerInfo) > 0 {
		details.WriteString("\nContainers:\n")
		for _, container := range pod.ContainerInfo {
			status := "🟢"
			if !container.Ready {
				status = "🔴"
			}
			details.WriteString(fmt.Sprintf("  %s %s (%s)\n", status, container.Name, container.State))
		}
	}

	return details.String()
}

// loadPodLogs returns a command to load logs for the selected pod
func (v *PodsView) loadPodLogs(ctx ViewContext) tea.Cmd {
	if v.selectedPod >= len(ctx.Pods) || v.selectedPod < 0 {
		return nil
	}

	selectedPod := ctx.Pods[v.selectedPod]
	return func() tea.Msg {
		// This would typically interface with the resource client to load logs
		// For now, return a message that indicates logs should be loaded
		return messages.LoadPodLogsMsg{
			PodName:   selectedPod.Name,
			Namespace: selectedPod.Namespace,
		}
	}
}