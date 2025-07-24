package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
)

// StatusBarComponent renders the bottom status bar
type StatusBarComponent struct {
	BaseComponent

	// Status content
	leftContent   string
	centerContent string
	rightContent  string
	notification  string
	errorMessage  string

	// Key hints
	keyHints []KeyHint
	
	// Styles
	normalStyle       lipgloss.Style
	keyStyle          lipgloss.Style
	descStyle         lipgloss.Style
	notificationStyle lipgloss.Style
	errorStyle        lipgloss.Style
	separatorStyle    lipgloss.Style
}

// KeyHint represents a keyboard hint
type KeyHint struct {
	Key         string
	Description string
}

// NewStatusBarComponent creates a new status bar component
func NewStatusBarComponent() *StatusBarComponent {
	return &StatusBarComponent{
		keyHints: []KeyHint{
			{Key: "?", Description: "help"},
			{Key: "tab", Description: "switch"},
			{Key: "↑↓", Description: "navigate"},
			{Key: "enter", Description: "select"},
			{Key: "q", Description: "quit"},
		},

		normalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorWhite)),

		keyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorCyan)).
			Bold(true),

		descStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorGray)),

		notificationStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorGreen)),

		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorRed)).
			Bold(true),

		separatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(constants.ColorDarkGray)),
	}
}

// Init initializes the status bar component
func (s *StatusBarComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status bar component
func (s *StatusBarComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case StatusUpdateMsg:
		s.leftContent = msg.Left
		s.centerContent = msg.Center
		s.rightContent = msg.Right
	
	case NotificationMsg:
		s.notification = msg.Message
		s.errorMessage = ""
		// Clear notification after timeout
		return tea.Tick(msg.Duration, func(time.Time) tea.Msg {
			return ClearNotificationMsg{}
		}), nil
	
	case ErrorMsg:
		s.errorMessage = msg.Message
		s.notification = ""
	
	case ClearNotificationMsg:
		s.notification = ""
	
	case UpdateKeyHintsMsg:
		s.keyHints = msg.Hints
	}
	
	return nil, nil
}

// View renders the status bar component
func (s *StatusBarComponent) View() string {
	if s.width == 0 {
		return ""
	}

	// Separator line
	separator := s.separatorStyle.Render(strings.Repeat("─", s.width))

	// Build status line
	var statusLine string
	
	// Priority: error > notification > normal status
	if s.errorMessage != "" {
		statusLine = s.renderErrorLine()
	} else if s.notification != "" {
		statusLine = s.renderNotificationLine()
	} else {
		statusLine = s.renderNormalStatus()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		separator,
		statusLine,
	)
}

// renderNormalStatus renders the normal status bar
func (s *StatusBarComponent) renderNormalStatus() string {
	// Build key hints
	hints := s.renderKeyHints()
	hintsWidth := lipgloss.Width(hints)

	// Build left content
	left := s.leftContent
	if left == "" {
		left = " " // Add padding
	} else {
		left = " " + left
	}
	leftWidth := lipgloss.Width(left)

	// Build right content
	right := s.rightContent
	if right != "" {
		right = right + " "
	}
	rightWidth := lipgloss.Width(right)

	// Calculate spacing for center content
	availableWidth := s.width - leftWidth - rightWidth - hintsWidth
	
	if availableWidth > 0 {
		// Enough space for all content
		padding := strings.Repeat(" ", availableWidth)
		return left + padding + right + hints
	} else if s.width > hintsWidth+10 {
		// Show only hints with some status
		return s.renderCompactStatus()
	} else {
		// Very narrow, show minimal hints
		return s.renderMinimalHints()
	}
}

// renderKeyHints renders the keyboard hints
func (s *StatusBarComponent) renderKeyHints() string {
	var hints []string
	for _, hint := range s.keyHints {
		key := s.keyStyle.Render(hint.Key)
		desc := s.descStyle.Render(hint.Description)
		hints = append(hints, fmt.Sprintf("%s %s", key, desc))
	}
	return strings.Join(hints, " • ")
}

// renderCompactStatus renders a compact version of the status bar
func (s *StatusBarComponent) renderCompactStatus() string {
	compactHints := []KeyHint{
		{Key: "?", Description: "help"},
		{Key: "tab", Description: "switch"},
		{Key: "q", Description: "quit"},
	}
	
	var hints []string
	for _, hint := range compactHints {
		hints = append(hints, fmt.Sprintf("%s %s", 
			s.keyStyle.Render(hint.Key),
			s.descStyle.Render(hint.Description),
		))
	}
	
	return " " + strings.Join(hints, " • ")
}

// renderMinimalHints renders minimal hints for very narrow terminals
func (s *StatusBarComponent) renderMinimalHints() string {
	return s.descStyle.Render(" Press ? for help")
}

// renderNotificationLine renders a notification message
func (s *StatusBarComponent) renderNotificationLine() string {
	icon := "✓"
	message := fmt.Sprintf(" %s %s", icon, s.notification)
	
	// Center the notification
	return lipgloss.Place(
		s.width, 1,
		lipgloss.Center, lipgloss.Center,
		s.notificationStyle.Render(message),
	)
}

// renderErrorLine renders an error message
func (s *StatusBarComponent) renderErrorLine() string {
	icon := "✗"
	message := fmt.Sprintf(" %s %s", icon, s.errorMessage)
	
	// Center the error
	return lipgloss.Place(
		s.width, 1,
		lipgloss.Center, lipgloss.Center,
		s.errorStyle.Render(message),
	)
}

// SetStatus updates the status bar content
func (s *StatusBarComponent) SetStatus(left, center, right string) {
	s.leftContent = left
	s.centerContent = center
	s.rightContent = right
}

// SetNotification displays a notification message
func (s *StatusBarComponent) SetNotification(message string) {
	s.notification = message
	s.errorMessage = ""
}

// SetError displays an error message
func (s *StatusBarComponent) SetError(message string) {
	s.errorMessage = message
	s.notification = ""
}

// Messages for status bar updates
type (
	// StatusUpdateMsg updates the status bar content
	StatusUpdateMsg struct {
		Left   string
		Center string
		Right  string
	}

	// NotificationMsg displays a temporary notification
	NotificationMsg struct {
		Message  string
		Duration time.Duration
	}

	// ErrorMsg displays an error message
	ErrorMsg struct {
		Message string
	}

	// ClearNotificationMsg clears the notification
	ClearNotificationMsg struct{}

	// UpdateKeyHintsMsg updates the keyboard hints
	UpdateKeyHintsMsg struct {
		Hints []KeyHint
	}
)