// Code generated by mockery v2.36.1. DO NOT EDIT.

package clworkosmock

import (
	context "context"
	url "net/url"

	mock "github.com/stretchr/testify/mock"

	usermanagement "github.com/workos/workos-go/v4/pkg/usermanagement"
)

// MockUserManagement is an autogenerated mock type for the UserManagement type
type MockUserManagement struct {
	mock.Mock
}

type MockUserManagement_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserManagement) EXPECT() *MockUserManagement_Expecter {
	return &MockUserManagement_Expecter{mock: &_m.Mock}
}

// AuthenticateWithCode provides a mock function with given fields: ctx, opts
func (_m *MockUserManagement) AuthenticateWithCode(ctx context.Context, opts usermanagement.AuthenticateWithCodeOpts) (usermanagement.AuthenticateResponse, error) {
	ret := _m.Called(ctx, opts)

	var r0 usermanagement.AuthenticateResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, usermanagement.AuthenticateWithCodeOpts) (usermanagement.AuthenticateResponse, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, usermanagement.AuthenticateWithCodeOpts) usermanagement.AuthenticateResponse); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(usermanagement.AuthenticateResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, usermanagement.AuthenticateWithCodeOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserManagement_AuthenticateWithCode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AuthenticateWithCode'
type MockUserManagement_AuthenticateWithCode_Call struct {
	*mock.Call
}

// AuthenticateWithCode is a helper method to define mock.On call
//   - ctx context.Context
//   - opts usermanagement.AuthenticateWithCodeOpts
func (_e *MockUserManagement_Expecter) AuthenticateWithCode(ctx interface{}, opts interface{}) *MockUserManagement_AuthenticateWithCode_Call {
	return &MockUserManagement_AuthenticateWithCode_Call{Call: _e.mock.On("AuthenticateWithCode", ctx, opts)}
}

func (_c *MockUserManagement_AuthenticateWithCode_Call) Run(run func(ctx context.Context, opts usermanagement.AuthenticateWithCodeOpts)) *MockUserManagement_AuthenticateWithCode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(usermanagement.AuthenticateWithCodeOpts))
	})
	return _c
}

func (_c *MockUserManagement_AuthenticateWithCode_Call) Return(_a0 usermanagement.AuthenticateResponse, _a1 error) *MockUserManagement_AuthenticateWithCode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserManagement_AuthenticateWithCode_Call) RunAndReturn(run func(context.Context, usermanagement.AuthenticateWithCodeOpts) (usermanagement.AuthenticateResponse, error)) *MockUserManagement_AuthenticateWithCode_Call {
	_c.Call.Return(run)
	return _c
}

// AuthenticateWithRefreshToken provides a mock function with given fields: ctx, opts
func (_m *MockUserManagement) AuthenticateWithRefreshToken(ctx context.Context, opts usermanagement.AuthenticateWithRefreshTokenOpts) (usermanagement.RefreshAuthenticationResponse, error) {
	ret := _m.Called(ctx, opts)

	var r0 usermanagement.RefreshAuthenticationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, usermanagement.AuthenticateWithRefreshTokenOpts) (usermanagement.RefreshAuthenticationResponse, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, usermanagement.AuthenticateWithRefreshTokenOpts) usermanagement.RefreshAuthenticationResponse); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(usermanagement.RefreshAuthenticationResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, usermanagement.AuthenticateWithRefreshTokenOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserManagement_AuthenticateWithRefreshToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AuthenticateWithRefreshToken'
type MockUserManagement_AuthenticateWithRefreshToken_Call struct {
	*mock.Call
}

// AuthenticateWithRefreshToken is a helper method to define mock.On call
//   - ctx context.Context
//   - opts usermanagement.AuthenticateWithRefreshTokenOpts
func (_e *MockUserManagement_Expecter) AuthenticateWithRefreshToken(ctx interface{}, opts interface{}) *MockUserManagement_AuthenticateWithRefreshToken_Call {
	return &MockUserManagement_AuthenticateWithRefreshToken_Call{Call: _e.mock.On("AuthenticateWithRefreshToken", ctx, opts)}
}

func (_c *MockUserManagement_AuthenticateWithRefreshToken_Call) Run(run func(ctx context.Context, opts usermanagement.AuthenticateWithRefreshTokenOpts)) *MockUserManagement_AuthenticateWithRefreshToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(usermanagement.AuthenticateWithRefreshTokenOpts))
	})
	return _c
}

func (_c *MockUserManagement_AuthenticateWithRefreshToken_Call) Return(_a0 usermanagement.RefreshAuthenticationResponse, _a1 error) *MockUserManagement_AuthenticateWithRefreshToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserManagement_AuthenticateWithRefreshToken_Call) RunAndReturn(run func(context.Context, usermanagement.AuthenticateWithRefreshTokenOpts) (usermanagement.RefreshAuthenticationResponse, error)) *MockUserManagement_AuthenticateWithRefreshToken_Call {
	_c.Call.Return(run)
	return _c
}

// GetAuthorizationURL provides a mock function with given fields: opts
func (_m *MockUserManagement) GetAuthorizationURL(opts usermanagement.GetAuthorizationURLOpts) (*url.URL, error) {
	ret := _m.Called(opts)

	var r0 *url.URL
	var r1 error
	if rf, ok := ret.Get(0).(func(usermanagement.GetAuthorizationURLOpts) (*url.URL, error)); ok {
		return rf(opts)
	}
	if rf, ok := ret.Get(0).(func(usermanagement.GetAuthorizationURLOpts) *url.URL); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*url.URL)
		}
	}

	if rf, ok := ret.Get(1).(func(usermanagement.GetAuthorizationURLOpts) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserManagement_GetAuthorizationURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAuthorizationURL'
type MockUserManagement_GetAuthorizationURL_Call struct {
	*mock.Call
}

// GetAuthorizationURL is a helper method to define mock.On call
//   - opts usermanagement.GetAuthorizationURLOpts
func (_e *MockUserManagement_Expecter) GetAuthorizationURL(opts interface{}) *MockUserManagement_GetAuthorizationURL_Call {
	return &MockUserManagement_GetAuthorizationURL_Call{Call: _e.mock.On("GetAuthorizationURL", opts)}
}

func (_c *MockUserManagement_GetAuthorizationURL_Call) Run(run func(opts usermanagement.GetAuthorizationURLOpts)) *MockUserManagement_GetAuthorizationURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(usermanagement.GetAuthorizationURLOpts))
	})
	return _c
}

func (_c *MockUserManagement_GetAuthorizationURL_Call) Return(_a0 *url.URL, _a1 error) *MockUserManagement_GetAuthorizationURL_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserManagement_GetAuthorizationURL_Call) RunAndReturn(run func(usermanagement.GetAuthorizationURLOpts) (*url.URL, error)) *MockUserManagement_GetAuthorizationURL_Call {
	_c.Call.Return(run)
	return _c
}

// GetLogoutURL provides a mock function with given fields: opts
func (_m *MockUserManagement) GetLogoutURL(opts usermanagement.GetLogoutURLOpts) (*url.URL, error) {
	ret := _m.Called(opts)

	var r0 *url.URL
	var r1 error
	if rf, ok := ret.Get(0).(func(usermanagement.GetLogoutURLOpts) (*url.URL, error)); ok {
		return rf(opts)
	}
	if rf, ok := ret.Get(0).(func(usermanagement.GetLogoutURLOpts) *url.URL); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*url.URL)
		}
	}

	if rf, ok := ret.Get(1).(func(usermanagement.GetLogoutURLOpts) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserManagement_GetLogoutURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLogoutURL'
type MockUserManagement_GetLogoutURL_Call struct {
	*mock.Call
}

// GetLogoutURL is a helper method to define mock.On call
//   - opts usermanagement.GetLogoutURLOpts
func (_e *MockUserManagement_Expecter) GetLogoutURL(opts interface{}) *MockUserManagement_GetLogoutURL_Call {
	return &MockUserManagement_GetLogoutURL_Call{Call: _e.mock.On("GetLogoutURL", opts)}
}

func (_c *MockUserManagement_GetLogoutURL_Call) Run(run func(opts usermanagement.GetLogoutURLOpts)) *MockUserManagement_GetLogoutURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(usermanagement.GetLogoutURLOpts))
	})
	return _c
}

func (_c *MockUserManagement_GetLogoutURL_Call) Return(_a0 *url.URL, _a1 error) *MockUserManagement_GetLogoutURL_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserManagement_GetLogoutURL_Call) RunAndReturn(run func(usermanagement.GetLogoutURLOpts) (*url.URL, error)) *MockUserManagement_GetLogoutURL_Call {
	_c.Call.Return(run)
	return _c
}

// GetUser provides a mock function with given fields: ctx, opts
func (_m *MockUserManagement) GetUser(ctx context.Context, opts usermanagement.GetUserOpts) (usermanagement.User, error) {
	ret := _m.Called(ctx, opts)

	var r0 usermanagement.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, usermanagement.GetUserOpts) (usermanagement.User, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, usermanagement.GetUserOpts) usermanagement.User); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(usermanagement.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, usermanagement.GetUserOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserManagement_GetUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUser'
type MockUserManagement_GetUser_Call struct {
	*mock.Call
}

// GetUser is a helper method to define mock.On call
//   - ctx context.Context
//   - opts usermanagement.GetUserOpts
func (_e *MockUserManagement_Expecter) GetUser(ctx interface{}, opts interface{}) *MockUserManagement_GetUser_Call {
	return &MockUserManagement_GetUser_Call{Call: _e.mock.On("GetUser", ctx, opts)}
}

func (_c *MockUserManagement_GetUser_Call) Run(run func(ctx context.Context, opts usermanagement.GetUserOpts)) *MockUserManagement_GetUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(usermanagement.GetUserOpts))
	})
	return _c
}

func (_c *MockUserManagement_GetUser_Call) Return(_a0 usermanagement.User, _a1 error) *MockUserManagement_GetUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserManagement_GetUser_Call) RunAndReturn(run func(context.Context, usermanagement.GetUserOpts) (usermanagement.User, error)) *MockUserManagement_GetUser_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUserManagement creates a new instance of MockUserManagement. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserManagement(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserManagement {
	mock := &MockUserManagement{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
