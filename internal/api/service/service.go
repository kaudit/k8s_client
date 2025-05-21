// Package service provides a high-level API for interacting with Kubernetes Services.
// It wraps the Kubernetes client-go implementation with additional validation and error handling.
package service

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

// ServiceAPI provides high-level methods for retrieving Kubernetes services.
// It handles input validation and supports pagination for list operations.
type ServiceAPI struct {
	client kubernetes.Interface
}

// NewServiceAPI creates a new ServiceAPI instance using the provided Kubernetes client.
// It returns an implementation of the api.ServiceAPI interface.
func NewServiceAPI(client kubernetes.Interface) api.ServiceAPI {
	return &ServiceAPI{
		client: client,
	}
}

// GetServiceByName retrieves a specific Service by namespace and name.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace of the service (must be non-empty).
//   - name: Name of the service (must be non-empty).
//
// Returns the matched *corev1.Service or an error if not found or invalid.
func (s *ServiceAPI) GetServiceByName(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(name, "required"); err != nil {
		return nil, fmt.Errorf("invalid service name: %w", err)
	}

	svc, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %q in namespace %q: %w", name, namespace, err)
	}

	return svc, nil
}

// ListServicesByLabel lists services by namespace and label selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace scope for the query (must be non-empty).
//   - labelSelector: Kubernetes label selector syntax (e.g., "app=myapp,tier=frontend").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching services across all pages or an error if validation fails or API calls fail.
func (s *ServiceAPI) ListServicesByLabel(ctx context.Context, namespace string, labelSelector string,
	timeoutSeconds time.Duration, limit int64) ([]corev1.Service, error) {

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

	return s.loopForResult(ctx, namespace, opts)
}

// ListServicesByField lists services by namespace and field selector with pagination support.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//   - namespace: Namespace scope for the query (must be non-empty).
//   - fieldSelector: Kubernetes field selector syntax (e.g., "metadata.name=my-service").
//   - timeoutSeconds: Timeout duration for the API call (must be at least 1s).
//   - limit: Maximum number of results per page (must be greater than 0).
//
// Returns all matching services across all pages or an error if validation fails or API calls fail.
func (s *ServiceAPI) ListServicesByField(ctx context.Context, namespace string, fieldSelector string,
	timeoutSeconds time.Duration, limit int64) ([]corev1.Service, error) {

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

	return s.loopForResult(ctx, namespace, opts)
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
// until all matching services are collected.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace to query.
//   - opts: List options including selectors, limit, and timeout.
//
// Returns the complete list of services across all pages or an error if any API call fails.
func (s *ServiceAPI) loopForResult(ctx context.Context, namespace string,
	opts metav1.ListOptions) ([]corev1.Service, error) {

	var result []corev1.Service

	for {
		list, err := s.client.CoreV1().Services(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list services in namespace %q: %w", namespace, err)
		}

		result = append(result, list.Items...)

		if list.Continue == "" {
			break
		}

		opts.Continue = list.Continue
	}

	return result, nil
}
