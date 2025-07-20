package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/katyella/lazyoc/internal/errors"
	"github.com/katyella/lazyoc/internal/logging"
	"github.com/katyella/lazyoc/internal/ui/components"
	"github.com/katyella/lazyoc/internal/ui/messages"
	"github.com/katyella/lazyoc/internal/ui/models"
	"github.com/katyella/lazyoc/internal/ui/navigation"
)

// TUI wraps the App model and implements the tea.Model interface
type TUI struct {
	*models.App
	
	// Navigation system
	navController *navigation.NavigationController
	helpComponent *navigation.HelpComponent
	
	// Layout components
	layoutManager *components.LayoutManager
	header        *components.HeaderComponent
	tabs          *components.TabComponent
	contentPane   *components.ContentPane
	detailPane    *components.DetailPane
	logPane       *components.LogPane
	statusBar     *components.StatusBarComponent
	
	// CRITICAL: Ready state pattern from research
	isReady               bool // Never render until first WindowSizeMsg
	componentsInitialized bool
}

// NewTUI creates a new TUI instance
func NewTUI(version string, debug bool) *TUI {
	app := models.NewApp(version)
	app.Debug = debug
	app.Logger = logging.SetupLogger(debug)
	
	logging.Info(app.Logger, "Initializing LazyOC TUI v%s", version)
	
	// Create TUI with empty layout components (will be initialized on first resize)
	tui := &TUI{
		App:           app,
		navController: navigation.NewNavigationController(),
		helpComponent: navigation.NewHelpComponent(80, 24), // Default size, will be updated
	}
	
	// Set up navigation callbacks
	tui.setupNavigationCallbacks()
	
	return tui
}

// Init implements tea.Model interface - called once at startup
func (t *TUI) Init() tea.Cmd {
	logging.Debug(t.Logger, "TUI Init() called")
	
	// Return a batch of initial commands
	return tea.Batch(
		// Get initial terminal size
		tea.WindowSize(),
		// Initial setup command
		func() tea.Msg {
			return messages.InitMsg{}
		},
	)
}

// Update implements tea.Model interface - handles messages and updates state
func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logging.Debug(t.Logger, "TUI Update() received message: %T", msg)
	
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	
	// Terminal window size changes
	case tea.WindowSizeMsg:
		// CRITICAL: This is our first valid sizing info - mark as ready
		if !t.isReady {
			t.isReady = true
			logging.Debug(t.Logger, "TUI marked ready after receiving WindowSizeMsg")
		}
		
		// Ensure minimum terminal size to prevent layout issues
		width, height := msg.Width, msg.Height
		minWidth, minHeight := 60, 15 // Absolute minimum for basic functionality
		
		if width < minWidth {
			width = minWidth
		}
		if height < minHeight {
			height = minHeight
		}
		
		t.SetDimensions(width, height)
		
		// Initialize components ONLY after we're ready with valid dimensions
		if !t.componentsInitialized && t.isReady {
			t.testLayoutManagerOnly(width, height)
			t.componentsInitialized = true
		} else if t.componentsInitialized {
			// Update existing layout with new dimensions
			if t.layoutManager != nil {
				t.layoutManager.UpdateDimensions(width, height)
				t.updateAllComponentDimensions()
			}
		}
		
		// Update help component dimensions
		if t.helpComponent != nil {
			t.helpComponent.SetDimensions(width, height)
		}
		
		logging.Debug(t.Logger, "Window resized to %dx%d (requested: %dx%d)", width, height, msg.Width, msg.Height)
		
	// Keyboard input
	case tea.KeyMsg:
		return t.handleKeyInputWithNavigation(msg)
		
	// Application initialization
	case messages.InitMsg:
		t.ClearLoading()
		t.State = models.StateMain
		// Initialize layout with default size if no WindowSizeMsg received yet
		if t.Width == 0 || t.Height == 0 && !t.componentsInitialized {
			t.SetDimensions(80, 24) // Default terminal size
			t.testLayoutManagerOnly(80, 24)
			t.componentsInitialized = true
		}
		logging.Info(t.Logger, "Application initialized successfully")
		
	// Error messages
	case messages.ErrorMsg:
		t.SetError(msg.Err)
		logging.Error(t.Logger, "Error occurred: %v", msg.Err)
		
	// Loading state changes
	case messages.LoadingMsg:
		if msg.Loading {
			t.SetLoading(msg.Message)
		} else {
			t.ClearLoading()
		}
		
	// Status messages
	case messages.StatusMsg:
		logging.Info(t.Logger, "Status: %s", msg.Message)
		
	// Connection events
	case messages.ConnectedMsg:
		logging.Info(t.Logger, "Connected to cluster: %s, namespace: %s", msg.ClusterName, msg.Namespace)
		
	case messages.DisconnectedMsg:
		logging.Info(t.Logger, "Disconnected from cluster")
		
	// Navigation messages
	case navigation.NavigationMsg:
		return t.handleNavigationMessage(msg)
		
	case navigation.ModeChangeMsg:
		return t.handleModeChange(msg)
		
	case navigation.SearchMsg:
		return t.handleSearchMessage(msg)
		
	case navigation.CommandMsg:
		return t.handleCommandMessage(msg)
	}
	
	// Update all components if they're initialized - CRITICAL for BubbleTea pattern
	if t.componentsInitialized {
		// Update log pane
		if t.logPane != nil {
			var cmd tea.Cmd
			t.logPane, cmd = t.logPane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
		// Update other components as needed
		if t.contentPane != nil {
			var cmd tea.Cmd
			t.contentPane, cmd = t.contentPane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
		if t.detailPane != nil {
			var cmd tea.Cmd
			t.detailPane, cmd = t.detailPane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	
	// Return updated model and any commands to run
	if len(cmds) > 0 {
		return t, tea.Batch(cmds...)
	}
	return t, nil
}

// handleKeyInput processes keyboard input and returns updated model and commands
func (t *TUI) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings that work in any state
	switch msg.String() {
	case "ctrl+c", "q":
		logging.Info(t.Logger, "User requested quit")
		return t, tea.Quit
		
	case "ctrl+d":
		// Toggle debug mode
		t.Debug = !t.Debug
		t.Logger = logging.SetupLogger(t.Debug)
		logging.Info(t.Logger, "Debug mode toggled: %v", t.Debug)
		return t, nil
	}
	
	// State-specific keybindings
	switch t.State {
	case models.StateMain:
		return t.handleMainStateKeys(msg)
	case models.StateHelp:
		return t.handleHelpStateKeys(msg)
	case models.StateError:
		return t.handleErrorStateKeys(msg)
	case models.StateLoading:
		// No input handling during loading
		return t, nil
	}
	
	return t, nil
}

// handleMainStateKeys handles keyboard input in the main application state
func (t *TUI) handleMainStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		t.ToggleHelp()
		return t, nil
		
	case "tab", "l":
		t.NextTab()
		logging.Debug(t.Logger, "Switched to tab: %s", t.GetTabName(t.ActiveTab))
		return t, nil
		
	case "shift+tab", "h":
		t.PrevTab()
		logging.Debug(t.Logger, "Switched to tab: %s", t.GetTabName(t.ActiveTab))
		return t, nil
		
	case "r", "f5":
		// Trigger refresh
		return t, func() tea.Msg {
			return messages.RefreshMsg{}
		}
	}
	
	return t, nil
}

// handleHelpStateKeys handles keyboard input in the help overlay state
func (t *TUI) handleHelpStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc":
		t.ToggleHelp()
		return t, nil
	}
	return t, nil
}

// handleErrorStateKeys handles keyboard input in the error state
func (t *TUI) handleErrorStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		t.ClearError()
		return t, nil
	}
	return t, nil
}

// View implements tea.Model interface - renders the current state
// This method must be PURE - no state mutations or side effects allowed
func (t *TUI) View() string {
	// CRITICAL: Never render until we receive WindowSizeMsg (prevents header issues)
	if !t.isReady {
		return "Initializing LazyOC... (Press q to quit)"
	}
	
	// Handle zero dimensions with simple fallback (no state mutation)
	if t.Width == 0 || t.Height == 0 {
		return "Initializing LazyOC... (Press q to quit)"
	}
	
	switch t.State {
	case models.StateLoading:
		return t.renderLoading()
	case models.StateError:
		return t.renderError()
	case models.StateHelp:
		return t.renderHelp()
	case models.StateMain:
		return t.renderMain()
	default:
		// Fallback for unknown state (no state mutation)
		return "Unknown state - Press q to quit"
	}
}

// renderLoading renders the loading state
func (t *TUI) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(t.Width).
		Height(t.Height).
		Align(lipgloss.Center, lipgloss.Center)
		
	content := fmt.Sprintf("🚀 %s\n\nLoading LazyOC v%s...", t.LoadingMessage, t.Version)
	
	return style.Render(content)
}

// renderError renders the error state
func (t *TUI) renderError() string {
	style := lipgloss.NewStyle().
		Width(t.Width).
		Height(t.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("9")) // Red
		
	errorMsg := "An error occurred"
	if t.LastError != nil {
		errorMsg = t.LastError.Error()
		
		// Show additional context for AppError
		if appErr, ok := t.LastError.(*errors.AppError); ok {
			errorMsg = fmt.Sprintf("%s\n\nType: %s", errorMsg, appErr.GetTypeString())
		}
	}
		
	content := fmt.Sprintf("❌ Error\n\n%s\n\nPress ESC or Enter to continue", errorMsg)
	
	return style.Render(content)
}

// renderHelp renders the help overlay using the navigation system
func (t *TUI) renderHelp() string {
	if t.helpComponent != nil && t.navController != nil {
		// Update help component dimensions
		t.helpComponent.SetDimensions(t.Width, t.Height)
		
		// Render using the navigation registry
		return t.helpComponent.Render(t.navController.GetRegistry())
	}
	
	// Fallback to simple help if navigation system isn't available
	helpText := `
📖 LazyOC Help - Navigation System Unavailable

Basic Keys:
  q, Ctrl+C    Quit application
  ?            Toggle help
  Tab          Switch tabs
  
Press ? or ESC to close help
`
	
	style := lipgloss.NewStyle().
		Width(t.Width).
		Height(t.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")) // Blue
		
	return style.Render(helpText)
}

// renderMain renders the main application interface with defensive nil checks
func (t *TUI) renderMain() string {
	// Try to use initialized components first, fall back to simple rendering
	if t.layoutManager != nil && t.header != nil && t.tabs != nil && t.contentPane != nil && t.detailPane != nil && t.logPane != nil && t.statusBar != nil {
		return t.renderWithComponents()
	}
	
	// Fallback to simple rendering if components aren't ready
	return t.renderSimple()
}

// renderWithComponents renders using the initialized layout components
func (t *TUI) renderWithComponents() string {
	// Update component content BEFORE rendering to avoid recursion
	t.updateComponentsContent()
	
	var parts []string
	var usedHeight int
	
	// Header - measure dynamically
	if t.header != nil {
		headerView := t.header.Render()
		parts = append(parts, headerView)
		usedHeight += lipgloss.Height(headerView)
	}
	
	// Tabs - measure dynamically  
	if t.tabs != nil {
		tabsView := t.tabs.Render()
		parts = append(parts, tabsView)
		usedHeight += lipgloss.Height(tabsView)
	}
	
	// Status bar - render first to measure
	var statusView string
	if t.statusBar != nil {
		statusView = t.statusBar.Render()
	} else {
		// Fallback status
		statusView = lipgloss.NewStyle().
			Width(t.Width).
			Foreground(lipgloss.Color("8")).
			Background(lipgloss.Color("0")).
			Render(fmt.Sprintf("Ready • %s • Debug: %v", t.GetTabName(t.ActiveTab), t.Debug))
	}
	usedHeight += lipgloss.Height(statusView)
	
	// Calculate remaining height for main area DYNAMICALLY
	remainingHeight := t.Height - usedHeight
	if remainingHeight < 3 {
		remainingHeight = 3 // Minimum content height
	}
	
	// Main content area with calculated height
	mainArea := t.renderMainAreaWithHeight(remainingHeight)
	parts = append(parts, mainArea)
	
	// Add status bar last
	parts = append(parts, statusView)
	
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderMainAreaWithHeight renders the main content area with specified height
func (t *TUI) renderMainAreaWithHeight(availableHeight int) string {
	var rows []string
	
	// Top row: content pane and detail pane side by side
	var topRow []string
	
	// Content pane (always present) - calculate height for log pane space
	logHeight := 0
	if t.logPane != nil && t.layoutManager.LogPaneVisible {
		logHeight = t.logPane.GetEffectiveHeight() + 1 // +1 for spacing
	}
	
	contentHeight := availableHeight - logHeight
	if contentHeight < 3 {
		contentHeight = 3
	}
	
	// Update content pane height dynamically
	if t.contentPane != nil {
		t.contentPane.SetDimensions(t.contentPane.Width, contentHeight)
		topRow = append(topRow, t.contentPane.Render())
	}
	
	// Detail pane (if visible)
	if t.detailPane != nil && t.layoutManager.DetailPaneVisible {
		t.detailPane.SetDimensions(t.detailPane.Width, contentHeight)
		topRow = append(topRow, t.detailPane.Render())
	}
	
	if len(topRow) > 0 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, topRow...))
	}
	
	// Bottom row: log pane (if visible)
	if t.logPane != nil && t.layoutManager.LogPaneVisible {
		logContent := t.logPane.Render()
		if logContent != "" {
			rows = append(rows, logContent)
		}
	}
	
	if len(rows) == 0 {
		// Fallback if no components are ready
		return "No components ready"
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderMainArea renders the main content area with proper layout (legacy method)
func (t *TUI) renderMainArea() string {
	return t.renderMainAreaWithHeight(t.Height - 6) // Fallback with estimate
}

// renderSimple provides simple fallback rendering when components aren't ready
func (t *TUI) renderSimple() string {
	header := lipgloss.NewStyle().
		Width(t.Width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Render("🚀 LazyOC v" + t.Version)
		
	tabs := lipgloss.NewStyle().
		Width(t.Width).
		Align(lipgloss.Center).
		Render("[ Pods ] [ Services ] [ Deployments ] [ ConfigMaps ] [ Secrets ]")
	
	contentHeight := t.Height - 4 // Account for header, tabs, status
	if contentHeight < 1 {
		contentHeight = 1
	}
	
	content := lipgloss.NewStyle().
		Width(t.Width).
		Height(contentHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1).
		Render(fmt.Sprintf("📦 %s Resources\n\nNo cluster connected yet.\n\nUse Tab/Shift+Tab or h/l to navigate tabs\nPress ? for help\nPress q to quit", t.GetTabName(t.ActiveTab)))
	
	status := lipgloss.NewStyle().
		Width(t.Width).
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0")).
		Render(fmt.Sprintf("Ready • %s • Debug: %v", t.GetTabName(t.ActiveTab), t.Debug))
	
	return lipgloss.JoinVertical(lipgloss.Left, header, tabs, content, status)
}

// renderHeader renders the application header
func (t *TUI) renderHeader() string {
	style := lipgloss.NewStyle().
		Width(t.Width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("12")). // Blue
		Bold(true)
		
	return style.Render(fmt.Sprintf("🚀 LazyOC v%s", t.Version))
}

// renderTabs renders the tab navigation
func (t *TUI) renderTabs() string {
	var tabs []string
	
	allTabs := []models.TabType{
		models.TabPods,
		models.TabServices, 
		models.TabDeployments,
		models.TabConfigMaps,
		models.TabSecrets,
	}
	
	for _, tab := range allTabs {
		name := t.GetTabName(tab)
		if tab == t.ActiveTab {
			// Active tab style
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")). // White
				Background(lipgloss.Color("12")). // Blue
				Padding(0, 1).
				Bold(true)
			tabs = append(tabs, style.Render(name))
		} else {
			// Inactive tab style
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")). // Gray
				Padding(0, 1)
			tabs = append(tabs, style.Render(name))
		}
	}
	
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	
	// Center the tab bar
	style := lipgloss.NewStyle().
		Width(t.Width).
		Align(lipgloss.Center)
		
	return style.Render(tabBar)
}

// renderContent renders the main content area
func (t *TUI) renderContent() string {
	// For now, show placeholder content
	content := fmt.Sprintf("📦 %s Resources\n\n", t.GetTabName(t.ActiveTab))
	content += "No cluster connected yet.\n\n"
	content += "Use Tab/Shift+Tab or h/l to navigate tabs\n"
	content += "Press ? for help"
	
	return content
}

// renderStatusBar renders the bottom status bar
func (t *TUI) renderStatusBar() string {
	leftStatus := fmt.Sprintf("Ready • %s", t.GetTabName(t.ActiveTab))
	rightStatus := fmt.Sprintf("Debug: %v", t.Debug)
	
	// Calculate spacing
	statusWidth := t.Width - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus)
	spacing := ""
	if statusWidth > 0 {
		spacing = lipgloss.NewStyle().Width(statusWidth).Render("")
	}
	
	statusBar := lipgloss.JoinHorizontal(lipgloss.Top,
		leftStatus,
		spacing, 
		rightStatus,
	)
	
	style := lipgloss.NewStyle().
		Width(t.Width).
		Foreground(lipgloss.Color("8")). // Gray
		Background(lipgloss.Color("0"))   // Black
		
	return style.Render(statusBar)
}

// safeInitializeLayout safely initializes or updates the layout components with error handling
func (t *TUI) safeInitializeLayout(width, height int) {
	// Add error recovery
	defer func() {
		if r := recover(); r != nil {
			logging.Error(t.Logger, "Layout initialization panic: %v", r)
		}
	}()
	
	t.initializeOrUpdateLayout(width, height)
}

// initializeOrUpdateLayout initializes or updates the layout components
func (t *TUI) initializeOrUpdateLayout(width, height int) {
	// Initialize layout manager if not exists
	if t.layoutManager == nil {
		t.layoutManager = components.NewLayoutManager(width, height)
	} else {
		t.layoutManager.UpdateDimensions(width, height)
	}
	
	// Initialize header component
	if t.header == nil {
		headerDims := t.layoutManager.GetPanelDimensions(components.PanelHeader)
		t.header = components.NewHeaderComponent(headerDims.Width, headerDims.Height, t.Version)
	} else {
		headerDims := t.layoutManager.GetPanelDimensions(components.PanelHeader)
		t.header.SetDimensions(headerDims.Width, headerDims.Height)
	}
	
	// Initialize tabs component
	if t.tabs == nil {
		tabsDims := t.layoutManager.GetPanelDimensions(components.PanelTabs)
		t.tabs = components.CreateKubernetesTabComponent(tabsDims.Width, tabsDims.Height)
	} else {
		tabsDims := t.layoutManager.GetPanelDimensions(components.PanelTabs)
		t.tabs.SetDimensions(tabsDims.Width, tabsDims.Height)
	}
	
	// Initialize content pane
	if t.contentPane == nil {
		contentDims := t.layoutManager.GetPanelDimensions(components.PanelMain)
		t.contentPane = components.NewContentPane(contentDims.Width, contentDims.Height)
		t.contentPane.SetTitle("Resources")
	} else {
		contentDims := t.layoutManager.GetPanelDimensions(components.PanelMain)
		t.contentPane.SetDimensions(contentDims.Width, contentDims.Height)
	}
	
	// Initialize detail pane
	if t.detailPane == nil {
		detailDims := t.layoutManager.GetPanelDimensions(components.PanelDetail)
		t.detailPane = components.NewDetailPane(detailDims.Width, detailDims.Height)
	} else {
		detailDims := t.layoutManager.GetPanelDimensions(components.PanelDetail)
		t.detailPane.SetDimensions(detailDims.Width, detailDims.Height)
	}
	
	// Initialize log pane
	if t.logPane == nil {
		logDims := t.layoutManager.GetPanelDimensions(components.PanelLog)
		t.logPane = components.NewLogPane(logDims.Width, logDims.Height)
		
		// Add some initial log entries
		t.logPane.Info("LazyOC application started", "tui")
		t.logPane.Debug("Layout system initialized", "layout")
	} else {
		logDims := t.layoutManager.GetPanelDimensions(components.PanelLog)
		t.logPane.SetDimensions(logDims.Width, logDims.Height)
	}
	
	// Initialize status bar
	if t.statusBar == nil {
		statusDims := t.layoutManager.GetPanelDimensions(components.PanelStatusBar)
		t.statusBar = components.NewStatusBarComponent(statusDims.Width, statusDims.Height)
		t.statusBar.SetStatus("Ready", components.StatusInfo)
		t.statusBar.SetKeyHints(components.CreateDefaultKeyHints())
	} else {
		statusDims := t.layoutManager.GetPanelDimensions(components.PanelStatusBar)
		t.statusBar.SetDimensions(statusDims.Width, statusDims.Height)
	}
}

// updateComponentsContent safely updates component content and state (renamed to avoid recursion)
func (t *TUI) updateComponentsContent() {
	// Early return if essential components aren't ready
	if t.layoutManager == nil {
		return
	}
	
	if t.header != nil {
		// Update header with current state - placeholder for now
		t.header.SetClusterInfo("", "default", false)
	}
	
	if t.tabs != nil {
		// Sync tab state with app state
		activeTabID := t.GetTabName(t.ActiveTab)
		t.tabs.SetActiveTabByID(strings.ToLower(activeTabID))
	}
	
	if t.contentPane != nil {
		// Update content based on active tab
		activeTabName := t.GetTabName(t.ActiveTab)
		placeholder := fmt.Sprintf("📦 %s Resources\n\nNo cluster connected yet.\n\nUse Tab/Shift+Tab or h/l to navigate tabs\nPress ? for help", activeTabName)
		t.contentPane.SetContent(placeholder)
		t.contentPane.SetTitle(activeTabName)
		
		// Set focus state
		t.contentPane.SetFocus(t.layoutManager.GetFocus() == components.PanelMain)
	}
	
	if t.detailPane != nil {
		// Update detail pane visibility
		t.detailPane.SetVisible(t.layoutManager.DetailPaneVisible)
		t.detailPane.SetFocus(t.layoutManager.GetFocus() == components.PanelDetail)
	}
	
	if t.logPane != nil {
		// Update log pane visibility
		t.logPane.SetVisible(t.layoutManager.LogPaneVisible)
		t.logPane.SetFocus(t.layoutManager.GetFocus() == components.PanelLog)
	}
	
	if t.statusBar != nil {
		// Update status bar with current state
		t.statusBar.SetActivePanel(string(rune('0' + int(t.layoutManager.GetFocus()))))
		t.statusBar.UpdateTimestamp()
	}
}

// testLayoutManagerOnly initializes layout manager + header for testing
func (t *TUI) testLayoutManagerOnly(width, height int) {
	defer func() {
		if r := recover(); r != nil {
			logging.Error(t.Logger, "Component initialization panic: %v", r)
		}
	}()
	
	// Initialize layout manager if not exists
	if t.layoutManager == nil {
		t.layoutManager = components.NewLayoutManager(width, height)
		logging.Debug(t.Logger, "Layout manager initialized: %dx%d", width, height)
	} else {
		t.layoutManager.UpdateDimensions(width, height)
		logging.Debug(t.Logger, "Layout manager updated: %dx%d", width, height)
	}
	
	// Test header component
	if t.header == nil {
		headerDims := t.layoutManager.GetPanelDimensions(components.PanelHeader)
		t.header = components.NewHeaderComponent(headerDims.Width, headerDims.Height, t.Version)
		logging.Debug(t.Logger, "Header component initialized: %dx%d", headerDims.Width, headerDims.Height)
	} else {
		headerDims := t.layoutManager.GetPanelDimensions(components.PanelHeader)
		t.header.SetDimensions(headerDims.Width, headerDims.Height)
		logging.Debug(t.Logger, "Header component updated: %dx%d", headerDims.Width, headerDims.Height)
	}
	
	// Test tabs component
	if t.tabs == nil {
		tabsDims := t.layoutManager.GetPanelDimensions(components.PanelTabs)
		t.tabs = components.CreateKubernetesTabComponent(tabsDims.Width, tabsDims.Height)
		logging.Debug(t.Logger, "Tabs component initialized: %dx%d", tabsDims.Width, tabsDims.Height)
	} else {
		tabsDims := t.layoutManager.GetPanelDimensions(components.PanelTabs)
		t.tabs.SetDimensions(tabsDims.Width, tabsDims.Height)
		logging.Debug(t.Logger, "Tabs component updated: %dx%d", tabsDims.Width, tabsDims.Height)
	}
	
	// Test content pane component (likely culprit - has viewport)
	if t.contentPane == nil {
		contentDims := t.layoutManager.GetPanelDimensions(components.PanelMain)
		t.contentPane = components.NewContentPane(contentDims.Width, contentDims.Height)
		t.contentPane.SetTitle("Resources")
		logging.Debug(t.Logger, "Content pane initialized: %dx%d", contentDims.Width, contentDims.Height)
	} else {
		contentDims := t.layoutManager.GetPanelDimensions(components.PanelMain)
		t.contentPane.SetDimensions(contentDims.Width, contentDims.Height)
		logging.Debug(t.Logger, "Content pane updated: %dx%d", contentDims.Width, contentDims.Height)
	}
	
	// Test detail pane component
	if t.detailPane == nil {
		detailDims := t.layoutManager.GetPanelDimensions(components.PanelDetail)
		t.detailPane = components.NewDetailPane(detailDims.Width, detailDims.Height)
		logging.Debug(t.Logger, "Detail pane initialized: %dx%d", detailDims.Width, detailDims.Height)
	} else {
		detailDims := t.layoutManager.GetPanelDimensions(components.PanelDetail)
		t.detailPane.SetDimensions(detailDims.Width, detailDims.Height)
		logging.Debug(t.Logger, "Detail pane updated: %dx%d", detailDims.Width, detailDims.Height)
	}
	
	// Test log pane component (potential culprit - might have viewport or complex init)
	if t.logPane == nil {
		logDims := t.layoutManager.GetPanelDimensions(components.PanelLog)
		t.logPane = components.NewLogPane(logDims.Width, logDims.Height)
		
		// Add log entries AFTER dimensions are set
		// (Don't add them here during initialization)
		logging.Debug(t.Logger, "Log pane initialized: %dx%d", logDims.Width, logDims.Height)
	} else {
		logDims := t.layoutManager.GetPanelDimensions(components.PanelLog)
		t.logPane.SetDimensions(logDims.Width, logDims.Height)
		logging.Debug(t.Logger, "Log pane updated: %dx%d", logDims.Width, logDims.Height)
	}
	
	// Add status bar component to complete the layout
	if t.statusBar == nil {
		statusDims := t.layoutManager.GetPanelDimensions(components.PanelStatusBar)
		t.statusBar = components.NewStatusBarComponent(statusDims.Width, statusDims.Height)
		t.statusBar.SetStatus("Ready", components.StatusInfo)
		t.statusBar.SetKeyHints(components.CreateDefaultKeyHints())
		logging.Debug(t.Logger, "Status bar initialized: %dx%d", statusDims.Width, statusDims.Height)
	} else {
		statusDims := t.layoutManager.GetPanelDimensions(components.PanelStatusBar)
		t.statusBar.SetDimensions(statusDims.Width, statusDims.Height)
		logging.Debug(t.Logger, "Status bar updated: %dx%d", statusDims.Width, statusDims.Height)
	}
	
	// Add some initial log entries after proper initialization
	if t.logPane != nil && t.logPane.Ready {
		t.logPane.Info("LazyOC application started", "tui")
		t.logPane.Debug("Layout system initialized", "layout") 
		t.logPane.Debug("All components initialized successfully", "tui")
		t.logPane.Info("Kubernetes TUI ready for cluster connection", "app")
		t.logPane.Warn("No cluster connected - connect to view resources", "cluster")
	}
}

// updateAllComponentDimensions updates dimensions for all initialized components
func (t *TUI) updateAllComponentDimensions() {
	if t.layoutManager == nil {
		return
	}
	
	// Update header
	if t.header != nil {
		headerDims := t.layoutManager.GetPanelDimensions(components.PanelHeader)
		t.header.SetDimensions(headerDims.Width, headerDims.Height)
		logging.Debug(t.Logger, "Header dimensions updated: %dx%d", headerDims.Width, headerDims.Height)
	}
	
	// Update tabs
	if t.tabs != nil {
		tabsDims := t.layoutManager.GetPanelDimensions(components.PanelTabs)
		t.tabs.SetDimensions(tabsDims.Width, tabsDims.Height)
		logging.Debug(t.Logger, "Tabs dimensions updated: %dx%d", tabsDims.Width, tabsDims.Height)
	}
	
	// Update content pane
	if t.contentPane != nil {
		contentDims := t.layoutManager.GetPanelDimensions(components.PanelMain)
		t.contentPane.SetDimensions(contentDims.Width, contentDims.Height)
		logging.Debug(t.Logger, "Content pane dimensions updated: %dx%d", contentDims.Width, contentDims.Height)
	}
	
	// Update detail pane
	if t.detailPane != nil {
		detailDims := t.layoutManager.GetPanelDimensions(components.PanelDetail)
		t.detailPane.SetDimensions(detailDims.Width, detailDims.Height)
		logging.Debug(t.Logger, "Detail pane dimensions updated: %dx%d", detailDims.Width, detailDims.Height)
	}
	
	// Update log pane
	if t.logPane != nil {
		logDims := t.layoutManager.GetPanelDimensions(components.PanelLog)
		t.logPane.SetDimensions(logDims.Width, logDims.Height)
		logging.Debug(t.Logger, "Log pane dimensions updated: %dx%d", logDims.Width, logDims.Height)
	}
	
	// Update status bar
	if t.statusBar != nil {
		statusDims := t.layoutManager.GetPanelDimensions(components.PanelStatusBar)
		t.statusBar.SetDimensions(statusDims.Width, statusDims.Height)
		logging.Debug(t.Logger, "Status bar dimensions updated: %dx%d", statusDims.Width, statusDims.Height)
	}
}

// setupNavigationCallbacks sets up callbacks for navigation actions
func (t *TUI) setupNavigationCallbacks() {
	t.navController.SetCallback(navigation.ActionQuit, func() tea.Cmd {
		logging.Info(t.Logger, "User requested quit via navigation")
		return tea.Quit
	})
	
	t.navController.SetCallback(navigation.ActionToggleHelp, func() tea.Cmd {
		t.ToggleHelp()
		return nil
	})
	
	t.navController.SetCallback(navigation.ActionToggleDebug, func() tea.Cmd {
		t.Debug = !t.Debug
		t.Logger = logging.SetupLogger(t.Debug)
		logging.Info(t.Logger, "Debug mode toggled: %v", t.Debug)
		return nil
	})
	
	t.navController.SetCallback(navigation.ActionRefresh, func() tea.Cmd {
		return func() tea.Msg {
			return messages.RefreshMsg{}
		}
	})
}

// handleKeyInputWithNavigation processes keyboard input using the navigation system
func (t *TUI) handleKeyInputWithNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Process the key through the navigation controller
	cmds, handled := t.navController.ProcessKeyMsg(msg)
	
	if handled {
		// Navigation system handled the key
		if len(cmds) > 0 {
			return t, tea.Batch(cmds...)
		}
		return t, nil
	}
	
	// Fall back to the original key handling for unhandled keys
	return t.handleKeyInput(msg)
}

// handleNavigationMessage handles navigation-specific messages
func (t *TUI) handleNavigationMessage(msg navigation.NavigationMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Update focus in layout manager
	if t.layoutManager != nil {
		t.layoutManager.SetFocus(msg.Panel)
		
		// Update component focus states
		if t.contentPane != nil {
			t.contentPane.SetFocus(msg.Panel == components.PanelMain)
		}
		if t.detailPane != nil {
			t.detailPane.SetFocus(msg.Panel == components.PanelDetail)
		}
		if t.logPane != nil {
			t.logPane.SetFocus(msg.Panel == components.PanelLog)
		}
	}
	
	// Route actions to appropriate components
	switch msg.Panel {
	case components.PanelMain:
		if t.contentPane != nil {
			cmds = append(cmds, t.routeActionToContentPane(msg.Action))
		}
	case components.PanelDetail:
		if t.detailPane != nil {
			cmds = append(cmds, t.routeActionToDetailPane(msg.Action))
		}
	case components.PanelLog:
		if t.logPane != nil {
			cmds = append(cmds, t.routeActionToLogPane(msg.Action))
		}
	}
	
	// Handle tab navigation
	switch msg.Action {
	case navigation.ActionNextTab:
		t.NextTab()
		logging.Debug(t.Logger, "Navigation: Switched to tab: %s", t.GetTabName(t.ActiveTab))
	case navigation.ActionPrevTab:
		t.PrevTab()
		logging.Debug(t.Logger, "Navigation: Switched to tab: %s", t.GetTabName(t.ActiveTab))
	}
	
	if len(cmds) > 0 {
		return t, tea.Batch(cmds...)
	}
	return t, nil
}

// handleModeChange handles navigation mode changes
func (t *TUI) handleModeChange(msg navigation.ModeChangeMsg) (tea.Model, tea.Cmd) {
	logging.Debug(t.Logger, "Navigation mode changed: %s -> %s", msg.OldMode, msg.NewMode)
	
	// Update help component with new mode
	if t.helpComponent != nil {
		t.helpComponent.SetCurrentMode(msg.NewMode)
	}
	
	// Update status bar to show current mode
	if t.statusBar != nil {
		modeIndicator := t.navController.GetRegistry().GetModeIndicator()
		t.statusBar.SetActivePanel(modeIndicator)
	}
	
	return t, nil
}

// handleSearchMessage handles search-related messages
func (t *TUI) handleSearchMessage(msg navigation.SearchMsg) (tea.Model, tea.Cmd) {
	if msg.Complete {
		logging.Info(t.Logger, "Search executed: %s", msg.Query)
		// TODO: Implement search functionality
	}
	return t, nil
}

// handleCommandMessage handles command-related messages
func (t *TUI) handleCommandMessage(msg navigation.CommandMsg) (tea.Model, tea.Cmd) {
	if msg.Complete {
		logging.Info(t.Logger, "Command executed: %s", msg.Command)
		// TODO: Implement command functionality
	}
	return t, nil
}

// routeActionToContentPane routes navigation actions to the content pane
func (t *TUI) routeActionToContentPane(action navigation.KeyAction) tea.Cmd {
	// Create a mock KeyMsg to pass to the content pane
	// This is a bridge between the navigation system and component updates
	var keyStr string
	
	switch action {
	case navigation.ActionMoveUp:
		keyStr = "up"
	case navigation.ActionMoveDown:
		keyStr = "down"
	case navigation.ActionMoveLeft:
		keyStr = "left"
	case navigation.ActionMoveRight:
		keyStr = "right"
	case navigation.ActionPageUp:
		keyStr = "pageup"
	case navigation.ActionPageDown:
		keyStr = "pagedown"
	case navigation.ActionGoToTop:
		keyStr = "home"
	case navigation.ActionGoToBottom:
		keyStr = "end"
	case navigation.ActionSelect:
		keyStr = "enter"
	default:
		return nil
	}
	
	return func() tea.Msg {
		// This creates a message that the content pane can handle
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
	}
}

// routeActionToDetailPane routes navigation actions to the detail pane
func (t *TUI) routeActionToDetailPane(action navigation.KeyAction) tea.Cmd {
	// Similar routing for detail pane actions
	switch action {
	case navigation.ActionToggleCollapse:
		if t.detailPane != nil {
			t.detailPane.ToggleCollapse()
		}
	case navigation.ActionToggleVisible:
		if t.detailPane != nil {
			t.detailPane.Toggle()
		}
	}
	return nil
}

// routeActionToLogPane routes navigation actions to the log pane
func (t *TUI) routeActionToLogPane(action navigation.KeyAction) tea.Cmd {
	switch action {
	case navigation.ActionClearLogs:
		if t.logPane != nil {
			t.logPane.ClearLogs()
		}
	case navigation.ActionToggleAutoscroll:
		if t.logPane != nil {
			t.logPane.ToggleAutoScroll()
		}
	case navigation.ActionTogglePause:
		if t.logPane != nil {
			t.logPane.TogglePause()
		}
	}
	return nil
}