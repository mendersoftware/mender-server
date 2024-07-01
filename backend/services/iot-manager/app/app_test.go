// Copyright 2024 Northern.tech AS
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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/services/iot-manager/client"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iotcore"
	coreMocks "github.com/mendersoftware/mender-server/services/iot-manager/client/iotcore/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iothub"
	hubMocks "github.com/mendersoftware/mender-server/services/iot-manager/client/iothub/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/store"
	storeMocks "github.com/mendersoftware/mender-server/services/iot-manager/store/mocks"
)

var (
	contextMatcher  = mock.MatchedBy(func(ctx context.Context) bool { return true })
	validConnString = &model.ConnectionString{
		HostName: "localhost:8080",
		Key:      crypto.String("not-so-secret-key"),
		Name:     "foobar",
	}
)

type JSONIterator json.Decoder

func (iter *JSONIterator) Next(ctx context.Context) bool {
	dec := (*json.Decoder)(iter)
	return dec.More()
}

func (iter *JSONIterator) Decode(v interface{}) error {
	dec := (*json.Decoder)(iter)
	return dec.Decode(v)
}

func (iter *JSONIterator) Close(ctx context.Context) error {
	return nil
}

func TestNew(t *testing.T) {
	app := New(nil, nil, nil)
	app = app.WithWebhooksTimeout(10)

	assert.NotNil(t, app)
}

func TestRunAndLogError(t *testing.T) {
	var b bytes.Buffer
	logger := log.NewEmpty()
	logger.Logger.Out = &b
	ctx := context.Background()
	ctx = log.WithContext(ctx, logger)
	run := false
	const errString = "unique string caught by logger"
	runAndLogError(ctx, func() error {
		run = true
		return errors.New(errString)
	})
	assert.True(t, run)
	assert.Contains(t, b.String(), errString)
}

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		Name string

		PingReturn    error
		ExpectedError error
	}{
		{
			Name:          "db Ping failed",
			PingReturn:    errors.New("failed to connect to db"),
			ExpectedError: errors.New("failed to connect to db"),
		},
		{
			Name: "db Ping successful",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("Ping",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
			).Return(tc.PingReturn)
			app := New(store, nil, nil)

			ctx := context.Background()
			err := app.HealthCheck(ctx)
			if tc.ExpectedError != nil {
				assert.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIntegrations(t *testing.T) {
	t.Parallel()
	integrationID := uuid.New()
	type testCase struct {
		Name     string
		Store    func(t *testing.T, self *testCase) *storeMocks.DataStore
		Expected []model.Integration

		// Credentials model.Credentials
		Error error
	}

	testCases := []testCase{
		{
			Name: "ok",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrations", contextMatcher, mock.AnythingOfType("model.IntegrationFilter")).
					Return([]model.Integration{{
						ID:          integrationID,
						Provider:    model.ProviderIoTHub,
						Credentials: model.Credentials{},
					}}, nil)
				return ds
			},
			Expected: []model.Integration{{
				ID:          integrationID,
				Provider:    model.ProviderIoTHub,
				Credentials: model.Credentials{},
			}},
		},
		{
			Name: "error: object not found",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrations", contextMatcher, mock.AnythingOfType("model.IntegrationFilter")).
					Return(nil, errors.New("store error: error retrieving integrations collection results"))
				return ds
			},
			Error: errors.New("store error: error retrieving integrations collection results"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			res, err := app.GetIntegrations(ctx)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.Equal(t, tc.Expected, res)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIntegrationByID(t *testing.T) {
	t.Parallel()

	integrationID := uuid.New()
	type testCase struct {
		Name string

		ID    uuid.UUID
		Store func(t *testing.T, self *testCase) *storeMocks.DataStore

		Integration *model.Integration
		Error       error
	}
	testCases := []testCase{{
		Name: "ok",

		ID: integrationID,
		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrationById", contextMatcher, self.ID).
				Return(self.Integration, nil)
			return ds
		},

		Integration: &model.Integration{
			ID:       integrationID,
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
	}, {
		Name: "error/not found",

		ID: integrationID,
		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrationById", contextMatcher, self.ID).
				Return(nil, store.ErrObjectNotFound)
			return ds
		},

		Error: ErrIntegrationNotFound,
	}, {
		Name: "error/internal",

		ID: integrationID,
		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrationById", contextMatcher, self.ID).
				Return(nil, errors.New("internal error"))
			return ds
		},

		Error: errors.New("internal error"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			store := tc.Store(t, &tc)
			app := New(store, nil, nil)
			integration, err := app.GetIntegrationById(context.Background(), tc.ID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Integration, integration)
			}
		})
	}
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name                  string
		Store                 func(t *testing.T, self *testCase) *storeMocks.DataStore
		CreateIntegrationData model.Integration
		Error                 error
	}

	testCases := []testCase{
		{
			Name: "integration created",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
					Return(&self.CreateIntegrationData, nil)
				return ds
			},
			CreateIntegrationData: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type: model.CredentialTypeSAS,
					ConnectionString: &model.ConnectionString{
						HostName: "localhost",
						Key:      crypto.String("secret"),
						Name:     "foobar",
					},
				},
			},
		},
		{
			Name: "create integration error",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
					Return(nil, errors.New("error creating the integration"))
				return ds
			},
			Error: errors.New("error creating the integration"),
		},
		{
			Name: "error: integration already exists",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
					Return(nil, store.ErrObjectExists)
				return ds
			},
			Error: ErrIntegrationExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("CreateIntegration",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				mock.AnythingOfType("model.Integration"),
			).Return(nil, tc.Error)
			app := New(store, nil, nil)

			ctx := context.Background()
			_, err := app.CreateIntegration(ctx, tc.CreateIntegrationData)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetIntegrationCredentials(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))
	type testCase struct {
		Name        string
		Store       func(t *testing.T, self *testCase) *storeMocks.DataStore
		Credentials model.Credentials
		Error       error
	}

	testCases := []testCase{
		{
			Name: "ok",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(nil)
				return ds
			},
			Credentials: model.Credentials{},
		},
		{
			Name: "error: object not found",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(store.ErrObjectNotFound)
				return ds
			},
			Error: ErrIntegrationNotFound,
		},
		{
			Name: "error: unexpected error",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(errors.New("unexpected error"))
				return ds
			},
			Error: errors.New("unexpected error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			err := app.SetIntegrationCredentials(ctx, integrationID, tc.Credentials)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveIntegration(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))
	type testCase struct {
		Name  string
		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Error error
	}

	testCases := []testCase{
		{
			Name: "ok",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrationById", contextMatcher, integrationID).
					Return(&model.Integration{
						ID:       uuid.Nil,
						Provider: model.ProviderIoTHub,
					}, nil).
					Once()
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, nil).
					Once()
				ds.On("RemoveIntegration", contextMatcher, integrationID).
					Return(nil).
					Once()
				return ds
			},
		},
		{
			Name: "error: get devices by integration ID issue",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrationById", contextMatcher, integrationID).
					Return(&model.Integration{
						ID:       uuid.Nil,
						Provider: model.ProviderIoTHub,
					}, nil).
					Once()
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, errors.New("some error: error retrieving integration collection results"))
				return ds
			},
			Error: errors.New("some error: error retrieving integration collection results"),
		},
		{
			Name: "error: devices with given integration ID exist",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrationById", contextMatcher, integrationID).
					Return(&model.Integration{
						ID:       uuid.Nil,
						Provider: model.ProviderIoTHub,
					}, nil).
					Once()
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(true, nil)
				return ds
			},
			Error: ErrCannotRemoveIntegration,
		},
		{
			Name: "error: integration not found",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrationById", contextMatcher, integrationID).
					Return(nil, store.ErrObjectNotFound).
					Once()
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, nil).
					Once()
				ds.On("RemoveIntegration", contextMatcher, integrationID).
					Return(store.ErrObjectNotFound).
					Once()
				return ds
			},
			Error: ErrIntegrationNotFound,
		},
		{
			Name: "error: unexpected error",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrationById", contextMatcher, integrationID).
					Return(&model.Integration{
						ID:       uuid.Nil,
						Provider: model.ProviderIoTHub,
					}, nil).
					Once()
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, nil)
				ds.On("RemoveIntegration", contextMatcher, integrationID).
					Return(errors.New("unexpected mongo error"))
				return ds
			},
			Error: errors.New("unexpected mongo error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			err := app.RemoveIntegration(ctx, integrationID)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvisionDevice(t *testing.T) {
	t.Parallel()
	testIntegrations := map[model.Provider]model.Integration{
		model.ProviderWebhook: {
			Provider: model.ProviderWebhook,
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			Credentials: model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
					Secret: func() *model.HexSecret {
						sec := model.HexSecret([]byte{'1', '2', '3'})
						return &sec
					}(),
				},
			},
		},
		model.ProviderIoTHub: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
		model.ProviderIoTCore: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID: func() *string {
						s := "1234567890"
						return &s
					}(),
					SecretAccessKey: func() *crypto.String {
						var s crypto.String = "1234567890"
						return &s
					}(),
					Region: func() *string {
						s := "eu-north-south-1"
						return &s
					}(),
					DevicePolicyName: func() *string {
						s := "gibAccess"
						return &s
					}(),
				},
			},
		},
	}

	type testCase struct {
		Name   string
		Device model.DeviceEvent
		Status model.Status

		Store        func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core         func(t *testing.T, self *testCase) *coreMocks.Client
		Hub          func(t *testing.T, self *testCase) *hubMocks.Client
		RoundTripper func(t *testing.T, req *http.Request) (*http.Response, error)
	}
	testCases := []testCase{{
		Name: "ok, webhook integration",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},
		Status: model.StatusAccepted,

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
					// Add an empty (ignored) provider
					{Provider: model.ProviderEmpty},
				}, nil).
				Once().
				On("UpsertDeviceIntegrations",
					contextMatcher,
					self.Device.ID,
					[]uuid.UUID{}).
				Return(new(model.Device), nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.Device.ID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.True(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			return new(hubMocks.Client)
		},
		Core: func(t *testing.T, self *testCase) *coreMocks.Client {
			return new(coreMocks.Client)
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusOK)
			return w.Result(), nil
		},
	}, {
		Name: "error/webhook returns error code",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("UpsertDeviceIntegrations",
					contextMatcher,
					self.Device.ID,
					[]uuid.UUID{}).
				Return(new(model.Device), nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.Device.ID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
						if assert.NotNil(t, stat.StatusCode) {
							assert.Equal(t, http.StatusInternalServerError, *stat.StatusCode)
						}
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
			assert.Contains(t, req.Header, client.ParamSignature)
			assert.Contains(t, req.Header, client.ParamAlgorithmType)
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusInternalServerError)
			return w.Result(), nil
		},
	}, {
		Name: "error/webhook fails to send request",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("UpsertDeviceIntegrations",
					contextMatcher,
					self.Device.ID,
					[]uuid.UUID{}).
				Return(new(model.Device), nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.Device.ID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()

			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
			assert.Contains(t, req.Header, client.ParamSignature)
			assert.Contains(t, req.Header, client.ParamAlgorithmType)
			return nil, errors.New("internal error")
		},
	}, {
		Name: "error/webhook fails to create request",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{{
					// Fails to validate
					Provider: model.ProviderWebhook,
				}}, nil).
				Once().
				On("UpsertDeviceIntegrations",
					contextMatcher,
					self.Device.ID,
					[]uuid.UUID{}).
				Return(new(model.Device), nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.Device.ID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
	}, {
		Name: "error/UpsertDeviceIntegrations with error stack",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration(nil), nil).
				Once().
				On("UpsertDeviceIntegrations",
					contextMatcher,
					self.Device.ID,
					[]uuid.UUID{}).
				Return(new(model.Device), errors.New("internal error")).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceProvisioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.Device.ID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 0)
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			return nil, errors.New("internal error")
		},
	}, {
		Name: "error/GetIntegrations/internal error",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return(nil, errors.New("internal error"))
			return mockedStore
		},
	}, {
		Name: "ok/GetIntegrations/not found",
		Device: model.DeviceEvent{
			ID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		},

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return(nil, store.ErrObjectNotFound)
			return mockedStore
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			defer ds.AssertExpectations(t)
			core := new(coreMocks.Client)
			if tc.Core != nil {
				core = tc.Core(t, &tc)
			}
			defer core.AssertExpectations(t)
			hub := new(hubMocks.Client)
			if tc.Hub != nil {
				hub = tc.Hub(t, &tc)
			}
			defer hub.AssertExpectations(t)
			client := &http.Client{
				Transport: roundTripperFunc(
					func(req *http.Request) (*http.Response, error) {
						return tc.RoundTripper(t, req)
					},
				),
			}

			a := &app{
				store:         ds,
				iothubClient:  hub,
				iotcoreClient: core,
				httpClient:    client,
			}

			err := a.ProvisionDevice(ctx, tc.Device)
			assert.NoError(t, err)

			// wait for the completion of the async go routine
			time.Sleep(500 * time.Millisecond)
		})
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (tRT roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tRT(req)
}

func TestDecommissionDevice(t *testing.T) {
	t.Parallel()
	testIntegrations := map[model.Provider]model.Integration{
		model.ProviderWebhook: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			Provider: model.ProviderWebhook,
			Credentials: model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
					Secret: func() *model.HexSecret {
						sec := model.HexSecret([]byte{'1', '2', '3'})
						return &sec
					}(),
				},
			},
		},
		model.ProviderIoTHub: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
		model.ProviderIoTCore: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID: func() *string {
						s := "1234567890"
						return &s
					}(),
					SecretAccessKey: func() *crypto.String {
						var s crypto.String = "1234567890"
						return &s
					}(),
					Region: func() *string {
						s := "eu-north-south-1"
						return &s
					}(),
					DevicePolicyName: func() *string {
						s := "gibAccess"
						return &s
					}(),
				},
			},
		},
	}

	type testCase struct {
		Name     string
		DeviceID string

		Store        func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core         func(t *testing.T, self *testCase) *coreMocks.Client
		Hub          func(t *testing.T, self *testCase) *hubMocks.Client
		RoundTripper func(t *testing.T, req *http.Request) (*http.Response, error)
	}
	testCases := []testCase{{
		Name:     "ok/all the integrations",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetDevice", contextMatcher, self.DeviceID).
				Return(&model.Device{
					ID: self.DeviceID,
					IntegrationIDs: []uuid.UUID{
						testIntegrations[model.ProviderWebhook].ID,
						testIntegrations[model.ProviderIoTHub].ID,
						testIntegrations[model.ProviderIoTCore].ID,
					},
				}, nil).
				Once().
				On("GetIntegrations",
					contextMatcher,
					model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
					testIntegrations[model.ProviderIoTHub],
					testIntegrations[model.ProviderIoTCore],
					// Add an empty (ignored) provider
					{Provider: model.ProviderEmpty},
				}, nil).
				Once().
				On("DeleteDevice", contextMatcher, self.DeviceID).
				Return(nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 3)
					for _, stat := range event.DeliveryStatus {
						assert.True(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)
			hub.On("DeleteDevice",
				contextMatcher,
				testIntegrations[model.ProviderIoTHub].
					Credentials.ConnectionString,
				self.DeviceID).
				Return(nil).
				Once()
			return hub
		},
		Core: func(t *testing.T, self *testCase) *coreMocks.Client {
			core := new(coreMocks.Client)
			core.On("DeleteDevice",
				contextMatcher,
				*testIntegrations[model.ProviderIoTCore].
					Credentials.AWSCredentials,
				self.DeviceID).
				Return(nil).
				Once()
			return core
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusOK)
			return w.Result(), nil
		},
	}, {
		Name:     "error/webhook returns error code",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("DeleteDevice", contextMatcher, self.DeviceID).
				Return(nil).
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
						if assert.NotNil(t, stat.StatusCode) {
							assert.Equal(t, *stat.StatusCode, http.StatusInternalServerError)
						}
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
			assert.Contains(t, req.Header, client.ParamSignature)
			assert.Contains(t, req.Header, client.ParamAlgorithmType)
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusInternalServerError)
			return w.Result(), nil
		},
	}, {
		Name:     "error/webhook fails to send request",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("DeleteDevice", contextMatcher, self.DeviceID).
				Return(nil).
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()

			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
			assert.Contains(t, req.Header, client.ParamSignature)
			assert.Contains(t, req.Header, client.ParamAlgorithmType)
			return nil, errors.New("internal error")
		},
	}, {
		Name:     "error/webhook fails to create request",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{{
					// Fails to validate
					Provider: model.ProviderWebhook,
				}}, nil).
				Once().
				On("DeleteDevice", contextMatcher, self.DeviceID).
				Return(nil).
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
	}, {
		Name:     "error/device not found (DeleteDevice) with error stack",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("DeleteDevice", contextMatcher, self.DeviceID).
				Return(store.ErrObjectNotFound).
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceDecommissioned, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			return nil, errors.New("internal error")
		},
	}, {
		Name:     "error: device not found in db in GetDeviceIntegrations",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
				Return(nil, store.ErrObjectNotFound)
			return mockedStore
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer recover()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			defer ds.AssertExpectations(t)
			core := new(coreMocks.Client)
			if tc.Core != nil {
				core = tc.Core(t, &tc)
			}
			defer core.AssertExpectations(t)
			hub := new(hubMocks.Client)
			if tc.Hub != nil {
				hub = tc.Hub(t, &tc)
			}
			defer hub.AssertExpectations(t)
			client := &http.Client{
				Transport: roundTripperFunc(
					func(req *http.Request) (*http.Response, error) {
						return tc.RoundTripper(t, req)
					},
				),
			}

			a := &app{
				store:         ds,
				iothubClient:  hub,
				iotcoreClient: core,
				httpClient:    client,
			}

			err := a.DecommissionDevice(ctx, tc.DeviceID)
			assert.NoError(t, err)

			// wait for the completion of the async go routine
			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestSetDeviceStatus(t *testing.T) {
	t.Parallel()
	testIntegrations := map[model.Provider]model.Integration{
		model.ProviderWebhook: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			Provider: model.ProviderWebhook,
			Credentials: model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
					Secret: func() *model.HexSecret {
						sec := model.HexSecret([]byte{'1', '2', '3'})
						return &sec
					}(),
				},
			},
		},
		model.ProviderIoTHub: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
		model.ProviderIoTCore: {
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID: func() *string {
						s := "1234567890"
						return &s
					}(),
					SecretAccessKey: func() *crypto.String {
						var s crypto.String = "1234567890"
						return &s
					}(),
					Region: func() *string {
						s := "eu-north-south-1"
						return &s
					}(),
					DevicePolicyName: func() *string {
						s := "gibAccess"
						return &s
					}(),
				},
			},
		},
	}

	type testCase struct {
		Name     string
		DeviceID string
		Status   model.Status

		Store        func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core         func(t *testing.T, self *testCase) *coreMocks.Client
		Hub          func(t *testing.T, self *testCase) *hubMocks.Client
		RoundTripper func(t *testing.T, req *http.Request) (*http.Response, error)
	}
	testCases := []testCase{{
		Name:     "ok, all the integrations",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
		Status:   model.StatusAccepted,

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
					testIntegrations[model.ProviderIoTHub],
					testIntegrations[model.ProviderIoTCore],
					// Add an empty (ignored) provider
					{Provider: model.ProviderEmpty},
				}, nil).
				Once().
				On("GetDevice", contextMatcher, self.DeviceID).
				Return(&model.Device{
					ID: self.DeviceID,
					IntegrationIDs: []uuid.UUID{
						testIntegrations[model.ProviderWebhook].ID,
						testIntegrations[model.ProviderIoTHub].ID,
						testIntegrations[model.ProviderIoTCore].ID,
					},
				}, nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 3)
					for _, stat := range event.DeliveryStatus {
						assert.True(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		Hub: func(t *testing.T, self *testCase) *hubMocks.Client {
			hub := new(hubMocks.Client)
			hubDev := &iothub.Device{
				DeviceID: self.DeviceID,
				Status:   iothub.StatusDisabled,
			}
			hub.On("GetDevice",
				contextMatcher,
				testIntegrations[model.ProviderIoTHub].
					Credentials.ConnectionString,
				self.DeviceID).
				Return(hubDev, nil).
				Once().
				On("UpsertDevice",
					contextMatcher,
					testIntegrations[model.ProviderIoTHub].
						Credentials.ConnectionString,
					self.DeviceID,
					mock.MatchedBy(func(actual *iothub.Device) bool {
						return actual.Status == iothub.StatusEnabled &&
							actual.DeviceID == self.DeviceID
					})).
				Run(func(args mock.Arguments) {
					hubDev.Status = iothub.StatusEnabled
				}).
				Return(hubDev, nil)
			return hub
		},
		Core: func(t *testing.T, self *testCase) *coreMocks.Client {
			core := new(coreMocks.Client)
			core.On("UpsertDevice",
				contextMatcher,
				*testIntegrations[model.ProviderIoTCore].
					Credentials.AWSCredentials,
				self.DeviceID,
				mock.MatchedBy(func(actual *iotcore.Device) bool {
					return actual.Status == iotcore.StatusEnabled
				}),
				*testIntegrations[model.ProviderIoTCore].
					Credentials.AWSCredentials.DevicePolicyName).
				Return(new(iotcore.Device), nil).
				Once()
			return core
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusOK)
			return w.Result(), nil
		},
	}, {
		Name:     "error/webhook returns error code",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
						if assert.NotNil(t, stat.StatusCode) {
							assert.Equal(t, *stat.StatusCode, http.StatusInternalServerError)
						}
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
			assert.Contains(t, req.Header, client.ParamSignature)
			assert.Contains(t, req.Header, client.ParamAlgorithmType)
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusInternalServerError)
			return w.Result(), nil
		},
	}, {
		Name:     "error/webhook fails to send request",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()

			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			var event model.WebhookEvent
			err := json.Unmarshal(b, &event)
			assert.NoError(t, err)
			assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
			assert.Contains(t, req.Header, client.ParamSignature)
			assert.Contains(t, req.Header, client.ParamAlgorithmType)
			return nil, &net.DNSError{Err: "unknown", Name: "localhost"}
		},
	}, {
		Name:     "error/webhook fails to create request",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{{
					// Fails to validate
					Provider: model.ProviderWebhook,
				}}, nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(nil).
				Once()
			return mockedStore
		},
	}, {
		Name:     "error/SaveEvent internal error",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations",
				contextMatcher,
				model.IntegrationFilter{}).
				Return([]model.Integration{
					testIntegrations[model.ProviderWebhook],
				}, nil).
				Once().
				On("SaveEvent", contextMatcher, mock.AnythingOfType("model.Event")).
				Run(func(args mock.Arguments) {
					event := args.Get(1).(model.Event)
					assert.Equal(t, model.EventTypeDeviceStatusChanged, event.Type)
					if assert.IsType(t, model.DeviceEvent{}, event.Data) {
						assert.Equal(t,
							self.DeviceID,
							event.Data.(model.DeviceEvent).ID)
					}
					assert.Len(t, event.DeliveryStatus, 1)
					for _, stat := range event.DeliveryStatus {
						assert.False(t, stat.Success)
					}
				}).
				Return(errors.New("internal error")).
				Once()
			return mockedStore
		},
		RoundTripper: func(t *testing.T, req *http.Request) (*http.Response, error) {
			return nil, errors.New("internal error")
		},
	}, {
		Name:     "ok: integration does not exist",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
				Return(nil, store.ErrObjectNotFound)
			return mockedStore
		},
	}, {
		Name:     "error: error getting integrations",
		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			mockedStore := new(storeMocks.DataStore)
			mockedStore.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
				Return(nil, io.ErrClosedPipe)
			return mockedStore
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			defer ds.AssertExpectations(t)
			core := new(coreMocks.Client)
			if tc.Core != nil {
				core = tc.Core(t, &tc)
			}
			defer core.AssertExpectations(t)
			hub := new(hubMocks.Client)
			if tc.Hub != nil {
				hub = tc.Hub(t, &tc)
			}
			defer hub.AssertExpectations(t)
			client := &http.Client{
				Transport: roundTripperFunc(
					func(req *http.Request) (*http.Response, error) {
						return tc.RoundTripper(t, req)
					},
				),
			}

			a := &app{
				store:         ds,
				iothubClient:  hub,
				iotcoreClient: core,
				httpClient:    client,
			}

			err := a.SetDeviceStatus(ctx, tc.DeviceID, tc.Status)
			assert.NoError(t, err)

			// wait for the completion of the async go routine
			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestGetDevice(t *testing.T) {
	testCases := []struct {
		Name string

		DeviceID       string
		GetDevice      *model.Device
		GetDeviceError error
		Error          error
	}{
		{
			Name: "ok",

			DeviceID: "1",
			GetDevice: &model.Device{
				ID: "1",
			},
		},
		{
			Name: "ok, device doesn't exist",

			DeviceID:       "1",
			GetDeviceError: store.ErrObjectNotFound,
			Error:          ErrDeviceNotFound,
		},
		{
			Name: "ko, device retrieval error",

			DeviceID:       "1",
			GetDeviceError: errors.New("error getting the device"),
			Error:          errors.New("error getting the device"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetDevice",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				tc.DeviceID,
			).Return(tc.GetDevice, tc.GetDeviceError)
			app := New(store, nil, nil)

			ctx := context.Background()
			device, err := app.GetDevice(ctx, tc.DeviceID)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetDevice, device)
			}
		})
	}
}

func TestGetDeviceStateIntegration(t *testing.T) {
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		DeviceID      string
		IntegrationID uuid.UUID

		GetDeviceByIntegrationID       *model.Device
		GetDeviceByIntegrationIDError  error
		GetIntegration                 *model.Integration
		GetIntegrationError            error
		GetDeviceStateIntegration      *model.DeviceState
		GetDeviceStateIntegrationError error
	}{
		{
			Name: "ko, device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationIDError:  store.ErrObjectNotFound,
			GetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationIDError:  errors.New("internal error"),
			GetDeviceStateIntegrationError: errors.New("failed to retrieve the device: internal error"),
		},
		{
			Name: "ko, integration not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			GetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			GetIntegrationError:            errors.New("internal error"),
			GetDeviceStateIntegrationError: errors.New("failed to retrieve the integration: internal error"),
		},
		{
			Name: "ko, unknown integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.Provider("super-secret-provider"),
			},
			GetDeviceStateIntegrationError: ErrUnknownIntegration,
		},
		{
			Name: "ko, no connection string",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.ProviderIoTHub,
			},
			GetDeviceStateIntegrationError: ErrNoCredentials,
		},
		{
			Name: "ko, IoT Core, no credentials",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.ProviderIoTCore,
			},
			GetDeviceStateIntegrationError: ErrNoCredentials,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetDeviceByIntegrationID",
				contextMatcher,
				tc.DeviceID,
				tc.IntegrationID,
			).Return(
				tc.GetDeviceByIntegrationID,
				tc.GetDeviceByIntegrationIDError,
			)
			if tc.GetDeviceByIntegrationIDError == nil {
				store.On("GetIntegrationById",
					contextMatcher,
					tc.IntegrationID,
				).Return(
					tc.GetIntegration,
					tc.GetIntegrationError,
				)
			}
			app := New(store, nil, nil)

			ctx := context.Background()
			state, err := app.GetDeviceStateIntegration(ctx, tc.DeviceID, tc.IntegrationID)
			if tc.GetDeviceStateIntegrationError != nil {
				assert.EqualError(t, err, tc.GetDeviceStateIntegrationError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetDeviceStateIntegration, state)
			}
		})
	}
}

func TestSetDeviceStateIntegration(t *testing.T) {
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		DeviceID      string
		IntegrationID uuid.UUID

		GetDeviceByIntegrationID       *model.Device
		GetDeviceByIntegrationIDError  error
		GetIntegration                 *model.Integration
		GetIntegrationError            error
		SetDeviceStateIntegration      *model.DeviceState
		SetDeviceStateIntegrationError error
	}{
		{
			Name: "ko, device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			SetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationIDError:  errors.New("internal error"),
			SetDeviceStateIntegrationError: errors.New("failed to retrieve the device: internal error"),
		},
		{
			Name: "ko, integration not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			SetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			GetIntegrationError:            errors.New("internal error"),
			SetDeviceStateIntegrationError: errors.New("failed to retrieve the integration: internal error"),
		},
		{
			Name: "ko, unknown integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.Provider("super-secret-provider"),
			},
			SetDeviceStateIntegrationError: ErrUnknownIntegration,
		},
		{
			Name: "ko, no connection string",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.ProviderIoTHub,
			},
			SetDeviceStateIntegrationError: ErrNoCredentials,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetDeviceByIntegrationID",
				contextMatcher,
				tc.DeviceID,
				tc.IntegrationID,
			).Return(
				tc.GetDeviceByIntegrationID,
				tc.GetDeviceByIntegrationIDError,
			)
			if tc.GetDeviceByIntegrationID != nil && tc.GetDeviceByIntegrationIDError == nil {
				store.On("GetIntegrationById",
					contextMatcher,
					tc.IntegrationID,
				).Return(
					tc.GetIntegration,
					tc.GetIntegrationError,
				)
			}
			app := New(store, nil, nil)

			ctx := context.Background()
			state := &model.DeviceState{}
			state, err := app.SetDeviceStateIntegration(ctx, tc.DeviceID, tc.IntegrationID, state)
			if tc.SetDeviceStateIntegrationError != nil {
				assert.EqualError(t, err, tc.SetDeviceStateIntegrationError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.SetDeviceStateIntegration, state)
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	t.Parallel()
	fltr := model.EventsFilter{
		Limit: 10,
	}
	ds := new(storeMocks.DataStore)
	defer ds.AssertExpectations(t)
	ds.On("GetEvents", contextMatcher, fltr).
		Return([]model.Event{}, nil)
	app := New(ds, nil, nil)
	events, err := app.GetEvents(context.Background(), fltr)
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}
