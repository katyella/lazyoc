package resources

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeletePod deletes a pod
func (c *K8sResourceClient) DeletePod(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	err := c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod %s/%s: %w", namespace, name, err)
	}

	return nil
}

// DeleteService deletes a service
func (c *K8sResourceClient) DeleteService(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	err := c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service %s/%s: %w", namespace, name, err)
	}

	return nil
}

// DeleteDeployment deletes a deployment
func (c *K8sResourceClient) DeleteDeployment(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s/%s: %w", namespace, name, err)
	}

	return nil
}

// DeleteConfigMap deletes a config map
func (c *K8sResourceClient) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	err := c.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete configmap %s/%s: %w", namespace, name, err)
	}

	return nil
}

// DeleteSecret deletes a secret
func (c *K8sResourceClient) DeleteSecret(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	err := c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete secret %s/%s: %w", namespace, name, err)
	}

	return nil
}