// Package namespace provides a high-level API for interacting with Kubernetes Namespaces.
// It wraps the Kubernetes client-go implementation with additional validation and error handling.
package namespace

import (
	"context"
	"fmt"
	"time"

	"github.com/kaudit/val"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/kaudit/k8s_client"
)

// NamespaceAPI provides high-level methods for retrieving Kubernetes namespaces.
// It handles input validation and supports pagination for list operations.
type NamespaceAPI struct {
	client kubernetes.Interface
}

// NewNamespaceAPI creates a new NamespaceAPI instance using the provided Kubernetes client.
// It returns an implementation of the api.NamespaceAPI interface.
func NewNamespaceAPI(client kubernetes.Interface) api.NamespaceAPI {
	return &NamespaceAPI{
		client: client,
	}
}

// GetNamespaceByName retrieves a specific Namespace by name.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - name: Name of the namespace (must be non-empty).
//
// Returns the matched *corev1.Namespace or an error if not found or invalid.
func (n *NamespaceAPI) GetNamespaceByName(ctx context.Context, name string) (*corev1.Namespace, error) {
	if err := val.ValidateWithTag(name, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace name: %w", err)
	}

	ns, err := n.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %q: %w", name, err)
	}
	return ns, nil
}

// ListNamespacesByLabel lists namespaces by label selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - labelSelector: Kubernetes label selector syntax (e.g., "app=myapp,tier=frontend").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching namespaces across all pages or an error if validation fails or API calls fail.
func (n *NamespaceAPI) ListNamespacesByLabel(ctx context.Context, labelSelector string,
	timeoutSeconds time.Duration, limit int64) ([]corev1.Namespace, error) {

	if err := validateInput(timeoutSeconds, limit); err != nil {
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

	return n.loopForResult(ctx, opts)
}

// ListNamespacesByField lists namespaces by field selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - fieldSelector: Kubernetes field selector syntax (e.g., "metadata.name=my-namespace").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching namespaces across all pages or an error if validation fails or API calls fail.
func (n *NamespaceAPI) ListNamespacesByField(ctx context.Context, fieldSelector string,
	timeoutSeconds time.Duration, limit int64) ([]corev1.Namespace, error) {

	if err := validateInput(timeoutSeconds, limit); err != nil {
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

	return n.loopForResult(ctx, opts)
}

// validateInput validates common input parameters for list operations.
// It checks that timeout is at least 1 second and limit is positive.
// Returns an error with detailed information if validation fails.
func validateInput(timeoutSeconds time.Duration, limit int64) error {
	if err := val.ValidateWithTag(timeoutSeconds, "required,min=1s"); err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	if err := val.ValidateWithTag(limit, "required,gt=0"); err != nil {
		return fmt.Errorf("invalid limit: %w", err)
	}

	return nil
}

// loopForResult handles pagination for list operations by repeatedly fetching pages of results
// until all matching namespaces are collected.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - opts: List options including selectors, limit, and timeout.
//
// Returns the complete list of namespaces across all pages or an error if any API call fails.
func (n *NamespaceAPI) loopForResult(ctx context.Context, opts metav1.ListOptions) ([]corev1.Namespace, error) {
	var result []corev1.Namespace

	for {
		list, err := n.client.CoreV1().Namespaces().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list namespaces: %w", err)
		}

		result = append(result, list.Items...)

		if list.Continue == "" {
			break
		}

		opts.Continue = list.Continue
	}

	return result, nil
}
