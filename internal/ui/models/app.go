// Package models defines the core data structures and state management for the LazyOC UI.
// It provides application state models, view modes, and tab management for the terminal interface.
package models

import (
	"log"
	"time"
)

// AppState represents the current state/mode of the application
type AppState int

const (
	// StateLoading indicates the application is initializing or loading data
	StateLoading AppState = iota

	// StateMain indicates the application is in normal operation mode
	StateMain

	// StateHelp indicates the help overlay is currently displayed
	StateHelp

	// StateError indicates the application is displaying an error state
	StateError
)

// ViewMode represents the currently active view/panel
type ViewMode int

const (
	ViewResourceList ViewMode = iota
	ViewResourceDetail
	ViewLogs
)

// TabType represents the different resource tabs
type TabType int

const (
	TabPods TabType = iota
	TabServices
	TabDeployments
	TabConfigMaps
	TabSecrets
	// OpenShift-specific tabs
	TabBuildConfigs
	TabImageStreams
	TabRoutes
)

// App represents the main application model
type App struct {
	// Core application state
	State     AppState
	ViewMode  ViewMode
	ActiveTab TabType

	// UI dimensions
	Width  int
	Height int

	// Error state
	LastError error
	ErrorTime time.Time

	// Debug mode
	Debug bool

	// Logger for debugging
	Logger *log.Logger

	// Loading state
	Loading        bool
	LoadingMessage string

	// Help state
	ShowHelp bool

	// Application metadata
	Version   string
	StartTime time.Time
}

// NewApp creates a new application model with default values
func NewApp(version string) *App {
	return &App{
		State:          StateLoading,
		ViewMode:       ViewResourceList,
		ActiveTab:      TabPods,
		Debug:          false,
		Loading:        true,
		LoadingMessage: "Initializing LazyOC...",
		Version:        version,
		StartTime:      time.Now(),
	}
}

// SetError sets the application error state
func (a *App) SetError(err error) {
	a.LastError = err
	a.ErrorTime = time.Now()
	a.State = StateError
	a.Loading = false

	if a.Logger != nil {
		a.Logger.Printf("Error: %v", err)
	}
}

// ClearError clears the error state and returns to main state
func (a *App) ClearError() {
	a.LastError = nil
	a.State = StateMain
}

// SetLoading sets the loading state with a message
func (a *App) SetLoading(message string) {
	a.Loading = true
	a.LoadingMessage = message
}

// ClearLoading clears the loading state
func (a *App) ClearLoading() {
	a.Loading = false
	a.LoadingMessage = ""
}

// ToggleHelp toggles the help overlay
func (a *App) ToggleHelp() {
	a.ShowHelp = !a.ShowHelp
	if a.ShowHelp {
		a.State = StateHelp
	} else {
		a.State = StateMain
	}
}

// SetDimensions updates the terminal dimensions
func (a *App) SetDimensions(width, height int) {
	a.Width = width
	a.Height = height
}

// IsReady returns true if the app is ready for normal operation
func (a *App) IsReady() bool {
	return a.State == StateMain && !a.Loading
}

// GetStatusMessage returns the current status message
func (a *App) GetStatusMessage() string {
	switch a.State {
	case StateLoading:
		return a.LoadingMessage
	case StateError:
		if a.LastError != nil {
			return a.LastError.Error()
		}
		return "Unknown error occurred"
	default:
		return "Ready"
	}
}

// NextTab switches to the next resource tab
func (a *App) NextTab() {
	// Get all available tabs in order (matching constants.ResourceTabs)
	tabs := []TabType{
		TabPods, TabServices, TabDeployments, TabConfigMaps, TabSecrets,
		TabBuildConfigs, TabImageStreams, TabRoutes,
	}

	// Find current tab index and move to next
	for i, tab := range tabs {
		if tab == a.ActiveTab {
			a.ActiveTab = tabs[(i+1)%len(tabs)]
			return
		}
	}
	// Fallback to first tab if current tab not found
	a.ActiveTab = TabPods
}

// PrevTab switches to the previous resource tab
func (a *App) PrevTab() {
	// Get all available tabs in order (matching constants.ResourceTabs)
	tabs := []TabType{
		TabPods, TabServices, TabDeployments, TabConfigMaps, TabSecrets,
		TabBuildConfigs, TabImageStreams, TabRoutes,
	}

	// Find current tab index and move to previous
	for i, tab := range tabs {
		if tab == a.ActiveTab {
			a.ActiveTab = tabs[(i-1+len(tabs))%len(tabs)]
			return
		}
	}
	// Fallback to first tab if current tab not found
	a.ActiveTab = TabPods
}

// GetTabName returns the display name for a tab
func (a *App) GetTabName(tab TabType) string {
	switch tab {
	case TabPods:
		return "Pods"
	case TabServices:
		return "Services"
	case TabDeployments:
		return "Deployments"
	case TabConfigMaps:
		return "ConfigMaps"
	case TabSecrets:
		return "Secrets"
	// OpenShift-specific tabs
	case TabBuildConfigs:
		return "BuildConfigs"
	case TabImageStreams:
		return "ImageStreams"
	case TabRoutes:
		return "Routes"
	default:
		return "Unknown"
	}
}
