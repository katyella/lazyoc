package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/katyella/lazyoc/internal/k8s"
	"github.com/katyella/lazyoc/internal/k8s/auth"
	"github.com/katyella/lazyoc/internal/k8s/monitor"
	"github.com/katyella/lazyoc/internal/k8s/projects"
	"github.com/katyella/lazyoc/internal/k8s/resources"
	"github.com/katyella/lazyoc/internal/logging"
	"github.com/katyella/lazyoc/internal/ui/messages"
	"github.com/katyella/lazyoc/internal/ui/models"
	"github.com/katyella/lazyoc/internal/ui/navigation"
	"k8s.io/client-go/kubernetes"
)

// SimplifiedTUI is a streamlined version without complex component initialization
type SimplifiedTUI struct {
	*models.App
	
	// Navigation system (keep this as it works well)
	navController *navigation.NavigationController
	
	// Kubernetes client integration
	k8sClient       k8s.Client
	resourceClient  resources.ResourceClient
	connMonitor     monitor.ConnectionMonitor
	authProvider    auth.AuthProvider
	projectManager  projects.ProjectManager
	projectFactory  *projects.DefaultProjectManagerFactory
	
	// Connection state
	connected      bool
	connecting     bool
	connectionErr  error
	namespace      string
	context        string
	
	// Resource data
	pods           []resources.PodInfo
	selectedPod    int
	loadingPods    bool
	
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
	
	// Project switching modal
	showProjectModal   bool
	projectList        []projects.ProjectInfo
	selectedProject    int
	currentProject     *projects.ProjectInfo
	loadingProjects    bool
	switchingProject   bool
	projectModalHeight int
	projectError       string
	
	// Theme
	theme string
	
	// Kubeconfig path
	KubeconfigPath string
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
		mainContent:   "",  // Will be set by updateMainContent
		logContent:    []string{"LazyOC started"},
		detailContent: "Select a resource to view details",
		namespace:     "default", // default namespace
		pods:          []resources.PodInfo{},
		selectedPod:   0,
	}
	
	// Set up navigation callbacks
	tui.setupNavigationCallbacks()
	
	// Initialize main content
	tui.updateMainContent()
	
	return tui
}

// SetKubeconfig sets the kubeconfig path and returns a command to initialize the connection
func (t *SimplifiedTUI) SetKubeconfig(kubeconfigPath string) tea.Cmd {
	if kubeconfigPath == "" {
		// Try default location
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}
	
	return tea.Batch(
		// First send connecting message
		func() tea.Msg {
			return messages.ConnectingMsg{KubeconfigPath: kubeconfigPath}
		},
		// Then initialize the client
		t.InitializeK8sClient(kubeconfigPath),
	)
}

// Init implements tea.Model
func (t *SimplifiedTUI) Init() tea.Cmd {
	var cmds []tea.Cmd
	
	// Basic initialization commands
	cmds = append(cmds, 
		tea.WindowSize(),
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return messages.InitMsg{}
		}),
	)
	
	// If kubeconfig is provided, initialize the connection
	if t.KubeconfigPath != "" {
		cmds = append(cmds, t.SetKubeconfig(t.KubeconfigPath))
	} else {
		// Try default kubeconfig location
		home, err := os.UserHomeDir()
		if err == nil {
			defaultPath := filepath.Join(home, ".kube", "config")
			if _, err := os.Stat(defaultPath); err == nil {
				cmds = append(cmds, t.SetKubeconfig(defaultPath))
			} else {
				// No kubeconfig found - send message
				cmds = append(cmds, func() tea.Msg {
					return messages.NoKubeconfigMsg{
						Message: "No kubeconfig found at ~/.kube/config",
					}
				})
			}
		} else {
			// Couldn't get home dir - send message
			cmds = append(cmds, func() tea.Msg {
				return messages.NoKubeconfigMsg{
					Message: "No kubeconfig specified",
				}
			})
		}
	}
	
	return tea.Batch(cmds...)
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
		
		// Special handling for project modal
		if t.showProjectModal {
			return t.handleProjectModalKeys(msg)
		}
		
		// Normal key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return t, tea.Quit
			
		case "ctrl+p":
			// Open project switching modal
			if t.connected {
				return t, t.openProjectModal()
			}
			return t, nil
			
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
			
		case "r":
			// Refresh pod list
			if t.connected && t.focusedPanel == 0 {
				return t, t.loadPods()
			}
			return t, nil
			
		case "j":
			if t.focusedPanel == 0 && len(t.pods) > 0 {
				// Move selection down in pod list
				t.selectedPod = (t.selectedPod + 1) % len(t.pods)
				t.updatePodDisplay()
			} else if t.focusedPanel == 0 && t.showLogs {
				// Move focus down to logs panel
				t.focusedPanel = 2
			} else if t.focusedPanel == 1 && t.showLogs {
				// Move focus from details to logs
				t.focusedPanel = 2
			}
			return t, nil
			
		case "k":
			if t.focusedPanel == 0 && len(t.pods) > 0 {
				// Move selection up in pod list
				t.selectedPod = t.selectedPod - 1
				if t.selectedPod < 0 {
					t.selectedPod = len(t.pods) - 1
				}
				t.updatePodDisplay()
			} else if t.focusedPanel == 2 {
				// Move focus up from logs to main panel
				t.focusedPanel = 0
			}
			return t, nil
			
		case "down":
			// Panel navigation
			if t.focusedPanel == 0 && t.showLogs {
				t.focusedPanel = 2
			} else if t.focusedPanel == 1 && t.showLogs {
				t.focusedPanel = 2
			}
			return t, nil
			
		case "up":
			// Panel navigation
			if t.focusedPanel == 2 {
				t.focusedPanel = 0
			}
			return t, nil
		}
		
	case messages.InitMsg:
		t.ClearLoading()
		t.State = models.StateMain
		// Don't overwrite log entries from Init()
		// Just update main content for current tab
		t.updateMainContent()
		logging.Info(t.Logger, "Application initialized successfully")
		
	case messages.ConnectionSuccess:
		t.connected = true
		t.connecting = false
		t.connectionErr = nil
		t.context = msg.Context
		t.namespace = msg.Namespace
		t.logContent = append(t.logContent, fmt.Sprintf("✅ Connected to %s", msg.Context))
		
		// Initialize project manager after successful connection
		t.initializeProjectManager()
		
		// Load pods automatically after connection
		return t, t.loadPods()
		
	case messages.ConnectionError:
		t.connected = false
		t.connecting = false
		t.connectionErr = msg.Err
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Connection failed: %v", msg.Err))
		t.updatePodDisplay()
		
	case messages.PodsLoaded:
		t.pods = msg.Pods
		t.loadingPods = false
		t.selectedPod = 0
		t.updatePodDisplay()
		t.logContent = append(t.logContent, fmt.Sprintf("Loaded %d pods from namespace %s", len(msg.Pods), t.namespace))
		
	case messages.LoadPodsError:
		t.loadingPods = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load pods: %v", msg.Err))
		t.updatePodDisplay()
		
	case messages.NoKubeconfigMsg:
		t.logContent = append(t.logContent, fmt.Sprintf("⚠️  %s", msg.Message))
		t.logContent = append(t.logContent, "💡 To connect: Run 'oc login' or use --kubeconfig flag")
		t.updateMainContent()
		
	case messages.ConnectingMsg:
		t.connecting = true
		t.logContent = append(t.logContent, fmt.Sprintf("Found kubeconfig at: %s", msg.KubeconfigPath))
		t.logContent = append(t.logContent, "Connecting to cluster...")
		
	case ProjectListLoadedMsg:
		t.loadingProjects = false
		t.projectList = msg.Projects
		t.selectedProject = 0
		// Find current project index
		for i, proj := range t.projectList {
			if t.currentProject != nil && proj.Name == t.currentProject.Name {
				t.selectedProject = i
				break
			}
		}
		
	case ProjectSwitchedMsg:
		t.showProjectModal = false
		t.switchingProject = false
		t.projectError = "" // Clear any errors on successful switch
		t.currentProject = &msg.Project
		t.namespace = msg.Project.Name
		t.logContent = append(t.logContent, fmt.Sprintf("Switched to %s '%s'", msg.Project.Type, msg.Project.Name))
		// Update main content to ensure tabs are visible
		t.updateMainContent()
		// Reload pods for the new project
		if t.connected {
			return t, t.loadPods()
		}
		
	case ProjectErrorMsg:
		t.loadingProjects = false
		t.switchingProject = false
		t.projectError = msg.Error
		t.logContent = append(t.logContent, fmt.Sprintf("Project error: %s", msg.Error))
		// Keep modal open to show error
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
	
	// Show project modal if active
	if t.showProjectModal {
		return t.renderProjectModal()
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
		// Single line header - show connection status inline
		title := fmt.Sprintf("🚀 LazyOC v%s", t.Version)
		var status string
		if t.connecting {
			status = " - ⟳ Connecting..."
		} else if t.connected {
			status = fmt.Sprintf(" - ● %s", t.context)
		} else {
			status = " - ○ Disconnected"
		}
		return headerStyle.Render(title + status)
	}
	
	// Two line header
	line1 := headerStyle.Render(fmt.Sprintf("🚀 LazyOC v%s", t.Version))
	
	// Connection status
	var statusText string
	var statusColor lipgloss.Color
	
	if t.connecting {
		statusText = "⟳ Connecting..."
		statusColor = primaryColor
	} else if t.connected {
		statusText = fmt.Sprintf("● Connected to %s (namespace: %s)", t.context, t.namespace)
		statusColor = lipgloss.Color("2") // green
	} else {
		statusText = "○ Not connected - Run 'oc login' or use --kubeconfig"
		statusColor = errorColor
	}
	
	line2 := lipgloss.NewStyle().
		Width(t.width).
		Align(lipgloss.Center).
		Foreground(statusColor).
		Render(statusText)
		
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
  j/k        Move down/up in pod list
  h/l        Previous/Next tab (in main panel)
  arrow keys Navigate tabs/list
  1/2/3      Jump to main/detail/log panel
  
Commands:
  ?          Toggle help
  ctrl+p     Switch project/namespace
  d          Toggle details panel
  L          Toggle log panel (shift+l)
  r          Refresh pod list
  t          Toggle theme
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
	if !t.connected {
		t.mainContent = fmt.Sprintf(`📦 %s

❌ Not connected to any cluster

To connect to a cluster:
1. Run 'oc login <cluster-url>' in your terminal
2. Or start LazyOC with: lazyoc --kubeconfig /path/to/config

Press 'q' to quit`, tabName)
		return
	}
	
	if t.ActiveTab == 0 { // Pods tab
		t.updatePodDisplay()
	} else {
		t.mainContent = fmt.Sprintf("📦 %s Resources\n\nComing soon...\n\nUse h/l or arrow keys to navigate tabs\nPress ? for help", tabName)
	}
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

// K8s Integration Methods

// InitializeK8sClient initializes the Kubernetes client with the given kubeconfig path
func (t *SimplifiedTUI) InitializeK8sClient(kubeconfigPath string) tea.Cmd {
	return func() tea.Msg {
		
		logging.Info(t.Logger, "🔄 Starting K8s client initialization with kubeconfig: %s", kubeconfigPath)
		
		// Create auth provider
		logging.Info(t.Logger, "📝 Creating auth provider")
		t.authProvider = auth.NewKubeconfigProvider(kubeconfigPath)
		
		// Authenticate with shorter timeout to avoid hanging
		logging.Info(t.Logger, "🔐 Starting authentication (timeout: 5s)")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		config, err := t.authProvider.Authenticate(ctx)
		if err != nil {
			logging.Error(t.Logger, "❌ Authentication failed: %v", err)
			return messages.ConnectionError{Err: fmt.Errorf("authentication failed: %w", err)}
		}
		logging.Info(t.Logger, "✅ Authentication successful")
		
		// Create clientset directly (no need for duplicate client factory)
		logging.Info(t.Logger, "🔧 Creating Kubernetes clientset")
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			logging.Error(t.Logger, "❌ Clientset creation failed: %v", err)
			return messages.ConnectionError{Err: fmt.Errorf("clientset creation failed: %w", err)}
		}
		logging.Info(t.Logger, "✅ Clientset created successfully")
		
		// Create a simple client factory for backward compatibility
		logging.Info(t.Logger, "🏭 Setting up client factory")
		k8sClient := k8s.NewClientFactory()
		k8sClient.SetClientset(clientset)
		k8sClient.SetConfig(config)
		
		// Create resource client
		logging.Info(t.Logger, "📦 Getting namespace and context info")
		namespace := t.authProvider.GetNamespace()
		clusterContext := t.authProvider.GetContext()
		logging.Info(t.Logger, "📍 Namespace: %s, Context: %s", namespace, clusterContext)
		
		logging.Info(t.Logger, "🔗 Creating resource client")
		resourceClient := resources.NewK8sResourceClient(clientset, namespace)
		
		// TODO: Re-enable connection monitor once connection issue is resolved
		// connMonitor := monitor.NewK8sConnectionMonitor(t.authProvider, resourceClient)
		var connMonitor monitor.ConnectionMonitor = nil
		
		// Test connection with a separate, shorter timeout
		logging.Info(t.Logger, "🧪 Testing connection (timeout: 3s)")
		testCtx, testCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer testCancel()
		
		err = resourceClient.TestConnection(testCtx)
		if err != nil {
			logging.Error(t.Logger, "❌ Connection test failed: %v", err)
			return messages.ConnectionError{Err: fmt.Errorf("connection test failed: %w", err)}
		}
		logging.Info(t.Logger, "✅ Connection test successful")
		
		// Store everything in the success message
		logging.Info(t.Logger, "💾 Storing connection components")
		t.k8sClient = k8sClient
		t.resourceClient = resourceClient
		t.connMonitor = connMonitor
		
		logging.Info(t.Logger, "🎉 K8s client initialization complete!")
		return messages.ConnectionSuccess{
			Context:   clusterContext,
			Namespace: namespace,
		}
	}
}

// loadPods fetches pods from the current namespace
func (t *SimplifiedTUI) loadPods() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.LoadPodsError{Err: fmt.Errorf("not connected to cluster")}
		}
		
		t.loadingPods = true
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		opts := resources.ListOptions{
			Namespace: t.namespace,
		}
		
		podList, err := t.resourceClient.ListPods(ctx, opts)
		if err != nil {
			t.loadingPods = false
			return messages.LoadPodsError{Err: err}
		}
		
		t.loadingPods = false
		return messages.PodsLoaded{Pods: podList.Items}
	}
}

// updatePodDisplay updates the main content with pod information
func (t *SimplifiedTUI) updatePodDisplay() {
	if !t.connected {
		t.mainContent = `📦 Pods

❌ Not connected to any cluster

To connect to a cluster:
1. Run 'oc login <cluster-url>' in your terminal
2. Or start LazyOC with: lazyoc --kubeconfig /path/to/config

Press 'q' to quit`
		return
	}
	
	if t.loadingPods {
		t.mainContent = "📦 Pods\n\nLoading pods..."
		return
	}
	
	if len(t.pods) == 0 {
		t.mainContent = fmt.Sprintf("📦 Pods in %s\n\nNo pods found in this namespace.", t.namespace)
		return
	}
	
	// Build pod list display
	var content strings.Builder
	content.WriteString(fmt.Sprintf("📦 Pods in %s\n\n", t.namespace))
	
	// Header
	content.WriteString("NAME                                    STATUS    READY   AGE\n")
	content.WriteString("────────────────────────────────────    ──────    ─────   ───\n")
	
	// Pod rows
	for i, pod := range t.pods {
		// Highlight selected pod
		prefix := "  "
		if i == t.selectedPod && t.focusedPanel == 0 {
			prefix = "▶ "
		}
		
		// Truncate name if too long
		name := pod.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}
		
		content.WriteString(fmt.Sprintf("%s%-38s  %-8s  %-5s   %s\n",
			prefix, name, pod.Phase, pod.Ready, pod.Age))
	}
	
	t.mainContent = content.String()
	
	// Update detail pane with selected pod info
	if t.selectedPod < len(t.pods) && t.selectedPod >= 0 {
		t.updatePodDetails(t.pods[t.selectedPod])
	}
}

// updatePodDetails updates the detail pane with pod information
func (t *SimplifiedTUI) updatePodDetails(pod resources.PodInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("📄 Pod Details: %s\n\n", pod.Name))
	
	details.WriteString(fmt.Sprintf("Namespace:  %s\n", pod.Namespace))
	details.WriteString(fmt.Sprintf("Status:     %s\n", pod.Phase))
	details.WriteString(fmt.Sprintf("Ready:      %s\n", pod.Ready))
	details.WriteString(fmt.Sprintf("Restarts:   %d\n", pod.Restarts))
	details.WriteString(fmt.Sprintf("Age:        %s\n", pod.Age))
	details.WriteString(fmt.Sprintf("Node:       %s\n", pod.Node))
	details.WriteString(fmt.Sprintf("IP:         %s\n", pod.IP))
	
	if len(pod.ContainerInfo) > 0 {
		details.WriteString("\nContainers:\n")
		for _, container := range pod.ContainerInfo {
			status := "🟢"
			if !container.Ready {
				status = "🔴"
			}
			details.WriteString(fmt.Sprintf("  %s %s (%s)\n", status, container.Name, container.State))
		}
	}
	
	t.detailContent = details.String()
}

// Project-related message types
type ProjectListLoadedMsg struct {
	Projects []projects.ProjectInfo
}

type ProjectSwitchedMsg struct {
	Project projects.ProjectInfo
}

type ProjectErrorMsg struct {
	Error string
}

// openProjectModal opens the project switching modal
func (t *SimplifiedTUI) openProjectModal() tea.Cmd {
	t.showProjectModal = true
	t.loadingProjects = true
	t.switchingProject = false
	t.projectError = "" // Clear any previous errors
	t.projectModalHeight = min(t.height-6, 15) // Leave space for borders and headers
	
	return tea.Batch(
		t.loadProjectList(),
		t.getCurrentProject(),
	)
}

// loadProjectList loads the list of available projects/namespaces
func (t *SimplifiedTUI) loadProjectList() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if t.projectManager == nil {
			return ProjectErrorMsg{Error: "Project manager not initialized"}
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		projectList, err := t.projectManager.List(ctx, projects.ListOptions{
			IncludeQuotas: false, // Don't load quotas for the list view
			IncludeLimits: false,
		})
		if err != nil {
			return ProjectErrorMsg{Error: fmt.Sprintf("Failed to load projects: %v", err)}
		}
		
		return ProjectListLoadedMsg{Projects: projectList}
	})
}

// getCurrentProject loads the current project information
func (t *SimplifiedTUI) getCurrentProject() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if t.projectManager == nil {
			return nil // No error, just skip
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		current, err := t.projectManager.GetCurrent(ctx)
		if err == nil && current != nil {
			t.currentProject = current
		}
		
		return nil // We handle current project setting in the loadProjectList response
	})
}

// handleProjectModalKeys handles keyboard input when the project modal is open
func (t *SimplifiedTUI) handleProjectModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if t.loadingProjects || t.switchingProject {
		// Only allow escape while loading or switching
		if msg.String() == "esc" {
			t.showProjectModal = false
			t.loadingProjects = false
			t.switchingProject = false
			t.updateMainContent() // Ensure tabs are visible when modal closes
			return t, nil
		}
		return t, nil
	}
	
	switch msg.String() {
	case "esc":
		t.showProjectModal = false
		t.updateMainContent() // Ensure tabs are visible when modal closes
		return t, nil
		
	case "enter":
		// Switch to selected project (prevent double-switching)
		if !t.switchingProject && len(t.projectList) > 0 && t.selectedProject >= 0 && t.selectedProject < len(t.projectList) {
			t.switchingProject = true
			t.projectError = "" // Clear error when attempting a switch
			return t, t.switchToProject(t.projectList[t.selectedProject])
		}
		return t, nil
		
	case "j", "down":
		if len(t.projectList) > 0 {
			t.selectedProject = (t.selectedProject + 1) % len(t.projectList)
			// Don't clear error - let user navigate while seeing the error
		}
		return t, nil
		
	case "k", "up":
		if len(t.projectList) > 0 {
			t.selectedProject = t.selectedProject - 1
			if t.selectedProject < 0 {
				t.selectedProject = len(t.projectList) - 1
			}
			// Don't clear error - let user navigate while seeing the error
		}
		return t, nil
		
	case "r":
		// Refresh project list and clear errors
		t.loadingProjects = true
		t.projectError = ""
		return t, t.loadProjectList()
	}
	
	return t, nil
}

// switchToProject switches to the specified project
func (t *SimplifiedTUI) switchToProject(project projects.ProjectInfo) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if t.projectManager == nil {
			return ProjectErrorMsg{Error: "Project manager not initialized"}
		}
		
		// Check if we're already on this project
		if t.currentProject != nil && t.currentProject.Name == project.Name {
			return ProjectSwitchedMsg{Project: project} // Just close modal, no actual switch needed
		}
		
		logging.Info(t.Logger, "🔄 Switching to %s: %s", project.Type, project.Name)
		
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Increased timeout
		defer cancel()
		
		result, err := t.projectManager.SwitchTo(ctx, project.Name)
		if err != nil {
			logging.Error(t.Logger, "❌ Failed to switch to %s '%s': %v", project.Type, project.Name, err)
			return ProjectErrorMsg{Error: fmt.Sprintf("Failed to switch to %s '%s': %v", project.Type, project.Name, err)}
		}
		
		if !result.Success {
			logging.Error(t.Logger, "❌ Project switch failed: %s", result.Message)
			return ProjectErrorMsg{Error: result.Message}
		}
		
		logging.Info(t.Logger, "✅ Successfully switched to %s: %s", project.Type, project.Name)
		
		// Return success with the project info
		if result.ProjectInfo != nil {
			return ProjectSwitchedMsg{Project: *result.ProjectInfo}
		}
		return ProjectSwitchedMsg{Project: project}
	})
}

// renderProjectModal renders the project switching modal
func (t *SimplifiedTUI) renderProjectModal() string {
	modalWidth := min(t.width-4, 60)
	modalHeight := t.projectModalHeight
	
	// Create the modal box with error styling if needed
	borderColor := lipgloss.Color("12") // Default blue
	if t.projectError != "" {
		borderColor = lipgloss.Color("9") // Red for errors
	}
	
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1).
		Align(lipgloss.Center)
	
	var content strings.Builder
	
	// Header
	if t.currentProject != nil {
		content.WriteString(fmt.Sprintf("Current %s: %s\n\n", t.currentProject.Type, t.currentProject.Name))
	} else {
		content.WriteString("Switch Project/Namespace\n\n")
	}
	
	// Show error prominently if there is one
	if t.projectError != "" {
		// Format error message for better display
		errorMsg := t.projectError
		maxErrorLen := modalWidth - 10 // Leave space for padding and borders
		
		// Truncate very long errors and add ellipsis
		if len(errorMsg) > maxErrorLen {
			errorMsg = errorMsg[:maxErrorLen-3] + "..."
		}
		
		content.WriteString("❌ Switch Failed\n\n")
		content.WriteString(fmt.Sprintf("%s\n\n", errorMsg))
		
		// Still show project list even with error so user can try another project
		if len(t.projectList) > 0 {
			content.WriteString("Select a different project:\n\n")
		}
	}
	
	if t.loadingProjects {
		content.WriteString("Loading projects...")
	} else if t.switchingProject {
		selectedProject := ""
		if len(t.projectList) > 0 && t.selectedProject >= 0 && t.selectedProject < len(t.projectList) {
			selectedProject = t.projectList[t.selectedProject].Name
		}
		content.WriteString(fmt.Sprintf("Switching to: %s\n\nPlease wait...", selectedProject))
	} else if len(t.projectList) == 0 && t.projectError == "" {
		content.WriteString("No projects found")
	} else if len(t.projectList) > 0 {
		// List projects
		maxItems := modalHeight - 6 // Account for header, footer, padding
		startIdx := max(0, t.selectedProject-maxItems/2)
		endIdx := min(len(t.projectList), startIdx+maxItems)
		
		for i := startIdx; i < endIdx; i++ {
			project := t.projectList[i]
			
			prefix := "  "
			if i == t.selectedProject {
				prefix = "▶ "
			}
			
			// Show project type icon
			typeIcon := "📦" // namespace
			if project.Type == projects.ProjectTypeOpenShiftProject {
				typeIcon = "🎯" // project
			}
			
			// Current project indicator
			currentIndicator := ""
			if t.currentProject != nil && project.Name == t.currentProject.Name {
				currentIndicator = " (current)"
			}
			
			line := fmt.Sprintf("%s%s %s%s", prefix, typeIcon, project.Name, currentIndicator)
			if project.DisplayName != "" && project.DisplayName != project.Name {
				line += fmt.Sprintf(" - %s", project.DisplayName)
			}
			
			content.WriteString(line + "\n")
		}
		
		// Show scroll indicator if needed
		if len(t.projectList) > maxItems {
			content.WriteString(fmt.Sprintf("\n[%d/%d projects]", t.selectedProject+1, len(t.projectList)))
		}
	}
	
	// Footer
	content.WriteString("\n\n")
	if t.loadingProjects {
		content.WriteString("Press 'esc' to cancel")
	} else if t.switchingProject {
		content.WriteString("Switching project... • esc: cancel")
	} else if t.projectError != "" {
		content.WriteString("↑↓/j,k: select different • enter: try selected • r: refresh • esc: cancel")
	} else {
		content.WriteString("↑↓/j,k: navigate • enter: switch • r: refresh • esc: cancel")
	}
	
	modal := modalStyle.Render(content.String())
	
	// Center the modal on screen
	return lipgloss.Place(t.width, t.height, lipgloss.Center, lipgloss.Center, modal)
}

// initializeProjectManager initializes the project manager after K8s client is ready
func (t *SimplifiedTUI) initializeProjectManager() {
	if t.k8sClient == nil {
		logging.Warn(t.Logger, "Cannot initialize project manager: K8s client not ready")
		return
	}
	
	config := t.k8sClient.GetConfig()
	if config == nil {
		logging.Error(t.Logger, "Failed to get K8s config for project manager: config is nil")
		return
	}
	
	clientset := t.k8sClient.GetClientset()
	if clientset == nil {
		logging.Error(t.Logger, "Failed to get clientset for project manager: clientset is nil")
		return
	}
	
	// Create project manager factory
	factory, err := projects.NewProjectManagerFactory(clientset, config, t.KubeconfigPath)
	if err != nil {
		logging.Error(t.Logger, "Failed to create project manager factory: %v", err)
		return
	}
	
	t.projectFactory = factory
	
	// Create the appropriate project manager
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	manager, err := factory.CreateAutoDetectManager(ctx)
	if err != nil {
		logging.Error(t.Logger, "Failed to create project manager: %v", err)
		return
	}
	
	t.projectManager = manager
	logging.Info(t.Logger, "✅ Project manager initialized for %s", manager.GetClusterType())
	
	// Load current project info
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		current, err := manager.GetCurrent(ctx)
		if err == nil && current != nil {
			t.currentProject = current
			logging.Info(t.Logger, "Current %s: %s", current.Type, current.Name)
		}
	}()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

