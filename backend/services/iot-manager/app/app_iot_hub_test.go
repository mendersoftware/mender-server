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

package app

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/iot-manager/client"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/devauth"
	mdevauth "github.com/mendersoftware/mender-server/services/iot-manager/client/devauth/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iothub"
	hubMocks "github.com/mendersoftware/mender-server/services/iot-manager/client/iothub/mocks"
	wfMocks "github.com/mendersoftware/mender-server/services/iot-manager/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/store"
	storeMocks "github.com/mendersoftware/mender-server/services/iot-manager/store/mocks"
)

func TestProvisionDeviceIoTHub(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	connString := &model.ConnectionString{
		HostName: "localhost",
		Key:      crypto.String("secret"),
		Name:     "foobar",
	}
	type testCase struct {
		Name        string
		DeviceID    string
		Integration model.Integration

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Hub   func(t *testing.T, self *testCase) *hubMocks.Client
		Wf    func(t *testing.T, self *testCase) *wfMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},

			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("UpsertDevice", contextMatcher, connString, self.DeviceID).
					Return(&iothub.Device{
						DeviceID: self.DeviceID,
						Auth: &iothub.Auth{
							Type: iothub.AuthTypeSymmetric,
							SymmetricKey: &iothub.SymmetricKey{
								Primary:   iothub.Key("key1"),
								Secondary: iothub.Key("key2"),
							},
						},
					}, nil).
					On("UpdateDeviceTwin", contextMatcher, connString, self.DeviceID,
						&iothub.DeviceTwinUpdate{
							Tags: map[string]interface{}{"mender": true},
						}).
					Return(nil)
				return hub
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				wf := new(wfMocks.Client)
				primKey := &model.ConnectionString{
					Key:      crypto.String("key1"),
					DeviceID: self.DeviceID,
					HostName: connString.HostName,
				}
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyPrimaryKey: primKey.String(),
					}).Return(nil)
				return wf
			},
		},
		{
			Name:     "error, no connection string",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},

			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("UpsertDevice", contextMatcher, connString, self.DeviceID).
					Return(&iothub.Device{
						DeviceID: self.DeviceID,
						Auth: &iothub.Auth{
							Type: iothub.AuthTypeNone,
						},
					}, nil)
				return hub
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				return new(wfMocks.Client)
			},

			Error: ErrNoDeviceConnectionString,
		},
		{
			Name:     "error, failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},

			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("UpsertDevice", contextMatcher, connString, self.DeviceID).
					Return(nil, errors.New("internal error"))
				return hub
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				return new(wfMocks.Client)
			},

			Error: errors.New("failed to update iothub devices: internal error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			ds := new(storeMocks.DataStore)
			defer ds.AssertExpectations(t)

			wf := tc.Wf(t, &tc)
			defer wf.AssertExpectations(t)

			a := New(ds, wf, nil)

			hub := tc.Hub(t, &tc)
			defer hub.AssertExpectations(t)
			a = a.WithIoTHub(hub)

			err := a.(*app).provisionIoTHubDevice(ctx, tc.DeviceID, tc.Integration)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecommissionDeviceIoTHub(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	connString := &model.ConnectionString{
		HostName: "localhost",
		Key:      crypto.String("secret"),
		Name:     "foobar",
	}
	type testCase struct {
		Name        string
		DeviceID    string
		Integration model.Integration

		Hub func(t *testing.T, self *testCase) *hubMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("DeleteDevice", contextMatcher, connString, self.DeviceID).
					Return(nil)
				return hub
			},
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},
		},
		{
			Name:     "error, no connection string",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				return hub
			},
			Integration: model.Integration{
				ID:          integrationID,
				Provider:    model.ProviderIoTHub,
				Credentials: model.Credentials{},
			},
			Error: ErrNoCredentials,
		},
		{
			Name:     "error, failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("DeleteDevice", contextMatcher, connString, self.DeviceID).
					Return(errors.New("failed to delete IoT Hub device: store: unexpected error"))
				return hub
			},
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},
			Error: errors.New("failed to delete IoT Hub device: store: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := new(storeMocks.DataStore)
			defer ds.AssertExpectations(t)

			a := New(ds, nil, nil)

			hub := tc.Hub(t, &tc)
			defer hub.AssertExpectations(t)
			a = a.WithIoTHub(hub)

			err := a.(*app).decommissionIoTHubDevice(ctx, tc.DeviceID, tc.Integration)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetDeviceStatusIoTHub(t *testing.T) {
	t.Parallel()
	connString := &model.ConnectionString{
		HostName: "localhost",
		Key:      crypto.String("secret"),
		Name:     "foobar",
	}
	type testCase struct {
		Name string

		ConnStr     *model.ConnectionString
		DeviceID    string
		Status      model.Status
		Integration model.Integration

		Hub func(t *testing.T, self *testCase) *hubMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name: "ok",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Integration: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				dev := &iothub.Device{
					DeviceID: "foobar",
					Status:   iothub.StatusDisabled,
				}
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(dev, nil).
					On("UpsertDevice", contextMatcher, connString, self.DeviceID,
						mock.MatchedBy(func(dev *iothub.Device) bool {
							return dev.Status == iothub.StatusEnabled
						})).
					Return(dev, nil)
				return hub
			},
		},
		{
			Name: "ok, device already has matching status",

			Status:   model.StatusDecommissioned,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Integration: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				dev := &iothub.Device{
					DeviceID: self.DeviceID,
					Status:   iothub.StatusDisabled,
				}
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(dev, nil)
				return hub
			},
		},
		{
			Name: "error, fail to update device",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				dev := &iothub.Device{
					DeviceID: self.DeviceID,
					Status:   iothub.StatusDisabled,
				}
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(dev, nil).
					On("UpsertDevice", contextMatcher, connString, self.DeviceID,
						mock.MatchedBy(func(dev *iothub.Device) bool {
							return dev.Status == iothub.StatusEnabled
						})).
					Return(nil, errors.New("failed to update IoT Hub device: hub: unexpected error"))
				return hub
			},
			Error: errors.New("failed to update IoT Hub device: hub: unexpected error"),
		},
		{
			Name: "error, get device",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Integration: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: connString,
				},
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(nil, errors.New("failed to retrieve device from IoT Hub: hub: unexpected error"))
				return hub
			},
			Error: errors.New("failed to retrieve device from IoT Hub: hub: unexpected error"),
		},
		{
			Name: "error, no connection string",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Integration: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type: model.CredentialTypeSAS,
				},
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				return hub
			},
			Error: ErrNoCredentials,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := new(storeMocks.DataStore)
			defer ds.AssertExpectations(t)

			a := New(ds, nil, nil)

			hub := tc.Hub(t, &tc)
			defer hub.AssertExpectations(t)
			a = a.WithIoTHub(hub)

			err := a.(*app).setDeviceStatusIoTHub(ctx, tc.DeviceID, tc.Status, tc.Integration)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveIoTHubMetadata(t *testing.T) {
	testCases := []struct {
		Name string
		In   map[string]interface{}
		Out  map[string]interface{}
	}{
		{
			Name: "ok",
			In: map[string]interface{}{
				"key": "value",
			},
			Out: map[string]interface{}{
				"key": "value",
			},
		},
		{
			Name: "ok, remove $metadata",
			In: map[string]interface{}{
				"key": "value",
				"$metadata": map[string]interface{}{
					"metadata": "value",
				},
			},
			Out: map[string]interface{}{
				"key": "value",
			},
		},
		{
			Name: "ok, remove $version",
			In: map[string]interface{}{
				"key":      "value",
				"$version": 1,
			},
			Out: map[string]interface{}{
				"key": "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			out := removeIoTHubMetadata(tc.In)
			assert.Equal(t, tc.Out, out)
		})
	}
}

func TestGetDeviceStateIoTHub(t *testing.T) {
	integration := &model.Integration{
		Credentials: model.Credentials{
			ConnectionString: &model.ConnectionString{HostName: "dummy"},
		},
	}
	testCases := []struct {
		Name string

		DeviceID    string
		Integration *model.Integration

		IoTHubClient              func(t *testing.T) *hubMocks.Client
		GetDeviceStateIoTHub      *model.DeviceState
		GetDeviceStateIoTHubError error
	}{
		{
			Name: "ok",

			DeviceID:    "1",
			Integration: integration,

			IoTHubClient: func(t *testing.T) *hubMocks.Client {
				hub := &hubMocks.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"key": "value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil)
				return hub
			},
			GetDeviceStateIoTHub: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
				Reported: map[string]interface{}{
					"another-key": "another-value",
				},
			},
		},
		{
			Name: "ko, no connection string",

			DeviceID:    "1",
			Integration: &model.Integration{},

			GetDeviceStateIoTHubError: ErrNoCredentials,
		},
		{
			Name: "ko, error retrieving the device twin",

			DeviceID:    "1",
			Integration: integration,

			IoTHubClient: func(t *testing.T) *hubMocks.Client {
				hub := &hubMocks.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(nil, errors.New("internal error"))
				return hub
			},
			GetDeviceStateIoTHubError: errors.New("failed to get the device twin: internal error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var iotHubClient iothub.Client
			if tc.IoTHubClient != nil {
				client := tc.IoTHubClient(t)
				defer client.AssertExpectations(t)
				iotHubClient = client
			}
			app := New(nil, nil, nil).WithIoTHub(iotHubClient)

			ctx := context.Background()
			state, err := app.GetDeviceStateIoTHub(ctx, tc.DeviceID, tc.Integration)
			if tc.GetDeviceStateIoTHubError != nil {
				assert.EqualError(t, err, tc.GetDeviceStateIoTHubError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetDeviceStateIoTHub, state)
			}
		})
	}
}

func TestSetDeviceStateIoTHub(t *testing.T) {
	integration := &model.Integration{
		Credentials: model.Credentials{
			ConnectionString: &model.ConnectionString{HostName: "dummy"},
		},
	}
	testCases := []struct {
		Name string

		DeviceID    string
		Integration *model.Integration
		DeviceState *model.DeviceState

		IoTHubClient              func(t *testing.T) *hubMocks.Client
		SetDeviceStateIoTHub      *model.DeviceState
		SetDeviceStateIoTHubError error
	}{
		{
			Name: "ok",

			DeviceID:    "1",
			Integration: integration,
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},

			IoTHubClient: func(t *testing.T) *hubMocks.Client {
				hub := &hubMocks.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					ETag: "etag",
					Tags: map[string]interface{}{
						"tag": "value",
					},
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"another-key": "another-value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil).Once()

				hub.On("UpdateDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
					&iothub.DeviceTwinUpdate{
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.UpdateProperties{
							Desired: map[string]interface{}{
								"key": "value",
							},
						},
						ETag:    "etag",
						Replace: true,
					},
				).Return(nil)

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					Tags: map[string]interface{}{
						"tag": "value",
					},
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"key": "value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil).Once()

				return hub
			},
			SetDeviceStateIoTHub: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
				Reported: map[string]interface{}{
					"another-key": "another-value",
				},
			},
		},
		{
			Name: "ko, no connection string",

			DeviceID:    "1",
			Integration: &model.Integration{},

			SetDeviceStateIoTHubError: ErrNoCredentials,
		},
		{
			Name: "ko, error setting the device twin",

			DeviceID:    "1",
			Integration: integration,
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},

			IoTHubClient: func(t *testing.T) *hubMocks.Client {
				hub := &hubMocks.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					ETag: "etag",
					Tags: map[string]interface{}{
						"tag": "value",
					},
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"another-key": "another-value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil).Once()

				hub.On("UpdateDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
					&iothub.DeviceTwinUpdate{
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.UpdateProperties{
							Desired: map[string]interface{}{
								"key": "value",
							},
						},
						ETag:    "etag",
						Replace: true,
					},
				).Return(errors.New("internal error"))

				return hub
			},
			SetDeviceStateIoTHubError: errors.New("failed to update the device twin: internal error"),
		},
		{
			Name: "ko, error conflict setting the device twin",

			DeviceID:    "1",
			Integration: integration,
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},

			IoTHubClient: func(t *testing.T) *hubMocks.Client {
				hub := &hubMocks.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					ETag: "etag",
					Tags: map[string]interface{}{
						"tag": "value",
					},
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"another-key": "another-value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil).Once()

				hub.On("UpdateDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
					&iothub.DeviceTwinUpdate{
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.UpdateProperties{
							Desired: map[string]interface{}{
								"key": "value",
							},
						},
						ETag:    "etag",
						Replace: true,
					},
				).Return(client.NewHTTPError(http.StatusPreconditionFailed))

				return hub
			},
			SetDeviceStateIoTHubError: ErrDeviceStateConflict,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var iotHubClient iothub.Client
			if tc.IoTHubClient != nil {
				client := tc.IoTHubClient(t)
				defer client.AssertExpectations(t)
				iotHubClient = client
			}
			app := New(nil, nil, nil).WithIoTHub(iotHubClient)

			ctx := context.Background()
			state, err := app.SetDeviceStateIoTHub(ctx, tc.DeviceID, tc.Integration, tc.DeviceState)
			if tc.SetDeviceStateIoTHubError != nil {
				assert.EqualError(t, err, tc.SetDeviceStateIoTHubError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.SetDeviceStateIoTHub, state)
			}
		})
	}
}

func TestSyncIoTHubDevices(t *testing.T) {
	// FIXME(alf) this should be covered by an acceptance test
	t.Parallel()
	noLogger := log.NewEmpty()
	noLogger.Logger.Out = io.Discard
	type testCase struct {
		Name string

		DeviceIDs   []string
		Integration model.Integration
		FailEarly   bool

		DataStore func(t *testing.T, self *testCase) *storeMocks.DataStore
		Devauth   func(t *testing.T, self *testCase) *mdevauth.Client
		Hub       func(t *testing.T, self *testCase) *hubMocks.Client
		Wf        func(t *testing.T, self *testCase) *wfMocks.Client

		Error error
	}
	testCases := []testCase{{
		Name: "ok/10 devices in all cases",

		DeviceIDs: []string{
			"38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e", // Accepted & not in hub
			"72334767-ff25-48ef-ae10-9dcf4f98587d", // Accepted
			"1280cb45-e941-47fb-922e-8dc55006d127", // Accepted
			"6b7ed385-91ca-4499-a118-3e6b863a9082", // Accepted
			"4e8e5b20-5558-486c-891c-41e3a4d309a4", // Accepted
			"49900bc3-9f2b-4b84-ad0d-bec7313b866b", // Rejected
			"3146cc4d-21eb-4f67-bdb8-96e3222b1b4b", // Rejected
			"02d9ab3e-ca1c-4a61-bf06-b23a224935d4", // No auth
			"a4a32db1-047d-4b4b-9f4a-b86a6c16ab90", // Not in deviceauth
			"1434a240-e556-4acf-b96d-ac66a20f82de", // Not in deviceauth
		},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},

		DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)

			ds.On("GetDevice",
				contextMatcher,
				"a4a32db1-047d-4b4b-9f4a-b86a6c16ab90").
				Return(&model.Device{
					ID:             "a4a32db1-047d-4b4b-9f4a-b86a6c16ab90",
					IntegrationIDs: []uuid.UUID{self.Integration.ID},
				}, nil).
				Once().
				On("GetDevice",
					contextMatcher,
					"1434a240-e556-4acf-b96d-ac66a20f82de").
				Return(&model.Device{
					ID:             "a4a32db1-047d-4b4b-9f4a-b86a6c16ab90",
					IntegrationIDs: []uuid.UUID{},
				}, nil).
				Once()
			ds.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{self.Integration}, nil).
				Twice()

			ds.On("DeleteDevice",
				contextMatcher,
				"a4a32db1-047d-4b4b-9f4a-b86a6c16ab90").
				Return(errors.New("internal error")).
				On("DeleteDevice",
					contextMatcher,
					"1434a240-e556-4acf-b96d-ac66a20f82de").
				Return(store.ErrObjectNotFound).
				Once()
			ds.On("SaveEvent",
				contextMatcher,
				mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t,
						model.EventTypeDeviceDecommissioned,
						event.Type)
				}).
				Return(nil)

			return ds
		},
		Devauth: func(t *testing.T, self *testCase) *mdevauth.Client {
			da := new(mdevauth.Client)
			authSets := make([]devauth.Device, 0, len(self.DeviceIDs))
			for i, id := range self.DeviceIDs[:8] {
				var status model.Status
				if i < 5 {
					status = model.StatusAccepted
				} else if i < 7 {
					status = model.StatusRejected
				} else if i < 8 {
					status = model.StatusNoAuth
				} else {
					continue
				}
				authSets = append(authSets, devauth.Device{
					ID:     id,
					Status: status,
				})
			}
			da.On("GetDevices", contextMatcher, self.DeviceIDs).
				Return(authSets, nil)
			return da
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)
			hub.On("DeleteDevice",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				"a4a32db1-047d-4b4b-9f4a-b86a6c16ab90").
				Return(nil).
				Once()

			twins := make([]iothub.DeviceTwin, len(self.DeviceIDs)-3)
			for i, id := range self.DeviceIDs[1:8] {
				twin := iothub.DeviceTwin{
					DeviceID: id,
					Status:   iothub.StatusEnabled,
				}
				dev := iothub.Device{
					DeviceID: id,
					Status:   iothub.StatusEnabled,
					Auth: &iothub.Auth{
						Type: iothub.AuthTypeSymmetric,
						SymmetricKey: &iothub.SymmetricKey{
							Primary:   iothub.Key("secret"),
							Secondary: iothub.Key("key"),
						},
					},
				}
				twins[i] = twin
				if i >= 4 {
					hub.On("GetDevice",
						contextMatcher,
						self.Integration.Credentials.ConnectionString,
						id).
						Return(&dev, nil).
						Once()
					devUpdate := dev
					devUpdate.Status = iothub.StatusDisabled
					hub.On("UpsertDevice",
						contextMatcher,
						self.Integration.Credentials.ConnectionString,
						id,
						&devUpdate).
						Return(&devUpdate, nil).
						Once()
				}
			}
			hub.On("GetDeviceTwins",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				self.DeviceIDs[:8]).
				Return(twins, nil)

			devUpdate := &iothub.Device{
				DeviceID: "38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
				Status:   iothub.StatusEnabled,
				Auth: &iothub.Auth{
					Type: iothub.AuthTypeSymmetric,
					SymmetricKey: &iothub.SymmetricKey{
						Primary:   iothub.Key("secret"),
						Secondary: iothub.Key("key"),
					},
				},
			}
			hub.On("UpsertDevice",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				self.DeviceIDs[0],
				mock.AnythingOfType("*iothub.Device")).
				Return(devUpdate, nil).
				Once()
			hub.On("UpdateDeviceTwin",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				self.DeviceIDs[0],
				mock.AnythingOfType("*iothub.DeviceTwinUpdate")).
				Return(nil)

			return hub
		},
		Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
			wf := new(wfMocks.Client)
			devCS := &model.ConnectionString{
				Key:      crypto.String("secret"),
				DeviceID: "38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
				HostName: validConnString.HostName,
			}
			wf.On("ProvisionExternalDevice",
				contextMatcher,
				"38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
				map[string]string{
					confKeyPrimaryKey: devCS.String(),
				}).Return(nil).Once()
			return wf
		},
	}, {
		Name: "continue on errors",

		DeviceIDs: []string{
			"38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e", // fail decommission
			"e7f14597-6a9a-44fa-84ea-cda8e221e207", // fail hub.GetDevice
			"e4297565-ea38-4031-8287-664d52878890", // fail hub.UpsertDevice
			"8de57bea-be8a-45d5-9147-a148cddf4f09", // fail provisionDevice
		},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},

		DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return(nil, errors.New("internal error"))
			return ds
		},
		Devauth: func(t *testing.T, self *testCase) *mdevauth.Client {
			da := new(mdevauth.Client)
			authSets := make([]devauth.Device, 0, len(self.DeviceIDs))
			for _, id := range self.DeviceIDs[1:] {
				authSets = append(authSets, devauth.Device{
					ID:     id,
					Status: model.StatusAccepted,
				})
			}
			da.On("GetDevices", contextMatcher, self.DeviceIDs).
				Return(authSets, nil)
			return da
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)

			devIDs := make([]string, len(self.DeviceIDs)-1)
			copy(devIDs, self.DeviceIDs)
			devIDs[0] = self.DeviceIDs[len(self.DeviceIDs)-1]
			hub.On("GetDeviceTwins",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				[]string{
					"8de57bea-be8a-45d5-9147-a148cddf4f09",
					"e7f14597-6a9a-44fa-84ea-cda8e221e207",
					"e4297565-ea38-4031-8287-664d52878890",
				},
			).Return([]iothub.DeviceTwin{{
				DeviceID: "e7f14597-6a9a-44fa-84ea-cda8e221e207",
				Status:   iothub.StatusDisabled,
			}, {
				DeviceID: "e4297565-ea38-4031-8287-664d52878890",
				Status:   iothub.StatusDisabled,
			}}, nil)

			devUpdated := iothub.Device{
				DeviceID: "e4297565-ea38-4031-8287-664d52878890",
				Status:   iothub.StatusEnabled,
				Auth: &iothub.Auth{
					Type: iothub.AuthTypeSymmetric,
					SymmetricKey: &iothub.SymmetricKey{
						Primary:   iothub.Key("secret"),
						Secondary: iothub.Key("key"),
					},
				},
			}
			devProvisioned := devUpdated
			devProvisioned.DeviceID = "8de57bea-be8a-45d5-9147-a148cddf4f09"
			hub.On("GetDevice",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				"e7f14597-6a9a-44fa-84ea-cda8e221e207").
				Return(nil, errors.New("internal error")).
				Once().
				On("GetDevice",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					"e4297565-ea38-4031-8287-664d52878890").
				Return(&devUpdated, nil).
				Once()

			hub.On("UpsertDevice",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				"e4297565-ea38-4031-8287-664d52878890",
				&devUpdated).
				Return(nil, errors.New("internal error")).
				Once().
				On("UpsertDevice",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					"8de57bea-be8a-45d5-9147-a148cddf4f09",
					&iothub.Device{
						DeviceID: "8de57bea-be8a-45d5-9147-a148cddf4f09",
						Status:   iothub.StatusEnabled,
					},
				).
				Return(nil, errors.New("internal error"))

			return hub
		},
		Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
			wf := new(wfMocks.Client)
			return wf
		},
	}, {
		Name: "error/devauth",

		DeviceIDs: []string{
			"38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
			"e7f14597-6a9a-44fa-84ea-cda8e221e207",
			"e4297565-ea38-4031-8287-664d52878890",
			"8de57bea-be8a-45d5-9147-a148cddf4f09",
		},

		DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			return ds
		},
		Devauth: func(t *testing.T, self *testCase) *mdevauth.Client {
			da := new(mdevauth.Client)
			da.On("GetDevices", contextMatcher, self.DeviceIDs).
				Return(nil, client.NewHTTPError(500))
			return da
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)
			return hub
		},
		Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
			wf := new(wfMocks.Client)
			return wf
		},
		FailEarly: true,
		Error:     errors.New("app: failed to lookup device authentication"),
	}, {
		Name: "error/decommission device",

		DeviceIDs: []string{
			"38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
		},

		DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
				Return(nil, errors.New("internal error"))
			return ds
		},
		Devauth: func(t *testing.T, self *testCase) *mdevauth.Client {
			da := new(mdevauth.Client)
			da.On("GetDevices", contextMatcher, self.DeviceIDs).
				Return([]devauth.Device{}, nil)
			return da
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)
			return hub
		},
		Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
			wf := new(wfMocks.Client)
			return wf
		},
		FailEarly: true,
		Error:     errors.New("app: failed to decommission device"),
	}, {
		Name: "error/get twins",

		DeviceIDs: []string{
			"38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
		},

		DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			return ds
		},
		Devauth: func(t *testing.T, self *testCase) *mdevauth.Client {
			da := new(mdevauth.Client)
			da.On("GetDevices", contextMatcher, self.DeviceIDs).
				Return([]devauth.Device{{
					ID:     self.DeviceIDs[0],
					Status: model.StatusAccepted,
				}}, nil)
			return da
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)
			hub.On("GetDeviceTwins",
				contextMatcher,
				self.Integration.Credentials.ConnectionString,
				self.DeviceIDs).
				Return(nil, client.NewHTTPError(500))
			return hub
		},
		Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
			wf := new(wfMocks.Client)
			return wf
		},
		FailEarly: true,
		Error:     errors.New("app: failed to get devices from IoT Hub"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := log.WithContext(context.Background(), noLogger)

			ds := tc.DataStore(t, &tc)
			da := tc.Devauth(t, &tc)
			hub := tc.Hub(t, &tc)
			wf := tc.Wf(t, &tc)
			defer da.AssertExpectations(t)
			defer ds.AssertExpectations(t)
			defer hub.AssertExpectations(t)
			defer wf.AssertExpectations(t)

			app := New(ds, wf, da).WithIoTHub(hub).(*app)
			err := app.syncIoTHubDevices(ctx, tc.DeviceIDs, tc.Integration, tc.FailEarly)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyDeviceTwinIotHub(t *testing.T) {
	t.Parallel()
	noLogger := log.NewEmpty()
	noLogger.Logger.Out = io.Discard

	// parse public key
	pubkeyStr := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzogVU7RGDilbsoUt/DdH
VJvcepl0A5+xzGQ50cq1VE/Dyyy8Zp0jzRXCnnu9nu395mAFSZGotZVr+sWEpO3c
yC3VmXdBZmXmQdZqbdD/GuixJOYfqta2ytbIUPRXFN7/I7sgzxnXWBYXYmObYvdP
okP0mQanY+WKxp7Q16pt1RoqoAd0kmV39g13rFl35muSHbSBoAW3GBF3gO+mF5Ty
1ddp/XcgLOsmvNNjY+2HOD5F/RX0fs07mWnbD7x+xz7KEKjF+H7ZpkqCwmwCXaf0
iyYyh1852rti3Afw4mDxuVSD7sd9ggvYMc0QHIpQNkD4YWOhNiE1AB0zH57VbUYG
UwIDAQAB
-----END PUBLIC KEY-----`
	var pubkey model.PublicKey
	err := pubkey.UnmarshalText([]byte(pubkeyStr))
	if err != nil {
		panic(err)
	}

	type testCase struct {
		Name string

		Integration model.Integration
		Req         model.PreauthRequest

		DataStore func(t *testing.T, self *testCase) *storeMocks.DataStore
		Hub       func(t *testing.T, self *testCase) *hubMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name: "ok",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(&iothub.DeviceTwin{
						ETag: "etag",
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.TwinProperties{
							Desired: map[string]interface{}{
								"another-key": "another-value",
							},
							Reported: map[string]interface{}{
								"id_data": map[string]interface{}{"foo": "bar"},
								"pubkey":  pubkeyStr,
							},
						},
					}, nil).Once()

				return hub
			},
		},
		{
			Name: "error getting integrations",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return(nil, errors.New("some error")).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)

				return hub
			},
			Error: errors.New("failed to retrieve integration: some error"),
		},
		{
			Name: "error getting device twin",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(nil, errors.New("some error")).Once()

				return hub
			},
			Error: errors.New("failed to get module twin from integration: some error"),
		},
		{
			Name: "error - identity data missing in the module twin",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(&iothub.DeviceTwin{
						ETag: "etag",
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.TwinProperties{
							Desired: map[string]interface{}{
								"another-key": "another-value",
							},
							Reported: map[string]interface{}{
								"pubkey": pubkeyStr,
							},
						},
					}, nil).Once()

				return hub
			},
			Error: errors.New("missing identity data"),
		},
		{
			Name: "error - reported identity data does not match",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(&iothub.DeviceTwin{
						ETag: "etag",
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.TwinProperties{
							Desired: map[string]interface{}{
								"another-key": "another-value",
							},
							Reported: map[string]interface{}{
								"id_data": map[string]interface{}{"foo": "baz"},
								"pubkey":  pubkeyStr,
							},
						},
					}, nil).Once()

				return hub
			},
			Error: errors.New(`reported "id_data" does not match request`),
		},
		{
			Name: "error - missing pubkey",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(&iothub.DeviceTwin{
						ETag: "etag",
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.TwinProperties{
							Desired: map[string]interface{}{
								"another-key": "another-value",
							},
							Reported: map[string]interface{}{
								"id_data": map[string]interface{}{"foo": "bar"},
							},
						},
					}, nil).Once()

				return hub
			},
			Error: errors.New("missing pubkey"),
		},
		{
			Name: "error - invalid public key from twin",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(&iothub.DeviceTwin{
						ETag: "etag",
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.TwinProperties{
							Desired: map[string]interface{}{
								"another-key": "another-value",
							},
							Reported: map[string]interface{}{
								"id_data": map[string]interface{}{"foo": "bar"},
								"pubkey":  "foo",
							},
						},
					}, nil).Once()

				return hub
			},
			Error: errors.New("invalid public key from twin: invalid public key format"),
		},
		{
			Name: "key does not match",

			Integration: model.Integration{
				ID:       uuid.New(),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type:             model.CredentialTypeSAS,
					ConnectionString: validConnString,
				},
			},

			Req: model.PreauthRequest{
				DeviceID: "foo",
				IdentityData: map[string]interface{}{
					"foo": "bar",
				},
				PublicKey: pubkey,
			},
			DataStore: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)

				ds.On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
					Return([]model.Integration{self.Integration}, nil).
					Once()

				return ds
			},
			Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
				hub := new(hubMocks.Client)
				hub.On("GetDeviceTwin",
					contextMatcher,
					self.Integration.Credentials.ConnectionString,
					self.Req.DeviceID).
					Return(&iothub.DeviceTwin{
						ETag: "etag",
						Tags: map[string]interface{}{
							"tag": "value",
						},
						Properties: iothub.TwinProperties{
							Desired: map[string]interface{}{
								"another-key": "another-value",
							},
							Reported: map[string]interface{}{
								"id_data": map[string]interface{}{"foo": "bar"},
								"pubkey": `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzogVU7RGDilbsoUt/DdH
VJvcepl0A5+xzGQ50cq1VE/Dyyy8Zp0jzRXCnnu9nu395mAFSZGotZVr+sWEpO3c
yC3VmXdBZmXmQdZqbdD/GuixJOYfqta2ytbIUPRXFN7/I7sgzxnXWBYXYmObYvdP
okP0mQanY+WKxp7Q16pt1RoqoAd0kmV39g13rFl35muSHbSBoAW3GBF3gO+mF5Ty
1ddp/XcgLOsmvNNjY+2HOD5F/RX0fs07mWnbD7x+xz7KEKjF+H7ZpkqCwmwCXaf0
iyYyh1852rti3Afw4mDxuVSD7sd9ggvYMc0QHIpQNkD4YWOhNiE1AB0zH57VbUYG
UwIDAQAC
-----END PUBLIC KEY-----`,
							},
						},
					}, nil).Once()

				return hub
			},
			Error: errors.New("key does not match"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := log.WithContext(context.Background(), noLogger)

			ds := tc.DataStore(t, &tc)
			hub := tc.Hub(t, &tc)
			defer ds.AssertExpectations(t)
			defer hub.AssertExpectations(t)

			app := New(ds, nil, nil).WithIoTHub(hub).(*app)
			err := app.VerifyDeviceTwin(ctx, tc.Req)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
