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

package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	storeMocks "github.com/mendersoftware/mender-server/services/iot-manager/store/mocks"
)

type testError struct{}

func (testError) Error() string {
	return "this is a test"
}

func TestDeviceHasIntegration(t *testing.T) {
	t.Parallel()
	var (
		DeviceID      = uuid.NewString()
		IntegrationID = uuid.New()
	)
	type testCase struct {
		Name string

		Init func(t *testing.T) (*device, *storeMocks.DataStore)

		HasIntegration bool
		Error          error
	}

	testCases := []testCase{{
		Name: "ok",

		Init: func(t *testing.T) (*device, *storeMocks.DataStore) {
			getter := new(storeMocks.DataStore)
			getter.On("GetDevice", contextMatcher, DeviceID).
				Return(&model.Device{
					ID: DeviceID,
					IntegrationIDs: []uuid.UUID{
						IntegrationID,
					},
				}, nil).
				Once()
			device := newDevice(DeviceID, getter)

			return device, getter
		},

		HasIntegration: true,
	}, {
		// Should only call GetDevice once
		Name: "ok/test twice",

		Init: func(t *testing.T) (*device, *storeMocks.DataStore) {
			getter := new(storeMocks.DataStore)
			getter.On("GetDevice", contextMatcher, DeviceID).
				Return(&model.Device{
					ID: DeviceID,
					IntegrationIDs: []uuid.UUID{
						IntegrationID,
					},
				}, nil).
				Once()
			device := newDevice(DeviceID, getter)
			has, err := device.HasIntegration(context.Background(), uuid.Nil)
			assert.NoError(t, err)
			assert.False(t, has)

			return device, getter
		},

		HasIntegration: true,
	}, {
		Name: "error/GetDevice",

		Init: func(t *testing.T) (*device, *storeMocks.DataStore) {
			getter := new(storeMocks.DataStore)
			getter.On("GetDevice", contextMatcher, DeviceID).
				Return(nil, testError{}).
				Once()
			device := newDevice(DeviceID, getter)
			return device, getter
		},

		Error: testError{},
	}, {
		Name: "error/multiple",

		Init: func(t *testing.T) (*device, *storeMocks.DataStore) {
			getter := new(storeMocks.DataStore)
			getter.On("GetDevice", contextMatcher, DeviceID).
				Return(nil, testError{}).
				Once()
			device := newDevice(DeviceID, getter)
			_, err := device.HasIntegration(context.Background(), uuid.Nil)
			assert.ErrorIs(t, err, testError{})

			return device, getter
		},

		Error: testError{},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			dev, getter := tc.Init(t)
			defer getter.AssertExpectations(t)

			canIHaz, err := dev.HasIntegration(ctx, IntegrationID)
			assert.Equal(t, tc.HasIntegration, canIHaz)
			if tc.Error != nil {
				assert.ErrorIs(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
