package events

import (
	tea "github.com/charmbracelet/bubbletea"
)

// MouseHandler handles mouse events for the TUI
type MouseHandler struct {
	enabled bool
}

// NewMouseHandler creates a new mouse handler
func NewMouseHandler(enabled bool) *MouseHandler {
	return &MouseHandler{
		enabled: enabled,
	}
}

// MouseEvent represents a mouse event with context
type MouseEvent struct {
	Mouse        tea.MouseMsg
	Width        int
	Height       int
	ShowDetails  bool
	ShowLogs     bool
	ActiveTab    uint
	FocusedPanel int
}

// HandleMouseEvent processes mouse input and returns appropriate commands
func (h *MouseHandler) HandleMouseEvent(event MouseEvent) []tea.Cmd {
	if !h.enabled {
		return nil
	}

	var cmds []tea.Cmd

	switch event.Mouse.Type {
	case tea.MouseLeft:
		cmds = append(cmds, h.handleLeftClick(event))

	case tea.MouseWheelUp:
		cmds = append(cmds, h.handleScrollUp(event))

	case tea.MouseWheelDown:
		cmds = append(cmds, h.handleScrollDown(event))

	case tea.MouseRight:
		cmds = append(cmds, h.handleRightClick(event))
	}

	return cmds
}

// handleLeftClick handles left mouse clicks
func (h *MouseHandler) handleLeftClick(event MouseEvent) tea.Cmd {
	x, y := event.Mouse.X, event.Mouse.Y

	// Determine which area was clicked
	clickArea := h.determineClickArea(x, y, event)

	switch clickArea {
	case ClickAreaTabs:
		return h.handleTabClick(x, event)

	case ClickAreaMainPanel:
		return h.handleMainPanelClick(x, y, event)

	case ClickAreaDetailsPanel:
		return h.handleDetailsPanelClick(x, y, event)

	case ClickAreaLogsPanel:
		return h.handleLogsPanelClick(x, y, event)

	case ClickAreaStatusBar:
		return h.handleStatusBarClick(x, y, event)

	default:
		return nil
	}
}

// handleScrollUp handles mouse wheel up events
func (h *MouseHandler) handleScrollUp(event MouseEvent) tea.Cmd {
	x, y := event.Mouse.X, event.Mouse.Y
	clickArea := h.determineClickArea(x, y, event)

	switch clickArea {
	case ClickAreaMainPanel:
		// Scroll up in main content (e.g., pod list)
		return func() tea.Msg {
			return ScrollMainContentMsg{Direction: -1}
		}

	case ClickAreaLogsPanel:
		// Scroll up in logs
		return func() tea.Msg {
			return ScrollLogsMsg{Direction: -1}
		}

	case ClickAreaDetailsPanel:
		// Scroll up in details
		return func() tea.Msg {
			return ScrollDetailsMsg{Direction: -1}
		}
	}

	return nil
}

// handleScrollDown handles mouse wheel down events
func (h *MouseHandler) handleScrollDown(event MouseEvent) tea.Cmd {
	x, y := event.Mouse.X, event.Mouse.Y
	clickArea := h.determineClickArea(x, y, event)

	switch clickArea {
	case ClickAreaMainPanel:
		// Scroll down in main content (e.g., pod list)
		return func() tea.Msg {
			return ScrollMainContentMsg{Direction: 1}
		}

	case ClickAreaLogsPanel:
		// Scroll down in logs
		return func() tea.Msg {
			return ScrollLogsMsg{Direction: 1}
		}

	case ClickAreaDetailsPanel:
		// Scroll down in details
		return func() tea.Msg {
			return ScrollDetailsMsg{Direction: 1}
		}
	}

	return nil
}

// handleRightClick handles right mouse clicks
func (h *MouseHandler) handleRightClick(event MouseEvent) tea.Cmd {
	x, y := event.Mouse.X, event.Mouse.Y
	clickArea := h.determineClickArea(x, y, event)

	switch clickArea {
	case ClickAreaMainPanel:
		// Show context menu for main panel
		return func() tea.Msg {
			return ShowContextMenuMsg{
				X:        x,
				Y:        y,
				Context:  "main",
				ActiveTab: event.ActiveTab,
			}
		}

	case ClickAreaLogsPanel:
		// Show context menu for logs panel
		return func() tea.Msg {
			return ShowContextMenuMsg{
				X:       x,
				Y:       y,
				Context: "logs",
			}
		}
	}

	return nil
}

// handleTabClick handles clicks on the tab bar
func (h *MouseHandler) handleTabClick(x int, event MouseEvent) tea.Cmd {
	// Calculate which tab was clicked based on x position
	// This is a simplified calculation - in a real implementation,
	// you'd need to know the exact tab positions
	tabWidth := event.Width / 5 // Assuming 5 tabs
	clickedTab := x / tabWidth

	if clickedTab >= 0 && clickedTab < 5 {
		return func() tea.Msg {
			return SwitchToTabMsg{TabIndex: uint(clickedTab)}
		}
	}

	return nil
}

// handleMainPanelClick handles clicks in the main panel
func (h *MouseHandler) handleMainPanelClick(x, y int, event MouseEvent) tea.Cmd {
	// Focus the main panel
	focusCmd := func() tea.Msg {
		return FocusPanelMsg{Panel: 0}
	}

	// Calculate which item was clicked (simplified)
	// In a real implementation, you'd need to know the exact line positions
	headerLines := 5 // Rough estimate for header and table header
	if y > headerLines {
		itemIndex := y - headerLines - 1
		if itemIndex >= 0 {
			return tea.Batch(
				focusCmd,
				func() tea.Msg {
					return SelectItemMsg{
						Index:     itemIndex,
						ActiveTab: event.ActiveTab,
					}
				},
			)
		}
	}

	return focusCmd
}

// handleDetailsPanelClick handles clicks in the details panel
func (h *MouseHandler) handleDetailsPanelClick(x, y int, event MouseEvent) tea.Cmd {
	// Focus the details panel
	return func() tea.Msg {
		return FocusPanelMsg{Panel: 1}
	}
}

// handleLogsPanelClick handles clicks in the logs panel
func (h *MouseHandler) handleLogsPanelClick(x, y int, event MouseEvent) tea.Cmd {
	// Focus the logs panel
	return func() tea.Msg {
		return FocusPanelMsg{Panel: 2}
	}
}

// handleStatusBarClick handles clicks on the status bar
func (h *MouseHandler) handleStatusBarClick(x, y int, event MouseEvent) tea.Cmd {
	// Different actions based on where in the status bar the click occurred
	if x < event.Width/3 {
		// Left section - connection status
		return func() tea.Msg {
			return ShowConnectionInfoMsg{}
		}
	} else if x > 2*event.Width/3 {
		// Right section - help hints
		return func() tea.Msg {
			return ShowHelpMsg{}
		}
	} else {
		// Middle section - project info
		return func() tea.Msg {
			return OpenProjectModalMsg{}
		}
	}
}

// determineClickArea determines which area of the UI was clicked
func (h *MouseHandler) determineClickArea(x, y int, event MouseEvent) ClickArea {
	// This is a simplified implementation
	// In a real implementation, you'd need to track the exact positions of each UI element

	// Header area (first 2-3 lines)
	if y <= 2 {
		return ClickAreaHeader
	}

	// Tab area (next line)
	if y == 3 {
		return ClickAreaTabs
	}

	// Status bar (last line)
	if y >= event.Height-1 {
		return ClickAreaStatusBar
	}

	// Calculate content area dimensions
	contentStartY := 4
	contentEndY := event.Height - 2
	contentHeight := contentEndY - contentStartY

	// Determine panel layout
	if event.ShowDetails && event.ShowLogs {
		// Three-panel layout
		mainWidth := int(float64(event.Width) * 0.4)
		logHeight := int(float64(contentHeight) * 0.33)

		if y >= contentEndY-logHeight {
			// Logs panel (bottom)
			return ClickAreaLogsPanel
		} else if x > mainWidth {
			// Details panel (right)
			return ClickAreaDetailsPanel
		} else {
			// Main panel (left)
			return ClickAreaMainPanel
		}
	} else if event.ShowDetails {
		// Two-panel layout (main + details)
		mainWidth := int(float64(event.Width) * 0.6)
		if x > mainWidth {
			return ClickAreaDetailsPanel
		} else {
			return ClickAreaMainPanel
		}
	} else if event.ShowLogs {
		// Two-panel layout (main + logs)
		logHeight := int(float64(contentHeight) * 0.33)
		if y >= contentEndY-logHeight {
			return ClickAreaLogsPanel
		} else {
			return ClickAreaMainPanel
		}
	} else {
		// Single panel layout
		return ClickAreaMainPanel
	}
}

// SetEnabled enables or disables mouse support
func (h *MouseHandler) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// IsEnabled returns whether mouse support is enabled
func (h *MouseHandler) IsEnabled() bool {
	return h.enabled
}

// ClickArea represents different clickable areas of the UI
type ClickArea int

const (
	ClickAreaUnknown ClickArea = iota
	ClickAreaHeader
	ClickAreaTabs
	ClickAreaMainPanel
	ClickAreaDetailsPanel
	ClickAreaLogsPanel
	ClickAreaStatusBar
)

// Message types for mouse events
type ScrollMainContentMsg struct {
	Direction int // 1 for down, -1 for up
}

type ScrollDetailsMsg struct {
	Direction int // 1 for down, -1 for up
}

type ShowContextMenuMsg struct {
	X         int
	Y         int
	Context   string
	ActiveTab uint
}

type SwitchToTabMsg struct {
	TabIndex uint
}

type SelectItemMsg struct {
	Index     int
	ActiveTab uint
}

type ShowConnectionInfoMsg struct{}

type ShowHelpMsg struct{}