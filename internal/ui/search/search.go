package search

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/katyella/lazyoc/internal/k8s/resources"
)

// SearchType represents the type of search
type SearchType int

const (
	SearchTypeContains SearchType = iota
	SearchTypeExact
	SearchTypeRegex
)

// ResourceFilter provides search and filtering capabilities for resources
type ResourceFilter struct {
	query      string
	searchType SearchType
	regex      *regexp.Regexp
}

// NewResourceFilter creates a new resource filter
func NewResourceFilter() *ResourceFilter {
	return &ResourceFilter{
		searchType: SearchTypeContains,
	}
}

// SetQuery sets the search query
func (f *ResourceFilter) SetQuery(query string) error {
	f.query = strings.ToLower(query)

	// If query starts with /, treat as regex
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") && len(query) > 2 {
		pattern := query[1 : len(query)-1]
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		f.regex = regex
		f.searchType = SearchTypeRegex
	} else if strings.HasPrefix(query, "=") {
		f.query = strings.ToLower(query[1:])
		f.searchType = SearchTypeExact
	} else {
		f.searchType = SearchTypeContains
	}

	return nil
}

// GetQuery returns the current search query
func (f *ResourceFilter) GetQuery() string {
	return f.query
}

// IsActive returns true if a search is active
func (f *ResourceFilter) IsActive() bool {
	return f.query != ""
}

// Clear clears the search query
func (f *ResourceFilter) Clear() {
	f.query = ""
	f.regex = nil
	f.searchType = SearchTypeContains
}

// FilterPods filters a list of pods based on the search query
func (f *ResourceFilter) FilterPods(pods []resources.PodInfo) []resources.PodInfo {
	if !f.IsActive() {
		return pods
	}

	var filtered []resources.PodInfo
	for _, pod := range pods {
		if f.matchesPod(pod) {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}

// FilterServices filters a list of services based on the search query
func (f *ResourceFilter) FilterServices(services []resources.ServiceInfo) []resources.ServiceInfo {
	if !f.IsActive() {
		return services
	}

	var filtered []resources.ServiceInfo
	for _, svc := range services {
		if f.matchesService(svc) {
			filtered = append(filtered, svc)
		}
	}
	return filtered
}

// FilterDeployments filters a list of deployments based on the search query
func (f *ResourceFilter) FilterDeployments(deployments []resources.DeploymentInfo) []resources.DeploymentInfo {
	if !f.IsActive() {
		return deployments
	}

	var filtered []resources.DeploymentInfo
	for _, dep := range deployments {
		if f.matchesDeployment(dep) {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

// FilterConfigMaps filters a list of config maps based on the search query
func (f *ResourceFilter) FilterConfigMaps(cms []resources.ConfigMapInfo) []resources.ConfigMapInfo {
	if !f.IsActive() {
		return cms
	}

	var filtered []resources.ConfigMapInfo
	for _, cm := range cms {
		if f.matchesConfigMap(cm) {
			filtered = append(filtered, cm)
		}
	}
	return filtered
}

// FilterSecrets filters a list of secrets based on the search query
func (f *ResourceFilter) FilterSecrets(secrets []resources.SecretInfo) []resources.SecretInfo {
	if !f.IsActive() {
		return secrets
	}

	var filtered []resources.SecretInfo
	for _, secret := range secrets {
		if f.matchesSecret(secret) {
			filtered = append(filtered, secret)
		}
	}
	return filtered
}

// matchesPod checks if a pod matches the search query
func (f *ResourceFilter) matchesPod(pod resources.PodInfo) bool {
	searchText := strings.ToLower(pod.Name) + " " +
		strings.ToLower(pod.Namespace) + " " +
		strings.ToLower(pod.Status) + " " +
		strings.ToLower(pod.Node)

	// Add container names
	for _, container := range pod.ContainerInfo {
		searchText += " " + strings.ToLower(container.Name)
	}

	// Add labels
	for k, v := range pod.Labels {
		searchText += " " + strings.ToLower(k) + "=" + strings.ToLower(v)
	}

	return f.matches(searchText, pod.Name)
}

// matchesService checks if a service matches the search query
func (f *ResourceFilter) matchesService(svc resources.ServiceInfo) bool {
	searchText := strings.ToLower(svc.Name) + " " +
		strings.ToLower(svc.Namespace) + " " +
		strings.ToLower(svc.Type) + " " +
		svc.ClusterIP

	// Add ports
	for _, port := range svc.Ports {
		searchText += " " + port
	}

	// Add selector string
	searchText += " " + strings.ToLower(svc.Selector)

	// Add labels
	for k, v := range svc.Labels {
		searchText += " " + strings.ToLower(k) + "=" + strings.ToLower(v)
	}

	return f.matches(searchText, svc.Name)
}

// matchesDeployment checks if a deployment matches the search query
func (f *ResourceFilter) matchesDeployment(dep resources.DeploymentInfo) bool {
	searchText := strings.ToLower(dep.Name) + " " +
		strings.ToLower(dep.Namespace)

	// Add deployment status info
	searchText += " " + strings.ToLower(dep.Strategy) + " " + strings.ToLower(dep.Condition)

	// Add labels
	for k, v := range dep.Labels {
		searchText += " " + strings.ToLower(k) + "=" + strings.ToLower(v)
	}

	return f.matches(searchText, dep.Name)
}

// matchesConfigMap checks if a config map matches the search query
func (f *ResourceFilter) matchesConfigMap(cm resources.ConfigMapInfo) bool {
	searchText := strings.ToLower(cm.Name) + " " +
		strings.ToLower(cm.Namespace)

	// Add data count info
	searchText += fmt.Sprintf(" %d", cm.DataCount)

	return f.matches(searchText, cm.Name)
}

// matchesSecret checks if a secret matches the search query
func (f *ResourceFilter) matchesSecret(secret resources.SecretInfo) bool {
	searchText := strings.ToLower(secret.Name) + " " +
		strings.ToLower(secret.Namespace) + " " +
		strings.ToLower(secret.Type)

	// Add data count info (don't expose actual data)
	searchText += fmt.Sprintf(" %d", secret.DataCount)

	return f.matches(searchText, secret.Name)
}

// matches performs the actual matching based on search type
func (f *ResourceFilter) matches(searchText, originalName string) bool {
	switch f.searchType {
	case SearchTypeContains:
		return strings.Contains(searchText, f.query)
	case SearchTypeExact:
		return strings.ToLower(originalName) == f.query
	case SearchTypeRegex:
		if f.regex != nil {
			return f.regex.MatchString(originalName)
		}
	}
	return false
}
