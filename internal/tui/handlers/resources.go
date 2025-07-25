package handlers

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ResourceHandler handles resource-related operations
type ResourceHandler struct {
	// Callbacks for resource operations
	onLoad   func(resourceType string) tea.Cmd
	onSelect func(resourceType string, index int) tea.Cmd
	onDelete func(resourceType string, name string) tea.Cmd
	onEdit   func(resourceType string, name string) tea.Cmd
	onExec   func(resourceType string, name string, container string) tea.Cmd
	onLogs   func(resourceType string, name string, container string) tea.Cmd
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{}
}

// SetLoadCallback sets the callback for loading resources
func (r *ResourceHandler) SetLoadCallback(fn func(resourceType string) tea.Cmd) {
	r.onLoad = fn
}

// SetSelectCallback sets the callback for selecting a resource
func (r *ResourceHandler) SetSelectCallback(fn func(resourceType string, index int) tea.Cmd) {
	r.onSelect = fn
}

// SetDeleteCallback sets the callback for deleting a resource
func (r *ResourceHandler) SetDeleteCallback(fn func(resourceType string, name string) tea.Cmd) {
	r.onDelete = fn
}

// SetEditCallback sets the callback for editing a resource
func (r *ResourceHandler) SetEditCallback(fn func(resourceType string, name string) tea.Cmd) {
	r.onEdit = fn
}

// SetExecCallback sets the callback for executing into a container
func (r *ResourceHandler) SetExecCallback(fn func(resourceType string, name string, container string) tea.Cmd) {
	r.onExec = fn
}

// SetLogsCallback sets the callback for viewing logs
func (r *ResourceHandler) SetLogsCallback(fn func(resourceType string, name string, container string) tea.Cmd) {
	r.onLogs = fn
}

// LoadResources triggers loading of resources
func (r *ResourceHandler) LoadResources(resourceType string) tea.Cmd {
	if r.onLoad != nil {
		return r.onLoad(resourceType)
	}
	return nil
}

// SelectResource handles resource selection
func (r *ResourceHandler) SelectResource(resourceType string, index int) tea.Cmd {
	if r.onSelect != nil {
		return r.onSelect(resourceType, index)
	}
	return nil
}

// DeleteResource handles resource deletion
func (r *ResourceHandler) DeleteResource(resourceType string, name string) tea.Cmd {
	if r.onDelete != nil {
		return r.onDelete(resourceType, name)
	}
	return nil
}

// EditResource handles resource editing
func (r *ResourceHandler) EditResource(resourceType string, name string) tea.Cmd {
	if r.onEdit != nil {
		return r.onEdit(resourceType, name)
	}
	return nil
}

// ExecIntoContainer handles executing into a container
func (r *ResourceHandler) ExecIntoContainer(resourceType string, name string, container string) tea.Cmd {
	if r.onExec != nil {
		return r.onExec(resourceType, name, container)
	}
	return nil
}

// ViewLogs handles viewing logs
func (r *ResourceHandler) ViewLogs(resourceType string, name string, container string) tea.Cmd {
	if r.onLogs != nil {
		return r.onLogs(resourceType, name, container)
	}
	return nil
}

// Common resource messages

// ResourceLoadingMsg indicates resources are being loaded
type ResourceLoadingMsg struct {
	ResourceType string
}

// ResourceLoadedMsg indicates resources have been loaded
type ResourceLoadedMsg struct {
	ResourceType string
	Count        int
	Duration     time.Duration
}

// ResourceErrorMsg indicates an error occurred
type ResourceErrorMsg struct {
	ResourceType string
	Operation    string
	Error        error
}

// ResourceSelectedMsg indicates a resource was selected
type ResourceSelectedMsg struct {
	ResourceType string
	Name         string
	Namespace    string
	Index        int
}

// ResourceDeletedMsg indicates a resource was deleted
type ResourceDeletedMsg struct {
	ResourceType string
	Name         string
	Namespace    string
}

// ResourceUpdatedMsg indicates a resource was updated
type ResourceUpdatedMsg struct {
	ResourceType string
	Name         string
	Namespace    string
}

// Helper functions for creating messages

// NewResourceError creates a resource error message
func NewResourceError(resourceType, operation string, err error) ResourceErrorMsg {
	return ResourceErrorMsg{
		ResourceType: resourceType,
		Operation:    operation,
		Error:        err,
	}
}

// NewResourceLoading creates a resource loading message
func NewResourceLoading(resourceType string) ResourceLoadingMsg {
	return ResourceLoadingMsg{
		ResourceType: resourceType,
	}
}

// NewResourceLoaded creates a resource loaded message
func NewResourceLoaded(resourceType string, count int, duration time.Duration) ResourceLoadedMsg {
	return ResourceLoadedMsg{
		ResourceType: resourceType,
		Count:        count,
		Duration:     duration,
	}
}

// FormatResourceError formats a resource error for display
func FormatResourceError(err ResourceErrorMsg) string {
	return fmt.Sprintf("Error %s %s: %v", err.Operation, err.ResourceType, err.Error)
}
