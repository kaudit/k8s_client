package service

import (
	"context"
	"fmt"

	"github.com/kaudit/val"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/kaudit/k8s_client"
)

// ServiceAPI provides high-level methods for retrieving Kubernetes services.
type ServiceAPI struct {
	client kubernetes.Interface
}

// NewServiceAPI creates a new ServiceAPI instance using the provided client.
func NewServiceAPI(client kubernetes.Interface) api.ServiceAPI {
	return &ServiceAPI{
		client: client,
	}
}

// GetServiceByName retrieves a specific Service by namespace and name.
//
// Parameters:
//   - ctx: Context for cancellation.
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

// ListServicesByLabel lists services by namespace and label selector.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace scope.
//   - labelSelector: Kubernetes label selector syntax.
//
// Returns all matching services or an error.
func (s *ServiceAPI) ListServicesByLabel(ctx context.Context, namespace string, labelSelector string) ([]corev1.Service, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(labelSelector, "required,k8s_label_selector"); err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	list, err := s.client.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list services by label in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}

// ListServicesByField lists services by namespace and field selector.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace scope.
//   - fieldSelector: Kubernetes field selector syntax.
//
// Returns all matching services or an error.
func (s *ServiceAPI) ListServicesByField(ctx context.Context, namespace string, fieldSelector string) ([]corev1.Service, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(fieldSelector, "required,k8s_field_selector"); err != nil {
		return nil, fmt.Errorf("invalid field selector: %w", err)
	}

	opts := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	list, err := s.client.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list services by field in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}
