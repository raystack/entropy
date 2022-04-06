// Code generated by mockery v2.10.1. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/odpf/entropy/domain"
	mock "github.com/stretchr/testify/mock"
)

// Module is an autogenerated mock type for the Module type
type Module struct {
	mock.Mock
}

type Module_Expecter struct {
	mock *mock.Mock
}

func (_m *Module) EXPECT() *Module_Expecter {
	return &Module_Expecter{mock: &_m.Mock}
}

// Act provides a mock function with given fields: r, action, params
func (_m *Module) Act(r *domain.Resource, action string, params map[string]interface{}) (map[string]interface{}, error) {
	ret := _m.Called(r, action, params)

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func(*domain.Resource, string, map[string]interface{}) map[string]interface{}); ok {
		r0 = rf(r, action, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*domain.Resource, string, map[string]interface{}) error); ok {
		r1 = rf(r, action, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Module_Act_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Act'
type Module_Act_Call struct {
	*mock.Call
}

// Act is a helper method to define mock.On call
//  - r *domain.Resource
//  - action string
//  - params map[string]interface{}
func (_e *Module_Expecter) Act(r interface{}, action interface{}, params interface{}) *Module_Act_Call {
	return &Module_Act_Call{Call: _e.mock.On("Act", r, action, params)}
}

func (_c *Module_Act_Call) Run(run func(r *domain.Resource, action string, params map[string]interface{})) *Module_Act_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*domain.Resource), args[1].(string), args[2].(map[string]interface{}))
	})
	return _c
}

func (_c *Module_Act_Call) Return(_a0 map[string]interface{}, _a1 error) *Module_Act_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Apply provides a mock function with given fields: r
func (_m *Module) Apply(r *domain.Resource) (domain.ResourceStatus, error) {
	ret := _m.Called(r)

	var r0 domain.ResourceStatus
	if rf, ok := ret.Get(0).(func(*domain.Resource) domain.ResourceStatus); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Get(0).(domain.ResourceStatus)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*domain.Resource) error); ok {
		r1 = rf(r)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Module_Apply_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Apply'
type Module_Apply_Call struct {
	*mock.Call
}

// Apply is a helper method to define mock.On call
//  - r *domain.Resource
func (_e *Module_Expecter) Apply(r interface{}) *Module_Apply_Call {
	return &Module_Apply_Call{Call: _e.mock.On("Apply", r)}
}

func (_c *Module_Apply_Call) Run(run func(r *domain.Resource)) *Module_Apply_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*domain.Resource))
	})
	return _c
}

func (_c *Module_Apply_Call) Return(_a0 domain.ResourceStatus, _a1 error) *Module_Apply_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// ID provides a mock function with given fields:
func (_m *Module) ID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Module_ID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ID'
type Module_ID_Call struct {
	*mock.Call
}

// ID is a helper method to define mock.On call
func (_e *Module_Expecter) ID() *Module_ID_Call {
	return &Module_ID_Call{Call: _e.mock.On("ID")}
}

func (_c *Module_ID_Call) Run(run func()) *Module_ID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Module_ID_Call) Return(_a0 string) *Module_ID_Call {
	_c.Call.Return(_a0)
	return _c
}

// Log provides a mock function with given fields: ctx, r, filter
func (_m *Module) Log(ctx context.Context, r *domain.Resource, filter map[string]string) (<-chan domain.LogChunk, error) {
	ret := _m.Called(ctx, r, filter)

	var r0 <-chan domain.LogChunk
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Resource, map[string]string) <-chan domain.LogChunk); ok {
		r0 = rf(ctx, r, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan domain.LogChunk)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *domain.Resource, map[string]string) error); ok {
		r1 = rf(ctx, r, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Module_Log_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Log'
type Module_Log_Call struct {
	*mock.Call
}

// Log is a helper method to define mock.On call
//  - ctx context.Context
//  - r *domain.Resource
//  - filter map[string]string
func (_e *Module_Expecter) Log(ctx interface{}, r interface{}, filter interface{}) *Module_Log_Call {
	return &Module_Log_Call{Call: _e.mock.On("Log", ctx, r, filter)}
}

func (_c *Module_Log_Call) Run(run func(ctx context.Context, r *domain.Resource, filter map[string]string)) *Module_Log_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*domain.Resource), args[2].(map[string]string))
	})
	return _c
}

func (_c *Module_Log_Call) Return(_a0 <-chan domain.LogChunk, _a1 error) *Module_Log_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Validate provides a mock function with given fields: r
func (_m *Module) Validate(r *domain.Resource) error {
	ret := _m.Called(r)

	var r0 error
	if rf, ok := ret.Get(0).(func(*domain.Resource) error); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Module_Validate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Validate'
type Module_Validate_Call struct {
	*mock.Call
}

// Validate is a helper method to define mock.On call
//  - r *domain.Resource
func (_e *Module_Expecter) Validate(r interface{}) *Module_Validate_Call {
	return &Module_Validate_Call{Call: _e.mock.On("Validate", r)}
}

func (_c *Module_Validate_Call) Run(run func(r *domain.Resource)) *Module_Validate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*domain.Resource))
	})
	return _c
}

func (_c *Module_Validate_Call) Return(_a0 error) *Module_Validate_Call {
	_c.Call.Return(_a0)
	return _c
}
