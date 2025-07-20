package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailPane represents a collapsible detail pane for showing resource information
type DetailPane struct {
	viewport.Model
	Width     int
	Height    int
	MinWidth  int
	MaxWidth  int
	
	// State management
	Visible   bool
	Collapsed bool
	Focused   bool
	
	// Content
	Title       string
	Content     string
	ResourceType string
	ResourceName string
	LastUpdated  time.Time
	
	// Display options
	ShowHeader    bool
	ShowScrollBar bool
	ShowTimestamp bool
	WrapContent   bool
	
	// Animation/transition
	CollapseDuration time.Duration
	IsAnimating     bool
	
	// Styling
	HeaderStyle      lipgloss.Style
	TitleStyle       lipgloss.Style
	ContentStyle     lipgloss.Style
	CollapsedStyle   lipgloss.Style
	BorderStyle      lipgloss.Style
	TimestampStyle   lipgloss.Style
	ScrollBarStyle   lipgloss.Style
}

// NewDetailPane creates a new detail pane
func NewDetailPane(width, height int) *DetailPane {
	vp := viewport.New(width-2, height-3) // Account for border and header
	vp.HighPerformanceRendering = true
	
	return &DetailPane{
		Model:      vp,
		Width:      width,
		Height:     height,
		MinWidth:   20,
		MaxWidth:   width / 2,
		
		Visible:    true,
		Collapsed:  false,
		Focused:    false,
		
		Title:       "Details",
		Content:     "",
		ResourceType: "",
		ResourceName: "",
		LastUpdated: time.Now(),
		
		ShowHeader:    true,
		ShowScrollBar: true,
		ShowTimestamp: true,
		WrapContent:   true,
		
		CollapseDuration: 200 * time.Millisecond,
		IsAnimating:     false,
		
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Padding(0, 1),
			
		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true),
			
		ContentStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1),
			
		CollapsedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Align(lipgloss.Center),
			
		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")),
			
		TimestampStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Align(lipgloss.Right),
			
		ScrollBarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}

// SetVisible sets the visibility of the detail pane
func (dp *DetailPane) SetVisible(visible bool) {
	dp.Visible = visible
}

// IsVisible returns whether the detail pane is visible
func (dp *DetailPane) IsVisible() bool {
	return dp.Visible
}

// Toggle toggles the visibility of the detail pane
func (dp *DetailPane) Toggle() {
	dp.Visible = !dp.Visible
}

// Collapse collapses the detail pane to show only the header
func (dp *DetailPane) Collapse() {
	dp.Collapsed = true
}

// Expand expands the detail pane to show full content
func (dp *DetailPane) Expand() {
	dp.Collapsed = false
}

// ToggleCollapse toggles the collapsed state
func (dp *DetailPane) ToggleCollapse() {
	dp.Collapsed = !dp.Collapsed
}

// IsCollapsed returns whether the detail pane is collapsed
func (dp *DetailPane) IsCollapsed() bool {
	return dp.Collapsed
}

// SetFocus sets the focus state
func (dp *DetailPane) SetFocus(focused bool) {
	dp.Focused = focused
	
	// Update border style based on focus
	if focused {
		dp.BorderStyle = dp.BorderStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12"))
	} else {
		dp.BorderStyle = dp.BorderStyle.
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))
	}
}

// SetDimensions updates the pane dimensions
func (dp *DetailPane) SetDimensions(width, height int) {
	dp.Width = width
	dp.Height = height
	
	// Enforce width constraints
	if width < dp.MinWidth {
		dp.Width = dp.MinWidth
	} else if width > dp.MaxWidth {
		dp.Width = dp.MaxWidth
	}
	
	// Update viewport size
	vpHeight := height - 2 // Border
	if dp.ShowHeader {
		vpHeight -= 1 // Header
	}
	
	vpWidth := width - 2 // Border
	if dp.ShowScrollBar {
		vpWidth -= 1 // Scroll bar
	}
	
	dp.Model.Width = vpWidth
	dp.Model.Height = vpHeight
}

// SetMinMaxWidth sets the minimum and maximum widths
func (dp *DetailPane) SetMinMaxWidth(min, max int) {
	dp.MinWidth = min
	dp.MaxWidth = max
}

// SetResourceInfo sets the resource information to display
func (dp *DetailPane) SetResourceInfo(resourceType, resourceName string, content string) {
	dp.ResourceType = resourceType
	dp.ResourceName = resourceName
	dp.Content = content
	dp.LastUpdated = time.Now()
	
	// Update title
	if resourceName != "" {
		dp.Title = fmt.Sprintf("%s: %s", resourceType, resourceName)
	} else {
		dp.Title = resourceType
	}
	
	// Update viewport content
	dp.Model.SetContent(dp.formatContent())
}

// SetContent sets the detail content directly
func (dp *DetailPane) SetContent(content string) {
	dp.Content = content
	dp.LastUpdated = time.Now()
	dp.Model.SetContent(dp.formatContent())
}

// AppendContent appends content to the existing content
func (dp *DetailPane) AppendContent(content string) {
	if dp.Content != "" {
		dp.Content += "\n" + content
	} else {
		dp.Content = content
	}
	dp.LastUpdated = time.Now()
	dp.Model.SetContent(dp.formatContent())
}

// ClearContent clears all content
func (dp *DetailPane) ClearContent() {
	dp.Content = ""
	dp.ResourceType = ""
	dp.ResourceName = ""
	dp.Title = "Details"
	dp.Model.SetContent("")
}

// Update handles Bubble Tea messages
func (dp *DetailPane) Update(msg tea.Msg) (*DetailPane, tea.Cmd) {
	if !dp.Visible || dp.Collapsed {
		return dp, nil
	}
	
	var cmd tea.Cmd
	dp.Model, cmd = dp.Model.Update(msg)
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if dp.Focused {
			switch msg.String() {
			case "c":
				dp.ToggleCollapse()
			case "h":
				dp.SetVisible(false)
			case "r":
				// Refresh content (placeholder for future refresh functionality)
				dp.LastUpdated = time.Now()
			}
		}
	}
	
	return dp, cmd
}

// Render renders the detail pane
func (dp *DetailPane) Render() string {
	if !dp.Visible {
		return ""
	}
	
	if dp.Collapsed {
		return dp.renderCollapsed()
	}
	
	return dp.renderExpanded()
}

// renderCollapsed renders the collapsed state (header only)
func (dp *DetailPane) renderCollapsed() string {
	collapsedContent := dp.CollapsedStyle.Render("▶ " + dp.Title + " (collapsed)")
	
	style := dp.BorderStyle.
		Width(dp.Width).
		Height(3) // Minimum height for collapsed state
	
	return style.Render(collapsedContent)
}

// renderExpanded renders the expanded state with full content
func (dp *DetailPane) renderExpanded() string {
	var content strings.Builder
	
	// Header
	if dp.ShowHeader {
		header := dp.renderHeader()
		content.WriteString(header)
		content.WriteString("\n")
	}
	
	// Content area
	viewportContent := dp.Model.View()
	if dp.ShowScrollBar {
		viewportContent = dp.addScrollBar(viewportContent)
	}
	content.WriteString(viewportContent)
	
	// Apply border
	style := dp.BorderStyle.
		Width(dp.Width).
		Height(dp.Height)
	
	return style.Render(content.String())
}

// renderHeader renders the header section
func (dp *DetailPane) renderHeader() string {
	// Left side: Title and collapse indicator
	leftSide := "▼ " + dp.TitleStyle.Render(dp.Title)
	
	// Right side: Timestamp (if enabled)
	var rightSide string
	if dp.ShowTimestamp {
		rightSide = dp.TimestampStyle.Render(dp.LastUpdated.Format("15:04:05"))
	}
	
	// Calculate spacing
	leftWidth := lipgloss.Width(leftSide)
	rightWidth := lipgloss.Width(rightSide)
	availableWidth := dp.Width - 4 // Account for border and padding
	spacingWidth := availableWidth - leftWidth - rightWidth
	
	var spacing string
	if spacingWidth > 0 {
		spacing = strings.Repeat(" ", spacingWidth)
	}
	
	headerContent := leftSide + spacing + rightSide
	return dp.HeaderStyle.Render(headerContent)
}

// addScrollBar adds a scroll bar to the content
func (dp *DetailPane) addScrollBar(content string) string {
	if !dp.ShowScrollBar {
		return content
	}
	
	lines := strings.Split(content, "\n")
	
	// Calculate scroll bar
	totalLines := strings.Count(dp.Content, "\n") + 1
	visibleLines := dp.Model.Height
	scrollTop := dp.Model.YOffset
	
	var scrollBar []string
	for i := 0; i < len(lines); i++ {
		scrollPos := i + scrollTop
		char := " "
		
		if totalLines > visibleLines {
			scrollRatio := float64(scrollPos) / float64(totalLines-visibleLines)
			barHeight := visibleLines
			barPos := int(scrollRatio * float64(barHeight-1))
			
			if i == barPos {
				char = "█"
			} else if scrollPos < totalLines {
				char = "│"
			}
		}
		
		scrollBar = append(scrollBar, dp.ScrollBarStyle.Render(char))
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

// formatContent formats the content for display
func (dp *DetailPane) formatContent() string {
	if dp.Content == "" {
		return "No details available.\n\nSelect a resource to view its details here."
	}
	
	content := dp.Content
	
	// Add wrap formatting if enabled
	if dp.WrapContent {
		content = dp.wrapText(content, dp.Model.Width-2)
	}
	
	return content
}

// wrapText wraps text to the specified width
func (dp *DetailPane) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	lines := strings.Split(text, "\n")
	var wrappedLines []string
	
	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}
		
		// Wrap long lines
		for len(line) > width {
			// Find the last space within the width limit
			wrapIndex := width
			if spaceIndex := strings.LastIndex(line[:width], " "); spaceIndex != -1 {
				wrapIndex = spaceIndex
			}
			
			wrappedLines = append(wrappedLines, line[:wrapIndex])
			line = line[wrapIndex:]
			
			// Remove leading space from continuation line
			line = strings.TrimLeft(line, " ")
		}
		
		// Add remaining part
		if len(line) > 0 {
			wrappedLines = append(wrappedLines, line)
		}
	}
	
	return strings.Join(wrappedLines, "\n")
}

// GetEffectiveWidth returns the current effective width (0 if collapsed or invisible)
func (dp *DetailPane) GetEffectiveWidth() int {
	if !dp.Visible || dp.Collapsed {
		return 0
	}
	return dp.Width
}

// GetEffectiveHeight returns the current effective height
func (dp *DetailPane) GetEffectiveHeight() int {
	if !dp.Visible {
		return 0
	}
	if dp.Collapsed {
		return 3 // Minimum height for collapsed header
	}
	return dp.Height
}

// SetShowScrollBar toggles scroll bar display
func (dp *DetailPane) SetShowScrollBar(show bool) {
	dp.ShowScrollBar = show
	dp.SetDimensions(dp.Width, dp.Height) // Recalculate viewport dimensions
}

// SetShowHeader toggles header display
func (dp *DetailPane) SetShowHeader(show bool) {
	dp.ShowHeader = show
	dp.SetDimensions(dp.Width, dp.Height) // Recalculate viewport dimensions
}

// SetShowTimestamp toggles timestamp display
func (dp *DetailPane) SetShowTimestamp(show bool) {
	dp.ShowTimestamp = show
}

// GetScrollPercentage returns the current scroll percentage
func (dp *DetailPane) GetScrollPercentage() float64 {
	if dp.Collapsed || !dp.Visible {
		return 0.0
	}
	
	totalLines := strings.Count(dp.Content, "\n") + 1
	if totalLines <= dp.Model.Height {
		return 1.0
	}
	
	return float64(dp.Model.YOffset) / float64(totalLines-dp.Model.Height)
}