package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// HeaderComponent represents the application header
type HeaderComponent struct {
	Width   int
	Height  int
	Version string

	// Cluster connection info
	ClusterName      string
	Namespace        string
	IsConnected      bool
	ConnectionStatus string

	// Display options
	ShowVersion   bool
	ShowCluster   bool
	ShowNamespace bool
	ShowTimestamp bool

	// Styling
	TitleStyle        lipgloss.Style
	ClusterStyle      lipgloss.Style
	NamespaceStyle    lipgloss.Style
	DisconnectedStyle lipgloss.Style
	TimestampStyle    lipgloss.Style
}

// NewHeaderComponent creates a new header component
func NewHeaderComponent(width, height int, version string) *HeaderComponent {
	return &HeaderComponent{
		Width:   width,
		Height:  height,
		Version: version,

		ClusterName:      "",
		Namespace:        "default",
		IsConnected:      false,
		ConnectionStatus: "Disconnected",

		ShowVersion:   true,
		ShowCluster:   true,
		ShowNamespace: true,
		ShowTimestamp: false,

		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true),

		ClusterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true),

		NamespaceStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(false),

		DisconnectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(false),

		TimestampStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Bold(false),
	}
}

// SetClusterInfo updates the cluster connection information
func (h *HeaderComponent) SetClusterInfo(clusterName, namespace string, connected bool) {
	h.ClusterName = clusterName
	h.Namespace = namespace
	h.IsConnected = connected

	if connected {
		h.ConnectionStatus = "Connected"
	} else {
		h.ConnectionStatus = "Disconnected"
	}
}

// SetDimensions updates the header dimensions
func (h *HeaderComponent) SetDimensions(width, height int) {
	h.Width = width
	h.Height = height
}

// Render renders the header component
func (h *HeaderComponent) Render() string {
	if h.Height == 1 {
		return h.renderSingleLine()
	} else if h.Height == 2 {
		return h.renderTwoLines()
	}
	return h.renderMultiLine()
}

// renderSingleLine renders a compact single-line header
func (h *HeaderComponent) renderSingleLine() string {
	// Create title
	title := h.TitleStyle.Render("🚀 LazyOC")

	if h.ShowVersion {
		title += " " + h.TitleStyle.Render("v"+h.Version)
	}

	// Create connection info
	var connInfo string
	if h.IsConnected && h.ShowCluster {
		connInfo = h.ClusterStyle.Render("● " + h.ClusterName)
		if h.ShowNamespace {
			connInfo += " " + h.NamespaceStyle.Render("["+h.Namespace+"]")
		}
	} else {
		connInfo = h.DisconnectedStyle.Render("○ " + h.ConnectionStatus)
	}

	// Calculate spacing
	titleWidth := lipgloss.Width(title)
	connWidth := lipgloss.Width(connInfo)
	spacingWidth := h.Width - titleWidth - connWidth

	var spacing string
	if spacingWidth > 0 {
		spacing = strings.Repeat(" ", spacingWidth)
	}

	line := title + spacing + connInfo

	// Apply container styling
	style := lipgloss.NewStyle().
		Width(h.Width).
		Height(1).
		Align(lipgloss.Left)

	return style.Render(line)
}

// renderTwoLines renders a two-line header with more information
func (h *HeaderComponent) renderTwoLines() string {
	// Line 1: Title and version
	title := h.TitleStyle.Render("🚀 LazyOC")
	if h.ShowVersion {
		title += " " + h.TitleStyle.Render("v"+h.Version)
	}

	var timestamp string
	if h.ShowTimestamp {
		timestamp = h.TimestampStyle.Render(time.Now().Format("15:04:05"))
	}

	// Calculate spacing for line 1
	titleWidth := lipgloss.Width(title)
	timestampWidth := lipgloss.Width(timestamp)
	spacingWidth1 := h.Width - titleWidth - timestampWidth

	var spacing1 string
	if spacingWidth1 > 0 {
		spacing1 = strings.Repeat(" ", spacingWidth1)
	}

	line1 := title + spacing1 + timestamp

	// Line 2: Connection information
	var line2 string
	if h.IsConnected {
		clusterInfo := h.ClusterStyle.Render("● Connected to: " + h.ClusterName)
		var namespaceInfo string
		if h.ShowNamespace {
			namespaceInfo = h.NamespaceStyle.Render("Namespace: " + h.Namespace)
		}

		// Calculate spacing for line 2
		clusterWidth := lipgloss.Width(clusterInfo)
		namespaceWidth := lipgloss.Width(namespaceInfo)
		spacingWidth2 := h.Width - clusterWidth - namespaceWidth

		var spacing2 string
		if spacingWidth2 > 0 {
			spacing2 = strings.Repeat(" ", spacingWidth2)
		}

		line2 = clusterInfo + spacing2 + namespaceInfo
	} else {
		line2 = h.DisconnectedStyle.Render("○ " + h.ConnectionStatus)

		// Center the disconnected message
		line2Width := lipgloss.Width(line2)
		if line2Width < h.Width {
			leftPadding := (h.Width - line2Width) / 2
			rightPadding := h.Width - line2Width - leftPadding
			line2 = strings.Repeat(" ", leftPadding) + line2 + strings.Repeat(" ", rightPadding)
		}
	}

	// Combine lines
	result := lipgloss.JoinVertical(lipgloss.Left, line1, line2)

	// Apply container styling
	style := lipgloss.NewStyle().
		Width(h.Width).
		Height(2)

	return style.Render(result)
}

// renderMultiLine renders a multi-line header with full information
func (h *HeaderComponent) renderMultiLine() string {
	lines := make([]string, 0, h.Height)

	// Line 1: Centered title
	titleLine := h.TitleStyle.Render("🚀 LazyOC - Kubernetes Resource Viewer")
	if h.ShowVersion {
		titleLine += " " + h.TitleStyle.Render("v"+h.Version)
	}

	titleStyle := lipgloss.NewStyle().
		Width(h.Width).
		Align(lipgloss.Center)
	lines = append(lines, titleStyle.Render(titleLine))

	// Line 2: Connection status
	var statusLine string
	if h.IsConnected {
		statusLine = h.ClusterStyle.Render(fmt.Sprintf("● Connected to cluster: %s", h.ClusterName))
	} else {
		statusLine = h.DisconnectedStyle.Render("○ " + h.ConnectionStatus)
	}

	statusStyle := lipgloss.NewStyle().
		Width(h.Width).
		Align(lipgloss.Center)
	lines = append(lines, statusStyle.Render(statusLine))

	// Line 3 (if height >= 3): Namespace and timestamp
	if h.Height >= 3 {
		var infoLine string
		if h.IsConnected && h.ShowNamespace {
			infoLine = h.NamespaceStyle.Render(fmt.Sprintf("Namespace: %s", h.Namespace))
		}

		if h.ShowTimestamp {
			timestamp := h.TimestampStyle.Render(time.Now().Format("2006-01-02 15:04:05"))
			if infoLine != "" {
				// Space them out
				infoWidth := lipgloss.Width(infoLine)
				timestampWidth := lipgloss.Width(timestamp)
				spacingWidth := h.Width - infoWidth - timestampWidth

				if spacingWidth > 0 {
					spacing := strings.Repeat(" ", spacingWidth)
					infoLine = infoLine + spacing + timestamp
				} else {
					infoLine = infoLine + " " + timestamp
				}
			} else {
				infoLine = timestamp
			}
		}

		infoStyle := lipgloss.NewStyle().
			Width(h.Width).
			Align(lipgloss.Center)
		lines = append(lines, infoStyle.Render(infoLine))
	}

	// Add empty lines if needed to fill height
	for len(lines) < h.Height {
		emptyStyle := lipgloss.NewStyle().Width(h.Width)
		lines = append(lines, emptyStyle.Render(""))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// SetShowVersion toggles version display
func (h *HeaderComponent) SetShowVersion(show bool) {
	h.ShowVersion = show
}

// SetShowCluster toggles cluster name display
func (h *HeaderComponent) SetShowCluster(show bool) {
	h.ShowCluster = show
}

// SetShowNamespace toggles namespace display
func (h *HeaderComponent) SetShowNamespace(show bool) {
	h.ShowNamespace = show
}

// SetShowTimestamp toggles timestamp display
func (h *HeaderComponent) SetShowTimestamp(show bool) {
	h.ShowTimestamp = show
}

// GetConnectionStatus returns the current connection status
func (h *HeaderComponent) GetConnectionStatus() string {
	return h.ConnectionStatus
}

// IsClusterConnected returns whether a cluster is connected
func (h *HeaderComponent) IsClusterConnected() bool {
	return h.IsConnected
}

// GetClusterName returns the current cluster name
func (h *HeaderComponent) GetClusterName() string {
	return h.ClusterName
}

// GetNamespace returns the current namespace
func (h *HeaderComponent) GetNamespace() string {
	return h.Namespace
}
