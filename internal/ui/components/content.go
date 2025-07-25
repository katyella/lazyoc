package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ContentPane represents the main scrollable content area
type ContentPane struct {
	viewport.Model
	Width  int
	Height int

	// Content management
	Content     string
	Title       string
	ItemCount   int
	SelectedRow int

	// Display options
	ShowLineNumbers bool
	ShowScrollBar   bool
	ShowTitle       bool
	WrapText        bool

	// Styling
	TitleStyle     lipgloss.Style
	ContentStyle   lipgloss.Style
	SelectedStyle  lipgloss.Style
	LineNumStyle   lipgloss.Style
	ScrollBarStyle lipgloss.Style

	// State
	Focused bool
}

// NewContentPane creates a new scrollable content pane
func NewContentPane(width, height int) *ContentPane {
	vp := viewport.New(width-2, height-2) // Account for borders
	vp.HighPerformanceRendering = true

	return &ContentPane{
		Model:       vp,
		Width:       width,
		Height:      height,
		Title:       "",
		ItemCount:   0,
		SelectedRow: 0,

		ShowLineNumbers: false,
		ShowScrollBar:   true,
		ShowTitle:       true,
		WrapText:        true,

		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Padding(0, 1),

		ContentStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),

		SelectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("12")).
			Bold(true),

		LineNumStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Width(4).
			Align(lipgloss.Right),

		ScrollBarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),

		Focused: false,
	}
}

// SetContent sets the content to display
func (cp *ContentPane) SetContent(content string) {
	cp.Content = content
	cp.Model.SetContent(content)
	cp.calculateItemCount()
}

// AppendContent appends content to the existing content
func (cp *ContentPane) AppendContent(content string) {
	if cp.Content != "" {
		cp.Content += "\n" + content
	} else {
		cp.Content = content
	}
	cp.Model.SetContent(cp.Content)
	cp.calculateItemCount()
}

// SetTitle sets the pane title
func (cp *ContentPane) SetTitle(title string) {
	cp.Title = title
}

// SetDimensions updates the pane dimensions
func (cp *ContentPane) SetDimensions(width, height int) {
	cp.Width = width
	cp.Height = height

	// Update viewport size accounting for borders and title
	vpHeight := height - 2 // Borders
	if cp.ShowTitle && cp.Title != "" {
		vpHeight -= 1 // Title
	}

	vpWidth := width - 2 // Borders
	if cp.ShowScrollBar {
		vpWidth -= 1 // Scroll bar
	}

	cp.Model.Width = vpWidth
	cp.Model.Height = vpHeight
}

// SetFocus sets the focus state
func (cp *ContentPane) SetFocus(focused bool) {
	cp.Focused = focused
}

// Update handles Bubble Tea messages
func (cp *ContentPane) Update(msg tea.Msg) (*ContentPane, tea.Cmd) {
	var cmd tea.Cmd
	cp.Model, cmd = cp.Model.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cp.Focused {
			switch msg.String() {
			case "j", "down":
				if cp.SelectedRow < cp.ItemCount-1 {
					cp.SelectedRow++
					cp.scrollToSelected()
				}
			case "k", "up":
				if cp.SelectedRow > 0 {
					cp.SelectedRow--
					cp.scrollToSelected()
				}
			case "g":
				cp.SelectedRow = 0
				cp.Model.GotoTop()
			case "G":
				cp.SelectedRow = cp.ItemCount - 1
				cp.Model.GotoBottom()
			case "ctrl+d":
				cp.Model.HalfViewDown()
				cp.updateSelectedFromScroll()
			case "ctrl+u":
				cp.Model.HalfViewUp()
				cp.updateSelectedFromScroll()
			case "ctrl+f":
				cp.Model.ViewDown()
				cp.updateSelectedFromScroll()
			case "ctrl+b":
				cp.Model.ViewUp()
				cp.updateSelectedFromScroll()
			}
		}
	}

	return cp, cmd
}

// Render renders the content pane
func (cp *ContentPane) Render() string {
	// Build the content area
	var content strings.Builder

	// Add title if enabled
	if cp.ShowTitle && cp.Title != "" {
		titleLine := cp.TitleStyle.Render(cp.Title)
		titleWidth := lipgloss.Width(titleLine)

		// Pad title to full width
		if titleWidth < cp.Width-2 {
			padding := strings.Repeat(" ", cp.Width-2-titleWidth)
			titleLine += padding
		}

		content.WriteString(titleLine)
		content.WriteString("\n")
	}

	// Get viewport content
	viewportContent := cp.renderViewportContent()
	content.WriteString(viewportContent)

	// Apply border styling
	borderStyle := lipgloss.NewStyle().
		Width(cp.Width).
		Height(cp.Height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8"))

	if cp.Focused {
		borderStyle = borderStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12"))
	}

	return borderStyle.Render(content.String())
}

// renderViewportContent renders the viewport content with optional enhancements
func (cp *ContentPane) renderViewportContent() string {
	baseContent := cp.Model.View()

	// Add line numbers if enabled
	if cp.ShowLineNumbers {
		baseContent = cp.addLineNumbers(baseContent)
	}

	// Add scroll bar if enabled
	if cp.ShowScrollBar {
		baseContent = cp.addScrollBar(baseContent)
	}

	return baseContent
}

// addLineNumbers adds line numbers to the content
func (cp *ContentPane) addLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for i, line := range lines {
		lineNum := cp.LineNumStyle.Render(fmt.Sprintf("%3d", i+1))

		// Highlight selected line
		if i == cp.SelectedRow && cp.Focused {
			line = cp.SelectedStyle.Render(line)
		}

		result.WriteString(lineNum)
		result.WriteString(" ")
		result.WriteString(line)

		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// addScrollBar adds a scroll bar to the content
func (cp *ContentPane) addScrollBar(content string) string {
	lines := strings.Split(content, "\n")

	// Calculate scroll bar position
	totalLines := strings.Count(cp.Content, "\n") + 1
	visibleLines := cp.Model.Height
	scrollTop := cp.Model.YOffset

	var scrollBar []string
	for i := 0; i < len(lines); i++ {
		scrollPos := i + scrollTop
		char := " "

		if totalLines > visibleLines {
			// Calculate if this position should show scroll indicator
			scrollRatio := float64(scrollPos) / float64(totalLines-visibleLines)
			barHeight := visibleLines
			barPos := int(scrollRatio * float64(barHeight-1))

			if i == barPos {
				char = "█"
			} else if scrollPos < totalLines {
				char = "│"
			}
		}

		scrollBar = append(scrollBar, cp.ScrollBarStyle.Render(char))
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

// calculateItemCount calculates the number of items/lines in the content
func (cp *ContentPane) calculateItemCount() {
	if cp.Content == "" {
		cp.ItemCount = 0
	} else {
		cp.ItemCount = strings.Count(cp.Content, "\n") + 1
	}
}

// scrollToSelected scrolls the viewport to show the selected item
func (cp *ContentPane) scrollToSelected() {
	// Convert selected row to viewport position
	if cp.SelectedRow < cp.Model.YOffset {
		// Selected item is above visible area
		cp.Model.SetYOffset(cp.SelectedRow)
	} else if cp.SelectedRow >= cp.Model.YOffset+cp.Model.Height {
		// Selected item is below visible area
		cp.Model.SetYOffset(cp.SelectedRow - cp.Model.Height + 1)
	}
}

// updateSelectedFromScroll updates the selected row based on scroll position
func (cp *ContentPane) updateSelectedFromScroll() {
	// Keep selected row within visible area
	minRow := cp.Model.YOffset
	maxRow := cp.Model.YOffset + cp.Model.Height - 1

	if cp.SelectedRow < minRow {
		cp.SelectedRow = minRow
	} else if cp.SelectedRow > maxRow {
		cp.SelectedRow = maxRow
	}

	// Ensure selected row is within content bounds
	if cp.SelectedRow >= cp.ItemCount {
		cp.SelectedRow = cp.ItemCount - 1
	}
	if cp.SelectedRow < 0 {
		cp.SelectedRow = 0
	}
}

// GetSelectedRow returns the currently selected row
func (cp *ContentPane) GetSelectedRow() int {
	return cp.SelectedRow
}

// SetSelectedRow sets the selected row
func (cp *ContentPane) SetSelectedRow(row int) {
	if row >= 0 && row < cp.ItemCount {
		cp.SelectedRow = row
		cp.scrollToSelected()
	}
}

// ClearContent clears all content
func (cp *ContentPane) ClearContent() {
	cp.Content = ""
	cp.Model.SetContent("")
	cp.ItemCount = 0
	cp.SelectedRow = 0
}

// SetShowLineNumbers toggles line number display
func (cp *ContentPane) SetShowLineNumbers(show bool) {
	cp.ShowLineNumbers = show
}

// SetShowScrollBar toggles scroll bar display
func (cp *ContentPane) SetShowScrollBar(show bool) {
	cp.ShowScrollBar = show
	cp.SetDimensions(cp.Width, cp.Height) // Recalculate dimensions
}

// SetShowTitle toggles title display
func (cp *ContentPane) SetShowTitle(show bool) {
	cp.ShowTitle = show
	cp.SetDimensions(cp.Width, cp.Height) // Recalculate dimensions
}

// GetScrollPercentage returns the current scroll percentage
func (cp *ContentPane) GetScrollPercentage() float64 {
	if cp.ItemCount <= cp.Model.Height {
		return 1.0 // Everything is visible
	}

	return float64(cp.Model.YOffset) / float64(cp.ItemCount-cp.Model.Height)
}

// IsAtTop returns whether the viewport is at the top
func (cp *ContentPane) IsAtTop() bool {
	return cp.Model.AtTop()
}

// IsAtBottom returns whether the viewport is at the bottom
func (cp *ContentPane) IsAtBottom() bool {
	return cp.Model.AtBottom()
}
