package components

import (
	"github.com/charmbracelet/lipgloss"
)

// TabComponent represents a tabbed navigation component
type TabComponent struct {
	Width       int
	Height      int
	Tabs        []Tab
	ActiveIndex int

	// Styling options
	ActiveStyle    lipgloss.Style
	InactiveStyle  lipgloss.Style
	SeparatorStyle lipgloss.Style
	ContainerStyle lipgloss.Style

	// Layout options
	Alignment      lipgloss.Position
	ShowSeparators bool
	TabPadding     int
	TabSpacing     int
}

// Tab represents a single tab
type Tab struct {
	ID         string
	Label      string
	Icon       string
	Enabled    bool
	Badge      string
	BadgeColor lipgloss.Color
	Tooltip    string
}

// NewTabComponent creates a new tab component
func NewTabComponent(width, height int) *TabComponent {
	return &TabComponent{
		Width:       width,
		Height:      height,
		Tabs:        make([]Tab, 0),
		ActiveIndex: 0,

		ActiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("12")). // Blue
			Bold(true).
			Padding(0, 1),

		InactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Padding(0, 1),

		SeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			SetString(" | "),

		ContainerStyle: lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center),

		Alignment:      lipgloss.Center,
		ShowSeparators: true,
		TabPadding:     1,
		TabSpacing:     0,
	}
}

// AddTab adds a new tab to the component
func (tc *TabComponent) AddTab(tab Tab) {
	tc.Tabs = append(tc.Tabs, tab)
}

// SetActiveTab sets the active tab by index
func (tc *TabComponent) SetActiveTab(index int) bool {
	if index >= 0 && index < len(tc.Tabs) && tc.Tabs[index].Enabled {
		tc.ActiveIndex = index
		return true
	}
	return false
}

// SetActiveTabByID sets the active tab by ID
func (tc *TabComponent) SetActiveTabByID(id string) bool {
	for i, tab := range tc.Tabs {
		if tab.ID == id && tab.Enabled {
			tc.ActiveIndex = i
			return true
		}
	}
	return false
}

// GetActiveTab returns the currently active tab
func (tc *TabComponent) GetActiveTab() *Tab {
	if tc.ActiveIndex >= 0 && tc.ActiveIndex < len(tc.Tabs) {
		return &tc.Tabs[tc.ActiveIndex]
	}
	return nil
}

// NextTab moves to the next enabled tab
func (tc *TabComponent) NextTab() bool {
	if len(tc.Tabs) == 0 {
		return false
	}

	start := tc.ActiveIndex
	for i := 0; i < len(tc.Tabs); i++ {
		next := (tc.ActiveIndex + 1 + i) % len(tc.Tabs)
		if tc.Tabs[next].Enabled {
			tc.ActiveIndex = next
			return tc.ActiveIndex != start
		}
	}
	return false
}

// PrevTab moves to the previous enabled tab
func (tc *TabComponent) PrevTab() bool {
	if len(tc.Tabs) == 0 {
		return false
	}

	start := tc.ActiveIndex
	for i := 0; i < len(tc.Tabs); i++ {
		prev := (tc.ActiveIndex - 1 - i + len(tc.Tabs)) % len(tc.Tabs)
		if tc.Tabs[prev].Enabled {
			tc.ActiveIndex = prev
			return tc.ActiveIndex != start
		}
	}
	return false
}

// SetDimensions updates the tab component dimensions
func (tc *TabComponent) SetDimensions(width, height int) {
	tc.Width = width
	tc.Height = height
	tc.ContainerStyle = tc.ContainerStyle.Width(width).Height(height)
}

// Render renders the tab component
func (tc *TabComponent) Render() string {
	if len(tc.Tabs) == 0 {
		return tc.ContainerStyle.Render("No tabs available")
	}

	renderedTabs := make([]string, 0, len(tc.Tabs))

	for i, tab := range tc.Tabs {
		if !tab.Enabled {
			continue
		}

		// Build tab content
		content := ""
		if tab.Icon != "" {
			content += tab.Icon + " "
		}
		content += tab.Label
		if tab.Badge != "" {
			badgeStyle := lipgloss.NewStyle()
			if tab.BadgeColor != "" {
				badgeStyle = badgeStyle.Foreground(tab.BadgeColor)
			}
			content += " " + badgeStyle.Render(tab.Badge)
		}

		// Apply appropriate style
		var tabStyle lipgloss.Style
		if i == tc.ActiveIndex {
			tabStyle = tc.ActiveStyle
		} else {
			tabStyle = tc.InactiveStyle
		}

		// Apply additional padding if specified
		if tc.TabPadding > 0 {
			tabStyle = tabStyle.Padding(0, tc.TabPadding)
		}

		renderedTab := tabStyle.Render(content)
		renderedTabs = append(renderedTabs, renderedTab)
	}

	// Join tabs with separators if enabled
	var tabBar string
	if tc.ShowSeparators && len(renderedTabs) > 1 {
		separator := tc.SeparatorStyle.String()
		tabBar = renderedTabs[0]
		for i := 1; i < len(renderedTabs); i++ {
			tabBar = lipgloss.JoinHorizontal(lipgloss.Top, tabBar, separator, renderedTabs[i])
		}
	} else {
		// Join with spacing
		if tc.TabSpacing > 0 {
			spacer := lipgloss.NewStyle().Width(tc.TabSpacing).Render("")
			tabBar = renderedTabs[0]
			for i := 1; i < len(renderedTabs); i++ {
				tabBar = lipgloss.JoinHorizontal(lipgloss.Top, tabBar, spacer, renderedTabs[i])
			}
		} else {
			tabBar = lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
		}
	}

	// Apply container styling with alignment
	return tc.ContainerStyle.Render(tabBar)
}

// SetAlignment sets the tab alignment
func (tc *TabComponent) SetAlignment(align lipgloss.Position) {
	tc.Alignment = align
	tc.ContainerStyle = tc.ContainerStyle.Align(align, lipgloss.Center)
}

// EnableTab enables or disables a tab by index
func (tc *TabComponent) EnableTab(index int, enabled bool) bool {
	if index >= 0 && index < len(tc.Tabs) {
		tc.Tabs[index].Enabled = enabled

		// If we disabled the active tab, move to the next enabled tab
		if !enabled && index == tc.ActiveIndex {
			return tc.NextTab()
		}
		return true
	}
	return false
}

// EnableTabByID enables or disables a tab by ID
func (tc *TabComponent) EnableTabByID(id string, enabled bool) bool {
	for i, tab := range tc.Tabs {
		if tab.ID == id {
			return tc.EnableTab(i, enabled)
		}
	}
	return false
}

// UpdateTabBadge updates the badge for a tab
func (tc *TabComponent) UpdateTabBadge(index int, badge string, color lipgloss.Color) bool {
	if index >= 0 && index < len(tc.Tabs) {
		tc.Tabs[index].Badge = badge
		tc.Tabs[index].BadgeColor = color
		return true
	}
	return false
}

// UpdateTabBadgeByID updates the badge for a tab by ID
func (tc *TabComponent) UpdateTabBadgeByID(id string, badge string, color lipgloss.Color) bool {
	for i, tab := range tc.Tabs {
		if tab.ID == id {
			return tc.UpdateTabBadge(i, badge, color)
		}
	}
	return false
}

// ClearBadges clears all tab badges
func (tc *TabComponent) ClearBadges() {
	for i := range tc.Tabs {
		tc.Tabs[i].Badge = ""
	}
}

// GetTabCount returns the total number of tabs
func (tc *TabComponent) GetTabCount() int {
	return len(tc.Tabs)
}

// GetEnabledTabCount returns the number of enabled tabs
func (tc *TabComponent) GetEnabledTabCount() int {
	count := 0
	for _, tab := range tc.Tabs {
		if tab.Enabled {
			count++
		}
	}
	return count
}

// CreateKubernetesTabComponent creates a pre-configured tab component for Kubernetes resources
func CreateKubernetesTabComponent(width, height int) *TabComponent {
	tc := NewTabComponent(width, height)

	// Add Kubernetes resource tabs
	tc.AddTab(Tab{
		ID:      "pods",
		Label:   "Pods",
		Icon:    "📦",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "services",
		Label:   "Services",
		Icon:    "🔗",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "deployments",
		Label:   "Deployments",
		Icon:    "🚀",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "configmaps",
		Label:   "ConfigMaps",
		Icon:    "⚙️",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "secrets",
		Label:   "Secrets",
		Icon:    "🔐",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "ingress",
		Label:   "Ingress",
		Icon:    "🌐",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "volumes",
		Label:   "Volumes",
		Icon:    "💾",
		Enabled: true,
	})

	tc.AddTab(Tab{
		ID:      "nodes",
		Label:   "Nodes",
		Icon:    "🖥️",
		Enabled: true,
	})

	return tc
}

// CreateSimpleTabComponent creates a simple tab component with basic styling
func CreateSimpleTabComponent(width, height int, tabNames []string) *TabComponent {
	tc := NewTabComponent(width, height)

	// Simplified styling
	tc.ActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 2)

	tc.InactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 2)

	tc.ShowSeparators = false
	tc.TabSpacing = 1

	// Add tabs from names
	for _, name := range tabNames {
		tc.AddTab(Tab{
			ID:      name,
			Label:   name,
			Enabled: true,
		})
	}

	return tc
}
