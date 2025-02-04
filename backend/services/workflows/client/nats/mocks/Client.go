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

	nats "github.com/mendersoftware/mender-server/services/workflows/client/nats"
	nats_go "github.com/nats-io/nats.go"
	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Client) Close() {
	_m.Called()
}

// CreateConsumer provides a mock function with given fields: name, upsert, config
func (_m *Client) CreateConsumer(name string, upsert bool, config nats.ConsumerConfig) error {
	ret := _m.Called(name, upsert, config)

	if len(ret) == 0 {
		panic("no return value specified for CreateConsumer")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool, nats.ConsumerConfig) error); ok {
		r0 = rf(name, upsert, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetConsumerConfig provides a mock function with given fields: name
func (_m *Client) GetConsumerConfig(name string) (*nats.ConsumerConfig, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for GetConsumerConfig")
	}

	var r0 *nats.ConsumerConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*nats.ConsumerConfig, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) *nats.ConsumerConfig); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*nats.ConsumerConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsConnected provides a mock function with given fields:
func (_m *Client) IsConnected() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsConnected")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// JetStreamCreateStream provides a mock function with given fields: streamName
func (_m *Client) JetStreamCreateStream(streamName string) error {
	ret := _m.Called(streamName)

	if len(ret) == 0 {
		panic("no return value specified for JetStreamCreateStream")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(streamName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// JetStreamPublish provides a mock function with given fields: _a0, _a1
func (_m *Client) JetStreamPublish(_a0 string, _a1 []byte) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for JetStreamPublish")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// JetStreamSubscribe provides a mock function with given fields: ctx, subj, durable, q
func (_m *Client) JetStreamSubscribe(ctx context.Context, subj string, durable string, q chan *nats_go.Msg) (nats.UnsubscribeFunc, error) {
	ret := _m.Called(ctx, subj, durable, q)

	if len(ret) == 0 {
		panic("no return value specified for JetStreamSubscribe")
	}

	var r0 nats.UnsubscribeFunc
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, chan *nats_go.Msg) (nats.UnsubscribeFunc, error)); ok {
		return rf(ctx, subj, durable, q)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, chan *nats_go.Msg) nats.UnsubscribeFunc); ok {
		r0 = rf(ctx, subj, durable, q)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.UnsubscribeFunc)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, chan *nats_go.Msg) error); ok {
		r1 = rf(ctx, subj, durable, q)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StreamName provides a mock function with given fields:
func (_m *Client) StreamName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for StreamName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
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
