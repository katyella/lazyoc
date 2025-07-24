package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/tui/components"
)

// PodsView handles the pods resource view
type PodsView struct {
	panel *components.PanelComponent
	pods  []PodItem
	style PodStyle
}

// PodItem represents a pod in the list
type PodItem struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
	Restarts  int
	Age       string
	Node      string
}

// PodStyle contains styles for pod rendering
type PodStyle struct {
	headerStyle  lipgloss.Style
	runningStyle lipgloss.Style
	pendingStyle lipgloss.Style
	failedStyle  lipgloss.Style
	unknownStyle lipgloss.Style
}

// NewPodsView creates a new pods view
func NewPodsView() *PodsView {
	view := &PodsView{
		panel: components.NewPanelComponent("Pods"),
		pods:  make([]PodItem, 0),
		style: PodStyle{
			headerStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(constants.ColorCyan)),
			runningStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorGreen)),
			pendingStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorYellow)),
			failedStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorRed)),
			unknownStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorGray)),
		},
	}
	
	view.panel.EnableSelection()
	return view
}

// Init initializes the pods view
func (v *PodsView) Init() tea.Cmd {
	return v.panel.Init()
}

// Update handles messages for the pods view
func (v *PodsView) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case PodsLoadedMsg:
		v.SetPods(msg.Pods)
		return nil, nil
	}
	
	return v.panel.Update(msg)
}

// View renders the pods view
func (v *PodsView) View() string {
	return v.panel.View()
}

// SetPods updates the pod list
func (v *PodsView) SetPods(pods []PodItem) {
	v.pods = pods
	v.updateContent()
}

// GetSelectedPod returns the currently selected pod
func (v *PodsView) GetSelectedPod() *PodItem {
	idx := v.panel.GetSelectedIndex()
	if idx > 0 && idx <= len(v.pods) { // Account for header row
		return &v.pods[idx-1]
	}
	return nil
}

// updateContent updates the panel content with pod list
func (v *PodsView) updateContent() {
	if len(v.pods) == 0 {
		v.panel.SetContent("No pods found")
		return
	}
	
	// Create header
	header := fmt.Sprintf("%-40s %-12s %-7s %-8s %-5s %-20s",
		"NAME", "STATUS", "READY", "RESTARTS", "AGE", "NODE")
	
	lines := []string{v.style.headerStyle.Render(header)}
	
	// Add pods
	for _, pod := range v.pods {
		// Choose style based on status
		var style lipgloss.Style
		switch strings.ToLower(pod.Status) {
		case "running":
			style = v.style.runningStyle
		case "pending", "containercreating", "podinitiating":
			style = v.style.pendingStyle
		case "failed", "error", "crashloopbackoff", "imagepullbackoff":
			style = v.style.failedStyle
		default:
			style = v.style.unknownStyle
		}
		
		line := fmt.Sprintf("%-40s %-12s %-7s %-8d %-5s %-20s",
			truncate(pod.Name, 40),
			truncate(pod.Status, 12),
			pod.Ready,
			pod.Restarts,
			pod.Age,
			truncate(pod.Node, 20),
		)
		
		lines = append(lines, style.Render(line))
	}
	
	v.panel.SetContentLines(lines)
}

// Component interface implementation
func (v *PodsView) Focus() error     { return v.panel.Focus() }
func (v *PodsView) Blur() error      { return v.panel.Blur() }
func (v *PodsView) IsFocused() bool  { return v.panel.IsFocused() }
func (v *PodsView) SetSize(w, h int) { v.panel.SetSize(w, h) }
func (v *PodsView) GetSize() (int, int) { return v.panel.GetSize() }

// Messages
type PodsLoadedMsg struct {
	Pods []PodItem
}

// Helper functions
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}