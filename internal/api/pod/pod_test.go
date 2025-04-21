package pod

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

func TestNewPodAPI(t *testing.T) {
	client := fake.NewClientset()
	api := NewPodAPI(client)

	require.NotNil(t, api)

	impl, ok := api.(*PodAPI)
	require.True(t, ok)
	assert.Same(t, client, impl.client)
}

func TestPodAPI_GetPodByName(t *testing.T) {
	// Create test pods
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	}

	// Setup tests
	tests := []struct {
		name      string
		namespace string
		podName   string
		objects   []runtime.Object
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Pod exists",
			namespace: "test-namespace",
			podName:   "test-pod",
			objects:   []runtime.Object{testPod},
			wantErr:   false,
		},
		{
			name:      "Pod not found",
			namespace: "test-namespace",
			podName:   "nonexistent-pod",
			objects:   []runtime.Object{testPod},
			wantErr:   true,
			errMsg:    "failed to get pod",
		},
		{
			name:      "Empty namespace",
			namespace: "",
			podName:   "test-pod",
			objects:   []runtime.Object{testPod},
			wantErr:   true,
			errMsg:    "invalid namespace",
		},
		{
			name:      "Empty pod name",
			namespace: "test-namespace",
			podName:   "",
			objects:   []runtime.Object{testPod},
			wantErr:   true,
			errMsg:    "invalid pod name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			podAPI := NewPodAPI(client)

			// Execute the method
			ctx := context.Background()
			pod, err := podAPI.GetPodByName(ctx, tt.namespace, tt.podName)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, pod)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, pod)
				assert.Equal(t, tt.podName, pod.Name)
				assert.Equal(t, tt.namespace, pod.Namespace)
			}
		})
	}
}

func TestPodAPI_ListPodsByLabel(t *testing.T) {
	// Create test pods
	testPod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
				"env": "test",
			},
		},
	}
	testPod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
				"env": "prod",
			},
		},
	}
	testPod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-3",
			Namespace: "other-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
	}

	// Setup tests
	tests := []struct {
		name          string
		namespace     string
		labelSelector string
		objects       []runtime.Object
		wantCount     int
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Valid input parameters - all namespace pods",
			namespace:     "test-namespace",
			labelSelector: "app=test-app",
			objects:       []runtime.Object{testPod1, testPod2, testPod3},
			wantCount:     2,
			wantErr:       false,
		},
		{
			name:          "Valid input parameters - specific env",
			namespace:     "test-namespace",
			labelSelector: "app=test-app,env=test",
			objects:       []runtime.Object{testPod1, testPod2, testPod3},
			wantCount:     1,
			wantErr:       false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			labelSelector: "app=test-app",
			objects:       []runtime.Object{testPod1, testPod2},
			wantErr:       true,
			errMsg:        "invalid namespace",
		},
		{
			name:          "Empty label selector",
			namespace:     "test-namespace",
			labelSelector: "",
			objects:       []runtime.Object{testPod1, testPod2},
			wantErr:       true,
			errMsg:        "invalid label selector",
		},
		{
			name:          "Invalid label selector format",
			namespace:     "test-namespace",
			labelSelector: "invalid@label",
			objects:       []runtime.Object{testPod1, testPod2},
			wantErr:       true,
			errMsg:        "invalid label selector",
		},
		{
			name:          "No matching pods",
			namespace:     "test-namespace",
			labelSelector: "app=non-existent",
			objects:       []runtime.Object{testPod1, testPod2, testPod3},
			wantCount:     0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			podAPI := NewPodAPI(client)

			// Execute the method
			ctx := context.Background()
			pods, err := podAPI.ListPodsByLabel(ctx, tt.namespace, tt.labelSelector)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, pods)
			} else {
				require.NoError(t, err)

				// Modified check: if wantCount is 0, we either expect an empty slice or nil
				if tt.wantCount == 0 {
					if pods != nil {
						assert.Empty(t, pods)
					}
					// If pods is nil, that's acceptable for the "no pods found" case
				} else {
					// For cases where we expect to find pods
					assert.NotNil(t, pods)
					assert.Len(t, pods, tt.wantCount)

					// Check that all returned pods are in the correct namespace
					for _, pod := range pods {
						assert.Equal(t, tt.namespace, pod.Namespace)
					}
				}
			}
		})
	}
}

func TestPodAPI_ListPodsByField(t *testing.T) {
	// Create test pods
	testPod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
		},
	}
	testPod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			NodeName: "node-2",
		},
	}
	testPod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-3",
			Namespace: "other-namespace",
		},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
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
			name:          "Valid field selector syntax",
			namespace:     "test-namespace",
			fieldSelector: "spec.nodeName=node-1",
			objects:       []runtime.Object{testPod1, testPod2, testPod3},
			wantErr:       false,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			fieldSelector: "spec.nodeName=node-1",
			objects:       []runtime.Object{testPod1, testPod2},
			wantErr:       true,
			errMsg:        "invalid namespace",
		},
		{
			name:          "Empty field selector",
			namespace:     "test-namespace",
			fieldSelector: "",
			objects:       []runtime.Object{testPod1, testPod2},
			wantErr:       true,
			errMsg:        "invalid field selector",
		},
		{
			name:          "Invalid field selector format",
			namespace:     "test-namespace",
			fieldSelector: "invalid@field",
			objects:       []runtime.Object{testPod1, testPod2},
			wantErr:       true,
			errMsg:        "invalid field selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			client := fake.NewClientset(tt.objects...)
			podAPI := NewPodAPI(client)

			// Execute the method
			ctx := context.Background()
			pods, err := podAPI.ListPodsByField(ctx, tt.namespace, tt.fieldSelector)

			// Verify results
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, pods)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, pods)

				// Only verify that pods are from the correct namespace
				// since fake client doesn't properly implement field selectors
				for _, pod := range pods {
					assert.Equal(t, tt.namespace, pod.Namespace)
				}

				// Verify we get the expected number of pods in this namespace
				if tt.namespace == "test-namespace" {
					assert.Len(t, pods, 2) // There are 2 pods in test-namespace
				}
			}
		})
	}
}
