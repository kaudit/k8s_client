package deployment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewDeploymentAPI(t *testing.T) {
	client := fake.NewClientset()
	api := NewDeploymentAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*DeploymentAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestDeploymentAPI_GetDeploymentByName(t *testing.T) {
	// Setup a deployment with desired characteristics
	replicas := int32(3)
	testDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 3,
		},
	}

	// Create fake clientset with test deployment
	fakeClient := fake.NewClientset(testDeployment)

	// Initialize deployment API
	deploymentAPI := NewDeploymentAPI(fakeClient)

	// Test cases
	tests := []struct {
		name           string
		namespace      string
		deploymentName string
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "Successfully get deployment",
			namespace:      "test-namespace",
			deploymentName: "test-deployment",
			wantErr:        false,
		},
		{
			name:           "Empty namespace",
			namespace:      "",
			deploymentName: "test-deployment",
			wantErr:        true,
			errorContains:  "invalid namespace",
		},
		{
			name:           "Empty deployment name",
			namespace:      "test-namespace",
			deploymentName: "",
			wantErr:        true,
			errorContains:  "invalid deployment name",
		},
		{
			name:           "Deployment not found",
			namespace:      "test-namespace",
			deploymentName: "nonexistent-deployment",
			wantErr:        true,
			errorContains:  "failed to get deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			deployment, err := deploymentAPI.GetDeploymentByName(ctx, tt.namespace, tt.deploymentName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, deployment)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, deployment)
				assert.Equal(t, tt.deploymentName, deployment.Name)
				assert.Equal(t, tt.namespace, deployment.Namespace)
				assert.Equal(t, int32(3), *deployment.Spec.Replicas)
				assert.Equal(t, int32(3), deployment.Status.ReadyReplicas)
				assert.Equal(t, "test-app", deployment.Spec.Selector.MatchLabels["app"])
			}
		})
	}
}

func TestDeploymentAPI_ListDeploymentsByLabel(t *testing.T) {
	// Setup test deployments
	replicas := int32(3)
	testDeployments := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment-1",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "test-app",
					"environment": "production",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment-2",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "test-app",
					"environment": "staging",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-deployment",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "other-app",
					"environment": "production",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		},
	}

	// Create fake clientset
	fakeClient := fake.NewClientset(testDeployments[0], testDeployments[1], testDeployments[2])

	// Initialize deployment API
	deploymentAPI := NewDeploymentAPI(fakeClient)

	// Test cases
	tests := []struct {
		name          string
		namespace     string
		labelSelector string
		expectedCount int
		expectedNames []string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "List deployments by app label",
			namespace:     "test-namespace",
			labelSelector: "app=test-app",
			expectedCount: 2,
			expectedNames: []string{"test-deployment-1", "test-deployment-2"},
			wantErr:       false,
		},
		{
			name:          "List deployments by environment label",
			namespace:     "test-namespace",
			labelSelector: "environment=production",
			expectedCount: 2,
			expectedNames: []string{"test-deployment-1", "other-deployment"},
			wantErr:       false,
		},
		{
			name:          "List deployments with multiple labels",
			namespace:     "test-namespace",
			labelSelector: "app=test-app,environment=production",
			expectedCount: 1,
			expectedNames: []string{"test-deployment-1"},
			wantErr:       false,
		},
		{
			name:          "No results",
			namespace:     "test-namespace",
			labelSelector: "app=nonexistent",
			expectedCount: 0,
			expectedNames: []string{},
			wantErr:       false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			labelSelector: "app=test-app",
			wantErr:       true,
			errorContains: "invalid namespace",
		},
		{
			name:          "Empty label selector",
			namespace:     "test-namespace",
			labelSelector: "",
			wantErr:       true,
			errorContains: "invalid label selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			deployments, err := deploymentAPI.ListDeploymentsByLabel(ctx, tt.namespace, tt.labelSelector)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, deployments, tt.expectedCount)

				// Check if all expected deployments are present
				if tt.expectedCount > 0 {
					foundNames := make([]string, len(deployments))
					for i, deployment := range deployments {
						foundNames[i] = deployment.Name
					}

					for _, expectedName := range tt.expectedNames {
						assert.Contains(t, foundNames, expectedName)
					}
				}
			}
		})
	}
}

func TestDeploymentAPI_ListDeploymentsByField(t *testing.T) {
	// Setup test deployments
	replicas := int32(3)
	testDeployments := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment-1",
				Namespace: "test-namespace",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment-2",
				Namespace: "other-namespace",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		},
	}

	// Create fake clientset with both test deployments
	fakeClient := fake.NewClientset(testDeployments[0], testDeployments[1])

	// Initialize deployment API
	deploymentAPI := NewDeploymentAPI(fakeClient)

	// Test cases
	tests := []struct {
		name          string
		namespace     string
		fieldSelector string
		expectedCount int
		wantErr       bool
		errorContains string
	}{
		{
			name:          "List deployments by field",
			namespace:     "test-namespace",
			fieldSelector: "metadata.name=test-deployment-1",
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			fieldSelector: "metadata.name=test-deployment-1",
			wantErr:       true,
			errorContains: "invalid namespace",
		},
		{
			name:          "Empty field selector",
			namespace:     "test-namespace",
			fieldSelector: "",
			wantErr:       true,
			errorContains: "invalid field selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			deployments, err := deploymentAPI.ListDeploymentsByField(ctx, tt.namespace, tt.fieldSelector)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, deployments, tt.expectedCount)
			}
		})
	}
}
