package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/tui/components"
	"github.com/katyella/lazyoc/internal/tui/state"
)

// App is the main TUI application
type App struct {
	// Core properties
	version string
	debug   bool

	// Components
	registry     *components.ComponentRegistry
	header       *components.HeaderComponent
	tabs         *components.TabsComponent
	statusBar    *components.StatusBarComponent
	mainPanel    *components.PanelComponent
	detailsPanel *components.PanelComponent
	logsPanel    *components.PanelComponent

	// Layout
	layout *components.LayoutManager

	// State
	stateManager *state.Manager

	// Terminal size
	width  int
	height int
}

// NewApp creates a new TUI application
func NewApp(version string, debug bool) *App {
	app := &App{
		version:      version,
		debug:        debug,
		registry:     components.NewComponentRegistry(),
		layout:       components.NewLayoutManager(),
		stateManager: state.NewManager(),
	}

	// Initialize components
	app.initializeComponents()

	return app
}

// initializeComponents creates and registers all components
func (a *App) initializeComponents() {
	// Create components
	a.header = components.NewHeaderComponent("LazyOC", a.version)
	a.tabs = components.NewTabsComponent()
	a.statusBar = components.NewStatusBarComponent()
	a.mainPanel = components.NewPanelComponent("Resources")
	a.detailsPanel = components.NewPanelComponent("Details")
	a.logsPanel = components.NewPanelComponent("Logs")

	// Enable selection for main panel
	a.mainPanel.EnableSelection()

	// Register components
	a.registry.Register("header", a.header)
	a.registry.Register("tabs", a.tabs)
	a.registry.Register("statusBar", a.statusBar)
	a.registry.Register("main", a.mainPanel)
	a.registry.Register("details", a.detailsPanel)
	a.registry.Register("logs", a.logsPanel)

	// Set initial focus
	_ = a.mainPanel.Focus()
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	// Initialize all components
	cmds := []tea.Cmd{
		a.registry.InitAll(),
		tea.WindowSize(), // Get initial window size
	}

	// Subscribe to state changes
	a.stateManager.Subscribe(a.handleStateChange)

	return tea.Batch(cmds...)
}

// Update handles all messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "?":
			// TODO: Show help modal
			return a, nil
		case "tab":
			// Switch to next tab
			cmd, _ := a.tabs.Update(components.TabChangeMsg{Direction: components.TabNext})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "shift+tab":
			// Switch to previous tab
			cmd, _ := a.tabs.Update(components.TabChangeMsg{Direction: components.TabPrev})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "ctrl+j":
			// Focus next panel
			a.focusNextPanel()
		case "ctrl+k":
			// Focus previous panel
			a.focusPreviousPanel()
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.layout.SetSize(msg.Width, msg.Height)
		a.updateComponentSizes()

	case components.TabChangedMsg:
		// Handle tab change
		a.stateManager.SetActiveTab(msg.Index, msg.Name)
		// TODO: Load appropriate resources
	}

	// Update all components with the message
	if cmd := a.registry.UpdateAll(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update state manager
	if cmd := a.stateManager.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Initializing..."
	}

	// Get layout dimensions
	dims := a.layout.GetDimensions()

	// Render components
	header := a.header.View()
	tabs := a.tabs.View()
	main := a.mainPanel.View()
	details := a.detailsPanel.View()
	logs := a.logsPanel.View()
	statusBar := a.statusBar.View()

	// Combine using layout
	return components.RenderLayout(dims, header, tabs, main, details, logs, statusBar)
}

// updateComponentSizes updates all component sizes based on layout
func (a *App) updateComponentSizes() {
	dims := a.layout.GetDimensions()

	// Update header and tabs (full width)
	a.header.SetSize(dims.Width, dims.HeaderHeight)
	a.tabs.SetSize(dims.Width, dims.TabBarHeight)
	a.statusBar.SetSize(dims.Width, dims.StatusBarHeight)

	// Update content panels
	a.mainPanel.SetSize(dims.MainWidth, dims.MainHeight)
	a.detailsPanel.SetSize(dims.DetailsWidth, dims.MainHeight)
	a.logsPanel.SetSize(dims.Width, dims.LogsHeight)
}

// focusNextPanel moves focus to the next panel
func (a *App) focusNextPanel() {
	panels := []string{"main", "details", "logs"}
	currentFocus := -1

	// Find current focus
	for i, name := range panels {
		if comp, ok := a.registry.Get(name); ok && comp.IsFocused() {
			currentFocus = i
			_ = comp.Blur()
			break
		}
	}

	// Focus next panel
	nextFocus := (currentFocus + 1) % len(panels)
	if comp, ok := a.registry.Get(panels[nextFocus]); ok {
		_ = comp.Focus()
	}
}

// focusPreviousPanel moves focus to the previous panel
func (a *App) focusPreviousPanel() {
	panels := []string{"main", "details", "logs"}
	currentFocus := -1

	// Find current focus
	for i, name := range panels {
		if comp, ok := a.registry.Get(name); ok && comp.IsFocused() {
			currentFocus = i
			_ = comp.Blur()
			break
		}
	}

	// Focus previous panel
	prevFocus := (currentFocus - 1 + len(panels)) % len(panels)
	if comp, ok := a.registry.Get(panels[prevFocus]); ok {
		_ = comp.Focus()
	}
}

// handleStateChange handles state change notifications
func (a *App) handleStateChange(change state.StateChange) {
	switch change.Type {
	case state.ChangeTypeConnection:
		// Update header with connection state
		if connState, ok := change.NewValue.(state.ConnectionState); ok {
			a.header.SetConnectionState(
				components.ConnectionState(connState.Status),
				components.ClusterInfo{
					Type:    connState.ClusterType,
					Version: connState.ClusterVersion,
					Context: connState.Context,
				},
				connState.Namespace,
			)
		}

	case state.ChangeTypeResources:
		// Update main panel with resources
		// TODO: Implement resource rendering

	case state.ChangeTypeSelection:
		// Update details panel
		// TODO: Implement detail rendering
	}
}

// Public methods for external control

// SetKubeconfig sets the kubeconfig path
func (a *App) SetKubeconfig(path string) error {
	return a.stateManager.SetKubeconfig(path)
}

// Connect initiates a connection to the cluster
func (a *App) Connect() tea.Cmd {
	// TODO: Implement connection logic
	return nil
}

// LoadResources loads resources for the current tab
func (a *App) LoadResources() tea.Cmd {
	// TODO: Implement resource loading
	return nil
}
