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
	k8sClient      k8s.Client
	resourceClient resources.ResourceClient
	connMonitor    monitor.ConnectionMonitor
	authProvider   auth.AuthProvider
	
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
		
		// Create auth provider
		t.authProvider = auth.NewKubeconfigProvider(kubeconfigPath)
		
		// Authenticate
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		config, err := t.authProvider.Authenticate(ctx)
		if err != nil {
			return messages.ConnectionError{Err: fmt.Errorf("authentication failed: %w", err)}
		}
		
		// Create client factory
		k8sClient := k8s.NewClientFactory()
		err = k8sClient.Initialize()
		if err != nil {
			return messages.ConnectionError{Err: fmt.Errorf("client initialization failed: %w", err)}
		}
		
		// Create clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return messages.ConnectionError{Err: fmt.Errorf("clientset creation failed: %w", err)}
		}
		
		// Create resource client
		namespace := t.authProvider.GetNamespace()
		clusterContext := t.authProvider.GetContext()
		resourceClient := resources.NewK8sResourceClient(clientset, namespace)
		
		// Create connection monitor
		connMonitor := monitor.NewK8sConnectionMonitor(t.authProvider, resourceClient)
		err = connMonitor.Start(context.Background())
		if err != nil {
			logging.Warn(t.Logger, "Failed to start connection monitor: %v", err)
		}
		
		// Test connection
		err = resourceClient.TestConnection(ctx)
		if err != nil {
			return messages.ConnectionError{Err: fmt.Errorf("connection test failed: %w", err)}
		}
		
		// Store everything in the success message
		t.k8sClient = k8sClient
		t.resourceClient = resourceClient
		t.connMonitor = connMonitor
		
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