package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/katyella/lazyoc/internal/logging"
	"github.com/katyella/lazyoc/internal/ui/messages"
	"github.com/katyella/lazyoc/internal/ui/models"
	"github.com/katyella/lazyoc/internal/ui/navigation"
)

// SimplifiedTUI is a streamlined version without complex component initialization
type SimplifiedTUI struct {
	*models.App
	
	// Navigation system (keep this as it works well)
	navController *navigation.NavigationController
	
	// Simple state instead of components
	width         int
	height        int
	ready         bool
	showHelp      bool
	focusedPanel  int
	
	// Content
	mainContent   string
	logContent    []string
	detailContent string
	
	// Visibility
	showDetails   bool
	showLogs      bool
	
	// Theme
	theme string
}

// NewSimplifiedTUI creates a new simplified TUI instance
func NewSimplifiedTUI(version string, debug bool) *SimplifiedTUI {
	app := models.NewApp(version)
	app.Debug = debug
	app.Logger = logging.SetupLogger(debug)
	
	logging.Info(app.Logger, "Initializing Simplified LazyOC TUI v%s", version)
	
	tui := &SimplifiedTUI{
		App:           app,
		navController: navigation.NewNavigationController(),
		theme:         "dark", // default theme
		showDetails:   true,
		showLogs:      true,
		focusedPanel:  0, // 0=main, 1=details, 2=logs
		mainContent:   fmt.Sprintf("📦 %s Resources\n\nNo cluster connected yet.\n\nPress ? for help", "Pods"),
		logContent:    []string{"LazyOC started", "Waiting for cluster connection..."},
		detailContent: "Select a resource to view details",
	}
	
	// Set up navigation callbacks
	tui.setupNavigationCallbacks()
	
	return tui
}

// Init implements tea.Model
func (t *SimplifiedTUI) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		func() tea.Msg {
			return messages.InitMsg{}
		},
	)
}

// Update implements tea.Model
func (t *SimplifiedTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.ready = true
		logging.Debug(t.Logger, "Window size: %dx%d", t.width, t.height)
		
	case tea.KeyMsg:
		// Special handling for help mode
		if t.showHelp {
			if msg.String() == "?" || msg.String() == "esc" {
				t.showHelp = false
				return t, nil
			}
			return t, nil
		}
		
		// Normal key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return t, tea.Quit
			
		case "?":
			t.showHelp = true
			return t, nil
			
		case "tab":
			t.focusedPanel = (t.focusedPanel + 1) % 3
			return t, nil
			
		case "shift+tab":
			t.focusedPanel = (t.focusedPanel + 2) % 3
			return t, nil
			
		case "d":
			t.showDetails = !t.showDetails
			return t, nil
			
		case "L":
			t.showLogs = !t.showLogs
			return t, nil
			
		case "h":
			if t.focusedPanel == 0 { // Only navigate tabs when in main panel
				t.PrevTab()
				t.updateMainContent()
			}
			return t, nil
			
		case "l":
			if t.focusedPanel == 0 { // Only navigate tabs when in main panel
				t.NextTab()
				t.updateMainContent()
			}
			return t, nil
			
		case "left":
			t.PrevTab()
			t.updateMainContent()
			return t, nil
			
		case "right":
			t.NextTab()
			t.updateMainContent()
			return t, nil
			
		case "j", "down":
			// Move focus down through panels
			if t.focusedPanel == 0 && t.showLogs {
				t.focusedPanel = 2
			} else if t.focusedPanel == 1 && t.showLogs {
				t.focusedPanel = 2
			}
			return t, nil
			
		case "k", "up":
			// Move focus up through panels
			if t.focusedPanel == 2 {
				t.focusedPanel = 0
			}
			return t, nil
			
		case "1":
			t.focusedPanel = 0 // Focus main panel
			return t, nil
			
		case "2":
			if t.showDetails {
				t.focusedPanel = 1 // Focus details panel
			}
			return t, nil
			
		case "3":
			if t.showLogs {
				t.focusedPanel = 2 // Focus logs panel
			}
			return t, nil
			
		case "t":
			// Toggle theme
			if t.theme == "dark" {
				t.theme = "light"
			} else {
				t.theme = "dark"
			}
			return t, nil
		}
		
	case messages.InitMsg:
		t.ClearLoading()
		t.State = models.StateMain
		// Add initial log entries
		t.logContent = append(t.logContent, "Application initialized")
		// Update main content for current tab
		t.updateMainContent()
		logging.Info(t.Logger, "Application initialized successfully")
	}
	
	return t, nil
}

// View implements tea.Model
func (t *SimplifiedTUI) View() string {
	// Don't render until we have dimensions
	if !t.ready || t.width == 0 || t.height == 0 {
		return "Initializing LazyOC..."
	}
	
	// Show help overlay if active
	if t.showHelp {
		return t.renderHelp()
	}
	
	// Render main interface
	return t.renderMain()
}

// renderMain renders the main interface using direct rendering
func (t *SimplifiedTUI) renderMain() string {
	var sections []string
	
	// Header (1-2 lines based on height)
	headerHeight := 2
	if t.height < 20 {
		headerHeight = 1
	}
	sections = append(sections, t.renderHeader(headerHeight))
	
	// Tabs (1 line)
	sections = append(sections, t.renderTabs())
	
	// Calculate remaining height
	usedHeight := headerHeight + 1 + 1 // header + tabs + status
	remainingHeight := t.height - usedHeight
	
	// Main content area
	if remainingHeight > 3 {
		sections = append(sections, t.renderContent(remainingHeight))
	}
	
	// Status bar (1 line)
	sections = append(sections, t.renderStatusBar())
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders a themed header
func (t *SimplifiedTUI) renderHeader(height int) string {
	primaryColor, errorColor := t.getThemeColors()
	headerStyle := lipgloss.NewStyle().
		Width(t.width).
		Align(lipgloss.Center).
		Foreground(primaryColor).
		Bold(true)
	
	if height == 1 {
		return headerStyle.Render(fmt.Sprintf("🚀 LazyOC v%s", t.Version))
	}
	
	// Two line header
	line1 := headerStyle.Render(fmt.Sprintf("🚀 LazyOC v%s", t.Version))
	line2 := lipgloss.NewStyle().
		Width(t.width).
		Align(lipgloss.Center).
		Foreground(errorColor).
		Render("○ Disconnected")
		
	return lipgloss.JoinVertical(lipgloss.Left, line1, line2)
}

// renderTabs renders the tab bar
func (t *SimplifiedTUI) renderTabs() string {
	tabs := []string{"Pods", "Services", "Deployments", "ConfigMaps", "Secrets"}
	var tabViews []string
	
	for i, tab := range tabs {
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == int(t.ActiveTab) {
			style = style.
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("12")).
				Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("8"))
		}
		tabViews = append(tabViews, style.Render(tab))
	}
	
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
	return lipgloss.NewStyle().
		Width(t.width).
		Align(lipgloss.Center).
		Render(tabBar)
}

// renderContent renders the main content area
func (t *SimplifiedTUI) renderContent(availableHeight int) string {
	// Calculate dimensions
	mainWidth := t.width
	if t.showDetails {
		mainWidth = t.width * 2 / 3
	}
	
	logHeight := 0
	if t.showLogs && availableHeight > 10 {
		logHeight = availableHeight / 3
		if logHeight < 5 {
			logHeight = 5
		}
	}
	
	mainHeight := availableHeight - logHeight
	
	// Main panel with theming
	primaryColor, _ := t.getThemeColors()
	borderColor := lipgloss.Color("240") // gray
	if t.focusedPanel == 0 {
		borderColor = primaryColor
	}
	
	mainStyle := lipgloss.NewStyle().
		Width(mainWidth - 2).
		Height(mainHeight - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1)
	
	mainPanel := mainStyle.Render(t.mainContent)
	
	// Detail panel
	var detailPanel string
	if t.showDetails {
		detailWidth := t.width - mainWidth
		detailBorderColor := lipgloss.Color("240") // Default gray
		if t.focusedPanel == 1 {
			detailBorderColor = lipgloss.Color("12") // Blue when focused
		}
		
		detailStyle := lipgloss.NewStyle().
			Width(detailWidth - 2).
			Height(mainHeight - 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(detailBorderColor).
			Padding(1)
		
		detailPanel = detailStyle.Render(t.detailContent)
	}
	
	// Combine main and detail panels
	var topSection string
	if t.showDetails {
		topSection = lipgloss.JoinHorizontal(lipgloss.Top, mainPanel, detailPanel)
	} else {
		topSection = mainPanel
	}
	
	// Log panel
	if t.showLogs && logHeight > 0 {
		logBorderColor := lipgloss.Color("240") // Default gray
		if t.focusedPanel == 2 {
			logBorderColor = lipgloss.Color("12") // Blue when focused
		}
		
		logStyle := lipgloss.NewStyle().
			Width(t.width - 2).
			Height(logHeight - 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(logBorderColor).
			Padding(1)
		
		// Join last N log entries
		logText := strings.Join(t.logContent[max(0, len(t.logContent)-10):], "\n")
		logPanel := logStyle.Render(logText)
		
		return lipgloss.JoinVertical(lipgloss.Left, topSection, logPanel)
	}
	
	return topSection
}

// renderStatusBar renders the status bar
func (t *SimplifiedTUI) renderStatusBar() string {
	panels := []string{"Main", "Details", "Logs"}
	// Style hints with different colors
	hintsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")) // Dimmer gray
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true) // White bold
	
	hints := fmt.Sprintf("%s help %s %s switch %s %s details %s %s logs %s %s quit",
		keyStyle.Render("?"), hintsStyle.Render("•"),
		keyStyle.Render("tab"), hintsStyle.Render("•"),
		keyStyle.Render("d"), hintsStyle.Render("•"),
		keyStyle.Render("L"), hintsStyle.Render("•"),
		keyStyle.Render("q"))
	
	// Add focus indicator
	focusIndicator := "◆"
	if t.focusedPanel >= 0 && t.focusedPanel < len(panels) {
		focusIndicator = fmt.Sprintf("◆ %s", panels[t.focusedPanel])
	}
	
	left := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")). // Blue for focused panel
		Bold(true).
		Render(focusIndicator)
	
	statusStyle := lipgloss.NewStyle().
		Width(t.width).
		Background(lipgloss.Color("236")). // Darker gray background
		Foreground(lipgloss.Color("15"))   // White text
		
	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	hintsWidth := lipgloss.Width(hints)
	spacing := t.width - leftWidth - hintsWidth
	if spacing < 0 {
		spacing = 1
	}
	
	status := left + strings.Repeat(" ", spacing) + hints
	return statusStyle.Render(status)
}

// renderHelp renders a simple help overlay
func (t *SimplifiedTUI) renderHelp() string {
	helpText := `📖 LazyOC Help

Navigation:
  tab        Next panel
  shift+tab  Previous panel
  j/k        Move focus down/up
  h/l        Previous/Next tab (in main panel)
  arrow keys Navigate tabs
  1/2/3      Jump to main/detail/log panel
  
Commands:
  ?          Toggle help
  d          Toggle details panel
  L          Toggle log panel (shift+l)
  q          Quit
  
Press ? or ESC to close`

	// Simple centered help box with better styling
	helpStyle := lipgloss.NewStyle().
		Width(60).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Background(lipgloss.Color("235")).
		Padding(1, 2).
		Align(lipgloss.Left)
		
	help := helpStyle.Render(helpText)
	
	// Center in screen
	return lipgloss.Place(
		t.width,
		t.height,
		lipgloss.Center,
		lipgloss.Center,
		help,
	)
}

// setupNavigationCallbacks sets up navigation callbacks
func (t *SimplifiedTUI) setupNavigationCallbacks() {
	t.navController.SetCallback(navigation.ActionQuit, func() tea.Cmd {
		return tea.Quit
	})
}

// updateMainContent updates the main content based on the active tab
func (t *SimplifiedTUI) updateMainContent() {
	tabName := t.GetTabName(t.ActiveTab)
	t.mainContent = fmt.Sprintf("📦 %s Resources\n\nNo cluster connected yet.\n\nUse h/l or arrow keys to navigate tabs\nPress ? for help", tabName)
}

// getThemeColors returns primary and error colors based on current theme
func (t *SimplifiedTUI) getThemeColors() (lipgloss.Color, lipgloss.Color) {
	if t.theme == "light" {
		return lipgloss.Color("4"), lipgloss.Color("1") // dark blue, dark red
	}
	return lipgloss.Color("12"), lipgloss.Color("9") // blue, red
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}