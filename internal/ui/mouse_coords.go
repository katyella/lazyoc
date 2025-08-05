package ui

import (
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/logging"
)

// ClickType represents the type of mouse click target
type ClickType int

const (
	ClickTab ClickType = iota
	ClickResource
	ClickPanel
	ClickUnhandled
)

// ClickTarget represents what was clicked
type ClickTarget struct {
	Type          ClickType
	Panel         int
	TabIndex      int
	ResourceIndex int
}

// MouseCoordinator handles mouse coordinate calculations
type MouseCoordinator struct {
	tui *TUI
}

// NewMouseCoordinator creates a new MouseCoordinator instance
func NewMouseCoordinator(tui *TUI) *MouseCoordinator {
	return &MouseCoordinator{tui: tui}
}

// GetClickTarget determines what was clicked based on coordinates
func (m *MouseCoordinator) GetClickTarget(x, y int) ClickTarget {
	headerHeight := m.tui.getHeaderHeight()

	logging.Debug(m.tui.Logger, "MouseCoordinator: analyzing click at X=%d, Y=%d, headerHeight=%d",
		x, y, headerHeight)

	// Check if click is on tabs row
	if y == headerHeight {
		tabIndex := m.calculateTabIndex(x)
		logging.Debug(m.tui.Logger, "MouseCoordinator: tab click detected, tabIndex=%d", tabIndex)
		return ClickTarget{Type: ClickTab, TabIndex: tabIndex}
	}

	// Determine which panel was clicked
	panel := m.getPanelFromCoordinates(x, y)

	// Check if click is in actual resource rows (not headers)
	resourceListStartY := headerHeight + 1
	contentHeaderLines := m.getContentHeaderLines()
	firstResourceY := resourceListStartY + contentHeaderLines
	if panel == 0 && y >= firstResourceY {
		resourceIndex := m.calculateResourceIndex(y)
		// Validate that the resource index is within bounds and not a separator/header click
		if resourceIndex >= 0 && m.isValidResourceClick(resourceIndex) {
			logging.Debug(m.tui.Logger, "MouseCoordinator: resource click detected, panel=%d, resourceIndex=%d",
				panel, resourceIndex)
			return ClickTarget{Type: ClickResource, Panel: panel, ResourceIndex: resourceIndex}
		} else {
			logging.Debug(m.tui.Logger, "MouseCoordinator: invalid resource click, resourceIndex=%d", resourceIndex)
		}
	}

	// Generic panel click
	if panel >= 0 {
		logging.Debug(m.tui.Logger, "MouseCoordinator: panel click detected, panel=%d", panel)
		return ClickTarget{Type: ClickPanel, Panel: panel}
	}

	logging.Debug(m.tui.Logger, "MouseCoordinator: unhandled click")
	return ClickTarget{Type: ClickUnhandled}
}

// calculateTabIndex determines which tab was clicked based on x coordinate
func (m *MouseCoordinator) calculateTabIndex(x int) int {
	tabs := constants.ResourceTabs

	// Calculate actual tab positions accounting for padding and centering
	var tabWidths []int
	totalTabsWidth := 0

	for _, tab := range tabs {
		// Each tab has padding of 1 on each side, so width = len(name) + 2
		tabWidth := len(tab) + 2
		tabWidths = append(tabWidths, tabWidth)
		totalTabsWidth += tabWidth
	}

	// Calculate starting position (center-aligned)
	startX := (m.tui.width - totalTabsWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Find which tab was clicked
	currentX := startX

	for i, tabWidth := range tabWidths {
		if x >= currentX && x < currentX+tabWidth {
			return i
		}
		currentX += tabWidth
	}

	// Return -1 if click is outside tab area
	return -1
}

// calculateResourceIndex determines which resource was clicked based on y coordinate
func (m *MouseCoordinator) calculateResourceIndex(y int) int {
	headerHeight := m.tui.getHeaderHeight()
	resourceListStartY := headerHeight + 1 // header + tabs

	// Different resource types have different content header structures
	contentHeaderLines := m.getContentHeaderLines()
	firstResourceY := resourceListStartY + contentHeaderLines
	resourceIndex := y - firstResourceY

	logging.Debug(m.tui.Logger, "MouseCoordinator: calculateResourceIndex Y=%d, resourceListStartY=%d, contentHeaderLines=%d, firstResourceY=%d, resourceIndex=%d, headerHeight=%d, activeTab=%d",
		y, resourceListStartY, contentHeaderLines, firstResourceY, resourceIndex, headerHeight, int(m.tui.ActiveTab))

	return resourceIndex
}

// getContentHeaderLines returns the number of header lines for the current resource type
func (m *MouseCoordinator) getContentHeaderLines() int {
	// All resource types follow the same pattern:
	// blank space
	// blank space
	// Line 2: "📦 Resource Name\n"
	// Line 3: "\n" (empty line)
	// Line 4: "NAME ... AGE\n" (header)
	// Line 6: "────────\n" (separator)
	// Line 7+: actual resource data

	return 6 // All tabs have exactly 4 header lines before resource data
}

// getPanelFromCoordinates determines which panel was clicked based on coordinates
func (m *MouseCoordinator) getPanelFromCoordinates(x, y int) int {
	// If no details or logs shown, everything is main panel
	if !m.tui.showDetails && !m.tui.showLogs {
		return 0
	}

	// Calculate main panel width
	mainWidth := m.tui.width
	if m.tui.showDetails {
		mainWidth = int(float64(m.tui.width) * constants.MainPanelWidthRatio)
	}

	// Check horizontal position for main vs details
	if m.tui.showDetails && x >= mainWidth {
		return 1 // Details panel
	}

	// Check vertical position for logs
	if m.tui.showLogs {
		// Calculate where logs panel starts
		// This is a heuristic - bottom 1/3 of screen
		logStartY := m.tui.height - (m.tui.height / 3)
		if y >= logStartY {
			return 2 // Logs panel
		}
	}

	return 0 // Main panel
}

// isValidResourceClick validates that the resource index corresponds to an actual resource
func (m *MouseCoordinator) isValidResourceClick(resourceIndex int) bool {
	switch m.tui.ActiveTab {
	case 0: // Pods
		return resourceIndex >= 0 && resourceIndex < len(m.tui.pods)
	case 1: // Services
		return resourceIndex >= 0 && resourceIndex < len(m.tui.services)
	case 2: // Deployments
		return resourceIndex >= 0 && resourceIndex < len(m.tui.deployments)
	case 3: // ConfigMaps
		return resourceIndex >= 0 && resourceIndex < len(m.tui.configMaps)
	case 4: // Secrets
		return resourceIndex >= 0 && resourceIndex < len(m.tui.secrets)
	case 5: // BuildConfigs
		return resourceIndex >= 0 && resourceIndex < len(m.tui.buildConfigs)
	case 6: // ImageStreams
		return resourceIndex >= 0 && resourceIndex < len(m.tui.imageStreams)
	case 7: // Routes
		return resourceIndex >= 0 && resourceIndex < len(m.tui.routes)
	default:
		return false
	}
}
