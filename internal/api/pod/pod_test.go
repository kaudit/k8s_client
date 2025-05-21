package pod

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

func TestPodAPI_New(t *testing.T) {
	client := fake.NewClientset()
	api := NewPodAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*PodAPI)
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

func TestPodAPI_GetPodByName(t *testing.T) {
	// Setup a pod with desired characteristics
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
				"env": "test",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image:latest",
				},
			},
			NodeName: "test-node",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	// Create fake clientset with test pod
	fakeClient := fake.NewClientset(testPod)

	// Initialize pod API
	podAPI := NewPodAPI(fakeClient)

	// Test cases
	tests := []struct {
		name          string
		namespace     string
		podName       string
		wantErr       bool
		errorContains string
	}{
		{
			name:      "Successfully get pod",
			namespace: "test-namespace",
			podName:   "test-pod",
			wantErr:   false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			podName:       "test-pod",
			wantErr:       true,
			errorContains: "invalid namespace",
		},
		{
			name:          "Empty pod name",
			namespace:     "test-namespace",
			podName:       "",
			wantErr:       true,
			errorContains: "invalid pod name",
		},
		{
			name:          "Pod not found",
			namespace:     "test-namespace",
			podName:       "nonexistent-pod",
			wantErr:       true,
			errorContains: "failed to get pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pod, err := podAPI.GetPodByName(ctx, tt.namespace, tt.podName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, pod)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, pod)
				assert.Equal(t, tt.podName, pod.Name)
				assert.Equal(t, tt.namespace, pod.Namespace)
				assert.Equal(t, "test-app", pod.Labels["app"])
				assert.Equal(t, "test", pod.Labels["env"])
				assert.Equal(t, "test-container", pod.Spec.Containers[0].Name)
				assert.Equal(t, "test-image:latest", pod.Spec.Containers[0].Image)
				assert.Equal(t, corev1.PodRunning, pod.Status.Phase)
			}
		})
	}
}

func TestPodAPI_ListPodsByLabel(t *testing.T) {
	// Setup test pods
	testPods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-1",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "test-app",
					"environment": "production",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-2",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "test-app",
					"environment": "staging",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"app":         "other-app",
					"environment": "production",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	}

	// Create fake clientset
	fakeClient := fake.NewClientset(testPods[0], testPods[1], testPods[2])

	// Initialize pod API
	podAPI := NewPodAPI(fakeClient)

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
			name:           "List pods by app label",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"pod-1", "pod-2"},
			wantErr:        false,
		},
		{
			name:           "List pods by environment label",
			namespace:      "test-namespace",
			labelSelector:  "environment=production",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  2,
			expectedNames:  []string{"pod-1", "other-pod"},
			wantErr:        false,
		},
		{
			name:           "List pods with multiple labels",
			namespace:      "test-namespace",
			labelSelector:  "app=test-app,environment=production",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			expectedNames:  []string{"pod-1"},
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
			pods, err := podAPI.ListPodsByLabel(ctx,
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
				assert.Len(t, pods, testCase.expectedCount)

				// Check if all expected pods are present
				if testCase.expectedCount > 0 {
					foundNames := make([]string, len(pods))
					for i, pod := range pods {
						foundNames[i] = pod.Name
					}

					for _, expectedName := range testCase.expectedNames {
						assert.Contains(t, foundNames, expectedName)
					}
				}
			}
		})
	}
}

func TestPodAPI_ListPodsByField(t *testing.T) {
	// Setup test pods
	testPods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-1",
				Namespace: "test-namespace",
			},
			Spec: corev1.PodSpec{
				NodeName: "node-1",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-2",
				Namespace: "test-namespace",
			},
			Spec: corev1.PodSpec{
				NodeName: "node-2",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-3",
				Namespace: "other-namespace",
			},
			Spec: corev1.PodSpec{
				NodeName: "node-1",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	}

	// Create fake clientset with test pods
	fakeClient := fake.NewClientset(testPods[0], testPods[1], testPods[2])

	// Initialize pod API
	podAPI := NewPodAPI(fakeClient)

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
			name:           "List pods by node name",
			namespace:      "test-namespace",
			fieldSelector:  "spec.nodeName=node-1",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "List pods by status phase",
			namespace:      "test-namespace",
			fieldSelector:  "status.phase=Running",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name:           "No matching pods",
			namespace:      "test-namespace",
			fieldSelector:  "spec.nodeName=nonexistent",
			timeoutSeconds: 2 * time.Second,
			limit:          1,
			expectedCount:  0,
			wantErr:        false,
		},
		{
			name:           "Empty namespace",
			namespace:      "",
			fieldSelector:  "spec.nodeName=node-1",
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
			fieldSelector:  "spec.nodeName=node-1",
			timeoutSeconds: 500 * time.Millisecond,
			limit:          1,
			wantErr:        true,
			errorContains:  "invalid timeout",
		},
		{
			name:           "Invalid limit",
			namespace:      "test-namespace",
			fieldSelector:  "spec.nodeName=node-1",
			timeoutSeconds: 2 * time.Second,
			limit:          0,
			wantErr:        true,
			errorContains:  "invalid limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pods, err := podAPI.ListPodsByField(ctx,
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
				assert.Nil(t, pods)
			} else {
				require.NoError(t, err)

				// Note: The fake client doesn't properly support field selectors
				// So we can't reliably test the number of results in this case
				// But we can verify we got a valid response and all pods are from the correct namespace
				if pods != nil {
					for _, pod := range pods {
						assert.Equal(t, tt.namespace, pod.Namespace)
					}
				}
			}
		})
	}
}
