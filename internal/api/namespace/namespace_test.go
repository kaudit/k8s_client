package namespace

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

func TestNamespaceAPI_New(t *testing.T) {
	client := fake.NewClientset()
	api := NewNamespaceAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*NamespaceAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestValidateInput(t *testing.T) {
	testCases := []struct {
		name           string
		timeoutSeconds time.Duration
		limit          int64
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "Valid input",
			timeoutSeconds: 2 * time.Second,
			limit:          2,
			wantErr:        false,
		},
		{
			name:           "invalid timeout",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          2,
			wantErr:        true,
			errMsg:         "invalid timeout",
		},
		{
			name:           "invalid limit - zero value",
			timeoutSeconds: 2 * time.Second,
			limit:          0,
			wantErr:        true,
			errMsg:         "invalid limit",
		},
		{
			name:           "invalid limit - negative value",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
			wantErr:        true,
			errMsg:         "invalid limit",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateInput(testCase.timeoutSeconds, testCase.limit)
			if testCase.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNamespaceAPI_GetNamespaceByName(t *testing.T) {
	// Setup a namespace with desired characteristics
	testNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"env": "test",
				"app": "service-a",
			},
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}

	// Create fake clientset with test namespace
	fakeClient := fake.NewClientset(testNamespace)

	// Initialize namespace API
	namespaceAPI := NewNamespaceAPI(fakeClient)

	// Test cases
	tests := []struct {
		name          string
		namespaceName string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "Successfully get namespace",
			namespaceName: "test-namespace",
			wantErr:       false,
		},
		{
			name:          "Empty namespace name",
			namespaceName: "",
			wantErr:       true,
			errorContains: "invalid namespace name",
		},
		{
			name:          "Namespace not found",
			namespaceName: "nonexistent-namespace",
			wantErr:       true,
			errorContains: "failed to get namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			namespaceName, err := namespaceAPI.GetNamespaceByName(ctx, tt.namespaceName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Empty(t, namespaceName)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.namespaceName, namespaceName)
			}
		})
	}
}

func TestNamespaceAPI_ListNamespacesByLabel(t *testing.T) {
	// Setup test namespaces
	testNamespaces := []*corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-1",
				Labels: map[string]string{
					"env": "test",
					"app": "service-a",
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-2",
				Labels: map[string]string{
					"env": "test",
					"app": "service-b",
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prod-namespace",
				Labels: map[string]string{
					"env": "prod",
					"app": "service-a",
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
	}

	// Create fake clientset
	fakeClient := fake.NewClientset(testNamespaces[0], testNamespaces[1], testNamespaces[2])

	// Initialize namespace API
	namespaceAPI := NewNamespaceAPI(fakeClient)

	// Test cases
	tests := []struct {
		name           string
		labelSelector  string
		timeoutSeconds time.Duration
		limit          int64
		expectedCount  int
		expectedNames  []string
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "List namespaces by env label",
			labelSelector:  "env=test",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"test-namespace-1", "test-namespace-2"},
			wantErr:        false,
		},
		{
			name:           "List namespaces by app label",
			labelSelector:  "app=service-a",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"test-namespace-1", "prod-namespace"},
			wantErr:        false,
		},
		{
			name:           "List namespaces with multiple labels",
			labelSelector:  "env=test,app=service-a",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			expectedNames:  []string{"test-namespace-1"},
			wantErr:        false,
		},
		{
			name:           "No results",
			labelSelector:  "env=dev",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  0,
			expectedNames:  []string{},
			wantErr:        false,
		},
		{
			name:           "Empty label selector",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			labelSelector:  "",
			wantErr:        true,
			errorContains:  "invalid label selector",
		},
		{
			name:           "Invalid timeout",
			timeoutSeconds: 2 * time.Millisecond,
			limit:          1,
			labelSelector:  "env=test",
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			timeoutSeconds: 2 * time.Second,
			limit:          -1,
			labelSelector:  "env=test",
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			namespaceNames, err := namespaceAPI.ListNamespacesByLabel(ctx,
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
				assert.Len(t, namespaceNames, testCase.expectedCount)

				// Check if all expected namespaces are present
				if testCase.expectedCount > 0 {
					for _, expectedName := range testCase.expectedNames {
						assert.Contains(t, namespaceNames, expectedName)
					}
				}
			}
		})
	}
}

func TestNamespaceAPI_ListNamespacesByField(t *testing.T) {
	// Setup test namespaces
	testNamespaces := []*corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-1",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-2",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceTerminating,
			},
		},
	}

	// Create fake clientset with both test namespaces
	fakeClient := fake.NewClientset(testNamespaces[0], testNamespaces[1])

	// Initialize namespace API
	namespaceAPI := NewNamespaceAPI(fakeClient)

	// Test cases
	tests := []struct {
		name           string
		fieldSelector  string
		timeoutSeconds time.Duration
		limit          int64
		expectedCount  int
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "List namespaces by name",
			fieldSelector:  "metadata.name=test-namespace-1",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "List namespaces by status",
			fieldSelector:  "status.phase=Active",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "No matching namespaces",
			fieldSelector:  "metadata.name=nonexistent",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  0,
			wantErr:        false,
		},
		{
			name:           "Empty field selector",
			fieldSelector:  "",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid field selector",
		},
		{
			name:           "Invalid field selector format",
			fieldSelector:  "invalid-format",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid field selector",
		},
		{
			name:           "Invalid timeout",
			fieldSelector:  "metadata.name=test-namespace-1",
			timeoutSeconds: 500 * time.Millisecond,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			fieldSelector:  "metadata.name=test-namespace-1",
			timeoutSeconds: 2 * time.Second,
			limit:          0,
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			namespaceNames, err := namespaceAPI.ListNamespacesByField(ctx,
				tt.fieldSelector,
				tt.timeoutSeconds,
				tt.limit,
			)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, namespaceNames)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, namespaceNames)
			}
		})
	}
}
