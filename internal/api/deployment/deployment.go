package deployment

import (
	"context"
	"fmt"

	"github.com/kaudit/val"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/kaudit/k8s_client"
)

// DeploymentAPI provides high-level methods for retrieving Kubernetes deployments.
type DeploymentAPI struct {
	client kubernetes.Interface
}

// NewDeploymentAPI creates a new DeploymentAPI instance using the provided client.
func NewDeploymentAPI(client kubernetes.Interface) api.DeploymentAPI {
	return &DeploymentAPI{
		client: client,
	}
}

// GetDeploymentByName retrieves a specific Deployment by namespace and name.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace of the deployment (must be non-empty).
//   - name: Name of the deployment (must be non-empty).
//
// Returns the matched *appsv1.Deployment or an error if not found or invalid.
func (d *DeploymentAPI) GetDeploymentByName(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(name, "required"); err != nil {
		return nil, fmt.Errorf("invalid deployment name: %w", err)
	}

	deploy, err := d.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %q in namespace %q: %w", name, namespace, err)
	}

	return deploy, nil
}

// ListDeploymentsByLabel lists deployments by namespace and label selector.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace scope.
//   - labelSelector: Kubernetes label selector syntax.
//
// Returns all matching deployments or an error.
func (d *DeploymentAPI) ListDeploymentsByLabel(ctx context.Context, namespace string, labelSelector string) ([]appsv1.Deployment, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(labelSelector, "required,k8s_label_selector"); err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	list, err := d.client.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments by label in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}

// ListDeploymentsByField lists deployments by namespace and field selector.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace scope.
//   - fieldSelector: Kubernetes field selector syntax.
//
// Returns all matching deployments or an error.
func (d *DeploymentAPI) ListDeploymentsByField(ctx context.Context, namespace string, fieldSelector string) ([]appsv1.Deployment, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(fieldSelector, "required,k8s_field_selector"); err != nil {
		return nil, fmt.Errorf("invalid field selector: %w", err)
	}

	opts := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	list, err := d.client.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments by field in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}
