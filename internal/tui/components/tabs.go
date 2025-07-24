package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
)

// TabsComponent renders the resource tabs
type TabsComponent struct {
	BaseComponent

	// Tab state
	tabs         []string
	activeTab    int
	disabled     bool
	disableReason string

	// Styles
	activeTabStyle   lipgloss.Style
	inactiveTabStyle lipgloss.Style
	separatorStyle   lipgloss.Style
	disabledStyle    lipgloss.Style
}

// NewTabsComponent creates a new tabs component
func NewTabsComponent() *TabsComponent {
	return &TabsComponent{
		tabs:      constants.ResourceTabs,
		activeTab: 0,

		activeTabStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(constants.ColorWhite)).
			Background(lipgloss.Color(constants.ColorBlue)).
			Padding(0, 2),

		inactiveTabStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorGray)).
			Padding(0, 2),

		separatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorDarkGray)),

		disabledStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorGray)).
			Italic(true),
	}
}

// Init initializes the tabs component
func (t *TabsComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages for the tabs component
func (t *TabsComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case TabChangeMsg:
		if msg.Direction == TabNext {
			t.NextTab()
		} else {
			t.PrevTab()
		}
		return TabChangedCmd(t.activeTab, t.tabs[t.activeTab]), nil
	
	case TabSelectMsg:
		if msg.Index >= 0 && msg.Index < len(t.tabs) {
			t.activeTab = msg.Index
			return TabChangedCmd(t.activeTab, t.tabs[t.activeTab]), nil
		}
	
	case DisableTabsMsg:
		t.disabled = true
		t.disableReason = msg.Reason
	
	case EnableTabsMsg:
		t.disabled = false
		t.disableReason = ""
	}
	
	return nil, nil
}

// View renders the tabs component
func (t *TabsComponent) View() string {
	if t.width == 0 {
		return ""
	}

	// If disabled, show the reason
	if t.disabled {
		message := t.disableReason
		if message == "" {
			message = "Not connected"
		}
		centered := lipgloss.Place(
			t.width, 1,
			lipgloss.Center, lipgloss.Center,
			t.disabledStyle.Render(message),
		)
		separator := t.separatorStyle.Render(strings.Repeat("─", t.width))
		return lipgloss.JoinVertical(lipgloss.Left, centered, separator)
	}

	// Render tabs
	var renderedTabs []string
	for i, tab := range t.tabs {
		if i == t.activeTab {
			renderedTabs = append(renderedTabs, t.activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, t.inactiveTabStyle.Render(tab))
		}
	}

	tabLine := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	// Add padding
	tabLineWidth := lipgloss.Width(stripANSI(tabLine))
	if tabLineWidth < t.width {
		padding := strings.Repeat(" ", t.width-tabLineWidth)
		tabLine = tabLine + padding
	}

	// Add separator
	separator := t.separatorStyle.Render(strings.Repeat("─", t.width))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabLine,
		separator,
	)
}

// NextTab switches to the next tab
func (t *TabsComponent) NextTab() {
	if !t.disabled && len(t.tabs) > 0 {
		t.activeTab = (t.activeTab + 1) % len(t.tabs)
	}
}

// PrevTab switches to the previous tab
func (t *TabsComponent) PrevTab() {
	if !t.disabled && len(t.tabs) > 0 {
		t.activeTab = (t.activeTab - 1 + len(t.tabs)) % len(t.tabs)
	}
}

// GetActiveTab returns the currently active tab
func (t *TabsComponent) GetActiveTab() (int, string) {
	if t.activeTab >= 0 && t.activeTab < len(t.tabs) {
		return t.activeTab, t.tabs[t.activeTab]
	}
	return 0, ""
}

// SetActiveTab sets the active tab by index
func (t *TabsComponent) SetActiveTab(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.activeTab = index
	}
}

// Messages for tab interactions
type (
	// TabChangeMsg requests a tab change
	TabChangeMsg struct {
		Direction TabDirection
	}

	// TabSelectMsg selects a specific tab
	TabSelectMsg struct {
		Index int
	}

	// TabChangedMsg is sent when the active tab changes
	TabChangedMsg struct {
		Index int
		Name  string
	}

	// DisableTabsMsg disables tab interactions
	DisableTabsMsg struct {
		Reason string
	}

	// EnableTabsMsg enables tab interactions
	EnableTabsMsg struct{}
)

// TabDirection represents the direction of tab navigation
type TabDirection int

const (
	TabNext TabDirection = iota
	TabPrev
)

// TabChangedCmd creates a command that notifies of a tab change
func TabChangedCmd(index int, name string) tea.Cmd {
	return func() tea.Msg {
		return TabChangedMsg{Index: index, Name: name}
	}
}