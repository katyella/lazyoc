package commands

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/katyella/lazyoc/internal/constants"
	"github.com/katyella/lazyoc/internal/k8s/projects"
	"github.com/katyella/lazyoc/internal/k8s/resources"
	"github.com/katyella/lazyoc/internal/ui/messages"
)

// CommandExecutor handles command execution for the TUI
type CommandExecutor struct {
	resourceClient resources.ResourceClient
	projectManager projects.ProjectManager
	kubeconfigPath string
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(resourceClient resources.ResourceClient, projectManager projects.ProjectManager, kubeconfigPath string) *CommandExecutor {
	return &CommandExecutor{
		resourceClient: resourceClient,
		projectManager: projectManager,
		kubeconfigPath: kubeconfigPath,
	}
}

// ExecuteCommand processes various command types and returns appropriate tea commands
func (e *CommandExecutor) ExecuteCommand(cmd ExecutableCommand) tea.Cmd {
	switch c := cmd.(type) {
	case *LoadPodsCommand:
		return e.executeLoadPods(c)
	case *LoadPodLogsCommand:
		return e.executeLoadPodLogs(c)
	case *LoadServicesCommand:
		return e.executeLoadServices(c)
	case *LoadDeploymentsCommand:
		return e.executeLoadDeployments(c)
	case *LoadConfigMapsCommand:
		return e.executeLoadConfigMaps(c)
	case *LoadSecretsCommand:
		return e.executeLoadSecrets(c)
	case *LoadProjectsCommand:
		return e.executeLoadProjects(c)
	case *SwitchProjectCommand:
		return e.executeSwitchProject(c)
	case *TestConnectionCommand:
		return e.executeTestConnection(c)
	case *RefreshResourcesCommand:
		return e.executeRefreshResources(c)
	default:
		return func() tea.Msg {
			return CommandErrorMsg{
				Err:     fmt.Errorf("unknown command type: %T", cmd),
				Command: cmd,
			}
		}
	}
}

// executeLoadPods loads pods from the current namespace
func (e *CommandExecutor) executeLoadPods(cmd *LoadPodsCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return messages.LoadPodsError{Err: fmt.Errorf("resource client not available")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: cmd.Namespace,
		}

		podList, err := e.resourceClient.ListPods(ctx, opts)
		if err != nil {
			return messages.LoadPodsError{Err: err}
		}

		return messages.PodsLoaded{Pods: podList.Items}
	}
}

// executeLoadPodLogs loads logs from a specific pod
func (e *CommandExecutor) executeLoadPodLogs(cmd *LoadPodLogsCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return PodLogsErrorMsg{
				Err:     fmt.Errorf("resource client not available"),
				PodName: cmd.PodName,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		// Use first container if no container specified
		containerName := cmd.ContainerName
		if containerName == "" && len(cmd.ContainerInfo) > 0 {
			containerName = cmd.ContainerInfo[0].Name
		}

		// Set up log options
		tailLines := int64(constants.DefaultPodLogTailLines)
		logOpts := resources.LogOptions{
			TailLines:  &tailLines,
			Timestamps: true,
		}

		// Fetch logs
		logsStr, err := e.resourceClient.GetPodLogs(ctx, cmd.Namespace, cmd.PodName, containerName, logOpts)
		if err != nil {
			return PodLogsErrorMsg{Err: err, PodName: cmd.PodName}
		}

		// Split logs into lines
		logLines := []string{}
		if logsStr != "" {
			lines := strings.Split(strings.TrimSpace(logsStr), "\n")
			for _, line := range lines {
				if line != "" {
					logLines = append(logLines, line)
				}
			}
		}

		if len(logLines) == 0 {
			logLines = []string{constants.NoLogsAvailableMessage}
		}

		return PodLogsLoadedMsg{Logs: logLines, PodName: cmd.PodName}
	}
}

// executeLoadServices loads services from the current namespace
func (e *CommandExecutor) executeLoadServices(cmd *LoadServicesCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return ResourcesLoadErrorMsg{
				Err:          fmt.Errorf("resource client not available"),
				ResourceType: "services",
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: cmd.Namespace,
		}

		serviceList, err := e.resourceClient.ListServices(ctx, opts)
		if err != nil {
			return ResourcesLoadErrorMsg{Err: err, ResourceType: "services"}
		}

		return ServicesLoadedMsg{Services: serviceList.Items}
	}
}

// executeLoadDeployments loads deployments from the current namespace
func (e *CommandExecutor) executeLoadDeployments(cmd *LoadDeploymentsCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return ResourcesLoadErrorMsg{
				Err:          fmt.Errorf("resource client not available"),
				ResourceType: "deployments",
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: cmd.Namespace,
		}

		deploymentList, err := e.resourceClient.ListDeployments(ctx, opts)
		if err != nil {
			return ResourcesLoadErrorMsg{Err: err, ResourceType: "deployments"}
		}

		return DeploymentsLoadedMsg{Deployments: deploymentList.Items}
	}
}

// executeLoadConfigMaps loads configmaps from the current namespace
func (e *CommandExecutor) executeLoadConfigMaps(cmd *LoadConfigMapsCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return ResourcesLoadErrorMsg{
				Err:          fmt.Errorf("resource client not available"),
				ResourceType: "configmaps",
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: cmd.Namespace,
		}

		configMapList, err := e.resourceClient.ListConfigMaps(ctx, opts)
		if err != nil {
			return ResourcesLoadErrorMsg{Err: err, ResourceType: "configmaps"}
		}

		return ConfigMapsLoadedMsg{ConfigMaps: configMapList.Items}
	}
}

// executeLoadSecrets loads secrets from the current namespace
func (e *CommandExecutor) executeLoadSecrets(cmd *LoadSecretsCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return ResourcesLoadErrorMsg{
				Err:          fmt.Errorf("resource client not available"),
				ResourceType: "secrets",
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		opts := resources.ListOptions{
			Namespace: cmd.Namespace,
		}

		secretList, err := e.resourceClient.ListSecrets(ctx, opts)
		if err != nil {
			return ResourcesLoadErrorMsg{Err: err, ResourceType: "secrets"}
		}

		return SecretsLoadedMsg{Secrets: secretList.Items}
	}
}

// executeLoadProjects loads available projects/namespaces
func (e *CommandExecutor) executeLoadProjects(cmd *LoadProjectsCommand) tea.Cmd {
	return func() tea.Msg {
		if e.projectManager == nil {
			return ProjectErrorMsg{Error: "Project manager not available"}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultOperationTimeout)
		defer cancel()

		projectList, err := e.projectManager.List(ctx, projects.ListOptions{
			IncludeQuotas: false,
			IncludeLimits: false,
		})
		if err != nil {
			return ProjectErrorMsg{Error: fmt.Sprintf("Failed to load projects: %v", err)}
		}

		return ProjectListLoadedMsg{Projects: projectList}
	}
}

// executeSwitchProject switches to the specified project
func (e *CommandExecutor) executeSwitchProject(cmd *SwitchProjectCommand) tea.Cmd {
	return func() tea.Msg {
		if e.projectManager == nil {
			return ProjectErrorMsg{Error: "Project manager not available"}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.ClusterDetectionTimeout)
		defer cancel()

		result, err := e.projectManager.SwitchTo(ctx, cmd.ProjectName)
		if err != nil {
			return ProjectErrorMsg{Error: fmt.Sprintf("Failed to switch to project '%s': %v", cmd.ProjectName, err)}
		}

		if !result.Success {
			return ProjectErrorMsg{Error: result.Message}
		}

		// Return success with the project info
		if result.ProjectInfo != nil {
			return ProjectSwitchedMsg{Project: *result.ProjectInfo}
		}

		// Fallback - create basic project info
		return ProjectSwitchedMsg{
			Project: projects.ProjectInfo{
				Name: cmd.ProjectName,
				Type: projects.ProjectTypeKubernetesNamespace,
			},
		}
	}
}

// executeTestConnection tests the connection to the cluster
func (e *CommandExecutor) executeTestConnection(cmd *TestConnectionCommand) tea.Cmd {
	return func() tea.Msg {
		if e.resourceClient == nil {
			return messages.ConnectionError{Err: fmt.Errorf("resource client not available")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), constants.ConnectionTestTimeout)
		defer cancel()

		err := e.resourceClient.TestConnection(ctx)
		if err != nil {
			return messages.ConnectionError{Err: fmt.Errorf("connection test failed: %w", err)}
		}

		return ConnectionTestSuccessMsg{}
	}
}

// executeRefreshResources refreshes all resources for the current namespace
func (e *CommandExecutor) executeRefreshResources(cmd *RefreshResourcesCommand) tea.Cmd {
	return tea.Batch(
		e.executeLoadPods(&LoadPodsCommand{Namespace: cmd.Namespace}),
		e.executeLoadServices(&LoadServicesCommand{Namespace: cmd.Namespace}),
		e.executeLoadDeployments(&LoadDeploymentsCommand{Namespace: cmd.Namespace}),
		e.executeLoadConfigMaps(&LoadConfigMapsCommand{Namespace: cmd.Namespace}),
		e.executeLoadSecrets(&LoadSecretsCommand{Namespace: cmd.Namespace}),
	)
}

// SetResourceClient updates the resource client
func (e *CommandExecutor) SetResourceClient(client resources.ResourceClient) {
	e.resourceClient = client
}

// SetProjectManager updates the project manager
func (e *CommandExecutor) SetProjectManager(manager projects.ProjectManager) {
	e.projectManager = manager
}

// Message types for command results
type CommandErrorMsg struct {
	Err     error
	Command ExecutableCommand
}

type PodLogsLoadedMsg struct {
	Logs    []string
	PodName string
}

type PodLogsErrorMsg struct {
	Err     error
	PodName string
}

type ServicesLoadedMsg struct {
	Services []resources.ServiceInfo
}

type DeploymentsLoadedMsg struct {
	Deployments []resources.DeploymentInfo
}

type ConfigMapsLoadedMsg struct {
	ConfigMaps []resources.ConfigMapInfo
}

type SecretsLoadedMsg struct {
	Secrets []resources.SecretInfo
}

type ResourcesLoadErrorMsg struct {
	Err          error
	ResourceType string
}

type ProjectListLoadedMsg struct {
	Projects []projects.ProjectInfo
}

type ProjectSwitchedMsg struct {
	Project projects.ProjectInfo
}

type ProjectErrorMsg struct {
	Error string
}

type ConnectionTestSuccessMsg struct{}