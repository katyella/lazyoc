package ui

import (
	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/logging"
)

// FocusManager handles panel focus state and transitions
type FocusManager struct {
	tui *TUI
}

// NewFocusManager creates a new FocusManager instance
func NewFocusManager(tui *TUI) *FocusManager {
	return &FocusManager{tui: tui}
}

// FocusPanel sets focus to a specific panel if it's available
func (f *FocusManager) FocusPanel(panel int) bool {
	if panel < 0 || panel >= constants.PanelCount {
		logging.Debug(f.tui.Logger, "FocusManager: invalid panel %d", panel)
		return false
	}
	
	// Check if panel is available
	switch panel {
	case 0: // Main panel - always available
		break
	case 1: // Details panel
		if !f.tui.showDetails {
			logging.Debug(f.tui.Logger, "FocusManager: details panel not shown, cannot focus")
			return false
		}
	case 2: // Logs panel
		if !f.tui.showLogs {
			logging.Debug(f.tui.Logger, "FocusManager: logs panel not shown, cannot focus")
			return false
		}
	}
	
	oldPanel := f.tui.focusedPanel
	f.tui.focusedPanel = panel
	logging.Debug(f.tui.Logger, "FocusManager: switched focus from panel %d to panel %d", oldPanel, panel)
	return true
}

// CycleFocus moves focus to the next available panel
func (f *FocusManager) CycleFocus() {
	nextPanel := (f.tui.focusedPanel + 1) % constants.PanelCount
	
	// Keep cycling until we find an available panel or return to current
	startPanel := f.tui.focusedPanel
	for nextPanel != startPanel {
		if f.FocusPanel(nextPanel) {
			return
		}
		nextPanel = (nextPanel + 1) % constants.PanelCount
	}
}

// CycleFocusReverse moves focus to the previous available panel
func (f *FocusManager) CycleFocusReverse() {
	nextPanel := (f.tui.focusedPanel + constants.PanelCount - 1) % constants.PanelCount
	
	// Keep cycling until we find an available panel or return to current
	startPanel := f.tui.focusedPanel
	for nextPanel != startPanel {
		if f.FocusPanel(nextPanel) {
			return
		}
		nextPanel = (nextPanel + constants.PanelCount - 1) % constants.PanelCount
	}
}

// GetFocusedPanel returns the currently focused panel
func (f *FocusManager) GetFocusedPanel() int {
	return f.tui.focusedPanel
}

// IsMainPanelFocused returns true if the main panel is focused
func (f *FocusManager) IsMainPanelFocused() bool {
	return f.tui.focusedPanel == 0
}

// IsDetailsPanelFocused returns true if the details panel is focused
func (f *FocusManager) IsDetailsPanelFocused() bool {
	return f.tui.focusedPanel == 1 && f.tui.showDetails
}

// IsLogsPanelFocused returns true if the logs panel is focused
func (f *FocusManager) IsLogsPanelFocused() bool {
	return f.tui.focusedPanel == 2 && f.tui.showLogs
}