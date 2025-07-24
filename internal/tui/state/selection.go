package state

import (
	"sync"
)

// SelectionManager manages resource selection state
type SelectionManager struct {
	mu sync.RWMutex
	
	// Selection state per resource type
	selections map[string]*ResourceSelection
	
	// Global selected resource
	current *SelectedResource
	
	// Observers
	observers []SelectionObserver
}

// ResourceSelection tracks selection for a resource type
type ResourceSelection struct {
	ResourceType  string
	SelectedIndex int
	SelectedID    string
	Items         []SelectableItem
}

// SelectableItem represents an item that can be selected
type SelectableItem struct {
	ID        string
	Name      string
	Namespace string
	Metadata  map[string]interface{}
}

// SelectedResource represents the currently selected resource
type SelectedResource struct {
	Type      string
	Name      string
	Namespace string
	ID        string
	Details   map[string]interface{}
}

// SelectionObserver receives selection change notifications
type SelectionObserver interface {
	OnSelectionChanged(old, new *SelectedResource)
}

// NewSelectionManager creates a new selection manager
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{
		selections: make(map[string]*ResourceSelection),
		observers:  make([]SelectionObserver, 0),
	}
}

// AddObserver adds a selection observer
func (s *SelectionManager) AddObserver(observer SelectionObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// RemoveObserver removes a selection observer
func (s *SelectionManager) RemoveObserver(observer SelectionObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// SetItems sets the selectable items for a resource type
func (s *SelectionManager) SetItems(resourceType string, items []SelectableItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	selection, exists := s.selections[resourceType]
	if !exists {
		selection = &ResourceSelection{
			ResourceType:  resourceType,
			SelectedIndex: 0,
			Items:         items,
		}
		s.selections[resourceType] = selection
	} else {
		selection.Items = items
		// Adjust index if out of bounds
		if selection.SelectedIndex >= len(items) {
			selection.SelectedIndex = len(items) - 1
			if selection.SelectedIndex < 0 {
				selection.SelectedIndex = 0
			}
		}
	}
	
	// Update selected ID
	if selection.SelectedIndex < len(items) {
		selection.SelectedID = items[selection.SelectedIndex].ID
	} else {
		selection.SelectedID = ""
	}
}

// SelectByIndex selects an item by index for a resource type
func (s *SelectionManager) SelectByIndex(resourceType string, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	selection, exists := s.selections[resourceType]
	if !exists || len(selection.Items) == 0 {
		return nil
	}
	
	if index < 0 || index >= len(selection.Items) {
		return nil
	}
	
	oldCurrent := s.current
	selection.SelectedIndex = index
	item := selection.Items[index]
	selection.SelectedID = item.ID
	
	// Update current selection
	s.current = &SelectedResource{
		Type:      resourceType,
		Name:      item.Name,
		Namespace: item.Namespace,
		ID:        item.ID,
		Details:   item.Metadata,
	}
	
	// Notify observers
	s.notifyObservers(oldCurrent, s.current)
	
	return nil
}

// SelectByID selects an item by ID for a resource type
func (s *SelectionManager) SelectByID(resourceType string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	selection, exists := s.selections[resourceType]
	if !exists || len(selection.Items) == 0 {
		return nil
	}
	
	// Find item by ID
	for i, item := range selection.Items {
		if item.ID == id {
			oldCurrent := s.current
			selection.SelectedIndex = i
			selection.SelectedID = id
			
			s.current = &SelectedResource{
				Type:      resourceType,
				Name:      item.Name,
				Namespace: item.Namespace,
				ID:        item.ID,
				Details:   item.Metadata,
			}
			
			s.notifyObservers(oldCurrent, s.current)
			return nil
		}
	}
	
	return nil
}

// GetSelectedIndex returns the selected index for a resource type
func (s *SelectionManager) GetSelectedIndex(resourceType string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if selection, exists := s.selections[resourceType]; exists {
		return selection.SelectedIndex
	}
	return 0
}

// GetSelectedItem returns the selected item for a resource type
func (s *SelectionManager) GetSelectedItem(resourceType string) *SelectableItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	selection, exists := s.selections[resourceType]
	if !exists || selection.SelectedIndex >= len(selection.Items) {
		return nil
	}
	
	return &selection.Items[selection.SelectedIndex]
}

// GetCurrent returns the currently selected resource
func (s *SelectionManager) GetCurrent() *SelectedResource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Clear clears the selection for a resource type
func (s *SelectionManager) Clear(resourceType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.selections, resourceType)
	
	if s.current != nil && s.current.Type == resourceType {
		oldCurrent := s.current
		s.current = nil
		s.notifyObservers(oldCurrent, nil)
	}
}

// ClearAll clears all selections
func (s *SelectionManager) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	oldCurrent := s.current
	s.selections = make(map[string]*ResourceSelection)
	s.current = nil
	
	if oldCurrent != nil {
		s.notifyObservers(oldCurrent, nil)
	}
}

// notifyObservers notifies all observers of a selection change
func (s *SelectionManager) notifyObservers(old, new *SelectedResource) {
	for _, observer := range s.observers {
		observer.OnSelectionChanged(old, new)
	}
}

// MoveSelection moves the selection by a delta
func (s *SelectionManager) MoveSelection(resourceType string, delta int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	selection, exists := s.selections[resourceType]
	if !exists || len(selection.Items) == 0 {
		return nil
	}
	
	newIndex := selection.SelectedIndex + delta
	
	// Clamp to valid range
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(selection.Items) {
		newIndex = len(selection.Items) - 1
	}
	
	if newIndex != selection.SelectedIndex {
		s.mu.Unlock()
		err := s.SelectByIndex(resourceType, newIndex)
		s.mu.Lock()
		return err
	}
	
	return nil
}