package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/tui/components"
)

// ServicesView handles the services resource view
type ServicesView struct {
	panel    *components.PanelComponent
	services []ServiceItem
	style    ServiceStyle
}

// ServiceItem represents a service in the list
type ServiceItem struct {
	Name       string
	Namespace  string
	Type       string
	ClusterIP  string
	ExternalIP string
	Ports      string
	Age        string
	Selectors  map[string]string
}

// ServiceStyle contains styles for service rendering
type ServiceStyle struct {
	headerStyle       lipgloss.Style
	clusterIPStyle    lipgloss.Style
	nodePortStyle     lipgloss.Style
	loadBalancerStyle lipgloss.Style
	externalNameStyle lipgloss.Style
}

// NewServicesView creates a new services view
func NewServicesView() *ServicesView {
	view := &ServicesView{
		panel:    components.NewPanelComponent("Services"),
		services: make([]ServiceItem, 0),
		style: ServiceStyle{
			headerStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(constants.ColorCyan)),
			clusterIPStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorWhite)),
			nodePortStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorBlue)),
			loadBalancerStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorGreen)),
			externalNameStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorYellow)),
		},
	}

	view.panel.EnableSelection()
	return view
}

// Init initializes the services view
func (v *ServicesView) Init() tea.Cmd {
	return v.panel.Init()
}

// Update handles messages for the services view
func (v *ServicesView) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case ServicesLoadedMsg:
		v.SetServices(msg.Services)
		return nil, nil
	}

	return v.panel.Update(msg)
}

// View renders the services view
func (v *ServicesView) View() string {
	return v.panel.View()
}

// SetServices updates the service list
func (v *ServicesView) SetServices(services []ServiceItem) {
	v.services = services
	v.updateContent()
}

// GetSelectedService returns the currently selected service
func (v *ServicesView) GetSelectedService() *ServiceItem {
	idx := v.panel.GetSelectedIndex()
	if idx > 0 && idx <= len(v.services) { // Account for header row
		return &v.services[idx-1]
	}
	return nil
}

// updateContent updates the panel content with service list
func (v *ServicesView) updateContent() {
	if len(v.services) == 0 {
		v.panel.SetContent("No services found")
		return
	}

	// Create header
	header := fmt.Sprintf("%-30s %-12s %-15s %-15s %-20s %-5s",
		"NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORTS", "AGE")

	lines := []string{v.style.headerStyle.Render(header)}

	// Add services
	for _, svc := range v.services {
		// Choose style based on type
		var style lipgloss.Style
		switch strings.ToLower(svc.Type) {
		case "clusterip":
			style = v.style.clusterIPStyle
		case "nodeport":
			style = v.style.nodePortStyle
		case "loadbalancer":
			style = v.style.loadBalancerStyle
		case "externalname":
			style = v.style.externalNameStyle
		default:
			style = v.style.clusterIPStyle
		}

		externalIP := svc.ExternalIP
		if externalIP == "" {
			externalIP = "<none>"
		}

		line := fmt.Sprintf("%-30s %-12s %-15s %-15s %-20s %-5s",
			truncate(svc.Name, 30),
			truncate(svc.Type, 12),
			truncate(svc.ClusterIP, 15),
			truncate(externalIP, 15),
			truncate(svc.Ports, 20),
			svc.Age,
		)

		lines = append(lines, style.Render(line))
	}

	v.panel.SetContentLines(lines)
}

// Component interface implementation
func (v *ServicesView) Focus() error        { return v.panel.Focus() }
func (v *ServicesView) Blur() error         { return v.panel.Blur() }
func (v *ServicesView) IsFocused() bool     { return v.panel.IsFocused() }
func (v *ServicesView) SetSize(w, h int)    { v.panel.SetSize(w, h) }
func (v *ServicesView) GetSize() (int, int) { return v.panel.GetSize() }

// Messages
type ServicesLoadedMsg struct {
	Services []ServiceItem
}
