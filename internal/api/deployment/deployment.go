// Package deployment provides a high-level API for interacting with Kubernetes Deployments.
// It wraps the Kubernetes client-go implementation with additional validation and error handling.
package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/kaudit/val"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/kaudit/k8s_client"
)

// DeploymentAPI provides high-level methods for retrieving Kubernetes deployments.
// It handles input validation and supports pagination for list operations.
type DeploymentAPI struct {
	client kubernetes.Interface
}

// NewDeploymentAPI creates a new DeploymentAPI instance using the provided Kubernetes client.
// It returns an implementation of the api.DeploymentAPI interface.
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

// ListDeploymentsByLabel lists deployments by namespace and label selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace scope for the query (must be non-empty).
//   - labelSelector: Kubernetes label selector syntax (e.g., "app=myapp,tier=frontend").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching deployments across all pages or an error if validation fails or API calls fail.
func (d *DeploymentAPI) ListDeploymentsByLabel(ctx context.Context, namespace string, labelSelector string,
	timeoutSeconds time.Duration, limit int64) ([]appsv1.Deployment, error) {

	if err := validateInput(namespace, timeoutSeconds, limit); err != nil {
		return nil, err
	}
	if err := val.ValidateWithTag(labelSelector, "required,k8s_label_selector"); err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	seconds := int64(timeoutSeconds.Seconds())

	opts := metav1.ListOptions{
		LabelSelector:  labelSelector,
		Limit:          limit,
		TimeoutSeconds: &seconds,
	}

	return d.loopForResult(ctx, namespace, opts)
}

// ListDeploymentsByField lists deployments by namespace and field selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace scope for the query (must be non-empty).
//   - fieldSelector: Kubernetes field selector syntax (e.g., "metadata.name=my-deployment").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching deployments across all pages or an error if validation fails or API calls fail.
func (d *DeploymentAPI) ListDeploymentsByField(ctx context.Context, namespace string, fieldSelector string,
	timeoutSeconds time.Duration, limit int64) ([]appsv1.Deployment, error) {

	if err := validateInput(namespace, timeoutSeconds, limit); err != nil {
		return nil, err
	}
	if err := val.ValidateWithTag(fieldSelector, "required,k8s_field_selector"); err != nil {
		return nil, fmt.Errorf("invalid field selector: %w", err)
	}

	seconds := int64(timeoutSeconds.Seconds())

	opts := metav1.ListOptions{
		FieldSelector:  fieldSelector,
		Limit:          limit,
		TimeoutSeconds: &seconds,
	}

	return d.loopForResult(ctx, namespace, opts)
}

// validateInput validates common input parameters for list operations.
// It checks that namespace is non-empty, timeout is at least 1 second, and limit is positive.
// Returns an error with detailed information if validation fails.
func validateInput(namespace string, timeoutSeconds time.Duration, limit int64) error {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return fmt.Errorf("invalid namespace: %w", err)
	}

	if err := val.ValidateWithTag(timeoutSeconds, "required,min=1s"); err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	if err := val.ValidateWithTag(limit, "required,gt=0"); err != nil {
		return fmt.Errorf("invalid limit: %w", err)
	}

	return nil
}

// loopForResult handles pagination for list operations by repeatedly fetching pages of results
// until all matching deployments are collected.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace to query.
//   - opts: List options including selectors, limit, and timeout.
//
// Returns the complete list of deployments across all pages or an error if any API call fails.
func (d *DeploymentAPI) loopForResult(ctx context.Context, namespace string,
	opts metav1.ListOptions) ([]appsv1.Deployment, error) {

	var result []appsv1.Deployment

	for {
		list, err := d.client.AppsV1().Deployments(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list deployments in namespace %q: %w", namespace, err)
		}

		result = append(result, list.Items...)

		if list.Continue == "" {
			break
		}

		opts.Continue = list.Continue
	}

	return result, nil
}
