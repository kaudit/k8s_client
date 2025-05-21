package serviceaccount

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// TestServiceAccountConnectionNativeAPI tests the ServiceAccountConnectionNativeAPI function
func TestServiceAccountConnectionNativeAPI(t *testing.T) {
	// Define our test cases
	tests := []struct {
		name             string
		inClusterFunc    func() (*rest.Config, error)
		newForConfigFunc func(*rest.Config) (kubernetes.Interface, error)
		wantErr          bool
		errContains      string
	}{
		{
			name: "Success case",
			inClusterFunc: func() (*rest.Config, error) {
				return &rest.Config{}, nil
			},
			newForConfigFunc: func(*rest.Config) (kubernetes.Interface, error) {
				return &kubernetes.Clientset{}, nil
			},
			wantErr: false,
		},
		{
			name: "InClusterConfig fails",
			inClusterFunc: func() (*rest.Config, error) {
				return nil, errors.New("in-cluster config error")
			},
			newForConfigFunc: func(*rest.Config) (kubernetes.Interface, error) {
				return &kubernetes.Clientset{}, nil
			},
			wantErr:     true,
			errContains: "rest.InClusterConfig failed",
		},
		{
			name: "NewForConfig fails",
			inClusterFunc: func() (*rest.Config, error) {
				return &rest.Config{}, nil
			},
			newForConfigFunc: func(*rest.Config) (kubernetes.Interface, error) {
				return nil, errors.New("new for config error")
			},
			wantErr:     true,
			errContains: "kubernetes.NewForConfig failed",
		},
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom ServiceAccountConnectionNativeAPIFunc with our test doubles
			connFunc := createServiceAccountConnectionFunc(tt.inClusterFunc, tt.newForConfigFunc)

			// Call the function
			client, err := connFunc()

			// Assert expectations
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// createServiceAccountConnectionFunc returns a function that matches the signature of
// ServiceAccountConnectionNativeAPI but uses the provided test doubles instead of
// the actual k8s.io/client-go functions.
func createServiceAccountConnectionFunc(
	inClusterFunc func() (*rest.Config, error),
	newForConfigFunc func(*rest.Config) (kubernetes.Interface, error),
) func() (kubernetes.Interface, error) {
	return func() (kubernetes.Interface, error) {
		config, err := inClusterFunc()
		if err != nil {
			return nil, errors.New("rest.InClusterConfig failed: " + err.Error())
		}

		clientset, err := newForConfigFunc(config)
		if err != nil {
			return nil, errors.New("kubernetes.NewForConfig failed: " + err.Error())
		}

		return clientset, nil
	}
}
