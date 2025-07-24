package services

import (
	"fmt"
	"sync"
)

// ServiceLocator provides dependency injection for services
type ServiceLocator struct {
	mu       sync.RWMutex
	services map[string]interface{}
}

// Global service locator instance
var (
	globalLocator *ServiceLocator
	once          sync.Once
)

// GetServiceLocator returns the global service locator
func GetServiceLocator() *ServiceLocator {
	once.Do(func() {
		globalLocator = NewServiceLocator()
	})
	return globalLocator
}

// NewServiceLocator creates a new service locator
func NewServiceLocator() *ServiceLocator {
	return &ServiceLocator{
		services: make(map[string]interface{}),
	}
}

// Register registers a service
func (s *ServiceLocator) Register(name string, service interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}
	
	s.services[name] = service
	return nil
}

// Unregister removes a service
func (s *ServiceLocator) Unregister(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.services, name)
}

// Get retrieves a service by name
func (s *ServiceLocator) Get(name string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	service, exists := s.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	
	return service, nil
}

// GetKubernetesService retrieves the Kubernetes service
func (s *ServiceLocator) GetKubernetesService() (*KubernetesService, error) {
	service, err := s.Get("kubernetes")
	if err != nil {
		return nil, err
	}
	
	k8s, ok := service.(*KubernetesService)
	if !ok {
		return nil, fmt.Errorf("service 'kubernetes' is not of type *KubernetesService")
	}
	
	return k8s, nil
}

// GetLogsService retrieves the logs service
func (s *ServiceLocator) GetLogsService() (*LogsService, error) {
	service, err := s.Get("logs")
	if err != nil {
		return nil, err
	}
	
	logs, ok := service.(*LogsService)
	if !ok {
		return nil, fmt.Errorf("service 'logs' is not of type *LogsService")
	}
	
	return logs, nil
}

// GetWatchService retrieves the watch service
func (s *ServiceLocator) GetWatchService() (*WatchService, error) {
	service, err := s.Get("watch")
	if err != nil {
		return nil, err
	}
	
	watch, ok := service.(*WatchService)
	if !ok {
		return nil, fmt.Errorf("service 'watch' is not of type *WatchService")
	}
	
	return watch, nil
}

// InitializeServices initializes all services
func (s *ServiceLocator) InitializeServices() error {
	// Create Kubernetes service
	k8s := NewKubernetesService()
	if err := s.Register("kubernetes", k8s); err != nil {
		return err
	}
	
	// Create logs service
	logs := NewLogsService(k8s)
	if err := s.Register("logs", logs); err != nil {
		return err
	}
	
	// Create watch service
	watch := NewWatchService(k8s)
	if err := s.Register("watch", watch); err != nil {
		return err
	}
	
	return nil
}

// Shutdown shuts down all services
func (s *ServiceLocator) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Stop watch service
	if watch, err := s.GetWatchService(); err == nil {
		watch.StopAllWatchers()
	}
	
	// Stop logs service
	if logs, err := s.GetLogsService(); err == nil {
		logs.StopAllStreams()
	}
	
	// Disconnect from Kubernetes
	if k8s, err := s.GetKubernetesService(); err == nil {
		k8s.Disconnect()
	}
	
	// Clear all services
	s.services = make(map[string]interface{})
}

// ServiceConfig contains configuration for services
type ServiceConfig struct {
	KubeconfigPath string
	Context        string
	Namespace      string
	WatchConfig    *WatchConfig
}

// ConfigureServices configures all services with the given config
func (s *ServiceLocator) ConfigureServices(config *ServiceConfig) error {
	// Configure Kubernetes service
	k8s, err := s.GetKubernetesService()
	if err != nil {
		return err
	}
	
	if err := k8s.Connect(config.KubeconfigPath, config.Context); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}
	
	if config.Namespace != "" {
		if err := k8s.SetNamespace(config.Namespace); err != nil {
			return fmt.Errorf("failed to set namespace: %w", err)
		}
	}
	
	// Configure watch service
	if config.WatchConfig != nil {
		watch, err := s.GetWatchService()
		if err != nil {
			return err
		}
		
		if err := watch.ConfigureWatch(config.WatchConfig); err != nil {
			return fmt.Errorf("failed to configure watch: %w", err)
		}
	}
	
	return nil
}

// Helper functions for common service operations

// ConnectToCluster connects to a Kubernetes cluster
func ConnectToCluster(kubeconfig, context string) error {
	locator := GetServiceLocator()
	k8s, err := locator.GetKubernetesService()
	if err != nil {
		return err
	}
	
	return k8s.Connect(kubeconfig, context)
}

// StartWatchingResources starts watching the specified resource types
func StartWatchingResources(resourceTypes []string) error {
	locator := GetServiceLocator()
	watch, err := locator.GetWatchService()
	if err != nil {
		return err
	}
	
	for _, resourceType := range resourceTypes {
		if err := watch.StartWatching(resourceType); err != nil {
			return fmt.Errorf("failed to watch %s: %w", resourceType, err)
		}
	}
	
	return nil
}

// StartLogStreaming starts streaming logs for a pod
func StartLogStreaming(podName, containerName string, lines int64) error {
	locator := GetServiceLocator()
	logs, err := locator.GetLogsService()
	if err != nil {
		return err
	}
	
	return logs.StartStreaming(podName, containerName, lines)
}