// Code generated by mockery v2.36.1. DO NOT EDIT.

package clworkosmock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	organizations "github.com/workos/workos-go/v4/pkg/organizations"
)

// MockOrganizations is an autogenerated mock type for the Organizations type
type MockOrganizations struct {
	mock.Mock
}

type MockOrganizations_Expecter struct {
	mock *mock.Mock
}

func (_m *MockOrganizations) EXPECT() *MockOrganizations_Expecter {
	return &MockOrganizations_Expecter{mock: &_m.Mock}
}

// CreateOrganization provides a mock function with given fields: ctx, opts
func (_m *MockOrganizations) CreateOrganization(ctx context.Context, opts organizations.CreateOrganizationOpts) (organizations.Organization, error) {
	ret := _m.Called(ctx, opts)

	var r0 organizations.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organizations.CreateOrganizationOpts) (organizations.Organization, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organizations.CreateOrganizationOpts) organizations.Organization); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(organizations.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organizations.CreateOrganizationOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockOrganizations_CreateOrganization_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateOrganization'
type MockOrganizations_CreateOrganization_Call struct {
	*mock.Call
}

// CreateOrganization is a helper method to define mock.On call
//   - ctx context.Context
//   - opts organizations.CreateOrganizationOpts
func (_e *MockOrganizations_Expecter) CreateOrganization(ctx interface{}, opts interface{}) *MockOrganizations_CreateOrganization_Call {
	return &MockOrganizations_CreateOrganization_Call{Call: _e.mock.On("CreateOrganization", ctx, opts)}
}

func (_c *MockOrganizations_CreateOrganization_Call) Run(run func(ctx context.Context, opts organizations.CreateOrganizationOpts)) *MockOrganizations_CreateOrganization_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organizations.CreateOrganizationOpts))
	})
	return _c
}

func (_c *MockOrganizations_CreateOrganization_Call) Return(_a0 organizations.Organization, _a1 error) *MockOrganizations_CreateOrganization_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockOrganizations_CreateOrganization_Call) RunAndReturn(run func(context.Context, organizations.CreateOrganizationOpts) (organizations.Organization, error)) *MockOrganizations_CreateOrganization_Call {
	_c.Call.Return(run)
	return _c
}

// GetOrganization provides a mock function with given fields: ctx, opts
func (_m *MockOrganizations) GetOrganization(ctx context.Context, opts organizations.GetOrganizationOpts) (organizations.Organization, error) {
	ret := _m.Called(ctx, opts)

	var r0 organizations.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organizations.GetOrganizationOpts) (organizations.Organization, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organizations.GetOrganizationOpts) organizations.Organization); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(organizations.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organizations.GetOrganizationOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockOrganizations_GetOrganization_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOrganization'
type MockOrganizations_GetOrganization_Call struct {
	*mock.Call
}

// GetOrganization is a helper method to define mock.On call
//   - ctx context.Context
//   - opts organizations.GetOrganizationOpts
func (_e *MockOrganizations_Expecter) GetOrganization(ctx interface{}, opts interface{}) *MockOrganizations_GetOrganization_Call {
	return &MockOrganizations_GetOrganization_Call{Call: _e.mock.On("GetOrganization", ctx, opts)}
}

func (_c *MockOrganizations_GetOrganization_Call) Run(run func(ctx context.Context, opts organizations.GetOrganizationOpts)) *MockOrganizations_GetOrganization_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organizations.GetOrganizationOpts))
	})
	return _c
}

func (_c *MockOrganizations_GetOrganization_Call) Return(_a0 organizations.Organization, _a1 error) *MockOrganizations_GetOrganization_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockOrganizations_GetOrganization_Call) RunAndReturn(run func(context.Context, organizations.GetOrganizationOpts) (organizations.Organization, error)) *MockOrganizations_GetOrganization_Call {
	_c.Call.Return(run)
	return _c
}

// ListOrganizations provides a mock function with given fields: ctx, opts
func (_m *MockOrganizations) ListOrganizations(ctx context.Context, opts organizations.ListOrganizationsOpts) (organizations.ListOrganizationsResponse, error) {
	ret := _m.Called(ctx, opts)

	var r0 organizations.ListOrganizationsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organizations.ListOrganizationsOpts) (organizations.ListOrganizationsResponse, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organizations.ListOrganizationsOpts) organizations.ListOrganizationsResponse); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(organizations.ListOrganizationsResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organizations.ListOrganizationsOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockOrganizations_ListOrganizations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListOrganizations'
type MockOrganizations_ListOrganizations_Call struct {
	*mock.Call
}

// ListOrganizations is a helper method to define mock.On call
//   - ctx context.Context
//   - opts organizations.ListOrganizationsOpts
func (_e *MockOrganizations_Expecter) ListOrganizations(ctx interface{}, opts interface{}) *MockOrganizations_ListOrganizations_Call {
	return &MockOrganizations_ListOrganizations_Call{Call: _e.mock.On("ListOrganizations", ctx, opts)}
}

func (_c *MockOrganizations_ListOrganizations_Call) Run(run func(ctx context.Context, opts organizations.ListOrganizationsOpts)) *MockOrganizations_ListOrganizations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organizations.ListOrganizationsOpts))
	})
	return _c
}

func (_c *MockOrganizations_ListOrganizations_Call) Return(_a0 organizations.ListOrganizationsResponse, _a1 error) *MockOrganizations_ListOrganizations_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockOrganizations_ListOrganizations_Call) RunAndReturn(run func(context.Context, organizations.ListOrganizationsOpts) (organizations.ListOrganizationsResponse, error)) *MockOrganizations_ListOrganizations_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateOrganization provides a mock function with given fields: ctx, opts
func (_m *MockOrganizations) UpdateOrganization(ctx context.Context, opts organizations.UpdateOrganizationOpts) (organizations.Organization, error) {
	ret := _m.Called(ctx, opts)

	var r0 organizations.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organizations.UpdateOrganizationOpts) (organizations.Organization, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organizations.UpdateOrganizationOpts) organizations.Organization); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Get(0).(organizations.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organizations.UpdateOrganizationOpts) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockOrganizations_UpdateOrganization_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateOrganization'
type MockOrganizations_UpdateOrganization_Call struct {
	*mock.Call
}

// UpdateOrganization is a helper method to define mock.On call
//   - ctx context.Context
//   - opts organizations.UpdateOrganizationOpts
func (_e *MockOrganizations_Expecter) UpdateOrganization(ctx interface{}, opts interface{}) *MockOrganizations_UpdateOrganization_Call {
	return &MockOrganizations_UpdateOrganization_Call{Call: _e.mock.On("UpdateOrganization", ctx, opts)}
}

func (_c *MockOrganizations_UpdateOrganization_Call) Run(run func(ctx context.Context, opts organizations.UpdateOrganizationOpts)) *MockOrganizations_UpdateOrganization_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organizations.UpdateOrganizationOpts))
	})
	return _c
}

func (_c *MockOrganizations_UpdateOrganization_Call) Return(_a0 organizations.Organization, _a1 error) *MockOrganizations_UpdateOrganization_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockOrganizations_UpdateOrganization_Call) RunAndReturn(run func(context.Context, organizations.UpdateOrganizationOpts) (organizations.Organization, error)) *MockOrganizations_UpdateOrganization_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockOrganizations creates a new instance of MockOrganizations. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockOrganizations(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOrganizations {
	mock := &MockOrganizations{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
