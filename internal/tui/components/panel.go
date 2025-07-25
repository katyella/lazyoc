package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
)

// PanelComponent is a base component for content panels
type PanelComponent struct {
	BaseComponent

	// Panel properties
	title       string
	content     string
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style

	// Scrolling
	scrollOffset int
	contentLines []string

	// Selection
	selectedIndex int
	selectable    bool
}

// NewPanelComponent creates a new panel component
func NewPanelComponent(title string) *PanelComponent {
	return &PanelComponent{
		title: title,

		borderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(constants.ColorGray)),

		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(constants.ColorBlue)),
	}
}

// Init initializes the panel component
func (p *PanelComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages for the panel component
func (p *PanelComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	if !p.focused {
		return nil, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.selectable {
				p.moveSelection(-1)
			} else {
				p.scroll(-1)
			}
		case "down", "j":
			if p.selectable {
				p.moveSelection(1)
			} else {
				p.scroll(1)
			}
		case "pgup":
			if p.selectable {
				p.moveSelection(-10)
			} else {
				p.scroll(-10)
			}
		case "pgdown":
			if p.selectable {
				p.moveSelection(10)
			} else {
				p.scroll(10)
			}
		case "home", "g":
			p.scrollOffset = 0
			p.selectedIndex = 0
		case "end", "G":
			if p.selectable && len(p.contentLines) > 0 {
				p.selectedIndex = len(p.contentLines) - 1
			} else if len(p.contentLines) > p.height-4 {
				p.scrollOffset = len(p.contentLines) - (p.height - 4)
			}
		}
	}

	return nil, nil
}

// View renders the panel component
func (p *PanelComponent) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	// Update border style based on focus
	borderStyle := p.borderStyle.
		Width(p.width).
		Height(p.height)

	if p.focused {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color(constants.ColorBlue))
	}

	// Prepare content
	content := p.prepareContent()

	// For now, just render the content with border
	// Title handling can be added later with a newer lipgloss version
	return borderStyle.Render(content)
}

// SetContent sets the panel content
func (p *PanelComponent) SetContent(content string) {
	p.content = content
	p.contentLines = strings.Split(content, "\n")

	// Adjust scroll if needed
	p.adjustScroll()
}

// SetContentLines sets the panel content as lines
func (p *PanelComponent) SetContentLines(lines []string) {
	p.contentLines = lines
	p.content = strings.Join(lines, "\n")

	// Adjust scroll if needed
	p.adjustScroll()
}

// EnableSelection enables item selection in the panel
func (p *PanelComponent) EnableSelection() {
	p.selectable = true
}

// GetSelectedIndex returns the currently selected index
func (p *PanelComponent) GetSelectedIndex() int {
	return p.selectedIndex
}

// SetSelectedIndex sets the selected index
func (p *PanelComponent) SetSelectedIndex(index int) {
	if index >= 0 && index < len(p.contentLines) {
		p.selectedIndex = index
		p.ensureSelectedVisible()
	}
}

// prepareContent prepares the content for rendering
func (p *PanelComponent) prepareContent() string {
	if len(p.contentLines) == 0 {
		return p.content
	}

	// Calculate visible area
	visibleHeight := p.height - 4 // Account for borders and padding
	if visibleHeight <= 0 {
		return ""
	}

	// Determine visible lines
	startLine := p.scrollOffset
	endLine := startLine + visibleHeight
	if endLine > len(p.contentLines) {
		endLine = len(p.contentLines)
	}

	// Build visible content
	var lines []string
	for i := startLine; i < endLine; i++ {
		line := p.contentLines[i]

		// Highlight selected line if selection is enabled
		if p.selectable && i == p.selectedIndex {
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(constants.ColorDarkGray)).
				Width(p.width - 4) // Account for borders and padding
			line = selectedStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Add scroll indicators
	if p.scrollOffset > 0 {
		lines[0] = "↑ " + lines[0]
	}
	if endLine < len(p.contentLines) {
		lastIdx := len(lines) - 1
		lines[lastIdx] = lines[lastIdx] + " ↓"
	}

	return strings.Join(lines, "\n")
}

// scroll adjusts the scroll offset
func (p *PanelComponent) scroll(delta int) {
	p.scrollOffset += delta
	p.adjustScroll()
}

// moveSelection moves the selection
func (p *PanelComponent) moveSelection(delta int) {
	p.selectedIndex += delta

	// Clamp to valid range
	if p.selectedIndex < 0 {
		p.selectedIndex = 0
	} else if p.selectedIndex >= len(p.contentLines) {
		p.selectedIndex = len(p.contentLines) - 1
	}

	// Ensure selected item is visible
	p.ensureSelectedVisible()
}

// adjustScroll ensures scroll offset is within valid range
func (p *PanelComponent) adjustScroll() {
	maxScroll := len(p.contentLines) - (p.height - 4)
	if maxScroll < 0 {
		maxScroll = 0
	}

	if p.scrollOffset < 0 {
		p.scrollOffset = 0
	} else if p.scrollOffset > maxScroll {
		p.scrollOffset = maxScroll
	}
}

// ensureSelectedVisible adjusts scroll to keep selection visible
func (p *PanelComponent) ensureSelectedVisible() {
	visibleHeight := p.height - 4

	// Scroll up if selection is above visible area
	if p.selectedIndex < p.scrollOffset {
		p.scrollOffset = p.selectedIndex
	}

	// Scroll down if selection is below visible area
	if p.selectedIndex >= p.scrollOffset+visibleHeight {
		p.scrollOffset = p.selectedIndex - visibleHeight + 1
	}

	p.adjustScroll()
}
