package state

import (
	"strings"
	"sync"
)

// FilterManager manages resource filtering and search state
type FilterManager struct {
	mu sync.RWMutex
	
	// Filters per resource type
	filters map[string]*ResourceFilter
	
	// Global search
	globalSearch string
	
	// Observers
	observers []FilterObserver
}

// ResourceFilter represents filters for a resource type
type ResourceFilter struct {
	ResourceType string
	SearchTerm   string
	Labels       map[string]string
	Namespaces   []string
	Statuses     []string
	Custom       map[string]interface{}
}

// FilterObserver receives filter change notifications
type FilterObserver interface {
	OnFilterChanged(resourceType string, filter *ResourceFilter)
}

// NewFilterManager creates a new filter manager
func NewFilterManager() *FilterManager {
	return &FilterManager{
		filters:   make(map[string]*ResourceFilter),
		observers: make([]FilterObserver, 0),
	}
}

// AddObserver adds a filter observer
func (f *FilterManager) AddObserver(observer FilterObserver) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.observers = append(f.observers, observer)
}

// RemoveObserver removes a filter observer
func (f *FilterManager) RemoveObserver(observer FilterObserver) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	for i, obs := range f.observers {
		if obs == observer {
			f.observers = append(f.observers[:i], f.observers[i+1:]...)
			break
		}
	}
}

// SetSearch sets the search term for a resource type
func (f *FilterManager) SetSearch(resourceType string, searchTerm string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	filter := f.getOrCreateFilter(resourceType)
	filter.SearchTerm = searchTerm
	
	f.notifyObservers(resourceType, filter)
}

// SetGlobalSearch sets the global search term
func (f *FilterManager) SetGlobalSearch(searchTerm string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.globalSearch = searchTerm
	
	// Update all resource filters
	for resourceType, filter := range f.filters {
		filter.SearchTerm = searchTerm
		f.notifyObservers(resourceType, filter)
	}
}

// SetLabels sets label filters for a resource type
func (f *FilterManager) SetLabels(resourceType string, labels map[string]string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	filter := f.getOrCreateFilter(resourceType)
	filter.Labels = labels
	
	f.notifyObservers(resourceType, filter)
}

// AddLabel adds a single label filter
func (f *FilterManager) AddLabel(resourceType string, key, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	filter := f.getOrCreateFilter(resourceType)
	if filter.Labels == nil {
		filter.Labels = make(map[string]string)
	}
	filter.Labels[key] = value
	
	f.notifyObservers(resourceType, filter)
}

// RemoveLabel removes a label filter
func (f *FilterManager) RemoveLabel(resourceType string, key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	filter := f.getOrCreateFilter(resourceType)
	if filter.Labels != nil {
		delete(filter.Labels, key)
		f.notifyObservers(resourceType, filter)
	}
}

// SetNamespaces sets namespace filters
func (f *FilterManager) SetNamespaces(resourceType string, namespaces []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	filter := f.getOrCreateFilter(resourceType)
	filter.Namespaces = namespaces
	
	f.notifyObservers(resourceType, filter)
}

// SetStatuses sets status filters
func (f *FilterManager) SetStatuses(resourceType string, statuses []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	filter := f.getOrCreateFilter(resourceType)
	filter.Statuses = statuses
	
	f.notifyObservers(resourceType, filter)
}

// GetFilter returns the filter for a resource type
func (f *FilterManager) GetFilter(resourceType string) *ResourceFilter {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	return f.filters[resourceType]
}

// ClearFilter clears filters for a resource type
func (f *FilterManager) ClearFilter(resourceType string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if _, exists := f.filters[resourceType]; exists {
		delete(f.filters, resourceType)
		f.notifyObservers(resourceType, nil)
	}
}

// ClearAllFilters clears all filters
func (f *FilterManager) ClearAllFilters() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	for resourceType := range f.filters {
		f.notifyObservers(resourceType, nil)
	}
	
	f.filters = make(map[string]*ResourceFilter)
	f.globalSearch = ""
}

// ApplyFilter checks if an item matches the filter
func (f *FilterManager) ApplyFilter(resourceType string, item map[string]interface{}) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	filter, exists := f.filters[resourceType]
	if !exists || filter == nil {
		return true
	}
	
	// Check search term
	if filter.SearchTerm != "" {
		if !f.matchesSearch(item, filter.SearchTerm) {
			return false
		}
	}
	
	// Check namespace filter
	if len(filter.Namespaces) > 0 {
		namespace := f.getItemNamespace(item)
		if !f.containsString(filter.Namespaces, namespace) {
			return false
		}
	}
	
	// Check status filter
	if len(filter.Statuses) > 0 {
		status := f.getItemStatus(item)
		if !f.containsString(filter.Statuses, status) {
			return false
		}
	}
	
	// Check label filters
	if len(filter.Labels) > 0 {
		itemLabels := f.getItemLabels(item)
		for key, value := range filter.Labels {
			if itemValue, exists := itemLabels[key]; !exists || itemValue != value {
				return false
			}
		}
	}
	
	return true
}

// getOrCreateFilter gets or creates a filter for a resource type
func (f *FilterManager) getOrCreateFilter(resourceType string) *ResourceFilter {
	filter, exists := f.filters[resourceType]
	if !exists {
		filter = &ResourceFilter{
			ResourceType: resourceType,
			Labels:       make(map[string]string),
			Namespaces:   make([]string, 0),
			Statuses:     make([]string, 0),
			Custom:       make(map[string]interface{}),
		}
		f.filters[resourceType] = filter
	}
	return filter
}

// notifyObservers notifies all observers of a filter change
func (f *FilterManager) notifyObservers(resourceType string, filter *ResourceFilter) {
	for _, observer := range f.observers {
		observer.OnFilterChanged(resourceType, filter)
	}
}

// Helper functions

func (f *FilterManager) matchesSearch(item map[string]interface{}, searchTerm string) bool {
	search := strings.ToLower(searchTerm)
	
	// Check name
	if name, ok := item["name"].(string); ok {
		if strings.Contains(strings.ToLower(name), search) {
			return true
		}
	}
	
	// Check namespace
	if namespace, ok := item["namespace"].(string); ok {
		if strings.Contains(strings.ToLower(namespace), search) {
			return true
		}
	}
	
	// Check labels
	if labels, ok := item["labels"].(map[string]interface{}); ok {
		for key, value := range labels {
			if strings.Contains(strings.ToLower(key), search) {
				return true
			}
			if strValue, ok := value.(string); ok {
				if strings.Contains(strings.ToLower(strValue), search) {
					return true
				}
			}
		}
	}
	
	return false
}

func (f *FilterManager) getItemNamespace(item map[string]interface{}) string {
	if metadata, ok := item["metadata"].(map[string]interface{}); ok {
		if namespace, ok := metadata["namespace"].(string); ok {
			return namespace
		}
	}
	return ""
}

func (f *FilterManager) getItemStatus(item map[string]interface{}) string {
	if status, ok := item["status"].(map[string]interface{}); ok {
		if phase, ok := status["phase"].(string); ok {
			return phase
		}
	}
	return ""
}

func (f *FilterManager) getItemLabels(item map[string]interface{}) map[string]string {
	labels := make(map[string]string)
	
	if metadata, ok := item["metadata"].(map[string]interface{}); ok {
		if itemLabels, ok := metadata["labels"].(map[string]interface{}); ok {
			for key, value := range itemLabels {
				if strValue, ok := value.(string); ok {
					labels[key] = strValue
				}
			}
		}
	}
	
	return labels
}

func (f *FilterManager) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}