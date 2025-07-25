package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/ui/errors"
)

// ErrorDisplayComponent renders error messages with recovery options
type ErrorDisplayComponent struct {
	width    int
	height   int
	theme    string
	errors   []*errors.UserFriendlyError
	selected int // Selected recovery action

	// Styles
	containerStyle lipgloss.Style
}

// NewErrorDisplayComponent creates a new error display component
func NewErrorDisplayComponent(theme string) *ErrorDisplayComponent {
	return &ErrorDisplayComponent{
		theme:    theme,
		selected: 0,
		errors:   make([]*errors.UserFriendlyError, 0),
	}
}

// SetDimensions sets the component dimensions
func (e *ErrorDisplayComponent) SetDimensions(width, height int) {
	e.width = width
	e.height = height
	e.updateStyles()
}

// AddError adds an error to the display
func (e *ErrorDisplayComponent) AddError(err *errors.UserFriendlyError) {
	e.errors = append(e.errors, err)

	// Keep only the last 10 errors
	if len(e.errors) > 10 {
		e.errors = e.errors[1:]
	}
}

// ClearErrors clears all errors
func (e *ErrorDisplayComponent) ClearErrors() {
	e.errors = make([]*errors.UserFriendlyError, 0)
	e.selected = 0
}

// GetLatestError returns the most recent error
func (e *ErrorDisplayComponent) GetLatestError() *errors.UserFriendlyError {
	if len(e.errors) == 0 {
		return nil
	}
	return e.errors[len(e.errors)-1]
}

// HasErrors returns true if there are errors to display
func (e *ErrorDisplayComponent) HasErrors() bool {
	return len(e.errors) > 0
}

// RenderInline renders a compact error message for inline display
func (e *ErrorDisplayComponent) RenderInline() string {
	if !e.HasErrors() {
		return ""
	}

	latestError := e.GetLatestError()
	icon := latestError.GetIcon()

	// Determine style based on severity
	var style lipgloss.Style
	switch latestError.Severity {
	case errors.ErrorSeverityCritical:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	case errors.ErrorSeverityError:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
	case errors.ErrorSeverityWarning:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	case errors.ErrorSeverityInfo:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White
	}

	// Truncate message for inline display
	message := latestError.GetDisplayMessage()
	maxLen := e.width - 10 // Account for icon and padding
	if len(message) > maxLen && maxLen > 0 {
		message = message[:maxLen-3] + "..."
	}

	return style.Render(fmt.Sprintf("%s %s", icon, message))
}

// RenderModal renders a detailed error modal with recovery options
func (e *ErrorDisplayComponent) RenderModal() string {
	if !e.HasErrors() {
		return ""
	}

	latestError := e.GetLatestError()

	// Calculate modal dimensions - responsive with better limits
	maxModalWidth := 150  // Increased for better readability
	minModalWidth := 70   // For very narrow terminals
	targetFraction := 0.9 // Use 90% of screen width

	// Compute modal width
	modalWidth := int(float64(e.width) * targetFraction)
	if modalWidth < minModalWidth {
		modalWidth = minModalWidth
	}
	if modalWidth > maxModalWidth {
		modalWidth = maxModalWidth
	}
	// Final safety check
	if modalWidth > e.width-4 {
		modalWidth = e.width - 4
	}

	// Calculate modal height based on content
	baseHeight := 15 // Minimum height for title, message, and footer
	errorActions := len(errors.GetRecoveryActions(latestError))
	contentHeight := baseHeight + errorActions + 2 // Add space for actions

	// Add extra lines for technical details if present
	if latestError.TechnicalDetail != "" {
		contentHeight += 3
	}

	// Add extra lines for suggested action if present
	if latestError.GetSuggestedAction() != "" {
		contentHeight += 2
	}

	modalHeight := contentHeight
	if modalHeight > e.height-8 {
		modalHeight = e.height - 8
	}
	if modalHeight < 15 {
		modalHeight = 15
	}

	// Create modal container
	var borderColor lipgloss.Color
	switch latestError.Severity {
	case errors.ErrorSeverityCritical:
		borderColor = lipgloss.Color("9") // Red
	case errors.ErrorSeverityError:
		borderColor = lipgloss.Color("9") // Red
	case errors.ErrorSeverityWarning:
		borderColor = lipgloss.Color("11") // Yellow
	default:
		borderColor = lipgloss.Color("12") // Blue
	}

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Background(lipgloss.Color("236")). // Dark background
		Foreground(lipgloss.Color("15"))   // White text

	var content strings.Builder

	// Title with icon
	titleStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true).
		Width(modalWidth - 6) // Account for padding and borders

	content.WriteString(titleStyle.Render(fmt.Sprintf("%s %s", latestError.GetIcon(), latestError.Title)))
	content.WriteString("\n\n")

	// Message
	messageStyle := lipgloss.NewStyle().
		Width(modalWidth - 6). // Account for padding and borders
		Foreground(lipgloss.Color("15"))

	content.WriteString(messageStyle.Render(latestError.GetDisplayMessage()))
	content.WriteString("\n\n")

	// Technical details if available (collapsible)
	if latestError.TechnicalDetail != "" {
		detailStyle := lipgloss.NewStyle().
			Width(modalWidth - 6).             // Account for padding and borders
			Foreground(lipgloss.Color("242")). // Dimmer gray
			Italic(true)

		content.WriteString(detailStyle.Render("Technical details:"))
		content.WriteString("\n")

		// Truncate technical details if too long
		detail := latestError.TechnicalDetail
		maxDetailLen := (modalWidth - 6) * 2 // Max 2 lines to save space
		if len(detail) > maxDetailLen {
			detail = detail[:maxDetailLen-3] + "..."
		}

		content.WriteString(detailStyle.Render(detail))
		content.WriteString("\n\n")
	}

	// Recovery actions
	actions := errors.GetRecoveryActions(latestError)
	if len(actions) > 0 {
		actionHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")). // Cyan
			Bold(true)

		content.WriteString(actionHeaderStyle.Render("Recovery Options:"))
		content.WriteString("\n")

		for i, action := range actions {
			var actionStyle lipgloss.Style
			if i == e.selected {
				actionStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("0")).  // Black text
					Background(lipgloss.Color("14")). // Cyan background
					Bold(true).
					Padding(0, 1)
			} else {
				actionStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("15")).
					Padding(0, 1)
			}

			prefix := "  "
			if i == e.selected {
				prefix = "▶ "
			}

			actionText := fmt.Sprintf("%s%s", prefix, action.Name)
			if action.Description != "" {
				actionText += fmt.Sprintf(" - %s", action.Description)
			}

			content.WriteString(actionStyle.Render(actionText))
			content.WriteString("\n")
		}

		content.WriteString("\n")
	}

	// Suggested action
	if latestError.GetSuggestedAction() != "" {
		suggestionStyle := lipgloss.NewStyle().
			Width(modalWidth - 6).            // Account for padding and borders
			Foreground(lipgloss.Color("10")). // Green
			Italic(true)

		content.WriteString(suggestionStyle.Render("💡 " + latestError.GetSuggestedAction()))
		content.WriteString("\n\n")
	}

	// Footer with timestamp and controls
	footerStyle := lipgloss.NewStyle().
		Width(modalWidth - 6).             // Account for padding and borders
		Foreground(lipgloss.Color("242")). // Dim gray
		Align(lipgloss.Center)

	timestamp := latestError.Timestamp.Format("15:04:05")
	footer := fmt.Sprintf("Occurred at %s • Press 'esc' to dismiss • Use ↑↓ to select action • Enter to execute", timestamp)
	content.WriteString(footerStyle.Render(footer))

	return containerStyle.Render(content.String())
}

// MoveSelection moves the selection up or down in recovery actions
func (e *ErrorDisplayComponent) MoveSelection(direction int) {
	if !e.HasErrors() {
		return
	}

	latestError := e.GetLatestError()
	actions := errors.GetRecoveryActions(latestError)

	if len(actions) == 0 {
		return
	}

	e.selected += direction
	if e.selected < 0 {
		e.selected = len(actions) - 1
	} else if e.selected >= len(actions) {
		e.selected = 0
	}
}

// GetSelectedAction returns the currently selected recovery action
func (e *ErrorDisplayComponent) GetSelectedAction() *errors.RecoveryAction {
	if !e.HasErrors() {
		return nil
	}

	latestError := e.GetLatestError()
	actions := errors.GetRecoveryActions(latestError)

	if len(actions) == 0 || e.selected < 0 || e.selected >= len(actions) {
		return nil
	}

	return &actions[e.selected]
}

// updateStyles updates the component styles based on current theme and dimensions
func (e *ErrorDisplayComponent) updateStyles() {
	// This would be expanded with proper theme support
	e.containerStyle = lipgloss.NewStyle().
		Width(e.width).
		Height(e.height)
}
