package kubeconfig

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocksauth "github.com/kaudit/k8s_client/mocks/K8sAuthLoader"
)

func TestNewKubeConfigConnection(t *testing.T) {
	// Arrange
	mockLoader := &mocksauth.MockK8sAuthLoader{}

	// Act
	conn := NewKubeConfigConnection(mockLoader)

	// Assert
	require.NotNil(t, conn)
	assert.Same(t, mockLoader, conn.authLoader)
}

func TestGetRestConfig(t *testing.T) {
	t.Run("invalid kubeconfig format", func(t *testing.T) {
		// Arrange
		invalidKubeconfig := []byte(`invalid yaml`)

		// Act
		config, err := getRestConfig(invalidKubeconfig)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "clientcmd.Load failed")
		assert.Nil(t, config)
	})

	t.Run("missing current-context", func(t *testing.T) {
		// Arrange - this is valid YAML but missing current-context
		noContextKubeconfig := []byte(`
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://kubernetes.default.svc
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
# current-context is missing
users:
- name: test-user
  user: {}
`)

		// Act
		config, err := getRestConfig(noContextKubeconfig)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "clientCfg.ClientConfig failed")
		assert.Nil(t, config)
	})

	t.Run("valid kubeconfig", func(t *testing.T) {
		// A valid kubeconfig for testing the parsing functionality
		validKubeconfig := []byte(`
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://fake-kubernetes.example.com:6443
    insecure-skip-tls-verify: true
  name: fake-cluster
contexts:
- context:
    cluster: fake-cluster
    namespace: default
    user: fake-admin
  name: fake-context
current-context: fake-context
preferences: {}
users:
- name: fake-admin
  user:
    username: admin
    password: admin-password
`)

		config, err := getRestConfig(validKubeconfig)

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.NotEmpty(t, config.Host)
		assert.Equal(t, "https://fake-kubernetes.example.com:6443", config.Host)
	})
}

func TestKubeConfigConnection_NativeAPI(t *testing.T) {
	t.Run("successful client creation", func(t *testing.T) {
		// A valid but fake kubeconfig for testing purposes
		var validKubeConfigData = []byte(`
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://fake-kubernetes.example.com:6443
    insecure-skip-tls-verify: true
  name: fake-cluster
contexts:
- context:
    cluster: fake-cluster
    namespace: default
    user: fake-admin
  name: fake-context
current-context: fake-context
preferences: {}
users:
- name: fake-admin
  user:
    username: admin
    password: admin-password
`)
		// Arrange - use our valid but fake kubeconfig
		mockLoader := &mocksauth.MockK8sAuthLoader{}
		mockLoader.On("Load").Return(validKubeConfigData, nil).Once()

		conn := NewKubeConfigConnection(mockLoader)

		// Act
		client, err := conn.NativeAPI()

		// Assert
		if err != nil {
			// The error should be about connection failure, not parsing failure
			assert.NotContains(t, err.Error(), "clientcmd.Load failed")
			// Most likely the error will be about connecting to the fake server
			assert.Contains(t, err.Error(), "kubernetes.NewForConfig failed")
		} else {
			assert.NotNil(t, client)
		}

		mockLoader.AssertExpectations(t)
	})

	t.Run("loader error", func(t *testing.T) {
		// Arrange
		mockLoader := &mocksauth.MockK8sAuthLoader{}
		mockLoader.On("Load").Return(nil, errors.New("load error")).Once()

		conn := NewKubeConfigConnection(mockLoader)

		// Act
		client, err := conn.NativeAPI()

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authLoader.Load failed")
		assert.Nil(t, client)
		mockLoader.AssertExpectations(t)
	})

	t.Run("getRestConfig error", func(t *testing.T) {
		// Arrange
		mockLoader := &mocksauth.MockK8sAuthLoader{}
		mockLoader.On("Load").Return([]byte(`invalid yaml`), nil).Once()

		conn := NewKubeConfigConnection(mockLoader)

		// Act
		client, err := conn.NativeAPI()

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "getRestConfig failed")
		assert.Nil(t, client)
		mockLoader.AssertExpectations(t)
	})
}
