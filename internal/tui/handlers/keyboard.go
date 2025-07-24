package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
)

// KeyHandler handles keyboard input
type KeyHandler struct {
	bindings map[string]func() tea.Cmd
}

// NewKeyHandler creates a new keyboard handler
func NewKeyHandler() *KeyHandler {
	return &KeyHandler{
		bindings: make(map[string]func() tea.Cmd),
	}
}

// Bind adds a key binding
func (k *KeyHandler) Bind(key string, action func() tea.Cmd) {
	k.bindings[key] = action
}

// BindMultiple adds multiple key bindings for the same action
func (k *KeyHandler) BindMultiple(keys []string, action func() tea.Cmd) {
	for _, key := range keys {
		k.bindings[key] = action
	}
}

// Handle processes a key message
func (k *KeyHandler) Handle(msg tea.KeyMsg) tea.Cmd {
	if action, ok := k.bindings[msg.String()]; ok {
		return action()
	}
	return nil
}

// DefaultKeyBindings returns standard key bindings
func DefaultKeyBindings() map[string]string {
	return map[string]string{
		// Navigation
		"up":    "Move up",
		"k":     "Move up",
		"down":  "Move down", 
		"j":     "Move down",
		"left":  "Move left",
		"h":     "Move left",
		"right": "Move right",
		"pgup":  "Page up",
		"pgdown": "Page down",
		"home":  "Go to top",
		"g":     "Go to top",
		"end":   "Go to bottom",
		"G":     "Go to bottom",
		
		// Tabs
		"tab":       "Next tab",
		"shift+tab": "Previous tab",
		"1":         "Go to tab 1",
		"2":         "Go to tab 2",
		"3":         "Go to tab 3",
		"4":         "Go to tab 4",
		"5":         "Go to tab 5",
		
		// Actions
		"enter":  "Select/Open",
		"space":  "Select/Toggle",
		"d":      "Delete",
		"e":      "Edit",
		"v":      "View",
		"l":      "Show logs",
		"r":      "Refresh",
		"f":      "Follow logs",
		"s":      "Shell/Exec",
		
		// Panels
		"ctrl+j": "Focus next panel",
		"ctrl+k": "Focus previous panel",
		"ctrl+h": "Hide/Show details",
		"ctrl+l": "Hide/Show logs",
		
		// General
		"?":      "Show help",
		"ctrl+c": "Quit",
		"q":      "Quit",
		"esc":    "Cancel/Back",
		"/":      "Search",
		"n":      "Next search result",
		"N":      "Previous search result",
	}
}

// GlobalKeyHandler handles global key bindings that work across all views
type GlobalKeyHandler struct {
	*KeyHandler
	quitKeys   []string
	helpKeys   []string
	searchKeys []string
}

// NewGlobalKeyHandler creates a new global key handler
func NewGlobalKeyHandler() *GlobalKeyHandler {
	return &GlobalKeyHandler{
		KeyHandler: NewKeyHandler(),
		quitKeys:   []string{"ctrl+c", "q"},
		helpKeys:   []string{"?"},
		searchKeys: []string{"/"},
	}
}

// IsQuitKey checks if the key should quit the application
func (g *GlobalKeyHandler) IsQuitKey(key string) bool {
	for _, qk := range g.quitKeys {
		if key == qk {
			return true
		}
	}
	return false
}

// IsHelpKey checks if the key should show help
func (g *GlobalKeyHandler) IsHelpKey(key string) bool {
	for _, hk := range g.helpKeys {
		if key == hk {
			return true
		}
	}
	return false
}

// IsSearchKey checks if the key should start search
func (g *GlobalKeyHandler) IsSearchKey(key string) bool {
	for _, sk := range g.searchKeys {
		if key == sk {
			return true
		}
	}
	return false
}