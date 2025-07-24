package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/tui/components"
)

// LogsView handles the logs panel
type LogsView struct {
	panel        *components.PanelComponent
	resourceType string
	resourceName string
	container    string
	logs         []LogEntry
	follow       bool
	showTimestamp bool
	style        LogsStyle
}

// LogEntry represents a single log line
type LogEntry struct {
	Timestamp time.Time
	Line      string
	Level     LogLevel
}

// LogLevel represents the log level
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
	LogLevelDebug
)

// LogsStyle contains styles for logs rendering
type LogsStyle struct {
	headerStyle    lipgloss.Style
	timestampStyle lipgloss.Style
	infoStyle      lipgloss.Style
	warnStyle      lipgloss.Style
	errorStyle     lipgloss.Style
	debugStyle     lipgloss.Style
}

// NewLogsView creates a new logs view
func NewLogsView() *LogsView {
	return &LogsView{
		panel:         components.NewPanelComponent("Logs"),
		logs:          make([]LogEntry, 0),
		follow:        true,
		showTimestamp: true,
		style: LogsStyle{
			headerStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(constants.ColorCyan)),
			timestampStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorDarkGray)),
			infoStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorWhite)),
			warnStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorYellow)),
			errorStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorRed)),
			debugStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(constants.ColorGray)),
		},
	}
}

// Init initializes the logs view
func (v *LogsView) Init() tea.Cmd {
	return v.panel.Init()
}

// Update handles messages for the logs view
func (v *LogsView) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case StartLogsMsg:
		v.StartLogs(msg.Type, msg.Name, msg.Container)
		return nil, nil
		
	case LogsReceivedMsg:
		v.AppendLogs(msg.Entries)
		return nil, nil
		
	case StopLogsMsg:
		v.Clear()
		return nil, nil
		
	case tea.KeyMsg:
		if v.panel.IsFocused() {
			switch msg.String() {
			case "f":
				v.follow = !v.follow
				v.updateContent()
			case "t":
				v.showTimestamp = !v.showTimestamp
				v.updateContent()
			case "c":
				v.Clear()
			}
		}
	}
	
	return v.panel.Update(msg)
}

// View renders the logs view
func (v *LogsView) View() string {
	return v.panel.View()
}

// StartLogs starts showing logs for a resource
func (v *LogsView) StartLogs(resourceType, name, container string) {
	v.resourceType = resourceType
	v.resourceName = name
	v.container = container
	v.logs = make([]LogEntry, 0)
	v.updateContent()
}

// AppendLogs adds new log entries
func (v *LogsView) AppendLogs(entries []LogEntry) {
	v.logs = append(v.logs, entries...)
	
	// Keep only last N entries to prevent memory issues
	maxLogs := 1000
	if len(v.logs) > maxLogs {
		v.logs = v.logs[len(v.logs)-maxLogs:]
	}
	
	v.updateContent()
}

// Clear clears the logs view
func (v *LogsView) Clear() {
	v.resourceType = ""
	v.resourceName = ""
	v.container = ""
	v.logs = make([]LogEntry, 0)
	v.panel.SetContent("No logs to display")
}

// updateContent updates the panel content with logs
func (v *LogsView) updateContent() {
	if v.resourceType == "" || v.resourceName == "" {
		v.panel.SetContent("No logs to display")
		return
	}
	
	var lines []string
	
	// Header
	header := fmt.Sprintf("Logs: %s/%s", v.resourceType, v.resourceName)
	if v.container != "" {
		header += fmt.Sprintf(" [%s]", v.container)
	}
	if v.follow {
		header += " (following)"
	}
	lines = append(lines, v.style.headerStyle.Render(header))
	lines = append(lines, "")
	
	// Status line
	status := fmt.Sprintf("Lines: %d | Follow: %v | Timestamps: %v",
		len(v.logs), v.follow, v.showTimestamp)
	lines = append(lines, v.style.timestampStyle.Render(status))
	lines = append(lines, "")
	
	// Logs
	if len(v.logs) == 0 {
		lines = append(lines, v.style.infoStyle.Render("Waiting for logs..."))
	} else {
		for _, entry := range v.logs {
			lines = append(lines, v.formatLogEntry(entry))
		}
	}
	
	v.panel.SetContentLines(lines)
	
	// Auto-scroll to bottom if following
	if v.follow && len(v.logs) > 0 {
		v.panel.SetSelectedIndex(len(lines) - 1)
	}
}

// formatLogEntry formats a single log entry
func (v *LogsView) formatLogEntry(entry LogEntry) string {
	var parts []string
	
	// Timestamp
	if v.showTimestamp {
		ts := entry.Timestamp.Format("15:04:05.000")
		parts = append(parts, v.style.timestampStyle.Render(ts))
	}
	
	// Choose style based on log level
	var style lipgloss.Style
	switch entry.Level {
	case LogLevelWarn:
		style = v.style.warnStyle
	case LogLevelError:
		style = v.style.errorStyle
	case LogLevelDebug:
		style = v.style.debugStyle
	default:
		style = v.style.infoStyle
	}
	
	// Log line
	parts = append(parts, style.Render(entry.Line))
	
	return strings.Join(parts, " ")
}

// DetectLogLevel attempts to detect the log level from the log line
func DetectLogLevel(line string) LogLevel {
	lower := strings.ToLower(line)
	
	if strings.Contains(lower, "error") || strings.Contains(lower, "err") ||
		strings.Contains(lower, "fatal") || strings.Contains(lower, "panic") {
		return LogLevelError
	}
	
	if strings.Contains(lower, "warn") || strings.Contains(lower, "warning") {
		return LogLevelWarn
	}
	
	if strings.Contains(lower, "debug") || strings.Contains(lower, "trace") {
		return LogLevelDebug
	}
	
	return LogLevelInfo
}

// Component interface implementation
func (v *LogsView) Focus() error     { return v.panel.Focus() }
func (v *LogsView) Blur() error      { return v.panel.Blur() }
func (v *LogsView) IsFocused() bool  { return v.panel.IsFocused() }
func (v *LogsView) SetSize(w, h int) { v.panel.SetSize(w, h) }
func (v *LogsView) GetSize() (int, int) { return v.panel.GetSize() }

// Messages
type StartLogsMsg struct {
	Type      string
	Name      string
	Container string
}

type LogsReceivedMsg struct {
	Entries []LogEntry
}

type StopLogsMsg struct{}