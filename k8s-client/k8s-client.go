package k8sclient

import (
	"errors"
	"fmt"

	"github.com/kaudit/val"

	api "github.com/kaudit/k8s_client"
	"github.com/kaudit/k8s_client/internal/api/deployment"
	"github.com/kaudit/k8s_client/internal/api/namespace"
	"github.com/kaudit/k8s_client/internal/api/pod"
	"github.com/kaudit/k8s_client/internal/api/service"
	"github.com/kaudit/k8s_client/internal/connection/kubeconfig"
	"github.com/kaudit/k8s_client/internal/connection/serviceaccount"
)

var ErrAlreadyConfigured = errors.New("k8s client already configured")

// K8sClient provides a centralized access point to high-level Kubernetes API abstractions.
//
// It encapsulates typed interfaces for interacting with Pods, Services, Deployments,
// and Namespaces — each exposed through domain-specific interface contracts.
//
// All API implementations are stateless, thread-safe, and validated via typed input contracts.
type K8sClient struct {
	pods        api.PodAPI        `validator:"required"`
	services    api.ServiceAPI    `validator:"required"`
	deployments api.DeploymentAPI `validator:"required"`
	namespaces  api.NamespaceAPI  `validator:"required"`
}

type K8sClientOption func(*K8sClient) error

// WithKubeConfigLoader creates a K8sClientOption that configures the client to use
// authentication via a kubeconfig file. This option requires a K8sAuthLoader implementation
// that can load the kubeconfig data.
//
// This option is mutually exclusive with WithServiceAccount. Using both options in the same
// client initialization will result in the last applied option overriding previous authentication
// configuration.
func WithKubeConfigLoader(loader api.K8sAuthLoader) K8sClientOption {
	return func(k8sClient *K8sClient) error {
		err := val.ValidateStruct(k8sClient)
		if err == nil {
			return ErrAlreadyConfigured
		}

		n, err := kubeconfig.NewKubeConfigConnection(loader).NativeAPI()
		if err != nil {
			return fmt.Errorf("failed to init k8s client: %w", err)
		}

		k8sClient.pods = pod.NewPodAPI(n)
		k8sClient.services = service.NewServiceAPI(n)
		k8sClient.deployments = deployment.NewDeploymentAPI(n)
		k8sClient.namespaces = namespace.NewNamespaceAPI(n)

		return nil
	}
}

// WithServiceAccount creates a K8sClientOption that configures the client to use
// in-cluster authentication via the service account token mounted in the pod.
// This option should be used when the application is running inside a Kubernetes cluster.
//
// This option is mutually exclusive with WithKubeConfigLoader. Using both options in the same
// client initialization will result in the last applied option overriding previous authentication
// configuration.
func WithServiceAccount() K8sClientOption {
	return func(k8sClient *K8sClient) error {
		err := val.ValidateStruct(k8sClient)
		if err == nil {
			return ErrAlreadyConfigured
		}

		n, err := serviceaccount.ServiceAccountConnectionNativeAPI()
		if err != nil {
			return fmt.Errorf("failed to init k8s client with service account: %w", err)
		}

		k8sClient.pods = pod.NewPodAPI(n)
		k8sClient.services = service.NewServiceAPI(n)
		k8sClient.deployments = deployment.NewDeploymentAPI(n)
		k8sClient.namespaces = namespace.NewNamespaceAPI(n)

		return nil
	}
}

func NewK8sClient(options ...K8sClientOption) (*K8sClient, error) {
	client := &K8sClient{}

	for _, option := range options {
		if err := option(client); err != nil {
			return nil, fmt.Errorf("failed to configure k8s client: %w", err)
		}
	}

	err := val.ValidateStruct(client)
	if err != nil {
		return nil, fmt.Errorf("failed to validate k8s client: %w", err)
	}

	return client, nil
}

// GetPodAPI exposes the PodAPI interface, allowing access to pod-specific operations.
func (k *K8sClient) GetPodAPI() api.PodAPI { return k.pods }

// GetServiceAPI exposes the ServiceAPI interface for service-level operations.
func (k *K8sClient) GetServiceAPI() api.ServiceAPI {
	return k.services
}

// GetDeploymentAPI exposes the DeploymentAPI interface for managing deployments.
func (k *K8sClient) GetDeploymentAPI() api.DeploymentAPI {
	return k.deployments
}

// GetNamespaceAPI exposes the NamespaceAPI interface for managing namespaces.
func (k *K8sClient) GetNamespaceAPI() api.NamespaceAPI {
	return k.namespaces
}
