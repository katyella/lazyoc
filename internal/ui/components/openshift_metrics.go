package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// OpenShiftMetrics provides visualization for OpenShift resource relationships and metrics
type OpenShiftMetrics struct {
	width  int
	height int
}

// NewOpenShiftMetrics creates a new OpenShift metrics component
func NewOpenShiftMetrics() *OpenShiftMetrics {
	return &OpenShiftMetrics{}
}

// SetSize sets the component size
func (m *OpenShiftMetrics) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// RenderBuildPipeline renders a build pipeline visualization
func (m *OpenShiftMetrics) RenderBuildPipeline(buildConfigs []resources.BuildConfigInfo, builds []resources.BuildInfo, imageStreams []resources.ImageStreamInfo) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		PaddingBottom(1)

	b.WriteString(titleStyle.Render("Build Pipeline Overview"))
	b.WriteString("\n\n")

	// Build statistics
	totalBuilds := len(builds)
	successfulBuilds := 0
	failedBuilds := 0
	runningBuilds := 0

	for _, build := range builds {
		switch build.Phase {
		case "Complete":
			successfulBuilds++
		case "Failed", "Error", "Cancelled":
			failedBuilds++
		case "Running", "Pending":
			runningBuilds++
		}
	}

	// Create metrics box
	metricsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("3")).
		Padding(1).
		Width(50)

	metrics := fmt.Sprintf(`Build Statistics:
  Total Builds: %d
  Successful: %d (%.1f%%)
  Failed: %d (%.1f%%)
  Running: %d

BuildConfigs: %d
ImageStreams: %d`,
		totalBuilds,
		successfulBuilds,
		func() float64 {
			if totalBuilds > 0 {
				return float64(successfulBuilds) / float64(totalBuilds) * 100
			}
			return 0
		}(),
		failedBuilds,
		func() float64 {
			if totalBuilds > 0 {
				return float64(failedBuilds) / float64(totalBuilds) * 100
			}
			return 0
		}(),
		runningBuilds,
		len(buildConfigs),
		len(imageStreams),
	)

	b.WriteString(metricsStyle.Render(metrics))
	b.WriteString("\n\n")

	// Build pipeline flow
	if len(buildConfigs) > 0 {
		pipelineStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")).
			Padding(1)

		pipeline := m.renderPipelineFlow(buildConfigs, imageStreams)
		b.WriteString(pipelineStyle.Render(pipeline))
	}

	return b.String()
}

// RenderResourceTopology renders resource relationship topology
func (m *OpenShiftMetrics) RenderResourceTopology(buildConfigs []resources.BuildConfigInfo, deploymentConfigs []resources.DeploymentConfigInfo, routes []resources.RouteInfo, imageStreams []resources.ImageStreamInfo) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		PaddingBottom(1)

	b.WriteString(titleStyle.Render("Resource Topology"))
	b.WriteString("\n\n")

	// Create topology map
	topology := m.buildTopologyMap(buildConfigs, deploymentConfigs, routes, imageStreams)
	
	topologyStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1).
		Width(80)

	b.WriteString(topologyStyle.Render(topology))

	return b.String()
}

// RenderHealthDashboard renders an OpenShift health dashboard
func (m *OpenShiftMetrics) RenderHealthDashboard(buildConfigs []resources.BuildConfigInfo, builds []resources.BuildInfo, routes []resources.RouteInfo) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		PaddingBottom(1)

	b.WriteString(titleStyle.Render("OpenShift Health Dashboard"))
	b.WriteString("\n\n")

	// Build health
	buildHealth := m.calculateBuildHealth(builds)
	routeHealth := m.calculateRouteHealth(routes)

	healthStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1)

	health := fmt.Sprintf(`System Health Overview:

Build System Health: %s
  Recent Build Success Rate: %.1f%%
  Active Builds: %d
  Build Failures (last 24h): %d

Route Health: %s
  Total Routes: %d
  Routes with TLS: %d
  Secure Routes: %.1f%%

Overall Status: %s`,
		buildHealth.Status,
		buildHealth.SuccessRate,
		buildHealth.ActiveBuilds,
		buildHealth.RecentFailures,
		routeHealth.Status,
		routeHealth.TotalRoutes,
		routeHealth.SecureRoutes,
		routeHealth.SecurePercentage,
		m.getOverallStatus(buildHealth, routeHealth),
	)

	b.WriteString(healthStyle.Render(health))

	return b.String()
}

// Helper methods

func (m *OpenShiftMetrics) renderPipelineFlow(buildConfigs []resources.BuildConfigInfo, imageStreams []resources.ImageStreamInfo) string {
	var b strings.Builder

	b.WriteString("Build Pipeline Flow:\n\n")

	for i, bc := range buildConfigs {
		if i >= 3 { // Limit to first 3 for display
			b.WriteString("... and more\n")
			break
		}

		// Source -> Build -> Image
		sourceInfo := "Source"
		if bc.Source.Git != nil {
			sourceInfo = "Git"
		}

		output := "Image"
		if bc.Output.To != nil {
			output = bc.Output.To.Name
		}

		flow := fmt.Sprintf("%s[%s] -> [Build:%s] -> [%s] -> [Deploy]",
			strings.Repeat("  ", i),
			sourceInfo,
			bc.Strategy,
			output,
		)

		// Add status indicator
		statusColor := "2" // green
		if bc.FailedBuilds > bc.SuccessBuilds {
			statusColor = "1" // red
		} else if bc.SuccessBuilds == 0 {
			statusColor = "3" // yellow
		}

		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
		b.WriteString(statusStyle.Render(flow))
		b.WriteString("\n")
	}

	return b.String()
}

func (m *OpenShiftMetrics) buildTopologyMap(buildConfigs []resources.BuildConfigInfo, deploymentConfigs []resources.DeploymentConfigInfo, routes []resources.RouteInfo, imageStreams []resources.ImageStreamInfo) string {
	var b strings.Builder

	b.WriteString("Resource Relationships:\n\n")

	// Group resources by application (using labels or naming conventions)
	apps := make(map[string]*AppResources)
	
	// Simple grouping by name prefix (before first -)
	for _, bc := range buildConfigs {
		appName := getAppName(bc.Name)
		if apps[appName] == nil {
			apps[appName] = &AppResources{}
		}
		apps[appName].BuildConfigs = append(apps[appName].BuildConfigs, bc.Name)
	}

	for _, dc := range deploymentConfigs {
		appName := getAppName(dc.Name)
		if apps[appName] == nil {
			apps[appName] = &AppResources{}
		}
		apps[appName].DeploymentConfigs = append(apps[appName].DeploymentConfigs, dc.Name)
	}

	for _, route := range routes {
		appName := getAppName(route.Name)
		if apps[appName] == nil {
			apps[appName] = &AppResources{}
		}
		apps[appName].Routes = append(apps[appName].Routes, route.Name)
	}

	for _, is := range imageStreams {
		appName := getAppName(is.Name)
		if apps[appName] == nil {
			apps[appName] = &AppResources{}
		}
		apps[appName].ImageStreams = append(apps[appName].ImageStreams, is.Name)
	}

	// Render topology
	count := 0
	for appName, resources := range apps {
		if count >= 3 { // Limit display
			b.WriteString("... and more applications\n")
			break
		}

		appStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
		b.WriteString(appStyle.Render(fmt.Sprintf("Application: %s", appName)))
		b.WriteString("\n")

		if len(resources.BuildConfigs) > 0 {
			b.WriteString(fmt.Sprintf("  └─ BuildConfigs: %s\n", strings.Join(resources.BuildConfigs, ", ")))
		}
		if len(resources.ImageStreams) > 0 {
			b.WriteString(fmt.Sprintf("  └─ ImageStreams: %s\n", strings.Join(resources.ImageStreams, ", ")))
		}
		if len(resources.DeploymentConfigs) > 0 {
			b.WriteString(fmt.Sprintf("  └─ DeploymentConfigs: %s\n", strings.Join(resources.DeploymentConfigs, ", ")))
		}
		if len(resources.Routes) > 0 {
			b.WriteString(fmt.Sprintf("  └─ Routes: %s\n", strings.Join(resources.Routes, ", ")))
		}
		b.WriteString("\n")
		count++
	}

	return b.String()
}

func (m *OpenShiftMetrics) calculateBuildHealth(builds []resources.BuildInfo) BuildHealth {
	if len(builds) == 0 {
		return BuildHealth{Status: "Unknown", SuccessRate: 0}
	}

	successful := 0
	failed := 0
	active := 0

	for _, build := range builds {
		switch build.Phase {
		case "Complete":
			successful++
		case "Failed", "Error", "Cancelled":
			failed++
		case "Running", "Pending":
			active++
		}
	}

	total := successful + failed
	successRate := 0.0
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	status := "Healthy"
	if successRate < 70 {
		status = "Degraded"
	}
	if successRate < 50 {
		status = "Unhealthy"
	}

	return BuildHealth{
		Status:         status,
		SuccessRate:    successRate,
		ActiveBuilds:   active,
		RecentFailures: failed,
	}
}

func (m *OpenShiftMetrics) calculateRouteHealth(routes []resources.RouteInfo) RouteHealth {
	total := len(routes)
	secure := 0

	for _, route := range routes {
		if route.TLS != nil {
			secure++
		}
	}

	securePercentage := 0.0
	if total > 0 {
		securePercentage = float64(secure) / float64(total) * 100
	}

	status := "Healthy"
	if securePercentage < 50 {
		status = "Warning"
	}

	return RouteHealth{
		Status:           status,
		TotalRoutes:      total,
		SecureRoutes:     secure,
		SecurePercentage: securePercentage,
	}
}

func (m *OpenShiftMetrics) getOverallStatus(buildHealth BuildHealth, routeHealth RouteHealth) string {
	if buildHealth.Status == "Unhealthy" || routeHealth.Status == "Warning" {
		return "⚠️  Needs Attention"
	}
	if buildHealth.Status == "Degraded" {
		return "⚡ Degraded Performance"
	}
	return "✅ Healthy"
}

func getAppName(resourceName string) string {
	parts := strings.Split(resourceName, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return resourceName
}

// Supporting types

type AppResources struct {
	BuildConfigs      []string
	DeploymentConfigs []string
	Routes            []string
	ImageStreams      []string
}

type BuildHealth struct {
	Status         string
	SuccessRate    float64
	ActiveBuilds   int
	RecentFailures int
}

type RouteHealth struct {
	Status           string
	TotalRoutes      int
	SecureRoutes     int
	SecurePercentage float64
}