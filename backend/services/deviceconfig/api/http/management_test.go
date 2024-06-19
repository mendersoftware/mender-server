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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	mapp "github.com/mendersoftware/mender-server/services/deviceconfig/app/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/stretchr/testify/assert"
)

const (
	professionalToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJwcm9mZXNzaW9uYWwifQ.MuqE9VLRSEIIxEK5FsenGi9H5SnaH9nAOKW-RVxZycE"
	enterpriseToken   = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4"
	osToken           = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJvcyJ9.IAkVl6GRbNAameEBozs49UYoFHjthxJxGGd4ZjmgV_Q"
)

func TestSetConfiguration(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		TenantID string
		Request  *http.Request

		App    *mapp.App
		Error  *rest.Error
		Status int
	}{
		{
			Name: "ok",

			Request: func() *http.Request {
				body, _ := json.Marshal(map[string]interface{}{
					"key0": "value0",
				})

				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("SetConfiguration",
					contextMatcher,
					mock.AnythingOfType("string"),
					model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
				).Return(nil)
				return app
			}(),
			Status: http.StatusNoContent,
		},

		{
			Name: "error from SetConfiguration",

			Request: func() *http.Request {
				body, _ := json.Marshal(map[string]interface{}{
					"key0": "value0",
				})

				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("SetConfiguration",
					contextMatcher,
					mock.AnythingOfType("string"),
					model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
				).Return(errors.New("some error"))
				return app
			}(),
			Status: http.StatusInternalServerError,
		},

		{
			Name: "error no auth",

			Request: func() *http.Request {
				body, _ := json.Marshal(map[string]interface{}{
					"configured": []map[string]interface{}{
						{
							"key":   "key0",
							"value": "value0",
						},
					},
				})

				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusUnauthorized,
		},

		{
			Name: "error bad token format",

			Request: func() *http.Request {
				body, _ := json.Marshal(map[string]interface{}{
					"configured": []map[string]interface{}{
						{
							"key":   "key0",
							"value": "value0",
						},
					},
				})

				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusUnauthorized,
		},

		{
			Name: "error url not found",

			Request: func() *http.Request {
				body, _ := json.Marshal(map[string]interface{}{
					"configured": []map[string]interface{}{
						{
							"key":   "key0",
							"value": "value0",
						},
					},
				})

				repl := strings.NewReplacer(
					":device_id", "",
				)
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusNotFound,
		},

		{
			Name: "error bad request not a valid json in body",

			Request: func() *http.Request {
				body := []byte("not a valid json text")

				repl := strings.NewReplacer(
					":device_id", "id",
				)
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusBadRequest,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			defer tc.App.AssertExpectations(t)
			router := NewRouter(tc.App)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tc.Request)
			assert.Equal(t, tc.Status, w.Code)
			if tc.Error != nil {
				b, _ := json.Marshal(tc.Error)
				assert.JSONEq(t, string(b), string(w.Body.Bytes()))
			}
		})
	}
}

func TestGetConfiguration(t *testing.T) {
	t.Parallel()

	device := model.Device{
		ID: uuid.New().String(),
		ConfiguredAttributes: []model.Attribute{
			{
				Key:   "key0",
				Value: "value0",
			},
			{
				Key:   "key2",
				Value: "value2",
			},
		},
		ReportedAttributes: []model.Attribute{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key3",
				Value: "value3",
			},
		},
		UpdatedTS: ptrNow(),
		ReportTS:  ptrNow(),
	}

	testCases := []struct {
		Name string

		TenantID string
		Request  *http.Request

		App    *mapp.App
		Error  *rest.Error
		Status int
	}{
		{
			Name: "ok",

			Request: func() *http.Request {
				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("GetDevice",
					contextMatcher,
					mock.AnythingOfType("string"),
				).Return(device, nil)
				return app
			}(),
			Status: http.StatusOK,
		},

		{
			Name: "error no auth",

			Request: func() *http.Request {
				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusUnauthorized,
		},

		{
			Name: "error bad token format",

			Request: func() *http.Request {
				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusUnauthorized,
		},

		{
			Name: "error url not found",

			Request: func() *http.Request {
				repl := strings.NewReplacer(
					":device_id", "",
				)
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusNotFound,
		},

		{
			Name: "error device not found",

			Request: func() *http.Request {
				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("GetDevice",
					contextMatcher,
					mock.AnythingOfType("string"),
				).Return(device, store.ErrDeviceNoExist)
				return app
			}(),
			Status: http.StatusNotFound,
		},

		{
			Name: "error internal error on GetDevice",

			Request: func() *http.Request {
				repl := strings.NewReplacer(
					":device_id", uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				)
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIManagement+repl.Replace(URIConfiguration),
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("GetDevice",
					contextMatcher,
					mock.AnythingOfType("string"),
				).Return(device, errors.New("some other error"))
				return app
			}(),
			Status: http.StatusInternalServerError,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			defer tc.App.AssertExpectations(t)
			router := NewRouter(tc.App)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tc.Request)
			assert.Equal(t, tc.Status, w.Code)
			if w.Code == http.StatusOK {
				var d map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &d)
				t.Logf("got: %+v", d)
				assert.Equal(t, d["configured"], attributes2Map(device.ConfiguredAttributes))
			}
			if tc.Error != nil {
				b, _ := json.Marshal(tc.Error)
				assert.JSONEq(t, string(b), string(w.Body.Bytes()))
			}
		})
	}
}

func TestDeployConfiguration(t *testing.T) {
	t.Parallel()

	deviceID := uuid.New().String()

	testCases := map[string]struct {
		deviceID                string
		device                  model.Device
		requestBody             string
		getDeviceErr            error
		callGetDevice           bool
		deployConfiguration     model.DeployConfigurationResponse
		deployConfigurationErr  error
		callDeployConfiguration bool
		areDevicesInGroupRsp    bool
		areDevicesInGroupError  error
		token                   string
		status                  int
	}{
		"ok": {
			deviceID: deviceID,
			device: model.Device{
				ID: deviceID,
				ConfiguredAttributes: []model.Attribute{
					{
						Key:   "key0",
						Value: "value0",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
				ReportedAttributes: []model.Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key3",
						Value: "value3",
					},
				},
				UpdatedTS: ptrNow(),
				ReportTS:  ptrNow(),
			},
			requestBody: "{\"retries\": 0}",
			deployConfiguration: model.DeployConfigurationResponse{
				DeploymentID: uuid.New(),
			},
			callGetDevice:           true,
			callDeployConfiguration: true,
			status:                  200,
		},
		"ok, update control map": {
			deviceID: deviceID,
			device: model.Device{
				ID: deviceID,
				ConfiguredAttributes: []model.Attribute{
					{
						Key:   "key0",
						Value: "value0",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
				ReportedAttributes: []model.Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key3",
						Value: "value3",
					},
				},
				UpdatedTS: ptrNow(),
				ReportTS:  ptrNow(),
			},
			requestBody: "{\"retries\": 0, \"update_contrlo_map\":\"bar\", \"configuration\":{\"foo\":\"bar\"}}",
			deployConfiguration: model.DeployConfigurationResponse{
				DeploymentID: uuid.New(),
			},
			callGetDevice:           true,
			callDeployConfiguration: true,
			status:                  200,
		},
		"ko, update control map": {
			deviceID: deviceID,
			device: model.Device{
				ID: deviceID,
				ConfiguredAttributes: []model.Attribute{
					{
						Key:   "key0",
						Value: "value0",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
				ReportedAttributes: []model.Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key3",
						Value: "value3",
					},
				},
				UpdatedTS: ptrNow(),
				ReportTS:  ptrNow(),
			},
			requestBody: "{\"retries\": 3, \"update_control_map\":{\"foo\":\"bar\"}}",
			deployConfiguration: model.DeployConfigurationResponse{
				DeploymentID: uuid.New(),
			},
			callGetDevice:           true,
			callDeployConfiguration: false,
			token:                   osToken,
			status:                  403,
		},
		"ko, error in DeployConfiguration": {
			deviceID: deviceID,
			device: model.Device{
				ID: deviceID,
				ConfiguredAttributes: []model.Attribute{
					{
						Key:   "key0",
						Value: "value0",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
				ReportedAttributes: []model.Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key3",
						Value: "value3",
					},
				},
				UpdatedTS: ptrNow(),
				ReportTS:  ptrNow(),
			},
			requestBody:             "{\"retries\": 0}",
			deployConfigurationErr:  errors.New("generic error"),
			callGetDevice:           true,
			callDeployConfiguration: true,
			status:                  500,
		},
		"ko, device not found": {
			deviceID:      deviceID,
			getDeviceErr:  store.ErrDeviceNoExist,
			callGetDevice: true,
			status:        404,
		},
		"ko, error in GetDevice": {
			deviceID:      deviceID,
			requestBody:   "",
			getDeviceErr:  errors.New("generic error"),
			callGetDevice: true,
			status:        500,
		},
		"ko, bad request body": {
			deviceID:    deviceID,
			requestBody: "",
			device: model.Device{
				ID: deviceID,
				ConfiguredAttributes: []model.Attribute{
					{
						Key:   "key0",
						Value: "value0",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
				ReportedAttributes: []model.Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key3",
						Value: "value3",
					},
				},
				UpdatedTS: ptrNow(),
				ReportTS:  ptrNow(),
			},
			callGetDevice: true,
			status:        400,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			app := new(mapp.App)
			defer app.AssertExpectations(t)
			if tc.callGetDevice {
				app.On("GetDevice",
					contextMatcher,
					mock.AnythingOfType("string"),
				).Return(tc.device, tc.getDeviceErr)
			}

			if tc.callDeployConfiguration {
				app.On("DeployConfiguration",
					contextMatcher,
					tc.device,
					mock.AnythingOfType("model.DeployConfigurationRequest"),
				).Return(tc.deployConfiguration, tc.deployConfigurationErr)
			}

			router := NewRouter(app)

			repl := strings.NewReplacer(
				":device_id", tc.deviceID,
			)

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIManagement+repl.Replace(URIDeployConfiguration),
				bytes.NewReader([]byte(tc.requestBody)),
			)
			req.Header.Set("Content-Type", "application/json")
			if len(tc.token) > 0 {
				req.Header.Set("Authorization", tc.token)
			} else {
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func attributes2Map(attributes []model.Attribute) map[string]interface{} {
	configurationMap := make(map[string]interface{}, len(attributes))
	for _, a := range attributes {
		switch a.Value.(type) {
		case string:
			configurationMap[a.Key] = a.Value.(string)
		}
	}

	return configurationMap
}
