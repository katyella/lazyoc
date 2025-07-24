package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
)

// HeaderComponent renders the application header
type HeaderComponent struct {
	BaseComponent

	// Header content
	title           string
	version         string
	connectionState ConnectionState
	clusterInfo     ClusterInfo
	namespace       string

	// Styles
	titleStyle      lipgloss.Style
	infoStyle       lipgloss.Style
	separatorStyle  lipgloss.Style
	connectedStyle  lipgloss.Style
	disconnectedStyle lipgloss.Style
}

// ConnectionState represents the connection status
type ConnectionState int

const (
	ConnectionStateDisconnected ConnectionState = iota
	ConnectionStateConnecting
	ConnectionStateConnected
	ConnectionStateError
)

// ClusterInfo contains cluster information
type ClusterInfo struct {
	Type    string // "Kubernetes" or "OpenShift"
	Version string
	Context string
}

// NewHeaderComponent creates a new header component
func NewHeaderComponent(title, version string) *HeaderComponent {
	return &HeaderComponent{
		title:   title,
		version: version,
		
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(constants.ColorBlue)),
		
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorGray)),
		
		separatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorDarkGray)),
		
		connectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorGreen)),
		
		disconnectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorYellow)),
	}
}

// Init initializes the header component
func (h *HeaderComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages for the header component
func (h *HeaderComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case ConnectionStateMsg:
		h.connectionState = msg.State
		h.clusterInfo = msg.ClusterInfo
		h.namespace = msg.Namespace
	}
	return nil, nil
}

// View renders the header component
func (h *HeaderComponent) View() string {
	if h.width == 0 {
		return ""
	}

	// Build title line
	titleText := fmt.Sprintf("%s v%s", h.title, h.version)
	title := h.titleStyle.Render(titleText)

	// Build connection status
	connectionStatus := h.renderConnectionStatus()

	// Calculate spacing
	titleWidth := lipgloss.Width(titleText)
	statusWidth := lipgloss.Width(stripANSI(connectionStatus))
	spacing := h.width - titleWidth - statusWidth - 2 // 2 for margins

	// Combine title and status
	var topLine string
	if spacing > 0 {
		topLine = fmt.Sprintf(" %s%s%s ", title, strings.Repeat(" ", spacing), connectionStatus)
	} else {
		// Not enough space, show only title
		topLine = " " + title + " "
	}

	// Add separator
	separator := h.separatorStyle.Render(strings.Repeat("─", h.width))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topLine,
		separator,
	)
}

// renderConnectionStatus renders the connection status section
func (h *HeaderComponent) renderConnectionStatus() string {
	switch h.connectionState {
	case ConnectionStateDisconnected:
		return h.disconnectedStyle.Render(constants.StatusDisconnected)
	
	case ConnectionStateConnecting:
		return h.disconnectedStyle.Render(constants.StatusConnecting)
	
	case ConnectionStateConnected:
		clusterType := h.clusterInfo.Type
		if clusterType == "" {
			clusterType = "Kubernetes"
		}
		
		status := fmt.Sprintf("● %s", clusterType)
		if h.clusterInfo.Version != "" {
			status += fmt.Sprintf(" %s", h.clusterInfo.Version)
		}
		if h.namespace != "" && h.namespace != constants.DefaultNamespace {
			status += fmt.Sprintf(" | %s", h.namespace)
		}
		if h.clusterInfo.Context != "" {
			status += fmt.Sprintf(" [%s]", h.clusterInfo.Context)
		}
		
		return h.connectedStyle.Render(status)
	
	case ConnectionStateError:
		return h.disconnectedStyle.Render("✗ Connection Error")
	
	default:
		return ""
	}
}

// SetConnectionState updates the connection state
func (h *HeaderComponent) SetConnectionState(state ConnectionState, info ClusterInfo, namespace string) {
	h.connectionState = state
	h.clusterInfo = info
	h.namespace = namespace
}

// Messages for header updates
type ConnectionStateMsg struct {
	State       ConnectionState
	ClusterInfo ClusterInfo
	Namespace   string
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(str string) string {
	// Simple implementation - in production, use a proper ANSI stripping library
	// This is a placeholder that removes common color codes
	result := str
	for _, code := range []string{
		"\033[0m", "\033[1m", "\033[2m", "\033[30m", "\033[31m", 
		"\033[32m", "\033[33m", "\033[34m", "\033[35m", "\033[36m", 
		"\033[37m", "\033[90m", "\033[91m", "\033[92m", "\033[93m",
		"\033[94m", "\033[95m", "\033[96m", "\033[97m",
	} {
		result = strings.ReplaceAll(result, code, "")
	}
	return result
}