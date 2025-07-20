package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// PanelConfig holds configuration for all panels
type PanelConfig struct {
	ShowBorders    bool
	FocusedStyle   lipgloss.Style
	UnfocusedStyle lipgloss.Style
	ContentPadding int
}

// DefaultPanelConfig returns default panel configuration
func DefaultPanelConfig() PanelConfig {
	return PanelConfig{
		ShowBorders: true,
		FocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(0, 1),
		UnfocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1),
		ContentPadding: 1,
	}
}

// FlexContainer represents a flexible container for arranging panels
type FlexContainer struct {
	Direction  FlexDirection
	Children   []FlexItem
	Width      int
	Height     int
	Gap        int
	Padding    int
	JustifyContent JustifyType
	AlignItems     AlignType
}

// FlexDirection defines the direction of flex container
type FlexDirection int

const (
	Row FlexDirection = iota
	Column
)

// JustifyType defines how items are justified in the flex container
type JustifyType int

const (
	JustifyStart JustifyType = iota
	JustifyEnd
	JustifyCenter
	JustifySpaceBetween
	JustifySpaceAround
)

// AlignType defines how items are aligned in the flex container
type AlignType int

const (
	AlignStart AlignType = iota
	AlignEnd
	AlignCenter
	AlignStretch
)

// FlexItem represents an item in a flex container
type FlexItem struct {
	Content    string
	FlexGrow   int
	FlexShrink int
	FlexBasis  int
	MinWidth   int
	MinHeight  int
	MaxWidth   int
	MaxHeight  int
	Padding    int
	Margin     int
	Style      lipgloss.Style
}

// NewFlexContainer creates a new flex container
func NewFlexContainer(direction FlexDirection, width, height int) *FlexContainer {
	return &FlexContainer{
		Direction:      direction,
		Children:       make([]FlexItem, 0),
		Width:          width,
		Height:         height,
		Gap:            1,
		Padding:        0,
		JustifyContent: JustifyStart,
		AlignItems:     AlignStart,
	}
}

// AddChild adds a child item to the flex container
func (fc *FlexContainer) AddChild(item FlexItem) {
	fc.Children = append(fc.Children, item)
}

// SetJustifyContent sets the justify-content property
func (fc *FlexContainer) SetJustifyContent(justify JustifyType) {
	fc.JustifyContent = justify
}

// SetAlignItems sets the align-items property
func (fc *FlexContainer) SetAlignItems(align AlignType) {
	fc.AlignItems = align
}

// SetGap sets the gap between items
func (fc *FlexContainer) SetGap(gap int) {
	fc.Gap = gap
}

// Render renders the flex container and its children
func (fc *FlexContainer) Render() string {
	if len(fc.Children) == 0 {
		return ""
	}

	// Calculate available space
	availableWidth := fc.Width - 2*fc.Padding
	availableHeight := fc.Height - 2*fc.Padding

	// Calculate sizes for each child
	childSizes := fc.calculateChildSizes(availableWidth, availableHeight)

	// Render children with calculated sizes
	renderedChildren := make([]string, len(fc.Children))
	for i, child := range fc.Children {
		size := childSizes[i]
		style := child.Style.
			Width(size.Width).
			Height(size.Height)
		
		if child.Padding > 0 {
			style = style.Padding(0, child.Padding)
		}
		
		renderedChildren[i] = style.Render(child.Content)
	}

	// Arrange children based on direction
	var result string
	if fc.Direction == Row {
		result = fc.arrangeRow(renderedChildren, availableWidth, availableHeight)
	} else {
		result = fc.arrangeColumn(renderedChildren, availableWidth, availableHeight)
	}

	// Apply container padding if needed
	if fc.Padding > 0 {
		containerStyle := lipgloss.NewStyle().
			Width(fc.Width).
			Height(fc.Height).
			Padding(fc.Padding)
		result = containerStyle.Render(result)
	}

	return result
}

// calculateChildSizes calculates the sizes for each child based on flex properties
func (fc *FlexContainer) calculateChildSizes(availableWidth, availableHeight int) []struct{ Width, Height int } {
	sizes := make([]struct{ Width, Height int }, len(fc.Children))
	
	if fc.Direction == Row {
		// Calculate widths for row layout
		totalFlexGrow := 0
		totalBasis := 0
		gapSpace := (len(fc.Children) - 1) * fc.Gap
		
		for _, child := range fc.Children {
			totalFlexGrow += child.FlexGrow
			if child.FlexBasis > 0 {
				totalBasis += child.FlexBasis
			}
		}
		
		remainingWidth := availableWidth - totalBasis - gapSpace
		
		for i, child := range fc.Children {
			width := child.FlexBasis
			if child.FlexGrow > 0 && totalFlexGrow > 0 {
				width += (remainingWidth * child.FlexGrow) / totalFlexGrow
			}
			
			// Apply constraints
			if child.MinWidth > 0 && width < child.MinWidth {
				width = child.MinWidth
			}
			if child.MaxWidth > 0 && width > child.MaxWidth {
				width = child.MaxWidth
			}
			
			height := availableHeight
			if child.MinHeight > 0 && height < child.MinHeight {
				height = child.MinHeight
			}
			if child.MaxHeight > 0 && height > child.MaxHeight {
				height = child.MaxHeight
			}
			
			sizes[i] = struct{ Width, Height int }{width, height}
		}
	} else {
		// Calculate heights for column layout
		totalFlexGrow := 0
		totalBasis := 0
		gapSpace := (len(fc.Children) - 1) * fc.Gap
		
		for _, child := range fc.Children {
			totalFlexGrow += child.FlexGrow
			if child.FlexBasis > 0 {
				totalBasis += child.FlexBasis
			}
		}
		
		remainingHeight := availableHeight - totalBasis - gapSpace
		
		for i, child := range fc.Children {
			height := child.FlexBasis
			if child.FlexGrow > 0 && totalFlexGrow > 0 {
				height += (remainingHeight * child.FlexGrow) / totalFlexGrow
			}
			
			// Apply constraints
			if child.MinHeight > 0 && height < child.MinHeight {
				height = child.MinHeight
			}
			if child.MaxHeight > 0 && height > child.MaxHeight {
				height = child.MaxHeight
			}
			
			width := availableWidth
			if child.MinWidth > 0 && width < child.MinWidth {
				width = child.MinWidth
			}
			if child.MaxWidth > 0 && width > child.MaxWidth {
				width = child.MaxWidth
			}
			
			sizes[i] = struct{ Width, Height int }{width, height}
		}
	}
	
	return sizes
}

// arrangeRow arranges children in a row
func (fc *FlexContainer) arrangeRow(children []string, availableWidth, availableHeight int) string {
	if len(children) == 0 {
		return ""
	}
	
	// Add gaps between children
	gapStyle := lipgloss.NewStyle().Width(fc.Gap)
	gap := gapStyle.Render("")
	
	result := children[0]
	for i := 1; i < len(children); i++ {
		result = lipgloss.JoinHorizontal(lipgloss.Top, result, gap, children[i])
	}
	
	return result
}

// arrangeColumn arranges children in a column
func (fc *FlexContainer) arrangeColumn(children []string, availableWidth, availableHeight int) string {
	if len(children) == 0 {
		return ""
	}
	
	// Add gaps between children
	if fc.Gap > 0 {
		gapStyle := lipgloss.NewStyle().Height(fc.Gap).Width(availableWidth)
		gap := gapStyle.Render("")
		
		result := children[0]
		for i := 1; i < len(children); i++ {
			result = lipgloss.JoinVertical(lipgloss.Left, result, gap, children[i])
		}
		return result
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, children...)
}

// CreateMainLayout creates the main application layout using flex containers
func CreateMainLayout(lm *LayoutManager) *FlexContainer {
	// Main vertical container
	mainContainer := NewFlexContainer(Column, lm.Width, lm.Height)
	mainContainer.SetGap(0)
	
	// Header (fixed height)
	headerDims := lm.GetPanelDimensions(PanelHeader)
	headerItem := FlexItem{
		Content:   CreateHeaderContent(lm),
		FlexBasis: headerDims.Height,
		Style:     lm.getHeaderStyle(),
	}
	mainContainer.AddChild(headerItem)
	
	// Tabs (fixed height)
	tabsDims := lm.GetPanelDimensions(PanelTabs)
	tabsItem := FlexItem{
		Content:   CreateTabsContent(lm),
		FlexBasis: tabsDims.Height,
		Style:     lm.getTabsStyle(),
	}
	mainContainer.AddChild(tabsItem)
	
	// Content area (flexible)
	contentContainer := NewFlexContainer(Row, lm.Width, 0)
	contentContainer.SetGap(1)
	
	// Main content pane
	mainItem := FlexItem{
		Content:   CreateMainContent(lm),
		FlexGrow:  1,
		MinWidth:  lm.MinPanelWidth,
		Style:     lm.getMainContentStyle(),
	}
	contentContainer.AddChild(mainItem)
	
	// Detail pane (if visible)
	if lm.DetailPaneVisible {
		detailDims := lm.GetPanelDimensions(PanelDetail)
		detailItem := FlexItem{
			Content:   CreateDetailPaneContent(lm),
			FlexBasis: detailDims.Width,
			MinWidth:  lm.MinPanelWidth,
			Style:     lm.getDetailPaneStyle(),
		}
		contentContainer.AddChild(detailItem)
	}
	
	// Render content container and add to main container
	contentItem := FlexItem{
		Content:  contentContainer.Render(),
		FlexGrow: 1,
	}
	mainContainer.AddChild(contentItem)
	
	// Log pane (if visible)
	if lm.LogPaneVisible {
		logDims := lm.GetPanelDimensions(PanelLog)
		logItem := FlexItem{
			Content:   CreateLogPaneContent(lm),
			FlexBasis: logDims.Height,
			Style:     lm.getLogPaneStyle(),
		}
		mainContainer.AddChild(logItem)
	}
	
	// Status bar (fixed height)
	statusDims := lm.GetPanelDimensions(PanelStatusBar)
	statusItem := FlexItem{
		Content:   CreateStatusBarContent(lm),
		FlexBasis: statusDims.Height,
		Style:     lm.getStatusBarStyle(),
	}
	mainContainer.AddChild(statusItem)
	
	return mainContainer
}

// Helper methods for LayoutManager to get styles for each panel
func (lm *LayoutManager) getHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(lm.Width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("12")).
		Bold(true)
}

func (lm *LayoutManager) getTabsStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(lm.Width).
		Align(lipgloss.Center)
}

func (lm *LayoutManager) getMainContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lm.GetBorderStyle(PanelMain)).
		BorderForeground(lm.GetBorderColor(PanelMain)).
		Padding(0, 1)
}

func (lm *LayoutManager) getDetailPaneStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lm.GetBorderStyle(PanelDetail)).
		BorderForeground(lm.GetBorderColor(PanelDetail)).
		Padding(0, 1)
}

func (lm *LayoutManager) getLogPaneStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lm.GetBorderStyle(PanelLog)).
		BorderForeground(lm.GetBorderColor(PanelLog)).
		Padding(0, 1)
}

func (lm *LayoutManager) getStatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(lm.Width).
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0"))
}

// Placeholder content creators (to be implemented with actual components)
func CreateHeaderContent(lm *LayoutManager) string {
	return "🚀 LazyOC - Kubernetes Resource Viewer"
}

func CreateTabsContent(lm *LayoutManager) string {
	return "[ Pods ] [ Services ] [ Deployments ] [ ConfigMaps ] [ Secrets ]"
}

func CreateMainContent(lm *LayoutManager) string {
	dims := lm.GetPanelDimensions(PanelMain)
	content := "Main Content Area\n\n"
	content += fmt.Sprintf("Dimensions: %dx%d\n", dims.Width, dims.Height)
	content += "Focus: Panel" + fmt.Sprintf("%d", int(lm.FocusedPanel)) + "\n"
	content += "\nSelect resources from the tabs above."
	return content
}

func CreateDetailPaneContent(lm *LayoutManager) string {
	if !lm.DetailPaneVisible {
		return ""
	}
	dims := lm.GetPanelDimensions(PanelDetail)
	content := "Detail Pane\n\n"
	content += fmt.Sprintf("Dimensions: %dx%d\n", dims.Width, dims.Height)
	content += "\nResource details will appear here."
	return content
}

func CreateLogPaneContent(lm *LayoutManager) string {
	if !lm.LogPaneVisible {
		return ""
	}
	dims := lm.GetPanelDimensions(PanelLog)
	content := "Log Pane\n\n"
	content += fmt.Sprintf("Dimensions: %dx%d\n", dims.Width, dims.Height)
	content += fmt.Sprintf("[%s] Application started\n", time.Now().Format("15:04:05"))
	content += "[15:04:06] Layout system initialized"
	return content
}

func CreateStatusBarContent(lm *LayoutManager) string {
	leftStatus := "Ready • Connected to cluster"
	rightStatus := fmt.Sprintf("Focus: Panel%d", int(lm.FocusedPanel))
	
	// Calculate spacing
	statusWidth := lm.Width - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus)
	spacing := ""
	if statusWidth > 0 {
		spacing = strings.Repeat(" ", statusWidth)
	}
	
	return leftStatus + spacing + rightStatus
}