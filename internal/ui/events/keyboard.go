package events

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/ui/messages"
)

// KeyboardHandler handles keyboard events for the TUI
type KeyboardHandler struct {
	// State for handling different input modes
	helpMode         bool
	errorModalMode   bool
	projectModalMode bool
}

// NewKeyboardHandler creates a new keyboard handler
func NewKeyboardHandler() *KeyboardHandler {
	return &KeyboardHandler{
		helpMode:         false,
		errorModalMode:   false,
		projectModalMode: false,
	}
}

// KeyEvent represents a keyboard event with context
type KeyEvent struct {
	Key          tea.KeyMsg
	FocusedPanel int
	ShowHelp     bool
	ShowError    bool
	ShowProject  bool
	Connected    bool
	ActiveTab    uint
}

// HandleKeyEvent processes keyboard input and returns appropriate commands
func (h *KeyboardHandler) HandleKeyEvent(event KeyEvent) []tea.Cmd {
	var cmds []tea.Cmd

	// Special handling for help mode
	if event.ShowHelp {
		return h.handleHelpKeys(event.Key)
	}

	// Special handling for error modal
	if event.ShowError {
		return h.handleErrorModalKeys(event.Key)
	}

	// Special handling for project modal
	if event.ShowProject {
		return h.handleProjectModalKeys(event.Key)
	}

	// Normal key handling
	switch event.Key.String() {
	case "ctrl+c", "q":
		cmds = append(cmds, tea.Quit)

	case "esc":
		// Close any open modals or return to normal mode
		if event.ShowError || event.ShowProject {
			cmds = append(cmds, h.closeModal())
		}

	case "r":
		// Manual retry/reconnect or refresh
		if !event.Connected {
			cmds = append(cmds, h.retryConnection())
		} else if event.ActiveTab == 0 {
			cmds = append(cmds, h.refreshPods())
		}

	case "ctrl+p":
		// Open project switching modal
		if event.Connected {
			cmds = append(cmds, h.openProjectModal())
		}

	case "?":
		cmds = append(cmds, h.toggleHelp())

	case "tab":
		cmds = append(cmds, h.nextPanel())

	case "shift+tab":
		cmds = append(cmds, h.prevPanel())

	case "d":
		cmds = append(cmds, h.toggleDetails())

	case "L":
		cmds = append(cmds, h.toggleLogs())

	case "e":
		// Show error modal if there are errors
		cmds = append(cmds, h.showErrorModal())

	case "t":
		cmds = append(cmds, h.toggleTheme())

	// Panel-specific navigation
	case "1":
		cmds = append(cmds, h.focusPanel(0))

	case "2":
		cmds = append(cmds, h.focusPanel(1))

	case "3":
		cmds = append(cmds, h.focusPanel(2))

	// Tab navigation
	case "h", "left":
		cmds = append(cmds, h.prevTab())

	case "l", "right":
		cmds = append(cmds, h.nextTab())

	// Directional navigation
	case "j", "down":
		cmds = append(cmds, h.handleDownNavigation(event))

	case "k", "up":
		cmds = append(cmds, h.handleUpNavigation(event))

	// Log scrolling
	case "pgup":
		if event.FocusedPanel == 2 {
			cmds = append(cmds, h.scrollLogsPageUp())
		}

	case "pgdn":
		if event.FocusedPanel == 2 {
			cmds = append(cmds, h.scrollLogsPageDown())
		}

	case "home":
		if event.FocusedPanel == 2 {
			cmds = append(cmds, h.scrollLogsHome())
		}

	case "end":
		if event.FocusedPanel == 2 {
			cmds = append(cmds, h.scrollLogsEnd())
		}
	}

	return cmds
}

// handleHelpKeys handles keyboard input when help is shown
func (h *KeyboardHandler) handleHelpKeys(key tea.KeyMsg) []tea.Cmd {
	switch key.String() {
	case "?", "esc":
		return []tea.Cmd{h.toggleHelp()}
	}
	return nil
}

// handleErrorModalKeys handles keyboard input when error modal is shown
func (h *KeyboardHandler) handleErrorModalKeys(key tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch key.String() {
	case "esc", "q":
		cmds = append(cmds, h.closeModal())

	case "up":
		cmds = append(cmds, h.moveErrorSelection(-1))

	case "down":
		cmds = append(cmds, h.moveErrorSelection(1))

	case "enter":
		cmds = append(cmds, h.executeSelectedAction())

	case "c":
		cmds = append(cmds, h.clearErrors())
	}

	return cmds
}

// handleProjectModalKeys handles keyboard input when project modal is shown
func (h *KeyboardHandler) handleProjectModalKeys(key tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch key.String() {
	case "esc":
		cmds = append(cmds, h.closeModal())

	case "enter":
		cmds = append(cmds, h.switchToSelectedProject())

	case "j", "down":
		cmds = append(cmds, h.moveProjectSelection(1))

	case "k", "up":
		cmds = append(cmds, h.moveProjectSelection(-1))

	case "r":
		cmds = append(cmds, h.refreshProjects())
	}

	return cmds
}

// handleDownNavigation handles down/j key navigation
func (h *KeyboardHandler) handleDownNavigation(event KeyEvent) tea.Cmd {
	switch event.FocusedPanel {
	case 0:
		// Main panel - navigate in list or move to logs
		return h.navigateDown(event.ActiveTab)
	case 1:
		// Details panel - move to logs if available
		return h.focusPanel(2)
	case 2:
		// Logs panel - scroll down
		return h.scrollLogsDown()
	}
	return nil
}

// handleUpNavigation handles up/k key navigation
func (h *KeyboardHandler) handleUpNavigation(event KeyEvent) tea.Cmd {
	switch event.FocusedPanel {
	case 0:
		// Main panel - navigate in list
		return h.navigateUp(event.ActiveTab)
	case 2:
		// Logs panel - scroll up or move to main
		return h.scrollLogsUp()
	}
	return nil
}

// Navigation commands
func (h *KeyboardHandler) navigateDown(activeTab uint) tea.Cmd {
	return func() tea.Msg {
		return NavigateDownMsg{ActiveTab: activeTab}
	}
}

func (h *KeyboardHandler) navigateUp(activeTab uint) tea.Cmd {
	return func() tea.Msg {
		return NavigateUpMsg{ActiveTab: activeTab}
	}
}

func (h *KeyboardHandler) nextTab() tea.Cmd {
	return func() tea.Msg {
		return messages.TabSwitchMsg{TabIndex: 1} // Next tab
	}
}

func (h *KeyboardHandler) prevTab() tea.Cmd {
	return func() tea.Msg {
		return messages.TabSwitchMsg{TabIndex: -1} // Previous tab
	}
}

// Panel commands
func (h *KeyboardHandler) focusPanel(panel int) tea.Cmd {
	return func() tea.Msg {
		return FocusPanelMsg{Panel: panel}
	}
}

func (h *KeyboardHandler) nextPanel() tea.Cmd {
	return func() tea.Msg {
		return NextPanelMsg{}
	}
}

func (h *KeyboardHandler) prevPanel() tea.Cmd {
	return func() tea.Msg {
		return PrevPanelMsg{}
	}
}

// Modal commands
func (h *KeyboardHandler) toggleHelp() tea.Cmd {
	return func() tea.Msg {
		return messages.HelpToggleMsg{}
	}
}

func (h *KeyboardHandler) closeModal() tea.Cmd {
	return func() tea.Msg {
		return CloseModalMsg{}
	}
}

func (h *KeyboardHandler) openProjectModal() tea.Cmd {
	return func() tea.Msg {
		return OpenProjectModalMsg{}
	}
}

func (h *KeyboardHandler) showErrorModal() tea.Cmd {
	return func() tea.Msg {
		return ShowErrorModalMsg{}
	}
}

// View toggle commands
func (h *KeyboardHandler) toggleDetails() tea.Cmd {
	return func() tea.Msg {
		return ToggleDetailsMsg{}
	}
}

func (h *KeyboardHandler) toggleLogs() tea.Cmd {
	return func() tea.Msg {
		return ToggleLogsMsg{}
	}
}

func (h *KeyboardHandler) toggleTheme() tea.Cmd {
	return func() tea.Msg {
		return ToggleThemeMsg{}
	}
}

// Connection commands
func (h *KeyboardHandler) retryConnection() tea.Cmd {
	return func() tea.Msg {
		return RetryConnectionMsg{}
	}
}

func (h *KeyboardHandler) refreshPods() tea.Cmd {
	return func() tea.Msg {
		return messages.RefreshMsg{}
	}
}

// Log scrolling commands
func (h *KeyboardHandler) scrollLogsDown() tea.Cmd {
	return func() tea.Msg {
		return ScrollLogsMsg{Direction: 1}
	}
}

func (h *KeyboardHandler) scrollLogsUp() tea.Cmd {
	return func() tea.Msg {
		return ScrollLogsMsg{Direction: -1}
	}
}

func (h *KeyboardHandler) scrollLogsPageDown() tea.Cmd {
	return func() tea.Msg {
		return ScrollLogsPageMsg{Direction: 1}
	}
}

func (h *KeyboardHandler) scrollLogsPageUp() tea.Cmd {
	return func() tea.Msg {
		return ScrollLogsPageMsg{Direction: -1}
	}
}

func (h *KeyboardHandler) scrollLogsHome() tea.Cmd {
	return func() tea.Msg {
		return ScrollLogsHomeMsg{}
	}
}

func (h *KeyboardHandler) scrollLogsEnd() tea.Cmd {
	return func() tea.Msg {
		return ScrollLogsEndMsg{}
	}
}

// Error modal commands
func (h *KeyboardHandler) moveErrorSelection(direction int) tea.Cmd {
	return func() tea.Msg {
		return MoveErrorSelectionMsg{Direction: direction}
	}
}

func (h *KeyboardHandler) executeSelectedAction() tea.Cmd {
	return func() tea.Msg {
		return ExecuteSelectedActionMsg{}
	}
}

func (h *KeyboardHandler) clearErrors() tea.Cmd {
	return func() tea.Msg {
		return ClearErrorsMsg{}
	}
}

// Project modal commands
func (h *KeyboardHandler) moveProjectSelection(direction int) tea.Cmd {
	return func() tea.Msg {
		return MoveProjectSelectionMsg{Direction: direction}
	}
}

func (h *KeyboardHandler) switchToSelectedProject() tea.Cmd {
	return func() tea.Msg {
		return SwitchToSelectedProjectMsg{}
	}
}

func (h *KeyboardHandler) refreshProjects() tea.Cmd {
	return func() tea.Msg {
		return RefreshProjectsMsg{}
	}
}

// Message types for keyboard events
type NavigateDownMsg struct {
	ActiveTab uint
}

type NavigateUpMsg struct {
	ActiveTab uint
}

type FocusPanelMsg struct {
	Panel int
}

type NextPanelMsg struct{}

type PrevPanelMsg struct{}

type CloseModalMsg struct{}

type OpenProjectModalMsg struct{}

type ShowErrorModalMsg struct{}

type ToggleDetailsMsg struct{}

type ToggleLogsMsg struct{}

type ToggleThemeMsg struct{}

type RetryConnectionMsg struct{}

type ScrollLogsMsg struct {
	Direction int // 1 for down, -1 for up
}

type ScrollLogsPageMsg struct {
	Direction int // 1 for down, -1 for up
}

type ScrollLogsHomeMsg struct{}

type ScrollLogsEndMsg struct{}

type MoveErrorSelectionMsg struct {
	Direction int
}

type ExecuteSelectedActionMsg struct{}

type ClearErrorsMsg struct{}

type MoveProjectSelectionMsg struct {
	Direction int
}

type SwitchToSelectedProjectMsg struct{}

type RefreshProjectsMsg struct{}
