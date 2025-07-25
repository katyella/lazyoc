package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/tui/components"
)

// DetailsView handles the details panel for selected resources
type DetailsView struct {
	panel        *components.PanelComponent
	resourceType string
	resourceName string
	content      map[string]interface{}
	style        DetailsStyle
}

// DetailsStyle contains styles for details rendering
type DetailsStyle struct {
	sectionStyle lipgloss.Style
	keyStyle     lipgloss.Style
	valueStyle   lipgloss.Style
	listStyle    lipgloss.Style
	errorStyle   lipgloss.Style
}

// NewDetailsView creates a new details view
func NewDetailsView() *DetailsView {
	return &DetailsView{
		panel:   components.NewPanelComponent("Details"),
		content: make(map[string]interface{}),
		style: DetailsStyle{
			sectionStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(constants.ColorCyan)).
				MarginTop(1),
			keyStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorBlue)),
			valueStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorWhite)),
			listStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorGray)).
				MarginLeft(2),
			errorStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorRed)),
		},
	}
}

// Init initializes the details view
func (v *DetailsView) Init() tea.Cmd {
	return v.panel.Init()
}

// Update handles messages for the details view
func (v *DetailsView) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case ResourceSelectedMsg:
		v.ShowResource(msg.Type, msg.Name, msg.Details)
		return nil, nil
	case ClearDetailsMsg:
		v.Clear()
		return nil, nil
	}

	return v.panel.Update(msg)
}

// View renders the details view
func (v *DetailsView) View() string {
	return v.panel.View()
}

// ShowResource displays details for a specific resource
func (v *DetailsView) ShowResource(resourceType, name string, details map[string]interface{}) {
	v.resourceType = resourceType
	v.resourceName = name
	v.content = details
	v.updateContent()
}

// Clear clears the details view
func (v *DetailsView) Clear() {
	v.resourceType = ""
	v.resourceName = ""
	v.content = make(map[string]interface{})
	v.panel.SetContent("Select a resource to view details")
}

// updateContent updates the panel content with resource details
func (v *DetailsView) updateContent() {
	if v.resourceType == "" || v.resourceName == "" {
		v.panel.SetContent("Select a resource to view details")
		return
	}

	var lines []string

	// Header
	header := fmt.Sprintf("%s: %s", v.resourceType, v.resourceName)
	lines = append(lines, v.style.sectionStyle.Render(header))
	lines = append(lines, "")

	// Process content sections
	if metadata, ok := v.content["metadata"].(map[string]interface{}); ok {
		lines = append(lines, v.renderSection("Metadata", metadata)...)
	}

	if spec, ok := v.content["spec"].(map[string]interface{}); ok {
		lines = append(lines, v.renderSection("Spec", spec)...)
	}

	if status, ok := v.content["status"].(map[string]interface{}); ok {
		lines = append(lines, v.renderSection("Status", status)...)
	}

	// Any other top-level sections
	for key, value := range v.content {
		if key != "metadata" && key != "spec" && key != "status" {
			if section, ok := value.(map[string]interface{}); ok {
				lines = append(lines, v.renderSection(strings.Title(key), section)...)
			}
		}
	}

	v.panel.SetContentLines(lines)
}

// renderSection renders a section of details
func (v *DetailsView) renderSection(title string, data map[string]interface{}) []string {
	var lines []string

	lines = append(lines, v.style.sectionStyle.Render(title))

	for key, value := range data {
		lines = append(lines, v.renderKeyValue(key, value)...)
	}

	lines = append(lines, "") // Empty line after section
	return lines
}

// renderKeyValue renders a key-value pair
func (v *DetailsView) renderKeyValue(key string, value interface{}) []string {
	var lines []string

	keyStr := v.style.keyStyle.Render(fmt.Sprintf("  %s:", key))

	switch val := value.(type) {
	case string:
		if val == "" {
			val = "<none>"
		}
		lines = append(lines, fmt.Sprintf("%s %s", keyStr, v.style.valueStyle.Render(val)))

	case []interface{}:
		if len(val) == 0 {
			lines = append(lines, fmt.Sprintf("%s %s", keyStr, v.style.valueStyle.Render("<none>")))
		} else {
			lines = append(lines, keyStr)
			for _, item := range val {
				lines = append(lines, v.style.listStyle.Render(fmt.Sprintf("- %v", item)))
			}
		}

	case map[string]interface{}:
		lines = append(lines, keyStr)
		for k, mapVal := range val {
			lines = append(lines, v.style.listStyle.Render(fmt.Sprintf("%s: %v", k, mapVal)))
		}

	default:
		lines = append(lines, fmt.Sprintf("%s %s", keyStr, v.style.valueStyle.Render(fmt.Sprintf("%v", value))))
	}

	return lines
}

// Component interface implementation
func (v *DetailsView) Focus() error        { return v.panel.Focus() }
func (v *DetailsView) Blur() error         { return v.panel.Blur() }
func (v *DetailsView) IsFocused() bool     { return v.panel.IsFocused() }
func (v *DetailsView) SetSize(w, h int)    { v.panel.SetSize(w, h) }
func (v *DetailsView) GetSize() (int, int) { return v.panel.GetSize() }

// Messages
type ResourceSelectedMsg struct {
	Type    string
	Name    string
	Details map[string]interface{}
}

type ClearDetailsMsg struct{}
