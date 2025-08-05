package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/logging"
	"github.com/katyella/lazyoc/internal/ui/models"
)

// MouseHandler handles mouse events
type MouseHandler struct {
	tui          *TUI
	navigator    *Navigator
	focusManager *FocusManager
	coordinator  *MouseCoordinator
}

// NewMouseHandler creates a new MouseHandler instance
func NewMouseHandler(tui *TUI, navigator *Navigator, focusManager *FocusManager, coordinator *MouseCoordinator) *MouseHandler {
	return &MouseHandler{
		tui:          tui,
		navigator:    navigator,
		focusManager: focusManager,
		coordinator:  coordinator,
	}
}

// Handle processes mouse events
func (m *MouseHandler) Handle(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Ignore mouse events in modal states
	if m.tui.showHelp || m.tui.showErrorModal || m.tui.showProjectModal || m.tui.showSecretModal {
		return m.tui, nil
	}

	logging.Debug(m.tui.Logger, "MouseHandler: received event Type=%v, X=%d, Y=%d", msg.Type, msg.X, msg.Y)

	switch msg.Type {
	case tea.MouseLeft:
		return m.handleMouseClick(msg.X, msg.Y)
	case tea.MouseWheelUp:
		return m.handleMouseWheel(-1, msg.X, msg.Y)
	case tea.MouseWheelDown:
		return m.handleMouseWheel(1, msg.X, msg.Y)
	}

	return m.tui, nil
}

// handleMouseClick processes mouse click events
func (m *MouseHandler) handleMouseClick(x, y int) (tea.Model, tea.Cmd) {
	target := m.coordinator.GetClickTarget(x, y)
	
	switch target.Type {
	case ClickTab:
		return m.handleTabClick(target.TabIndex)
	case ClickResource:
		return m.handleResourceClick(target.ResourceIndex)
	case ClickPanel:
		return m.handlePanelClick(target.Panel)
	case ClickUnhandled:
		logging.Debug(m.tui.Logger, "MouseHandler: unhandled click")
	}
	
	return m.tui, nil
}

// handleTabClick processes clicks on tabs
func (m *MouseHandler) handleTabClick(tabIndex int) (tea.Model, tea.Cmd) {
	if tabIndex < 0 || tabIndex >= len(constants.ResourceTabs) {
		logging.Debug(m.tui.Logger, "MouseHandler: invalid tab index %d", tabIndex)
		return m.tui, nil
	}
	
	oldTab := m.tui.ActiveTab
	m.tui.ActiveTab = models.TabType(tabIndex)
	
	logging.Debug(m.tui.Logger, "MouseHandler: switched from tab %d to tab %d", 
		int(oldTab), tabIndex)
	
	// Call handleTabSwitch to update the content display
	return m.tui, m.tui.handleTabSwitch()
}

// handleResourceClick processes clicks on resource list items
func (m *MouseHandler) handleResourceClick(resourceIndex int) (tea.Model, tea.Cmd) {
	if !m.tui.connected {
		logging.Debug(m.tui.Logger, "MouseHandler: resource click ignored - not connected")
		return m.tui, nil
	}
	
	logging.Debug(m.tui.Logger, "MouseHandler: selecting resource %d in tab %d", 
		resourceIndex, int(m.tui.ActiveTab))
	
	m.navigator.SelectResource(resourceIndex)
	
	// For pods, handle log loading if needed
	if m.tui.ActiveTab == 0 && m.tui.logViewMode == constants.PodLogViewMode {
		m.tui.clearPodLogs()
		return m.tui, tea.Batch(m.tui.loadPodLogs(), m.tui.startPodLogRefreshTimer())
	}
	
	return m.tui, m.tui.loadPodLogs()
}

// handlePanelClick processes clicks that change panel focus
func (m *MouseHandler) handlePanelClick(panel int) (tea.Model, tea.Cmd) {
	if panel != m.focusManager.GetFocusedPanel() {
		logging.Debug(m.tui.Logger, "MouseHandler: switching focus to panel %d", panel)
		m.focusManager.FocusPanel(panel)
	}
	return m.tui, nil
}

// handleMouseWheel processes mouse wheel scroll events
func (m *MouseHandler) handleMouseWheel(direction int, x, y int) (tea.Model, tea.Cmd) {
	if !m.tui.connected {
		return m.tui, nil
	}
	
	// Determine which panel the wheel event is in
	panel := m.coordinator.getPanelFromCoordinates(x, y)
	
	logging.Debug(m.tui.Logger, "MouseHandler: wheel event in panel %d, direction=%d", panel, direction)
	
	switch panel {
	case 0: // Main panel - scroll resource list
		return m.handleResourceListScroll(direction)
	
	case 1: // Details panel - scroll details content
		if m.tui.showDetails {
			logging.Debug(m.tui.Logger, "MouseHandler: wheel in details panel (not implemented)")
			// Details scrolling can be implemented later
		}
	
	case 2: // Logs panel - scroll logs
		if m.tui.showLogs {
			return m.handleLogScroll(direction)
		}
	}
	
	return m.tui, nil
}

// handleResourceListScroll scrolls through resource lists
func (m *MouseHandler) handleResourceListScroll(direction int) (tea.Model, tea.Cmd) {
	if direction > 0 {
		m.navigator.SelectNextResource()
	} else {
		m.navigator.SelectPreviousResource()
	}
	
	// For pods, handle log loading if needed
	if m.tui.ActiveTab == 0 && m.tui.logViewMode == constants.PodLogViewMode {
		m.tui.clearPodLogs()
		return m.tui, tea.Batch(m.tui.loadPodLogs(), m.tui.startPodLogRefreshTimer())
	}
	
	return m.tui, m.tui.loadPodLogs()
}

// handleLogScroll scrolls through log content
func (m *MouseHandler) handleLogScroll(direction int) (tea.Model, tea.Cmd) {
	if direction > 0 {
		// Scroll down
		maxScroll := m.tui.getMaxLogScrollOffset()
		if m.tui.logScrollOffset < maxScroll {
			m.tui.logScrollOffset++
			m.tui.userScrolled = true
			m.tui.tailMode = false
		} else {
			m.tui.tailMode = true
			m.tui.userScrolled = false
		}
	} else {
		// Scroll up
		if m.tui.logScrollOffset > 0 {
			m.tui.logScrollOffset--
			m.tui.userScrolled = true
			m.tui.tailMode = false
		}
	}
	
	logging.Debug(m.tui.Logger, "MouseHandler: log scroll direction=%d, offset=%d", 
		direction, m.tui.logScrollOffset)
	
	return m.tui, nil
}