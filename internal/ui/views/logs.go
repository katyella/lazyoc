package views

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/ui/messages"
)

// LogViewMode represents different log view modes
type LogViewMode string

const (
	LogViewModeApp LogViewMode = "app"
	LogViewModePod LogViewMode = "pod"
)

// LogsView handles log display and streaming
type LogsView struct {
	viewMode        LogViewMode
	podLogs         []string
	appLogs         []string
	logScrollOffset int
	userScrolled    bool
	loadingLogs     bool
	selectedPod     int
	maxLogLines     int
}

// NewLogsView creates a new logs view
func NewLogsView() *LogsView {
	return &LogsView{
		viewMode:        LogViewMode(constants.DefaultLogViewMode),
		podLogs:         []string{},
		appLogs:         []string{constants.InitialLogMessage},
		logScrollOffset: 0,
		userScrolled:    false,
		loadingLogs:     false,
		selectedPod:     0,
		maxLogLines:     constants.MaxLogLines,
	}
}

// GetType returns the view type
func (v *LogsView) GetType() ViewType {
	return ViewTypeLogs
}

// CanHandle returns true if this view can handle the given message
func (v *LogsView) CanHandle(msg tea.Msg) bool {
	switch msg.(type) {
	case PodLogsLoaded, PodLogsError, messages.LoadPodLogsMsg:
		return true
	case tea.KeyMsg:
		keyMsg := msg.(tea.KeyMsg)
		// Handle log navigation and mode switching
		switch keyMsg.String() {
		case "l", "j", "k", "pgup", "pgdn", "home", "end":
			return true
		}
	}
	return false
}

// Update handles messages for the logs view
func (v *LogsView) Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case PodLogsLoaded:
		v.loadingLogs = false
		v.podLogs = msg.Logs
		// Limit log lines to prevent memory issues
		if len(v.podLogs) > v.maxLogLines {
			v.podLogs = v.podLogs[len(v.podLogs)-v.maxLogLines:] // Keep last N lines
		}
		// Auto-scroll to bottom on new logs
		v.userScrolled = false
		v.logScrollOffset = v.getMaxLogScrollOffset()

	case PodLogsError:
		v.loadingLogs = false
		v.podLogs = []string{fmt.Sprintf("Failed to load logs: %v", msg.Err)}
		v.logScrollOffset = 0

	case messages.LoadPodLogsMsg:
		v.loadingLogs = true
		v.podLogs = []string{}
		v.logScrollOffset = 0

	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			// Toggle log view mode when in log panel
			if ctx.FocusedPanel == 2 {
				if v.viewMode == LogViewModeApp {
					v.viewMode = LogViewModePod
					// Auto-load pod logs if not loaded and we have a selected pod
					if len(v.podLogs) == 0 && len(ctx.Pods) > 0 && ctx.SelectedPod < len(ctx.Pods) {
						v.clearPodLogs()
						return v, v.loadPodLogs(ctx)
					}
				} else {
					v.viewMode = LogViewModeApp
				}
			}

		case "j":
			if ctx.FocusedPanel == 2 && v.viewMode == LogViewModePod && len(v.podLogs) > 0 {
				// Scroll down in pod logs
				maxScroll := v.getMaxLogScrollOffset()
				if v.logScrollOffset < maxScroll {
					v.logScrollOffset += 1
					v.userScrolled = true
				}
			}

		case "k":
			if ctx.FocusedPanel == 2 && v.viewMode == LogViewModePod && len(v.podLogs) > 0 {
				// Scroll up in pod logs
				if v.logScrollOffset > 0 {
					v.logScrollOffset -= 1
					v.userScrolled = true
				}
			}

		case "pgup":
			if ctx.FocusedPanel == 2 && v.viewMode == LogViewModePod && len(v.podLogs) > 0 {
				// Page up in pod logs
				pageSize := v.getLogPageSize(ctx)
				v.logScrollOffset = max(0, v.logScrollOffset-pageSize)
				v.userScrolled = true
			}

		case "pgdn":
			if ctx.FocusedPanel == 2 && v.viewMode == LogViewModePod && len(v.podLogs) > 0 {
				// Page down in pod logs
				pageSize := v.getLogPageSize(ctx)
				maxScroll := v.getMaxLogScrollOffset()
				v.logScrollOffset = min(maxScroll, v.logScrollOffset+pageSize)
				v.userScrolled = true
			}

		case "home":
			if ctx.FocusedPanel == 2 && v.viewMode == LogViewModePod && len(v.podLogs) > 0 {
				// Go to top of pod logs
				v.logScrollOffset = 0
				v.userScrolled = true
			}

		case "end":
			if ctx.FocusedPanel == 2 && v.viewMode == LogViewModePod && len(v.podLogs) > 0 {
				// Go to bottom of pod logs
				v.logScrollOffset = v.getMaxLogScrollOffset()
				v.userScrolled = false // At bottom means auto-scroll is re-enabled
			}
		}
	}

	return v, nil
}

// Render renders the logs view with specified dimensions
func (v *LogsView) Render(ctx ViewContext) string {
	return v.RenderWithDimensions(ctx, ctx.Width, ctx.Height/3)
}

// RenderWithDimensions renders the logs view with specific dimensions
func (v *LogsView) RenderWithDimensions(ctx ViewContext, width, height int) string {
	if height <= 0 {
		return ""
	}

	// Calculate content area dimensions
	const (
		borderOverhead    = 2 // top + bottom border
		paddingOverhead   = 2 // top + bottom padding
		logHeaderOverhead = 2 // header line + separator line
	)

	logPanelTotalOverhead := borderOverhead + paddingOverhead + logHeaderOverhead
	maxLogContentLines := height - logPanelTotalOverhead
	if maxLogContentLines < 1 {
		maxLogContentLines = 1
	}

	// Border color based on focus
	logBorderColor := lipgloss.Color("240") // Default gray
	if ctx.FocusedPanel == 2 {
		logBorderColor = lipgloss.Color("12") // Blue when focused
	}

	// Create log panel style
	logStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(logBorderColor).
		Padding(1)

	// Show logs based on current log view mode
	var logText string
	var logHeader string

	if v.viewMode == LogViewModePod {
		logText, logHeader = v.renderPodLogs(ctx, maxLogContentLines, width)
	} else {
		logText, logHeader = v.renderAppLogs(ctx, maxLogContentLines, width)
	}

	// Color the header based on log type with brighter colors
	headerStyle := lipgloss.NewStyle().Bold(true)
	if v.viewMode == LogViewModePod {
		headerStyle = headerStyle.Foreground(lipgloss.Color("207")) // Bright magenta for pod logs
	} else {
		headerStyle = headerStyle.Foreground(lipgloss.Color("51")) // Bright cyan for app logs
	}

	coloredHeader := headerStyle.Render(logHeader)

	separatorLength := len(logHeader)
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", separatorLength))

	fullLogText := fmt.Sprintf("%s\n%s\n%s", coloredHeader, separator, logText)

	return logStyle.Render(fullLogText)
}

// renderPodLogs renders pod logs content
func (v *LogsView) renderPodLogs(ctx ViewContext, maxLogContentLines, width int) (string, string) {
	var logText string
	var logHeader string

	if v.loadingLogs {
		logText = "🔄 Loading pod logs..."
		logHeader = "Pod Logs (Loading...)"
	} else if len(v.podLogs) > 0 {
		// Calculate visible lines strictly based on maxLogContentLines
		visibleLines := maxLogContentLines
		if visibleLines < 1 {
			visibleLines = 1
		}

		start := v.logScrollOffset
		end := start + visibleLines
		if end > len(v.podLogs) {
			end = len(v.podLogs)
		}
		if start >= len(v.podLogs) {
			start = max(0, len(v.podLogs)-visibleLines)
			end = len(v.podLogs)
		}

		visibleLogs := v.podLogs[start:end]

		// Apply coloring to each log line
		coloredLogs := []string{}
		totalLines := 0
		logWidth := width - constants.LogWidthPadding // Account for borders and padding

		for _, line := range visibleLogs {
			colored := v.colorizePodLog(line)

			// Count how many actual lines this log entry will render as
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

		if len(ctx.Pods) > 0 && ctx.SelectedPod < len(ctx.Pods) {
			logHeader = fmt.Sprintf("Pod Logs: %s", ctx.Pods[ctx.SelectedPod].Name)
		} else {
			logHeader = "Pod Logs"
		}
	} else {
		// Show message when no pod logs are available
		if len(ctx.Pods) > 0 && ctx.SelectedPod < len(ctx.Pods) {
			selectedPodName := ctx.Pods[ctx.SelectedPod].Name
			logText = fmt.Sprintf("📋 No logs loaded for pod '%s'", selectedPodName)
			logHeader = fmt.Sprintf("Pod Logs: %s (Not loaded)", selectedPodName)
		} else {
			logText = "📋 No pod selected"
			logHeader = "Pod Logs (No pod selected)"
		}
	}

	return logText, logHeader
}

// renderAppLogs renders application logs content
func (v *LogsView) renderAppLogs(ctx ViewContext, maxLogContentLines, width int) (string, string) {
	// Get recent logs but account for multiline entries
	startIdx := max(0, len(v.appLogs)-constants.LastNAppLogEntries) // Start with last 100 entries
	recentLogs := v.appLogs[startIdx:]

	// Apply coloring and count actual rendered lines
	coloredAppLogs := []string{}
	totalLines := 0
	logWidth := width - 6 // Account for borders and padding

	for _, line := range recentLogs {
		colored := v.colorizeAppLog(line)

		// Count how many actual lines this log entry will render as
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
	logText := strings.Join(coloredAppLogs, "\n")
	logHeader := "App Logs"

	return logText, logHeader
}

// AddAppLog adds a log entry to the application logs
func (v *LogsView) AddAppLog(message string) {
	v.appLogs = append(v.appLogs, message)
	// Keep a reasonable number of app logs in memory
	if len(v.appLogs) > constants.MaxAppLogEntries {
		v.appLogs = v.appLogs[len(v.appLogs)-constants.MaxAppLogEntries:]
	}
}

// GetAppLogs returns the current application logs
func (v *LogsView) GetAppLogs() []string {
	return v.appLogs
}

// clearPodLogs clears the current pod logs and sets loading state
func (v *LogsView) clearPodLogs() {
	v.podLogs = []string{}
	v.logScrollOffset = 0
	v.loadingLogs = true
	v.userScrolled = false // Reset scroll tracking
}

// loadPodLogs returns a command to load logs for the selected pod
func (v *LogsView) loadPodLogs(ctx ViewContext) tea.Cmd {
	if !ctx.Connected || len(ctx.Pods) == 0 || ctx.SelectedPod >= len(ctx.Pods) {
		return func() tea.Msg {
			return PodLogsError{Err: fmt.Errorf("no pod selected or not connected"), PodName: ""}
		}
	}

	selectedPod := ctx.Pods[ctx.SelectedPod]
	return func() tea.Msg {
		return messages.LoadPodLogsMsg{
			PodName:   selectedPod.Name,
			Namespace: selectedPod.Namespace,
		}
	}
}

// getMaxLogScrollOffset returns the maximum scroll offset for logs
func (v *LogsView) getMaxLogScrollOffset() int {
	if len(v.podLogs) == 0 {
		return 0
	}

	visibleLines := 10 // Default fallback
	maxScroll := len(v.podLogs) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// getLogPageSize returns the number of visible log lines per page
func (v *LogsView) getLogPageSize(ctx ViewContext) int {
	// Calculate log panel height similar to renderContent
	availableHeight := ctx.Height - 4 // header + tabs + status + margins
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

// SetViewMode sets the log view mode
func (v *LogsView) SetViewMode(mode LogViewMode) {
	v.viewMode = mode
}

// GetViewMode returns the current log view mode
func (v *LogsView) GetViewMode() LogViewMode {
	return v.viewMode
}

// colorizeAppLog applies color to app log messages based on content
func (v *LogsView) colorizeAppLog(logLine string) string {
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
func (v *LogsView) colorizePodLog(logLine string) string {
	// Define brighter, more readable color styles
	timestampStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))       // Brighter gray
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Bright red + bold
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))            // Orange/yellow
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))             // Bright green
	debugStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))            // Bright blue
	noticeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))           // Cyan for notice

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

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PodLogsLoaded message type
type PodLogsLoaded struct {
	Logs    []string
	PodName string
}

// PodLogsError message type
type PodLogsError struct {
	Err     error
	PodName string
}