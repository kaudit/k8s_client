package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewServiceAPI(t *testing.T) {
	client := fake.NewClientset()
	api := NewServiceAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*ServiceAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestServiceAPI_GetServiceByName(t *testing.T) {
	// Create test services
	testSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}

	// Setup tests
	tests := []struct {
		name      string
		namespace string
		svcName   string
		objects   []runtime.Object
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Service exists",
			namespace: "test-namespace",
			svcName:   "test-service",
			objects:   []runtime.Object{testSvc},
			wantErr:   false,
		},
		{
			name:      "Service not found",
			namespace: "test-namespace",
			svcName:   "nonexistent-service",
			objects:   []runtime.Object{testSvc},
			wantErr:   true,
			errMsg:    "failed to get service",
		},
		{
			name:      "Empty namespace",
			namespace: "",
			svcName:   "test-service",
			objects:   []runtime.Object{testSvc},
			wantErr:   true,
			errMsg:    "invalid namespace",
		},
		{
			name:      "Empty service name",
			namespace: "test-namespace",
			svcName:   "",
			objects:   []runtime.Object{testSvc},
			wantErr:   true,
			errMsg:    "invalid service name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			serviceAPI := NewServiceAPI(client)

			// Execute the method
			ctx := context.Background()
			svc, err := serviceAPI.GetServiceByName(ctx, tt.namespace, tt.svcName)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, svc)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, svc)
				assert.Equal(t, tt.svcName, svc.Name)
				assert.Equal(t, tt.namespace, svc.Namespace)
			}
		})
	}
}

func TestServiceAPI_ListServicesByLabel(t *testing.T) {
	// Create test services
	testSvc1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-1",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
				"env": "test",
			},
		},
	}
	testSvc2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-2",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
				"env": "prod",
			},
		},
	}

	// Setup tests
	tests := []struct {
		name          string
		namespace     string
		labelSelector string
		objects       []runtime.Object
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Valid input parameters",
			namespace:     "test-namespace",
			labelSelector: "app=test-app",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			labelSelector: "app=test-app",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       true,
			errMsg:        "invalid namespace",
		},
		{
			name:          "Empty label selector",
			namespace:     "test-namespace",
			labelSelector: "",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       true,
			errMsg:        "invalid label selector",
		},
		{
			name:          "Invalid label selector format",
			namespace:     "test-namespace",
			labelSelector: "invalid@label",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       true,
			errMsg:        "invalid label selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			serviceAPI := NewServiceAPI(client)

			// Execute the method
			ctx := context.Background()
			services, err := serviceAPI.ListServicesByLabel(ctx, tt.namespace, tt.labelSelector)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, services)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, services)
			}
		})
	}
}

func TestServiceAPI_ListServicesByField(t *testing.T) {
	// Create test services
	testSvc1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-1",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	testSvc2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-2",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
		},
	}

	// Setup tests
	tests := []struct {
		name          string
		namespace     string
		fieldSelector string
		objects       []runtime.Object
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Valid input parameters",
			namespace:     "test-namespace",
			fieldSelector: "metadata.name=service-1",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			fieldSelector: "metadata.name=service-1",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       true,
			errMsg:        "invalid namespace",
		},
		{
			name:          "Empty field selector",
			namespace:     "test-namespace",
			fieldSelector: "",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       true,
			errMsg:        "invalid field selector",
		},
		{
			name:          "Invalid field selector format",
			namespace:     "test-namespace",
			fieldSelector: "invalid-format",
			objects:       []runtime.Object{testSvc1, testSvc2},
			wantErr:       true,
			errMsg:        "invalid field selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			serviceAPI := NewServiceAPI(client)

			// Execute the method
			ctx := context.Background()
			services, err := serviceAPI.ListServicesByField(ctx, tt.namespace, tt.fieldSelector)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, services)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, services)
			}
		})
	}
}
