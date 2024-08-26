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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	mapp "github.com/mendersoftware/mender-server/services/deviceconfig/app/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
)

var contextMatcher = mock.MatchedBy(func(v context.Context) bool {
	return true
})

func TestAlive(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", URIInternal+URIAlive, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Nil(t, w.Body.Bytes())
}

func TestHealth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		Error      error
		StatusCode int
	}{{
		Name: "ok",

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error, from application layer",

		Error:      errors.New("mongo: Connection refused"),
		StatusCode: http.StatusInternalServerError,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := new(mapp.App)
			app.On("HealthCheck",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
			).Return(tc.Error)
			defer app.AssertExpectations(t)
			router := NewRouter(app)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", URIInternal+URIHealth, nil)
			req.Header.Set("X-Men-Requestid", "test")

			router.ServeHTTP(w, req)
			assert.Equal(t, tc.StatusCode, w.Code)
			if tc.Error == nil {
				assert.Nil(t, w.Body.Bytes())
			} else {
				err := rest.Error{
					Err:       tc.Error.Error(),
					RequestID: "test",
				}
				b, _ := json.Marshal(err)
				assert.Equal(t,
					string(b),
					w.Body.String(),
				)
			}
		})
	}
}

func TestProvisionTenant(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		TenantID string
		Request  *http.Request

		App    *mapp.App
		Error  *rest.Error
		Status int
	}{{
		Name: "ok",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"tenant_id": "0123456789abcdef01234567",
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenants,
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("ProvisionTenant",
				contextMatcher,
				model.NewTenant{
					TenantID: "0123456789abcdef01234567",
				},
			).Return(nil)
			return app
		}(),
		Status: http.StatusCreated,
	}, {
		Name: "error bad request body",

		Request: func() *http.Request {
			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenants,
				bytes.NewReader([]byte("tenant_id=foobar")),
			)
			req.Header.Set("X-Men-Requestid", "test")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return req
		}(),

		App: new(mapp.App),
		Error: &rest.Error{
			Err: "malformed request body: invalid character " +
				"'e' in literal true (expecting 'r')",
			RequestID: "test",
		},
		Status: http.StatusBadRequest,
	}, {
		Name: "error invalid request body",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"user_id": uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				),
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenants,
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: new(mapp.App),
		Error: &rest.Error{
			Err:       "invalid request body: tenant_id: cannot be blank.",
			RequestID: "test",
		},
		Status: http.StatusBadRequest,
	}, {
		Name: "error, internal (app) error",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"tenant_id": uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				),
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenants,
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("ProvisionTenant",
				contextMatcher,
				model.NewTenant{
					TenantID: uuid.NewSHA1(
						uuid.NameSpaceDNS, []byte("mender.io"),
					).String(),
				},
			).Return(errors.New("something went wrong!"))
			return app
		}(),
		Error: &rest.Error{
			Err:       http.StatusText(http.StatusInternalServerError),
			RequestID: "test",
		},
		Status: http.StatusInternalServerError,
	}}
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

func TestDeleteTenant(t *testing.T) {
	t.Parallel()
	const tenantID = "123456789012345678901234"

	testCases := []struct {
		Name string

		Request *http.Request

		App    *mapp.App
		Error  *rest.Error
		Status int
	}{{
		Name: "ok",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenant_id", tenantID,
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+URIInternal+repl.Replace(URITenant),
				nil,
			)
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("DeleteTenant",
				contextMatcher,
				tenantID,
			).Return(nil)
			return app
		}(),
		Status: http.StatusNoContent,
	}, {
		Name: "error, internal server error",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenant_id", "123456789012345678901234",
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+URIInternal+repl.Replace(URITenant),
				nil,
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("DeleteTenant",
				contextMatcher,
				tenantID,
			).Return(errors.New("Oh noez!"))
			return app
		}(),
		Error: &rest.Error{
			Err:       http.StatusText(http.StatusInternalServerError),
			RequestID: "test",
		},
		Status: http.StatusInternalServerError,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			defer tc.App.AssertExpectations(t)
			router := NewRouter(tc.App)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tc.Request)
			assert.Equal(t, tc.Status, w.Code)
		})
	}
}

func TestProvisionDevice(t *testing.T) {
	t.Parallel()
	newDeviceMatcher := func(expected string) interface{} {
		return mock.MatchedBy(func(dev model.NewDevice) bool {
			return assert.Equal(t, expected, dev.ID)
		})
	}
	testCases := []struct {
		Name string

		Request *http.Request

		App    *mapp.App
		Error  *rest.Error
		Status int
	}{{
		Name: "ok",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"device_id": uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				),
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenantDevices,
				bytes.NewReader(body),
			)
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("ProvisionDevice",
				contextMatcher,
				newDeviceMatcher(uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String()),
			).Return(nil)
			return app
		}(),
		Status: http.StatusCreated,
	}, {
		Name: "error bad request body",

		Request: func() *http.Request {
			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenantDevices,
				bytes.NewReader([]byte("device_id=foobar")),
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: new(mapp.App),
		Error: &rest.Error{
			Err: "malformed request body: invalid character " +
				"'d' looking for beginning of value",
			RequestID: "test",
		},
		Status: http.StatusBadRequest,
	}, {
		Name: "error invalid request body",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"user_id": uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				),
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenantDevices,
				bytes.NewReader(body),
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: new(mapp.App),
		Error: &rest.Error{
			Err:       "invalid request body: device_id: cannot be blank.",
			RequestID: "test",
		},
		Status: http.StatusBadRequest,
	}, {
		Name: "error, internal (app) error",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"device_id": uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				),
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenantDevices,
				bytes.NewReader(body),
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("ProvisionDevice",
				contextMatcher,
				newDeviceMatcher(uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String()),
			).Return(errors.New("something went wrong!"))
			return app
		}(),
		Error: &rest.Error{
			Err:       http.StatusText(http.StatusInternalServerError),
			RequestID: "test",
		},
		Status: http.StatusInternalServerError,
	}, {
		Name: "error, device already exists",

		Request: func() *http.Request {
			body, _ := json.Marshal(map[string]interface{}{
				"device_id": uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				),
			})

			req, _ := http.NewRequest("POST",
				"http://localhost"+URIInternal+URITenantDevices,
				bytes.NewReader(body),
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("ProvisionDevice",
				contextMatcher,
				newDeviceMatcher(uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String()),
			).Return(store.ErrDeviceAlreadyExists)
			return app
		}(),
		Error: &rest.Error{
			Err:       store.ErrDeviceAlreadyExists.Error(),
			RequestID: "test",
		},
		Status: http.StatusConflict,
	}}
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

func matchCTXIdentity(tenantID string) interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		if id := identity.FromContext(ctx); id != nil {
			return id.Tenant == tenantID
		}
		return false
	})
}

func TestUpdateConfiguration(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		DeviceID string
		TenantID string
		Body     interface{}
		App      func(t *testing.T, self *testCase) *mapp.App

		Code  int
		Error error
	}
	testCases := []testCase{{
		Name: "ok",

		DeviceID: "5526343c-69e4-48a2-9f44-d4542044294b",
		TenantID: "123456789012345678901234",
		Body: model.Attributes{{
			Key:   "key",
			Value: "value",
		}},
		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("UpdateConfiguration",
				matchCTXIdentity(self.TenantID),
				self.DeviceID,
				self.Body.(model.Attributes),
			).Return(nil).
				Once()
			return app
		},
		Code: http.StatusNoContent,
	}, {
		Name: "error/internal",

		DeviceID: "5526343c-69e4-48a2-9f44-d4542044294b",
		TenantID: "123456789012345678901234",
		Body: model.Attributes{{
			Key:   "key",
			Value: "value",
		}},
		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("UpdateConfiguration",
				matchCTXIdentity(self.TenantID),
				self.DeviceID,
				self.Body.(model.Attributes),
			).Return(errors.New("internal error"))
			return app
		},
		Code:  http.StatusInternalServerError,
		Error: errors.New(http.StatusText(http.StatusInternalServerError)),
	}, {
		Name: "error/too many attributes",

		DeviceID: "5526343c-69e4-48a2-9f44-d4542044294b",
		TenantID: "123456789012345678901234",
		Body: func() model.Attributes {
			attrs := make(model.Attributes, model.AttributesMaxLength+1)
			for i := range attrs {
				attrs[i] = model.Attribute{
					Key:   fmt.Sprintf("key%d", i),
					Value: fmt.Sprintf("value%d", i),
				}
			}
			return attrs
		}(),
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},
		Code:  http.StatusBadRequest,
		Error: errors.New("invalid request parameters"),
	}, {
		Name: "error/too many attributes",

		DeviceID: "5526343c-69e4-48a2-9f44-d4542044294b",
		TenantID: "123456789012345678901234",
		Body:     []byte("key:value"),
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},
		Code:  http.StatusBadRequest,
		Error: errors.New("malformed request parameters"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			l := log.NewEmpty()
			l.Logger.Out = io.Discard
			ctx := log.WithContext(context.Background(), l)
			app := tc.App(t, &tc)
			defer app.AssertExpectations(t)
			path := strings.NewReplacer(
				":tenant_id", tc.TenantID,
				":device_id", tc.DeviceID,
			).Replace(URIInternal + URITenant + URIConfiguration)

			var body io.Reader
			switch typ := tc.Body.(type) {
			case []byte:
				body = bytes.NewReader(typ)
			default:
				b, _ := json.Marshal(typ)
				body = bytes.NewReader(b)
			}
			req, _ := http.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"http://localhost:8080"+path,
				body,
			)

			w := httptest.NewRecorder()
			api := NewRouter(app)
			api.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code)
			if tc.Error != nil {
				var err rest.Error
				_ = json.Unmarshal(w.Body.Bytes(), &err)
				assert.Regexp(t, tc.Error.Error(), err.Error())
			} else {
				assert.Empty(t, w.Body.Bytes())
			}
		})
	}
}

func TestDecommissionDevice(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Request *http.Request

		App    *mapp.App
		Error  *rest.Error
		Status int
	}{{
		Name: "ok",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenant_id", "123456789012345678901234",
				":device_id", uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String(),
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+URIInternal+
					repl.Replace(URITenantDevice),
				nil,
			)
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("DecommissionDevice",
				contextMatcher,
				uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String(),
			).Return(nil)
			return app
		}(),
		Status: http.StatusNoContent,
	}, {
		Name: "error device not found",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenant_id", "123456789012345678901234",
				":device_id", uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String(),
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+URIInternal+
					repl.Replace(URITenantDevice),
				bytes.NewReader([]byte("device_id=foobar")),
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("DecommissionDevice",
				contextMatcher,
				uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String(),
			).Return(errors.Wrap(store.ErrDeviceNoExist, "mongo"))
			return app
		}(),
		Error: &rest.Error{
			Err:       store.ErrDeviceNoExist.Error(),
			RequestID: "test",
		},
		Status: http.StatusNotFound,
	}, {
		Name: "error, internal server error",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenant_id", "123456789012345678901234",
				":device_id", uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String(),
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+URIInternal+
					repl.Replace(URITenantDevice),
				bytes.NewReader([]byte("device_id=foobar")),
			)
			req.Header.Set("X-Men-Requestid", "test")
			return req
		}(),

		App: func() *mapp.App {
			app := new(mapp.App)
			app.On("DecommissionDevice",
				contextMatcher,
				uuid.NewSHA1(
					uuid.NameSpaceDNS, []byte("mender.io"),
				).String(),
			).Return(errors.New("Oh noez!"))
			return app
		}(),
		Error: &rest.Error{
			Err:       http.StatusText(http.StatusInternalServerError),
			RequestID: "test",
		},
		Status: http.StatusInternalServerError,
	}}
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

func TestInternalDeployConfiguration(t *testing.T) {
	// Keep the test brief since the rest is covered in management_test.go
	tenantID := "1123456789012345678901234"
	device := model.Device{
		ID: "6aff21b7-7b88-4182-98df-9b9df7787d67",
	}
	deploymentConfig := model.DeployConfigurationResponse{
		DeploymentID: uuid.New(),
	}
	contextMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		id := identity.FromContext(ctx)
		if !assert.NotNil(t, id) {
			return false
		}
		return assert.Equal(t, tenantID, id.Tenant)
	})
	deplReq := model.DeployConfigurationRequest{
		Retries: 1,
	}
	b, _ := json.Marshal(deplReq)

	app := new(mapp.App)
	app.On("GetDevice", contextMatcher, device.ID).
		Return(device, nil).
		On("DeployConfiguration",
			contextMatcher,
			device,
			deplReq,
		).
		Return(deploymentConfig, nil)

	router := NewRouter(app)
	repl := strings.NewReplacer(
		":"+pathParamTenantID,
		tenantID,
		":"+pathParamDeviceID,
		device.ID,
	)
	req, _ := http.NewRequest(
		http.MethodPost,
		URIInternal+repl.Replace(URITenant+URIDeployConfiguration),
		bytes.NewReader(b),
	)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equalf(t,
		http.StatusOK,
		w.Code,
		"unexpected HTTP status code: request body: '%s'",
		w.Body.String(),
	)
}
