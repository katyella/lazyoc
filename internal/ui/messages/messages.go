package messages

import tea "github.com/charmbracelet/bubbletea"

// WindowSizeMsg represents a terminal window resize event
type WindowSizeMsg struct {
	Width  int
	Height int
}

// ErrorMsg represents an error that occurred in the application
type ErrorMsg struct {
	Err error
}

// LoadingMsg represents a loading state change
type LoadingMsg struct {
	Message string
	Loading bool
}

// TabSwitchMsg represents a tab switch request
type TabSwitchMsg struct {
	TabIndex int
}

// HelpToggleMsg represents a help overlay toggle request
type HelpToggleMsg struct{}

// QuitMsg represents a quit application request
type QuitMsg struct{}

// KeyMsg represents a keyboard input message
type KeyMsg tea.KeyMsg

// InitMsg represents the initial application setup message
type InitMsg struct{}

// ReadyMsg represents when the application is ready for use
type ReadyMsg struct{}

// DebugToggleMsg represents a debug mode toggle
type DebugToggleMsg struct{}

// RefreshMsg represents a refresh request for data
type RefreshMsg struct{}

// StatusMsg represents a status bar message
type StatusMsg struct {
	Message string
	Type    StatusType
}

// StatusType represents the type of status message
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
)

// ConnectedMsg represents successful cluster connection
type ConnectedMsg struct {
	ClusterName string
	Namespace   string
}

// DisconnectedMsg represents cluster disconnection
type DisconnectedMsg struct{}

// SpinnerTick represents a spinner animation tick
type SpinnerTick struct{}

// LoadPodLogsMsg represents a request to load pod logs
type LoadPodLogsMsg struct {
	PodName   string
	Namespace string
}
