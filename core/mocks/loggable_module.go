// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	module "github.com/odpf/entropy/core/module"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/odpf/entropy/core/resource"
)

// LoggableModule is an autogenerated mock type for the Loggable type
type LoggableModule struct {
	mock.Mock
}

type LoggableModule_Expecter struct {
	mock *mock.Mock
}

func (_m *LoggableModule) EXPECT() *LoggableModule_Expecter {
	return &LoggableModule_Expecter{mock: &_m.Mock}
}

// Describe provides a mock function with given fields:
func (_m *LoggableModule) Describe() module.Desc {
	ret := _m.Called()

	var r0 module.Desc
	if rf, ok := ret.Get(0).(func() module.Desc); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(module.Desc)
	}

	return r0
}

// LoggableModule_Describe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Describe'
type LoggableModule_Describe_Call struct {
	*mock.Call
}

// Describe is a helper method to define mock.On call
func (_e *LoggableModule_Expecter) Describe() *LoggableModule_Describe_Call {
	return &LoggableModule_Describe_Call{Call: _e.mock.On("Describe")}
}

func (_c *LoggableModule_Describe_Call) Run(run func()) *LoggableModule_Describe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LoggableModule_Describe_Call) Return(_a0 module.Desc) *LoggableModule_Describe_Call {
	_c.Call.Return(_a0)
	return _c
}

// Log provides a mock function with given fields: ctx, spec, filter
func (_m *LoggableModule) Log(ctx context.Context, spec module.Spec, filter map[string]string) (<-chan module.LogChunk, error) {
	ret := _m.Called(ctx, spec, filter)

	var r0 <-chan module.LogChunk
	if rf, ok := ret.Get(0).(func(context.Context, module.Spec, map[string]string) <-chan module.LogChunk); ok {
		r0 = rf(ctx, spec, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan module.LogChunk)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.Spec, map[string]string) error); ok {
		r1 = rf(ctx, spec, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoggableModule_Log_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Log'
type LoggableModule_Log_Call struct {
	*mock.Call
}

// Log is a helper method to define mock.On call
//  - ctx context.Context
//  - spec module.Spec
//  - filter map[string]string
func (_e *LoggableModule_Expecter) Log(ctx interface{}, spec interface{}, filter interface{}) *LoggableModule_Log_Call {
	return &LoggableModule_Log_Call{Call: _e.mock.On("Log", ctx, spec, filter)}
}

func (_c *LoggableModule_Log_Call) Run(run func(ctx context.Context, spec module.Spec, filter map[string]string)) *LoggableModule_Log_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.Spec), args[2].(map[string]string))
	})
	return _c
}

func (_c *LoggableModule_Log_Call) Return(_a0 <-chan module.LogChunk, _a1 error) *LoggableModule_Log_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Plan provides a mock function with given fields: ctx, spec, act
func (_m *LoggableModule) Plan(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	ret := _m.Called(ctx, spec, act)

	var r0 *resource.Resource
	if rf, ok := ret.Get(0).(func(context.Context, module.Spec, module.ActionRequest) *resource.Resource); ok {
		r0 = rf(ctx, spec, act)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource.Resource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.Spec, module.ActionRequest) error); ok {
		r1 = rf(ctx, spec, act)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoggableModule_Plan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Plan'
type LoggableModule_Plan_Call struct {
	*mock.Call
}

// Plan is a helper method to define mock.On call
//  - ctx context.Context
//  - spec module.Spec
//  - act module.ActionRequest
func (_e *LoggableModule_Expecter) Plan(ctx interface{}, spec interface{}, act interface{}) *LoggableModule_Plan_Call {
	return &LoggableModule_Plan_Call{Call: _e.mock.On("Plan", ctx, spec, act)}
}

func (_c *LoggableModule_Plan_Call) Run(run func(ctx context.Context, spec module.Spec, act module.ActionRequest)) *LoggableModule_Plan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.Spec), args[2].(module.ActionRequest))
	})
	return _c
}

func (_c *LoggableModule_Plan_Call) Return(_a0 *resource.Resource, _a1 error) *LoggableModule_Plan_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Sync provides a mock function with given fields: ctx, spec
func (_m *LoggableModule) Sync(ctx context.Context, spec module.Spec) (*resource.Output, error) {
	ret := _m.Called(ctx, spec)

	var r0 *resource.Output
	if rf, ok := ret.Get(0).(func(context.Context, module.Spec) *resource.Output); ok {
		r0 = rf(ctx, spec)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource.Output)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.Spec) error); ok {
		r1 = rf(ctx, spec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoggableModule_Sync_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sync'
type LoggableModule_Sync_Call struct {
	*mock.Call
}

// Sync is a helper method to define mock.On call
//  - ctx context.Context
//  - spec module.Spec
func (_e *LoggableModule_Expecter) Sync(ctx interface{}, spec interface{}) *LoggableModule_Sync_Call {
	return &LoggableModule_Sync_Call{Call: _e.mock.On("Sync", ctx, spec)}
}

func (_c *LoggableModule_Sync_Call) Run(run func(ctx context.Context, spec module.Spec)) *LoggableModule_Sync_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.Spec))
	})
	return _c
}

func (_c *LoggableModule_Sync_Call) Return(_a0 *resource.Output, _a1 error) *LoggableModule_Sync_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
