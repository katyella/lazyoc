package resources

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	
	"github.com/katyella/lazyoc/internal/k8s"
)

// OpenShiftResourceClient provides operations for OpenShift-specific resources
type OpenShiftResourceClient struct {
	client k8s.OpenShiftClient
}

// NewOpenShiftResourceClient creates a new OpenShift resource client
func NewOpenShiftResourceClient(client k8s.OpenShiftClient) *OpenShiftResourceClient {
	return &OpenShiftResourceClient{
		client: client,
	}
}

// BuildConfigs

// ListBuildConfigs retrieves BuildConfigs from the specified namespace
func (c *OpenShiftResourceClient) ListBuildConfigs(ctx context.Context, opts ListOptions) (*ResourceList[BuildConfigInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	buildClient := c.client.GetBuildClient()
	buildConfigs, err := buildClient.BuildV1().BuildConfigs(opts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list BuildConfigs: %w", err)
	}

	items := make([]BuildConfigInfo, 0, len(buildConfigs.Items))
	for _, bc := range buildConfigs.Items {
		info := buildConfigToInfo(&bc)
		items = append(items, info)
	}

	return &ResourceList[BuildConfigInfo]{
		Items:     items,
		Total:     len(items),
		Namespace: opts.Namespace,
		Continue:  buildConfigs.Continue,
		Remaining: func() int64 { if buildConfigs.RemainingItemCount != nil { return *buildConfigs.RemainingItemCount }; return 0 }(),
	}, nil
}

// GetBuildConfig retrieves a specific BuildConfig
func (c *OpenShiftResourceClient) GetBuildConfig(ctx context.Context, namespace, name string) (*BuildConfigInfo, error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	buildClient := c.client.GetBuildClient()
	bc, err := buildClient.BuildV1().BuildConfigs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get BuildConfig %s: %w", name, err)
	}

	info := buildConfigToInfo(bc)
	return &info, nil
}

// TriggerBuild starts a new build from a BuildConfig
func (c *OpenShiftResourceClient) TriggerBuild(ctx context.Context, namespace, name string) (*BuildInfo, error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	buildClient := c.client.GetBuildClient()
	
	// Create a BuildRequest to trigger the build
	buildRequest := &buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	build, err := buildClient.BuildV1().BuildConfigs(namespace).Instantiate(ctx, name, buildRequest, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to trigger build for BuildConfig %s: %w", name, err)
	}

	info := buildToInfo(build)
	return &info, nil
}

// Builds

// ListBuilds retrieves Builds from the specified namespace
func (c *OpenShiftResourceClient) ListBuilds(ctx context.Context, opts ListOptions) (*ResourceList[BuildInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	buildClient := c.client.GetBuildClient()
	builds, err := buildClient.BuildV1().Builds(opts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Builds: %w", err)
	}

	items := make([]BuildInfo, 0, len(builds.Items))
	for _, build := range builds.Items {
		info := buildToInfo(&build)
		items = append(items, info)
	}

	return &ResourceList[BuildInfo]{
		Items:     items,
		Total:     len(items),
		Namespace: opts.Namespace,
		Continue:  builds.Continue,
		Remaining: func() int64 { if builds.RemainingItemCount != nil { return *builds.RemainingItemCount }; return 0 }(),
	}, nil
}

// ImageStreams

// ListImageStreams retrieves ImageStreams from the specified namespace
func (c *OpenShiftResourceClient) ListImageStreams(ctx context.Context, opts ListOptions) (*ResourceList[ImageStreamInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	imageClient := c.client.GetImageClient()
	imageStreams, err := imageClient.ImageV1().ImageStreams(opts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list ImageStreams: %w", err)
	}

	items := make([]ImageStreamInfo, 0, len(imageStreams.Items))
	for _, is := range imageStreams.Items {
		info := imageStreamToInfo(&is)
		items = append(items, info)
	}

	return &ResourceList[ImageStreamInfo]{
		Items:     items,
		Total:     len(items),
		Namespace: opts.Namespace,
		Continue:  imageStreams.Continue,
		Remaining: func() int64 { if imageStreams.RemainingItemCount != nil { return *imageStreams.RemainingItemCount }; return 0 }(),
	}, nil
}

// DeploymentConfigs

// ListDeploymentConfigs retrieves DeploymentConfigs from the specified namespace
func (c *OpenShiftResourceClient) ListDeploymentConfigs(ctx context.Context, opts ListOptions) (*ResourceList[DeploymentConfigInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	appsClient := c.client.GetAppsClient()
	dcs, err := appsClient.AppsV1().DeploymentConfigs(opts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list DeploymentConfigs: %w", err)
	}

	items := make([]DeploymentConfigInfo, 0, len(dcs.Items))
	for _, dc := range dcs.Items {
		info := deploymentConfigToInfo(&dc)
		items = append(items, info)
	}

	return &ResourceList[DeploymentConfigInfo]{
		Items:     items,
		Total:     len(items),
		Namespace: opts.Namespace,
		Continue:  dcs.Continue,
		Remaining: func() int64 { if dcs.RemainingItemCount != nil { return *dcs.RemainingItemCount }; return 0 }(),
	}, nil
}

// Routes

// ListRoutes retrieves Routes from the specified namespace
func (c *OpenShiftResourceClient) ListRoutes(ctx context.Context, opts ListOptions) (*ResourceList[RouteInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	routeClient := c.client.GetRouteClient()
	routes, err := routeClient.RouteV1().Routes(opts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
		Limit:         opts.Limit,
		Continue:      opts.Continue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Routes: %w", err)
	}

	items := make([]RouteInfo, 0, len(routes.Items))
	for _, route := range routes.Items {
		info := routeToInfo(&route)
		items = append(items, info)
	}

	return &ResourceList[RouteInfo]{
		Items:     items,
		Total:     len(items),
		Namespace: opts.Namespace,
		Continue:  routes.Continue,
		Remaining: func() int64 { if routes.RemainingItemCount != nil { return *routes.RemainingItemCount }; return 0 }(),
	}, nil
}

// Operators - Using dynamic client for flexibility

// ListOperators retrieves ClusterServiceVersions (Operators) from the specified namespace
func (c *OpenShiftResourceClient) ListOperators(ctx context.Context, opts ListOptions) (*ResourceList[OperatorInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	// For now, return empty list - will implement with dynamic client later
	return &ResourceList[OperatorInfo]{
		Items:     []OperatorInfo{},
		Total:     0,
		Namespace: opts.Namespace,
	}, nil
}

// ListSubscriptions retrieves Subscriptions from the specified namespace
func (c *OpenShiftResourceClient) ListSubscriptions(ctx context.Context, opts ListOptions) (*ResourceList[SubscriptionInfo], error) {
	if !c.client.IsOpenShift() {
		return nil, fmt.Errorf("not connected to an OpenShift cluster")
	}

	// For now, return empty list - will implement with dynamic client later
	return &ResourceList[SubscriptionInfo]{
		Items:     []SubscriptionInfo{},
		Total:     0,
		Namespace: opts.Namespace,
	}, nil
}

// Helper conversion functions

func buildConfigToInfo(bc *buildv1.BuildConfig) BuildConfigInfo {
	info := BuildConfigInfo{
		ResourceInfo: ResourceInfo{
			Name:        bc.Name,
			Namespace:   bc.Namespace,
			Kind:        "BuildConfig",
			APIVersion:  bc.APIVersion,
			Labels:      bc.Labels,
			Annotations: bc.Annotations,
			CreatedAt:   bc.CreationTimestamp.Time,
			Status:      "Ready", // BuildConfigs don't have a phase
		},
		Age: duration.HumanDuration(time.Since(bc.CreationTimestamp.Time)),
	}

	// Set strategy
	if bc.Spec.Strategy.Type != "" {
		info.Strategy = string(bc.Spec.Strategy.Type)
	}

	// Set source information
	if bc.Spec.Source.Git != nil {
		info.Source = BuildSource{
			Type: "Git",
			Git: &GitSource{
				URI: bc.Spec.Source.Git.URI,
				Ref: bc.Spec.Source.Git.Ref,
			},
			ContextDir: bc.Spec.Source.ContextDir,
		}
	}

	// Set output information
	if bc.Spec.Output.To != nil {
		info.Output = BuildOutput{
			To: &BuildOutputTo{
				Kind:      bc.Spec.Output.To.Kind,
				Name:      bc.Spec.Output.To.Name,
				Namespace: bc.Spec.Output.To.Namespace,
			},
		}
	}

	// Set build statistics
	info.SuccessBuilds = int(bc.Status.LastVersion)
	
	return info
}

func buildToInfo(build *buildv1.Build) BuildInfo {
	info := BuildInfo{
		ResourceInfo: ResourceInfo{
			Name:        build.Name,
			Namespace:   build.Namespace,
			Kind:        "Build",
			APIVersion:  build.APIVersion,
			Labels:      build.Labels,
			Annotations: build.Annotations,
			CreatedAt:   build.CreationTimestamp.Time,
			Status:      string(build.Status.Phase),
		},
		Phase:       string(build.Status.Phase),
		Message:     build.Status.Message,
		StartTime:   build.Status.StartTimestamp.Time,
		BuildConfig: build.Labels["buildconfig"],
		Age:         duration.HumanDuration(time.Since(build.CreationTimestamp.Time)),
	}

	// Set completion time and duration
	if build.Status.CompletionTimestamp != nil {
		info.CompletionTime = &build.Status.CompletionTimestamp.Time
		info.Duration = duration.HumanDuration(build.Status.CompletionTimestamp.Sub(build.Status.StartTimestamp.Time))
	} else if !build.Status.StartTimestamp.IsZero() {
		info.Duration = duration.HumanDuration(time.Since(build.Status.StartTimestamp.Time))
	}

	// Set strategy
	if build.Spec.Strategy.Type != "" {
		info.Strategy = string(build.Spec.Strategy.Type)
	}

	// Set output image
	if build.Status.OutputDockerImageReference != "" {
		info.OutputImage = build.Status.OutputDockerImageReference
	}

	return info
}

func imageStreamToInfo(is *imagev1.ImageStream) ImageStreamInfo {
	info := ImageStreamInfo{
		ResourceInfo: ResourceInfo{
			Name:        is.Name,
			Namespace:   is.Namespace,
			Kind:        "ImageStream",
			APIVersion:  is.APIVersion,
			Labels:      is.Labels,
			Annotations: is.Annotations,
			CreatedAt:   is.CreationTimestamp.Time,
			Status:      "Ready", // ImageStreams don't have a phase
		},
		DockerImageRepository:       is.Status.DockerImageRepository,
		PublicDockerImageRepository: is.Status.PublicDockerImageRepository,
		Age:                        duration.HumanDuration(time.Since(is.CreationTimestamp.Time)),
	}

	// Convert tags
	for _, tag := range is.Status.Tags {
		tagInfo := ImageStreamTag{
			Name:  tag.Tag,
			Items: make([]ImageStreamImage, 0, len(tag.Items)),
		}
		
		for _, item := range tag.Items {
			imageInfo := ImageStreamImage{
				Created:        item.Created.Time,
				DockerImageRef: item.DockerImageReference,
				Image:          item.Image,
				Generation:     item.Generation,
			}
			tagInfo.Items = append(tagInfo.Items, imageInfo)
		}
		
		info.Tags = append(info.Tags, tagInfo)
	}

	return info
}

func deploymentConfigToInfo(dc *appsv1.DeploymentConfig) DeploymentConfigInfo {
	info := DeploymentConfigInfo{
		ResourceInfo: ResourceInfo{
			Name:        dc.Name,
			Namespace:   dc.Namespace,
			Kind:        "DeploymentConfig",
			APIVersion:  dc.APIVersion,
			Labels:      dc.Labels,
			Annotations: dc.Annotations,
			CreatedAt:   dc.CreationTimestamp.Time,
			Status:      "Ready", // DeploymentConfigs don't have a simple phase
		},
		Replicas:          dc.Spec.Replicas,
		ReadyReplicas:     dc.Status.ReadyReplicas,
		UpdatedReplicas:   dc.Status.UpdatedReplicas,
		AvailableReplicas: dc.Status.AvailableReplicas,
		LatestVersion:     dc.Status.LatestVersion,
		Age:               duration.HumanDuration(time.Since(dc.CreationTimestamp.Time)),
	}

	// Set strategy
	info.Strategy = DeploymentStrategy{
		Type: string(dc.Spec.Strategy.Type),
	}

	return info
}

func routeToInfo(route *routev1.Route) RouteInfo {
	info := RouteInfo{
		ResourceInfo: ResourceInfo{
			Name:        route.Name,
			Namespace:   route.Namespace,
			Kind:        "Route",
			APIVersion:  route.APIVersion,
			Labels:      route.Labels,
			Annotations: route.Annotations,
			CreatedAt:   route.CreationTimestamp.Time,
			Status:      "Ready", // Routes don't have a simple phase
		},
		Host: route.Spec.Host,
		Path: route.Spec.Path,
		Service: RouteTargetRef{
			Kind:   route.Spec.To.Kind,
			Name:   route.Spec.To.Name,
			Weight: route.Spec.To.Weight,
		},
		Age: duration.HumanDuration(time.Since(route.CreationTimestamp.Time)),
	}

	// Set port
	if route.Spec.Port != nil {
		info.Port = &RoutePort{
			TargetPort: route.Spec.Port.TargetPort.String(),
		}
	}

	// Set TLS
	if route.Spec.TLS != nil {
		info.TLS = &TLSConfig{
			Termination:                   string(route.Spec.TLS.Termination),
			Certificate:                   route.Spec.TLS.Certificate,
			Key:                          route.Spec.TLS.Key,
			CACertificate:                route.Spec.TLS.CACertificate,
			DestinationCACertificate:     route.Spec.TLS.DestinationCACertificate,
			InsecureEdgeTerminationPolicy: string(route.Spec.TLS.InsecureEdgeTerminationPolicy),
		}
	}

	return info
}

func clusterServiceVersionToInfo(csv *operatorsv1alpha1.ClusterServiceVersion) OperatorInfo {
	info := OperatorInfo{
		ResourceInfo: ResourceInfo{
			Name:        csv.Name,
			Namespace:   csv.Namespace,
			Kind:        "ClusterServiceVersion",
			APIVersion:  csv.APIVersion,
			Labels:      csv.Labels,
			Annotations: csv.Annotations,
			CreatedAt:   csv.CreationTimestamp.Time,
			Status:      string(csv.Status.Phase),
		},
		Phase:       string(csv.Status.Phase),
		Version:     csv.Spec.Version.String(),
		DisplayName: csv.Spec.DisplayName,
		Description: csv.Spec.Description,
		Age:         duration.HumanDuration(time.Since(csv.CreationTimestamp.Time)),
	}

	// Set provider
	if csv.Spec.Provider.Name != "" {
		info.Provider = OperatorProvider{
			Name: csv.Spec.Provider.Name,
			URL:  csv.Spec.Provider.URL,
		}
	}

	return info
}

func subscriptionToInfo(sub *operatorsv1alpha1.Subscription) SubscriptionInfo {
	info := SubscriptionInfo{
		ResourceInfo: ResourceInfo{
			Name:        sub.Name,
			Namespace:   sub.Namespace,
			Kind:        "Subscription",
			APIVersion:  sub.APIVersion,
			Labels:      sub.Labels,
			Annotations: sub.Annotations,
			CreatedAt:   sub.CreationTimestamp.Time,
			Status:      string(sub.Status.State),
		},
		Channel:                sub.Spec.Channel,
		StartingCSV:            sub.Spec.StartingCSV,
		CurrentCSV:             sub.Status.CurrentCSV,
		InstalledCSV:           sub.Status.InstalledCSV,
		InstallPlanGeneration:  int64(sub.Status.InstallPlanGeneration),
		State:                  string(sub.Status.State),
		Age:                    duration.HumanDuration(time.Since(sub.CreationTimestamp.Time)),
	}

	// Set install plan ref
	if sub.Status.InstallPlanRef != nil {
		info.InstallPlanRef = &InstallPlanRef{
			APIVersion:      sub.Status.InstallPlanRef.APIVersion,
			Kind:            sub.Status.InstallPlanRef.Kind,
			Name:            sub.Status.InstallPlanRef.Name,
			Namespace:       sub.Status.InstallPlanRef.Namespace,
			ResourceVersion: sub.Status.InstallPlanRef.ResourceVersion,
			UID:             string(sub.Status.InstallPlanRef.UID),
		}
	}

	return info
}