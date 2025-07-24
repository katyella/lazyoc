package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/tui/components"
)

// DeploymentsView handles the deployments resource view
type DeploymentsView struct {
	panel       *components.PanelComponent
	deployments []DeploymentItem
	style       DeploymentStyle
}

// DeploymentItem represents a deployment in the list
type DeploymentItem struct {
	Name        string
	Namespace   string
	Ready       string
	UpToDate    int
	Available   int
	Age         string
	Containers  []string
	Images      []string
}

// DeploymentStyle contains styles for deployment rendering
type DeploymentStyle struct {
	headerStyle   lipgloss.Style
	healthyStyle  lipgloss.Style
	scalingStyle  lipgloss.Style
	unhealthyStyle lipgloss.Style
}

// NewDeploymentsView creates a new deployments view
func NewDeploymentsView() *DeploymentsView {
	view := &DeploymentsView{
		panel:       components.NewPanelComponent("Deployments"),
		deployments: make([]DeploymentItem, 0),
		style: DeploymentStyle{
			headerStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(constants.ColorCyan)),
			healthyStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorGreen)),
			scalingStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorYellow)),
			unhealthyStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorRed)),
		},
	}
	
	view.panel.EnableSelection()
	return view
}

// Init initializes the deployments view
func (v *DeploymentsView) Init() tea.Cmd {
	return v.panel.Init()
}

// Update handles messages for the deployments view
func (v *DeploymentsView) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case DeploymentsLoadedMsg:
		v.SetDeployments(msg.Deployments)
		return nil, nil
	}
	
	return v.panel.Update(msg)
}

// View renders the deployments view
func (v *DeploymentsView) View() string {
	return v.panel.View()
}

// SetDeployments updates the deployment list
func (v *DeploymentsView) SetDeployments(deployments []DeploymentItem) {
	v.deployments = deployments
	v.updateContent()
}

// GetSelectedDeployment returns the currently selected deployment
func (v *DeploymentsView) GetSelectedDeployment() *DeploymentItem {
	idx := v.panel.GetSelectedIndex()
	if idx > 0 && idx <= len(v.deployments) { // Account for header row
		return &v.deployments[idx-1]
	}
	return nil
}

// updateContent updates the panel content with deployment list
func (v *DeploymentsView) updateContent() {
	if len(v.deployments) == 0 {
		v.panel.SetContent("No deployments found")
		return
	}
	
	// Create header
	header := fmt.Sprintf("%-40s %-7s %-9s %-10s %-5s",
		"NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")
	
	lines := []string{v.style.headerStyle.Render(header)}
	
	// Add deployments
	for _, dep := range v.deployments {
		// Choose style based on ready state
		var style lipgloss.Style
		ready := dep.Ready
		if ready == "" {
			ready = "0/0"
		}
		
		// Parse ready state (e.g., "2/2" means 2 ready out of 2 desired)
		var readyCount, desiredCount int
		fmt.Sscanf(ready, "%d/%d", &readyCount, &desiredCount)
		
		if readyCount == desiredCount && desiredCount > 0 {
			style = v.style.healthyStyle
		} else if readyCount > 0 {
			style = v.style.scalingStyle
		} else {
			style = v.style.unhealthyStyle
		}
		
		line := fmt.Sprintf("%-40s %-7s %-9d %-10d %-5s",
			truncate(dep.Name, 40),
			dep.Ready,
			dep.UpToDate,
			dep.Available,
			dep.Age,
		)
		
		lines = append(lines, style.Render(line))
	}
	
	v.panel.SetContentLines(lines)
}

// Component interface implementation
func (v *DeploymentsView) Focus() error     { return v.panel.Focus() }
func (v *DeploymentsView) Blur() error      { return v.panel.Blur() }
func (v *DeploymentsView) IsFocused() bool  { return v.panel.IsFocused() }
func (v *DeploymentsView) SetSize(w, h int) { v.panel.SetSize(w, h) }
func (v *DeploymentsView) GetSize() (int, int) { return v.panel.GetSize() }

// Messages
type DeploymentsLoadedMsg struct {
	Deployments []DeploymentItem
}