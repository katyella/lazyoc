package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StatusType represents different types of status messages
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
	StatusLoading
)

// String returns the string representation of a status type
func (st StatusType) String() string {
	switch st {
	case StatusInfo:
		return "INFO"
	case StatusSuccess:
		return "SUCCESS"
	case StatusWarning:
		return "WARNING"
	case StatusError:
		return "ERROR"
	case StatusLoading:
		return "LOADING"
	default:
		return "UNKNOWN"
	}
}

// StatusBarComponent represents the application status bar
type StatusBarComponent struct {
	Width  int
	Height int

	// Status information
	LeftStatus   string
	CenterStatus string
	RightStatus  string
	StatusType   StatusType
	LastUpdated  time.Time

	// Connection info
	ClusterName     string
	Namespace       string
	IsConnected     bool
	ConnectionCount int

	// Application state
	ActivePanel string
	KeyHints    []KeyHint
	FocusMode   bool

	// Display options
	ShowKeyHints   bool
	ShowTimestamp  bool
	ShowConnection bool
	ShowResources  bool
	ShowFocus      bool
	AutoUpdate     bool

	// Styling
	BaseStyle   lipgloss.Style
	LeftStyle   lipgloss.Style
	CenterStyle lipgloss.Style
	RightStyle  lipgloss.Style

	// Status type styles
	infoStyle    lipgloss.Style
	successStyle lipgloss.Style
	warningStyle lipgloss.Style
	errorStyle   lipgloss.Style
	loadingStyle lipgloss.Style

	// Key hint styling
	keyHintStyle lipgloss.Style
}

// KeyHint represents a keyboard shortcut hint
type KeyHint struct {
	Key         string
	Description string
	Category    string
}

// NewStatusBarComponent creates a new status bar component
func NewStatusBarComponent(width, height int) *StatusBarComponent {
	return &StatusBarComponent{
		Width:  width,
		Height: height,

		LeftStatus:   "Ready",
		CenterStatus: "",
		RightStatus:  "",
		StatusType:   StatusInfo,
		LastUpdated:  time.Now(),

		ClusterName:     "",
		Namespace:       "default",
		IsConnected:     false,
		ConnectionCount: 0,

		ActivePanel: "main",
		KeyHints:    make([]KeyHint, 0),
		FocusMode:   false,

		ShowKeyHints:   true,
		ShowTimestamp:  false,
		ShowConnection: true,
		ShowResources:  true,
		ShowFocus:      true,
		AutoUpdate:     true,

		BaseStyle: lipgloss.NewStyle().
			Width(width).
			Height(height).
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("0")),  // Black

		LeftStyle: lipgloss.NewStyle().
			Bold(true),

		CenterStyle: lipgloss.NewStyle().
			Align(lipgloss.Center),

		RightStyle: lipgloss.NewStyle().
			Align(lipgloss.Right),

		// Status type specific styles
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")), // Blue

		successStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")), // Green

		warningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")), // Yellow

		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")), // Red

		loadingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")), // Cyan

		keyHintStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Bold(false),
	}
}

// SetDimensions updates the status bar dimensions
func (sb *StatusBarComponent) SetDimensions(width, height int) {
	sb.Width = width
	sb.Height = height
	sb.BaseStyle = sb.BaseStyle.Width(width).Height(height)
}

// SetStatus sets the main status message
func (sb *StatusBarComponent) SetStatus(message string, statusType StatusType) {
	sb.LeftStatus = message
	sb.StatusType = statusType
	sb.LastUpdated = time.Now()
}

// SetLeftStatus sets the left status message
func (sb *StatusBarComponent) SetLeftStatus(status string) {
	sb.LeftStatus = status
	sb.LastUpdated = time.Now()
}

// SetCenterStatus sets the center status message
func (sb *StatusBarComponent) SetCenterStatus(status string) {
	sb.CenterStatus = status
}

// SetRightStatus sets the right status message
func (sb *StatusBarComponent) SetRightStatus(status string) {
	sb.RightStatus = status
}

// SetClusterInfo updates cluster connection information
func (sb *StatusBarComponent) SetClusterInfo(clusterName, namespace string, connected bool) {
	sb.ClusterName = clusterName
	sb.Namespace = namespace
	sb.IsConnected = connected

	if connected {
		sb.ConnectionCount++
	}

	sb.LastUpdated = time.Now()
}

// SetActivePanel sets the currently active panel
func (sb *StatusBarComponent) SetActivePanel(panel string) {
	sb.ActivePanel = panel
}

// SetKeyHints sets the keyboard shortcut hints
func (sb *StatusBarComponent) SetKeyHints(hints []KeyHint) {
	sb.KeyHints = hints
}

// AddKeyHint adds a keyboard shortcut hint
func (sb *StatusBarComponent) AddKeyHint(key, description, category string) {
	hint := KeyHint{
		Key:         key,
		Description: description,
		Category:    category,
	}
	sb.KeyHints = append(sb.KeyHints, hint)
}

// ClearKeyHints clears all keyboard shortcut hints
func (sb *StatusBarComponent) ClearKeyHints() {
	sb.KeyHints = make([]KeyHint, 0)
}

// SetFocusMode sets focus mode (simplified display)
func (sb *StatusBarComponent) SetFocusMode(focus bool) {
	sb.FocusMode = focus
}

// Render renders the status bar
func (sb *StatusBarComponent) Render() string {
	if sb.FocusMode {
		return sb.renderFocusMode()
	}

	if sb.Height == 1 {
		return sb.renderSingleLine()
	}

	return sb.renderMultiLine()
}

// renderSingleLine renders a single-line status bar
func (sb *StatusBarComponent) renderSingleLine() string {
	// Build left section
	leftSection := sb.buildLeftSection()

	// Build center section
	centerSection := sb.buildCenterSection()

	// Build right section
	rightSection := sb.buildRightSection()

	// Calculate spacing
	leftWidth := lipgloss.Width(leftSection)
	centerWidth := lipgloss.Width(centerSection)
	rightWidth := lipgloss.Width(rightSection)

	// Calculate center position
	centerPosition := (sb.Width - centerWidth) / 2

	// Calculate spacing
	leftSpacing := centerPosition - leftWidth
	rightSpacing := sb.Width - centerPosition - centerWidth - rightWidth

	// Ensure minimum spacing
	if leftSpacing < 1 {
		leftSpacing = 1
	}
	if rightSpacing < 1 {
		rightSpacing = 1
	}

	// Build final status line
	var statusLine strings.Builder
	statusLine.WriteString(leftSection)
	statusLine.WriteString(strings.Repeat(" ", leftSpacing))
	statusLine.WriteString(centerSection)
	statusLine.WriteString(strings.Repeat(" ", rightSpacing))
	statusLine.WriteString(rightSection)

	// Truncate if too long
	line := statusLine.String()
	if len(line) > sb.Width {
		line = line[:sb.Width-3] + "..."
	}

	return sb.BaseStyle.Render(line)
}

// renderMultiLine renders a multi-line status bar
func (sb *StatusBarComponent) renderMultiLine() string {
	lines := make([]string, 0, sb.Height)

	// Line 1: Main status and connection info
	line1 := sb.buildMainStatusLine()
	lines = append(lines, line1)

	// Line 2 (if height >= 2): Key hints or additional info
	if sb.Height >= 2 {
		if sb.ShowKeyHints && len(sb.KeyHints) > 0 {
			line2 := sb.buildKeyHintLine()
			lines = append(lines, line2)
		} else {
			// Additional info line
			line2 := sb.buildAdditionalInfoLine()
			lines = append(lines, line2)
		}
	}

	// Additional lines for more key hints
	for i := 2; i < sb.Height; i++ {
		lines = append(lines, "")
	}

	// Join lines and apply styling
	content := strings.Join(lines, "\n")
	return sb.BaseStyle.Render(content)
}

// renderFocusMode renders a minimal focus mode status bar
func (sb *StatusBarComponent) renderFocusMode() string {
	focusContent := fmt.Sprintf("Focus: %s", sb.ActivePanel)

	if sb.IsConnected {
		focusContent += fmt.Sprintf(" • %s", sb.ClusterName)
	}

	// Center the content
	contentWidth := lipgloss.Width(focusContent)
	if contentWidth < sb.Width {
		padding := (sb.Width - contentWidth) / 2
		leftPad := strings.Repeat(" ", padding)
		rightPad := strings.Repeat(" ", sb.Width-contentWidth-padding)
		focusContent = leftPad + focusContent + rightPad
	}

	return sb.BaseStyle.Render(focusContent)
}

// buildLeftSection builds the left section of the status bar
func (sb *StatusBarComponent) buildLeftSection() string {
	// Apply status type styling
	var statusStyle lipgloss.Style
	switch sb.StatusType {
	case StatusInfo:
		statusStyle = sb.infoStyle
	case StatusSuccess:
		statusStyle = sb.successStyle
	case StatusWarning:
		statusStyle = sb.warningStyle
	case StatusError:
		statusStyle = sb.errorStyle
	case StatusLoading:
		statusStyle = sb.loadingStyle
	}

	// Add status indicator
	var indicator string
	switch sb.StatusType {
	case StatusSuccess:
		indicator = "✓ "
	case StatusWarning:
		indicator = "⚠ "
	case StatusError:
		indicator = "✗ "
	case StatusLoading:
		indicator = "⟳ "
	default:
		indicator = "● "
	}

	return statusStyle.Render(indicator + sb.LeftStatus)
}

// buildCenterSection builds the center section of the status bar
func (sb *StatusBarComponent) buildCenterSection() string {
	if sb.CenterStatus == "" {
		return ""
	}
	return sb.CenterStyle.Render(sb.CenterStatus)
}

// buildRightSection builds the right section of the status bar
func (sb *StatusBarComponent) buildRightSection() string {
	var rightParts []string

	// Add custom right status if set
	if sb.RightStatus != "" {
		rightParts = append(rightParts, sb.RightStatus)
	}

	// Add active panel if enabled
	if sb.ShowFocus && sb.ActivePanel != "" {
		rightParts = append(rightParts, fmt.Sprintf("Focus: %s", sb.ActivePanel))
	}

	// Add connection info if enabled
	if sb.ShowConnection && sb.IsConnected {
		connInfo := fmt.Sprintf("%s[%s]", sb.ClusterName, sb.Namespace)
		rightParts = append(rightParts, connInfo)
	}

	// Add timestamp if enabled
	if sb.ShowTimestamp {
		timestamp := sb.LastUpdated.Format("15:04:05")
		rightParts = append(rightParts, timestamp)
	}

	return strings.Join(rightParts, " • ")
}

// buildMainStatusLine builds the main status line for multi-line mode
func (sb *StatusBarComponent) buildMainStatusLine() string {
	left := sb.buildLeftSection()
	right := sb.buildRightSection()

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacingWidth := sb.Width - leftWidth - rightWidth

	var spacing string
	if spacingWidth > 0 {
		spacing = strings.Repeat(" ", spacingWidth)
	}

	return left + spacing + right
}

// buildKeyHintLine builds a line of keyboard hints
func (sb *StatusBarComponent) buildKeyHintLine() string {
	if len(sb.KeyHints) == 0 {
		return ""
	}

	var hints []string
	availableWidth := sb.Width - 4 // Account for padding

	for _, hint := range sb.KeyHints {
		hintText := fmt.Sprintf("%s:%s", hint.Key, hint.Description)
		hintStyled := sb.keyHintStyle.Render(hintText)

		// Check if we have space for this hint
		currentLength := 0
		for _, h := range hints {
			currentLength += lipgloss.Width(h) + 2 // +2 for separator
		}

		if currentLength+lipgloss.Width(hintStyled) <= availableWidth {
			hints = append(hints, hintStyled)
		} else {
			// If we can't fit this hint, add "..." and break
			if len(hints) > 0 {
				hints = append(hints, sb.keyHintStyle.Render("..."))
			}
			break
		}
	}

	if len(hints) == 0 {
		return ""
	}

	hintLine := strings.Join(hints, "  ")

	// Center the hints
	hintWidth := lipgloss.Width(hintLine)
	if hintWidth < sb.Width {
		padding := (sb.Width - hintWidth) / 2
		leftPad := strings.Repeat(" ", padding)
		rightPad := strings.Repeat(" ", sb.Width-hintWidth-padding)
		hintLine = leftPad + hintLine + rightPad
	}

	return hintLine
}

// buildAdditionalInfoLine builds an additional info line
func (sb *StatusBarComponent) buildAdditionalInfoLine() string {
	var infoParts []string

	if sb.ShowResources {
		infoParts = append(infoParts, "Resources: Ready")
	}

	if sb.IsConnected {
		infoParts = append(infoParts, fmt.Sprintf("Connected: %d times", sb.ConnectionCount))
	}

	if len(infoParts) == 0 {
		return ""
	}

	infoLine := strings.Join(infoParts, " • ")
	infoStyled := sb.keyHintStyle.Render(infoLine)

	// Center the info
	infoWidth := lipgloss.Width(infoStyled)
	if infoWidth < sb.Width {
		padding := (sb.Width - infoWidth) / 2
		leftPad := strings.Repeat(" ", padding)
		rightPad := strings.Repeat(" ", sb.Width-infoWidth-padding)
		infoStyled = leftPad + infoStyled + rightPad
	}

	return infoStyled
}

// UpdateTimestamp updates the last updated timestamp
func (sb *StatusBarComponent) UpdateTimestamp() {
	sb.LastUpdated = time.Now()
}

// GetStatus returns the current status information
func (sb *StatusBarComponent) GetStatus() (string, StatusType) {
	return sb.LeftStatus, sb.StatusType
}

// IsConnected returns whether a cluster is connected
func (sb *StatusBarComponent) IsClusterConnected() bool {
	return sb.IsConnected
}

// GetActivePanel returns the currently active panel
func (sb *StatusBarComponent) GetActivePanel() string {
	return sb.ActivePanel
}

// CreateDefaultKeyHints creates default keyboard shortcuts for LazyOC
func CreateDefaultKeyHints() []KeyHint {
	return []KeyHint{
		{"q", "quit", "global"},
		{"?", "help", "global"},
		{"tab", "next-tab", "navigation"},
		{"h/l", "prev/next", "navigation"},
		{"j/k", "up/down", "navigation"},
		{"r", "refresh", "action"},
		{"c", "collapse", "view"},
	}
}
