// Code generated by mockery v2.52.2. DO NOT EDIT.

package clconnectmock

import (
	context "context"

	client "github.com/ory/client-go"

	mock "github.com/stretchr/testify/mock"
)

// MockOry is an autogenerated mock type for the Ory type
type MockOry struct {
	mock.Mock
}

type MockOry_Expecter struct {
	mock *mock.Mock
}

func (_m *MockOry) EXPECT() *MockOry_Expecter {
	return &MockOry_Expecter{mock: &_m.Mock}
}

// Authenticate provides a mock function with given fields: ctx, cookie, allowAnonymous
func (_m *MockOry) Authenticate(ctx context.Context, cookie string, allowAnonymous bool) (*client.Session, error) {
	ret := _m.Called(ctx, cookie, allowAnonymous)

	if len(ret) == 0 {
		panic("no return value specified for Authenticate")
	}

	var r0 *client.Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) (*client.Session, error)); ok {
		return rf(ctx, cookie, allowAnonymous)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) *client.Session); ok {
		r0 = rf(ctx, cookie, allowAnonymous)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, bool) error); ok {
		r1 = rf(ctx, cookie, allowAnonymous)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockOry_Authenticate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Authenticate'
type MockOry_Authenticate_Call struct {
	*mock.Call
}

// Authenticate is a helper method to define mock.On call
//   - ctx context.Context
//   - cookie string
//   - allowAnonymous bool
func (_e *MockOry_Expecter) Authenticate(ctx interface{}, cookie interface{}, allowAnonymous interface{}) *MockOry_Authenticate_Call {
	return &MockOry_Authenticate_Call{Call: _e.mock.On("Authenticate", ctx, cookie, allowAnonymous)}
}

func (_c *MockOry_Authenticate_Call) Run(run func(ctx context.Context, cookie string, allowAnonymous bool)) *MockOry_Authenticate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(bool))
	})
	return _c
}

func (_c *MockOry_Authenticate_Call) Return(_a0 *client.Session, _a1 error) *MockOry_Authenticate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockOry_Authenticate_Call) RunAndReturn(run func(context.Context, string, bool) (*client.Session, error)) *MockOry_Authenticate_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockOry creates a new instance of MockOry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockOry(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOry {
	mock := &MockOry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
