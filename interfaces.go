package api

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// DeploymentAPI defines an interface for interacting with Kubernetes Deployments.
// It provides high-level methods for retrieving and listing Deployments with input
// validation and pagination support. Methods support retrieving individual Deployments
// by name and listing Deployments using label or field selectors, all within the
// context of a specific namespace.
type DeploymentAPI interface {
	GetDeploymentByName(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	ListDeploymentsByLabel(ctx context.Context, namespace string, labelSelector string,
		timeoutSeconds time.Duration, limit int64) ([]appsv1.Deployment, error)
	ListDeploymentsByField(ctx context.Context, namespace string, fieldSelector string,
		timeoutSeconds time.Duration, limit int64) ([]appsv1.Deployment, error)
}

// NamespaceAPI defines an interface for interacting with Kubernetes Namespaces.
// It provides high-level methods for retrieving and listing Namespaces with input
// validation and pagination support. Unlike other resources, Namespaces are cluster-wide
// objects and don't exist within other namespaces, so no namespace parameter is required
// for listing operations. Methods support both label and field selector based filtering.
type NamespaceAPI interface {
	GetNamespaceByName(ctx context.Context, name string) (*corev1.Namespace, error)
	ListNamespacesByLabel(ctx context.Context, labelSelector string, timeoutSeconds time.Duration,
		limit int64) ([]corev1.Namespace, error)
	ListNamespacesByField(ctx context.Context, fieldSelector string, timeoutSeconds time.Duration,
		limit int64) ([]corev1.Namespace, error)
}

// ServiceAPI defines an interface for interacting with Kubernetes Services.
// It provides high-level methods for retrieving and listing Services with input
// validation and pagination support. All list operations handle fetching multiple
// pages of results automatically. Methods support retrieving individual Services by
// name and listing Services that match particular label or field selectors within
// a specific namespace.
type ServiceAPI interface {
	GetServiceByName(ctx context.Context, namespace, name string) (*corev1.Service, error)
	ListServicesByLabel(ctx context.Context, namespace string, labelSelector string,
		timeoutSeconds time.Duration, limit int64) ([]corev1.Service, error)
	ListServicesByField(ctx context.Context, namespace string, fieldSelector string,
		timeoutSeconds time.Duration, limit int64) ([]corev1.Service, error)
}

// PodAPI defines an interface for interacting with Kubernetes Pods.
// It provides high-level methods for retrieving and listing Pods with input
// validation and pagination support. All list operations handle fetching multiple
// pages of results automatically. Methods support retrieving individual Pods by name
// and listing Pods that match specific criteria using label or field selectors within
// a specific namespace.
type PodAPI interface {
	GetPodByName(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	ListPodsByLabel(ctx context.Context, namespace string, labelSelector string,
		timeoutSeconds time.Duration, limit int64) ([]corev1.Pod, error)
	ListPodsByField(ctx context.Context, namespace string, fieldSelector string,
		timeoutSeconds time.Duration, limit int64) ([]corev1.Pod, error)
}

// K8sAuthLoader defines a mechanism for loading Kubernetes authentication configuration data.
// It encapsulates the details of obtaining authentication information from various sources,
// such as service account tokens or kubeconfig files.
type K8sAuthLoader interface {
	Load() ([]byte, error)
}
