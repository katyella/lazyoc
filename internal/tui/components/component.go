package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Component represents a UI component that can be rendered and handle events
type Component interface {
	// Init initializes the component and returns any initial commands
	Init() tea.Cmd

	// Update handles messages and updates the component state
	Update(msg tea.Msg) (tea.Cmd, error)

	// View renders the component to a string
	View() string

	// Focus sets the component as focused
	Focus() error

	// Blur removes focus from the component
	Blur() error

	// IsFocused returns whether the component is currently focused
	IsFocused() bool

	// SetSize updates the component's dimensions
	SetSize(width, height int)

	// GetSize returns the component's current dimensions
	GetSize() (width, height int)
}

// BaseComponent provides common functionality for all components
type BaseComponent struct {
	focused bool
	width   int
	height  int
}

// Focus sets the component as focused
func (b *BaseComponent) Focus() error {
	b.focused = true
	return nil
}

// Blur removes focus from the component
func (b *BaseComponent) Blur() error {
	b.focused = false
	return nil
}

// IsFocused returns whether the component is currently focused
func (b *BaseComponent) IsFocused() bool {
	return b.focused
}

// SetSize updates the component's dimensions
func (b *BaseComponent) SetSize(width, height int) {
	b.width = width
	b.height = height
}

// GetSize returns the component's current dimensions
func (b *BaseComponent) GetSize() (width, height int) {
	return b.width, b.height
}

// Message types for component communication
type (
	// FocusMsg is sent when a component should receive focus
	FocusMsg struct {
		ComponentID string
	}

	// BlurMsg is sent when a component should lose focus
	BlurMsg struct {
		ComponentID string
	}

	// ResizeMsg is sent when the terminal is resized
	ResizeMsg struct {
		Width  int
		Height int
	}

	// RefreshMsg is sent when a component should refresh its content
	RefreshMsg struct {
		ComponentID string
	}
)

// ComponentRegistry manages component registration and lookup
type ComponentRegistry struct {
	components map[string]Component
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]Component),
	}
}

// Register adds a component to the registry
func (r *ComponentRegistry) Register(id string, component Component) {
	r.components[id] = component
}

// Get retrieves a component by ID
func (r *ComponentRegistry) Get(id string) (Component, bool) {
	component, exists := r.components[id]
	return component, exists
}

// All returns all registered components
func (r *ComponentRegistry) All() map[string]Component {
	return r.components
}

// InitAll initializes all registered components
func (r *ComponentRegistry) InitAll() tea.Cmd {
	var cmds []tea.Cmd
	for _, component := range r.components {
		if cmd := component.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// UpdateAll updates all components with a message
func (r *ComponentRegistry) UpdateAll(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for _, component := range r.components {
		if cmd, err := component.Update(msg); err == nil && cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// ResizeAll resizes all components
func (r *ComponentRegistry) ResizeAll(width, height int) {
	for _, component := range r.components {
		component.SetSize(width, height)
	}
}