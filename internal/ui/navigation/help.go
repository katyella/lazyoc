package navigation

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/ui/components"
)

// HelpComponent represents a help overlay with keybinding information
type HelpComponent struct {
	Width  int
	Height int

	// Content
	Title       string
	HelpText    string
	CurrentMode NavigationMode

	// Display options
	ShowBorder   bool
	ShowModeInfo bool
	ShowQuickRef bool

	// Styling
	TitleStyle       lipgloss.Style
	HeaderStyle      lipgloss.Style
	ContentStyle     lipgloss.Style
	BorderStyle      lipgloss.Style
	ModeStyle        lipgloss.Style
	KeyStyle         lipgloss.Style
	DescriptionStyle lipgloss.Style
	CategoryStyle    lipgloss.Style
}

// NewHelpComponent creates a new help overlay component
func NewHelpComponent(width, height int) *HelpComponent {
	return &HelpComponent{
		Width:  width,
		Height: height,

		Title:       "LazyOC Help",
		CurrentMode: ModeNormal,

		ShowBorder:   true,
		ShowModeInfo: true,
		ShowQuickRef: true,

		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Align(lipgloss.Center).
			Margin(0, 0, 1, 0),

		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true).
			Underline(true).
			Margin(0, 0, 1, 0),

		ContentStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(1, 2),

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1),

		ModeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Background(lipgloss.Color("8")).
			Padding(0, 1).
			Margin(0, 0, 1, 0),

		KeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")).
			Bold(true).
			Width(12),

		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),

		CategoryStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true).
			Underline(true).
			Margin(1, 0, 0, 0),
	}
}

// SetDimensions updates the help component dimensions
func (h *HelpComponent) SetDimensions(width, height int) {
	h.Width = width
	h.Height = height
}

// SetCurrentMode updates the current navigation mode for context-sensitive help
func (h *HelpComponent) SetCurrentMode(mode NavigationMode) {
	h.CurrentMode = mode
}

// Render renders the help overlay
func (h *HelpComponent) Render(registry *KeybindingRegistry) string {
	// Always ensure we have valid dimensions
	if h.Width < 20 || h.Height < 8 {
		// Terminal too small, show minimal help
		return h.renderMinimalHelp()
	}

	// Calculate available space more conservatively
	availableHeight := h.Height - 4 // Border + padding (reduced from 6)
	availableWidth := h.Width - 6   // Border + padding (reduced from 8)

	// Ensure we have minimum space
	if availableHeight < 6 || availableWidth < 30 {
		// Terminal too small, show minimal help
		return h.renderMinimalHelp()
	}

	var content strings.Builder
	linesUsed := 0

	// Title (2 lines with spacing)
	if linesUsed < availableHeight-2 {
		title := h.TitleStyle.Width(availableWidth).Render("📖 " + h.Title)
		content.WriteString(title)
		content.WriteString("\n\n")
		linesUsed += 2
	}

	// Current mode indicator (2 lines with spacing)
	if h.ShowModeInfo && linesUsed < availableHeight-2 {
		modeText := fmt.Sprintf("Current Mode: %s", registry.GetModeIndicator())
		modeIndicator := h.ModeStyle.Render(modeText)
		content.WriteString(modeIndicator)
		content.WriteString("\n\n")
		linesUsed += 2
	}

	// Calculate remaining height for help sections
	remainingHeight := availableHeight - linesUsed - 1 // Save 1 line for close instruction

	if remainingHeight > 5 {
		// Show abbreviated help content that fits
		helpContent := h.generateConstrainedHelpContent(registry, remainingHeight-2)
		content.WriteString(helpContent)
		linesUsed += strings.Count(helpContent, "\n") + 1
	}

	// Close instruction
	if linesUsed < availableHeight {
		content.WriteString("\nPress ? or ESC to close help")
	}

	// CRITICAL: Ensure content doesn't exceed available space
	contentStr := content.String()
	contentLines := strings.Split(contentStr, "\n")

	// Truncate content if it exceeds available height
	if len(contentLines) > availableHeight {
		contentLines = contentLines[:availableHeight-1]
		contentLines = append(contentLines, "... (content truncated)")
		contentStr = strings.Join(contentLines, "\n")
	}

	// Apply content styling with strict constraints
	styledContent := h.ContentStyle.
		Width(availableWidth).
		MaxWidth(availableWidth).
		Height(availableHeight).
		MaxHeight(availableHeight).
		Render(contentStr)

	if h.ShowBorder {
		// Ensure border doesn't exceed terminal dimensions
		return h.BorderStyle.
			Width(h.Width).
			MaxWidth(h.Width).
			Height(h.Height).
			MaxHeight(h.Height).
			Render(styledContent)
	}

	return styledContent
}

// renderMinimalHelp renders a minimal help for very small terminals
func (h *HelpComponent) renderMinimalHelp() string {
	helpText := "📖 Help: q=quit ?=close esc=cancel hjkl=move tab=panels"

	// Truncate text if terminal is extremely small
	if h.Width < 50 {
		helpText = "📖 q=quit ?=close esc=cancel"
	}
	if h.Width < 30 {
		helpText = "q=quit ?=close"
	}

	style := lipgloss.NewStyle().
		Width(h.Width).
		MaxWidth(h.Width).
		Height(h.Height).
		MaxHeight(h.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12"))

	return style.Render(helpText)
}

// generateConstrainedHelpContent generates help content that fits within height limit
func (h *HelpComponent) generateConstrainedHelpContent(registry *KeybindingRegistry, maxHeight int) string {
	var content strings.Builder
	linesUsed := 0

	// Prioritize most important bindings for constrained space
	priorityBindings := h.getPriorityBindings(registry)

	// Show quick essential keys first
	if linesUsed < maxHeight-1 {
		content.WriteString(h.CategoryStyle.Render("Essential Keys"))
		content.WriteString("\n")
		linesUsed += 2

		for _, binding := range priorityBindings {
			if linesUsed >= maxHeight-1 {
				break
			}

			keyPart := h.KeyStyle.Render(binding.Key)
			descPart := h.DescriptionStyle.Render(binding.Description)

			line := lipgloss.JoinHorizontal(lipgloss.Left,
				"  ", keyPart, " ", descPart)
			content.WriteString(line)
			content.WriteString("\n")
			linesUsed++
		}
	}

	return content.String()
}

// getPriorityBindings returns the most important keybindings for quick reference
func (h *HelpComponent) getPriorityBindings(registry *KeybindingRegistry) []KeyBinding {
	allBindings := registry.GetAllBindings()
	var priority []KeyBinding

	// Essential keys in order of importance
	essentialKeys := []string{"q", "?", "esc", "h", "j", "k", "l", "tab", "enter", "r"}

	for _, key := range essentialKeys {
		if binding, exists := allBindings[key]; exists {
			priority = append(priority, binding)
		}
	}

	return priority
}


// GetContextualHints returns contextual keybinding hints for the status bar
func (h *HelpComponent) GetContextualHints(registry *KeybindingRegistry, currentPanel components.PanelType) []string {
	var hints []string

	mode := registry.GetMode()

	switch mode {
	case ModeNormal:
		hints = []string{
			"? help",
			"hjkl move",
			"tab panels",
			"/ search",
			"q quit",
		}

		// Add panel-specific hints
		switch currentPanel {
		case components.PanelMain:
			hints = append(hints, "enter select")
		case components.PanelLog:
			hints = append(hints, "C clear", "p pause")
		}

	case ModeSearch:
		hints = []string{
			"enter search",
			"esc cancel",
		}

	case ModeCommand:
		hints = []string{
			"enter execute",
			"esc cancel",
		}

	case ModeInsert:
		hints = []string{
			"esc normal",
		}
	}

	return hints
}

// FormatHints formats hints for display in the status bar
func (h *HelpComponent) FormatHints(hints []string) string {
	if len(hints) == 0 {
		return ""
	}

	var formattedHints []string
	for _, hint := range hints {
		formattedHints = append(formattedHints,
			h.KeyStyle.Copy().Width(0).Render(hint))
	}

	return strings.Join(formattedHints, " • ")
}
