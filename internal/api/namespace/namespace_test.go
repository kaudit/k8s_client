package namespace

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

func TestNewNamespaceAPI(t *testing.T) {
	client := fake.NewClientset()
	api := NewNamespaceAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*NamespaceAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestNamespaceAPI_GetNamespaceByName(t *testing.T) {
	// Create test namespaces
	testNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"env": "test",
			},
		},
	}

	tests := []struct {
		name    string
		nsName  string
		objects []runtime.Object
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Namespace exists",
			nsName:  "test-namespace",
			objects: []runtime.Object{testNS},
			wantErr: false,
		},
		{
			name:    "Namespace not found",
			nsName:  "nonexistent-namespace",
			objects: []runtime.Object{testNS},
			wantErr: true,
			errMsg:  "failed to get namespace",
		},
		{
			name:    "Empty namespace name",
			nsName:  "",
			objects: []runtime.Object{testNS},
			wantErr: true,
			errMsg:  "failed to validate namespace name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			nsAPI := NewNamespaceAPI(client)

			// Execute the method
			ctx := context.Background()
			ns, err := nsAPI.GetNamespaceByName(ctx, tt.nsName)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, ns)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, ns)
				assert.Equal(t, tt.nsName, ns.Name)
			}
		})
	}
}

func TestNamespaceAPI_ListNamespacesByLabel(t *testing.T) {
	// Create test namespaces
	testNS1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace-1",
			Labels: map[string]string{
				"env": "test",
				"app": "service-a",
			},
		},
	}

	testNS2 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace-2",
			Labels: map[string]string{
				"env": "test",
				"app": "service-b",
			},
		},
	}

	testNS3 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod-namespace",
			Labels: map[string]string{
				"env": "prod",
				"app": "service-a",
			},
		},
	}

	tests := []struct {
		name          string
		labelSelector string
		objects       []runtime.Object
		wantCount     int
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Filter by environment - test",
			labelSelector: "env=test",
			objects:       []runtime.Object{testNS1, testNS2, testNS3},
			wantCount:     2,
			wantErr:       false,
		},
		{
			name:          "Filter by environment and app",
			labelSelector: "env=test,app=service-a",
			objects:       []runtime.Object{testNS1, testNS2, testNS3},
			wantCount:     1,
			wantErr:       false,
		},
		{
			name:          "No matching namespaces",
			labelSelector: "env=dev",
			objects:       []runtime.Object{testNS1, testNS2, testNS3},
			wantCount:     0,
			wantErr:       false,
		},
		{
			name:          "Empty label selector",
			labelSelector: "",
			objects:       []runtime.Object{testNS1, testNS2},
			wantErr:       true,
			errMsg:        "failed to validate label selector",
		},
		{
			name:          "Invalid label selector format",
			labelSelector: "invalid@label",
			objects:       []runtime.Object{testNS1, testNS2},
			wantErr:       true,
			errMsg:        "failed to validate label selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			nsAPI := NewNamespaceAPI(client)

			// Execute the method
			ctx := context.Background()
			namespaces, err := nsAPI.ListNamespacesByLabel(ctx, tt.labelSelector)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, namespaces)
			} else {
				require.NoError(t, err)

				// For "No matching namespaces" case, handle both nil and empty slice
				if tt.wantCount == 0 {
					if namespaces != nil {
						assert.Empty(t, namespaces)
					}
				} else {
					assert.NotNil(t, namespaces)
					assert.Len(t, namespaces, tt.wantCount)
				}
			}
		})
	}
}

func TestNamespaceAPI_ListNamespacesByField(t *testing.T) {
	// Create test namespaces
	testNS1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace-1",
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}

	testNS2 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace-2",
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}

	tests := []struct {
		name          string
		fieldSelector string
		objects       []runtime.Object
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Empty field selector",
			fieldSelector: "",
			objects:       []runtime.Object{testNS1, testNS2},
			wantErr:       true,
			errMsg:        "failed to validate field selector",
		},
		{
			name:          "Invalid field selector format",
			fieldSelector: "invalid@field",
			objects:       []runtime.Object{testNS1, testNS2},
			wantErr:       true,
			errMsg:        "failed to validate field selector",
		},
		{
			name:          "Valid field selector syntax",
			fieldSelector: "metadata.name=test-namespace-1",
			objects:       []runtime.Object{testNS1, testNS2},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			nsAPI := NewNamespaceAPI(client)

			// Execute the method
			ctx := context.Background()
			namespaces, err := nsAPI.ListNamespacesByField(ctx, tt.fieldSelector)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, namespaces)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, namespaces)
				// We don't verify the count of results because fake clients
				// don't properly implement field selection
			}
		})
	}
}
