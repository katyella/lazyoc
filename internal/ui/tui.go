package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/katyella/lazyoc/internal/errors"
	"github.com/katyella/lazyoc/internal/logging"
	"github.com/katyella/lazyoc/internal/ui/messages"
	"github.com/katyella/lazyoc/internal/ui/models"
)

// TUI wraps the App model and implements the tea.Model interface
type TUI struct {
	*models.App
}

// NewTUI creates a new TUI instance
func NewTUI(version string, debug bool) *TUI {
	app := models.NewApp(version)
	app.Debug = debug
	app.Logger = logging.SetupLogger(debug)
	
	logging.Info(app.Logger, "Initializing LazyOC TUI v%s", version)
	
	return &TUI{
		App: app,
	}
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
		t.SetDimensions(msg.Width, msg.Height)
		logging.Debug(t.Logger, "Window resized to %dx%d", msg.Width, msg.Height)
		
	// Keyboard input
	case tea.KeyMsg:
		return t.handleKeyInput(msg)
		
	// Application initialization
	case messages.InitMsg:
		t.ClearLoading()
		t.State = models.StateMain
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
func (t *TUI) View() string {
	if t.Width == 0 || t.Height == 0 {
		return "Initializing..."
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
		return "Unknown state"
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

// renderHelp renders the help overlay
func (t *TUI) renderHelp() string {
	helpText := `
📖 LazyOC Help

Global Keys:
  q, Ctrl+C    Quit application
  ?            Toggle help
  Ctrl+D       Toggle debug mode

Navigation:
  Tab, l       Next tab
  Shift+Tab, h Previous tab
  r, F5        Refresh

Tabs:
  Pods         View pod resources
  Services     View service resources  
  Deployments  View deployment resources
  ConfigMaps   View configmap resources
  Secrets      View secret resources

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

// renderMain renders the main application interface
func (t *TUI) renderMain() string {
	header := t.renderHeader()
	tabs := t.renderTabs()
	content := t.renderContent()
	statusBar := t.renderStatusBar()
	
	// Calculate heights
	headerHeight := lipgloss.Height(header)
	tabsHeight := lipgloss.Height(tabs)
	statusHeight := 1
	contentHeight := t.Height - headerHeight - tabsHeight - statusHeight
	
	// Style the content area
	contentStyle := lipgloss.NewStyle().
		Width(t.Width).
		Height(contentHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8")) // Gray
		
	styledContent := contentStyle.Render(content)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabs,
		styledContent,
		statusBar,
	)
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