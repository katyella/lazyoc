package state

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Manager manages application state
type Manager struct {
	mu sync.RWMutex

	// State slices
	connection ConnectionState
	ui         UIState
	resources  ResourceState

	// Sub-managers
	selection *SelectionManager
	filters   *FilterManager

	// Subscribers
	subscribers []func(StateChange)
}

// ConnectionState represents the connection state
type ConnectionState struct {
	Status         ConnectionStatus
	Kubeconfig     string
	Context        string
	Namespace      string
	ClusterType    string
	ClusterVersion string
	Error          error
}

// ConnectionStatus represents the connection status
type ConnectionStatus int

const (
	ConnectionStatusDisconnected ConnectionStatus = iota
	ConnectionStatusConnecting
	ConnectionStatusConnected
	ConnectionStatusError
)

// UIState represents the UI state
type UIState struct {
	ActiveTab     int
	ActiveTabName string
	FocusedPanel  string
	ShowDetails   bool
	ShowLogs      bool
	ModalVisible  bool
	ModalType     string
}

// ResourceState represents the resource state
type ResourceState struct {
	Pods          []interface{} // TODO: Use proper types
	Services      []interface{}
	Deployments   []interface{}
	ConfigMaps    []interface{}
	Secrets       []interface{}
	SelectedIndex int
	SelectedItem  interface{}
}

// StateChange represents a state change event
type StateChange struct {
	Type     StateChangeType
	OldValue interface{}
	NewValue interface{}
}

// StateChangeType represents the type of state change
type StateChangeType int

const (
	ChangeTypeConnection StateChangeType = iota
	ChangeTypeUI
	ChangeTypeResources
	ChangeTypeSelection
)

// NewManager creates a new state manager
func NewManager() *Manager {
	m := &Manager{
		connection: ConnectionState{
			Status:    ConnectionStatusDisconnected,
			Namespace: "default",
		},
		ui: UIState{
			ActiveTab:    0,
			ShowDetails:  true,
			ShowLogs:     true,
			FocusedPanel: "main",
		},
		resources: ResourceState{
			Pods:        make([]interface{}, 0),
			Services:    make([]interface{}, 0),
			Deployments: make([]interface{}, 0),
			ConfigMaps:  make([]interface{}, 0),
			Secrets:     make([]interface{}, 0),
		},
		selection:   NewSelectionManager(),
		filters:     NewFilterManager(),
		subscribers: make([]func(StateChange), 0),
	}

	// Set up observer patterns
	m.selection.AddObserver(m)
	m.filters.AddObserver(m)

	return m
}

// Subscribe adds a state change subscriber
func (m *Manager) Subscribe(handler func(StateChange)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = append(m.subscribers, handler)
}

// notify notifies all subscribers of a state change
func (m *Manager) notify(change StateChange) {
	for _, handler := range m.subscribers {
		handler(change)
	}
}

// Update handles state update messages
func (m *Manager) Update(msg tea.Msg) tea.Cmd {
	// Handle state-related messages
	// TODO: Implement message handling
	return nil
}

// Connection state methods

// GetConnectionState returns the current connection state
func (m *Manager) GetConnectionState() ConnectionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connection
}

// SetConnectionStatus updates the connection status
func (m *Manager) SetConnectionStatus(status ConnectionStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.connection
	m.connection.Status = status

	m.notify(StateChange{
		Type:     ChangeTypeConnection,
		OldValue: oldState,
		NewValue: m.connection,
	})
}

// SetKubeconfig sets the kubeconfig path
func (m *Manager) SetKubeconfig(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connection.Kubeconfig = path
	return nil
}

// SetClusterInfo updates the cluster information
func (m *Manager) SetClusterInfo(clusterType, version, context, namespace string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.connection
	m.connection.ClusterType = clusterType
	m.connection.ClusterVersion = version
	m.connection.Context = context
	m.connection.Namespace = namespace

	m.notify(StateChange{
		Type:     ChangeTypeConnection,
		OldValue: oldState,
		NewValue: m.connection,
	})
}

// UI state methods

// GetUIState returns the current UI state
func (m *Manager) GetUIState() UIState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ui
}

// SetActiveTab updates the active tab
func (m *Manager) SetActiveTab(index int, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.ui
	m.ui.ActiveTab = index
	m.ui.ActiveTabName = name

	m.notify(StateChange{
		Type:     ChangeTypeUI,
		OldValue: oldState,
		NewValue: m.ui,
	})
}

// SetFocusedPanel updates the focused panel
func (m *Manager) SetFocusedPanel(panel string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.ui
	m.ui.FocusedPanel = panel

	m.notify(StateChange{
		Type:     ChangeTypeUI,
		OldValue: oldState,
		NewValue: m.ui,
	})
}

// SetPanelVisibility updates panel visibility
func (m *Manager) SetPanelVisibility(showDetails, showLogs bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.ui
	m.ui.ShowDetails = showDetails
	m.ui.ShowLogs = showLogs

	m.notify(StateChange{
		Type:     ChangeTypeUI,
		OldValue: oldState,
		NewValue: m.ui,
	})
}

// Resource state methods

// GetResourceState returns the current resource state
func (m *Manager) GetResourceState() ResourceState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.resources
}

// SetPods updates the pods list
func (m *Manager) SetPods(pods []interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.resources
	m.resources.Pods = pods

	m.notify(StateChange{
		Type:     ChangeTypeResources,
		OldValue: oldState,
		NewValue: m.resources,
	})
}

// SetSelectedIndex updates the selected resource index
func (m *Manager) SetSelectedIndex(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.resources
	m.resources.SelectedIndex = index

	// Update selected item based on active tab
	switch m.ui.ActiveTabName {
	case "Pods":
		if index >= 0 && index < len(m.resources.Pods) {
			m.resources.SelectedItem = m.resources.Pods[index]
		}
	case "Services":
		if index >= 0 && index < len(m.resources.Services) {
			m.resources.SelectedItem = m.resources.Services[index]
		}
		// TODO: Add other resource types
	}

	m.notify(StateChange{
		Type:     ChangeTypeSelection,
		OldValue: oldState,
		NewValue: m.resources,
	})
}

// GetSelectionManager returns the selection manager
func (m *Manager) GetSelectionManager() *SelectionManager {
	return m.selection
}

// GetFilterManager returns the filter manager
func (m *Manager) GetFilterManager() *FilterManager {
	return m.filters
}

// SelectionObserver implementation

// OnSelectionChanged handles selection changes
func (m *Manager) OnSelectionChanged(old, new *SelectedResource) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update resource state
	if new != nil {
		m.resources.SelectedItem = new.Details
	} else {
		m.resources.SelectedItem = nil
	}

	m.notify(StateChange{
		Type:     ChangeTypeSelection,
		OldValue: old,
		NewValue: new,
	})
}

// FilterObserver implementation

// OnFilterChanged handles filter changes
func (m *Manager) OnFilterChanged(resourceType string, filter *ResourceFilter) {
	// Notify subscribers of filter change
	m.notify(StateChange{
		Type:     ChangeTypeUI,
		OldValue: resourceType,
		NewValue: filter,
	})
}
