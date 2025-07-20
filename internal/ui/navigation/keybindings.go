package navigation

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyAction represents a keyboard action
type KeyAction string

const (
	// Global actions
	ActionQuit             KeyAction = "quit"
	ActionToggleHelp       KeyAction = "toggle_help"
	ActionToggleDebug      KeyAction = "toggle_debug"
	ActionRefresh          KeyAction = "refresh"
	ActionEscape           KeyAction = "escape"
	
	// Navigation actions
	ActionMoveUp           KeyAction = "move_up"
	ActionMoveDown         KeyAction = "move_down"
	ActionMoveLeft         KeyAction = "move_left"
	ActionMoveRight        KeyAction = "move_right"
	ActionPageUp           KeyAction = "page_up"
	ActionPageDown         KeyAction = "page_down"
	ActionHalfPageUp       KeyAction = "half_page_up"
	ActionHalfPageDown     KeyAction = "half_page_down"
	ActionGoToTop          KeyAction = "goto_top"
	ActionGoToBottom       KeyAction = "goto_bottom"
	
	// Panel navigation
	ActionNextPanel        KeyAction = "next_panel"
	ActionPrevPanel        KeyAction = "prev_panel"
	ActionNextTab          KeyAction = "next_tab"
	ActionPrevTab          KeyAction = "prev_tab"
	ActionFocusMain        KeyAction = "focus_main"
	ActionFocusDetail      KeyAction = "focus_detail"
	ActionFocusLog         KeyAction = "focus_log"
	
	// Modal actions
	ActionEnterSearch      KeyAction = "enter_search"
	ActionEnterCommand     KeyAction = "enter_command"
	ActionEnterInsert      KeyAction = "enter_insert"
	ActionEnterNormal      KeyAction = "enter_normal"
	
	// Content actions
	ActionSelect           KeyAction = "select"
	ActionDelete           KeyAction = "delete"
	ActionEdit             KeyAction = "edit"
	ActionCopy             KeyAction = "copy"
	ActionToggleCollapse   KeyAction = "toggle_collapse"
	ActionToggleVisible    KeyAction = "toggle_visible"
	
	// Log actions
	ActionClearLogs        KeyAction = "clear_logs"
	ActionToggleAutoscroll KeyAction = "toggle_autoscroll"
	ActionTogglePause      KeyAction = "toggle_pause"
	ActionFilterLevel      KeyAction = "filter_level"
)

// KeyBinding represents a key combination and its action
type KeyBinding struct {
	Key         string
	Action      KeyAction
	Description string
	Context     NavigationMode
}

// NavigationMode represents the current navigation mode
type NavigationMode string

const (
	ModeNormal  NavigationMode = "normal"
	ModeSearch  NavigationMode = "search"
	ModeCommand NavigationMode = "command"
	ModeInsert  NavigationMode = "insert"
)

// KeybindingRegistry manages keyboard shortcuts and their actions
type KeybindingRegistry struct {
	bindings map[NavigationMode]map[string]KeyBinding
	mode     NavigationMode
}

// NewKeybindingRegistry creates a new keybinding registry with default vim-like bindings
func NewKeybindingRegistry() *KeybindingRegistry {
	kr := &KeybindingRegistry{
		bindings: make(map[NavigationMode]map[string]KeyBinding),
		mode:     ModeNormal,
	}
	
	kr.initializeDefaultBindings()
	return kr
}

// initializeDefaultBindings sets up the default vim-like key bindings
func (kr *KeybindingRegistry) initializeDefaultBindings() {
	// Initialize maps for each mode
	kr.bindings[ModeNormal] = make(map[string]KeyBinding)
	kr.bindings[ModeSearch] = make(map[string]KeyBinding)
	kr.bindings[ModeCommand] = make(map[string]KeyBinding)
	kr.bindings[ModeInsert] = make(map[string]KeyBinding)
	
	// Normal mode bindings (vim-like)
	normalBindings := []KeyBinding{
		// Global actions
		{"q", ActionQuit, "Quit application", ModeNormal},
		{"ctrl+c", ActionQuit, "Quit application", ModeNormal},
		{"?", ActionToggleHelp, "Toggle help", ModeNormal},
		{"ctrl+d", ActionToggleDebug, "Toggle debug mode", ModeNormal},
		{"r", ActionRefresh, "Refresh", ModeNormal},
		{"f5", ActionRefresh, "Refresh", ModeNormal},
		{"esc", ActionEscape, "Escape/Cancel", ModeNormal},
		
		// Vim-like movement
		{"h", ActionMoveLeft, "Move left", ModeNormal},
		{"j", ActionMoveDown, "Move down", ModeNormal},
		{"k", ActionMoveUp, "Move up", ModeNormal},
		{"l", ActionMoveRight, "Move right", ModeNormal},
		
		// Alternative movement keys
		{"left", ActionMoveLeft, "Move left", ModeNormal},
		{"down", ActionMoveDown, "Move down", ModeNormal},
		{"up", ActionMoveUp, "Move up", ModeNormal},
		{"right", ActionMoveRight, "Move right", ModeNormal},
		
		// Page navigation
		{"ctrl+f", ActionPageDown, "Page down", ModeNormal},
		{"ctrl+b", ActionPageUp, "Page up", ModeNormal},
		{"ctrl+d", ActionHalfPageDown, "Half page down", ModeNormal},
		{"ctrl+u", ActionHalfPageUp, "Half page up", ModeNormal},
		{"pagedown", ActionPageDown, "Page down", ModeNormal},
		{"pageup", ActionPageUp, "Page up", ModeNormal},
		
		// Jump navigation (vim-style)
		{"gg", ActionGoToTop, "Go to top", ModeNormal},
		{"G", ActionGoToBottom, "Go to bottom", ModeNormal},
		{"home", ActionGoToTop, "Go to top", ModeNormal},
		{"end", ActionGoToBottom, "Go to bottom", ModeNormal},
		
		// Panel navigation
		{"tab", ActionNextPanel, "Next panel", ModeNormal},
		{"shift+tab", ActionPrevPanel, "Previous panel", ModeNormal},
		{"H", ActionPrevTab, "Previous tab", ModeNormal},
		{"L", ActionNextTab, "Next tab", ModeNormal},
		
		// Direct panel focus
		{"1", ActionFocusMain, "Focus main panel", ModeNormal},
		{"2", ActionFocusDetail, "Focus detail panel", ModeNormal},
		{"3", ActionFocusLog, "Focus log panel", ModeNormal},
		
		// Modal transitions
		{"/", ActionEnterSearch, "Enter search mode", ModeNormal},
		{":", ActionEnterCommand, "Enter command mode", ModeNormal},
		{"i", ActionEnterInsert, "Enter insert mode", ModeNormal},
		
		// Content actions
		{"enter", ActionSelect, "Select item", ModeNormal},
		{"space", ActionSelect, "Select item", ModeNormal},
		{"d", ActionDelete, "Delete item", ModeNormal},
		{"e", ActionEdit, "Edit item", ModeNormal},
		{"y", ActionCopy, "Copy item", ModeNormal},
		{"c", ActionToggleCollapse, "Toggle collapse", ModeNormal},
		{"v", ActionToggleVisible, "Toggle visibility", ModeNormal},
		
		// Log-specific actions
		{"C", ActionClearLogs, "Clear logs", ModeNormal},
		{"a", ActionToggleAutoscroll, "Toggle autoscroll", ModeNormal},
		{"p", ActionTogglePause, "Toggle pause", ModeNormal},
	}
	
	// Add normal mode bindings
	for _, binding := range normalBindings {
		kr.bindings[ModeNormal][binding.Key] = binding
	}
	
	// Search mode bindings
	searchBindings := []KeyBinding{
		{"esc", ActionEnterNormal, "Exit search mode", ModeSearch},
		{"ctrl+c", ActionEnterNormal, "Exit search mode", ModeSearch},
		{"enter", ActionSelect, "Execute search", ModeSearch},
	}
	
	for _, binding := range searchBindings {
		kr.bindings[ModeSearch][binding.Key] = binding
	}
	
	// Command mode bindings
	commandBindings := []KeyBinding{
		{"esc", ActionEnterNormal, "Exit command mode", ModeCommand},
		{"ctrl+c", ActionEnterNormal, "Exit command mode", ModeCommand},
		{"enter", ActionSelect, "Execute command", ModeCommand},
	}
	
	for _, binding := range commandBindings {
		kr.bindings[ModeCommand][binding.Key] = binding
	}
	
	// Insert mode bindings
	insertBindings := []KeyBinding{
		{"esc", ActionEnterNormal, "Exit insert mode", ModeInsert},
		{"ctrl+c", ActionEnterNormal, "Exit insert mode", ModeInsert},
	}
	
	for _, binding := range insertBindings {
		kr.bindings[ModeInsert][binding.Key] = binding
	}
}

// GetMode returns the current navigation mode
func (kr *KeybindingRegistry) GetMode() NavigationMode {
	return kr.mode
}

// SetMode sets the current navigation mode
func (kr *KeybindingRegistry) SetMode(mode NavigationMode) {
	kr.mode = mode
}

// GetAction returns the action for a given key in the current mode
func (kr *KeybindingRegistry) GetAction(key string) (KeyAction, bool) {
	if modeBindings, exists := kr.bindings[kr.mode]; exists {
		if binding, exists := modeBindings[key]; exists {
			return binding.Action, true
		}
	}
	return "", false
}

// GetBinding returns the full binding for a given key in the current mode
func (kr *KeybindingRegistry) GetBinding(key string) (KeyBinding, bool) {
	if modeBindings, exists := kr.bindings[kr.mode]; exists {
		if binding, exists := modeBindings[key]; exists {
			return binding, true
		}
	}
	return KeyBinding{}, false
}

// GetAllBindings returns all bindings for the current mode
func (kr *KeybindingRegistry) GetAllBindings() map[string]KeyBinding {
	if modeBindings, exists := kr.bindings[kr.mode]; exists {
		return modeBindings
	}
	return make(map[string]KeyBinding)
}

// GetBindingsForMode returns all bindings for a specific mode
func (kr *KeybindingRegistry) GetBindingsForMode(mode NavigationMode) map[string]KeyBinding {
	if modeBindings, exists := kr.bindings[mode]; exists {
		return modeBindings
	}
	return make(map[string]KeyBinding)
}

// AddBinding adds or updates a key binding
func (kr *KeybindingRegistry) AddBinding(key string, action KeyAction, description string, mode NavigationMode) {
	if kr.bindings[mode] == nil {
		kr.bindings[mode] = make(map[string]KeyBinding)
	}
	
	kr.bindings[mode][key] = KeyBinding{
		Key:         key,
		Action:      action,
		Description: description,
		Context:     mode,
	}
}

// RemoveBinding removes a key binding
func (kr *KeybindingRegistry) RemoveBinding(key string, mode NavigationMode) {
	if modeBindings, exists := kr.bindings[mode]; exists {
		delete(modeBindings, key)
	}
}

// ProcessKeyMsg processes a BubbleTea KeyMsg and returns the corresponding action
func (kr *KeybindingRegistry) ProcessKeyMsg(msg tea.KeyMsg) (KeyAction, bool) {
	keyStr := msg.String()
	
	// Handle special cases for complex key combinations
	switch {
	case keyStr == "g":
		// Handle 'gg' sequence for go-to-top
		// This would need state management for multi-key sequences
		return kr.GetAction("g")
	default:
		return kr.GetAction(keyStr)
	}
}

// GetHelpText returns formatted help text for the current mode
func (kr *KeybindingRegistry) GetHelpText() string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("=== %s MODE ===\n\n", strings.ToUpper(string(kr.mode))))
	
	// Group bindings by category
	categories := map[string][]KeyBinding{
		"Global":     {},
		"Navigation": {},
		"Panel":      {},
		"Content":    {},
		"Modal":      {},
	}
	
	for _, binding := range kr.GetAllBindings() {
		switch binding.Action {
		case ActionQuit, ActionToggleHelp, ActionToggleDebug, ActionRefresh, ActionEscape:
			categories["Global"] = append(categories["Global"], binding)
		case ActionMoveUp, ActionMoveDown, ActionMoveLeft, ActionMoveRight, 
			 ActionPageUp, ActionPageDown, ActionGoToTop, ActionGoToBottom:
			categories["Navigation"] = append(categories["Navigation"], binding)
		case ActionNextPanel, ActionPrevPanel, ActionNextTab, ActionPrevTab,
			 ActionFocusMain, ActionFocusDetail, ActionFocusLog:
			categories["Panel"] = append(categories["Panel"], binding)
		case ActionSelect, ActionDelete, ActionEdit, ActionCopy, ActionToggleCollapse:
			categories["Content"] = append(categories["Content"], binding)
		case ActionEnterSearch, ActionEnterCommand, ActionEnterInsert, ActionEnterNormal:
			categories["Modal"] = append(categories["Modal"], binding)
		}
	}
	
	// Format each category
	for category, bindings := range categories {
		if len(bindings) > 0 {
			sb.WriteString(fmt.Sprintf("%s:\n", category))
			for _, binding := range bindings {
				sb.WriteString(fmt.Sprintf("  %-12s %s\n", binding.Key, binding.Description))
			}
			sb.WriteString("\n")
		}
	}
	
	return sb.String()
}

// GetModeIndicator returns a string representation of the current mode for status display
func (kr *KeybindingRegistry) GetModeIndicator() string {
	switch kr.mode {
	case ModeNormal:
		return "NORMAL"
	case ModeSearch:
		return "SEARCH"
	case ModeCommand:
		return "COMMAND"
	case ModeInsert:
		return "INSERT"
	default:
		return "UNKNOWN"
	}
}