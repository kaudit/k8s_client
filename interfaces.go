package api

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// DeploymentAPI defines an interface for interacting with Kubernetes Deployments.
// This interface provides a simplified abstraction over the Kubernetes client-go library
// for common Deployment operations. It allows retrieving individual Deployments by name
// and listing Deployments by either label selectors or field selectors, all within the
// context of a specific namespace.
type DeploymentAPI interface {
	GetDeploymentByName(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	ListDeploymentsByLabel(ctx context.Context, namespace string, labelSelector string, timeoutSeconds time.Duration, limit int64) ([]appsv1.Deployment, error)
	ListDeploymentsByField(ctx context.Context, namespace string, fieldSelector string, timeoutSeconds time.Duration, limit int64) ([]appsv1.Deployment, error)
}

// NamespaceAPI defines an interface for interacting with Kubernetes Namespaces.
// This interface abstracts the underlying Kubernetes API operations related to Namespaces,
// providing methods to retrieve individual Namespaces by name and to list Namespaces
// matching certain criteria using label or field selectors. Unlike other resources,
// Namespaces are cluster-wide objects and don't exist within other namespaces,
// so no namespace parameter is required for listing operations.
type NamespaceAPI interface {
	GetNamespaceByName(ctx context.Context, name string) (*corev1.Namespace, error)
	ListNamespacesByLabel(ctx context.Context, labelSelector string, timeoutSeconds time.Duration, limit int64) ([]corev1.Namespace, error)
	ListNamespacesByField(ctx context.Context, fieldSelector string, timeoutSeconds time.Duration, limit int64) ([]corev1.Namespace, error)
}

// ServiceAPI defines an interface for interacting with Kubernetes Services.
// This interface provides simplified access to Service-related operations within
// the Kubernetes API. It includes functionality to retrieve individual Services by
// their name within a specific namespace, as well as list Services that match
// particular label or field selectors. Services provide network access to sets of Pods,
// and this interface helps abstract the details of how these Services are queried.
type ServiceAPI interface {
	GetServiceByName(ctx context.Context, namespace, name string) (*corev1.Service, error)
	ListServicesByLabel(ctx context.Context, namespace string, labelSelector string) ([]corev1.Service, error)
	ListServicesByField(ctx context.Context, namespace string, fieldSelector string) ([]corev1.Service, error)
}

// PodAPI defines an interface for interacting with Kubernetes Pods.
// This interface abstracts operations related to Pods, which are the smallest
// deployable units in Kubernetes that can be created and managed. It provides
// methods to retrieve individual Pods by name within a namespace, and to list
// Pods that match specific criteria using label or field selectors. Pods
// represent containers running on your cluster, and this interface simplifies
// interaction with them.
type PodAPI interface {
	GetPodByName(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	ListPodsByLabel(ctx context.Context, namespace string, labelSelector string) ([]corev1.Pod, error)
	ListPodsByField(ctx context.Context, namespace string, fieldSelector string) ([]corev1.Pod, error)
}

// K8sAuthLoader defines a mechanism for loading Kubernetes authentication configuration data.
// It supports loading from both a default source and a user-specified file path.
type K8sAuthLoader interface {
	Load() ([]byte, error)
}
