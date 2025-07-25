package services

import (
	"context"
	"sync"
	"time"
)

// WatchService manages resource watching
type WatchService struct {
	mu sync.RWMutex

	// Kubernetes service
	k8s *KubernetesService

	// Active watchers
	watchers map[string]*resourceWatcher

	// Observers
	observers []WatchObserver
}

// resourceWatcher represents an active resource watcher
type resourceWatcher struct {
	resourceType string
	context      context.Context
	cancel       context.CancelFunc
	eventChan    <-chan ResourceEvent
}

// WatchObserver receives watch events
type WatchObserver interface {
	OnResourceAdded(resourceType string, resource interface{})
	OnResourceModified(resourceType string, resource interface{})
	OnResourceDeleted(resourceType string, resource interface{})
	OnWatchError(resourceType string, err error)
}

// NewWatchService creates a new watch service
func NewWatchService(k8s *KubernetesService) *WatchService {
	return &WatchService{
		k8s:       k8s,
		watchers:  make(map[string]*resourceWatcher),
		observers: make([]WatchObserver, 0),
	}
}

// AddObserver adds a watch observer
func (w *WatchService) AddObserver(observer WatchObserver) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.observers = append(w.observers, observer)
}

// RemoveObserver removes a watch observer
func (w *WatchService) RemoveObserver(observer WatchObserver) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, obs := range w.observers {
		if obs == observer {
			w.observers = append(w.observers[:i], w.observers[i+1:]...)
			break
		}
	}
}

// StartWatching starts watching a resource type
func (w *WatchService) StartWatching(resourceType string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Stop existing watcher if any
	if existing, ok := w.watchers[resourceType]; ok {
		existing.cancel()
		delete(w.watchers, resourceType)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(context.Background())

	// Start watching
	eventChan, err := w.k8s.WatchResources(ctx, resourceType)
	if err != nil {
		cancel()
		return err
	}

	// Create watcher record
	watcher := &resourceWatcher{
		resourceType: resourceType,
		context:      ctx,
		cancel:       cancel,
		eventChan:    eventChan,
	}

	w.watchers[resourceType] = watcher

	// Start processing events
	go w.processEvents(watcher)

	return nil
}

// StopWatching stops watching a resource type
func (w *WatchService) StopWatching(resourceType string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if watcher, ok := w.watchers[resourceType]; ok {
		watcher.cancel()
		delete(w.watchers, resourceType)
	}
}

// StopAllWatchers stops all watchers
func (w *WatchService) StopAllWatchers() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, watcher := range w.watchers {
		watcher.cancel()
	}

	w.watchers = make(map[string]*resourceWatcher)
}

// IsWatching checks if a resource type is being watched
func (w *WatchService) IsWatching(resourceType string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	_, exists := w.watchers[resourceType]
	return exists
}

// GetActiveWatchers returns the list of active watchers
func (w *WatchService) GetActiveWatchers() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	types := make([]string, 0, len(w.watchers))
	for resourceType := range w.watchers {
		types = append(types, resourceType)
	}

	return types
}

// processEvents processes events from a watcher
func (w *WatchService) processEvents(watcher *resourceWatcher) {
	for {
		select {
		case event, ok := <-watcher.eventChan:
			if !ok {
				// Channel closed, watcher stopped
				return
			}

			// Process event
			w.handleEvent(watcher.resourceType, event)

		case <-watcher.context.Done():
			// Context cancelled, stop processing
			return
		}
	}
}

// handleEvent handles a resource event
func (w *WatchService) handleEvent(resourceType string, event ResourceEvent) {
	w.mu.RLock()
	observers := make([]WatchObserver, len(w.observers))
	copy(observers, w.observers)
	w.mu.RUnlock()

	// Notify observers based on event type
	for _, observer := range observers {
		switch event.Type {
		case EventAdded:
			observer.OnResourceAdded(resourceType, event.Resource)
		case EventModified:
			observer.OnResourceModified(resourceType, event.Resource)
		case EventDeleted:
			observer.OnResourceDeleted(resourceType, event.Resource)
		}
	}
}

// RefreshResource manually refreshes a resource
func (w *WatchService) RefreshResource(resourceType string) error {
	// This would trigger a manual refresh of the resource
	// Implementation depends on the specific resource type

	// For now, simulate a refresh by creating synthetic events
	switch resourceType {
	case "pods":
		pods, err := w.k8s.GetPods()
		if err != nil {
			w.notifyError(resourceType, err)
			return err
		}

		// Create synthetic added events for all pods
		for _, pod := range pods {
			event := ResourceEvent{
				Type:      EventModified,
				Resource:  pod,
				Timestamp: time.Now(),
			}
			w.handleEvent(resourceType, event)
		}

	case "services":
		services, err := w.k8s.GetServices()
		if err != nil {
			w.notifyError(resourceType, err)
			return err
		}

		for _, svc := range services {
			event := ResourceEvent{
				Type:      EventModified,
				Resource:  svc,
				Timestamp: time.Now(),
			}
			w.handleEvent(resourceType, event)
		}

	case "deployments":
		deployments, err := w.k8s.GetDeployments()
		if err != nil {
			w.notifyError(resourceType, err)
			return err
		}

		for _, dep := range deployments {
			event := ResourceEvent{
				Type:      EventModified,
				Resource:  dep,
				Timestamp: time.Now(),
			}
			w.handleEvent(resourceType, event)
		}

	default:
		// Unsupported resource type
		return nil
	}

	return nil
}

// notifyError notifies observers of a watch error
func (w *WatchService) notifyError(resourceType string, err error) {
	w.mu.RLock()
	observers := make([]WatchObserver, len(w.observers))
	copy(observers, w.observers)
	w.mu.RUnlock()

	for _, observer := range observers {
		observer.OnWatchError(resourceType, err)
	}
}

// WatchConfig configures watch behavior
type WatchConfig struct {
	// ResourceTypes to watch
	ResourceTypes []string

	// RefreshInterval for manual refresh (if watch not supported)
	RefreshInterval time.Duration

	// BufferSize for event channels
	BufferSize int
}

// DefaultWatchConfig returns default watch configuration
func DefaultWatchConfig() *WatchConfig {
	return &WatchConfig{
		ResourceTypes:   []string{"pods", "services", "deployments"},
		RefreshInterval: 5 * time.Second,
		BufferSize:      100,
	}
}

// ConfigureWatch configures the watch service
func (w *WatchService) ConfigureWatch(config *WatchConfig) error {
	// Stop all existing watchers
	w.StopAllWatchers()

	// Start watching configured resource types
	for _, resourceType := range config.ResourceTypes {
		if err := w.StartWatching(resourceType); err != nil {
			// Log error but continue with other resources
			continue
		}
	}

	return nil
}
