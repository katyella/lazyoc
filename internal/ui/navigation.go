package ui

import (
	"github.com/katyella/lazyoc/internal/logging"
	"github.com/katyella/lazyoc/internal/ui/models"
)

// Navigator handles resource navigation logic for both keyboard and mouse events
type Navigator struct {
	tui *TUI
}

// NewNavigator creates a new Navigator instance
func NewNavigator(tui *TUI) *Navigator {
	return &Navigator{tui: tui}
}

// SelectNextResource moves selection to the next resource in the current tab
func (n *Navigator) SelectNextResource() {
	n.moveResourceSelection(1)
}

// SelectPreviousResource moves selection to the previous resource in the current tab
func (n *Navigator) SelectPreviousResource() {
	n.moveResourceSelection(-1)
}

// SelectResource selects a specific resource by index in the current tab
func (n *Navigator) SelectResource(index int) {
	switch n.tui.ActiveTab {
	case models.TabPods:
		logging.Debug(n.tui.Logger, "Navigator: attempting to select pod %d, have %d pods total", index, len(n.tui.pods))
		if index >= 0 && index < len(n.tui.pods) {
			n.tui.selectedPod = index
			n.tui.updatePodDisplay()
			logging.Debug(n.tui.Logger, "Selected pod %d", index)
		} else {
			logging.Debug(n.tui.Logger, "Navigator: pod index %d out of bounds (0-%d)", index, len(n.tui.pods)-1)
		}
	case models.TabServices:
		if index >= 0 && index < len(n.tui.services) {
			n.tui.selectedService = index
			n.tui.updateServiceDisplay()
			logging.Debug(n.tui.Logger, "Selected service %d", index)
		}
	case models.TabDeployments:
		if index >= 0 && index < len(n.tui.deployments) {
			n.tui.selectedDeployment = index
			n.tui.updateDeploymentDisplay()
			logging.Debug(n.tui.Logger, "Selected deployment %d", index)
		}
	case models.TabConfigMaps:
		if index >= 0 && index < len(n.tui.configMaps) {
			n.tui.selectedConfigMap = index
			n.tui.updateConfigMapDisplay()
			logging.Debug(n.tui.Logger, "Selected configmap %d", index)
		}
	case models.TabSecrets:
		if index >= 0 && index < len(n.tui.secrets) {
			n.tui.selectedSecret = index
			n.tui.updateSecretDisplay()
			logging.Debug(n.tui.Logger, "Selected secret %d", index)
		}
	case models.TabBuildConfigs:
		if index >= 0 && index < len(n.tui.buildConfigs) {
			n.tui.selectedBuildConfig = index
			n.tui.updateBuildConfigDisplay()
			logging.Debug(n.tui.Logger, "Selected buildconfig %d", index)
		}
	case models.TabImageStreams:
		if index >= 0 && index < len(n.tui.imageStreams) {
			n.tui.selectedImageStream = index
			n.tui.updateImageStreamDisplay()
			logging.Debug(n.tui.Logger, "Selected imagestream %d", index)
		}
	case models.TabRoutes:
		if index >= 0 && index < len(n.tui.routes) {
			n.tui.selectedRoute = index
			n.tui.updateRouteDisplay()
			logging.Debug(n.tui.Logger, "Selected route %d", index)
		}
	}
}

// moveResourceSelection moves the selection by delta in the current tab
func (n *Navigator) moveResourceSelection(delta int) {
	switch n.tui.ActiveTab {
	case models.TabPods:
		n.movePodSelection(delta)
	case models.TabServices:
		n.moveServiceSelection(delta)
	case models.TabDeployments:
		n.moveDeploymentSelection(delta)
	case models.TabConfigMaps:
		n.moveConfigMapSelection(delta)
	case models.TabSecrets:
		n.moveSecretSelection(delta)
	case models.TabBuildConfigs:
		n.moveBuildConfigSelection(delta)
	case models.TabImageStreams:
		n.moveImageStreamSelection(delta)
	case models.TabRoutes:
		n.moveRouteSelection(delta)
	}
}

// Helper methods for each resource type
func (n *Navigator) movePodSelection(delta int) {
	if len(n.tui.pods) == 0 {
		return
	}
	
	newIndex := n.tui.selectedPod + delta
	if delta > 0 {
		// Moving down/forward
		n.tui.selectedPod = (newIndex) % len(n.tui.pods)
	} else {
		// Moving up/backward
		if newIndex < 0 {
			n.tui.selectedPod = len(n.tui.pods) - 1
		} else {
			n.tui.selectedPod = newIndex
		}
	}
	n.tui.updatePodDisplay()
}

func (n *Navigator) moveServiceSelection(delta int) {
	if len(n.tui.services) == 0 {
		return
	}
	
	newIndex := n.tui.selectedService + delta
	if delta > 0 {
		n.tui.selectedService = (newIndex) % len(n.tui.services)
	} else {
		if newIndex < 0 {
			n.tui.selectedService = len(n.tui.services) - 1
		} else {
			n.tui.selectedService = newIndex
		}
	}
	n.tui.updateServiceDisplay()
}

func (n *Navigator) moveDeploymentSelection(delta int) {
	if len(n.tui.deployments) == 0 {
		return
	}
	
	newIndex := n.tui.selectedDeployment + delta
	if delta > 0 {
		n.tui.selectedDeployment = (newIndex) % len(n.tui.deployments)
	} else {
		if newIndex < 0 {
			n.tui.selectedDeployment = len(n.tui.deployments) - 1
		} else {
			n.tui.selectedDeployment = newIndex
		}
	}
	n.tui.updateDeploymentDisplay()
}

func (n *Navigator) moveConfigMapSelection(delta int) {
	if len(n.tui.configMaps) == 0 {
		return
	}
	
	newIndex := n.tui.selectedConfigMap + delta
	if delta > 0 {
		n.tui.selectedConfigMap = (newIndex) % len(n.tui.configMaps)
	} else {
		if newIndex < 0 {
			n.tui.selectedConfigMap = len(n.tui.configMaps) - 1
		} else {
			n.tui.selectedConfigMap = newIndex
		}
	}
	n.tui.updateConfigMapDisplay()
}

func (n *Navigator) moveSecretSelection(delta int) {
	if len(n.tui.secrets) == 0 {
		return
	}
	
	newIndex := n.tui.selectedSecret + delta
	if delta > 0 {
		n.tui.selectedSecret = (newIndex) % len(n.tui.secrets)
	} else {
		if newIndex < 0 {
			n.tui.selectedSecret = len(n.tui.secrets) - 1
		} else {
			n.tui.selectedSecret = newIndex
		}
	}
	n.tui.updateSecretDisplay()
}

func (n *Navigator) moveBuildConfigSelection(delta int) {
	if len(n.tui.buildConfigs) == 0 {
		return
	}
	
	newIndex := n.tui.selectedBuildConfig + delta
	if delta > 0 {
		n.tui.selectedBuildConfig = (newIndex) % len(n.tui.buildConfigs)
	} else {
		if newIndex < 0 {
			n.tui.selectedBuildConfig = len(n.tui.buildConfigs) - 1
		} else {
			n.tui.selectedBuildConfig = newIndex
		}
	}
	n.tui.updateBuildConfigDisplay()
}

func (n *Navigator) moveImageStreamSelection(delta int) {
	if len(n.tui.imageStreams) == 0 {
		return
	}
	
	newIndex := n.tui.selectedImageStream + delta
	if delta > 0 {
		n.tui.selectedImageStream = (newIndex) % len(n.tui.imageStreams)
	} else {
		if newIndex < 0 {
			n.tui.selectedImageStream = len(n.tui.imageStreams) - 1
		} else {
			n.tui.selectedImageStream = newIndex
		}
	}
	n.tui.updateImageStreamDisplay()
}

func (n *Navigator) moveRouteSelection(delta int) {
	if len(n.tui.routes) == 0 {
		return
	}
	
	newIndex := n.tui.selectedRoute + delta
	if delta > 0 {
		n.tui.selectedRoute = (newIndex) % len(n.tui.routes)
	} else {
		if newIndex < 0 {
			n.tui.selectedRoute = len(n.tui.routes) - 1
		} else {
			n.tui.selectedRoute = newIndex
		}
	}
	n.tui.updateRouteDisplay()
}