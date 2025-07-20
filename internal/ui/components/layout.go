package components

import (
	"github.com/charmbracelet/lipgloss"
)

// LayoutManager handles the arrangement and styling of UI panels
type LayoutManager struct {
	Width  int
	Height int
	
	// Layout configuration
	HeaderHeight    int
	TabsHeight      int
	StatusBarHeight int
	MinPanelWidth   int
	MinPanelHeight  int
	
	// Panel visibility and sizing
	DetailPaneVisible bool
	LogPaneVisible    bool
	DetailPaneWidth   int
	LogPaneHeight     int
	
	// Focus management
	FocusedPanel PanelType
}

// PanelType represents different UI panels
type PanelType int

const (
	PanelHeader PanelType = iota
	PanelTabs
	PanelMain
	PanelDetail
	PanelLog
	PanelStatusBar
)

// PanelDimensions holds the calculated dimensions for a panel
type PanelDimensions struct {
	X      int
	Y      int
	Width  int
	Height int
}

// LayoutConfig holds configuration for layout behavior
type LayoutConfig struct {
	MinTerminalWidth  int
	MinTerminalHeight int
	HeaderHeight      int
	TabsHeight        int
	StatusBarHeight   int
	DetailPaneRatio   float64 // 0.0 to 1.0
	LogPaneRatio      float64 // 0.0 to 1.0
	MinPanelWidth     int
	MinPanelHeight    int
}

// DefaultLayoutConfig returns sensible defaults for the layout
func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		MinTerminalWidth:  60, // Reduced for better small screen support
		MinTerminalHeight: 15, // Reduced for better small screen support
		HeaderHeight:      2,  // Reduced from 3 to save space
		TabsHeight:        1,  // Reduced from 2 to save space
		StatusBarHeight:   1,
		DetailPaneRatio:   0.3,  // 30% of width
		LogPaneRatio:      0.25, // 25% of height
		MinPanelWidth:     15,   // Reduced for small screens
		MinPanelHeight:    3,    // Reduced for small screens
	}
}

// NewLayoutManager creates a new layout manager with default configuration
func NewLayoutManager(width, height int) *LayoutManager {
	config := DefaultLayoutConfig()
	
	return &LayoutManager{
		Width:           width,
		Height:          height,
		HeaderHeight:    config.HeaderHeight,
		TabsHeight:      config.TabsHeight,
		StatusBarHeight: config.StatusBarHeight,
		MinPanelWidth:   config.MinPanelWidth,
		MinPanelHeight:  config.MinPanelHeight,
		
		DetailPaneVisible: true,
		LogPaneVisible:    true,
		DetailPaneWidth:   int(float64(width) * config.DetailPaneRatio),
		LogPaneHeight:     int(float64(height) * config.LogPaneRatio),
		
		FocusedPanel: PanelMain,
	}
}

// UpdateDimensions updates the layout manager's dimensions and recalculates panel sizes
func (lm *LayoutManager) UpdateDimensions(width, height int) {
	lm.Width = width
	lm.Height = height
	
	// Apply responsive adjustments for small screens first
	lm.applyResponsiveAdjustments()
	
	// Recalculate panel sizes based on ratios
	config := DefaultLayoutConfig()
	if lm.DetailPaneVisible {
		lm.DetailPaneWidth = int(float64(width) * config.DetailPaneRatio)
	}
	if lm.LogPaneVisible {
		lm.LogPaneHeight = int(float64(height) * config.LogPaneRatio)
	}
	
	// Ensure minimum sizes
	if lm.DetailPaneWidth < lm.MinPanelWidth {
		lm.DetailPaneWidth = lm.MinPanelWidth
	}
	if lm.LogPaneHeight < lm.MinPanelHeight {
		lm.LogPaneHeight = lm.MinPanelHeight
	}
}

// applyResponsiveAdjustments adjusts layout for small terminal sizes
func (lm *LayoutManager) applyResponsiveAdjustments() {
	config := DefaultLayoutConfig()
	
	// For very small terminals, reduce component heights
	if lm.Height < 20 {
		lm.HeaderHeight = 1
		lm.TabsHeight = 1
		lm.StatusBarHeight = 1
		
		// Hide detail pane on very small screens
		if lm.Width < 80 {
			lm.DetailPaneVisible = false
		}
		
		// Reduce or hide log pane on very small screens
		if lm.Height < 15 {
			lm.LogPaneVisible = false
		}
	} else {
		// Restore default values for larger screens
		lm.HeaderHeight = config.HeaderHeight
		lm.TabsHeight = config.TabsHeight
		lm.StatusBarHeight = config.StatusBarHeight
	}
}

// GetPanelDimensions calculates the dimensions for a specific panel
func (lm *LayoutManager) GetPanelDimensions(panel PanelType) PanelDimensions {
	switch panel {
	case PanelHeader:
		return PanelDimensions{
			X:      0,
			Y:      0,
			Width:  lm.Width,
			Height: lm.HeaderHeight,
		}
		
	case PanelTabs:
		return PanelDimensions{
			X:      0,
			Y:      lm.HeaderHeight,
			Width:  lm.Width,
			Height: lm.TabsHeight,
		}
		
	case PanelMain:
		x := 0
		y := lm.HeaderHeight + lm.TabsHeight
		width := lm.Width
		height := lm.Height - lm.HeaderHeight - lm.TabsHeight - lm.StatusBarHeight
		
		// Adjust for detail pane if visible
		if lm.DetailPaneVisible {
			width -= lm.DetailPaneWidth
		}
		
		// Adjust for log pane if visible
		if lm.LogPaneVisible {
			height -= lm.LogPaneHeight
		}
		
		return PanelDimensions{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		}
		
	case PanelDetail:
		if !lm.DetailPaneVisible {
			return PanelDimensions{}
		}
		
		x := lm.Width - lm.DetailPaneWidth
		y := lm.HeaderHeight + lm.TabsHeight
		height := lm.Height - lm.HeaderHeight - lm.TabsHeight - lm.StatusBarHeight
		
		// Adjust for log pane if visible
		if lm.LogPaneVisible {
			height -= lm.LogPaneHeight
		}
		
		return PanelDimensions{
			X:      x,
			Y:      y,
			Width:  lm.DetailPaneWidth,
			Height: height,
		}
		
	case PanelLog:
		if !lm.LogPaneVisible {
			return PanelDimensions{}
		}
		
		x := 0
		y := lm.Height - lm.StatusBarHeight - lm.LogPaneHeight
		width := lm.Width
		
		// Adjust for detail pane if visible
		if lm.DetailPaneVisible {
			width -= lm.DetailPaneWidth
		}
		
		return PanelDimensions{
			X:      x,
			Y:      y,
			Width:  width,
			Height: lm.LogPaneHeight,
		}
		
	case PanelStatusBar:
		return PanelDimensions{
			X:      0,
			Y:      lm.Height - lm.StatusBarHeight,
			Width:  lm.Width,
			Height: lm.StatusBarHeight,
		}
		
	default:
		return PanelDimensions{}
	}
}

// ToggleDetailPane toggles the visibility of the detail pane
func (lm *LayoutManager) ToggleDetailPane() {
	lm.DetailPaneVisible = !lm.DetailPaneVisible
}

// ToggleLogPane toggles the visibility of the log pane
func (lm *LayoutManager) ToggleLogPane() {
	lm.LogPaneVisible = !lm.LogPaneVisible
}

// SetFocus sets the currently focused panel
func (lm *LayoutManager) SetFocus(panel PanelType) {
	lm.FocusedPanel = panel
}

// GetFocus returns the currently focused panel
func (lm *LayoutManager) GetFocus() PanelType {
	return lm.FocusedPanel
}

// IsMinimumSize checks if the terminal meets minimum size requirements
func (lm *LayoutManager) IsMinimumSize() bool {
	config := DefaultLayoutConfig()
	return lm.Width >= config.MinTerminalWidth && lm.Height >= config.MinTerminalHeight
}

// GetBorderStyle returns the border style for a panel based on focus
func (lm *LayoutManager) GetBorderStyle(panel PanelType) lipgloss.Border {
	if lm.FocusedPanel == panel {
		return lipgloss.RoundedBorder()
	}
	return lipgloss.NormalBorder()
}

// GetBorderColor returns the border color for a panel based on focus
func (lm *LayoutManager) GetBorderColor(panel PanelType) lipgloss.Color {
	if lm.FocusedPanel == panel {
		return lipgloss.Color("12") // Blue for focused
	}
	return lipgloss.Color("8") // Gray for unfocused
}

// StylePanel applies consistent styling to a panel
func (lm *LayoutManager) StylePanel(content string, panel PanelType) string {
	dims := lm.GetPanelDimensions(panel)
	
	style := lipgloss.NewStyle().
		Width(dims.Width).
		Height(dims.Height).
		Border(lm.GetBorderStyle(panel)).
		BorderForeground(lm.GetBorderColor(panel))
	
	return style.Render(content)
}