package pod

import (
	"context"
	"fmt"

	"github.com/kaudit/val"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/kaudit/k8s_client"
)

// PodAPI provides high-level methods for retrieving Kubernetes pods.
type PodAPI struct {
	client kubernetes.Interface
}

// NewPodAPI creates a new PodAPI instance using the provided client.
func NewPodAPI(client kubernetes.Interface) api.PodAPI {
	return &PodAPI{
		client: client,
	}
}

// GetPodByName retrieves a specific Pod by namespace and name.
//
// Parameters:
//   - ctx: Context for cancellation.
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

// ListPodsByLabel lists pods by namespace and label selector.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace scope.
//   - labelSelector: Kubernetes label selector syntax.
//
// Returns all matching pods or an error.
func (p *PodAPI) ListPodsByLabel(ctx context.Context, namespace string, labelSelector string) ([]corev1.Pod, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(labelSelector, "required,k8s_label_selector"); err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	list, err := p.client.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods by label in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}

// ListPodsByField lists pods by namespace and field selector.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - namespace: Namespace scope.
//   - fieldSelector: Kubernetes field selector syntax.
//
// Returns all matching pods or an error.
func (p *PodAPI) ListPodsByField(ctx context.Context, namespace string, fieldSelector string) ([]corev1.Pod, error) {
	if err := val.ValidateWithTag(namespace, "required"); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := val.ValidateWithTag(fieldSelector, "required,k8s_field_selector"); err != nil {
		return nil, fmt.Errorf("invalid field selector: %w", err)
	}

	opts := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	list, err := p.client.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods by field in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}
