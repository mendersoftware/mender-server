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
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/workflows"
	wf_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
	store_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/store/mocks"
)

func TestHealthCheck(t *testing.T) {
	err := errors.New("error")

	store := &store_mocks.DataStore{}
	store.On("Ping",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
	).Return(err)

	app := New(store, nil)

	ctx := context.Background()
	res := app.HealthCheck(ctx)
	assert.Equal(t, err, res)

	store.AssertExpectations(t)
}

func TestProvisionDevice(t *testing.T) {
	err := errors.New("error")
	const tenantID = "1234"
	const deviceID = "abcd"

	store := &store_mocks.DataStore{}
	store.On("ProvisionDevice",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
		tenantID,
		deviceID,
	).Return(err)

	app := New(store, nil)

	ctx := context.Background()
	res := app.ProvisionDevice(ctx, tenantID, &model.Device{ID: deviceID})
	assert.Equal(t, err, res)

	store.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	err := errors.New("error")
	const tenantID = "1234"
	const deviceID = "abcd"

	store := &store_mocks.DataStore{}
	store.On("DeleteDevice",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
		tenantID,
		deviceID,
	).Return(err)

	app := New(store, nil)

	ctx := context.Background()
	res := app.DeleteDevice(ctx, tenantID, deviceID)
	assert.Equal(t, err, res)

	store.AssertExpectations(t)
}

func TestGetDevice(t *testing.T) {
	err := errors.New("error")
	const tenantID = "1234"
	const deviceID = "abcd"
	device := &model.Device{
		ID: deviceID,
	}

	store := &store_mocks.DataStore{}
	store.On("GetDevice",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
		tenantID,
		"not-found",
	).Return(nil, nil)

	store.On("GetDevice",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
		tenantID,
		"error",
	).Return(nil, err)

	store.On("GetDevice",
		mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}),
		tenantID,
		deviceID,
	).Return(device, nil)

	app := New(store, nil)

	ctx := context.Background()
	_, res := app.GetDevice(ctx, tenantID, "error")
	assert.Equal(t, err, res)

	_, res = app.GetDevice(ctx, tenantID, "not-found")
	assert.Equal(t, ErrDeviceNotFound, res)

	dev, res := app.GetDevice(ctx, tenantID, deviceID)
	assert.NoError(t, res)
	assert.Equal(t, dev, device)

	store.AssertExpectations(t)
}

type brokenReader struct{}

func (r brokenReader) Read(b []byte) (int, error) {
	return 0, errors.New("broken reader")
}

func TestPrepareUserSession(t *testing.T) {
	testCases := []struct {
		Name string

		CTX     context.Context
		Session *model.Session

		Rand          io.Reader
		BadParameters bool

		StoreGetDevice    *model.Device
		StoreGetDeviceErr error

		StoreAllocSessErr error

		HaveAuditLogs         bool
		WorkflowsError        error
		StoreDeleteSessionErr error

		Erre error
	}{{
		Name: "ok",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},
		StoreGetDevice: &model.Device{
			ID:     "00000000-0000-0000-0000-000000000000",
			Status: model.DeviceStatusConnected,
		},
		StoreGetDeviceErr: nil,
		StoreAllocSessErr: nil,

		WorkflowsError: nil,
	}, {
		Name: "error, nil session",

		CTX:           context.Background(),
		BadParameters: true,

		Erre: errors.New("nil Session"),
	}, {
		Name: "error, RNG malfunction",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},
		BadParameters: true,

		Rand: brokenReader{},

		Erre: errors.New("^failed to generate session ID: broken reader$"),
	}, {
		Name: "error, RNG malfunction",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
		},
		BadParameters: true,

		Erre: errors.New("^app: cannot create invalid Session: " +
			"start_ts: cannot be blank.$"),
	}, {
		Name: "error, GetDevice internal error",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},

		StoreGetDeviceErr: errors.New("store: internal error"),

		Erre: errors.New("^store: internal error$"),
	}, {
		Name: "error, GetDevice not found",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},

		Erre: ErrDeviceNotFound,
	}, {
		Name: "error, GetDevice disconnected",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},
		StoreGetDevice: &model.Device{
			ID:     "00000000-0000-0000-0000-000000000000",
			Status: model.DeviceStatusDisconnected,
		},
		Erre: ErrDeviceNotConnected,
	}, {
		Name: "error, AllocateSession internal error",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},
		StoreGetDevice: &model.Device{
			ID:     "00000000-0000-0000-0000-000000000000",
			Status: model.DeviceStatusConnected,
		},
		StoreAllocSessErr: errors.New("store: internal error"),
		Erre:              errors.New("store: internal error"),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ds := new(store_mocks.DataStore)
			defer ds.AssertExpectations(t)
			wf := new(wf_mocks.Client)
			defer wf.AssertExpectations(t)
			uuid.SetRand(tc.Rand)
			defer uuid.SetRand(nil)
			app := New(
				ds,
				wf, Config{HaveAuditLogs: tc.HaveAuditLogs},
			)
			if tc.BadParameters {
				goto execTest
			}
			ds.On("GetDevice",
				tc.CTX,
				tc.Session.TenantID,
				tc.Session.DeviceID).
				Return(tc.StoreGetDevice, tc.StoreGetDeviceErr)
			if tc.StoreGetDeviceErr != nil ||
				(tc.StoreGetDeviceErr == nil &&
					tc.StoreGetDevice == nil) ||
				tc.StoreGetDevice.Status !=
					model.DeviceStatusConnected {
				goto execTest
			}
			ds.On("AllocateSession", tc.CTX, tc.Session).
				Return(tc.StoreAllocSessErr)
			if tc.StoreAllocSessErr != nil {
				goto execTest
			}
			if !tc.HaveAuditLogs {
				goto execTest
			}
			ds.On("DeleteSession",
				tc.CTX,
				mock.AnythingOfType("string")).
				Return(tc.Session, tc.StoreDeleteSessionErr)

		execTest:
			err := app.PrepareUserSession(tc.CTX, tc.Session)
			if tc.Erre != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Erre.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogUserSession(t *testing.T) {
	testCases := []struct {
		Name string

		CTX     context.Context
		Session *model.Session

		Rand          io.Reader
		BadParameters bool

		HaveAuditLogs         bool
		WorkflowsError        error
		StoreDeleteSessionErr error

		Erre error
	}{{
		Name: "ok, terminal",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			Types:    []string{model.SessionTypeTerminal},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},

		HaveAuditLogs:  true,
		WorkflowsError: nil,
	}, {
		Name: "ok, port forward",

		CTX: context.Background(),
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			Types:    []string{model.SessionTypePortForward},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},

		HaveAuditLogs:  true,
		WorkflowsError: nil,
	}, {
		Name: "error, SubmitAuditLog http error",

		CTX:           context.Background(),
		HaveAuditLogs: true,
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			Types:    []string{model.SessionTypeTerminal},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},
		WorkflowsError: errors.New("http error"),
		Erre: errors.New(
			"failed to submit audit log: http error",
		),
	}, {
		Name: "error, SubmitAuditLog http error and cleanup error",

		CTX:           context.Background(),
		HaveAuditLogs: true,
		Session: &model.Session{
			DeviceID: "00000000-0000-0000-0000-000000000000",
			UserID:   "00000000-0000-0000-0000-000000000001",
			Types:    []string{model.SessionTypeTerminal},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now(),
		},
		WorkflowsError:        errors.New("http error"),
		StoreDeleteSessionErr: errors.New("store: internal error"),
		Erre: errors.New(
			"failed to submit audit log: http error: failed to clean up " +
				"session state: store: internal error",
		),
	}}

	validateAuditLog := mock.MatchedBy(func(log workflows.AuditLog) bool {
		return assert.NoError(t, log.Validate())
	})

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ds := new(store_mocks.DataStore)
			defer ds.AssertExpectations(t)
			wf := new(wf_mocks.Client)
			defer wf.AssertExpectations(t)
			uuid.SetRand(tc.Rand)
			defer uuid.SetRand(nil)
			app := New(
				ds,
				wf, Config{HaveAuditLogs: tc.HaveAuditLogs},
			)
			if tc.BadParameters {
				goto execTest
			}
			wf.On("SubmitAuditLog",
				tc.CTX,
				validateAuditLog).
				Return(tc.WorkflowsError)
			if tc.WorkflowsError == nil {
				goto execTest
			}
			ds.On("DeleteSession",
				tc.CTX,
				mock.AnythingOfType("string")).
				Return(tc.Session, tc.StoreDeleteSessionErr)

		execTest:
			err := app.LogUserSession(tc.CTX, tc.Session, tc.Session.Types[0])
			if tc.Erre != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Erre.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFreeUserSession(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		SessionID string

		StoreDeleteSession    *model.Session
		StoreDeleteSessionErr error

		HaveAuditLogs bool
		WorkflowsErr  error

		Erre error
	}{{
		Name: "ok",

		SessionID: "00000000-0000-0000-0000-000000000000",

		StoreDeleteSession: &model.Session{
			ID:       "00000000-0000-0000-0000-000000000000",
			DeviceID: "00000000-0000-0000-0000-000000000001",
			UserID:   "00000000-0000-0000-0000-000000000002",
			Types:    []string{model.SessionTypeTerminal},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now().Add(-time.Hour),
		},
	}, {
		Name: "ok, with audit logs",

		SessionID: "00000000-0000-0000-0000-000000000000",

		StoreDeleteSession: &model.Session{
			ID:       "00000000-0000-0000-0000-000000000000",
			DeviceID: "00000000-0000-0000-0000-000000000001",
			UserID:   "00000000-0000-0000-0000-000000000002",
			Types:    []string{model.SessionTypeTerminal},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now().Add(-time.Hour),
		},
		HaveAuditLogs: true,
	}, {
		Name: "ok, with audit logs port forward",

		SessionID: "00000000-0000-0000-0000-000000000000",

		StoreDeleteSession: &model.Session{
			ID:       "00000000-0000-0000-0000-000000000000",
			DeviceID: "00000000-0000-0000-0000-000000000001",
			UserID:   "00000000-0000-0000-0000-000000000002",
			Types:    []string{model.SessionTypePortForward},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now().Add(-time.Hour),
		},
		HaveAuditLogs: true,
	}, {
		Name: "error, store.DeleteSession internal error",

		SessionID: "00000000-0000-0000-0000-000000000000",

		HaveAuditLogs:         true,
		StoreDeleteSessionErr: errors.New("store: internal error"),

		Erre: errors.New("store: internal error$"),
	}, {
		Name: "error, SubmitAuditLogs http error",

		SessionID: "00000000-0000-0000-0000-000000000000",

		StoreDeleteSession: &model.Session{
			ID:       "00000000-0000-0000-0000-000000000000",
			DeviceID: "00000000-0000-0000-0000-000000000001",
			UserID:   "00000000-0000-0000-0000-000000000002",
			Types:    []string{model.SessionTypeTerminal},
			TenantID: "000000000000000000000000",
			StartTS:  time.Now().Add(-time.Hour),
		},
		HaveAuditLogs: true,
		WorkflowsErr:  errors.New("http error"),

		Erre: errors.New("http error$"),
	}}

	workflowsMatcher := func(sess *model.Session) func(workflows.AuditLog) bool {
		return func(wf workflows.AuditLog) bool {
			return true
		}
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := new(store_mocks.DataStore)
			wf := new(wf_mocks.Client)
			defer ds.AssertExpectations(t)
			defer wf.AssertExpectations(t)
			app := New(ds, wf, Config{HaveAuditLogs: tc.HaveAuditLogs})
			ctx := context.Background()

			ds.On("DeleteSession", ctx, tc.SessionID).
				Return(tc.StoreDeleteSession, tc.StoreDeleteSessionErr)
			if tc.StoreDeleteSessionErr != nil || !tc.HaveAuditLogs {
				goto execTest
			}
			wf.On("SubmitAuditLog", ctx,
				mock.MatchedBy(workflowsMatcher(tc.StoreDeleteSession))).
				Return(tc.WorkflowsErr)

		execTest:
			types := []string{}
			if tc.StoreDeleteSession != nil {
				types = tc.StoreDeleteSession.Types
			}
			err := app.FreeUserSession(ctx, tc.SessionID, types)
			if tc.Erre != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Erre.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSessionRecording(t *testing.T) {
	testCases := []struct {
		Name                       string
		DbGetSessionRecordingError error
	}{
		{
			Name: "ok",
		},
		{
			Name:                       "error from the store",
			DbGetSessionRecordingError: errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			sessionId := "00000000-0000-0000-0000-000000000000"
			writer := io.Discard
			store := &store_mocks.DataStore{}
			store.On("WriteSessionRecords",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				sessionId,
				writer,
			).Return(tc.DbGetSessionRecordingError)
			app := New(store, nil)

			ctx := context.Background()
			err := app.GetSessionRecording(ctx, sessionId, writer)
			assert.Equal(t, tc.DbGetSessionRecordingError, err)
		})
	}
}

func TestSaveSessionRecording(t *testing.T) {
	testCases := []struct {
		Name                       string
		DbGetSessionRecordingError error
	}{
		{
			Name: "ok",
		},
		{
			Name:                       "error from the store",
			DbGetSessionRecordingError: errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			sessionId := "00000000-0000-0000-0000-000000000000"
			bytes := []byte("ls -al")
			store := &store_mocks.DataStore{}
			store.On("InsertSessionRecording",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				sessionId,
				bytes,
			).Return(tc.DbGetSessionRecordingError)
			app := New(store, nil)

			ctx := context.Background()
			err := app.SaveSessionRecording(ctx, sessionId, bytes)
			assert.Equal(t, tc.DbGetSessionRecordingError, err)
		})
	}
}

func TestGetRecorder(t *testing.T) {
	testCases := []struct {
		Name                       string
		DbGetSessionRecordingError error
	}{
		{
			Name: "ok",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			sessionId := "00000000-0000-0000-0000-000000000000"
			store := &store_mocks.DataStore{}
			app := New(store, nil)

			ctx := context.Background()
			r := app.GetRecorder(ctx, sessionId)
			assert.NotNil(t, r)
		})
	}
}

func TestDownloadFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		UserID   string
		DeviceID string
		Path     string

		HaveAuditLogs bool
		WorkflowsErr  error

		Err error
	}{
		{
			Name: "ok",

			UserID:   "00000000-0000-0000-0000-000000000000",
			DeviceID: "00000000-0000-0000-0000-000000000000",
			Path:     "/path/to/file",
		},
		{
			Name: "ok, with audit logs",

			UserID:        "00000000-0000-0000-0000-000000000000",
			DeviceID:      "00000000-0000-0000-0000-000000000000",
			Path:          "/path/to/file",
			HaveAuditLogs: true,
		},
		{
			Name: "ko, with audit logs",

			UserID:        "00000000-0000-0000-0000-000000000000",
			DeviceID:      "00000000-0000-0000-0000-000000000000",
			Path:          "/path/to/file",
			HaveAuditLogs: true,
			WorkflowsErr:  errors.New("generic error"),

			Err: errors.New("failed to submit audit log for file transfer: generic error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ds := new(store_mocks.DataStore)
			defer ds.AssertExpectations(t)

			wf := new(wf_mocks.Client)
			defer wf.AssertExpectations(t)

			app := New(ds, wf, Config{HaveAuditLogs: tc.HaveAuditLogs})
			ctx := context.Background()

			if tc.HaveAuditLogs {
				wf.On("SubmitAuditLog",
					ctx,
					mock.AnythingOfType("workflows.AuditLog"),
				).Return(tc.WorkflowsErr)
			}

			err := app.DownloadFile(ctx, tc.UserID, tc.DeviceID, tc.Path)
			if tc.Err != nil {
				assert.EqualError(t, err, tc.Err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		UserID   string
		DeviceID string
		Path     string

		HaveAuditLogs bool
		WorkflowsErr  error

		Err error
	}{
		{
			Name: "ok",

			UserID:   "00000000-0000-0000-0000-000000000000",
			DeviceID: "00000000-0000-0000-0000-000000000000",
			Path:     "/path/to/file",
		},
		{
			Name: "ok, with audit logs",

			UserID:        "00000000-0000-0000-0000-000000000000",
			DeviceID:      "00000000-0000-0000-0000-000000000000",
			Path:          "/path/to/file",
			HaveAuditLogs: true,
		},
		{
			Name: "ko, with audit logs",

			UserID:        "00000000-0000-0000-0000-000000000000",
			DeviceID:      "00000000-0000-0000-0000-000000000000",
			Path:          "/path/to/file",
			HaveAuditLogs: true,
			WorkflowsErr:  errors.New("generic error"),

			Err: errors.New("failed to submit audit log for file transfer: generic error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ds := new(store_mocks.DataStore)
			defer ds.AssertExpectations(t)

			wf := new(wf_mocks.Client)
			defer wf.AssertExpectations(t)

			app := New(ds, wf, Config{HaveAuditLogs: tc.HaveAuditLogs})
			ctx := context.Background()

			if tc.HaveAuditLogs {
				wf.On("SubmitAuditLog",
					ctx,
					mock.AnythingOfType("workflows.AuditLog"),
				).Return(tc.WorkflowsErr)
			}

			err := app.UploadFile(ctx, tc.UserID, tc.DeviceID, tc.Path)
			if tc.Err != nil {
				assert.EqualError(t, err, tc.Err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	t.Parallel()
	gracePeriod := 1 * time.Second

	test := New(nil, nil, Config{})
	test.Shutdown(gracePeriod)
	test.ShutdownDone()

	// verify the channel is closed
	testApp, _ := test.(*app)
	_, ok := <-testApp.shutdownDone
	assert.False(t, ok)
}

func TestShutdownCancels(t *testing.T) {
	t.Parallel()
	gracePeriod := 1 * time.Second

	app := New(nil, nil, Config{})

	// register shutdown cancels
	c1 := false
	app.RegisterShutdownCancel(func() {
		c1 = true
	})

	c2 := false
	app.RegisterShutdownCancel(func() {
		c2 = true
	})

	c3 := false
	id := app.RegisterShutdownCancel(func() {
		c3 = true
	})
	app.UnregisterShutdownCancel(id)

	t1 := time.Now()
	app.Shutdown(gracePeriod)

	assert.True(t, c1)
	assert.True(t, c2)
	assert.False(t, c3)

	elapsed := time.Now().Sub(t1)
	assert.Greater(t, elapsed, gracePeriod)

	app.ShutdownDone()
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

			ds := new(store_mocks.DataStore)
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
