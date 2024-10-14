// Copyright 2023 Northern.tech AS
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

	orchestrator "github.com/mendersoftware/mender-server/services/deviceauth/client/orchestrator"
	mock "github.com/stretchr/testify/mock"
)

// ClientRunner is an autogenerated mock type for the ClientRunner type
type ClientRunner struct {
	mock.Mock
}

// CheckHealth provides a mock function with given fields: ctx
func (_m *ClientRunner) CheckHealth(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for CheckHealth")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitDeviceDecommisioningJob provides a mock function with given fields: ctx, req
func (_m *ClientRunner) SubmitDeviceDecommisioningJob(ctx context.Context, req orchestrator.DecommissioningReq) error {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for SubmitDeviceDecommisioningJob")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, orchestrator.DecommissioningReq) error); ok {
		r0 = rf(ctx, req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitDeviceLimitWarning provides a mock function with given fields: ctx, devWarn
func (_m *ClientRunner) SubmitDeviceLimitWarning(ctx context.Context, devWarn orchestrator.DeviceLimitWarning) error {
	ret := _m.Called(ctx, devWarn)

	if len(ret) == 0 {
		panic("no return value specified for SubmitDeviceLimitWarning")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, orchestrator.DeviceLimitWarning) error); ok {
		r0 = rf(ctx, devWarn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitProvisionDeviceJob provides a mock function with given fields: ctx, req
func (_m *ClientRunner) SubmitProvisionDeviceJob(ctx context.Context, req orchestrator.ProvisionDeviceReq) error {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for SubmitProvisionDeviceJob")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, orchestrator.ProvisionDeviceReq) error); ok {
		r0 = rf(ctx, req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitReindexReporting provides a mock function with given fields: c, device
func (_m *ClientRunner) SubmitReindexReporting(c context.Context, device string) error {
	ret := _m.Called(c, device)

	if len(ret) == 0 {
		panic("no return value specified for SubmitReindexReporting")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(c, device)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitReindexReportingBatch provides a mock function with given fields: c, devices
func (_m *ClientRunner) SubmitReindexReportingBatch(c context.Context, devices []string) error {
	ret := _m.Called(c, devices)

	if len(ret) == 0 {
		panic("no return value specified for SubmitReindexReportingBatch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) error); ok {
		r0 = rf(c, devices)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitUpdateDeviceInventoryJob provides a mock function with given fields: ctx, req
func (_m *ClientRunner) SubmitUpdateDeviceInventoryJob(ctx context.Context, req orchestrator.UpdateDeviceInventoryReq) error {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for SubmitUpdateDeviceInventoryJob")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, orchestrator.UpdateDeviceInventoryReq) error); ok {
		r0 = rf(ctx, req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitUpdateDeviceStatusJob provides a mock function with given fields: ctx, req
func (_m *ClientRunner) SubmitUpdateDeviceStatusJob(ctx context.Context, req orchestrator.UpdateDeviceStatusReq) error {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for SubmitUpdateDeviceStatusJob")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, orchestrator.UpdateDeviceStatusReq) error); ok {
		r0 = rf(ctx, req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewClientRunner creates a new instance of ClientRunner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClientRunner(t interface {
	mock.TestingT
	Cleanup(func())
}) *ClientRunner {
	mock := &ClientRunner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
