package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/katyella/lazyoc/internal/constants"
)

// LayoutManager manages the overall screen layout
type LayoutManager struct {
	width  int
	height int

	// Layout configuration
	showDetails     bool
	showLogs        bool
	headerHeight    int
	statusBarHeight int
	tabBarHeight    int

	// Calculated dimensions
	contentHeight int
	mainWidth     int
	detailsWidth  int
	logsHeight    int
}

// NewLayoutManager creates a new layout manager
func NewLayoutManager() *LayoutManager {
	return &LayoutManager{
		showDetails:     true,
		showLogs:        true,
		headerHeight:    3, // Title + connection status + separator
		statusBarHeight: 2, // Status line + separator
		tabBarHeight:    2, // Tab bar + separator
	}
}

// SetSize updates the terminal dimensions
func (l *LayoutManager) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.recalculate()
}

// SetPanelVisibility updates which panels are visible
func (l *LayoutManager) SetPanelVisibility(showDetails, showLogs bool) {
	l.showDetails = showDetails
	l.showLogs = showLogs
	l.recalculate()
}

// recalculate updates all calculated dimensions
func (l *LayoutManager) recalculate() {
	// Calculate available content height
	l.contentHeight = l.height - l.headerHeight - l.statusBarHeight - l.tabBarHeight

	// Calculate panel widths
	if l.showDetails {
		l.mainWidth = int(float64(l.width) * constants.MainPanelWidthRatio)
		l.detailsWidth = l.width - l.mainWidth
	} else {
		l.mainWidth = l.width
		l.detailsWidth = 0
	}

	// Calculate logs height
	if l.showLogs {
		availableHeight := l.contentHeight
		l.logsHeight = int(float64(availableHeight) * constants.LogHeightRatio)
		if l.logsHeight < constants.DefaultLogHeight {
			l.logsHeight = constants.DefaultLogHeight
		}
		if l.logsHeight > availableHeight-constants.MinMainContentLines {
			l.logsHeight = availableHeight - constants.MinMainContentLines
		}
	} else {
		l.logsHeight = 0
	}
}

// GetDimensions returns the calculated dimensions for all panels
func (l *LayoutManager) GetDimensions() LayoutDimensions {
	return LayoutDimensions{
		Width:           l.width,
		Height:          l.height,
		HeaderHeight:    l.headerHeight,
		StatusBarHeight: l.statusBarHeight,
		TabBarHeight:    l.tabBarHeight,
		ContentHeight:   l.contentHeight,
		MainWidth:       l.mainWidth,
		DetailsWidth:    l.detailsWidth,
		LogsHeight:      l.logsHeight,
		MainHeight:      l.contentHeight - l.logsHeight,
	}
}

// LayoutDimensions contains all calculated layout dimensions
type LayoutDimensions struct {
	// Terminal dimensions
	Width  int
	Height int

	// Fixed heights
	HeaderHeight    int
	StatusBarHeight int
	TabBarHeight    int

	// Content area
	ContentHeight int

	// Panel dimensions
	MainWidth    int
	MainHeight   int
	DetailsWidth int
	LogsHeight   int
}

// RenderLayout combines components into the final layout
func RenderLayout(dimensions LayoutDimensions, header, tabs, main, details, logs, statusBar string) string {
	// Prepare content panels
	var content string

	if dimensions.DetailsWidth > 0 {
		// Side-by-side layout for main and details
		mainStyle := lipgloss.NewStyle().
			Width(dimensions.MainWidth).
			Height(dimensions.MainHeight)

		detailsStyle := lipgloss.NewStyle().
			Width(dimensions.DetailsWidth).
			Height(dimensions.MainHeight)

		topContent := lipgloss.JoinHorizontal(
			lipgloss.Top,
			mainStyle.Render(main),
			detailsStyle.Render(details),
		)

		if dimensions.LogsHeight > 0 {
			logsStyle := lipgloss.NewStyle().
				Width(dimensions.Width).
				Height(dimensions.LogsHeight)

			content = lipgloss.JoinVertical(
				lipgloss.Left,
				topContent,
				logsStyle.Render(logs),
			)
		} else {
			content = topContent
		}
	} else {
		// Full width main panel
		mainStyle := lipgloss.NewStyle().
			Width(dimensions.Width).
			Height(dimensions.MainHeight)

		if dimensions.LogsHeight > 0 {
			logsStyle := lipgloss.NewStyle().
				Width(dimensions.Width).
				Height(dimensions.LogsHeight)

			content = lipgloss.JoinVertical(
				lipgloss.Left,
				mainStyle.Render(main),
				logsStyle.Render(logs),
			)
		} else {
			content = mainStyle.Render(main)
		}
	}

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		statusBar,
	)
}

// CalculatePanelFocus determines which panel should receive focus based on coordinates
func CalculatePanelFocus(x, y int, dimensions LayoutDimensions) PanelType {
	// Check if in header or status bar
	if y < dimensions.HeaderHeight+dimensions.TabBarHeight {
		return PanelNone
	}
	if y >= dimensions.Height-dimensions.StatusBarHeight {
		return PanelNone
	}

	// Check if in logs area
	contentY := y - dimensions.HeaderHeight - dimensions.TabBarHeight
	if dimensions.LogsHeight > 0 && contentY >= dimensions.MainHeight {
		return PanelLogs
	}

	// Check main vs details
	if dimensions.DetailsWidth > 0 && x >= dimensions.MainWidth {
		return PanelDetails
	}

	return PanelMain
}

// PanelType represents the type of panel
type PanelType int

const (
	PanelNone PanelType = iota
	PanelMain
	PanelDetails
	PanelLogs
)

// String returns the string representation of the panel type
func (p PanelType) String() string {
	switch p {
	case PanelMain:
		return "main"
	case PanelDetails:
		return "details"
	case PanelLogs:
		return "logs"
	default:
		return "none"
	}
}
