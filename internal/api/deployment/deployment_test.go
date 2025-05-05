package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentAPI_New(t *testing.T) {
	client := fake.NewClientset()
	api := NewDeploymentAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*DeploymentAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestValidateInput(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		wantErr        bool
		errMsg         string
		namespace      string
		timeoutSeconds time.Duration
		limit          int64
	}{
		{
			name:           "Valid input",
			input:          "test-deployment",
			wantErr:        false,
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          2,
		},
		{
			name:           "empty namespace",
			input:          "test-deployment",
			wantErr:        true,
			errMsg:         "invalid namespace",
			namespace:      "",
			timeoutSeconds: 2 * time.Second,
			limit:          2,
		},
		{
			name:           "invalid timeout",
			input:          "test-deployment",
			wantErr:        true,
			errMsg:         "invalid timeout",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          2,
		},
		{
			name:           "invalid limit - zero value",
			input:          "test-deployment",
			wantErr:        true,
			errMsg:         "invalid limit",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          0,
		},
		{
			name:           "invalid limit - negative value",
			input:          "test-deployment",
			wantErr:        true,
			errMsg:         "invalid limit",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
		},
	}

	for _, testCase := range testCases {
		err := validateInput(testCase.namespace, testCase.timeoutSeconds, testCase.limit)
		if testCase.wantErr {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.errMsg)
		} else {
			require.NoError(t, err)
		}
	}
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
		name           string
		namespace      string
		labelSelector  string
		timeoutSeconds time.Duration
		limit          int64
		expectedCount  int
		expectedNames  []string
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "List deployments by app label",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"test-deployment-1", "test-deployment-2"},
			wantErr:        false,
		},
		{
			name:           "List deployments by environment label",
			namespace:      "test-namespace",
			labelSelector:  "environment=production",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"test-deployment-1", "other-deployment"},
			wantErr:        false,
		},
		{
			name:           "List deployments with multiple labels",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app,environment=production",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			expectedNames:  []string{"test-deployment-1"},
			wantErr:        false,
		},
		{
			name:           "No results",
			namespace:      "test-namespace",
			labelSelector:  "app=nonexistent",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  0,
			expectedNames:  []string{},
			wantErr:        false,
		},
		{
			name:           "Empty namespace",
			namespace:      "",
			labelSelector:  "app=test-app",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid namespace",
		},
		{
			name:           "Empty label selector",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			labelSelector:  "",
			wantErr:        true,
			errorContains:  "invalid label selector",
		},
		{
			name:           "Invalid timeout",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          1,
			labelSelector:  "",
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
			labelSelector:  "",
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			deployments, err := deploymentAPI.ListDeploymentsByLabel(ctx,
				testCase.namespace,
				testCase.labelSelector,
				testCase.timeoutSeconds,
				testCase.limit,
			)

			if testCase.wantErr {
				require.Error(t, err)
				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, deployments, testCase.expectedCount)

				// Check if all expected deployments are present
				if testCase.expectedCount > 0 {
					foundNames := make([]string, len(deployments))
					for i, deployment := range deployments {
						foundNames[i] = deployment.Name
					}

					for _, expectedName := range testCase.expectedNames {
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
		name           string
		namespace      string
		fieldSelector  string
		timeoutSeconds time.Duration
		limit          int64
		expectedCount  int
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "List deployments by field",
			namespace:      "test-namespace",
			fieldSelector:  "metadata.name=test-deployment-1",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "Empty namespace",
			namespace:      "",
			fieldSelector:  "metadata.name=test-deployment-1",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid namespace",
		},
		{
			name:           "Empty field selector",
			namespace:      "test-namespace",
			fieldSelector:  "",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid field selector",
		},
		{
			name:           "Invalid timeout",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          1,
			fieldSelector:  "metadata.name=test-deployment-1",
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
			fieldSelector:  "metadata.name=test-deployment-1",
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			deployments, err := deploymentAPI.ListDeploymentsByField(
				ctx,
				testCase.namespace,
				testCase.fieldSelector,
				testCase.timeoutSeconds,
				testCase.limit,
			)

			if testCase.wantErr {
				require.Error(t, err)
				if testCase.errorContains != "" {
					assert.Contains(t, err.Error(), testCase.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, deployments, testCase.expectedCount)
			}
		})
	}
}
