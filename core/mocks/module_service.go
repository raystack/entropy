// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	json "encoding/json"

	mock "github.com/stretchr/testify/mock"

	module "github.com/odpf/entropy/core/module"

	resource "github.com/odpf/entropy/core/resource"
)

// ModuleService is an autogenerated mock type for the ModuleService type
type ModuleService struct {
	mock.Mock
}

type ModuleService_Expecter struct {
	mock *mock.Mock
}

func (_m *ModuleService) EXPECT() *ModuleService_Expecter {
	return &ModuleService_Expecter{mock: &_m.Mock}
}

// GetOutput provides a mock function with given fields: ctx, res
func (_m *ModuleService) GetOutput(ctx context.Context, res module.ExpandedResource) (json.RawMessage, error) {
	ret := _m.Called(ctx, res)

	var r0 json.RawMessage
	if rf, ok := ret.Get(0).(func(context.Context, module.ExpandedResource) json.RawMessage); ok {
		r0 = rf(ctx, res)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(json.RawMessage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.ExpandedResource) error); ok {
		r1 = rf(ctx, res)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ModuleService_GetOutput_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOutput'
type ModuleService_GetOutput_Call struct {
	*mock.Call
}

// GetOutput is a helper method to define mock.On call
//  - ctx context.Context
//  - res module.ExpandedResource
func (_e *ModuleService_Expecter) GetOutput(ctx interface{}, res interface{}) *ModuleService_GetOutput_Call {
	return &ModuleService_GetOutput_Call{Call: _e.mock.On("GetOutput", ctx, res)}
}

func (_c *ModuleService_GetOutput_Call) Run(run func(ctx context.Context, res module.ExpandedResource)) *ModuleService_GetOutput_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.ExpandedResource))
	})
	return _c
}

func (_c *ModuleService_GetOutput_Call) Return(_a0 json.RawMessage, _a1 error) *ModuleService_GetOutput_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// PlanAction provides a mock function with given fields: ctx, res, act
func (_m *ModuleService) PlanAction(ctx context.Context, res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	ret := _m.Called(ctx, res, act)

	var r0 *module.Plan
	if rf, ok := ret.Get(0).(func(context.Context, module.ExpandedResource, module.ActionRequest) *module.Plan); ok {
		r0 = rf(ctx, res, act)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*module.Plan)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.ExpandedResource, module.ActionRequest) error); ok {
		r1 = rf(ctx, res, act)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ModuleService_PlanAction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PlanAction'
type ModuleService_PlanAction_Call struct {
	*mock.Call
}

// PlanAction is a helper method to define mock.On call
//  - ctx context.Context
//  - res module.ExpandedResource
//  - act module.ActionRequest
func (_e *ModuleService_Expecter) PlanAction(ctx interface{}, res interface{}, act interface{}) *ModuleService_PlanAction_Call {
	return &ModuleService_PlanAction_Call{Call: _e.mock.On("PlanAction", ctx, res, act)}
}

func (_c *ModuleService_PlanAction_Call) Run(run func(ctx context.Context, res module.ExpandedResource, act module.ActionRequest)) *ModuleService_PlanAction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.ExpandedResource), args[2].(module.ActionRequest))
	})
	return _c
}

func (_c *ModuleService_PlanAction_Call) Return(_a0 *module.Plan, _a1 error) *ModuleService_PlanAction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// StreamLogs provides a mock function with given fields: ctx, res, filter
func (_m *ModuleService) StreamLogs(ctx context.Context, res module.ExpandedResource, filter map[string]string) (<-chan module.LogChunk, error) {
	ret := _m.Called(ctx, res, filter)

	var r0 <-chan module.LogChunk
	if rf, ok := ret.Get(0).(func(context.Context, module.ExpandedResource, map[string]string) <-chan module.LogChunk); ok {
		r0 = rf(ctx, res, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan module.LogChunk)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.ExpandedResource, map[string]string) error); ok {
		r1 = rf(ctx, res, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ModuleService_StreamLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StreamLogs'
type ModuleService_StreamLogs_Call struct {
	*mock.Call
}

// StreamLogs is a helper method to define mock.On call
//  - ctx context.Context
//  - res module.ExpandedResource
//  - filter map[string]string
func (_e *ModuleService_Expecter) StreamLogs(ctx interface{}, res interface{}, filter interface{}) *ModuleService_StreamLogs_Call {
	return &ModuleService_StreamLogs_Call{Call: _e.mock.On("StreamLogs", ctx, res, filter)}
}

func (_c *ModuleService_StreamLogs_Call) Run(run func(ctx context.Context, res module.ExpandedResource, filter map[string]string)) *ModuleService_StreamLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.ExpandedResource), args[2].(map[string]string))
	})
	return _c
}

func (_c *ModuleService_StreamLogs_Call) Return(_a0 <-chan module.LogChunk, _a1 error) *ModuleService_StreamLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// SyncState provides a mock function with given fields: ctx, res
func (_m *ModuleService) SyncState(ctx context.Context, res module.ExpandedResource) (*resource.State, error) {
	ret := _m.Called(ctx, res)

	var r0 *resource.State
	if rf, ok := ret.Get(0).(func(context.Context, module.ExpandedResource) *resource.State); ok {
		r0 = rf(ctx, res)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource.State)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, module.ExpandedResource) error); ok {
		r1 = rf(ctx, res)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ModuleService_SyncState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SyncState'
type ModuleService_SyncState_Call struct {
	*mock.Call
}

// SyncState is a helper method to define mock.On call
//  - ctx context.Context
//  - res module.ExpandedResource
func (_e *ModuleService_Expecter) SyncState(ctx interface{}, res interface{}) *ModuleService_SyncState_Call {
	return &ModuleService_SyncState_Call{Call: _e.mock.On("SyncState", ctx, res)}
}

func (_c *ModuleService_SyncState_Call) Run(run func(ctx context.Context, res module.ExpandedResource)) *ModuleService_SyncState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(module.ExpandedResource))
	})
	return _c
}

func (_c *ModuleService_SyncState_Call) Return(_a0 *resource.State, _a1 error) *ModuleService_SyncState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
