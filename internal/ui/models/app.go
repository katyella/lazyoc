package models

import (
	"log"
	"time"
)

// AppState represents the current state/mode of the application
type AppState int

const (
	StateLoading AppState = iota
	StateMain
	StateHelp
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
	Loading bool
	LoadingMessage string

	// Help state
	ShowHelp bool

	// Application metadata
	Version string
	StartTime time.Time
}

// NewApp creates a new application model with default values
func NewApp(version string) *App {
	return &App{
		State:     StateLoading,
		ViewMode:  ViewResourceList,
		ActiveTab: TabPods,
		Debug:     false,
		Loading:   true,
		LoadingMessage: "Initializing LazyOC...",
		Version:   version,
		StartTime: time.Now(),
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
	switch a.ActiveTab {
	case TabPods:
		a.ActiveTab = TabServices
	case TabServices:
		a.ActiveTab = TabDeployments
	case TabDeployments:
		a.ActiveTab = TabConfigMaps
	case TabConfigMaps:
		a.ActiveTab = TabSecrets
	case TabSecrets:
		a.ActiveTab = TabPods
	}
}

// PrevTab switches to the previous resource tab
func (a *App) PrevTab() {
	switch a.ActiveTab {
	case TabPods:
		a.ActiveTab = TabSecrets
	case TabServices:
		a.ActiveTab = TabPods
	case TabDeployments:
		a.ActiveTab = TabServices
	case TabConfigMaps:
		a.ActiveTab = TabDeployments
	case TabSecrets:
		a.ActiveTab = TabConfigMaps
	}
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
	default:
		return "Unknown"
	}
}