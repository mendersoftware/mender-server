// Copyright 2021 Northern.tech AS
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

// Code generated by mockery v2.2.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/mendersoftware/mender-server/services/deviceconfig/model"

	uuid "github.com/google/uuid"
)

// DataStore is an autogenerated mock type for the DataStore type
type DataStore struct {
	mock.Mock
}

// Close provides a mock function with given fields: ctx
func (_m *DataStore) Close(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteDevice provides a mock function with given fields: ctx, devID
func (_m *DataStore) DeleteDevice(ctx context.Context, devID string) error {
	ret := _m.Called(ctx, devID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, devID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteTenant provides a mock function with given fields: ctx, tenant_id
func (_m *DataStore) DeleteTenant(ctx context.Context, tenant_id string) error {
	ret := _m.Called(ctx, tenant_id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenant_id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropDatabase provides a mock function with given fields: ctx
func (_m *DataStore) DropDatabase(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDevice provides a mock function with given fields: ctx, devID
func (_m *DataStore) GetDevice(ctx context.Context, devID string) (model.Device, error) {
	ret := _m.Called(ctx, devID)

	var r0 model.Device
	if rf, ok := ret.Get(0).(func(context.Context, string) model.Device); ok {
		r0 = rf(ctx, devID)
	} else {
		r0 = ret.Get(0).(model.Device)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, devID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InsertDevice provides a mock function with given fields: ctx, dev
func (_m *DataStore) InsertDevice(ctx context.Context, dev model.Device) error {
	ret := _m.Called(ctx, dev)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Device) error); ok {
		r0 = rf(ctx, dev)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Migrate provides a mock function with given fields: ctx, version, automigrate
func (_m *DataStore) Migrate(ctx context.Context, version string, automigrate bool) error {
	ret := _m.Called(ctx, version, automigrate)

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

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReplaceConfiguration provides a mock function with given fields: ctx, dev
func (_m *DataStore) ReplaceConfiguration(ctx context.Context, dev model.Device) error {
	ret := _m.Called(ctx, dev)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Device) error); ok {
		r0 = rf(ctx, dev)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReplaceReportedConfiguration provides a mock function with given fields: ctx, dev
func (_m *DataStore) ReplaceReportedConfiguration(ctx context.Context, dev model.Device) error {
	ret := _m.Called(ctx, dev)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Device) error); ok {
		r0 = rf(ctx, dev)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDeploymentID provides a mock function with given fields: ctx, devID, deploymentID
func (_m *DataStore) SetDeploymentID(ctx context.Context, devID string, deploymentID uuid.UUID) error {
	ret := _m.Called(ctx, devID, deploymentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) error); ok {
		r0 = rf(ctx, devID, deploymentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateConfiguration provides a mock function with given fields: ctx, deviceID, attrs
func (_m *DataStore) UpdateConfiguration(ctx context.Context, deviceID string, attrs model.Attributes) error {
	ret := _m.Called(ctx, deviceID, attrs)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Attributes) error); ok {
		r0 = rf(ctx, deviceID, attrs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
