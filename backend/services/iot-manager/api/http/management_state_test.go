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

package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	mapp "github.com/mendersoftware/mender-server/services/iot-manager/app/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

func TestGetDeviceState(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		Headers  http.Header
		DeviceID string

		App func(t *testing.T) *mapp.App

		StatusCode int
		Response   interface{}
	}{
		{
			Name: "ok",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID: "1",

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDevice",
					contextMatcher,
					"1",
				).Return(&model.Device{
					ID:             "1",
					IntegrationIDs: []uuid.UUID{integrationID},
				}, nil)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(&model.DeviceState{
					Desired: map[string]interface{}{
						"key": "value",
					},
				}, nil)
				return mapp
			},

			StatusCode: http.StatusOK,
			Response: model.DeviceStates{
				integrationID.String(): model.DeviceState{
					Desired: map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
		{
			Name: "error, get device",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID: "1",

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDevice",
					contextMatcher,
					"1",
				).Return(nil, errors.New("internal error"))
				return mapp
			},

			StatusCode: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       "internal error",
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, get device not found",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID: "1",

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDevice",
					contextMatcher,
					"1",
				).Return(nil, app.ErrDeviceNotFound)
				return mapp
			},

			StatusCode: http.StatusNotFound,
			Response: rest.Error{
				Err:       app.ErrDeviceNotFound.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, get device state integration",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID: "1",

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDevice",
					contextMatcher,
					"1",
				).Return(&model.Device{
					ID:             "1",
					IntegrationIDs: []uuid.UUID{integrationID},
				}, nil)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(nil, errors.New("internal error"))
				return mapp
			},

			StatusCode: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       "internal error",
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, get device state integration not found",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID: "1",

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDevice",
					contextMatcher,
					"1",
				).Return(&model.Device{
					ID:             "1",
					IntegrationIDs: []uuid.UUID{integrationID},
				}, nil)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(nil, app.ErrIntegrationNotFound)
				return mapp
			},

			StatusCode: http.StatusOK,
			Response:   model.DeviceStates{},
		},
		{
			Name: "error, empty device ID",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID: "",

			StatusCode: http.StatusBadRequest,
			Response: rest.Error{
				Err:       ErrEmptyDeviceID.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, invalid authorization header",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsDevice: true,
					Subject:  "829cbefb-70e7-438f-9ac5-35fd131c2f76",
					Tenant:   "123456789012345678901234",
				})},
			},
			DeviceID: "1",

			StatusCode: http.StatusForbidden,
			Response: rest.Error{
				Err:       ErrMissingUserAuthentication.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var testApp *mapp.App
			if tc.App == nil {
				testApp = new(mapp.App)
			} else {
				testApp = tc.App(t)
			}
			defer testApp.AssertExpectations(t)
			handler := NewRouter(testApp)
			w := httptest.NewRecorder()
			url := strings.Replace(APIURLDeviceState, ":id", tc.DeviceID, 1)
			req, _ := http.NewRequest("GET",
				"http://localhost"+
					APIURLManagement+
					url,
				nil,
			)
			for key := range tc.Headers {
				req.Header.Set(key, tc.Headers.Get(key))
			}

			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.StatusCode, w.Code, "invalid HTTP status code")
			b, _ := json.Marshal(tc.Response)
			assert.JSONEq(t, string(b), w.Body.String())
		})
	}
}

func TestGetDeviceStateIntegration(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		Headers             http.Header
		DeviceID            string
		IntegrationID       uuid.UUID
		IntegrationIDString string

		App func(t *testing.T) *mapp.App

		StatusCode int
		Response   interface{}
	}{
		{
			Name: "ok",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(&model.DeviceState{
					Desired: map[string]interface{}{
						"key": "value",
					},
				}, nil)
				return mapp
			},

			StatusCode: http.StatusOK,
			Response: model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			Name: "error, get device state integration",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(nil, errors.New("internal error"))
				return mapp
			},

			StatusCode: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       "internal error",
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, get device state integration not found",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(nil, app.ErrIntegrationNotFound)
				return mapp
			},

			StatusCode: http.StatusNotFound,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, get device state integration not found",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("GetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
				).Return(nil, app.ErrIntegrationNotFound)
				return mapp
			},

			StatusCode: http.StatusNotFound,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, empty device ID",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "",
			IntegrationID: integrationID,

			StatusCode: http.StatusBadRequest,
			Response: rest.Error{
				Err:       ErrEmptyDeviceID.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, invalid integration ID",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:            "1",
			IntegrationIDString: "2",

			StatusCode: http.StatusBadRequest,
			Response: rest.Error{
				Err:       ErrInvalidIntegrationID.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, invalid authorization header",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsDevice: true,
					Subject:  "829cbefb-70e7-438f-9ac5-35fd131c2f76",
					Tenant:   "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,

			StatusCode: http.StatusForbidden,
			Response: rest.Error{
				Err:       ErrMissingUserAuthentication.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var testApp *mapp.App
			if tc.App == nil {
				testApp = new(mapp.App)
			} else {
				testApp = tc.App(t)
			}
			defer testApp.AssertExpectations(t)
			handler := NewRouter(testApp)
			w := httptest.NewRecorder()
			url := strings.Replace(APIURLDeviceStateIntegration, ":id", tc.DeviceID, 1)
			if tc.IntegrationIDString == "" {
				tc.IntegrationIDString = tc.IntegrationID.String()
			}
			url = strings.Replace(url, ":integrationId", tc.IntegrationIDString, 1)
			req, _ := http.NewRequest("GET",
				"http://localhost"+
					APIURLManagement+
					url,
				nil,
			)
			for key := range tc.Headers {
				req.Header.Set(key, tc.Headers.Get(key))
			}

			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.StatusCode, w.Code, "invalid HTTP status code")
			b, _ := json.Marshal(tc.Response)
			assert.JSONEq(t, string(b), w.Body.String())
		})
	}
}

func TestSetDeviceStateIntegration(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		Headers             http.Header
		DeviceID            string
		IntegrationID       uuid.UUID
		IntegrationIDString string
		RequestBody         interface{}

		App func(t *testing.T) *mapp.App

		StatusCode int
		Response   interface{}
	}{
		{
			Name: "ok",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,
			RequestBody: map[string]interface{}{
				"desired": map[string]string{
					"key": "value",
				},
			},

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("SetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
					&model.DeviceState{
						Desired: map[string]interface{}{
							"key": "value",
						},
					},
				).Return(&model.DeviceState{
					Desired: map[string]interface{}{
						"key": "value",
					},
				}, nil)
				return mapp
			},

			StatusCode: http.StatusOK,
			Response: model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			Name: "error, set device state integration",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,
			RequestBody: map[string]interface{}{
				"desired": map[string]string{
					"key": "value",
				},
			},

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("SetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
					&model.DeviceState{
						Desired: map[string]interface{}{
							"key": "value",
						},
					},
				).Return(nil, errors.New("internal error"))
				return mapp
			},

			StatusCode: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       "internal error",
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, conflict set device state integration",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,
			RequestBody: map[string]interface{}{
				"desired": map[string]string{
					"key": "value",
				},
			},

			App: func(t *testing.T) *mapp.App {
				mapp := new(mapp.App)
				mapp.On("SetDeviceStateIntegration",
					contextMatcher,
					"1",
					integrationID,
					&model.DeviceState{
						Desired: map[string]interface{}{
							"key": "value",
						},
					},
				).Return(nil, app.ErrDeviceStateConflict)
				return mapp
			},

			StatusCode: http.StatusConflict,
			Response: rest.Error{
				Err:       app.ErrDeviceStateConflict.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, invalid payload",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,
			RequestBody:   nil,

			StatusCode: http.StatusBadRequest,
			Response: rest.Error{
				Err:       "malformed request body: invalid request",
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, empty device ID",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:      "",
			IntegrationID: integrationID,

			StatusCode: http.StatusBadRequest,
			Response: rest.Error{
				Err:       ErrEmptyDeviceID.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, invalid integration ID",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},
			DeviceID:            "1",
			IntegrationIDString: "2",

			StatusCode: http.StatusBadRequest,
			Response: rest.Error{
				Err:       ErrInvalidIntegrationID.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, invalid authorization header",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsDevice: true,
					Subject:  "829cbefb-70e7-438f-9ac5-35fd131c2f76",
					Tenant:   "123456789012345678901234",
				})},
			},
			DeviceID:      "1",
			IntegrationID: integrationID,

			StatusCode: http.StatusForbidden,
			Response: rest.Error{
				Err:       ErrMissingUserAuthentication.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var testApp *mapp.App
			if tc.App == nil {
				testApp = new(mapp.App)
			} else {
				testApp = tc.App(t)
			}
			defer testApp.AssertExpectations(t)
			handler := NewRouter(testApp)
			w := httptest.NewRecorder()
			url := strings.Replace(APIURLDeviceStateIntegration, ":id", tc.DeviceID, 1)
			if tc.IntegrationIDString == "" {
				tc.IntegrationIDString = tc.IntegrationID.String()
			}
			url = strings.Replace(url, ":integrationId", tc.IntegrationIDString, 1)
			var body io.Reader
			if tc.RequestBody != nil {
				b, _ := json.Marshal(tc.RequestBody)
				body = bytes.NewReader(b)
			}
			req, _ := http.NewRequest("PUT",
				"http://localhost"+
					APIURLManagement+
					url,
				body,
			)
			for key := range tc.Headers {
				req.Header.Set(key, tc.Headers.Get(key))
			}

			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.StatusCode, w.Code, "invalid HTTP status code")
			b, _ := json.Marshal(tc.Response)
			assert.JSONEq(t, string(b), w.Body.String())
		})
	}
}
