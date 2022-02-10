// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	domain "github.com/odpf/entropy/domain"
	mock "github.com/stretchr/testify/mock"
)

// ResourceRepository is an autogenerated mock type for the ResourceRepository type
type ResourceRepository struct {
	mock.Mock
}

type ResourceRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *ResourceRepository) EXPECT() *ResourceRepository_Expecter {
	return &ResourceRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: r
func (_m *ResourceRepository) Create(r *domain.Resource) error {
	ret := _m.Called(r)

	var r0 error
	if rf, ok := ret.Get(0).(func(*domain.Resource) error); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResourceRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ResourceRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//  - r *domain.Resource
func (_e *ResourceRepository_Expecter) Create(r interface{}) *ResourceRepository_Create_Call {
	return &ResourceRepository_Create_Call{Call: _e.mock.On("Create", r)}
}

func (_c *ResourceRepository_Create_Call) Run(run func(r *domain.Resource)) *ResourceRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*domain.Resource))
	})
	return _c
}

func (_c *ResourceRepository_Create_Call) Return(_a0 error) *ResourceRepository_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetByURN provides a mock function with given fields: urn
func (_m *ResourceRepository) GetByURN(urn string) (*domain.Resource, error) {
	ret := _m.Called(urn)

	var r0 *domain.Resource
	if rf, ok := ret.Get(0).(func(string) *domain.Resource); ok {
		r0 = rf(urn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Resource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(urn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResourceRepository_GetByURN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByURN'
type ResourceRepository_GetByURN_Call struct {
	*mock.Call
}

// GetByURN is a helper method to define mock.On call
//  - urn string
func (_e *ResourceRepository_Expecter) GetByURN(urn interface{}) *ResourceRepository_GetByURN_Call {
	return &ResourceRepository_GetByURN_Call{Call: _e.mock.On("GetByURN", urn)}
}

func (_c *ResourceRepository_GetByURN_Call) Run(run func(urn string)) *ResourceRepository_GetByURN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ResourceRepository_GetByURN_Call) Return(_a0 *domain.Resource, _a1 error) *ResourceRepository_GetByURN_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// List provides a mock function with given fields: parent, kind
func (_m *ResourceRepository) List(parent string, kind string) ([]*domain.Resource, error) {
	ret := _m.Called(parent, kind)

	var r0 []*domain.Resource
	if rf, ok := ret.Get(0).(func(string, string) []*domain.Resource); ok {
		r0 = rf(parent, kind)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Resource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(parent, kind)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResourceRepository_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type ResourceRepository_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//  - parent string
//  - kind string
func (_e *ResourceRepository_Expecter) List(parent interface{}, kind interface{}) *ResourceRepository_List_Call {
	return &ResourceRepository_List_Call{Call: _e.mock.On("List", parent, kind)}
}

func (_c *ResourceRepository_List_Call) Run(run func(parent string, kind string)) *ResourceRepository_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *ResourceRepository_List_Call) Return(_a0 []*domain.Resource, _a1 error) *ResourceRepository_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Migrate provides a mock function with given fields:
func (_m *ResourceRepository) Migrate() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResourceRepository_Migrate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Migrate'
type ResourceRepository_Migrate_Call struct {
	*mock.Call
}

// Migrate is a helper method to define mock.On call
func (_e *ResourceRepository_Expecter) Migrate() *ResourceRepository_Migrate_Call {
	return &ResourceRepository_Migrate_Call{Call: _e.mock.On("Migrate")}
}

func (_c *ResourceRepository_Migrate_Call) Run(run func()) *ResourceRepository_Migrate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ResourceRepository_Migrate_Call) Return(_a0 error) *ResourceRepository_Migrate_Call {
	_c.Call.Return(_a0)
	return _c
}

// Update provides a mock function with given fields: r
func (_m *ResourceRepository) Update(r *domain.Resource) error {
	ret := _m.Called(r)

	var r0 error
	if rf, ok := ret.Get(0).(func(*domain.Resource) error); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResourceRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ResourceRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//  - r *domain.Resource
func (_e *ResourceRepository_Expecter) Update(r interface{}) *ResourceRepository_Update_Call {
	return &ResourceRepository_Update_Call{Call: _e.mock.On("Update", r)}
}

func (_c *ResourceRepository_Update_Call) Run(run func(r *domain.Resource)) *ResourceRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*domain.Resource))
	})
	return _c
}

func (_c *ResourceRepository_Update_Call) Return(_a0 error) *ResourceRepository_Update_Call {
	_c.Call.Return(_a0)
	return _c
}
