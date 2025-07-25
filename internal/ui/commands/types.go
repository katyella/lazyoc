package commands

import (
	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// ExecutableCommand represents a command that can be executed by the CommandExecutor
type ExecutableCommand interface {
	// GetType returns the command type for identification
	GetType() ExecutableCommandType

	// GetDescription returns a human-readable description of the command
	GetDescription() string
}

// ExecutableCommandType represents different types of commands
type ExecutableCommandType int

const (
	ExecutableCommandTypeLoadPods ExecutableCommandType = iota
	ExecutableCommandTypeLoadPodLogs
	ExecutableCommandTypeLoadServices
	ExecutableCommandTypeLoadDeployments
	ExecutableCommandTypeLoadConfigMaps
	ExecutableCommandTypeLoadSecrets
	ExecutableCommandTypeLoadProjects
	ExecutableCommandTypeSwitchProject
	ExecutableCommandTypeTestConnection
	ExecutableCommandTypeRefreshResources
)

// LoadPodsCommand loads pods from a namespace
type LoadPodsCommand struct {
	Namespace string
}

func (c *LoadPodsCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadPods
}

func (c *LoadPodsCommand) GetDescription() string {
	return "Load pods from namespace " + c.Namespace
}

// LoadPodLogsCommand loads logs from a specific pod
type LoadPodLogsCommand struct {
	Namespace     string
	PodName       string
	ContainerName string
	ContainerInfo []resources.ContainerInfo
}

func (c *LoadPodLogsCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadPodLogs
}

func (c *LoadPodLogsCommand) GetDescription() string {
	return "Load logs from pod " + c.PodName
}

// LoadServicesCommand loads services from a namespace
type LoadServicesCommand struct {
	Namespace string
}

func (c *LoadServicesCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadServices
}

func (c *LoadServicesCommand) GetDescription() string {
	return "Load services from namespace " + c.Namespace
}

// LoadDeploymentsCommand loads deployments from a namespace
type LoadDeploymentsCommand struct {
	Namespace string
}

func (c *LoadDeploymentsCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadDeployments
}

func (c *LoadDeploymentsCommand) GetDescription() string {
	return "Load deployments from namespace " + c.Namespace
}

// LoadConfigMapsCommand loads configmaps from a namespace
type LoadConfigMapsCommand struct {
	Namespace string
}

func (c *LoadConfigMapsCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadConfigMaps
}

func (c *LoadConfigMapsCommand) GetDescription() string {
	return "Load configmaps from namespace " + c.Namespace
}

// LoadSecretsCommand loads secrets from a namespace
type LoadSecretsCommand struct {
	Namespace string
}

func (c *LoadSecretsCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadSecrets
}

func (c *LoadSecretsCommand) GetDescription() string {
	return "Load secrets from namespace " + c.Namespace
}

// LoadProjectsCommand loads available projects/namespaces
type LoadProjectsCommand struct{}

func (c *LoadProjectsCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeLoadProjects
}

func (c *LoadProjectsCommand) GetDescription() string {
	return "Load available projects/namespaces"
}

// SwitchProjectCommand switches to a specific project
type SwitchProjectCommand struct {
	ProjectName string
}

func (c *SwitchProjectCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeSwitchProject
}

func (c *SwitchProjectCommand) GetDescription() string {
	return "Switch to project " + c.ProjectName
}

// TestConnectionCommand tests the connection to the cluster
type TestConnectionCommand struct{}

func (c *TestConnectionCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeTestConnection
}

func (c *TestConnectionCommand) GetDescription() string {
	return "Test connection to cluster"
}

// RefreshResourcesCommand refreshes all resources for a namespace
type RefreshResourcesCommand struct {
	Namespace string
}

func (c *RefreshResourcesCommand) GetType() ExecutableCommandType {
	return ExecutableCommandTypeRefreshResources
}

func (c *RefreshResourcesCommand) GetDescription() string {
	return "Refresh all resources for namespace " + c.Namespace
}

// CommandFactory creates commands based on various inputs
type CommandFactory struct{}

// NewCommandFactory creates a new command factory
func NewCommandFactory() *CommandFactory {
	return &CommandFactory{}
}

// CreateLoadPodsCommand creates a command to load pods
func (f *CommandFactory) CreateLoadPodsCommand(namespace string) ExecutableCommand {
	return &LoadPodsCommand{Namespace: namespace}
}

// CreateLoadPodLogsCommand creates a command to load pod logs
func (f *CommandFactory) CreateLoadPodLogsCommand(namespace, podName, containerName string, containers []resources.ContainerInfo) ExecutableCommand {
	return &LoadPodLogsCommand{
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		ContainerInfo: containers,
	}
}

// CreateLoadServicesCommand creates a command to load services
func (f *CommandFactory) CreateLoadServicesCommand(namespace string) ExecutableCommand {
	return &LoadServicesCommand{Namespace: namespace}
}

// CreateLoadDeploymentsCommand creates a command to load deployments
func (f *CommandFactory) CreateLoadDeploymentsCommand(namespace string) ExecutableCommand {
	return &LoadDeploymentsCommand{Namespace: namespace}
}

// CreateLoadConfigMapsCommand creates a command to load configmaps
func (f *CommandFactory) CreateLoadConfigMapsCommand(namespace string) ExecutableCommand {
	return &LoadConfigMapsCommand{Namespace: namespace}
}

// CreateLoadSecretsCommand creates a command to load secrets
func (f *CommandFactory) CreateLoadSecretsCommand(namespace string) ExecutableCommand {
	return &LoadSecretsCommand{Namespace: namespace}
}

// CreateLoadProjectsCommand creates a command to load projects
func (f *CommandFactory) CreateLoadProjectsCommand() ExecutableCommand {
	return &LoadProjectsCommand{}
}

// CreateSwitchProjectCommand creates a command to switch projects
func (f *CommandFactory) CreateSwitchProjectCommand(projectName string) ExecutableCommand {
	return &SwitchProjectCommand{ProjectName: projectName}
}

// CreateTestConnectionCommand creates a command to test connection
func (f *CommandFactory) CreateTestConnectionCommand() ExecutableCommand {
	return &TestConnectionCommand{}
}

// CreateRefreshResourcesCommand creates a command to refresh all resources
func (f *CommandFactory) CreateRefreshResourcesCommand(namespace string) ExecutableCommand {
	return &RefreshResourcesCommand{Namespace: namespace}
}

// GetExecutableCommandTypeName returns a human-readable name for a command type
func GetExecutableCommandTypeName(cmdType ExecutableCommandType) string {
	switch cmdType {
	case ExecutableCommandTypeLoadPods:
		return "LoadPods"
	case ExecutableCommandTypeLoadPodLogs:
		return "LoadPodLogs"
	case ExecutableCommandTypeLoadServices:
		return "LoadServices"
	case ExecutableCommandTypeLoadDeployments:
		return "LoadDeployments"
	case ExecutableCommandTypeLoadConfigMaps:
		return "LoadConfigMaps"
	case ExecutableCommandTypeLoadSecrets:
		return "LoadSecrets"
	case ExecutableCommandTypeLoadProjects:
		return "LoadProjects"
	case ExecutableCommandTypeSwitchProject:
		return "SwitchProject"
	case ExecutableCommandTypeTestConnection:
		return "TestConnection"
	case ExecutableCommandTypeRefreshResources:
		return "RefreshResources"
	default:
		return "Unknown"
	}
}
