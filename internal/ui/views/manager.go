package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/katyella/lazyoc/internal/k8s/resources"
	"github.com/katyella/lazyoc/internal/ui/models"
)

// ViewType represents different view types in the application
type ViewType int

const (
	ViewTypePods ViewType = iota
	ViewTypeContainers
	ViewTypeLogs
	ViewTypeResources
)

// ViewContext contains data needed by views
type ViewContext struct {
	App          *models.App
	Width        int
	Height       int
	FocusedPanel int
	Connected    bool
	Namespace    string
	Pods         []resources.PodInfo
	SelectedPod  int
}

// View represents a view component that can render content and handle updates
type View interface {
	// Update handles tea messages and returns updated model and commands
	Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd)
	
	// Render returns the string representation of the view
	Render(ctx ViewContext) string
	
	// GetType returns the view type
	GetType() ViewType
	
	// CanHandle returns true if this view can handle the given message
	CanHandle(msg tea.Msg) bool
}

// ViewManager manages different views and handles view switching
type ViewManager struct {
	views       map[ViewType]View
	activeView  ViewType
	initialized bool
}

// NewViewManager creates a new view manager
func NewViewManager() *ViewManager {
	return &ViewManager{
		views:       make(map[ViewType]View),
		activeView:  ViewTypePods,
		initialized: false,
	}
}

// RegisterView registers a view with the manager
func (vm *ViewManager) RegisterView(viewType ViewType, view View) {
	vm.views[viewType] = view
}

// SetActiveView switches to the specified view
func (vm *ViewManager) SetActiveView(viewType ViewType) {
	if _, exists := vm.views[viewType]; exists {
		vm.activeView = viewType
	}
}

// GetActiveView returns the currently active view
func (vm *ViewManager) GetActiveView() ViewType {
	return vm.activeView
}

// Update handles messages for the active view
func (vm *ViewManager) Update(msg tea.Msg, ctx ViewContext) tea.Cmd {
	if view, exists := vm.views[vm.activeView]; exists {
		if view.CanHandle(msg) {
			updatedView, cmd := view.Update(msg, ctx)
			vm.views[vm.activeView] = updatedView
			return cmd
		}
	}
	return nil
}

// Render renders the active view
func (vm *ViewManager) Render(ctx ViewContext) string {
	if view, exists := vm.views[vm.activeView]; exists {
		return view.Render(ctx)
	}
	return "No active view"
}

// Initialize sets up the view manager with default views
func (vm *ViewManager) Initialize() {
	if vm.initialized {
		return
	}
	
	// Register default views
	vm.RegisterView(ViewTypePods, NewPodsView())
	vm.RegisterView(ViewTypeLogs, NewLogsView())
	vm.RegisterView(ViewTypeContainers, NewContainersView())
	vm.RegisterView(ViewTypeResources, NewResourcesView())
	
	vm.initialized = true
}

// IsInitialized returns whether the view manager has been initialized
func (vm *ViewManager) IsInitialized() bool {
	return vm.initialized
}