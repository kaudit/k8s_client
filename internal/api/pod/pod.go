// Package pod provides a high-level API for interacting with Kubernetes Pods.
// It wraps the Kubernetes client-go implementation with additional validation and error handling.
package pod

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

// PodAPI provides high-level methods for retrieving Kubernetes pods.
// It handles input validation and supports pagination for list operations.
type PodAPI struct {
	client kubernetes.Interface
}

// NewPodAPI creates a new PodAPI instance using the provided Kubernetes client.
// It returns an implementation of the api.PodAPI interface.
func NewPodAPI(client kubernetes.Interface) api.PodAPI {
	return &PodAPI{
		client: client,
	}
}

// GetPodByName retrieves a specific Pod by namespace and name.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace of the pod (must be non-empty).
//   - name: Name of the pod (must be non-empty).
//
// Returns the matched *corev1.Pod or an error if not found or invalid.
func (p *PodAPI) GetPodByName(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(name, "required"); err != nil {
		return nil, fmt.Errorf("invalid pod name: %w", err)
	}

	pod, err := p.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %q in namespace %q: %w", name, namespace, err)
	}

	return pod, nil
}

// ListPodsByLabel lists pods by namespace and label selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace scope for the query (must be non-empty).
//   - labelSelector: Kubernetes label selector syntax (e.g., "app=myapp,tier=frontend").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching pods across all pages or an error if validation fails or API calls fail.
func (p *PodAPI) ListPodsByLabel(ctx context.Context, namespace string, labelSelector string,
	timeoutSeconds time.Duration, limit int64) ([]corev1.Pod, error) {

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

	return p.loopForResult(ctx, namespace, opts)
}

// ListPodsByField lists pods by namespace and field selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace scope for the query (must be non-empty).
//   - fieldSelector: Kubernetes field selector syntax (e.g., "metadata.name=my-pod").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching pods across all pages or an error if validation fails or API calls fail.
func (p *PodAPI) ListPodsByField(ctx context.Context, namespace string, fieldSelector string,
	timeoutSeconds time.Duration, limit int64) ([]corev1.Pod, error) {

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

	return p.loopForResult(ctx, namespace, opts)
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
// until all matching pods are collected.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace to query.
//   - opts: List options including selectors, limit, and timeout.
//
// Returns the complete list of pods across all pages or an error if any API call fails.
func (p *PodAPI) loopForResult(ctx context.Context, namespace string,
	opts metav1.ListOptions) ([]corev1.Pod, error) {

	var result []corev1.Pod

	for {
		list, err := p.client.CoreV1().Pods(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pods in namespace %q: %w", namespace, err)
		}

		result = append(result, list.Items...)

		if list.Continue == "" {
			break
		}

		opts.Continue = list.Continue
	}

	return result, nil
}
