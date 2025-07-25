package constants

import "time"

// UI theme and display constants
const (
	// DefaultTheme is the default UI theme
	DefaultTheme = "dark"

	// DefaultNamespace is the default Kubernetes namespace
	DefaultNamespace = "default"
)

// UI dimensions
const (
	// MinTerminalWidth is the minimum terminal width required
	MinTerminalWidth = 80

	// MinTerminalHeight is the minimum terminal height required
	MinTerminalHeight = 24

	// StatusBarHeight is the height of the status bar in lines
	StatusBarHeight = 3
)

// ResourceTabs defines the available resource tabs in the UI
var ResourceTabs = []string{"Pods", "Services", "Deployments", "ConfigMaps", "Secrets"}

// PanelNames defines the available panels in the UI
var PanelNames = []string{"Main", "Details", "Logs"}

// Log view constants
const (
	// DefaultLogViewMode is the default log view mode
	DefaultLogViewMode = "app"

	// PodLogViewMode is the pod log view mode
	PodLogViewMode = "pod"
)

// UI Messages
const (
	// InitialLogMessage is the first message shown in logs
	InitialLogMessage = "LazyOC started"

	// DefaultDetailContent is shown when no resource is selected
	DefaultDetailContent = "Select a resource to view details"

	// InitializingMessage is shown during startup
	InitializingMessage = "Initializing LazyOC..."

	// ConnectingStatus is shown when connecting to cluster
	ConnectingStatus = "⟳ Connecting..."

	// NotConnectedMessage is shown when not connected
	NotConnectedMessage = "○ Not connected - Run 'oc login' or use --kubeconfig"

	// LoadingPodsMessage is shown when loading pods
	LoadingPodsMessage = "📦 Pods\n\nLoading pods..."

	// NoLogsAvailableMessage is shown when pod has no logs
	NoLogsAvailableMessage = "No logs available for this pod"

	// ComingSoonMessage is placeholder for unimplemented features
	ComingSoonMessage = "Coming soon..."
)

// Modal dimensions
const (
	// HelpModalWidth is the width of the help modal
	HelpModalWidth = 60

	// HelpModalHeight is the height of the help modal
	HelpModalHeight = 22

	// ProjectModalMinHeight is the minimum height for project modal
	ProjectModalMinHeight = 6

	// ProjectModalMaxHeight is the maximum height for project modal
	ProjectModalMaxHeight = 15

	// ProjectModalMinWidth is the minimum width for project modal
	ProjectModalMinWidth = 4

	// ProjectModalMaxWidth is the maximum width for project modal
	ProjectModalMaxWidth = 60
)

// Animation constants
const (
	// SpinnerAnimationInterval is the interval for spinner animation
	SpinnerAnimationInterval = 100 * time.Millisecond

	// InitialTickDelay is the initial delay before ticking
	InitialTickDelay = 100 * time.Millisecond
)

// Layout constants
const (
	// DefaultFocusedPanel is the initially focused panel (0=main)
	DefaultFocusedPanel = 0

	// PanelCount is the number of panels to cycle through
	PanelCount = 3

	// MainPanelWidthRatio is the width ratio for main panel when details shown
	MainPanelWidthRatio = 2.0 / 3.0

	// MinMainContentLines is the minimum lines for main content
	MinMainContentLines = 10

	// MinLogContentLines is the minimum lines for log content
	MinLogContentLines = 2

	// LogHeightRatio is the ratio of height for log panel
	LogHeightRatio = 1.0 / 3.0

	// DefaultLogHeight is the default log height when ratio calculation is too small
	DefaultLogHeight = 15

	// LogWidthPadding is the padding for log width calculations
	LogWidthPadding = 6

	// CompactStatusWidthThreshold is the width threshold for compact status
	CompactStatusWidthThreshold = 80

	// SingleLineHeaderHeightThreshold is the height threshold for single line header
	SingleLineHeaderHeightThreshold = 20

	// LastNAppLogEntries is the number of recent app log entries to show
	LastNAppLogEntries = 100

	// PodNameTruncateLength is the length to truncate pod names
	PodNameTruncateLength = 38

	// PodNameTruncateLengthCompact is the compact length to truncate pod names
	PodNameTruncateLengthCompact = 35

	// DefaultPodLogTailLines is the default number of lines to tail from pod logs
	DefaultPodLogTailLines = 100
)
