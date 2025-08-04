package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	"github.com/katyella/lazyoc/internal/ui/components"
	"github.com/katyella/lazyoc/internal/ui/errors"
	"github.com/katyella/lazyoc/internal/ui/messages"
	"github.com/katyella/lazyoc/internal/ui/models"
	"k8s.io/client-go/kubernetes"

	"github.com/katyella/lazyoc/internal/constants"
)

// TUI is the main terminal user interface for LazyOC
type TUI struct {
	*models.App

	// Kubernetes client integration
	k8sClient      k8s.Client
	resourceClient resources.ResourceClient
	connMonitor    monitor.ConnectionMonitor
	authProvider   auth.AuthProvider
	projectManager projects.ProjectManager
	projectFactory *projects.DefaultProjectManagerFactory

	// Connection state
	connected      bool
	connecting     bool
	connectionErr  error
	namespace      string
	context        string
	clusterVersion string

	// Resource data
	pods        []resources.PodInfo
	selectedPod int
	loadingPods bool

	// Kubernetes resource data
	services           []resources.ServiceInfo
	selectedService    int
	loadingServices    bool
	deployments        []resources.DeploymentInfo
	selectedDeployment int
	loadingDeployments bool
	configMaps         []resources.ConfigMapInfo
	selectedConfigMap  int
	loadingConfigMaps  bool
	secrets            []resources.SecretInfo
	selectedSecret     int
	loadingSecrets     bool

	// OpenShift resource data
	buildConfigs        []resources.BuildConfigInfo
	selectedBuildConfig int
	loadingBuildConfigs bool

	imageStreams        []resources.ImageStreamInfo
	selectedImageStream int
	loadingImageStreams bool

	routes        []resources.RouteInfo
	selectedRoute int
	loadingRoutes bool

	// Pod logs data
	podLogs         []string
	loadingLogs     bool
	logScrollOffset int
	maxLogLines     int
	userScrolled    bool            // Track if user manually scrolled
	lastLogTime     string          // Track last log timestamp for streaming
	tailMode        bool            // True when auto-scrolling to new logs
	seenLogLines    map[string]bool // Track seen log lines to prevent duplicates

	// Log view mode: "app", "pod", or "service"
	logViewMode string

	// Service logs data
	serviceLogs        []string
	serviceLogPods     []resources.PodInfo
	loadingServiceLogs bool

	// Simple state instead of components
	width        int
	height       int
	ready        bool
	showHelp     bool
	focusedPanel int

	// Content
	mainContent   string
	logContent    []string
	detailContent string

	// Visibility
	showDetails bool
	showLogs    bool

	// Project switching modal
	showProjectModal   bool
	projectList        []projects.ProjectInfo
	selectedProject    int
	currentProject     *projects.ProjectInfo
	loadingProjects    bool
	switchingProject   bool
	projectModalHeight int
	projectError       string

	// Error handling and recovery
	errorDisplay    *components.ErrorDisplayComponent
	showErrorModal  bool
	retryInProgress bool
	retryCount      int
	maxRetries      int

	// Secret viewing modal
	showSecretModal   bool
	secretModalData   map[string]string
	secretModalName   string
	secretModalKeys   []string
	selectedSecretKey int
	secretMasked      bool

	// Theme
	theme string

	// Kubeconfig path
	KubeconfigPath string
}

// NewTUI creates a new TUI instance
func NewTUI(version string, debug bool) *TUI {
	app := models.NewApp(version)
	app.Debug = debug
	app.Logger = logging.SetupLogger(debug)

	logging.Info(app.Logger, "Initializing Simplified LazyOC TUI v%s", version)

	tui := &TUI{
		App:           app,
		theme:         constants.DefaultTheme,
		showDetails:   true,
		showLogs:      true,
		focusedPanel:  constants.DefaultFocusedPanel,
		mainContent:   "", // Will be set by updateMainContent
		logContent:    []string{constants.InitialLogMessage},
		detailContent: constants.DefaultDetailContent,
		namespace:     constants.DefaultNamespace,
		pods:          []resources.PodInfo{},
		selectedPod:   0,
		// Pod logs
		podLogs:      []string{},
		maxLogLines:  constants.MaxLogLines,
		logViewMode:  constants.DefaultLogViewMode,
		tailMode:     true, // Start in tail mode by default
		seenLogLines: make(map[string]bool),
		// Error handling
		errorDisplay: components.NewErrorDisplayComponent("dark"),
		maxRetries:   constants.DefaultRetryAttempts,
	}

	// Initialize main content
	tui.updateMainContent()

	return tui
}

// SetKubeconfig sets the kubeconfig path and returns a command to initialize the connection
func (t *TUI) SetKubeconfig(kubeconfigPath string) tea.Cmd {
	if kubeconfigPath == "" {
		// Try default location
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfigPath = filepath.Join(home, constants.KubeConfigDir, constants.KubeConfigFile)
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
func (t *TUI) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Basic initialization commands
	cmds = append(cmds,
		tea.WindowSize(),
		tea.Tick(constants.InitialTickDelay, func(t time.Time) tea.Msg {
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
			defaultPath := filepath.Join(home, constants.KubeConfigDir, constants.KubeConfigFile)
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
func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.ready = true
		logging.Debug(t.Logger, "Window size: %dx%d", t.width, t.height)

	case tea.MouseMsg:
		return t.handleMouseEvent(msg)

	case tea.KeyMsg:
		// Special handling for help mode
		if t.showHelp {
			if msg.String() == "?" || msg.String() == "esc" {
				t.showHelp = false
				return t, nil
			}
			return t, nil
		}

		// Special handling for error modal
		if t.showErrorModal {
			return t.handleErrorModalKeys(msg)
		}

		// Special handling for project modal
		if t.showProjectModal {
			return t.handleProjectModalKeys(msg)
		}

		// Special handling for secret modal
		if t.showSecretModal {
			return t.handleSecretModalKeys(msg)
		}

		// Normal key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return t, tea.Quit

		case "esc":
			// Close error modal if open
			if t.showErrorModal {
				t.showErrorModal = false
				return t, nil
			}
			return t, nil

		case "r":
			// Manual retry/reconnect
			if !t.connected && !t.connecting {
				return t, t.InitializeK8sClient(t.KubeconfigPath)
			}
			// Refresh resources if connected
			if t.connected {
				switch t.ActiveTab {
				case 0: // Pods
					return t, t.loadPods()
				case 1: // Services
					t.loadingServices = true
					t.services = []resources.ServiceInfo{}
					t.updateMainContent()
					return t, t.loadServices()
				case 2: // Deployments
					t.loadingDeployments = true
					t.deployments = []resources.DeploymentInfo{}
					t.updateMainContent()
					return t, t.loadDeployments()
				case 3: // ConfigMaps
					t.loadingConfigMaps = true
					t.configMaps = []resources.ConfigMapInfo{}
					t.updateMainContent()
					return t, t.loadConfigMaps()
				case 4: // Secrets
					t.loadingSecrets = true
					t.secrets = []resources.SecretInfo{}
					t.updateMainContent()
					return t, t.loadSecrets()
				case 5: // BuildConfigs
					t.loadingBuildConfigs = true
					t.buildConfigs = []resources.BuildConfigInfo{}
					t.updateMainContent()
					return t, t.loadBuildConfigs()
				case 6: // ImageStreams
					t.loadingImageStreams = true
					t.imageStreams = []resources.ImageStreamInfo{}
					t.updateMainContent()
					return t, t.loadImageStreams()
				case 7: // Routes
					t.loadingRoutes = true
					t.routes = []resources.RouteInfo{}
					t.updateMainContent()
					return t, t.loadRoutes()
				}
			}
			return t, nil

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
			t.focusedPanel = (t.focusedPanel + 1) % constants.PanelCount
			return t, nil

		case "shift+tab":
			t.focusedPanel = (t.focusedPanel + 2) % constants.PanelCount
			return t, nil

		case "d":
			t.showDetails = !t.showDetails
			return t, nil

		case "L":
			// Handle different behavior based on current tab
			if t.ActiveTab == 1 && len(t.services) > 0 { // Services tab
				// Toggle service log mode or regular log mode
				if t.logViewMode == "service" {
					t.logViewMode = "app"
					t.showLogs = !t.showLogs
				} else {
					t.logViewMode = "service"
					t.showLogs = true
					// Stream logs for the selected service
					return t, t.streamServiceLogs()
				}
			} else {
				t.showLogs = !t.showLogs
			}
			return t, nil

		case "e":
			// Show error modal if there are errors
			if t.errorDisplay.HasErrors() {
				t.showErrorModal = true
			}
			return t, nil

		case "h":
			if t.focusedPanel == 0 { // Only navigate tabs when in main panel
				t.PrevTab()
				return t, t.handleTabSwitch()
			}
			return t, nil

		case "l":
			if t.focusedPanel == 2 { // Toggle log view when in log panel
				if t.logViewMode == constants.DefaultLogViewMode {
					t.logViewMode = constants.PodLogViewMode
					// Auto-load pod logs if not loaded and we have a selected pod
					if len(t.podLogs) == 0 && len(t.pods) > 0 && t.selectedPod < len(t.pods) {
						t.clearPodLogs() // This sets loadingLogs = true
						return t, tea.Batch(t.loadPodLogs(), t.startPodLogRefreshTimer())
					}
					// Start log refresh timer even if logs are already loaded
					return t, t.startPodLogRefreshTimer()
				} else {
					t.logViewMode = constants.DefaultLogViewMode
				}
			} else if t.focusedPanel == 0 { // Navigate tabs when in main panel
				t.NextTab()
				return t, t.handleTabSwitch()
			}
			return t, nil

		case "left":
			t.PrevTab()
			return t, t.handleTabSwitch()

		case "right":
			t.NextTab()
			return t, t.handleTabSwitch()

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

		// "r" case is already handled above for retry/refresh

		case "j":
			if t.focusedPanel == 0 {
				// Handle navigation based on current tab
				switch t.ActiveTab {
				case models.TabPods:
					if len(t.pods) > 0 {
						t.selectedPod = (t.selectedPod + 1) % len(t.pods)
						t.updatePodDisplay()
						// Clear logs and load logs for newly selected pod
						t.clearPodLogs()
						if t.logViewMode == constants.PodLogViewMode {
							return t, tea.Batch(t.loadPodLogs(), t.startPodLogRefreshTimer())
						}
						return t, t.loadPodLogs()
					}
				case models.TabServices:
					if len(t.services) > 0 {
						t.selectedService = (t.selectedService + 1) % len(t.services)
						t.updateServiceDisplay()
					}
				case models.TabDeployments:
					if len(t.deployments) > 0 {
						t.selectedDeployment = (t.selectedDeployment + 1) % len(t.deployments)
						t.updateDeploymentDisplay()
					}
				case models.TabConfigMaps:
					if len(t.configMaps) > 0 {
						t.selectedConfigMap = (t.selectedConfigMap + 1) % len(t.configMaps)
						t.updateConfigMapDisplay()
					}
				case models.TabSecrets:
					if len(t.secrets) > 0 {
						t.selectedSecret = (t.selectedSecret + 1) % len(t.secrets)
						t.updateSecretDisplay()
					}
				case models.TabBuildConfigs:
					if len(t.buildConfigs) > 0 {
						t.selectedBuildConfig = (t.selectedBuildConfig + 1) % len(t.buildConfigs)
						t.updateBuildConfigDisplay()
					}
				case models.TabImageStreams:
					if len(t.imageStreams) > 0 {
						t.selectedImageStream = (t.selectedImageStream + 1) % len(t.imageStreams)
						t.updateImageStreamDisplay()
					}
				case models.TabRoutes:
					if len(t.routes) > 0 {
						t.selectedRoute = (t.selectedRoute + 1) % len(t.routes)
						t.updateRouteDisplay()
					}
				}
			} else if t.focusedPanel == 0 && t.showLogs {
				// Move focus down to logs panel
				t.focusedPanel = 2
			} else if t.focusedPanel == 1 && t.showLogs {
				// Move focus from details to logs
				t.focusedPanel = 2
			} else if t.focusedPanel == 2 && t.logViewMode == "pod" && len(t.podLogs) > 0 {
				// Scroll down in pod logs - improved bounds checking
				maxScroll := t.getMaxLogScrollOffset()
				if t.logScrollOffset < maxScroll {
					t.logScrollOffset += 1
					t.userScrolled = true
					t.tailMode = false // Disable tail mode when manually scrolling
				} else {
					// If at bottom and trying to scroll down, enable tail mode
					t.tailMode = true
					t.userScrolled = false
				}
			}
			return t, nil

		case "k":
			if t.focusedPanel == 0 {
				// Handle navigation based on current tab
				switch t.ActiveTab {
				case models.TabPods:
					if len(t.pods) > 0 {
						t.selectedPod = t.selectedPod - 1
						if t.selectedPod < 0 {
							t.selectedPod = len(t.pods) - 1
						}
						t.updatePodDisplay()
						// Clear logs and load logs for newly selected pod
						t.clearPodLogs()
						if t.logViewMode == constants.PodLogViewMode {
							return t, tea.Batch(t.loadPodLogs(), t.startPodLogRefreshTimer())
						}
						return t, t.loadPodLogs()
					}
				case models.TabServices:
					if len(t.services) > 0 {
						t.selectedService = t.selectedService - 1
						if t.selectedService < 0 {
							t.selectedService = len(t.services) - 1
						}
						t.updateServiceDisplay()
					}
				case models.TabDeployments:
					if len(t.deployments) > 0 {
						t.selectedDeployment = t.selectedDeployment - 1
						if t.selectedDeployment < 0 {
							t.selectedDeployment = len(t.deployments) - 1
						}
						t.updateDeploymentDisplay()
					}
				case models.TabConfigMaps:
					if len(t.configMaps) > 0 {
						t.selectedConfigMap = t.selectedConfigMap - 1
						if t.selectedConfigMap < 0 {
							t.selectedConfigMap = len(t.configMaps) - 1
						}
						t.updateConfigMapDisplay()
					}
				case models.TabSecrets:
					if len(t.secrets) > 0 {
						t.selectedSecret = t.selectedSecret - 1
						if t.selectedSecret < 0 {
							t.selectedSecret = len(t.secrets) - 1
						}
						t.updateSecretDisplay()
					}
				case models.TabBuildConfigs:
					if len(t.buildConfigs) > 0 {
						t.selectedBuildConfig = t.selectedBuildConfig - 1
						if t.selectedBuildConfig < 0 {
							t.selectedBuildConfig = len(t.buildConfigs) - 1
						}
						t.updateBuildConfigDisplay()
					}
				case models.TabImageStreams:
					if len(t.imageStreams) > 0 {
						t.selectedImageStream = t.selectedImageStream - 1
						if t.selectedImageStream < 0 {
							t.selectedImageStream = len(t.imageStreams) - 1
						}
						t.updateImageStreamDisplay()
					}
				case models.TabRoutes:
					if len(t.routes) > 0 {
						t.selectedRoute = t.selectedRoute - 1
						if t.selectedRoute < 0 {
							t.selectedRoute = len(t.routes) - 1
						}
						t.updateRouteDisplay()
					}
				}
			} else if t.focusedPanel == 2 && t.logViewMode == "pod" && len(t.podLogs) > 0 {
				// Scroll up in pod logs - improved bounds checking
				if t.logScrollOffset > 0 {
					t.logScrollOffset -= 1
					t.userScrolled = true
					t.tailMode = false // Disable tail mode when manually scrolling up
				}
				// Stay in log panel even at the top
			} else if t.focusedPanel == 2 {
				// Stay in log panel for app logs too
				// Don't change focus - explicitly do nothing
				_ = 0 // Explicitly do nothing
			}
			return t, nil

		case "pgup":
			if t.focusedPanel == 2 && t.logViewMode == "pod" && len(t.podLogs) > 0 {
				// Page up in pod logs
				pageSize := t.getLogPageSize()
				t.logScrollOffset = max(0, t.logScrollOffset-pageSize)
				t.userScrolled = true
				t.tailMode = false // Disable tail mode when paging up
			}
			return t, nil

		case "pgdn":
			if t.focusedPanel == 2 && t.logViewMode == "pod" && len(t.podLogs) > 0 {
				// Page down in pod logs
				pageSize := t.getLogPageSize()
				maxScroll := t.getMaxLogScrollOffset()
				newOffset := min(maxScroll, t.logScrollOffset+pageSize)
				t.logScrollOffset = newOffset
				t.userScrolled = true
				// If we reached the bottom, enable tail mode
				if t.logScrollOffset >= maxScroll {
					t.tailMode = true
					t.userScrolled = false
				} else {
					t.tailMode = false
				}
			}
			return t, nil

		case "home":
			if t.focusedPanel == 2 && t.logViewMode == "pod" && len(t.podLogs) > 0 {
				// Go to top of pod logs
				t.logScrollOffset = 0
				t.userScrolled = true
				t.tailMode = false // Disable tail mode when going to top
			}
			return t, nil

		case "end":
			if t.focusedPanel == 2 && t.logViewMode == "pod" && len(t.podLogs) > 0 {
				// Go to bottom of pod logs and enable tail mode
				t.logScrollOffset = t.getMaxLogScrollOffset()
				t.userScrolled = false
				t.tailMode = true // Enable tail mode when going to bottom
			}
			return t, nil

		case "shift+t", "T":
			if t.focusedPanel == 2 && t.logViewMode == "pod" {
				// Toggle tail mode
				t.tailMode = !t.tailMode
				if t.tailMode {
					// If enabling tail mode, scroll to bottom
					t.logScrollOffset = t.getMaxLogScrollOffset()
					t.userScrolled = false
				}
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

		case "enter":
			// Handle Enter key for different tabs
			if t.focusedPanel == 0 {
				switch t.ActiveTab {
				case models.TabSecrets:
					if len(t.secrets) > 0 && t.selectedSecret < len(t.secrets) {
						return t, t.loadSecretData()
					}
				}
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

		// Reset retry counters on successful connection
		if t.retryCount > 0 {
			t.logContent = append(t.logContent, fmt.Sprintf("✨ Connection restored after %d retries", t.retryCount))
			t.retryCount = 0
		} else {
			t.logContent = append(t.logContent, fmt.Sprintf("✅ Connected to %s", msg.Context))
		}
		t.retryInProgress = false

		// Initialize project manager after successful connection
		t.initializeProjectManager()

		// Load cluster version information and pods
		return t, tea.Batch(
			t.loadClusterInfo(),
			t.loadPods(),
			t.startPodRefreshTimer(),
			t.startPodLogRefreshTimer(),
			t.startSpinnerAnimation(),
		)

	case messages.ConnectionError:
		t.connected = false
		t.connecting = false
		t.connectionErr = msg.Err
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Connection failed: %v", msg.Err))
		t.updatePodDisplay()

	case messages.PodsLoaded:
		// Store the previously selected pod name to preserve selection during refresh
		var previouslySelectedPodName string
		if len(t.pods) > 0 && t.selectedPod < len(t.pods) {
			previouslySelectedPodName = t.pods[t.selectedPod].Name
		}

		t.pods = msg.Pods
		t.loadingPods = false

		// Try to preserve the selected pod after refresh
		newSelectedPod := 0
		if previouslySelectedPodName != "" {
			for i, pod := range msg.Pods {
				if pod.Name == previouslySelectedPodName {
					newSelectedPod = i
					break
				}
			}
		}
		t.selectedPod = newSelectedPod

		// Only clear pod logs if we switched to a different pod or there's no previous selection
		if previouslySelectedPodName == "" || (len(msg.Pods) > 0 && newSelectedPod < len(msg.Pods) && msg.Pods[newSelectedPod].Name != previouslySelectedPodName) {
			t.podLogs = []string{}
			t.logScrollOffset = 0
			t.loadingLogs = false
		}

		t.updatePodDisplay()
		t.logContent = append(t.logContent, fmt.Sprintf("Loaded %d pods from namespace %s", len(msg.Pods), t.namespace))

	case messages.LoadPodsError:
		t.loadingPods = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load pods: %v", msg.Err))
		t.updatePodDisplay()

	// Kubernetes resource message handlers
	case messages.ServicesLoaded:
		// Store the previously selected service name to preserve selection during refresh
		var previouslySelectedServiceName string
		if len(t.services) > 0 && t.selectedService < len(t.services) {
			previouslySelectedServiceName = t.services[t.selectedService].Name
		}
		t.services = msg.Services
		t.loadingServices = false
		// Try to preserve the selected service after refresh
		newSelectedService := 0
		if previouslySelectedServiceName != "" {
			for i, svc := range msg.Services {
				if svc.Name == previouslySelectedServiceName {
					newSelectedService = i
					break
				}
			}
		}
		t.selectedService = newSelectedService
		t.updateServiceDisplay()
		t.logContent = append(t.logContent, fmt.Sprintf("Loaded %d services from namespace %s", len(msg.Services), t.namespace))
	case messages.ServicesLoadError:
		t.loadingServices = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load services: %v", msg.Err))
		t.updateServiceDisplay()
	case messages.DeploymentsLoaded:
		// Store the previously selected deployment name to preserve selection during refresh
		var previouslySelectedDeploymentName string
		if len(t.deployments) > 0 && t.selectedDeployment < len(t.deployments) {
			previouslySelectedDeploymentName = t.deployments[t.selectedDeployment].Name
		}
		t.deployments = msg.Deployments
		t.loadingDeployments = false
		// Try to preserve the selected deployment after refresh
		newSelectedDeployment := 0
		if previouslySelectedDeploymentName != "" {
			for i, deploy := range msg.Deployments {
				if deploy.Name == previouslySelectedDeploymentName {
					newSelectedDeployment = i
					break
				}
			}
		}
		t.selectedDeployment = newSelectedDeployment
		t.updateDeploymentDisplay()
		t.logContent = append(t.logContent, fmt.Sprintf("Loaded %d deployments from namespace %s", len(msg.Deployments), t.namespace))
	case messages.DeploymentsLoadError:
		t.loadingDeployments = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load deployments: %v", msg.Err))
		t.updateDeploymentDisplay()
	case messages.ConfigMapsLoaded:
		// Store the previously selected configmap name to preserve selection during refresh
		var previouslySelectedConfigMapName string
		if len(t.configMaps) > 0 && t.selectedConfigMap < len(t.configMaps) {
			previouslySelectedConfigMapName = t.configMaps[t.selectedConfigMap].Name
		}
		t.configMaps = msg.ConfigMaps
		t.loadingConfigMaps = false
		// Try to preserve the selected configmap after refresh
		newSelectedConfigMap := 0
		if previouslySelectedConfigMapName != "" {
			for i, cm := range msg.ConfigMaps {
				if cm.Name == previouslySelectedConfigMapName {
					newSelectedConfigMap = i
					break
				}
			}
		}
		t.selectedConfigMap = newSelectedConfigMap
		t.updateConfigMapDisplay()
		t.logContent = append(t.logContent, fmt.Sprintf("Loaded %d configmaps from namespace %s", len(msg.ConfigMaps), t.namespace))
	case messages.ConfigMapsLoadError:
		t.loadingConfigMaps = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load configmaps: %v", msg.Err))
		t.updateConfigMapDisplay()
	case messages.SecretsLoaded:
		// Store the previously selected secret name to preserve selection during refresh
		var previouslySelectedSecretName string
		if len(t.secrets) > 0 && t.selectedSecret < len(t.secrets) {
			previouslySelectedSecretName = t.secrets[t.selectedSecret].Name
		}
		t.secrets = msg.Secrets
		t.loadingSecrets = false
		// Try to preserve the selected secret after refresh
		newSelectedSecret := 0
		if previouslySelectedSecretName != "" {
			for i, secret := range msg.Secrets {
				if secret.Name == previouslySelectedSecretName {
					newSelectedSecret = i
					break
				}
			}
		}
		t.selectedSecret = newSelectedSecret
		t.updateSecretDisplay()
		t.logContent = append(t.logContent, fmt.Sprintf("Loaded %d secrets from namespace %s", len(msg.Secrets), t.namespace))
	case messages.SecretsLoadError:
		t.loadingSecrets = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load secrets: %v", msg.Err))
		t.updateSecretDisplay()

	// OpenShift resource message handlers
	case messages.BuildConfigsLoaded:
		t.buildConfigs = msg.BuildConfigs
		t.loadingBuildConfigs = false
		t.updateMainContent()

	case messages.BuildConfigsLoadError:
		t.buildConfigs = []resources.BuildConfigInfo{}
		t.loadingBuildConfigs = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load BuildConfigs: %v", msg.Err))
		t.updateMainContent()

	case messages.ImageStreamsLoaded:
		t.imageStreams = msg.ImageStreams
		t.loadingImageStreams = false
		t.updateMainContent()

	case messages.ImageStreamsLoadError:
		t.imageStreams = []resources.ImageStreamInfo{}
		t.loadingImageStreams = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load ImageStreams: %v", msg.Err))
		t.updateMainContent()

	case messages.RoutesLoaded:
		t.routes = msg.Routes
		t.loadingRoutes = false
		t.updateMainContent()

	case messages.RoutesLoadError:
		t.routes = []resources.RouteInfo{}
		t.loadingRoutes = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load Routes: %v", msg.Err))
		t.updateMainContent()

	case messages.ServiceLogsLoaded:
		t.serviceLogs = msg.Logs
		t.serviceLogPods = msg.Pods
		t.loadingServiceLogs = false
		t.logContent = t.serviceLogs
		t.logScrollOffset = 0 // Reset scroll to top
		t.userScrolled = false

	case messages.ServiceLogsLoadError:
		t.serviceLogs = []string{}
		t.serviceLogPods = []resources.PodInfo{}
		t.loadingServiceLogs = false
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load service logs: %v", msg.Err))

	case messages.SecretDataLoaded:
		t.secretModalData = msg.Data
		t.secretModalName = msg.SecretName
		t.secretModalKeys = msg.Keys
		t.selectedSecretKey = 0
		t.secretMasked = true // Start with masked view for security
		t.showSecretModal = true

	case messages.SecretDataLoadError:
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load secret data: %v", msg.Err))

	case messages.RefreshPods:
		// Automatically refresh pods and set up next refresh
		if t.connected && t.ActiveTab == 0 {
			return t, tea.Batch(t.loadPods(), t.startPodRefreshTimer())
		}
		return t, t.startPodRefreshTimer()

	case messages.RefreshPodLogs:
		// Automatically refresh pod logs and set up next refresh
		if t.connected && t.logViewMode == constants.PodLogViewMode && len(t.pods) > 0 && t.selectedPod < len(t.pods) {
			return t, tea.Batch(t.loadPodLogsInternal(true), t.startPodLogRefreshTimer())
		}
		return t, t.startPodLogRefreshTimer()

	case messages.NoKubeconfigMsg:
		t.logContent = append(t.logContent, fmt.Sprintf("⚠️  %s", msg.Message))
		t.logContent = append(t.logContent, "💡 To connect: Run 'oc login' or use --kubeconfig flag")
		t.updateMainContent()

	case messages.ConnectingMsg:
		t.connecting = true
		t.logContent = append(t.logContent, fmt.Sprintf("Found kubeconfig at: %s", msg.KubeconfigPath))
		t.logContent = append(t.logContent, "🔄 Connecting to cluster... (you should see spinner in status bar)")
		// Start spinner animation immediately
		return t, t.startSpinnerAnimation()

	case messages.ClusterInfoLoaded:
		t.clusterVersion = msg.Version
		// Only log if we have a real version (not error messages)
		if msg.Version != "" && !strings.Contains(msg.Version, "restricted") && !strings.Contains(msg.Version, "not available") {
			t.logContent = append(t.logContent, fmt.Sprintf("📊 Cluster version: %s", msg.Version))
		}

	case messages.ClusterInfoError:
		t.logContent = append(t.logContent, fmt.Sprintf("⚠️ Failed to load cluster info: %v", msg.Err))

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
		// Clear pod logs when switching projects
		t.clearPodLogs()
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

		// Create user-friendly error for project issues
		projectError := errors.NewUserFriendlyError(
			"Project Error",
			msg.Error,
			errors.ErrorSeverityWarning,
			errors.ErrorCategoryProject,
			nil,
		)
		t.errorDisplay.AddError(projectError)
		t.logContent = append(t.logContent, fmt.Sprintf("Project error: %s", msg.Error))
		// Keep modal open to show error

	case messages.SpinnerTick:
		// Continue spinner animation if we have active loading operations
		if t.connecting || t.loadingPods || t.loadingProjects || t.switchingProject {
			return t, t.startSpinnerAnimation()
		}

	case AutoRetryMsg:
		// Automatic retry for connection errors
		if !t.connected && !t.connecting && t.retryCount <= t.maxRetries {
			t.logContent = append(t.logContent, fmt.Sprintf("🔄 Attempting reconnection (attempt %d/%d)...", t.retryCount, t.maxRetries))
			return t, t.InitializeK8sClient(t.KubeconfigPath)
		}

	case RetrySuccessMsg:
		// Reset retry counter on successful connection
		t.retryCount = 0
		t.retryInProgress = false
		t.logContent = append(t.logContent, "✨ Connection restored successfully")

	case ManualRetryMsg:
		// Manual retry triggered by user
		t.retryInProgress = true
		if !t.connected {
			t.logContent = append(t.logContent, "🔄 Manual reconnection attempt...")
			return t, t.InitializeK8sClient(t.KubeconfigPath)
		}

	case PodLogsLoaded:
		// Pod logs successfully loaded (initial load)
		t.loadingLogs = false
		t.podLogs = msg.Logs

		// Extract and store the timestamp from the last log line for future streaming
		if len(t.podLogs) > 0 {
			lastLog := t.podLogs[len(t.podLogs)-1]
			t.lastLogTime = t.extractTimestampFromLogLine(lastLog)
		}

		// Limit log lines to prevent memory issues
		if len(t.podLogs) > t.maxLogLines {
			t.podLogs = t.podLogs[len(t.podLogs)-t.maxLogLines:] // Keep last N lines
		}
		// Auto-scroll to bottom on initial load
		t.userScrolled = false
		t.tailMode = true
		t.logScrollOffset = t.getMaxLogScrollOffset()
		t.logContent = append(t.logContent, fmt.Sprintf("📋 Loaded %d log lines from %s", len(msg.Logs), msg.PodName))

	case PodLogsRefreshed:
		// Pod logs refreshed with new content (streaming)
		if len(msg.Logs) > 0 {
			// Append new logs to existing logs
			t.podLogs = append(t.podLogs, msg.Logs...)

			// Update timestamp from the last new log
			lastLog := msg.Logs[len(msg.Logs)-1]
			t.lastLogTime = t.extractTimestampFromLogLine(lastLog)

			// Limit total log lines to prevent memory issues
			if len(t.podLogs) > t.maxLogLines {
				t.podLogs = t.podLogs[len(t.podLogs)-t.maxLogLines:] // Keep last N lines
			}

			// Auto-scroll only if in tail mode
			if t.tailMode {
				t.logScrollOffset = t.getMaxLogScrollOffset()
			}

			t.logContent = append(t.logContent, fmt.Sprintf("📋 Added %d new log lines from %s", len(msg.Logs), msg.PodName))
		}

	case PodLogsError:
		// Pod logs loading failed
		t.loadingLogs = false
		t.podLogs = []string{fmt.Sprintf("Failed to load logs: %v", msg.Err)}
		t.logScrollOffset = 0
		// Create user-friendly error
		userError := errors.MapKubernetesError(msg.Err)
		t.errorDisplay.AddError(userError)
		t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to load logs from %s: %s", msg.PodName, userError.GetDisplayMessage()))
	}

	return t, nil
}

// View implements tea.Model
func (t *TUI) View() string {
	// Don't render until we have dimensions
	if !t.ready || t.width == 0 || t.height == 0 {
		return constants.InitializingMessage
	}

	// Show help overlay if active
	if t.showHelp {
		return t.renderHelp()
	}

	// Show project modal if active
	if t.showProjectModal {
		return t.renderProjectModal()
	}

	// Show secret modal if active
	if t.showSecretModal {
		return t.renderSecretModal()
	}

	// Render main interface
	return t.renderMain()
}

// renderMain renders the main interface using direct rendering
func (t *TUI) renderMain() string {
	var sections []string

	// Header (1-2 lines based on height)
	headerHeight := 2
	if t.height < constants.SingleLineHeaderHeightThreshold {
		headerHeight = 1
	}
	sections = append(sections, t.renderHeader(headerHeight))

	// Tabs (1 line)
	sections = append(sections, t.renderTabs())

	// Calculate remaining height strictly
	// Fixed elements: header + tabs + status bar
	fixedHeight := headerHeight + 1 + 1 // header + tabs + status
	remainingHeight := t.height - fixedHeight

	// Main content area - ensure we always render content if we have any space
	if remainingHeight > 0 {
		sections = append(sections, t.renderContent(remainingHeight))
	}

	// Status bar (1 line)
	sections = append(sections, t.renderStatusBar())

	baseView := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Ensure output doesn't exceed terminal height
	lines := strings.Count(baseView, "\n") + 1
	if lines > t.height {
		// Truncate to fit terminal
		allLines := strings.Split(baseView, "\n")
		if len(allLines) > t.height {
			baseView = strings.Join(allLines[:t.height], "\n")
		}
	}

	// Overlay error modal if showing
	if t.showErrorModal {
		t.errorDisplay.SetDimensions(t.width, t.height)
		errorModal := t.errorDisplay.RenderModal()

		// Center the modal on screen using lipgloss.Place
		return lipgloss.Place(t.width, t.height, lipgloss.Center, lipgloss.Center, errorModal)
	}

	// Show help modal
	if t.showHelp {
		return t.renderHelp()
	}

	// Show project modal
	if t.showProjectModal {
		return t.renderProjectModal()
	}

	return baseView
}

// renderHeader renders a themed header
func (t *TUI) renderHeader(height int) string {
	primaryColor, errorColor := t.getThemeColors()
	headerStyle := lipgloss.NewStyle().
		Width(t.width).
		Align(lipgloss.Center).
		Foreground(primaryColor).
		Bold(true)

	if height == 1 {
		// Single line header - show connection status and project inline
		title := fmt.Sprintf("🚀 LazyOC v%s", t.Version)
		var status string
		if t.connecting {
			status = " - " + constants.ConnectingStatus
		} else if t.connected {
			projectInfo := t.getProjectDisplayInfo()
			status = fmt.Sprintf(" - ● %s (%s)", t.context, projectInfo)
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
		statusText = constants.ConnectingStatus
		statusColor = primaryColor
	} else if t.connected {
		projectInfo := t.getProjectDisplayInfo()
		statusText = fmt.Sprintf("● Connected to %s (%s)", t.context, projectInfo)
		statusColor = lipgloss.Color("2") // green
	} else {
		statusText = constants.NotConnectedMessage
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
func (t *TUI) renderTabs() string {
	tabs := constants.ResourceTabs
	var tabViews []string

	for i, tab := range tabs {
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == int(t.ActiveTab) {
			style = style.
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("12")).
				Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("244")) // Brighter gray for inactive tabs
		}
		tabViews = append(tabViews, style.Render(tab))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
	return lipgloss.NewStyle().
		Width(t.width).
		Align(lipgloss.Center).
		Render(tabBar)
}

// Constants for visual overhead
const (
	borderOverhead    = 2 // top + bottom border
	paddingOverhead   = 2 // top + bottom padding
	logHeaderOverhead = 2 // header line + separator line
)

// renderContent renders the main content area
func (t *TUI) renderContent(availableHeight int) string {
	// Calculate dimensions
	mainWidth := t.width
	if t.showDetails {
		mainWidth = int(float64(t.width) * constants.MainPanelWidthRatio)
	}

	// Calculate log panel's total overhead
	logPanelTotalOverhead := borderOverhead + paddingOverhead + logHeaderOverhead

	logHeight := 0
	maxLogContentLines := 0
	if t.showLogs && availableHeight > constants.MinMainContentLines {
		// Reserve at least 10 lines for main content and detail panel
		minMainContentHeight := constants.MinMainContentLines

		// Calculate maximum allowed log height
		maxAllowedLogHeight := availableHeight - minMainContentHeight

		// Target log height is 1/3 of available or 15 lines, whichever is smaller
		targetLogHeight := min(int(float64(availableHeight)*constants.LogHeightRatio), constants.DefaultLogHeight)

		// Apply constraints
		logHeight = min(targetLogHeight, maxAllowedLogHeight)

		// Ensure minimum log height includes overhead
		minLogHeight := logPanelTotalOverhead + constants.MinLogContentLines // At least 2 lines of content
		if logHeight < minLogHeight {
			logHeight = 0 // Don't show logs if we can't meet minimum
		}

		// Calculate actual visible lines for log content
		if logHeight > 0 {
			maxLogContentLines = logHeight - logPanelTotalOverhead
			if maxLogContentLines < 1 {
				maxLogContentLines = 1
			}
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

		// Show logs based on current log view mode
		var logText string
		var logHeader string
		if t.logViewMode == "pod" {
			// Pod logs mode
			if t.loadingLogs {
				logText = "🔄 Loading pod logs..."
				logHeader = "Pod Logs (Loading...)"
			} else if len(t.podLogs) > 0 {
				// Calculate visible lines strictly based on maxLogContentLines
				visibleLines := maxLogContentLines
				if visibleLines < 1 {
					visibleLines = 1
				}

				start := t.logScrollOffset
				end := start + visibleLines
				if end > len(t.podLogs) {
					end = len(t.podLogs)
				}
				if start >= len(t.podLogs) {
					start = max(0, len(t.podLogs)-visibleLines)
					end = len(t.podLogs)
				}

				visibleLogs := t.podLogs[start:end]

				// Apply coloring to each log line and count actual rendered lines
				// Account for both newlines and wrapped lines
				coloredLogs := []string{}
				totalLines := 0
				logWidth := t.width - constants.LogWidthPadding // Account for borders and padding

				for _, line := range visibleLogs {
					colored := t.colorizePodLog(line)

					// Count how many actual lines this log entry will render as
					// This includes both explicit newlines and wrapped lines
					lineCount := 0
					for _, subline := range strings.Split(colored, "\n") {
						// Calculate wrapped lines for each subline
						sublineLen := len(subline)
						if sublineLen == 0 {
							lineCount++
						} else {
							lineCount += (sublineLen + logWidth - 1) / logWidth
						}
					}

					// Only add if we have room
					if totalLines+lineCount <= maxLogContentLines {
						coloredLogs = append(coloredLogs, colored)
						totalLines += lineCount
					} else if totalLines < maxLogContentLines {
						// Just skip partially visible entries to avoid complexity
						break
					} else {
						break
					}
				}
				logText = strings.Join(coloredLogs, "\n")

				if len(t.pods) > 0 && t.selectedPod < len(t.pods) {
					tailIndicator := ""
					if t.tailMode {
						tailIndicator = " [TAIL]"
					}
					logHeader = fmt.Sprintf("Pod Logs: %s%s", t.pods[t.selectedPod].Name, tailIndicator)
				} else {
					logHeader = "Pod Logs"
				}
			} else {
				// Show message when no pod logs are available
				if len(t.pods) > 0 && t.selectedPod < len(t.pods) {
					selectedPodName := t.pods[t.selectedPod].Name
					logText = fmt.Sprintf("📋 No logs loaded for pod '%s'", selectedPodName)
					logHeader = fmt.Sprintf("Pod Logs: %s (Not loaded)", selectedPodName)
				} else {
					logText = "📋 No pod selected"
					logHeader = "Pod Logs (No pod selected)"
				}
			}
		} else {
			// App logs mode
			// Get recent logs but account for multiline entries
			startIdx := max(0, len(t.logContent)-constants.LastNAppLogEntries) // Start with last 100 entries
			recentLogs := t.logContent[startIdx:]

			// Apply coloring and count actual rendered lines
			// Account for both newlines and wrapped lines
			coloredAppLogs := []string{}
			totalLines := 0
			logWidth := t.width - 6 // Account for borders and padding

			for _, line := range recentLogs {
				colored := t.colorizeAppLog(line)

				// Count how many actual lines this log entry will render as
				// This includes both explicit newlines and wrapped lines
				lineCount := 0
				for _, subline := range strings.Split(colored, "\n") {
					// Calculate wrapped lines for each subline
					sublineLen := len(subline)
					if sublineLen == 0 {
						lineCount++
					} else {
						lineCount += (sublineLen + logWidth - 1) / logWidth
					}
				}

				// Only add if we have room
				if totalLines+lineCount <= maxLogContentLines {
					coloredAppLogs = append(coloredAppLogs, colored)
					totalLines += lineCount
				} else if totalLines < maxLogContentLines {
					// Just skip partially visible entries to avoid complexity
					break
				} else {
					break
				}
			}
			logText = strings.Join(coloredAppLogs, "\n")
			logHeader = "App Logs"
		}

		// Color the header based on log type with brighter colors
		headerStyle := lipgloss.NewStyle().Bold(true)
		if t.logViewMode == "pod" {
			headerStyle = headerStyle.Foreground(lipgloss.Color("207")) // Bright magenta for pod logs
		} else {
			headerStyle = headerStyle.Foreground(lipgloss.Color("51")) // Bright cyan for app logs
		}

		coloredHeader := headerStyle.Render(logHeader)

		separatorLength := len(logHeader)
		separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", separatorLength))

		fullLogText := fmt.Sprintf("%s\n%s\n%s", coloredHeader, separator, logText)

		logPanel := logStyle.Render(fullLogText)

		return lipgloss.JoinVertical(lipgloss.Left, topSection, logPanel)
	}

	return topSection
}

// renderStatusBar renders the status bar with enhanced connection information
func (t *TUI) renderStatusBar() string {
	// Style hints with different colors
	hintsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))         // Dimmer gray
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true) // White bold

	// Add error indicator to hints if there are errors
	errorHint := ""
	if t.errorDisplay.HasErrors() {
		errorHint = fmt.Sprintf("%s errors %s ", keyStyle.Render("e"), hintsStyle.Render("•"))
	}

	hints := fmt.Sprintf("%s%s help %s %s switch %s %s project %s %s retry %s %s details %s %s logs %s %s quit",
		errorHint,
		keyStyle.Render("?"), hintsStyle.Render("•"),
		keyStyle.Render("tab"), hintsStyle.Render("•"),
		keyStyle.Render("ctrl+p"), hintsStyle.Render("•"),
		keyStyle.Render("r"), hintsStyle.Render("•"),
		keyStyle.Render("d"), hintsStyle.Render("•"),
		keyStyle.Render("L"), hintsStyle.Render("•"),
		keyStyle.Render("q"))

	// Enhanced left section with connection status
	left := t.renderConnectionStatus()

	statusStyle := lipgloss.NewStyle().
		Width(t.width).
		Background(lipgloss.Color("236")). // Darker gray background
		Foreground(lipgloss.Color("15"))   // White text

	// Enhanced middle section with project and cluster info
	middle := t.renderClusterInfo()

	// Calculate spacing for three-column layout
	leftWidth := lipgloss.Width(left)
	middleWidth := lipgloss.Width(middle)
	hintsWidth := lipgloss.Width(hints)

	// Distribute remaining space
	totalContentWidth := leftWidth + middleWidth + hintsWidth
	remainingSpace := t.width - totalContentWidth

	var status string
	if remainingSpace < 2 || t.width < constants.CompactStatusWidthThreshold {
		// Compact layout for narrow screens
		status = t.renderCompactStatus(left, middle, hints)
	} else {
		// Three-column layout with full information
		leftSpacing := remainingSpace / 2
		rightSpacing := remainingSpace - leftSpacing
		if leftSpacing < 1 {
			leftSpacing = 1
		}
		if rightSpacing < 1 {
			rightSpacing = 1
		}
		status = left + strings.Repeat(" ", leftSpacing) + middle + strings.Repeat(" ", rightSpacing) + hints
	}

	return statusStyle.Render(status)
}

// renderConnectionStatus returns the connection status indicator
func (t *TUI) renderConnectionStatus() string {
	panels := constants.PanelNames

	// Focus indicator (existing functionality)
	focusIndicator := "◆"
	if t.focusedPanel >= 0 && t.focusedPanel < len(panels) {
		focusIndicator = fmt.Sprintf("◆ %s", panels[t.focusedPanel])
	}

	focusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")). // Blue for focused panel
		Bold(true)

	// Connection status indicator
	var statusIcon, statusText string
	var statusColor lipgloss.Color

	if t.connecting {
		statusIcon = t.getLoadingSpinner()
		statusText = "Connecting"
		statusColor = lipgloss.Color("11") // Yellow
	} else if t.connected {
		// Show refresh status when loading pods
		if t.loadingPods {
			statusIcon = t.getLoadingSpinner()
			statusText = "Refreshing"
			statusColor = lipgloss.Color("11") // Yellow for refreshing
		} else {
			statusIcon = "✅"
			statusText = "Connected"
			statusColor = lipgloss.Color("10") // Green
		}
	} else if t.connectionErr != nil {
		statusIcon = "❌"
		statusText = "Failed"
		statusColor = lipgloss.Color("9") // Red
	} else {
		statusIcon = "⚪"
		statusText = "Disconnected"
		statusColor = lipgloss.Color("8") // Gray
	}

	connectionStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	connectionInfo := connectionStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText))

	// Add error indicator if there are errors
	errorIndicator := ""
	if t.errorDisplay.HasErrors() {
		latestError := t.errorDisplay.GetLatestError()
		if latestError != nil {
			errorIcon := latestError.GetIcon()
			errorIndicator = fmt.Sprintf(" • %s", errorIcon)
		}
	}

	return fmt.Sprintf("%s • %s%s", focusStyle.Render(focusIndicator), connectionInfo, errorIndicator)
}

// renderClusterInfo returns cluster and project information
func (t *TUI) renderClusterInfo() string {
	if !t.connected {
		return ""
	}

	var parts []string

	// Project/Namespace info (enhanced from existing getProjectDisplayInfo)
	if t.currentProject != nil {
		var icon string
		if t.currentProject.Type == projects.ProjectTypeOpenShiftProject {
			icon = "🎯"
		} else {
			icon = "📦"
		}

		displayName := t.currentProject.Name
		if t.currentProject.DisplayName != "" && t.currentProject.DisplayName != t.currentProject.Name {
			displayName = t.currentProject.DisplayName
		}

		parts = append(parts, fmt.Sprintf("%s %s", icon, displayName))
	} else if t.namespace != "" {
		parts = append(parts, fmt.Sprintf("📦 %s", t.namespace))
	}

	// Cluster version info (only show if we have actual version, not error messages)
	if t.clusterVersion != "" && !strings.Contains(t.clusterVersion, "restricted") && !strings.Contains(t.clusterVersion, "not available") {
		parts = append(parts, fmt.Sprintf("⚙️ %s", t.clusterVersion))
	}

	// Loading indicators for ongoing operations (project loading only - pod loading shows in connection status)
	if t.loadingProjects {
		parts = append(parts, fmt.Sprintf("%s Loading projects", t.getLoadingSpinner()))
	}

	if len(parts) == 0 {
		return ""
	}

	clusterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Cyan for cluster info
		Bold(true)

	return clusterStyle.Render(strings.Join(parts, " • "))
}

// renderCompactStatus renders a compact status for narrow screens
func (t *TUI) renderCompactStatus(left, middle, hints string) string {
	// In compact mode, prioritize connection status and essential hints
	compactHints := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Render("? help • tab switch • q quit")

	availableWidth := t.width - lipgloss.Width(left) - lipgloss.Width(compactHints) - 2

	if availableWidth > 10 && middle != "" {
		// Include truncated cluster info if there's space
		if lipgloss.Width(middle) > availableWidth {
			middle = middle[:availableWidth-3] + "..."
		}
		return left + " " + middle + strings.Repeat(" ", t.width-lipgloss.Width(left)-lipgloss.Width(middle)-lipgloss.Width(compactHints)-2) + compactHints
	}

	// Just connection status and minimal hints
	spacing := t.width - lipgloss.Width(left) - lipgloss.Width(compactHints)
	if spacing < 1 {
		spacing = 1
	}
	return left + strings.Repeat(" ", spacing) + compactHints
}

// getLoadingSpinner returns an animated loading spinner based on current time
func (t *TUI) getLoadingSpinner() string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	// Use time to create animation effect
	index := (time.Now().UnixMilli() / 100) % int64(len(spinners))
	return spinners[index]
}

// renderHelp renders a simple help overlay
func (t *TUI) renderHelp() string {
	helpText := `📖 LazyOC Help

Navigation:
  tab        Next panel
  shift+tab  Previous panel
  j/k        Move down/up in pod list OR scroll logs
  h/l        Previous/Next tab (in main panel)
  arrow keys Navigate tabs/list
  1/2/3      Jump to main/detail/log panel
  
Log Scrolling (when in log panel):
  j/k        Scroll up/down line by line
  PgUp/PgDn  Scroll up/down page by page
  Home/End   Jump to top/bottom of logs
  T          Toggle tail mode (auto-scroll to new logs)
  
Commands:
  ?          Toggle help  
  l          Toggle app/pod logs (when in log panel) OR navigate tabs
  ctrl+p     Switch project/namespace
  d          Toggle details panel
  L          Toggle log panel (shift+l)
  r          Retry connection / Refresh
  e          Show error details (when errors exist)
  t          Toggle theme
  q          Quit
  
Press ? or ESC to close`

	// Simple centered help box with better styling
	helpStyle := lipgloss.NewStyle().
		Width(constants.HelpModalWidth).
		Height(constants.HelpModalHeight). // Increased height for additional log scrolling help
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

// updateMainContent updates the main content based on the active tab
func (t *TUI) updateMainContent() {
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

	switch t.ActiveTab {
	case 0: // Pods tab
		t.updatePodDisplay()
	case 1: // Services tab
		t.updateServiceDisplay()
	case 2: // Deployments tab
		t.updateDeploymentDisplay()
	case 3: // ConfigMaps tab
		t.updateConfigMapDisplay()
	case 4: // Secrets tab
		t.updateSecretDisplay()
	case 5: // BuildConfigs tab
		t.updateBuildConfigDisplay()
	case 6: // ImageStreams tab
		t.updateImageStreamDisplay()
	case 7: // Routes tab
		t.updateRouteDisplay()
	default:
		t.mainContent = fmt.Sprintf("📦 %s Resources\n\n%s\n\nUse h/l or arrow keys to navigate tabs\nPress ? for help", tabName, constants.ComingSoonMessage)
	}
}

// getThemeColors returns primary and error colors based on current theme
func (t *TUI) getThemeColors() (lipgloss.Color, lipgloss.Color) {
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
func (t *TUI) InitializeK8sClient(kubeconfigPath string) tea.Cmd {
	return func() tea.Msg {

		logging.Info(t.Logger, "🔄 Starting K8s client initialization with kubeconfig: %s", kubeconfigPath)

		// Create auth provider
		logging.Info(t.Logger, "📝 Creating auth provider")
		t.authProvider = auth.NewKubeconfigProvider(kubeconfigPath)

		// Authenticate with shorter timeout to avoid hanging
		logging.Info(t.Logger, "🔐 Starting authentication (timeout: 5s)")
		ctx, cancel := context.WithTimeout(context.Background(), constants.AuthenticationTimeout)
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

		// Initialize OpenShift detection and clients
		logging.Info(t.Logger, "🔍 Initializing OpenShift clients if available")
		err = k8sClient.InitializeOpenShiftAfterSetup()
		if err != nil {
			logging.Warn(t.Logger, "OpenShift client initialization failed (might be regular K8s): %v", err)
		} else {
			logging.Info(t.Logger, "✅ OpenShift clients initialized successfully")
		}

		// Create resource client
		logging.Info(t.Logger, "📦 Getting namespace and context info")
		namespace := t.authProvider.GetNamespace()
		clusterContext := t.authProvider.GetContext()
		logging.Info(t.Logger, "📍 Namespace: %s, Context: %s", namespace, clusterContext)

		logging.Info(t.Logger, "🔗 Creating project-aware resource client")

		// Create resource client with project manager if possible
		var resourceClient resources.ResourceClient

		// Create project manager factory
		projectFactory, err := projects.NewProjectManagerFactory(clientset, config, kubeconfigPath)
		if err != nil {
			logging.Warn(t.Logger, "⚠️ Failed to create project manager factory, falling back to namespace-only mode: %v", err)
			// Fallback to basic resource client without project manager
			resourceClient = resources.NewK8sResourceClientWithConfig(clientset, config, namespace)
		} else {
			// Create project manager with auto-detection
			projectManager, err := projectFactory.CreateAutoDetectManager(context.Background())
			if err != nil {
				logging.Warn(t.Logger, "⚠️ Failed to create project manager, falling back to namespace-only mode: %v", err)
				// Fallback to basic resource client without project manager
				resourceClient = resources.NewK8sResourceClientWithConfig(clientset, config, namespace)
			} else {
				logging.Info(t.Logger, "✅ Project manager created successfully")
				// Create resource client with project manager integration
				resourceClient = resources.NewK8sResourceClientWithProjectManagerAndConfig(clientset, config, namespace, projectManager)
			}
		}

		// Connection monitor is not currently used - dead code
		// connMonitor := monitor.NewK8sConnectionMonitor(t.authProvider, resourceClient)
		var connMonitor monitor.ConnectionMonitor = nil

		// Test connection with a separate, shorter timeout
		logging.Info(t.Logger, "🧪 Testing connection (timeout: 3s)")
		testCtx, testCancel := context.WithTimeout(context.Background(), constants.ConnectionTestTimeout)
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
func (t *TUI) loadPods() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.LoadPodsError{Err: fmt.Errorf("not connected to cluster")}
		}

		t.loadingPods = true

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
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

// loadServices loads services from the resource client
func (t *TUI) loadServices() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.ServicesLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		t.loadingServices = true

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: t.namespace,
		}

		serviceList, err := t.resourceClient.ListServices(ctx, opts)
		if err != nil {
			t.loadingServices = false
			return messages.ServicesLoadError{Err: err}
		}

		t.loadingServices = false
		return messages.ServicesLoaded{Services: serviceList.Items}
	}
}

// loadDeployments loads deployments from the resource client
func (t *TUI) loadDeployments() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.DeploymentsLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		t.loadingDeployments = true

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: t.namespace,
		}

		deploymentList, err := t.resourceClient.ListDeployments(ctx, opts)
		if err != nil {
			t.loadingDeployments = false
			return messages.DeploymentsLoadError{Err: err}
		}

		t.loadingDeployments = false
		return messages.DeploymentsLoaded{Deployments: deploymentList.Items}
	}
}

// streamServiceLogs streams logs for all pods behind the selected service
func (t *TUI) streamServiceLogs() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.ServiceLogsLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		if len(t.services) == 0 || t.selectedService >= len(t.services) {
			return messages.ServiceLogsLoadError{Err: fmt.Errorf("no service selected")}
		}

		selectedService := t.services[t.selectedService]

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		// Get pods for the selected service
		pods, err := t.resourceClient.GetPodsForService(ctx, t.namespace, selectedService.Name)
		if err != nil {
			return messages.ServiceLogsLoadError{Err: fmt.Errorf("failed to get pods for service %s: %w", selectedService.Name, err)}
		}

		if len(pods) == 0 {
			return messages.ServiceLogsLoaded{
				ServiceName: selectedService.Name,
				Pods:        pods,
				Logs:        []string{fmt.Sprintf("No pods found for service %s", selectedService.Name)},
			}
		}

		// Start streaming logs for all pods concurrently
		logChan := make(chan string, 100)
		ctx2, cancel2 := context.WithCancel(context.Background())

		// Start goroutines for each pod
		for _, pod := range pods {
			if len(pod.ContainerInfo) > 0 {
				containerName := pod.ContainerInfo[0].Name
				go func(podName, containerName string) {
					opts := resources.LogOptions{
						Follow:    true,
						TailLines: func() *int64 { i := int64(50); return &i }(),
					}

					logStream, err := t.resourceClient.StreamPodLogs(ctx2, t.namespace, podName, containerName, opts)
					if err != nil {
						logChan <- fmt.Sprintf("[%s/%s] Error streaming logs: %v", podName, containerName, err)
						return
					}

					for logLine := range logStream {
						select {
						case <-ctx2.Done():
							return
						case logChan <- fmt.Sprintf("[%s/%s] %s", podName, containerName, logLine):
						}
					}
				}(pod.Name, containerName)
			}
		}

		// Collect initial logs
		var allLogs []string
		timeout := time.After(2 * time.Second) // Wait up to 2 seconds for initial logs
	CollectLoop:
		for {
			select {
			case logLine := <-logChan:
				allLogs = append(allLogs, logLine)
				if len(allLogs) >= 100 { // Limit initial load
					break CollectLoop
				}
			case <-timeout:
				break CollectLoop
			}
		}

		cancel2() // Stop streaming for now - we'll implement continuous streaming later

		return messages.ServiceLogsLoaded{
			ServiceName: selectedService.Name,
			Pods:        pods,
			Logs:        allLogs,
		}
	}
}

// loadServiceLogs loads logs for all pods behind the selected service
func (t *TUI) loadServiceLogs() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.ServiceLogsLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		if len(t.services) == 0 || t.selectedService >= len(t.services) {
			return messages.ServiceLogsLoadError{Err: fmt.Errorf("no service selected")}
		}

		t.loadingServiceLogs = true
		selectedService := t.services[t.selectedService]

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		// Get pods for the selected service
		pods, err := t.resourceClient.GetPodsForService(ctx, t.namespace, selectedService.Name)
		if err != nil {
			t.loadingServiceLogs = false
			return messages.ServiceLogsLoadError{Err: fmt.Errorf("failed to get pods for service %s: %w", selectedService.Name, err)}
		}

		if len(pods) == 0 {
			t.loadingServiceLogs = false
			return messages.ServiceLogsLoaded{
				ServiceName: selectedService.Name,
				Pods:        pods,
				Logs:        []string{fmt.Sprintf("No pods found for service %s", selectedService.Name)},
			}
		}

		// Collect logs from all pods
		var allLogs []string
		for _, pod := range pods {
			// Get logs from the first container of each pod
			if len(pod.ContainerInfo) > 0 {
				containerName := pod.ContainerInfo[0].Name
				opts := resources.LogOptions{
					TailLines: func() *int64 { i := int64(100); return &i }(),
				}

				logs, err := t.resourceClient.GetPodLogs(ctx, pod.Namespace, pod.Name, containerName, opts)
				if err != nil {
					allLogs = append(allLogs, fmt.Sprintf("[%s/%s] Error getting logs: %v", pod.Name, containerName, err))
				} else {
					// Split logs by lines and add pod name prefix
					logLines := strings.Split(strings.TrimSpace(logs), "\n")
					for _, line := range logLines {
						if line != "" {
							allLogs = append(allLogs, fmt.Sprintf("[%s/%s] %s", pod.Name, containerName, line))
						}
					}
				}
			}
		}

		t.loadingServiceLogs = false
		return messages.ServiceLogsLoaded{
			ServiceName: selectedService.Name,
			Pods:        pods,
			Logs:        allLogs,
		}
	}
}

// loadConfigMaps loads configmaps from the resource client
func (t *TUI) loadConfigMaps() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.ConfigMapsLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		t.loadingConfigMaps = true

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: t.namespace,
		}

		configMapList, err := t.resourceClient.ListConfigMaps(ctx, opts)
		if err != nil {
			t.loadingConfigMaps = false
			return messages.ConfigMapsLoadError{Err: err}
		}

		t.loadingConfigMaps = false
		return messages.ConfigMapsLoaded{ConfigMaps: configMapList.Items}
	}
}

// loadSecrets loads secrets from the resource client
func (t *TUI) loadSecrets() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.SecretsLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		t.loadingSecrets = true

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: t.namespace,
		}

		secretList, err := t.resourceClient.ListSecrets(ctx, opts)
		if err != nil {
			t.loadingSecrets = false
			return messages.SecretsLoadError{Err: err}
		}

		t.loadingSecrets = false
		return messages.SecretsLoaded{Secrets: secretList.Items}
	}
}

// loadSecretData loads the data for the selected secret
func (t *TUI) loadSecretData() tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil {
			return messages.SecretDataLoadError{Err: fmt.Errorf("not connected to cluster")}
		}

		if len(t.secrets) == 0 || t.selectedSecret >= len(t.secrets) {
			return messages.SecretDataLoadError{Err: fmt.Errorf("no secret selected")}
		}

		selectedSecret := t.secrets[t.selectedSecret]

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		// Get the actual secret data
		secretData, err := t.resourceClient.GetSecretData(ctx, t.namespace, selectedSecret.Name)
		if err != nil {
			return messages.SecretDataLoadError{Err: fmt.Errorf("failed to get secret data %s: %w", selectedSecret.Name, err)}
		}

		// Get keys for navigation
		var keys []string
		for key := range secretData {
			keys = append(keys, key)
		}

		return messages.SecretDataLoaded{
			SecretName: selectedSecret.Name,
			Data:       secretData,
			Keys:       keys,
		}
	}
}

// startPodRefreshTimer returns a command that sets up automatic pod refresh
func (t *TUI) startPodRefreshTimer() tea.Cmd {
	return tea.Tick(constants.PodRefreshInterval, func(time.Time) tea.Msg {
		return messages.RefreshPods{}
	})
}

// startPodLogRefreshTimer returns a command that sets up automatic pod log refresh
func (t *TUI) startPodLogRefreshTimer() tea.Cmd {
	return tea.Tick(constants.PodLogRefreshInterval, func(time.Time) tea.Msg {
		return messages.RefreshPodLogs{}
	})
}

// startSpinnerAnimation returns a command that triggers spinner animation updates
func (t *TUI) startSpinnerAnimation() tea.Cmd {
	return tea.Tick(constants.SpinnerAnimationInterval, func(time.Time) tea.Msg {
		return messages.SpinnerTick{}
	})
}

// loadClusterInfo fetches cluster version and server information
func (t *TUI) loadClusterInfo() tea.Cmd {
	return func() tea.Msg {
		// Debug: Always send this message first
		return messages.ClusterInfoLoaded{
			Version:    "OpenShift (version API restricted)",
			ServerInfo: map[string]interface{}{"debug": "cluster info called"},
		}
	}
}

// clearPodLogs clears the current pod logs and sets loading state
func (t *TUI) clearPodLogs() {
	t.podLogs = []string{}
	t.logScrollOffset = 0
	t.loadingLogs = true
	t.userScrolled = false                 // Reset scroll tracking
	t.lastLogTime = ""                     // Reset timestamp tracking
	t.tailMode = true                      // Reset to tail mode
	t.seenLogLines = make(map[string]bool) // Clear seen logs map
}

// loadPodLogs fetches logs from the currently selected pod
func (t *TUI) loadPodLogs() tea.Cmd {
	return t.loadPodLogsInternal(false)
}

// loadPodLogsInternal fetches logs with option for streaming (append mode)
func (t *TUI) loadPodLogsInternal(isRefresh bool) tea.Cmd {
	return func() tea.Msg {
		if !t.connected || t.resourceClient == nil || len(t.pods) == 0 || t.selectedPod >= len(t.pods) {
			return PodLogsError{Err: fmt.Errorf("no pod selected or not connected"), PodName: ""}
		}

		selectedPod := t.pods[t.selectedPod]
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		// Get current project/namespace
		namespace := t.resourceClient.GetCurrentProject()
		if namespace == "" {
			namespace = t.namespace
		}

		// Use first container if available
		containerName := ""
		if len(selectedPod.ContainerInfo) > 0 {
			containerName = selectedPod.ContainerInfo[0].Name
		}

		// Set up log options based on whether this is initial load or refresh
		var logOpts resources.LogOptions
		if isRefresh && t.lastLogTime != "" {
			// For refresh, get logs since last timestamp using SinceSeconds
			// Parse last timestamp to calculate seconds ago
			sinceSeconds := int64(30) // Default to 30 seconds if parsing fails
			if lastTime, err := t.parseLogTimestamp(t.lastLogTime); err == nil {
				secondsAgo := int64(time.Since(lastTime).Seconds())
				if secondsAgo > 0 {
					sinceSeconds = secondsAgo + 1 // Add 1 second buffer to avoid duplicates
				}
			}

			logOpts = resources.LogOptions{
				SinceSeconds: &sinceSeconds,
				Timestamps:   true,
			}
		} else {
			// Initial load - get last 100 lines
			tailLines := int64(constants.DefaultPodLogTailLines)
			logOpts = resources.LogOptions{
				TailLines:  &tailLines,
				Timestamps: true,
			}
		}

		// Fetch logs
		logsStr, err := t.resourceClient.GetPodLogs(ctx, namespace, selectedPod.Name, containerName, logOpts)
		if err != nil {
			return PodLogsError{Err: err, PodName: selectedPod.Name}
		}

		// Split logs into lines and deduplicate
		var logLines []string
		if logsStr != "" {
			lines := strings.Split(strings.TrimSpace(logsStr), "\n")
			for _, line := range lines {
				if line != "" {
					// For refresh mode, check for duplicates
					if isRefresh {
						if !t.seenLogLines[line] {
							logLines = append(logLines, line)
							t.seenLogLines[line] = true
						}
					} else {
						// For initial load, accept all lines and mark as seen
						logLines = append(logLines, line)
						t.seenLogLines[line] = true
					}
				}
			}
		}

		// Return appropriate message based on load type
		if isRefresh {
			return PodLogsRefreshed{Logs: logLines, PodName: selectedPod.Name}
		} else {
			if len(logLines) == 0 {
				logLines = []string{constants.NoLogsAvailableMessage}
			}
			return PodLogsLoaded{Logs: logLines, PodName: selectedPod.Name}
		}
	}
}

// parseLogTimestamp parses a timestamp from a log line
func (t *TUI) parseLogTimestamp(timestamp string) (time.Time, error) {
	// Common Kubernetes log timestamp formats
	layouts := []string{
		"2006-01-02T15:04:05.999999999Z07:00", // RFC3339 with nanoseconds
		"2006-01-02T15:04:05.999999Z07:00",    // RFC3339 with microseconds
		"2006-01-02T15:04:05.999Z07:00",       // RFC3339 with milliseconds
		"2006-01-02T15:04:05Z07:00",           // RFC3339 basic
		"2006-01-02 15:04:05.999999999",       // Space-separated with nanoseconds
		"2006-01-02 15:04:05.999999",          // Space-separated with microseconds
		"2006-01-02 15:04:05.999",             // Space-separated with milliseconds
		"2006-01-02 15:04:05",                 // Space-separated basic
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, timestamp); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}

// extractTimestampFromLogLine extracts timestamp from a log line
func (t *TUI) extractTimestampFromLogLine(logLine string) string {
	// Use the same pattern as in colorizePodLog
	timestampPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[\.\d]*[Z]?[+-]?\d{0,2}:?\d{0,2}`)
	matches := timestampPattern.FindString(logLine)
	return matches
}

// updatePodDisplay updates the main content with pod information
func (t *TUI) updatePodDisplay() {
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
		t.mainContent = constants.LoadingPodsMessage
		return
	}

	if len(t.pods) == 0 {
		// Use project-aware display for no pods message
		if t.resourceClient != nil {
			currentProject := t.resourceClient.GetCurrentProject()
			if currentProject != "" {
				t.mainContent = fmt.Sprintf("📦 Pods in %s\n\nNo pods found in this project.", currentProject)
			} else {
				t.mainContent = fmt.Sprintf("📦 Pods in %s\n\nNo pods found in this namespace.", t.namespace)
			}
		} else {
			t.mainContent = fmt.Sprintf("📦 Pods in %s\n\nNo pods found in this namespace.", t.namespace)
		}
		return
	}

	// Build pod list display
	var content strings.Builder

	// Use project-aware display if resource client supports it
	if t.resourceClient != nil {
		currentProject := t.resourceClient.GetCurrentProject()
		if currentProject != "" {
			content.WriteString(fmt.Sprintf("📦 Pods in %s\n\n", currentProject))
		} else {
			content.WriteString(fmt.Sprintf("📦 Pods in %s\n\n", t.namespace))
		}
	} else {
		content.WriteString(fmt.Sprintf("📦 Pods in %s\n\n", t.namespace))
	}

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
		if len(name) > constants.PodNameTruncateLength {
			name = name[:constants.PodNameTruncateLengthCompact] + "..."
		}

		// Add status indicator with emoji
		statusIndicator := t.getPodStatusIndicator(pod.Phase)

		content.WriteString(fmt.Sprintf("%s%-38s  %s%-7s  %-5s   %s\n",
			prefix, name, statusIndicator, pod.Phase, pod.Ready, pod.Age))
	}

	t.mainContent = content.String()

	// Update detail pane with selected pod info
	if t.selectedPod < len(t.pods) && t.selectedPod >= 0 {
		t.updatePodDetails(t.pods[t.selectedPod])
	}
}

// getPodStatusIndicator returns an emoji indicator for pod status
func (t *TUI) getPodStatusIndicator(phase string) string {
	switch phase {
	case "Running":
		return "✅"
	case "Pending":
		return "⏳"
	case "Failed":
		return "❌"
	case "Succeeded":
		return "✨"
	case "Unknown":
		return "❓"
	default:
		return "⚪"
	}
}

// updatePodDetails updates the detail pane with pod information
func (t *TUI) updatePodDetails(pod resources.PodInfo) {
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

// updateBuildConfigDetails updates the detail pane with BuildConfig information
func (t *TUI) updateBuildConfigDetails(bc resources.BuildConfigInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🔨 BuildConfig Details: %s\n\n", bc.Name))

	details.WriteString(fmt.Sprintf("Namespace:  %s\n", bc.Namespace))
	details.WriteString(fmt.Sprintf("Status:     %s\n", bc.Status))
	details.WriteString(fmt.Sprintf("Strategy:   %s\n", bc.Strategy))
	details.WriteString(fmt.Sprintf("Age:        %s\n", bc.Age))

	// Source information
	details.WriteString(fmt.Sprintf("\nSource:\n"))
	details.WriteString(fmt.Sprintf("  Type:     %s\n", bc.Source.Type))
	if bc.Source.Git != nil {
		details.WriteString(fmt.Sprintf("  Git URI:  %s\n", bc.Source.Git.URI))
		if bc.Source.Git.Ref != "" {
			details.WriteString(fmt.Sprintf("  Git Ref:  %s\n", bc.Source.Git.Ref))
		}
	}
	if bc.Source.ContextDir != "" {
		details.WriteString(fmt.Sprintf("  Context:  %s\n", bc.Source.ContextDir))
	}

	// Output information
	details.WriteString(fmt.Sprintf("\nOutput:\n"))
	details.WriteString(fmt.Sprintf("  To:       %s\n", bc.Output.To))

	// Build statistics
	details.WriteString(fmt.Sprintf("\nBuilds:\n"))
	details.WriteString(fmt.Sprintf("  Success:  %d\n", bc.SuccessBuilds))
	details.WriteString(fmt.Sprintf("  Failed:   %d\n", bc.FailedBuilds))

	if bc.LastBuild != nil {
		details.WriteString(fmt.Sprintf("\nLast Build:\n"))
		details.WriteString(fmt.Sprintf("  Status:   %s\n", bc.LastBuild.Phase))
		details.WriteString(fmt.Sprintf("  Duration: %s\n", bc.LastBuild.Duration))
	}

	t.detailContent = details.String()
}

// updateImageStreamDetails updates the detail pane with ImageStream information
func (t *TUI) updateImageStreamDetails(is resources.ImageStreamInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🖼️ ImageStream Details: %s\n\n", is.Name))

	details.WriteString(fmt.Sprintf("Namespace:    %s\n", is.Namespace))
	details.WriteString(fmt.Sprintf("Status:       %s\n", is.Status))
	details.WriteString(fmt.Sprintf("Age:          %s\n", is.Age))
	details.WriteString(fmt.Sprintf("Repository:   %s\n", is.DockerImageRepository))

	if is.PublicDockerImageRepository != "" {
		details.WriteString(fmt.Sprintf("Public Repo:  %s\n", is.PublicDockerImageRepository))
	}

	// Tags information
	details.WriteString(fmt.Sprintf("\nTags (%d):\n", len(is.Tags)))
	if len(is.Tags) > 0 {
		for _, tag := range is.Tags {
			details.WriteString(fmt.Sprintf("  • %s", tag.Name))
			if len(tag.Items) > 0 {
				details.WriteString(fmt.Sprintf(" (%d images)", len(tag.Items)))
			}
			details.WriteString("\n")
		}
	} else {
		details.WriteString("  No tags available\n")
	}

	t.detailContent = details.String()
}

// updateRouteDetails updates the detail pane with Route information
func (t *TUI) updateRouteDetails(route resources.RouteInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🛣️ Route Details: %s\n\n", route.Name))

	details.WriteString(fmt.Sprintf("Namespace:    %s\n", route.Namespace))
	details.WriteString(fmt.Sprintf("Status:       %s\n", route.Status))
	details.WriteString(fmt.Sprintf("Host:         %s\n", route.Host))
	details.WriteString(fmt.Sprintf("Age:          %s\n", route.Age))

	if route.Path != "" {
		details.WriteString(fmt.Sprintf("Path:         %s\n", route.Path))
	}

	// Service information
	details.WriteString(fmt.Sprintf("\nTarget Service:\n"))
	details.WriteString(fmt.Sprintf("  Name:       %s\n", route.Service.Name))
	details.WriteString(fmt.Sprintf("  Kind:       %s\n", route.Service.Kind))

	if route.Port != nil {
		details.WriteString(fmt.Sprintf("  Port:       %s\n", route.Port.TargetPort))
	}

	// TLS information
	if route.TLS != nil {
		details.WriteString(fmt.Sprintf("\nTLS:\n"))
		details.WriteString(fmt.Sprintf("  Termination: %s\n", route.TLS.Termination))
		if route.TLS.InsecureEdgeTerminationPolicy != "" {
			details.WriteString(fmt.Sprintf("  Insecure:    %s\n", route.TLS.InsecureEdgeTerminationPolicy))
		}
	} else {
		details.WriteString(fmt.Sprintf("\nTLS:          None\n"))
	}

	if route.WildcardPolicy != "" {
		details.WriteString(fmt.Sprintf("Wildcard:     %s\n", route.WildcardPolicy))
	}

	t.detailContent = details.String()
}

// updateServiceDetails updates the detail pane with Service information
func (t *TUI) updateServiceDetails(svc resources.ServiceInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🔗 Service Details: %s\n\n", svc.Name))

	details.WriteString(fmt.Sprintf("Namespace:    %s\n", svc.Namespace))
	details.WriteString(fmt.Sprintf("Status:       %s\n", svc.Status))
	details.WriteString(fmt.Sprintf("Type:         %s\n", svc.Type))
	details.WriteString(fmt.Sprintf("Cluster IP:   %s\n", svc.ClusterIP))
	details.WriteString(fmt.Sprintf("Age:          %s\n", svc.Age))

	if len(svc.ExternalIPs) > 0 {
		details.WriteString(fmt.Sprintf("External IPs: %s\n", strings.Join(svc.ExternalIPs, ", ")))
	}

	// Ports information
	if len(svc.Ports) > 0 {
		details.WriteString(fmt.Sprintf("\nPorts:\n"))
		for _, port := range svc.Ports {
			details.WriteString(fmt.Sprintf("  • %s\n", port))
		}
	}

	// Selector information
	if svc.Selector != "" {
		details.WriteString(fmt.Sprintf("\nSelector:     %s\n", svc.Selector))
	}

	t.detailContent = details.String()
}

// updateDeploymentDetails updates the detail pane with Deployment information
func (t *TUI) updateDeploymentDetails(deploy resources.DeploymentInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🚀 Deployment Details: %s\n\n", deploy.Name))

	details.WriteString(fmt.Sprintf("Namespace:    %s\n", deploy.Namespace))
	details.WriteString(fmt.Sprintf("Status:       %s\n", deploy.Status))
	details.WriteString(fmt.Sprintf("Strategy:     %s\n", deploy.Strategy))
	details.WriteString(fmt.Sprintf("Age:          %s\n", deploy.Age))

	// Replica information
	details.WriteString(fmt.Sprintf("\nReplicas:\n"))
	details.WriteString(fmt.Sprintf("  Desired:    %d\n", deploy.Replicas))
	details.WriteString(fmt.Sprintf("  Ready:      %d\n", deploy.ReadyReplicas))
	details.WriteString(fmt.Sprintf("  Updated:    %d\n", deploy.UpdatedReplicas))
	details.WriteString(fmt.Sprintf("  Available:  %d\n", deploy.AvailableReplicas))

	// Condition information
	if deploy.Condition != "" {
		details.WriteString(fmt.Sprintf("\nCondition:    %s\n", deploy.Condition))
	}

	t.detailContent = details.String()
}

// updateConfigMapDetails updates the detail pane with ConfigMap information
func (t *TUI) updateConfigMapDetails(cm resources.ConfigMapInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("⚙️ ConfigMap Details: %s\n\n", cm.Name))

	details.WriteString(fmt.Sprintf("Namespace:    %s\n", cm.Namespace))
	details.WriteString(fmt.Sprintf("Status:       %s\n", cm.Status))
	details.WriteString(fmt.Sprintf("Data Count:   %d\n", cm.DataCount))
	details.WriteString(fmt.Sprintf("Age:          %s\n", cm.Age))

	// Labels information
	if len(cm.Labels) > 0 {
		details.WriteString(fmt.Sprintf("\nLabels:\n"))
		for key, value := range cm.Labels {
			details.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Annotations information
	if len(cm.Annotations) > 0 {
		details.WriteString(fmt.Sprintf("\nAnnotations:\n"))
		for key, value := range cm.Annotations {
			// Truncate long annotation values
			if len(value) > 60 {
				value = value[:57] + "..."
			}
			details.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	t.detailContent = details.String()
}

// updateSecretDetails updates the detail pane with Secret information
func (t *TUI) updateSecretDetails(secret resources.SecretInfo) {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("🔐 Secret Details: %s\n\n", secret.Name))

	details.WriteString(fmt.Sprintf("Namespace:    %s\n", secret.Namespace))
	details.WriteString(fmt.Sprintf("Status:       %s\n", secret.Status))
	details.WriteString(fmt.Sprintf("Type:         %s\n", secret.Type))
	details.WriteString(fmt.Sprintf("Data Count:   %d\n", secret.DataCount))
	details.WriteString(fmt.Sprintf("Age:          %s\n", secret.Age))

	// Security notice for secrets
	details.WriteString(fmt.Sprintf("\n🔒 Security:\n"))
	details.WriteString(fmt.Sprintf("  Secret data is protected and not displayed\n"))
	details.WriteString(fmt.Sprintf("  for security reasons.\n"))

	// Labels information
	if len(secret.Labels) > 0 {
		details.WriteString(fmt.Sprintf("\nLabels:\n"))
		for key, value := range secret.Labels {
			details.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Annotations information (non-sensitive)
	if len(secret.Annotations) > 0 {
		details.WriteString(fmt.Sprintf("\nAnnotations:\n"))
		for key, value := range secret.Annotations {
			// Truncate long annotation values
			if len(value) > 60 {
				value = value[:57] + "..."
			}
			details.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
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

// AutoRetryMsg is sent to trigger automatic retry
type AutoRetryMsg struct{}

// RetrySuccessMsg is sent when a retry succeeds
type RetrySuccessMsg struct{}

// ManualRetryMsg is sent when user manually triggers retry
type ManualRetryMsg struct{}

// PodLogsLoaded is sent when pod logs are successfully loaded
type PodLogsLoaded struct {
	Logs    []string
	PodName string
}

// PodLogsRefreshed is sent when pod logs are refreshed with new content
type PodLogsRefreshed struct {
	Logs    []string
	PodName string
}

// PodLogsError is sent when pod logs loading fails
type PodLogsError struct {
	Err     error
	PodName string
}

// openProjectModal opens the project switching modal
func (t *TUI) openProjectModal() tea.Cmd {
	t.showProjectModal = true
	t.loadingProjects = true
	t.switchingProject = false
	t.projectError = ""                                                                                   // Clear any previous errors
	t.projectModalHeight = min(t.height-constants.ProjectModalMinHeight, constants.ProjectModalMaxHeight) // Leave space for borders and headers

	return tea.Batch(
		t.loadProjectList(),
		t.getCurrentProject(),
	)
}

// loadProjectList loads the list of available projects/namespaces
func (t *TUI) loadProjectList() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if t.projectManager == nil {
			return ProjectErrorMsg{Error: "Project manager not initialized"}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
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
func (t *TUI) getCurrentProject() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if t.projectManager == nil {
			return nil // No error, just skip
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.AuthenticationTimeout)
		defer cancel()

		current, err := t.projectManager.GetCurrent(ctx)
		if err == nil && current != nil {
			t.currentProject = current
		}

		return nil // We handle current project setting in the loadProjectList response
	})
}

// handleProjectModalKeys handles keyboard input when the project modal is open
func (t *TUI) handleProjectModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
func (t *TUI) switchToProject(project projects.ProjectInfo) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if t.projectManager == nil {
			return ProjectErrorMsg{Error: "Project manager not initialized"}
		}

		// Check if we're already on this project
		if t.currentProject != nil && t.currentProject.Name == project.Name {
			return ProjectSwitchedMsg{Project: project} // Just close modal, no actual switch needed
		}

		logging.Info(t.Logger, "🔄 Switching to %s: %s", project.Type, project.Name)

		ctx, cancel := context.WithTimeout(context.Background(), constants.ClusterDetectionTimeout) // Increased timeout
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
func (t *TUI) renderProjectModal() string {
	modalWidth := min(t.width-constants.ProjectModalMinWidth, constants.ProjectModalMaxWidth)
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

// renderSecretModal renders the secret data viewing modal
func (t *TUI) renderSecretModal() string {
	if t.secretModalData == nil || len(t.secretModalKeys) == 0 {
		return t.renderMain()
	}

	primaryColor, _ := t.getThemeColors()

	// Modal dimensions
	modalWidth := min(80, t.width-4)
	modalHeight := min(20, t.height-4)

	// Modal style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1).
		Width(modalWidth - 4). // Account for border and padding
		Height(modalHeight - 4)

	var content strings.Builder

	// Title
	title := fmt.Sprintf("Secret: %s", t.secretModalName)
	if t.secretMasked {
		title += " (masked - press 'm' to toggle)"
	} else {
		title += " (visible - press 'm' to mask)"
	}
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")

	// Keys and values
	maxDisplayKeys := modalHeight - 8 // Leave room for title, instructions, etc.
	startIdx := 0
	if t.selectedSecretKey >= maxDisplayKeys {
		startIdx = t.selectedSecretKey - maxDisplayKeys + 1
	}

	for i := startIdx; i < len(t.secretModalKeys) && i < startIdx+maxDisplayKeys; i++ {
		key := t.secretModalKeys[i]
		value := t.secretModalData[key]

		// Cursor indicator
		prefix := "  "
		if i == t.selectedSecretKey {
			prefix = "► "
		}

		// Display value (masked or not)
		displayValue := value
		if t.secretMasked {
			displayValue = strings.Repeat("*", min(len(value), 20))
		} else {
			// Truncate long values
			if len(displayValue) > modalWidth-20 {
				displayValue = displayValue[:modalWidth-20] + "..."
			}
		}

		line := fmt.Sprintf("%s%s: %s", prefix, key, displayValue)
		if i == t.selectedSecretKey {
			line = lipgloss.NewStyle().Background(primaryColor).Foreground(lipgloss.Color("0")).Render(line)
		}
		content.WriteString(line + "\n")
	}

	// Instructions
	content.WriteString("\n")
	content.WriteString("j/k: navigate • m: toggle mask • c: copy selected • C: copy all as JSON\n")
	content.WriteString("esc/q: close")

	modal := modalStyle.Render(content.String())
	return lipgloss.Place(t.width, t.height, lipgloss.Center, lipgloss.Center, modal)
}

// initializeProjectManager initializes the project manager after K8s client is ready
func (t *TUI) initializeProjectManager() {
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
	ctx, cancel := context.WithTimeout(context.Background(), constants.AuthenticationTimeout)
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
		ctx, cancel := context.WithTimeout(context.Background(), constants.AuthenticationTimeout)
		defer cancel()

		current, err := manager.GetCurrent(ctx)
		if err == nil && current != nil {
			t.currentProject = current
			logging.Info(t.Logger, "Current %s: %s", current.Type, current.Name)
		}
	}()
}

// getProjectDisplayInfo returns formatted project context information for display
func (t *TUI) getProjectDisplayInfo() string {
	if t.currentProject == nil {
		// Fallback to namespace info if project not loaded yet
		if t.namespace != "" {
			return fmt.Sprintf("namespace: %s", t.namespace)
		}
		return "namespace: default"
	}

	// Show project type icon and name (more compact format)
	var icon string
	if t.currentProject.Type == projects.ProjectTypeOpenShiftProject {
		icon = "🎯"
	} else {
		icon = "📦"
	}

	// Use display name if available (OpenShift projects), otherwise use name
	displayName := t.currentProject.Name
	if t.currentProject.DisplayName != "" && t.currentProject.DisplayName != t.currentProject.Name {
		displayName = t.currentProject.DisplayName
	}

	return fmt.Sprintf("%s %s", icon, displayName)
}

// handleErrorModalKeys handles keyboard input when error modal is open
func (t *TUI) handleErrorModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Close error modal
		t.showErrorModal = false
		return t, nil

	case "up":
		// Move selection up in recovery actions
		t.errorDisplay.MoveSelection(-1)
		return t, nil

	case "down":
		// Move selection down in recovery actions
		t.errorDisplay.MoveSelection(1)
		return t, nil

	case "enter":
		// Execute selected recovery action
		action := t.errorDisplay.GetSelectedAction()
		if action != nil {
			return t, t.executeRecoveryAction(action)
		}
		return t, nil

	case "c":
		// Clear all errors
		t.errorDisplay.ClearErrors()
		t.showErrorModal = false
		return t, nil
	}

	return t, nil
}

// executeRecoveryAction executes the selected recovery action
func (t *TUI) executeRecoveryAction(action *errors.RecoveryAction) tea.Cmd {
	switch action.Name {
	case "Retry Connection":
		// Close modal and trigger reconnection
		t.showErrorModal = false
		if !t.connected && !t.connecting {
			t.logContent = append(t.logContent, "🔄 Manual reconnection initiated...")
			return t.InitializeK8sClient(t.KubeconfigPath)
		}

	case "Refresh Resources":
		// Close modal and refresh current resources
		t.showErrorModal = false
		if t.connected {
			switch t.ActiveTab {
			case 0: // Pods tab
				return t.loadPods()
			default:
				return t.loadPods() // Default to pods for now
			}
		}

	case "Refresh Projects":
		// Close modal and refresh project list
		t.showErrorModal = false
		if t.projectManager != nil {
			return t.loadProjectList()
		}

	case "Refresh Application":
		// Close modal and perform full refresh
		t.showErrorModal = false
		t.errorDisplay.ClearErrors()
		t.retryCount = 0
		if t.connected {
			return tea.Batch(
				t.loadPods(),
				t.loadProjectList(),
			)
		} else {
			return t.InitializeK8sClient(t.KubeconfigPath)
		}
	}

	// Close modal by default
	t.showErrorModal = false
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getMaxLogScrollOffset returns the maximum scroll offset for logs
func (t *TUI) getMaxLogScrollOffset() int {
	if len(t.podLogs) == 0 {
		return 0
	}

	visibleLines := t.getLogPageSize()
	maxScroll := len(t.podLogs) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// getLogPageSize returns the number of visible log lines per page
func (t *TUI) getLogPageSize() int {
	if !t.showLogs {
		return 10 // fallback
	}

	// Calculate log panel height similar to renderContent
	availableHeight := t.height - 4 // header + tabs + status + margins
	logHeight := availableHeight / 3
	if logHeight < 5 {
		logHeight = 5
	}

	// Account for border and padding
	visibleLines := logHeight - 4
	if visibleLines < 1 {
		visibleLines = 1
	}
	return visibleLines
}

// Log coloring helper functions

// colorizeAppLog applies color to app log messages based on content
func (t *TUI) colorizeAppLog(logLine string) string {
	// Define brighter, more readable color styles
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Bright red + bold
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))             // Orange
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))           // Bright green
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))              // Bright blue

	// Apply colors based on content patterns
	switch {
	case strings.Contains(logLine, "❌") || strings.Contains(logLine, "Failed") || strings.Contains(logLine, "Error"):
		return errorStyle.Render(logLine)
	case strings.Contains(logLine, "⟳") || strings.Contains(logLine, "retry") || strings.Contains(logLine, "Retry"):
		return warnStyle.Render(logLine)
	case strings.Contains(logLine, "✓") || strings.Contains(logLine, "Connected") || strings.Contains(logLine, "Loaded") || strings.Contains(logLine, "Switched"):
		return successStyle.Render(logLine)
	case strings.Contains(logLine, "🔄") || strings.Contains(logLine, "●") || strings.Contains(logLine, "Loading"):
		return infoStyle.Render(logLine)
	default:
		return logLine // No coloring for neutral messages
	}
}

// colorizePodLog applies color to pod log lines based on log level patterns
func (t *TUI) colorizePodLog(logLine string) string {
	// Define brighter, more readable color styles
	timestampStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))        // Brighter gray
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Bright red + bold
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))             // Orange/yellow
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))              // Bright green
	debugStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))             // Bright blue
	noticeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))            // Cyan for notice

	// Improved log level patterns - more comprehensive
	errorPattern := regexp.MustCompile(`(?i)\b(error|fatal|err|panic|exception|fail|critical)\b`)
	warnPattern := regexp.MustCompile(`(?i)\b(warn|warning|deprecated|caution)\b`)
	infoPattern := regexp.MustCompile(`(?i)\b(info|information|starting|started|listening)\b`)
	debugPattern := regexp.MustCompile(`(?i)\b(debug|trace|verbose)\b`)
	noticePattern := regexp.MustCompile(`(?i)\b(notice|configured|loaded|compiled)\b`)

	// More flexible timestamp pattern
	timestampPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[\.\d]*`)

	// Simple approach - color the entire line based on content
	switch {
	case errorPattern.MatchString(logLine):
		return errorStyle.Render(logLine)
	case warnPattern.MatchString(logLine):
		return warnStyle.Render(logLine)
	case infoPattern.MatchString(logLine):
		return infoStyle.Render(logLine)
	case debugPattern.MatchString(logLine):
		return debugStyle.Render(logLine)
	case noticePattern.MatchString(logLine):
		return noticeStyle.Render(logLine)
	case timestampPattern.MatchString(logLine):
		// If it's mainly a timestamp line, color it with timestamp style
		return timestampStyle.Render(logLine)
	default:
		return logLine // Default color for unmatched content
	}
}

// OpenShift resource display functions

func (t *TUI) updateBuildConfigDisplay() {
	if t.loadingBuildConfigs {
		t.mainContent = "🔨 BuildConfigs\n\nLoading BuildConfigs..."
		return
	}

	if len(t.buildConfigs) == 0 {
		t.mainContent = "🔨 BuildConfigs\n\nNo BuildConfigs found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("🔨 BuildConfigs\n\n")

	// Header
	header := fmt.Sprintf("%-30s %-20s %-15s %-10s %s", "NAME", "SOURCE", "STRATEGY", "BUILDS", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 90))
	content.WriteString("\n")

	// BuildConfig rows
	for i, bc := range t.buildConfigs {
		style := lipgloss.NewStyle()
		if i == t.selectedBuildConfig {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		sourceType := "Unknown"
		if bc.Source.Git != nil {
			sourceType = "Git"
		}

		buildsInfo := fmt.Sprintf("%d/%d", bc.SuccessBuilds, bc.SuccessBuilds+bc.FailedBuilds)

		row := fmt.Sprintf("%-30s %-20s %-15s %-10s %s",
			truncateString(bc.Name, 30),
			truncateString(sourceType, 20),
			truncateString(bc.Strategy, 15),
			buildsInfo,
			bc.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' to trigger build • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected BuildConfig info
	if t.selectedBuildConfig < len(t.buildConfigs) && t.selectedBuildConfig >= 0 {
		t.updateBuildConfigDetails(t.buildConfigs[t.selectedBuildConfig])
	}
}

func (t *TUI) updateImageStreamDisplay() {
	if t.loadingImageStreams {
		t.mainContent = "🖼️ ImageStreams\n\nLoading ImageStreams..."
		return
	}

	if len(t.imageStreams) == 0 {
		t.mainContent = "🖼️ ImageStreams\n\nNo ImageStreams found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("🖼️ ImageStreams\n\n")

	// Header
	header := fmt.Sprintf("%-35s %-40s %-8s %s", "NAME", "DOCKER REPOSITORY", "TAGS", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 95))
	content.WriteString("\n")

	// ImageStream rows
	for i, is := range t.imageStreams {
		style := lipgloss.NewStyle()
		if i == t.selectedImageStream {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		tagCount := len(is.Tags)
		repo := truncateString(is.DockerImageRepository, 40)

		row := fmt.Sprintf("%-35s %-40s %-8d %s",
			truncateString(is.Name, 35),
			repo,
			tagCount,
			is.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' for tag details • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected ImageStream info
	if t.selectedImageStream < len(t.imageStreams) && t.selectedImageStream >= 0 {
		t.updateImageStreamDetails(t.imageStreams[t.selectedImageStream])
	}
}

func (t *TUI) updateRouteDisplay() {
	if t.loadingRoutes {
		t.mainContent = "🛣️ Routes\n\nLoading Routes..."
		return
	}

	if len(t.routes) == 0 {
		t.mainContent = "🛣️ Routes\n\nNo Routes found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("🛣️ Routes\n\n")

	// Header
	header := fmt.Sprintf("%-25s %-40s %-20s %-8s %s", "NAME", "HOST", "SERVICE", "TLS", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 100))
	content.WriteString("\n")

	// Route rows
	for i, route := range t.routes {
		style := lipgloss.NewStyle()
		if i == t.selectedRoute {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		tlsStatus := "None"
		if route.TLS != nil {
			tlsStatus = route.TLS.Termination
		}

		row := fmt.Sprintf("%-25s %-40s %-20s %-8s %s",
			truncateString(route.Name, 25),
			truncateString(route.Host, 40),
			truncateString(route.Service.Name, 20),
			tlsStatus,
			route.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' for details • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected Route info
	if t.selectedRoute < len(t.routes) && t.selectedRoute >= 0 {
		t.updateRouteDetails(t.routes[t.selectedRoute])
	}
}

// updateServiceDisplay updates the main content with service information
func (t *TUI) updateServiceDisplay() {
	if t.loadingServices {
		t.mainContent = "🔗 Services\n\nLoading Services..."
		return
	}

	if len(t.services) == 0 {
		t.mainContent = "🔗 Services\n\nNo Services found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("🔗 Services\n\n")

	// Header
	header := fmt.Sprintf("%-30s %-15s %-20s %-30s %s", "NAME", "TYPE", "CLUSTER-IP", "PORTS", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 100))
	content.WriteString("\n")

	// Service rows
	for i, svc := range t.services {
		style := lipgloss.NewStyle()
		if i == t.selectedService {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		ports := strings.Join(svc.Ports, ",")
		if len(ports) > 30 {
			ports = ports[:27] + "..."
		}

		row := fmt.Sprintf("%-30s %-15s %-20s %-30s %s",
			truncateString(svc.Name, 30),
			svc.Type,
			truncateString(svc.ClusterIP, 20),
			ports,
			svc.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' for details • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected Service info
	if t.selectedService < len(t.services) && t.selectedService >= 0 {
		t.updateServiceDetails(t.services[t.selectedService])
	}
}

// updateDeploymentDisplay updates the main content with deployment information
func (t *TUI) updateDeploymentDisplay() {
	if t.loadingDeployments {
		t.mainContent = "🚀 Deployments\n\nLoading Deployments..."
		return
	}

	if len(t.deployments) == 0 {
		t.mainContent = "🚀 Deployments\n\nNo Deployments found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("🚀 Deployments\n\n")

	// Header
	header := fmt.Sprintf("%-30s %-10s %-10s %-10s %-15s %s", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "STRATEGY", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 95))
	content.WriteString("\n")

	// Deployment rows
	for i, deploy := range t.deployments {
		style := lipgloss.NewStyle()
		if i == t.selectedDeployment {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		ready := fmt.Sprintf("%d/%d", deploy.ReadyReplicas, deploy.Replicas)

		row := fmt.Sprintf("%-30s %-10s %-10d %-10d %-15s %s",
			truncateString(deploy.Name, 30),
			ready,
			deploy.UpdatedReplicas,
			deploy.AvailableReplicas,
			truncateString(deploy.Strategy, 15),
			deploy.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' for details • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected Deployment info
	if t.selectedDeployment < len(t.deployments) && t.selectedDeployment >= 0 {
		t.updateDeploymentDetails(t.deployments[t.selectedDeployment])
	}
}

// updateConfigMapDisplay updates the main content with configmap information
func (t *TUI) updateConfigMapDisplay() {
	if t.loadingConfigMaps {
		t.mainContent = "⚙️ ConfigMaps\n\nLoading ConfigMaps..."
		return
	}

	if len(t.configMaps) == 0 {
		t.mainContent = "⚙️ ConfigMaps\n\nNo ConfigMaps found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("⚙️ ConfigMaps\n\n")

	// Header
	header := fmt.Sprintf("%-30s %-10s %s", "NAME", "DATA", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 50))
	content.WriteString("\n")

	// ConfigMap rows
	for i, cm := range t.configMaps {
		style := lipgloss.NewStyle()
		if i == t.selectedConfigMap {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		row := fmt.Sprintf("%-30s %-10d %s",
			truncateString(cm.Name, 30),
			cm.DataCount,
			cm.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' for details • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected ConfigMap info
	if t.selectedConfigMap < len(t.configMaps) && t.selectedConfigMap >= 0 {
		t.updateConfigMapDetails(t.configMaps[t.selectedConfigMap])
	}
}

// updateSecretDisplay updates the main content with secret information
func (t *TUI) updateSecretDisplay() {
	if t.loadingSecrets {
		t.mainContent = "🔐 Secrets\n\nLoading Secrets..."
		return
	}

	if len(t.secrets) == 0 {
		t.mainContent = "🔐 Secrets\n\nNo Secrets found in current namespace.\n\nPress 'r' to refresh"
		return
	}

	var content strings.Builder
	content.WriteString("🔐 Secrets\n\n")

	// Header
	header := fmt.Sprintf("%-30s %-20s %-10s %s", "NAME", "TYPE", "DATA", "AGE")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", 70))
	content.WriteString("\n")

	// Secret rows
	for i, secret := range t.secrets {
		style := lipgloss.NewStyle()
		if i == t.selectedSecret {
			style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
		}

		row := fmt.Sprintf("%-30s %-20s %-10d %s",
			truncateString(secret.Name, 30),
			truncateString(secret.Type, 20),
			secret.DataCount,
			secret.Age,
		)

		content.WriteString(style.Render(row))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\nUse j/k or ↑↓ to navigate • Press 'enter' for details • Press 'r' to refresh")

	t.mainContent = content.String()

	// Update detail panel with selected Secret info
	if t.selectedSecret < len(t.secrets) && t.selectedSecret >= 0 {
		t.updateSecretDetails(t.secrets[t.selectedSecret])
	}
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// handleTabSwitch handles tab switching and auto-loading
func (t *TUI) handleTabSwitch() tea.Cmd {
	t.updateMainContent()

	// Auto-load data for resource tabs if needed
	if t.connected {
		switch t.ActiveTab {
		case 1: // Services
			if len(t.services) == 0 && !t.loadingServices {
				t.loadingServices = true
				return t.loadServices()
			}
		case 2: // Deployments
			if len(t.deployments) == 0 && !t.loadingDeployments {
				t.loadingDeployments = true
				return t.loadDeployments()
			}
		case 3: // ConfigMaps
			if len(t.configMaps) == 0 && !t.loadingConfigMaps {
				t.loadingConfigMaps = true
				return t.loadConfigMaps()
			}
		case 4: // Secrets
			if len(t.secrets) == 0 && !t.loadingSecrets {
				t.loadingSecrets = true
				return t.loadSecrets()
			}
		case 5: // BuildConfigs
			if len(t.buildConfigs) == 0 && !t.loadingBuildConfigs {
				t.loadingBuildConfigs = true
				return t.loadBuildConfigs()
			}
		case 6: // ImageStreams
			if len(t.imageStreams) == 0 && !t.loadingImageStreams {
				t.loadingImageStreams = true
				return t.loadImageStreams()
			}
		case 7: // Routes
			if len(t.routes) == 0 && !t.loadingRoutes {
				t.loadingRoutes = true
				return t.loadRoutes()
			}
		}
	}

	return nil
}

// OpenShift resource loading functions

func (t *TUI) loadBuildConfigs() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Check if we have an OpenShift client
		osClient, ok := t.k8sClient.(k8s.OpenShiftClient)
		if !ok || !osClient.IsOpenShift() {
			return messages.BuildConfigsLoadError{Err: fmt.Errorf("not connected to an OpenShift cluster")}
		}

		// Create a resource client for OpenShift
		resourceClient := resources.NewOpenShiftResourceClient(osClient)

		// Load BuildConfigs
		listOpts := resources.ListOptions{
			Namespace: t.namespace,
		}

		buildConfigList, err := resourceClient.ListBuildConfigs(context.Background(), listOpts)
		if err != nil {
			return messages.BuildConfigsLoadError{Err: err}
		}

		return messages.BuildConfigsLoaded{BuildConfigs: buildConfigList.Items}
	})
}

func (t *TUI) loadImageStreams() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Check if we have an OpenShift client
		osClient, ok := t.k8sClient.(k8s.OpenShiftClient)
		if !ok || !osClient.IsOpenShift() {
			return messages.ImageStreamsLoadError{Err: fmt.Errorf("not connected to an OpenShift cluster")}
		}

		// Create a resource client for OpenShift
		resourceClient := resources.NewOpenShiftResourceClient(osClient)

		// Load ImageStreams
		listOpts := resources.ListOptions{
			Namespace: t.namespace,
		}

		imageStreamList, err := resourceClient.ListImageStreams(context.Background(), listOpts)
		if err != nil {
			return messages.ImageStreamsLoadError{Err: err}
		}

		return messages.ImageStreamsLoaded{ImageStreams: imageStreamList.Items}
	})
}

func (t *TUI) loadRoutes() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Check if we have an OpenShift client
		osClient, ok := t.k8sClient.(k8s.OpenShiftClient)
		if !ok || !osClient.IsOpenShift() {
			return messages.RoutesLoadError{Err: fmt.Errorf("not connected to an OpenShift cluster")}
		}

		// Create a resource client for OpenShift
		resourceClient := resources.NewOpenShiftResourceClient(osClient)

		// Load Routes
		listOpts := resources.ListOptions{
			Namespace: t.namespace,
		}

		routeList, err := resourceClient.ListRoutes(context.Background(), listOpts)
		if err != nil {
			return messages.RoutesLoadError{Err: err}
		}

		return messages.RoutesLoaded{Routes: routeList.Items}
	})
}

// handleSecretModalKeys handles key input for the secret modal
func (t *TUI) handleSecretModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Close secret modal
		t.showSecretModal = false
		t.secretModalData = nil
		t.secretModalKeys = nil
		t.selectedSecretKey = 0
		return t, nil

	case "j", "down":
		if len(t.secretModalKeys) > 0 {
			t.selectedSecretKey = (t.selectedSecretKey + 1) % len(t.secretModalKeys)
		}
		return t, nil

	case "k", "up":
		if len(t.secretModalKeys) > 0 {
			t.selectedSecretKey = t.selectedSecretKey - 1
			if t.selectedSecretKey < 0 {
				t.selectedSecretKey = len(t.secretModalKeys) - 1
			}
		}
		return t, nil

	case "m":
		// Toggle masking
		t.secretMasked = !t.secretMasked
		return t, nil

	case "c":
		// Copy selected secret key to clipboard
		if len(t.secretModalKeys) > 0 && t.selectedSecretKey < len(t.secretModalKeys) {
			key := t.secretModalKeys[t.selectedSecretKey]
			value := t.secretModalData[key]
			return t, t.copyToClipboard(value)
		}
		return t, nil

	case "C":
		// Copy all secret data to clipboard as JSON
		return t, t.copySecretAsJSON()
	}

	return t, nil
}

// copyToClipboard copies text to clipboard
func (t *TUI) copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		// Use different clipboard commands based on OS
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			// Try xclip first, then xsel
			if _, err := exec.LookPath("xclip"); err == nil {
				cmd = exec.Command("xclip", "-selection", "clipboard")
			} else if _, err := exec.LookPath("xsel"); err == nil {
				cmd = exec.Command("xsel", "--clipboard", "--input")
			} else {
				t.logContent = append(t.logContent, "❌ No clipboard tool found (xclip or xsel required)")
				return nil
			}
		case "darwin":
			cmd = exec.Command("pbcopy")
		case "windows":
			cmd = exec.Command("clip")
		default:
			t.logContent = append(t.logContent, "❌ Clipboard not supported on this OS")
			return nil
		}

		if cmd != nil {
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err != nil {
				t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to copy to clipboard: %v", err))
			} else {
				t.logContent = append(t.logContent, "✅ Copied to clipboard")
			}
		}
		return nil
	}
}

// copySecretAsJSON copies all secret data as JSON to clipboard
func (t *TUI) copySecretAsJSON() tea.Cmd {
	return func() tea.Msg {
		if t.secretModalData == nil {
			return nil
		}

		jsonData, err := json.MarshalIndent(t.secretModalData, "", "  ")
		if err != nil {
			t.logContent = append(t.logContent, fmt.Sprintf("❌ Failed to serialize secret as JSON: %v", err))
			return nil
		}

		return t.copyToClipboard(string(jsonData))()
	}
}

// handleMouseEvent processes mouse interactions
func (t *TUI) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Ignore mouse events in modal states
	if t.showHelp || t.showErrorModal || t.showProjectModal || t.showSecretModal {
		return t, nil
	}

	switch msg.Type {
	case tea.MouseLeft:
		return t.handleMouseClick(msg.X, msg.Y)
	case tea.MouseWheelUp:
		return t.handleMouseWheel(-1, msg.X, msg.Y)
	case tea.MouseWheelDown:
		return t.handleMouseWheel(1, msg.X, msg.Y)
	}

	return t, nil
}

// handleMouseClick processes mouse click events
func (t *TUI) handleMouseClick(x, y int) (tea.Model, tea.Cmd) {
	headerHeight := t.getHeaderHeight()
	tabY := headerHeight

	// Check if click is on tabs row
	if y == tabY {
		return t.handleTabClick(x)
	}

	// Check if click is in resource list area
	resourceListStartY := headerHeight + 1 // header + tabs
	if y > resourceListStartY {
		return t.handleResourceClick(x, y-resourceListStartY-1) // -1 for 0-based indexing
	}

	return t, nil
}

// handleTabClick processes clicks on the tab bar
func (t *TUI) handleTabClick(x int) (tea.Model, tea.Cmd) {
	tabs := constants.ResourceTabs
	totalTabs := len(tabs)

	// Calculate tab width (approximately equal distribution)
	tabWidth := t.width / totalTabs
	if tabWidth < 1 {
		tabWidth = 1
	}

	// Find which tab was clicked
	clickedTab := x / tabWidth
	if clickedTab >= totalTabs {
		clickedTab = totalTabs - 1
	}

	// Switch to the clicked tab
	t.ActiveTab = models.TabType(clickedTab)
	logging.Debug(t.Logger, "Mouse click switched to tab %d (%s)", clickedTab, tabs[clickedTab])

	return t, nil
}

// handleResourceClick processes clicks on resource list items
func (t *TUI) handleResourceClick(x, y int) (tea.Model, tea.Cmd) {
	// Only handle clicks if we're connected and have resources
	if !t.connected {
		return t, nil
	}

	switch t.ActiveTab {
	case models.TabPods: // Pods
		if y < len(t.pods) {
			t.selectedPod = y
			t.updatePodDisplay()
			logging.Debug(t.Logger, "Mouse selected pod %d", y)
		}
	case models.TabServices: // Services
		if y < len(t.services) {
			t.selectedService = y
			t.updateServiceDisplay()
			logging.Debug(t.Logger, "Mouse selected service %d", y)
		}
	case models.TabDeployments: // Deployments
		if y < len(t.deployments) {
			t.selectedDeployment = y
			t.updateDeploymentDisplay()
			logging.Debug(t.Logger, "Mouse selected deployment %d", y)
		}
	case models.TabConfigMaps: // ConfigMaps
		if y < len(t.configMaps) {
			t.selectedConfigMap = y
			t.updateConfigMapDisplay()
			logging.Debug(t.Logger, "Mouse selected configmap %d", y)
		}
	case models.TabSecrets: // Secrets
		if y < len(t.secrets) {
			t.selectedSecret = y
			t.updateSecretDisplay()
			logging.Debug(t.Logger, "Mouse selected secret %d", y)
		}
	case models.TabBuildConfigs: // BuildConfigs
		if y < len(t.buildConfigs) {
			t.selectedBuildConfig = y
			t.updateBuildConfigDisplay()
			logging.Debug(t.Logger, "Mouse selected buildconfig %d", y)
		}
	case models.TabImageStreams: // ImageStreams
		if y < len(t.imageStreams) {
			t.selectedImageStream = y
			t.updateImageStreamDisplay()
			logging.Debug(t.Logger, "Mouse selected imagestream %d", y)
		}
	case models.TabRoutes: // Routes
		if y < len(t.routes) {
			t.selectedRoute = y
			t.updateRouteDisplay()
			logging.Debug(t.Logger, "Mouse selected route %d", y)
		}
	}

	return t, nil
}

// handleMouseWheel processes mouse wheel scroll events
func (t *TUI) handleMouseWheel(direction int, x, y int) (tea.Model, tea.Cmd) {
	// Only handle wheel events if we're connected
	if !t.connected {
		return t, nil
	}

	headerHeight := t.getHeaderHeight()
	resourceListStartY := headerHeight + 1

	// Check if wheel event is in resource list area
	if y > resourceListStartY {
		return t.handleResourceListScroll(direction)
	}

	return t, nil
}

// handleResourceListScroll scrolls through resource lists
func (t *TUI) handleResourceListScroll(direction int) (tea.Model, tea.Cmd) {
	switch t.ActiveTab {
	case models.TabPods: // Pods
		if direction > 0 { // Scroll down
			if t.selectedPod < len(t.pods)-1 {
				t.selectedPod++
				t.updatePodDisplay()
			}
		} else { // Scroll up
			if t.selectedPod > 0 {
				t.selectedPod--
				t.updatePodDisplay()
			}
		}
	case models.TabServices: // Services
		if direction > 0 {
			if t.selectedService < len(t.services)-1 {
				t.selectedService++
				t.updateServiceDisplay()
			}
		} else {
			if t.selectedService > 0 {
				t.selectedService--
				t.updateServiceDisplay()
			}
		}
	case models.TabDeployments: // Deployments
		if direction > 0 {
			if t.selectedDeployment < len(t.deployments)-1 {
				t.selectedDeployment++
				t.updateDeploymentDisplay()
			}
		} else {
			if t.selectedDeployment > 0 {
				t.selectedDeployment--
				t.updateDeploymentDisplay()
			}
		}
	case models.TabConfigMaps: // ConfigMaps
		if direction > 0 {
			if t.selectedConfigMap < len(t.configMaps)-1 {
				t.selectedConfigMap++
				t.updateConfigMapDisplay()
			}
		} else {
			if t.selectedConfigMap > 0 {
				t.selectedConfigMap--
				t.updateConfigMapDisplay()
			}
		}
	case models.TabSecrets: // Secrets
		if direction > 0 {
			if t.selectedSecret < len(t.secrets)-1 {
				t.selectedSecret++
				t.updateSecretDisplay()
			}
		} else {
			if t.selectedSecret > 0 {
				t.selectedSecret--
				t.updateSecretDisplay()
			}
		}
	case models.TabBuildConfigs: // BuildConfigs
		if direction > 0 {
			if t.selectedBuildConfig < len(t.buildConfigs)-1 {
				t.selectedBuildConfig++
				t.updateBuildConfigDisplay()
			}
		} else {
			if t.selectedBuildConfig > 0 {
				t.selectedBuildConfig--
				t.updateBuildConfigDisplay()
			}
		}
	case models.TabImageStreams: // ImageStreams
		if direction > 0 {
			if t.selectedImageStream < len(t.imageStreams)-1 {
				t.selectedImageStream++
				t.updateImageStreamDisplay()
			}
		} else {
			if t.selectedImageStream > 0 {
				t.selectedImageStream--
				t.updateImageStreamDisplay()
			}
		}
	case models.TabRoutes: // Routes
		if direction > 0 {
			if t.selectedRoute < len(t.routes)-1 {
				t.selectedRoute++
				t.updateRouteDisplay()
			}
		} else {
			if t.selectedRoute > 0 {
				t.selectedRoute--
				t.updateRouteDisplay()
			}
		}
	}

	return t, nil
}

// getHeaderHeight calculates the header height based on terminal size
func (t *TUI) getHeaderHeight() int {
	if t.height < constants.SingleLineHeaderHeightThreshold {
		return 1
	}
	return 2
}
