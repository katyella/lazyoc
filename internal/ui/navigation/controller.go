package navigation

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/ui/components"
)

// NavigationController manages the navigation state and handles routing
type NavigationController struct {
	registry     *KeybindingRegistry
	focusManager *FocusManager

	// Multi-key sequence handling
	pendingSequence string
	lastKeyTime     time.Time
	sequenceTimeout time.Duration

	// Search and command state
	searchQuery     string
	commandBuffer   string
	isSearchActive  bool
	isCommandActive bool

	// Event callbacks
	callbacks map[KeyAction]func() tea.Cmd
}

// FocusManager manages panel focus and navigation
type FocusManager struct {
	currentPanel  components.PanelType
	previousPanel components.PanelType
	focusOrder    []components.PanelType
	panelEnabled  map[components.PanelType]bool
}

// NavigationMsg represents navigation-related messages
type NavigationMsg struct {
	Action KeyAction
	Panel  components.PanelType
	Data   interface{}
}

// ModeChangeMsg represents mode change messages
type ModeChangeMsg struct {
	OldMode NavigationMode
	NewMode NavigationMode
}

// SearchMsg represents search-related messages
type SearchMsg struct {
	Query    string
	Active   bool
	Complete bool
}

// CommandMsg represents command-related messages
type CommandMsg struct {
	Command  string
	Active   bool
	Complete bool
}

// NewNavigationController creates a new navigation controller
func NewNavigationController() *NavigationController {
	nc := &NavigationController{
		registry:        NewKeybindingRegistry(),
		focusManager:    NewFocusManager(),
		sequenceTimeout: 500 * time.Millisecond,
		callbacks:       make(map[KeyAction]func() tea.Cmd),
	}

	return nc
}

// NewFocusManager creates a new focus manager
func NewFocusManager() *FocusManager {
	return &FocusManager{
		currentPanel:  components.PanelMain,
		previousPanel: components.PanelMain,
		focusOrder:    []components.PanelType{components.PanelMain, components.PanelDetail, components.PanelLog},
		panelEnabled: map[components.PanelType]bool{
			components.PanelMain:      true,
			components.PanelDetail:    true,
			components.PanelLog:       true,
			components.PanelStatusBar: false, // Status bar typically not focusable
			components.PanelTabs:      false, // Tabs handled differently
		},
	}
}

// SetCallback sets a callback function for a specific action
func (nc *NavigationController) SetCallback(action KeyAction, callback func() tea.Cmd) {
	nc.callbacks[action] = callback
}

// GetRegistry returns the keybinding registry
func (nc *NavigationController) GetRegistry() *KeybindingRegistry {
	return nc.registry
}

// GetFocusManager returns the focus manager
func (nc *NavigationController) GetFocusManager() *FocusManager {
	return nc.focusManager
}

// ProcessKeyMsg processes a keyboard message and returns appropriate commands
func (nc *NavigationController) ProcessKeyMsg(msg tea.KeyMsg) ([]tea.Cmd, bool) {
	var cmds []tea.Cmd
	keyStr := msg.String()

	// Handle multi-key sequences (like 'gg')
	if handled, cmd := nc.handleMultiKeySequence(keyStr); handled {
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return cmds, true
	}

	// Get action from registry
	action, exists := nc.registry.ProcessKeyMsg(msg)
	if !exists {
		return nil, false
	}

	// Handle mode transitions first
	if modeCmd := nc.handleModeTransitions(action); modeCmd != nil {
		cmds = append(cmds, modeCmd)
	}

	// Handle navigation actions
	if navCmd := nc.handleNavigationActions(action); navCmd != nil {
		cmds = append(cmds, navCmd)
	}

	// Handle general actions
	if actionCmd := nc.handleGeneralActions(action); actionCmd != nil {
		cmds = append(cmds, actionCmd)
	}

	// Execute registered callbacks
	if callback, exists := nc.callbacks[action]; exists {
		if cmd := callback(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cmds, len(cmds) > 0
}

// handleMultiKeySequence handles multi-key sequences like 'gg'
func (nc *NavigationController) handleMultiKeySequence(keyStr string) (bool, tea.Cmd) {
	now := time.Now()

	// Check for sequence timeout
	if !nc.lastKeyTime.IsZero() && now.Sub(nc.lastKeyTime) > nc.sequenceTimeout {
		nc.pendingSequence = ""
	}

	nc.lastKeyTime = now

	// Handle 'gg' sequence for go-to-top
	if keyStr == "g" {
		if nc.pendingSequence == "g" {
			// Second 'g' - execute go-to-top
			nc.pendingSequence = ""
			return true, func() tea.Msg {
				return NavigationMsg{Action: ActionGoToTop, Panel: nc.focusManager.currentPanel}
			}
		} else {
			// First 'g' - start sequence
			nc.pendingSequence = "g"
			return true, nil
		}
	}

	// If we have a pending sequence and this isn't completing it, clear it
	if nc.pendingSequence != "" && keyStr != "g" {
		nc.pendingSequence = ""
	}

	return false, nil
}

// handleModeTransitions handles mode transition actions
func (nc *NavigationController) handleModeTransitions(action KeyAction) tea.Cmd {
	var newMode NavigationMode
	var handled bool

	switch action {
	case ActionEnterSearch:
		newMode = ModeSearch
		handled = true
		nc.isSearchActive = true
		nc.searchQuery = ""

	case ActionEnterCommand:
		newMode = ModeCommand
		handled = true
		nc.isCommandActive = true
		nc.commandBuffer = ""

	case ActionEnterInsert:
		newMode = ModeInsert
		handled = true

	case ActionEnterNormal, ActionEscape:
		newMode = ModeNormal
		handled = true
		nc.isSearchActive = false
		nc.isCommandActive = false
		nc.searchQuery = ""
		nc.commandBuffer = ""
	}

	if handled {
		oldMode := nc.registry.GetMode()
		nc.registry.SetMode(newMode)

		return func() tea.Msg {
			return ModeChangeMsg{OldMode: oldMode, NewMode: newMode}
		}
	}

	return nil
}

// handleNavigationActions handles panel navigation actions
func (nc *NavigationController) handleNavigationActions(action KeyAction) tea.Cmd {
	switch action {
	case ActionNextPanel:
		nc.focusManager.NextPanel()
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: nc.focusManager.currentPanel}
		}

	case ActionPrevPanel:
		nc.focusManager.PrevPanel()
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: nc.focusManager.currentPanel}
		}

	case ActionFocusMain:
		nc.focusManager.SetFocus(components.PanelMain)
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: components.PanelMain}
		}

	case ActionFocusDetail:
		nc.focusManager.SetFocus(components.PanelDetail)
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: components.PanelDetail}
		}

	case ActionFocusLog:
		nc.focusManager.SetFocus(components.PanelLog)
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: components.PanelLog}
		}

	case ActionNextTab, ActionPrevTab:
		// Tab navigation actions - route to TUI
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: nc.focusManager.currentPanel}
		}

	case ActionMoveUp, ActionMoveDown, ActionMoveLeft, ActionMoveRight,
		ActionPageUp, ActionPageDown, ActionHalfPageUp, ActionHalfPageDown,
		ActionGoToTop, ActionGoToBottom:
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: nc.focusManager.currentPanel}
		}
	}

	return nil
}

// handleGeneralActions handles general application actions
func (nc *NavigationController) handleGeneralActions(action KeyAction) tea.Cmd {
	switch action {
	case ActionSelect:
		if nc.registry.GetMode() == ModeSearch && nc.isSearchActive {
			// Execute search
			query := nc.searchQuery
			nc.isSearchActive = false
			nc.registry.SetMode(ModeNormal)
			return func() tea.Msg {
				return SearchMsg{Query: query, Active: false, Complete: true}
			}
		} else if nc.registry.GetMode() == ModeCommand && nc.isCommandActive {
			// Execute command
			command := nc.commandBuffer
			nc.isCommandActive = false
			nc.registry.SetMode(ModeNormal)
			return func() tea.Msg {
				return CommandMsg{Command: command, Active: false, Complete: true}
			}
		} else {
			// Regular select action
			return func() tea.Msg {
				return NavigationMsg{Action: action, Panel: nc.focusManager.currentPanel}
			}
		}

	default:
		// Other actions that need to be routed
		return func() tea.Msg {
			return NavigationMsg{Action: action, Panel: nc.focusManager.currentPanel}
		}
	}

	return nil
}

// UpdateSearchQuery updates the search query (called when in search mode)
func (nc *NavigationController) UpdateSearchQuery(query string) tea.Cmd {
	nc.searchQuery = query
	return func() tea.Msg {
		return SearchMsg{Query: query, Active: true, Complete: false}
	}
}

// UpdateCommandBuffer updates the command buffer (called when in command mode)
func (nc *NavigationController) UpdateCommandBuffer(command string) tea.Cmd {
	nc.commandBuffer = command
	return func() tea.Msg {
		return CommandMsg{Command: command, Active: true, Complete: false}
	}
}

// GetCurrentMode returns the current navigation mode
func (nc *NavigationController) GetCurrentMode() NavigationMode {
	return nc.registry.GetMode()
}

// GetCurrentPanel returns the currently focused panel
func (nc *NavigationController) GetCurrentPanel() components.PanelType {
	return nc.focusManager.currentPanel
}

// GetSearchQuery returns the current search query
func (nc *NavigationController) GetSearchQuery() string {
	return nc.searchQuery
}

// GetCommandBuffer returns the current command buffer
func (nc *NavigationController) GetCommandBuffer() string {
	return nc.commandBuffer
}

// IsSearchActive returns whether search mode is active
func (nc *NavigationController) IsSearchActive() bool {
	return nc.isSearchActive
}

// IsCommandActive returns whether command mode is active
func (nc *NavigationController) IsCommandActive() bool {
	return nc.isCommandActive
}

// FocusManager methods

// GetCurrentFocus returns the currently focused panel
func (fm *FocusManager) GetCurrentFocus() components.PanelType {
	return fm.currentPanel
}

// GetPreviousFocus returns the previously focused panel
func (fm *FocusManager) GetPreviousFocus() components.PanelType {
	return fm.previousPanel
}

// SetFocus sets the focus to a specific panel
func (fm *FocusManager) SetFocus(panel components.PanelType) {
	if fm.panelEnabled[panel] {
		fm.previousPanel = fm.currentPanel
		fm.currentPanel = panel
	}
}

// NextPanel moves focus to the next panel in the focus order
func (fm *FocusManager) NextPanel() {
	currentIndex := fm.findPanelIndex(fm.currentPanel)
	if currentIndex == -1 {
		return
	}

	// Find next enabled panel
	for i := 1; i < len(fm.focusOrder); i++ {
		nextIndex := (currentIndex + i) % len(fm.focusOrder)
		nextPanel := fm.focusOrder[nextIndex]
		if fm.panelEnabled[nextPanel] {
			fm.SetFocus(nextPanel)
			return
		}
	}
}

// PrevPanel moves focus to the previous panel in the focus order
func (fm *FocusManager) PrevPanel() {
	currentIndex := fm.findPanelIndex(fm.currentPanel)
	if currentIndex == -1 {
		return
	}

	// Find previous enabled panel
	for i := 1; i < len(fm.focusOrder); i++ {
		prevIndex := (currentIndex - i + len(fm.focusOrder)) % len(fm.focusOrder)
		prevPanel := fm.focusOrder[prevIndex]
		if fm.panelEnabled[prevPanel] {
			fm.SetFocus(prevPanel)
			return
		}
	}
}

// SetPanelEnabled enables or disables a panel for focus
func (fm *FocusManager) SetPanelEnabled(panel components.PanelType, enabled bool) {
	fm.panelEnabled[panel] = enabled
}

// IsPanelEnabled returns whether a panel is enabled for focus
func (fm *FocusManager) IsPanelEnabled(panel components.PanelType) bool {
	return fm.panelEnabled[panel]
}

// findPanelIndex finds the index of a panel in the focus order
func (fm *FocusManager) findPanelIndex(panel components.PanelType) int {
	for i, p := range fm.focusOrder {
		if p == panel {
			return i
		}
	}
	return -1
}
