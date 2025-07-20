package components

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of a log level
func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Source    string
	Extra     map[string]interface{}
}

// Format formats the log entry as a string
func (le LogEntry) Format() string {
	timestamp := le.Timestamp.Format("15:04:05.000")
	level := le.Level.String()
	
	var line strings.Builder
	line.WriteString(fmt.Sprintf("[%s] %5s", timestamp, level))
	
	if le.Source != "" {
		line.WriteString(fmt.Sprintf(" [%s]", le.Source))
	}
	
	line.WriteString(": ")
	line.WriteString(le.Message)
	
	return line.String()
}

// LogPane represents a streaming log pane
type LogPane struct {
	viewport.Model
	Width     int
	Height    int
	MinHeight int
	
	// State management
	Visible      bool
	Collapsed    bool
	Focused      bool
	AutoScroll   bool
	Paused       bool
	Ready        bool  // Track if viewport is properly initialized
	
	// Log management
	entries      []LogEntry
	maxEntries   int
	currentLevel LogLevel
	filters      map[string]bool
	mutex        sync.RWMutex
	
	// Display options
	ShowHeader     bool
	ShowScrollBar  bool
	ShowTimestamp  bool
	ShowLevel      bool
	ShowSource     bool
	WrapLines      bool
	
	// Styling
	HeaderStyle    lipgloss.Style
	TitleStyle     lipgloss.Style
	BorderStyle    lipgloss.Style
	ScrollBarStyle lipgloss.Style
	
	// Level-specific styles
	debugStyle lipgloss.Style
	infoStyle  lipgloss.Style
	warnStyle  lipgloss.Style
	errorStyle lipgloss.Style
	fatalStyle lipgloss.Style
}

// NewLogPane creates a new log pane
func NewLogPane(width, height int) *LogPane {
	// DON'T create viewport yet - wait for proper initialization
	var vp viewport.Model
	
	return &LogPane{
		Model:        vp,
		Width:        width,
		Height:       height,
		MinHeight:    5,
		
		Visible:      true,
		Collapsed:    false,
		Focused:      false,
		AutoScroll:   true,
		Paused:       false,
		Ready:        false, // Not ready until properly sized
		
		entries:      make([]LogEntry, 0),
		maxEntries:   1000,
		currentLevel: LogLevelDebug,
		filters:      make(map[string]bool),
		
		ShowHeader:    true,
		ShowScrollBar: true,
		ShowTimestamp: true,
		ShowLevel:     true,
		ShowSource:    false,
		WrapLines:     true,
		
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true).
			Padding(0, 1),
			
		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true),
			
		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")),
			
		ScrollBarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
			
		// Level-specific styling
		debugStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")), // Gray
			
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")), // Blue
			
		warnStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")), // Yellow
			
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).  // Red
			Bold(true),
			
		fatalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("9")).  // Red background
			Bold(true),
	}
}

// SetVisible sets the visibility of the log pane
func (lp *LogPane) SetVisible(visible bool) {
	lp.Visible = visible
}

// IsVisible returns whether the log pane is visible
func (lp *LogPane) IsVisible() bool {
	return lp.Visible
}

// Toggle toggles the visibility of the log pane
func (lp *LogPane) Toggle() {
	lp.Visible = !lp.Visible
}

// Collapse collapses the log pane to show only the header
func (lp *LogPane) Collapse() {
	lp.Collapsed = true
}

// Expand expands the log pane to show full content
func (lp *LogPane) Expand() {
	lp.Collapsed = false
}

// ToggleCollapse toggles the collapsed state
func (lp *LogPane) ToggleCollapse() {
	lp.Collapsed = !lp.Collapsed
}

// SetFocus sets the focus state
func (lp *LogPane) SetFocus(focused bool) {
	lp.Focused = focused
	
	// Update border style based on focus
	if focused {
		lp.BorderStyle = lp.BorderStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12"))
	} else {
		lp.BorderStyle = lp.BorderStyle.
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))
	}
}

// SetDimensions updates the pane dimensions and initializes viewport if needed
func (lp *LogPane) SetDimensions(width, height int) {
	lp.Width = width
	lp.Height = height
	
	// Enforce minimum height
	if height < lp.MinHeight {
		lp.Height = lp.MinHeight
	}
	
	// Calculate viewport dimensions
	vpHeight := lp.Height - 2 // Border
	if lp.ShowHeader {
		vpHeight -= 1 // Header
	}
	
	vpWidth := width - 2 // Border
	if lp.ShowScrollBar {
		vpWidth -= 1 // Scroll bar
	}
	
	// Initialize or update viewport with proper dimensions
	if !lp.Ready && vpWidth > 0 && vpHeight > 0 {
		// First time setup - create viewport with proper dimensions
		lp.Model = viewport.New(vpWidth, vpHeight)
		lp.Model.HighPerformanceRendering = true
		lp.Ready = true
		
		// Viewport successfully initialized
		
		// Set initial content immediately since this is called from Update/Init path
		// This is safe because SetDimensions is only called from Update flow
		lp.refreshContent()
	} else if lp.Ready {
		// Update existing viewport
		lp.Model.Width = vpWidth
		lp.Model.Height = vpHeight
		
		// Don't refresh content here - it should be triggered by specific messages
	}
}

// AddLog adds a new log entry and triggers a refresh if ready
func (lp *LogPane) AddLog(level LogLevel, message, source string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Source:    source,
		Extra:     make(map[string]interface{}),
	}
	
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	// Add entry
	lp.entries = append(lp.entries, entry)
	
	// Trim entries if we exceed max
	if len(lp.entries) > lp.maxEntries {
		// Remove oldest entries
		excess := len(lp.entries) - lp.maxEntries
		lp.entries = lp.entries[excess:]
	}
	
	// If ready, refresh content immediately (this is safe if called from init/update context)
	if lp.Ready {
		lp.refreshContent()
		if lp.AutoScroll && !lp.Paused {
			lp.Model.GotoBottom()
		}
	}
}

// AddLogf adds a formatted log entry
func (lp *LogPane) AddLogf(level LogLevel, source, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	lp.AddLog(level, message, source)
}

// Debug adds a debug log entry
func (lp *LogPane) Debug(message, source string) {
	lp.AddLog(LogLevelDebug, message, source)
}

// Info adds an info log entry
func (lp *LogPane) Info(message, source string) {
	lp.AddLog(LogLevelInfo, message, source)
}

// Warn adds a warning log entry
func (lp *LogPane) Warn(message, source string) {
	lp.AddLog(LogLevelWarn, message, source)
}

// Error adds an error log entry
func (lp *LogPane) Error(message, source string) {
	lp.AddLog(LogLevelError, message, source)
}

// Fatal adds a fatal log entry
func (lp *LogPane) Fatal(message, source string) {
	lp.AddLog(LogLevelFatal, message, source)
}

// SetLogLevel sets the minimum log level to display (no auto-refresh)
func (lp *LogPane) SetLogLevel(level LogLevel) {
	lp.currentLevel = level
	// Don't auto-refresh here - it will be handled in Update()
}

// SetMaxEntries sets the maximum number of log entries to keep (no auto-refresh)
func (lp *LogPane) SetMaxEntries(max int) {
	lp.maxEntries = max
	
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	// Trim existing entries if necessary
	if len(lp.entries) > max {
		excess := len(lp.entries) - max
		lp.entries = lp.entries[excess:]
		// Don't auto-refresh here - it will be handled in Update()
	}
}

// ClearLogs clears all log entries (no auto-refresh)
func (lp *LogPane) ClearLogs() {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.entries = make([]LogEntry, 0)
	// Don't auto-refresh here - it will be handled in Update()
}

// ToggleAutoScroll toggles auto-scroll behavior
func (lp *LogPane) ToggleAutoScroll() {
	lp.AutoScroll = !lp.AutoScroll
}

// TogglePause toggles log streaming pause
func (lp *LogPane) TogglePause() {
	lp.Paused = !lp.Paused
}

// LogRefreshMsg indicates logs should be refreshed
type LogRefreshMsg struct{}

// Update handles Bubble Tea messages
func (lp *LogPane) Update(msg tea.Msg) (*LogPane, tea.Cmd) {
	if !lp.Visible || lp.Collapsed {
		return lp, nil
	}
	
	var cmd tea.Cmd
	lp.Model, cmd = lp.Model.Update(msg)
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle resize and refresh content if ready
		if lp.Ready {
			lp.refreshContent()
		}
		
	case LogRefreshMsg:
		// Handle explicit refresh requests
		if lp.Ready {
			lp.refreshContent()
			if lp.AutoScroll && !lp.Paused {
				lp.Model.GotoBottom()
			}
		}
		
	case tea.KeyMsg:
		if lp.Focused {
			switch msg.String() {
			case "c":
				lp.ClearLogs()
				lp.refreshContent() // Safe to call in Update()
			case "p":
				lp.TogglePause()
			case "a":
				lp.ToggleAutoScroll()
			case "l":
				lp.ToggleCollapse()
			case "1":
				lp.SetLogLevel(LogLevelDebug)
				lp.refreshContent() // Safe to call in Update()
			case "2":
				lp.SetLogLevel(LogLevelInfo)
				lp.refreshContent()
			case "3":
				lp.SetLogLevel(LogLevelWarn)
				lp.refreshContent()
			case "4":
				lp.SetLogLevel(LogLevelError)
				lp.refreshContent()
			case "5":
				lp.SetLogLevel(LogLevelFatal)
				lp.refreshContent()
			case "g":
				lp.Model.GotoTop()
			case "G":
				lp.Model.GotoBottom()
			}
		}
	}
	
	return lp, cmd
}

// Render renders the log pane
func (lp *LogPane) Render() string {
	if !lp.Visible {
		return ""
	}
	
	if lp.Collapsed {
		return lp.renderCollapsed()
	}
	
	return lp.renderExpanded()
}

// renderCollapsed renders the collapsed state
func (lp *LogPane) renderCollapsed() string {
	lp.mutex.RLock()
	entryCount := len(lp.entries)
	lp.mutex.RUnlock()
	
	collapsedContent := fmt.Sprintf("▶ Logs (%d entries) - %s", entryCount, lp.getStatusString())
	
	style := lp.BorderStyle.
		Width(lp.Width).
		Height(3) // Minimum height for collapsed state
	
	return style.Render(collapsedContent)
}

// renderExpanded renders the expanded state with full content
func (lp *LogPane) renderExpanded() string {
	var content strings.Builder
	
	// Header
	if lp.ShowHeader {
		header := lp.renderHeader()
		content.WriteString(header)
		content.WriteString("\n")
	}
	
	// Content area - use viewport's current content, don't modify it
	if lp.Ready {
		viewportContent := lp.Model.View()
		
		if lp.ShowScrollBar {
			viewportContent = lp.addScrollBar(viewportContent)
		}
		content.WriteString(viewportContent)
	} else {
		// Show loading state if not ready
		content.WriteString("Initializing log view...")
	}
	
	// Apply border
	style := lp.BorderStyle.
		Width(lp.Width).
		Height(lp.Height)
	
	return style.Render(content.String())
}

// renderHeader renders the header section
func (lp *LogPane) renderHeader() string {
	lp.mutex.RLock()
	entryCount := len(lp.entries)
	lp.mutex.RUnlock()
	
	// Left side: Title and collapse indicator
	leftSide := fmt.Sprintf("▼ Logs (%d)", entryCount)
	
	// Right side: Status
	rightSide := lp.getStatusString()
	
	// Calculate spacing
	leftWidth := lipgloss.Width(leftSide)
	rightWidth := lipgloss.Width(rightSide)
	availableWidth := lp.Width - 4 // Account for border and padding
	spacingWidth := availableWidth - leftWidth - rightWidth
	
	var spacing string
	if spacingWidth > 0 {
		spacing = strings.Repeat(" ", spacingWidth)
	}
	
	headerContent := leftSide + spacing + rightSide
	return lp.HeaderStyle.Render(headerContent)
}

// getStatusString returns the current status string
func (lp *LogPane) getStatusString() string {
	var status []string
	
	if lp.Paused {
		status = append(status, "PAUSED")
	}
	
	if !lp.AutoScroll {
		status = append(status, "NO-SCROLL")
	}
	
	status = append(status, lp.currentLevel.String())
	
	return strings.Join(status, " ")
}

// addScrollBar adds a scroll bar to the content
func (lp *LogPane) addScrollBar(content string) string {
	if !lp.ShowScrollBar {
		return content
	}
	
	lines := strings.Split(content, "\n")
	
	// Calculate scroll bar
	lp.mutex.RLock()
	totalEntries := len(lp.entries)
	lp.mutex.RUnlock()
	
	visibleLines := lp.Model.Height
	scrollTop := lp.Model.YOffset
	
	var scrollBar []string
	for i := 0; i < len(lines); i++ {
		scrollPos := i + scrollTop
		char := " "
		
		if totalEntries > visibleLines {
			scrollRatio := float64(scrollPos) / float64(totalEntries-visibleLines)
			barHeight := visibleLines
			barPos := int(scrollRatio * float64(barHeight-1))
			
			if i == barPos {
				char = "█"
			} else if scrollPos < totalEntries {
				char = "│"
			}
		}
		
		scrollBar = append(scrollBar, lp.ScrollBarStyle.Render(char))
	}
	
	// Combine content with scroll bar
	var result strings.Builder
	for i, line := range lines {
		result.WriteString(line)
		if i < len(scrollBar) {
			result.WriteString(scrollBar[i])
		}
		
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// buildContent creates the current viewport content string (for View() calls)
func (lp *LogPane) buildContent() string {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	var content strings.Builder
	
	for _, entry := range lp.entries {
		// Filter by log level
		if entry.Level < lp.currentLevel {
			continue
		}
		
		// Apply source filters if any
		if len(lp.filters) > 0 && !lp.filters[entry.Source] {
			continue
		}
		
		// Format and style the entry
		formattedEntry := lp.formatLogEntry(entry)
		content.WriteString(formattedEntry)
		content.WriteString("\n")
	}
	
	contentStr := content.String()
	if contentStr == "" {
		contentStr = "No log entries available"
	}
	
	return contentStr
}

// refreshContent should only be called from Update() - sets viewport content
func (lp *LogPane) refreshContent() {
	// Safety check: don't refresh if viewport isn't ready
	if !lp.Ready {
		return
	}
	
	contentStr := lp.buildContent()
	
	// CRITICAL: Only call SetContent in Update(), never in View()
	lp.Model.SetContent(contentStr)
	lp.Model.SetYOffset(0)
}

// formatLogEntry formats a log entry with appropriate styling
func (lp *LogPane) formatLogEntry(entry LogEntry) string {
	// Get base formatted string
	formatted := entry.Format()
	
	// DEBUG: Use bright colors to test visibility (disable styles temporarily)
	switch entry.Level {
	case LogLevelError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(formatted) // Bright red
	case LogLevelWarn:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(formatted) // Bright yellow
	case LogLevelInfo:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(formatted) // Bright green
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(formatted) // White
	}
}

// GetEffectiveHeight returns the current effective height
func (lp *LogPane) GetEffectiveHeight() int {
	if !lp.Visible {
		return 0
	}
	if lp.Collapsed {
		return 3 // Minimum height for collapsed header
	}
	return lp.Height
}

// SetShowScrollBar toggles scroll bar display
func (lp *LogPane) SetShowScrollBar(show bool) {
	lp.ShowScrollBar = show
	lp.SetDimensions(lp.Width, lp.Height) // Recalculate viewport dimensions
}

// SetShowHeader toggles header display
func (lp *LogPane) SetShowHeader(show bool) {
	lp.ShowHeader = show
	lp.SetDimensions(lp.Width, lp.Height) // Recalculate viewport dimensions
}

// GetLogCount returns the current number of log entries
func (lp *LogPane) GetLogCount() int {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	return len(lp.entries)
}
// GetEntryCount returns the number of log entries  
func (lp *LogPane) GetEntryCount() int {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	return len(lp.entries)
}
