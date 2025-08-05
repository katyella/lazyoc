package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/constants"
)

// KeyboardHandler handles keyboard events
type KeyboardHandler struct {
	tui          *TUI
	navigator    *Navigator
	focusManager *FocusManager
}

// NewKeyboardHandler creates a new KeyboardHandler instance
func NewKeyboardHandler(tui *TUI, navigator *Navigator, focusManager *FocusManager) *KeyboardHandler {
	return &KeyboardHandler{
		tui:          tui,
		navigator:    navigator,
		focusManager: focusManager,
	}
}

// Handle processes keyboard events
func (k *KeyboardHandler) Handle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Special handling for help mode
	if k.tui.showHelp {
		if msg.String() == "?" || msg.String() == "esc" {
			k.tui.showHelp = false
			return k.tui, nil
		}
		return k.tui, nil
	}

	// Special handling for error modal
	if k.tui.showErrorModal {
		return k.tui.handleErrorModalKeys(msg)
	}

	// Special handling for project modal
	if k.tui.showProjectModal {
		return k.tui.handleProjectModalKeys(msg)
	}

	// Special handling for secret modal
	if k.tui.showSecretModal {
		return k.tui.handleSecretModalKeys(msg)
	}

	// Normal key handling
	switch msg.String() {
	case "ctrl+c", "q":
		return k.tui, tea.Quit
		
	case "ctrl+p":
		return k.handleProjectSwitchKey()

	case "esc":
		// Close error modal if open
		if k.tui.showErrorModal {
			k.tui.showErrorModal = false
			return k.tui, nil
		}
		return k.tui, nil

	case "r":
		// Manual retry/reconnect
		if !k.tui.connected && !k.tui.connecting {
			return k.tui, k.tui.InitializeK8sClient(k.tui.KubeconfigPath)
		}
		return k.tui, nil

	case "?":
		k.tui.showHelp = !k.tui.showHelp
		return k.tui, nil

	case "tab":
		k.focusManager.CycleFocus()
		return k.tui, nil

	case "shift+tab":
		k.focusManager.CycleFocusReverse()
		return k.tui, nil

	case "enter":
		return k.handleEnterKey()

	case "p":
		return k.handleProjectKey()

	case "space":
		return k.handleSpaceKey()

	case "c":
		return k.handleCopyKey()

	case "d":
		return k.handleDetailsToggleKey()
		
	case "e":
		return k.handleErrorKey()

	case "t":
		return k.handleThemeToggleKey()
		
	case "T":
		return k.handleTailToggleKey()

	case "l":
		return k.handleLogToggleKey()
		
	case "L":
		return k.handleLogPanelToggleKey()

	case "j", "down":
		return k.handleDownKey()

	case "k", "up":
		return k.handleUpKey()

	case "h":
		return k.handleLeftTabKey()
		
	case "left":
		k.tui.PrevTab()
		return k.tui, k.tui.handleTabSwitch()

	case "right":
		k.tui.NextTab()
		return k.tui, k.tui.handleTabSwitch()

	case "1":
		k.focusManager.FocusPanel(0) // Focus main panel
		return k.tui, nil

	case "2":
		if k.tui.showDetails {
			k.focusManager.FocusPanel(1) // Focus details panel
		}
		return k.tui, nil

	case "3":
		if k.tui.showLogs {
			k.focusManager.FocusPanel(2) // Focus logs panel
		}
		return k.tui, nil
	}

	return k.tui, nil
}

// handleDownKey handles 'j' and down arrow key
func (k *KeyboardHandler) handleDownKey() (tea.Model, tea.Cmd) {
	if k.focusManager.IsMainPanelFocused() {
		k.navigator.SelectNextResource()
		// For pods, handle log loading
		if k.tui.ActiveTab == 0 && k.tui.logViewMode == constants.PodLogViewMode {
			k.tui.clearPodLogs()
			return k.tui, tea.Batch(k.tui.loadPodLogs(), k.tui.startPodLogRefreshTimer())
		}
		return k.tui, k.tui.loadPodLogs()
	} else if k.focusManager.IsMainPanelFocused() && k.tui.showLogs {
		// Move focus down to logs panel
		k.focusManager.FocusPanel(2)
	} else if k.focusManager.IsDetailsPanelFocused() && k.tui.showLogs {
		// Move focus from details to logs
		k.focusManager.FocusPanel(2)
	} else if k.focusManager.IsLogsPanelFocused() && k.tui.logViewMode == "pod" && len(k.tui.podLogs) > 0 {
		// Scroll down in pod logs
		maxScroll := k.tui.getMaxLogScrollOffset()
		if k.tui.logScrollOffset < maxScroll {
			k.tui.logScrollOffset += 1
			k.tui.userScrolled = true
			k.tui.tailMode = false
		} else {
			k.tui.tailMode = true
			k.tui.userScrolled = false
		}
	}
	return k.tui, nil
}

// handleUpKey handles 'k' and up arrow key
func (k *KeyboardHandler) handleUpKey() (tea.Model, tea.Cmd) {
	if k.focusManager.IsMainPanelFocused() {
		k.navigator.SelectPreviousResource()
		// For pods, handle log loading
		if k.tui.ActiveTab == 0 && k.tui.logViewMode == constants.PodLogViewMode {
			k.tui.clearPodLogs()
			return k.tui, tea.Batch(k.tui.loadPodLogs(), k.tui.startPodLogRefreshTimer())
		}
		return k.tui, k.tui.loadPodLogs()
	} else if k.focusManager.IsLogsPanelFocused() && k.tui.logViewMode == "pod" && len(k.tui.podLogs) > 0 {
		// Scroll up in pod logs
		if k.tui.logScrollOffset > 0 {
			k.tui.logScrollOffset -= 1
			k.tui.userScrolled = true
			k.tui.tailMode = false
		}
	} else if k.focusManager.IsLogsPanelFocused() {
		// Move focus up from logs
		if k.tui.showDetails {
			k.focusManager.FocusPanel(1) // Focus details panel
		} else {
			k.focusManager.FocusPanel(0) // Focus main panel
		}
	} else if k.focusManager.IsDetailsPanelFocused() {
		// Move focus up from details to main
		k.focusManager.FocusPanel(0)
	}
	return k.tui, nil
}

// Additional handler methods for other keys
func (k *KeyboardHandler) handleEnterKey() (tea.Model, tea.Cmd) {
	if k.focusManager.IsMainPanelFocused() {
		switch k.tui.ActiveTab {
		case 0: // Pods tab
			if len(k.tui.pods) > 0 {
				// Toggle details panel for the selected pod
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		case 1: // Services tab
			if len(k.tui.services) > 0 {
				// Toggle details panel for the selected service
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		case 2: // Deployments tab
			if len(k.tui.deployments) > 0 {
				// Toggle details panel for the selected deployment
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		case 3: // ConfigMaps tab
			if len(k.tui.configMaps) > 0 {
				// Toggle details panel for the selected configmap
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		case 4: // Secrets tab
			if len(k.tui.secrets) > 0 {
				// Load and show secret data in modal
				return k.tui, k.tui.loadSecretData()
			}
		case 5: // BuildConfigs tab
			if len(k.tui.buildConfigs) > 0 {
				// Toggle details panel for the selected buildconfig
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		case 6: // ImageStreams tab
			if len(k.tui.imageStreams) > 0 {
				// Toggle details panel for the selected imagestream
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		case 7: // Routes tab
			if len(k.tui.routes) > 0 {
				// Toggle details panel for the selected route
				k.tui.showDetails = !k.tui.showDetails
				return k.tui, nil
			}
		}
	} else if k.focusManager.IsLogsPanelFocused() {
		// Toggle log view when in log panel
		if k.tui.logViewMode == constants.DefaultLogViewMode {
			k.tui.logViewMode = constants.PodLogViewMode
		} else {
			k.tui.logViewMode = constants.DefaultLogViewMode
		}
		return k.tui, nil
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleProjectKey() (tea.Model, tea.Cmd) {
	// Open project switching modal if connected
	if k.tui.connected {
		return k.tui, k.tui.openProjectModal()
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleProjectSwitchKey() (tea.Model, tea.Cmd) {
	// Open project switching modal if connected (same as 'p' key)
	if k.tui.connected {
		return k.tui, k.tui.openProjectModal()
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleSpaceKey() (tea.Model, tea.Cmd) {
	// Toggle details panel
	k.tui.showDetails = !k.tui.showDetails
	return k.tui, nil
}

func (k *KeyboardHandler) handleCopyKey() (tea.Model, tea.Cmd) {
	// Copy selected resource info (placeholder - not implemented in original)
	return k.tui, nil
}

func (k *KeyboardHandler) handleDetailsToggleKey() (tea.Model, tea.Cmd) {
	// Toggle details panel
	k.tui.showDetails = !k.tui.showDetails
	return k.tui, nil
}

func (k *KeyboardHandler) handleErrorKey() (tea.Model, tea.Cmd) {
	// Show error modal if there are errors
	if k.tui.errorDisplay.HasErrors() {
		k.tui.showErrorModal = true
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleThemeToggleKey() (tea.Model, tea.Cmd) {
	// Toggle theme
	if k.tui.theme == "dark" {
		k.tui.theme = "light"
	} else {
		k.tui.theme = "dark"
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleTailToggleKey() (tea.Model, tea.Cmd) {
	// Toggle tail mode for logs
	if k.focusManager.IsLogsPanelFocused() {
		k.tui.tailMode = !k.tui.tailMode
		if k.tui.tailMode {
			k.tui.userScrolled = false
		}
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleLogToggleKey() (tea.Model, tea.Cmd) {
	// Different behavior based on current panel
	if k.focusManager.IsLogsPanelFocused() {
		// Toggle log view when in log panel
		if k.tui.logViewMode == constants.DefaultLogViewMode {
			k.tui.logViewMode = constants.PodLogViewMode
		} else {
			k.tui.logViewMode = constants.DefaultLogViewMode
		}
		return k.tui, nil
	} else if k.focusManager.IsMainPanelFocused() {
		// Navigate tabs when in main panel (h/l navigation)
		k.tui.NextTab()
		return k.tui, k.tui.handleTabSwitch()
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleLogPanelToggleKey() (tea.Model, tea.Cmd) {
	// Toggle logs panel (L key)
	k.tui.showLogs = !k.tui.showLogs
	if !k.tui.showLogs && k.focusManager.IsLogsPanelFocused() {
		k.focusManager.FocusPanel(0) // Focus main panel if logs were focused
	}
	return k.tui, nil
}

func (k *KeyboardHandler) handleLeftTabKey() (tea.Model, tea.Cmd) {
	// Navigate tabs when in main panel (h/l navigation)
	if k.focusManager.IsMainPanelFocused() {
		k.tui.PrevTab()
		return k.tui, k.tui.handleTabSwitch()
	}
	return k.tui, nil
}