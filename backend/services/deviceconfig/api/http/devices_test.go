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
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	mapp "github.com/mendersoftware/mender-server/services/deviceconfig/app/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
)

func ptrNow() *time.Time {
	now := time.Now()
	return &now
}

func TestDevicesSetConfiguration(t *testing.T) {
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

				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIuZGV2aWNlIjp0cnVlLCJtZW5kZXIucGxhbiI6ImVudGVycHJpc2UifQ.OuXQuUrH0T9CucKUwTyDJE7E1yXACVyQeiSjwxqm22Y")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("SetReportedConfiguration",
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
			Name: "error from SetReportedConfiguration",

			Request: func() *http.Request {
				body, _ := json.Marshal(map[string]interface{}{
					"key0": "value0",
				})

				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIuZGV2aWNlIjp0cnVlLCJtZW5kZXIucGxhbiI6ImVudGVycHJpc2UifQ.OuXQuUrH0T9CucKUwTyDJE7E1yXACVyQeiSjwxqm22Y")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				app.On("SetReportedConfiguration",
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

				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
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

				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
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
			Name: "error bad request not a valid json in body",

			Request: func() *http.Request {
				body := []byte("not a valid json text")

				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					bytes.NewReader(body),
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIuZGV2aWNlIjp0cnVlLCJtZW5kZXIucGxhbiI6ImVudGVycHJpc2UifQ.OuXQuUrH0T9CucKUwTyDJE7E1yXACVyQeiSjwxqm22Y")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusBadRequest,
		},
		{
			Name: "error forbidden not a device",

			Request: func() *http.Request {
				req, _ := http.NewRequest("PUT",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIudXNlciI6dHJ1ZSwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.ZeGSy7IaAAdKz7T71A3zml2VSObIAuJSLh3ypT6zd3U")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusForbidden,
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

func TestDevicesGetConfiguration(t *testing.T) {
	t.Parallel()

	device := model.Device{
		ID: uuid.New().String(),
		ConfiguredAttributes: []model.Attribute{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key3",
				Value: "value3",
			},
		},
		ReportedAttributes: []model.Attribute{
			{
				Key:   "key0",
				Value: "value0",
			},
			{
				Key:   "key2",
				Value: "value2",
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
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIuZGV2aWNlIjp0cnVlLCJtZW5kZXIucGxhbiI6ImVudGVycHJpc2UifQ.OuXQuUrH0T9CucKUwTyDJE7E1yXACVyQeiSjwxqm22Y")
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
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
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
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
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
			Name: "error forbidden not a device",

			Request: func() *http.Request {
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIudXNlciI6dHJ1ZSwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.ZeGSy7IaAAdKz7T71A3zml2VSObIAuJSLh3ypT6zd3U")
				return req
			}(),

			App: func() *mapp.App {
				app := new(mapp.App)
				return app
			}(),
			Status: http.StatusForbidden,
		},

		{
			Name: "error internal error on GetDevice",

			Request: func() *http.Request {
				req, _ := http.NewRequest("GET",
					"http://localhost"+URIDevices+URIDeviceConfiguration,
					nil,
				)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJlZWQxNGQ1NS1kOTk2LTQyY2QtODI0OC1lODA2NjYzODEwYTgiLCJtZW5kZXIuZGV2aWNlIjp0cnVlLCJtZW5kZXIucGxhbiI6ImVudGVycHJpc2UifQ.OuXQuUrH0T9CucKUwTyDJE7E1yXACVyQeiSjwxqm22Y")
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
				assert.Equal(t, d, attributes2Map(device.ConfiguredAttributes))
			}
			if tc.Error != nil {
				b, _ := json.Marshal(tc.Error)
				assert.JSONEq(t, string(b), string(w.Body.Bytes()))
			}
		})
	}
}
