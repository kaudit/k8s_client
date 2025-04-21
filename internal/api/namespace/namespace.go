package namespace

import (
	"context"
	"fmt"

	"github.com/kaudit/val"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/kaudit/k8s_client"
)

// NamespaceAPI provides high-level methods for retrieving and manipulating Kubernetes namespaces.
type NamespaceAPI struct {
	client kubernetes.Interface
}

// NewNamespaceAPI creates a new NamespaceAPI instance with the provided Kubernetes client.
//
// The client parameter should be a valid implementation of kubernetes.Interface.
//
// Returns an initialized *NamespaceAPI.
func NewNamespaceAPI(client kubernetes.Interface) api.NamespaceAPI {
	return &NamespaceAPI{
		client: client,
	}
}

// GetNamespaceByName retrieves a single Namespace object by its name.
//
// The name parameter is validated to ensure it is not empty.
// If the validation fails or if the retrieval from the Kubernetes API fails, an error is returned.
//
//   - ctx: The context to use for cancellation.
//   - name: The name of the Kubernetes namespace to retrieve.
//
// Returns a pointer to a corev1.Namespace object or an error if the namespace
// is not found or if any other retrieval error occurs.
func (n *NamespaceAPI) GetNamespaceByName(ctx context.Context, name string) (*corev1.Namespace, error) {
	err := val.ValidateWithTag(name, "required")
	if err != nil {
		return nil, fmt.Errorf("failed to validate namespace name: %w", err)
	}

	ns, err := n.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %q: %w", name, err)
	}
	return ns, nil
}

// ListNamespacesByLabel retrieves a list of Namespace objects filtered by a label selector.
//
// The labelSelector parameter is validated to ensure it uses a valid Kubernetes
// label selector syntax.
// If the validation fails or if the Kubernetes API call fails, an error is returned.
//
//   - ctx: The context to use for cancellation.
//   - labelSelector: The Kubernetes-compliant label selector string.
//
// Returns a slice of corev1.Namespace objects matching the label selector, or
// an error if the operation fails.
func (n *NamespaceAPI) ListNamespacesByLabel(ctx context.Context, labelSelector string) ([]corev1.Namespace, error) {
	err := val.ValidateWithTag(labelSelector, "required,k8s_label_selector")
	if err != nil {
		return nil, fmt.Errorf("failed to validate label selector: %w", err)
	}

	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	list, err := n.client.CoreV1().Namespaces().List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces by label %q: %w", labelSelector, err)
	}

	return list.Items, nil
}

// ListNamespacesByField retrieves a list of Namespace objects filtered by a field selector.
//
// The fieldSelector parameter is validated to ensure it uses a valid Kubernetes
// field selector syntax.
// An error is returned if the validation or the Kubernetes API call fails.
//
//   - ctx: The context to use for cancellation.
//   - fieldSelector: The Kubernetes-compliant field selector string.
//
// Returns a slice of corev1.Namespace objects matching the field selector, or
// an error if the operation fails.
func (n *NamespaceAPI) ListNamespacesByField(ctx context.Context, fieldSelector string) ([]corev1.Namespace, error) {
	err := val.ValidateWithTag(fieldSelector, "required,k8s_field_selector")
	if err != nil {
		return nil, fmt.Errorf("failed to validate field selector: %w", err)
	}

	opts := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	list, err := n.client.CoreV1().Namespaces().List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces by field %q: %w", fieldSelector, err)
	}

	return list.Items, nil
}
