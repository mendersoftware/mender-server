// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

// Code generated by mockery v2.45.1. DO NOT EDIT.

package mocks

import (
	context "context"

	deviceauth "github.com/mendersoftware/mender-server/services/reporting/client/deviceauth"
	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// GetDevices provides a mock function with given fields: ctx, tid, deviceIDs
func (_m *Client) GetDevices(ctx context.Context, tid string, deviceIDs []string) (map[string]deviceauth.DeviceAuthDevice, error) {
	ret := _m.Called(ctx, tid, deviceIDs)

	if len(ret) == 0 {
		panic("no return value specified for GetDevices")
	}

	var r0 map[string]deviceauth.DeviceAuthDevice
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) (map[string]deviceauth.DeviceAuthDevice, error)); ok {
		return rf(ctx, tid, deviceIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) map[string]deviceauth.DeviceAuthDevice); ok {
		r0 = rf(ctx, tid, deviceIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]deviceauth.DeviceAuthDevice)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tid, deviceIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
