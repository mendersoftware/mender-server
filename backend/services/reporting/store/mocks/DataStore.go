// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

// Code generated by mockery v2.45.1. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/mendersoftware/mender-server/services/reporting/model"
	mock "github.com/stretchr/testify/mock"
)

// DataStore is an autogenerated mock type for the DataStore type
type DataStore struct {
	mock.Mock
}

// Close provides a mock function with given fields: ctx
func (_m *DataStore) Close(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropDatabase provides a mock function with given fields: ctx
func (_m *DataStore) DropDatabase(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for DropDatabase")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetMapping provides a mock function with given fields: ctx, tenantID
func (_m *DataStore) GetMapping(ctx context.Context, tenantID string) (*model.Mapping, error) {
	ret := _m.Called(ctx, tenantID)

	if len(ret) == 0 {
		panic("no return value specified for GetMapping")
	}

	var r0 *model.Mapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Mapping, error)); ok {
		return rf(ctx, tenantID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Mapping); ok {
		r0 = rf(ctx, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Mapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Migrate provides a mock function with given fields: ctx, version, automigrate
func (_m *DataStore) Migrate(ctx context.Context, version string, automigrate bool) error {
	ret := _m.Called(ctx, version, automigrate)

	if len(ret) == 0 {
		panic("no return value specified for Migrate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) error); ok {
		r0 = rf(ctx, version, automigrate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MigrateLatest provides a mock function with given fields: ctx
func (_m *DataStore) MigrateLatest(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for MigrateLatest")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Ping provides a mock function with given fields: ctx
func (_m *DataStore) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAndGetMapping provides a mock function with given fields: ctx, tenantID, inventory
func (_m *DataStore) UpdateAndGetMapping(ctx context.Context, tenantID string, inventory []string) (*model.Mapping, error) {
	ret := _m.Called(ctx, tenantID, inventory)

	if len(ret) == 0 {
		panic("no return value specified for UpdateAndGetMapping")
	}

	var r0 *model.Mapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) (*model.Mapping, error)); ok {
		return rf(ctx, tenantID, inventory)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) *model.Mapping); ok {
		r0 = rf(ctx, tenantID, inventory)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Mapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenantID, inventory)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewDataStore creates a new instance of DataStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDataStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *DataStore {
	mock := &DataStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
