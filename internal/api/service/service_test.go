package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestServiceAPI_New(t *testing.T) {
	client := fake.NewClientset()
	api := NewServiceAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*ServiceAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestValidateInput(t *testing.T) {
	testCases := []struct {
		name           string
		namespace      string
		timeoutSeconds time.Duration
		limit          int64
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "Valid input",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          2,
			wantErr:        false,
		},
		{
			name:           "empty namespace",
			namespace:      "",
			timeoutSeconds: 2 * time.Second,
			limit:          2,
			wantErr:        true,
			errMsg:         "invalid namespace",
		},
		{
			name:           "invalid timeout",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          2,
			wantErr:        true,
			errMsg:         "invalid timeout",
		},
		{
			name:           "invalid limit - zero value",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          0,
			wantErr:        true,
			errMsg:         "invalid limit",
		},
		{
			name:           "invalid limit - negative value",
			namespace:      "test-namespace",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
			wantErr:        true,
			errMsg:         "invalid limit",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateInput(testCase.namespace, testCase.timeoutSeconds, testCase.limit)
			if testCase.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServiceAPI_GetServiceByName(t *testing.T) {
	// Setup a service with desired characteristics
	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
				"env": "test",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "test-app",
			},
		},
	}

	// Create fake clientset with test service
	fakeClient := fake.NewClientset(testService)

	// Initialize service API
	serviceAPI := NewServiceAPI(fakeClient)

	// Test cases
	tests := []struct {
		name          string
		namespace     string
		serviceName   string
		wantErr       bool
		errorContains string
	}{
		{
			name:        "Successfully get service",
			namespace:   "test-namespace",
			serviceName: "test-service",
			wantErr:     false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			serviceName:   "test-service",
			wantErr:       true,
			errorContains: "invalid namespace",
		},
		{
			name:          "Empty service name",
			namespace:     "test-namespace",
			serviceName:   "",
			wantErr:       true,
			errorContains: "invalid service name",
		},
		{
			name:          "Service not found",
			namespace:     "test-namespace",
			serviceName:   "nonexistent-service",
			wantErr:       true,
			errorContains: "failed to get service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service, err := serviceAPI.GetServiceByName(ctx, tt.namespace, tt.serviceName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, service)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, service)
				assert.Equal(t, tt.serviceName, service.Name)
				assert.Equal(t, tt.namespace, service.Namespace)
				assert.Equal(t, "test-app", service.Labels["app"])
				assert.Equal(t, "test", service.Labels["env"])
				assert.Equal(t, corev1.ServiceTypeClusterIP, service.Spec.Type)
				assert.Equal(t, int32(80), service.Spec.Ports[0].Port)
				assert.Equal(t, "test-app", service.Spec.Selector["app"])
			}
		})
	}
}

func TestServiceAPI_ListServicesByLabel(t *testing.T) {
	// Setup test services
	testServices := []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-1",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "test-app",
					"environment": "production",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-2",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "test-app",
					"environment": "staging",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-service",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "other-app",
					"environment": "production",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
			},
		},
	}

	// Create fake clientset
	fakeClient := fake.NewClientset(testServices[0], testServices[1], testServices[2])

	// Initialize service API
	serviceAPI := NewServiceAPI(fakeClient)

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
			name:           "List services by app label",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"service-1", "service-2"},
			wantErr:        false,
		},
		{
			name:           "List services by environment label",
			namespace:      "test-namespace",
			labelSelector:  "environment=production",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"service-1", "other-service"},
			wantErr:        false,
		},
		{
			name:           "List services with multiple labels",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app,environment=production",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			expectedNames:  []string{"service-1"},
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
			labelSelector:  "",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid label selector",
		},
		{
			name:           "Invalid timeout",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			services, err := serviceAPI.ListServicesByLabel(ctx,
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
				assert.Len(t, services, testCase.expectedCount)

				// Check if all expected services are present
				if testCase.expectedCount > 0 {
					foundNames := make([]string, len(services))
					for i, service := range services {
						foundNames[i] = service.Name
					}

					for _, expectedName := range testCase.expectedNames {
						assert.Contains(t, foundNames, expectedName)
					}
				}
			}
		})
	}
}

func TestServiceAPI_ListServicesByField(t *testing.T) {
	// Setup test services
	testServices := []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-1",
				Namespace: "test-namespace",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-2",
				Namespace: "test-namespace",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-3",
				Namespace: "other-namespace",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		},
	}

	// Create fake clientset with test services
	fakeClient := fake.NewClientset(testServices[0], testServices[1], testServices[2])

	// Initialize service API
	serviceAPI := NewServiceAPI(fakeClient)

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
			name:           "List services by name",
			namespace:      "test-namespace",
			fieldSelector:  "metadata.name=service-1",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "List services by type",
			namespace:      "test-namespace",
			fieldSelector:  "spec.type=ClusterIP",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "No matching services",
			namespace:      "test-namespace",
			fieldSelector:  "metadata.name=nonexistent",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  0,
			wantErr:        false,
		},
		{
			name:           "Empty namespace",
			namespace:      "",
			fieldSelector:  "metadata.name=service-1",
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
			name:           "Invalid field selector format",
			namespace:      "test-namespace",
			fieldSelector:  "invalid-format",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid field selector",
		},
		{
			name:           "Invalid timeout",
			namespace:      "test-namespace",
			fieldSelector:  "metadata.name=service-1",
			timeoutSeconds: 500 * time.Millisecond,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			namespace:      "test-namespace",
			fieldSelector:  "metadata.name=service-1",
			timeoutSeconds: 2 * time.Second,
			limit:          0,
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			services, err := serviceAPI.ListServicesByField(ctx,
				tt.namespace,
				tt.fieldSelector,
				tt.timeoutSeconds,
				tt.limit,
			)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, services)
			} else {
				require.NoError(t, err)

				// Note: The fake client doesn't properly support field selectors
				// So we can't reliably test the number of results in this case
				// But we can verify we got a valid response and all services are from the correct namespace
				if services != nil {
					for _, service := range services {
						assert.Equal(t, tt.namespace, service.Namespace)
					}
				}
			}
		})
	}
}
