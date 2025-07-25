package resources

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecOptions contains options for executing commands in pods
type ExecOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Command       []string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	TTY           bool
}

// ExecuteInPod executes a command in a pod
func (c *K8sResourceClient) ExecuteInPod(ctx context.Context, opts ExecOptions) error {
	if opts.Namespace == "" {
		opts.Namespace = c.currentNamespace
	}

	// Get the pod to ensure it exists and to get container info
	pod, err := c.clientset.CoreV1().Pods(opts.Namespace).Get(ctx, opts.PodName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s/%s: %w", opts.Namespace, opts.PodName, err)
	}

	// If no container specified, use the first container
	if opts.ContainerName == "" && len(pod.Spec.Containers) > 0 {
		opts.ContainerName = pod.Spec.Containers[0].Name
	}

	// Create the exec request
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(opts.PodName).
		Namespace(opts.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: opts.ContainerName,
			Command:   opts.Command,
			Stdin:     opts.Stdin != nil,
			Stdout:    opts.Stdout != nil,
			Stderr:    opts.Stderr != nil,
			TTY:       opts.TTY,
		}, scheme.ParameterCodec)

	// Get the config from the client
	// Note: This requires the restConfig to be stored in K8sResourceClient
	if c.restConfig == nil {
		return fmt.Errorf("REST config not available for exec operations")
	}
	config := c.restConfig

	// Create the executor
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Execute the command
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	})
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// GetPodContainers returns the list of containers in a pod
func (c *K8sResourceClient) GetPodContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
	}

	var containers []string
	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	return containers, nil
}
