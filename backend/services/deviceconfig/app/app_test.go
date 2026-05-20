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

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	client_mocks "github.com/mendersoftware/mender-server/pkg/api/client/mocks"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/utils/types"

	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	mstore "github.com/mendersoftware/mender-server/services/deviceconfig/store/mocks"
)

var contextMatcher = mock.MatchedBy(func(ctx context.Context) bool { return true })

func TestHealthCheck(t *testing.T) {
	t.Parallel()
	err := errors.New("error")

	store := &mstore.DataStore{}
	store.On("Ping",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
	).Return(err)

	app := New(store, nil, Config{})

	ctx := context.Background()
	res := app.HealthCheck(ctx)
	assert.Equal(t, err, res)

	store.AssertExpectations(t)
}

func TestProvisionTenant(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	const tenantID = "dummy"
	tenant := model.NewTenant{
		TenantID: tenantID,
	}

	ds := new(mstore.DataStore)
	ds.On("MigrateLatest",
		mock.MatchedBy(func(ctx context.Context) bool {
			id := identity.FromContext(ctx)
			assert.NotNil(t, id)
			assert.Equal(t, id.Tenant, tenantID)
			return true
		}),
	).Return(nil)

	defer ds.AssertExpectations(t)

	app := New(ds, nil, Config{})
	err := app.ProvisionTenant(ctx, tenant)
	assert.NoError(t, err)
}

func TestDeleteTenant(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		tenantId string

		dbErr  error
		outErr string
	}{
		{
			tenantId: "tenant1",
			dbErr:    errors.New("error"),
		},
		{
			tenantId: "tenant2",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			ds := new(mstore.DataStore)
			defer ds.AssertExpectations(t)
			ds.On("DeleteTenant",
				mock.MatchedBy(func(ctx context.Context) bool {
					ident := identity.FromContext(ctx)
					return assert.NotNil(t, ident) &&
						assert.Equal(t, tc.tenantId, ident.Tenant)
				}),
				tc.tenantId,
			).Return(tc.dbErr)
			app := New(ds, nil, Config{})
			err := app.DeleteTenant(ctx, tc.tenantId)

			if tc.dbErr != nil {
				assert.EqualError(t, err, tc.dbErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvisionDevice(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	dev := model.NewDevice{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
	}
	deviceMatcher := mock.MatchedBy(func(d model.Device) bool {
		if !assert.Equal(t, dev.ID, d.ID) {
			return false
		}
		return assert.WithinDuration(t, time.Now(), *d.UpdatedTS, time.Minute)
	})

	ds := new(mstore.DataStore)
	defer ds.AssertExpectations(t)
	ds.On("InsertDevice", ctx, deviceMatcher).Return(nil)

	app := New(ds, nil, Config{})
	err := app.ProvisionDevice(ctx, dev)
	assert.NoError(t, err)
}

func TestGetDevice(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	dev := model.NewDevice{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
	}
	device := model.Device{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
	}
	deviceMatcher := mock.MatchedBy(func(d model.Device) bool {
		if !assert.Equal(t, dev.ID, d.ID) {
			return false
		}
		return assert.WithinDuration(t, time.Now(), *d.UpdatedTS, time.Minute)
	})

	ds := new(mstore.DataStore)
	defer ds.AssertExpectations(t)
	ds.On("InsertDevice", ctx, deviceMatcher).Return(nil)
	ds.On("GetDevice", ctx, dev.ID).Return(device, nil)

	app := New(ds, nil, Config{})
	err := app.ProvisionDevice(ctx, dev)
	assert.NoError(t, err)

	d, err := app.GetDevice(ctx, dev.ID)
	assert.NoError(t, err)
	assert.Equal(t, dev.ID, d.ID)
}

func TestDecommissionDevice(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	devID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String()

	ds := new(mstore.DataStore)
	defer ds.AssertExpectations(t)
	ds.On("DeleteDevice", ctx, devID).Return(nil)

	app := New(ds, nil, Config{})
	err := app.DecommissionDevice(ctx, devID)
	assert.NoError(t, err)
}

func TestSetConfiguration(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	dev := model.NewDevice{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
	}
	device := model.Device{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
		ConfiguredAttributes: []model.Attribute{
			{
				Key:   "hostname",
				Value: "some0",
			},
		},
		ReportedAttributes: []model.Attribute{
			{
				Key:   "hostname",
				Value: "some0other",
			},
		},
	}
	deviceMatcher := mock.MatchedBy(func(d model.Device) bool {
		if !assert.Equal(t, dev.ID, d.ID) {
			return false
		}
		return assert.WithinDuration(t, time.Now(), *d.UpdatedTS, time.Minute)
	})

	ds := new(mstore.DataStore)
	defer ds.AssertExpectations(t)
	ds.On("InsertDevice", ctx, deviceMatcher).Return(nil)
	ds.On("ReplaceConfiguration", ctx, deviceMatcher).Return(nil)
	ds.On("GetDevice", ctx, dev.ID).Return(device, nil)

	app := New(ds, nil, Config{})
	err := app.ProvisionDevice(ctx, dev)
	assert.NoError(t, err)

	err = app.SetConfiguration(ctx, dev.ID, device.ConfiguredAttributes)
	assert.NoError(t, err)

	d, err := app.GetDevice(ctx, dev.ID)
	assert.NoError(t, err)

	assert.Equal(t, d.ConfiguredAttributes, device.ConfiguredAttributes)

	err = app.SetConfiguration(ctx, dev.ID, []model.Attribute{
		{
			Key:   "hostname",
			Value: "other",
		},
	})
	assert.NoError(t, err)

	d, err = app.GetDevice(ctx, dev.ID)
	assert.NoError(t, err)

	assert.NotEqual(t, device.ConfiguredAttributes, d.ConfiguredAttributes[0])

	err = app.SetConfiguration(ctx, dev.ID, []model.Attribute{
		{
			Key:   "hostname",
			Value: "",
		},
	})
	assert.NoError(t, err)
}

func TestUpdateConfiguration(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		CTX      context.Context
		DeviceID string
		Attrs    model.Attributes

		Store func(t *testing.T, self *testCase) *mstore.DataStore
		Wf    func(t *testing.T, self *testCase) client.WorkflowsOtherAPI

		Error error
	}
	testCases := []testCase{{
		Name: "ok/single tenant",

		CTX:      context.Background(),
		DeviceID: "e1ce5c7a-5819-4ee1-aff7-f4fb0b50009c",
		Attrs: model.Attributes{{
			Key:   "key",
			Value: "value",
		}},

		Store: func(t *testing.T, self *testCase) *mstore.DataStore {
			store := new(mstore.DataStore)
			store.On("UpdateConfiguration",
				contextMatcher,
				self.DeviceID,
				self.Attrs,
			).Return(nil).Once()
			return store
		},
		Wf: func(t *testing.T, self *testCase) client.WorkflowsOtherAPI {
			rt := client_mocks.NewMockRoundTripper(t)
			cfg := client.NewConfiguration()
			cfg.HTTPClient = &http.Client{Transport: rt}
			return client.NewAPIClient(cfg).WorkflowsOtherAPI
		},
	}, {
		Name: "ok/with audit",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Subject: "f363a096-3ef6-4871-be81-f39ca751d0c0",
			IsUser:  true,
		}),
		DeviceID: "e1ce5c7a-5819-4ee1-aff7-f4fb0b50009c",
		Attrs: model.Attributes{{
			Key:   "key",
			Value: "value",
		}},

		Store: func(t *testing.T, self *testCase) *mstore.DataStore {
			store := new(mstore.DataStore)
			store.On("UpdateConfiguration",
				contextMatcher,
				self.DeviceID,
				self.Attrs,
			).Return(nil).Once()
			return store
		},
		Wf: func(t *testing.T, self *testCase) client.WorkflowsOtherAPI {
			rt := client_mocks.NewMockRoundTripper(t)
			cfg := client.NewConfiguration()
			cfg.HTTPClient = &http.Client{Transport: rt}
			rt.EXPECT().RoundTrip(mock.Anything).
				Run(func(request *http.Request) {
					assert.Equal(t, "emit_auditlog", path.Base(request.URL.Path))
					var body struct {
						Auditlog  model.AuditLog `json:"auditlog"`
						TenantID  string         `json:"tenant_id"`
						RequestID string         `json:"request_id"`
					}
					id := identity.FromContext(request.Context())
					err := json.NewDecoder(request.Body).Decode(&body)
					assert.NoError(t, err)
					al := body.Auditlog
					assert.Equal(t, id.Subject, al.Actor.ID)
					assert.Equal(t, self.DeviceID, al.Object.ID)
					assert.WithinDuration(t, time.Now(), al.EventTS, 5*time.Minute)

				}).
				Return(
					&http.Response{
						StatusCode:    200,
						Header:        http.Header{"Content-Type": []string{"application/json"}},
						ContentLength: int64(len(`{"id": "123","name": "emit_auditlog"}`)),
						Body:          io.NopCloser(bytes.NewReader([]byte(`{"id": "123","name": "emit_auditlog"}`))),
					}, nil,
				).Once()
			return client.NewAPIClient(cfg).WorkflowsOtherAPI
		},
	}, {
		Name: "error/submitting audit",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Subject: "f363a096-3ef6-4871-be81-f39ca751d0c0",
			IsUser:  true,
		}),
		DeviceID: "e1ce5c7a-5819-4ee1-aff7-f4fb0b50009c",
		Attrs: model.Attributes{{
			Key:   "key",
			Value: "value",
		}},

		Store: func(t *testing.T, self *testCase) *mstore.DataStore {
			store := new(mstore.DataStore)
			store.On("UpdateConfiguration",
				contextMatcher,
				self.DeviceID,
				self.Attrs,
			).Return(nil).Once()
			return store
		},
		Wf: func(t *testing.T, self *testCase) client.WorkflowsOtherAPI {
			rt := client_mocks.NewMockRoundTripper(t)
			cfg := client.NewConfiguration()
			cfg.HTTPClient = &http.Client{Transport: rt}
			rt.EXPECT().RoundTrip(mock.MatchedBy(func(r *http.Request) bool {
				assert.Equal(t, "emit_auditlog", path.Base(r.URL.Path))
				return true

			})).Return(&http.Response{
				StatusCode: 500,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Status:     "internal error",
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error": "internal error"}`))),
			}, nil)
			return client.NewAPIClient(cfg).WorkflowsOtherAPI
		},

		Error: errors.New("failed to submit audit log for updating " +
			"the device configuration: internal error"),
	}, {
		Name: "error/internal",

		CTX:      context.Background(),
		DeviceID: "e1ce5c7a-5819-4ee1-aff7-f4fb0b50009c",
		Attrs: model.Attributes{{
			Key:   "key",
			Value: "value",
		}},

		Store: func(t *testing.T, self *testCase) *mstore.DataStore {
			store := new(mstore.DataStore)
			store.On("UpdateConfiguration",
				contextMatcher,
				self.DeviceID,
				self.Attrs,
			).Return(errors.New("internal error")).Once()
			return store
		},
		Wf: func(t *testing.T, self *testCase) client.WorkflowsOtherAPI {
			rt := client_mocks.NewMockRoundTripper(t)
			cfg := client.NewConfiguration()
			cfg.HTTPClient = &http.Client{Transport: rt}
			return client.NewAPIClient(cfg).WorkflowsOtherAPI
		},

		Error: errors.New("internal error"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := tc.Store(t, &tc)
			wf := tc.Wf(t, &tc)
			defer ds.AssertExpectations(t)

			app := New(ds, wf, Config{HaveAuditLogs: true})
			err := app.UpdateConfiguration(tc.CTX, tc.DeviceID, tc.Attrs)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.ErrorContains(t, err, tc.Error.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetConfigurationWithAuditLogs(t *testing.T) {
	const userID = "user-id"

	testCases := map[string]struct {
		err error
	}{
		"ok": {
			err: nil,
		},
		"error": {
			err: errors.New("workflows error"),
		},
	}

	t.Parallel()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			ctx = identity.WithContext(ctx, &identity.Identity{
				Subject: userID,
				IsUser:  true,
			})

			dev := model.NewDevice{
				ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
			}
			configuration := []model.Attribute{
				{
					Key:   "hostname",
					Value: "some0",
				},
			}

			deviceMatcher := mock.MatchedBy(func(d model.Device) bool {
				if !assert.Equal(t, dev.ID, d.ID) {
					return false
				}
				return assert.WithinDuration(t, time.Now(), *d.UpdatedTS, time.Minute)
			})

			ds := new(mstore.DataStore)
			defer ds.AssertExpectations(t)
			ds.On("InsertDevice", ctx, deviceMatcher).Return(nil)
			ds.On("ReplaceConfiguration", ctx, deviceMatcher).Return(nil)

			rt := client_mocks.NewMockRoundTripper(t)
			cfg := client.NewConfiguration()
			cfg.HTTPClient = &http.Client{
				Transport: rt,
			}
			wflows := client.NewAPIClient(cfg).WorkflowsOtherAPI
			roundTrip := rt.EXPECT().RoundTrip(mock.MatchedBy(func(r *http.Request) bool {
				return path.Base(r.URL.Path) == "emit_auditlog"
			}))
			roundTrip.Run(func(request *http.Request) {
				var body struct {
					Auditlog model.AuditLog `json:"auditlog"`
					TenantID string         `json:"tenant_id"`
				}
				err := json.NewDecoder(request.Body).Decode(&body)
				assert.NoError(t, err)
				log := body.Auditlog
				assert.Equal(t, model.ActionSetConfiguration, log.Action)
				assert.Equal(t, model.AuditLogActor{
					ID:   userID,
					Type: model.ActorUser,
				}, log.Actor)
				assert.Equal(t, model.AuditLogObject{
					ID:   dev.ID,
					Type: model.ObjectDevice,
				}, log.Object)
				assert.Equal(t, "{\"hostname\":\"some0\"}", log.Change)
				assert.WithinDuration(t, time.Now(), log.EventTS, time.Minute)
				if tc.err != nil {
					roundTrip.Return(nil, tc.err)
				} else {
					w := httptest.NewRecorder()
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(client.StartWorkflow201Response{
						Id: types.Pointer("123"),
					})
					roundTrip.Return(w.Result(), nil)
				}
			})

			app := New(ds, wflows, Config{HaveAuditLogs: true})
			err := app.ProvisionDevice(ctx, dev)
			assert.NoError(t, err)

			err = app.SetConfiguration(ctx, dev.ID, configuration)
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
			}
		})
	}
}

func TestSetReportedConfiguration(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	dev := model.NewDevice{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
	}
	device := model.Device{
		ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
		ConfiguredAttributes: []model.Attribute{
			{
				Key:   "hostname",
				Value: "some0",
			},
		},
		ReportedAttributes: []model.Attribute{
			{
				Key:   "hostname",
				Value: "some0other",
			},
		},
	}
	deviceMatcher := mock.MatchedBy(func(d model.Device) bool {
		if !assert.Equal(t, dev.ID, d.ID) {
			return false
		}
		return assert.WithinDuration(t, time.Now(), *d.UpdatedTS, time.Minute)
	})
	deviceMatcherReport := mock.MatchedBy(func(d model.Device) bool {
		if !assert.Equal(t, dev.ID, d.ID) {
			return false
		}
		return assert.WithinDuration(t, time.Now(), *d.ReportTS, time.Minute)
	})

	ds := new(mstore.DataStore)
	defer ds.AssertExpectations(t)
	ds.On("InsertDevice", ctx, deviceMatcher).Return(nil)
	ds.On("ReplaceReportedConfiguration", ctx, deviceMatcherReport).Return(nil)
	ds.On("GetDevice", ctx, dev.ID).Return(device, nil)

	app := New(ds, nil, Config{})
	err := app.ProvisionDevice(ctx, dev)
	assert.NoError(t, err)

	err = app.SetReportedConfiguration(ctx, dev.ID, device.ReportedAttributes)
	assert.NoError(t, err)

	d, err := app.GetDevice(ctx, dev.ID)
	assert.NoError(t, err)

	assert.Equal(t, d.ReportedAttributes, device.ReportedAttributes)

	err = app.SetReportedConfiguration(ctx, dev.ID, []model.Attribute{
		{
			Key:   "hostname",
			Value: "other",
		},
	})
	assert.NoError(t, err)

	d, err = app.GetDevice(ctx, dev.ID)
	assert.NoError(t, err)

	assert.NotEqual(t, device.ReportedAttributes, d.ReportedAttributes[0])

	err = app.SetReportedConfiguration(ctx, dev.ID, []model.Attribute{
		{
			Key:   "hostname",
			Value: "",
		},
	})
	assert.NoError(t, err)
}

func TestDeployConfiguration(t *testing.T) {
	t.Parallel()

	const userID = "user-id"

	testCases := map[string]struct {
		device  model.Device
		request model.DeployConfigurationRequest
		err     error
		wfErr   error
		dsErr   error
	}{
		"ok": {},
		"ko, deploy error": {
			err: errors.New("error"),
		},
		"ko, dsErr": {
			dsErr: errors.New("data store error"),
		},
		"ko, wfErr": {
			wfErr: errors.New("workflow error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx = identity.WithContext(ctx, &identity.Identity{
				Tenant:  "tenantID",
				IsUser:  true,
				Subject: userID,
			})

			ds := new(mstore.DataStore)
			defer ds.AssertExpectations(t)

			ds.On("SetDeploymentID",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				tc.device.ID,
				mock.AnythingOfType("uuid.UUID"),
			).Return(tc.dsErr)

			rt := client_mocks.NewMockRoundTripper(t)
			cfg := client.NewConfiguration()
			cfg.HTTPClient = &http.Client{
				Transport: rt,
			}
			wflows := client.NewAPIClient(cfg).WorkflowsOtherAPI

			configuration, _ := tc.device.ConfiguredAttributes.MarshalJSON()
			if tc.dsErr == nil {
				deployConfig := rt.EXPECT().
					RoundTrip(mock.MatchedBy(func(r *http.Request) bool {
						return path.Base(r.URL.Path) == "deploy_device_configuration"
					}))
				deployConfig.Run(func(request *http.Request) {
					var body struct {
						DeviceID      string         `json:"device_id"`
						TenantID      string         `json:"tenant_id"`
						Configuration []byte         `json:"configuration"`
						Retries       uint           `json:"retries"`
						UpdateCtrlMap map[string]any `json:"update_control_map"`
					}
					err := json.NewDecoder(request.Body).Decode(&body)
					require.NoError(t, err)
					assert.Equal(t, body.DeviceID, tc.device.ID)
					assert.Equal(t, body.TenantID, "tenantID")
					assert.Equal(t, body.Retries, tc.request.Retries)
					assert.Equal(t, body.UpdateCtrlMap, tc.request.UpdateControlMap)
					assert.Equal(t, body.Configuration, configuration)
					if tc.err == nil {
						w := httptest.NewRecorder()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(client.StartWorkflow201Response{
							Id: types.Pointer("1234"),
						})
						deployConfig.Return(w.Result(), nil)
					} else {
						deployConfig.Return(nil, tc.err)
					}
				}).
					Once()
			}

			if tc.dsErr == nil && tc.err == nil || tc.wfErr != nil {
				submitLog := rt.EXPECT().RoundTrip(mock.MatchedBy(func(r *http.Request) bool {
					return path.Base(r.URL.Path) == "emit_auditlog"
				}))
				submitLog.Run(func(request *http.Request) {
					var body struct {
						Auditlog model.AuditLog `json:"auditlog"`
						TenantID string         `json:"tenant_id"`
					}
					err := json.NewDecoder(request.Body).Decode(&body)
					assert.NoError(t, err)
					log := body.Auditlog
					assert.Equal(t, model.ActionDeployConfiguration, log.Action)
					assert.Equal(t, model.AuditLogActor{
						ID:   userID,
						Type: model.ActorUser,
					}, log.Actor)
					assert.Equal(t, model.AuditLogObject{
						ID:   tc.device.ID,
						Type: model.ObjectDevice,
					}, log.Object)
					assert.Equal(t, string(configuration), log.Change)
					assert.WithinDuration(t, time.Now(), log.EventTS, time.Minute)
					if tc.wfErr != nil {
						submitLog.Return(nil, tc.wfErr)
					} else {
						w := httptest.NewRecorder()
						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						json.NewEncoder(w).Encode(client.StartWorkflow201Response{
							Id: types.Pointer("123"),
						})
						submitLog.Return(w.Result(), nil)
					}
				})
			}

			app := New(ds, wflows, Config{HaveAuditLogs: true})
			_, err := app.DeployConfiguration(ctx, tc.device, tc.request)
			if tc.err != nil {
				assert.Error(t, err, tc.err)
			} else if tc.wfErr != nil {
				assert.Error(t, err, tc.wfErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
