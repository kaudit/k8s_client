// Code generated by mockery v2.51.1. DO NOT EDIT.

package mocksapi

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

// MockPodAPI is an autogenerated mock type for the PodAPI type
type MockPodAPI struct {
	mock.Mock
}

type MockPodAPI_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPodAPI) EXPECT() *MockPodAPI_Expecter {
	return &MockPodAPI_Expecter{mock: &_m.Mock}
}

// GetPodByName provides a mock function with given fields: ctx, namespace, name
func (_m *MockPodAPI) GetPodByName(ctx context.Context, namespace string, name string) (*v1.Pod, error) {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for GetPodByName")
	}

	var r0 *v1.Pod
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*v1.Pod, error)); ok {
		return rf(ctx, namespace, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.Pod); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Pod)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPodAPI_GetPodByName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPodByName'
type MockPodAPI_GetPodByName_Call struct {
	*mock.Call
}

// GetPodByName is a helper method to define mock.On call
//   - ctx context.Context
//   - namespace string
//   - name string
func (_e *MockPodAPI_Expecter) GetPodByName(ctx interface{}, namespace interface{}, name interface{}) *MockPodAPI_GetPodByName_Call {
	return &MockPodAPI_GetPodByName_Call{Call: _e.mock.On("GetPodByName", ctx, namespace, name)}
}

func (_c *MockPodAPI_GetPodByName_Call) Run(run func(ctx context.Context, namespace string, name string)) *MockPodAPI_GetPodByName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockPodAPI_GetPodByName_Call) Return(_a0 *v1.Pod, _a1 error) *MockPodAPI_GetPodByName_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPodAPI_GetPodByName_Call) RunAndReturn(run func(context.Context, string, string) (*v1.Pod, error)) *MockPodAPI_GetPodByName_Call {
	_c.Call.Return(run)
	return _c
}

// ListPodsByField provides a mock function with given fields: ctx, namespace, fieldSelector
func (_m *MockPodAPI) ListPodsByField(ctx context.Context, namespace string, fieldSelector string) ([]v1.Pod, error) {
	ret := _m.Called(ctx, namespace, fieldSelector)

	if len(ret) == 0 {
		panic("no return value specified for ListPodsByField")
	}

	var r0 []v1.Pod
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]v1.Pod, error)); ok {
		return rf(ctx, namespace, fieldSelector)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []v1.Pod); ok {
		r0 = rf(ctx, namespace, fieldSelector)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Pod)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, fieldSelector)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPodAPI_ListPodsByField_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListPodsByField'
type MockPodAPI_ListPodsByField_Call struct {
	*mock.Call
}

// ListPodsByField is a helper method to define mock.On call
//   - ctx context.Context
//   - namespace string
//   - fieldSelector string
func (_e *MockPodAPI_Expecter) ListPodsByField(ctx interface{}, namespace interface{}, fieldSelector interface{}) *MockPodAPI_ListPodsByField_Call {
	return &MockPodAPI_ListPodsByField_Call{Call: _e.mock.On("ListPodsByField", ctx, namespace, fieldSelector)}
}

func (_c *MockPodAPI_ListPodsByField_Call) Run(run func(ctx context.Context, namespace string, fieldSelector string)) *MockPodAPI_ListPodsByField_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockPodAPI_ListPodsByField_Call) Return(_a0 []v1.Pod, _a1 error) *MockPodAPI_ListPodsByField_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPodAPI_ListPodsByField_Call) RunAndReturn(run func(context.Context, string, string) ([]v1.Pod, error)) *MockPodAPI_ListPodsByField_Call {
	_c.Call.Return(run)
	return _c
}

// ListPodsByLabel provides a mock function with given fields: ctx, namespace, labelSelector
func (_m *MockPodAPI) ListPodsByLabel(ctx context.Context, namespace string, labelSelector string) ([]v1.Pod, error) {
	ret := _m.Called(ctx, namespace, labelSelector)

	if len(ret) == 0 {
		panic("no return value specified for ListPodsByLabel")
	}

	var r0 []v1.Pod
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]v1.Pod, error)); ok {
		return rf(ctx, namespace, labelSelector)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []v1.Pod); ok {
		r0 = rf(ctx, namespace, labelSelector)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Pod)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, labelSelector)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPodAPI_ListPodsByLabel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListPodsByLabel'
type MockPodAPI_ListPodsByLabel_Call struct {
	*mock.Call
}

// ListPodsByLabel is a helper method to define mock.On call
//   - ctx context.Context
//   - namespace string
//   - labelSelector string
func (_e *MockPodAPI_Expecter) ListPodsByLabel(ctx interface{}, namespace interface{}, labelSelector interface{}) *MockPodAPI_ListPodsByLabel_Call {
	return &MockPodAPI_ListPodsByLabel_Call{Call: _e.mock.On("ListPodsByLabel", ctx, namespace, labelSelector)}
}

func (_c *MockPodAPI_ListPodsByLabel_Call) Run(run func(ctx context.Context, namespace string, labelSelector string)) *MockPodAPI_ListPodsByLabel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockPodAPI_ListPodsByLabel_Call) Return(_a0 []v1.Pod, _a1 error) *MockPodAPI_ListPodsByLabel_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPodAPI_ListPodsByLabel_Call) RunAndReturn(run func(context.Context, string, string) ([]v1.Pod, error)) *MockPodAPI_ListPodsByLabel_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockPodAPI creates a new instance of MockPodAPI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPodAPI(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPodAPI {
	mock := &MockPodAPI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
