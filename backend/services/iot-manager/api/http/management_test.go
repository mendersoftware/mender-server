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

package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	mapp "github.com/mendersoftware/mender-server/services/iot-manager/app/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

func init() {
	model.SetTrustedHostnames([]string{"localhost", "*.azure-devices.net", "localhost"})
}

var (
	contextMatcher  = mock.MatchedBy(func(_ context.Context) bool { return true })
	validConnString = &model.ConnectionString{
		HostName: "localhost:8080",
		Key:      crypto.String("not-so-secret-key"),
		Name:     "foobar",
	}
)

func compareParameterValues(t *testing.T, expected interface{}) interface{} {
	return mock.MatchedBy(func(actual interface{}) bool {
		return assert.EqualValues(t, expected, actual)
	})
}

func GenerateJWT(id identity.Identity) string {
	JWT := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"alg":"HS256","typ":"JWT"}`),
	)
	b, _ := json.Marshal(id)
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(b)
	hash := hmac.New(sha256.New, []byte("hmac-sha256-secret"))
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(
		hash.Sum([]byte(JWT)),
	)
	return JWT
}

func TestGetIntegrations(t *testing.T) {
	t.Parallel()
	validConnStringString, _ := validConnString.MarshalText()
	testCases := []struct {
		Name string

		Headers http.Header

		App func(t *testing.T) *mapp.App

		StatusCode int
		Response   interface{}
	}{
		{
			Name: "ok",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetIntegrations", contextMatcher).
					Return([]model.Integration{
						{
							ID:       uuid.Nil,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             model.CredentialTypeSAS,
								ConnectionString: validConnString,
							},
						},
					}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response: []map[string]interface{}{{
				"id":       uuid.Nil,
				"provider": model.ProviderIoTHub,
				"credentials": map[string]interface{}{
					"type":              model.CredentialTypeSAS,
					"connection_string": string(validConnStringString),
				},
			}},
		},
		{
			Name: "ok empty integrations",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetIntegrations", contextMatcher).Return([]model.Integration{}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response:   []model.Integration{},
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
			StatusCode: http.StatusForbidden,
			Response: rest.Error{
				Err:       ErrMissingUserAuthentication.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
		{
			Name: "error, failed to retrieve integrations collection",

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

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetIntegrations", contextMatcher).Return(nil, errors.New("error retrieving integrations collection results"))
				return app
			},

			StatusCode: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       "error retrieving integrations collection results",
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
			req, _ := http.NewRequest("GET",
				"http://localhost"+
					APIURLManagement+
					APIURLIntegrations,
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

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))
	var jitter string
	for i := 0; i < 4096; i++ {
		jitter += "1"
	}
	testCases := []struct {
		Name string

		RequestBody interface{}
		RequestHdrs http.Header

		App func(t *testing.T) *mapp.App

		RspCode int
		Error   error
	}{{
		Name: "ok",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
			"credentials": map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App {
			a := new(mapp.App)
			a.On("CreateIntegration",
				contextMatcher,
				mock.AnythingOfType("model.Integration")).
				Return(&model.Integration{
					ID:       integrationID,
					Provider: model.ProviderIoTHub,
					Credentials: model.Credentials{
						Type:             model.CredentialTypeSAS,
						ConnectionString: validConnString,
					},
				}, nil)
			return a
		},

		RspCode: http.StatusCreated,
	}, {
		Name: "duplicate integration",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
			"credentials": map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App {
			a := new(mapp.App)
			a.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
				Return(nil, app.ErrIntegrationExists)
			return a
		},

		RspCode: http.StatusConflict,
		Error:   app.ErrIntegrationExists,
	}, {
		Name: "internal error",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
			"credentials": map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App {
			a := new(mapp.App)
			a.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
				Return(nil, errors.New("internal error"))
			return a
		},

		RspCode: http.StatusInternalServerError,
		Error:   errors.New("internal error"),
	}, {
		Name: "malformed request body",

		RequestBody: map[string]interface{}{
			"provider":    model.ProviderIoTHub,
			"credentials": 1234,
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App { return new(mapp.App) },

		RspCode: http.StatusBadRequest,
		Error:   errors.New("malformed request body: json:"),
	}, {
		Name: "error/forbidden",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
			})},
		},

		App: func(t *testing.T) *mapp.App { return new(mapp.App) },

		RspCode: http.StatusForbidden,
		Error:   ErrMissingUserAuthentication,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t)
			defer app.AssertExpectations(t)
			var body io.Reader
			if tc.RequestBody != nil {
				b, _ := json.Marshal(tc.RequestBody)
				body = bytes.NewReader(b)
			}
			req, _ := http.NewRequest(http.MethodPost,
				"http://localhost"+APIURLManagement+APIURLIntegrations,
				body,
			)
			for k, v := range tc.RequestHdrs {
				req.Header[k] = v
			}

			router := NewRouter(app)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.RspCode, w.Code)
			if tc.Error != nil {
				var erro rest.Error
				if assert.NotNil(t, w.Body) {
					err := json.Unmarshal(w.Body.Bytes(), &erro)
					require.NoError(t, err)
					assert.Regexp(t, tc.Error.Error(), erro.Error())
				}
			} else {
				assert.Empty(t, w.Body.String())
				location := w.Header().Get(hdrLocation)
				expectedLocation := APIURLManagement + "/integrations/" + integrationID.String()
				assert.Equal(t, expectedLocation, location)
			}
		})
	}
}

func TestGetIntegrationById(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))
	type testCase struct {
		Name string

		IntegrationID string
		Header        http.Header
		App           func(t *testing.T, self *testCase) *mapp.App

		Code     int
		Response interface{}
	}

	testCases := []testCase{{
		Name: "ok",

		IntegrationID: integrationID.String(),
		Header: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetIntegrationById", contextMatcher, integrationID).
				Return(self.Response, nil)
			return app
		},

		Code: http.StatusOK,
		Response: &model.Integration{
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
	}, {
		Name: "error/not found",

		IntegrationID: integrationID.String(),
		Header: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
			textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			appie.On("GetIntegrationById", contextMatcher, integrationID).
				Return(nil, app.ErrIntegrationNotFound)
			return appie
		},

		Code: http.StatusNotFound,
		Response: rest.Error{
			Err:       app.ErrIntegrationNotFound.Error(),
			RequestID: "test",
		},
	}, {
		Name: "error/not a uuid",

		IntegrationID: "not-a-uuid",
		Header: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
			textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		Code: http.StatusBadRequest,
		Response: rest.Error{
			Err:       "integration ID must be a valid UUID: invalid UUID length: 10",
			RequestID: "test",
		},
	}, {
		Name: "error/not found",

		IntegrationID: integrationID.String(),
		Header: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
			textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetIntegrationById", contextMatcher, integrationID).
				Return(nil, errors.New("internal error"))
			return app
		},

		Code: http.StatusInternalServerError,
		Response: rest.Error{
			Err:       "internal error",
			RequestID: "test",
		},
	}, {
		Name: "error/not found",

		IntegrationID: integrationID.String(),
		Header: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
				Tenant:  "123456789012345678901234",
			})},
			textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		Code: http.StatusForbidden,
		Response: rest.Error{
			Err:       ErrMissingUserAuthentication.Error(),
			RequestID: "test",
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t, &tc)
			defer app.AssertExpectations(t)
			repl := strings.NewReplacer(":id", tc.IntegrationID)
			req, _ := http.NewRequest(
				http.MethodGet,
				"http://localhost"+APIURLManagement+
					repl.Replace(APIURLIntegration),
				nil,
			)
			for k, v := range tc.Header {
				req.Header[k] = v
			}

			w := httptest.NewRecorder()
			handler := NewRouter(app)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code, "invalid HTTP status code")
			switch expected := tc.Response.(type) {
			case []byte:
				assert.Equal(t, expected, w.Body.Bytes(),
					"HTTP response body does not match expected value",
				)
			default:
				b, err := json.Marshal(expected)
				require.NoError(t, err, "test case error")
				assert.JSONEq(t, string(b), w.Body.String(),
					"HTTP response body does not match expected value",
				)
			}
		})
	}
}

func TestSetIntegrationCredentials(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))

	type testCase struct {
		Name string

		IntegrationID string
		Header        http.Header
		RequestBody   interface{}
		App           func(t *testing.T, self *testCase) *mapp.App

		Code     int
		Response interface{}
		Error    error
	}

	testCases := []testCase{
		{
			Name: "ok",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
			},
			RequestBody: map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				app := new(mapp.App)
				app.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(nil)
				return app
			},

			Code:     http.StatusNoContent,
			Response: nil,
		},
		{
			Name: "error, cannot parse path param",

			IntegrationID: "invalid_uuid",
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			App: func(t *testing.T, self *testCase) *mapp.App { return new(mapp.App) },

			Code: http.StatusBadRequest,
			Response: rest.Error{
				Err:       "integration ID must be a valid UUID: invalid UUID length: 12",
				RequestID: "test",
			},
			Error: errors.New("integration ID must be a valid UUID"),
		},
		{
			Name: "malformed request body",

			IntegrationID: integrationID.String(),
			RequestBody: map[string]interface{}{
				"connection_string": validConnString.String(),
			},
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewString(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
			},

			App: func(t *testing.T, self *testCase) *mapp.App { return new(mapp.App) },

			Code:  http.StatusBadRequest,
			Error: errors.New("malformed request body: type: cannot be blank."),
		},
		{
			Name: "error/forbidden",
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewString(),
					Tenant:  "123456789012345678901234",
				})},
			},

			App: func(t *testing.T, self *testCase) *mapp.App { return new(mapp.App) },

			Code:  http.StatusForbidden,
			Error: ErrMissingUserAuthentication,
		},
		{
			Name: "error, internal server error",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			RequestBody: map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(errors.New("random internal server error"))
				return appie
			},

			Code: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "test",
			},
			Error: errors.New("random internal server error"),
		},
		{
			Name: "error, integration not found",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			RequestBody: map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(app.ErrIntegrationNotFound)
				return appie
			},

			Code: http.StatusNotFound,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "test",
			},
			Error: app.ErrIntegrationNotFound,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t, &tc)
			defer app.AssertExpectations(t)
			repl := strings.NewReplacer(":id", tc.IntegrationID)
			var body io.Reader
			if tc.RequestBody != nil {
				b, _ := json.Marshal(tc.RequestBody)
				body = bytes.NewReader(b)
			}
			req, _ := http.NewRequest(
				http.MethodPut,
				"http://localhost"+APIURLManagement+
					repl.Replace(APIURLIntegration)+"/credentials",
				body,
			)
			for k, v := range tc.Header {
				req.Header[k] = v
			}

			w := httptest.NewRecorder()
			handler := NewRouter(app)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code, "invalid HTTP status code")

			if tc.Error != nil {
				var erro rest.Error
				if assert.NotNil(t, w.Body) {
					err := json.Unmarshal(w.Body.Bytes(), &erro)
					require.NoError(t, err)
					assert.Regexp(t, tc.Error.Error(), erro.Error())
				}
			} else {
				assert.Empty(t, w.Body.Bytes())
			}
		})
	}
}

func TestRemoveIntegration(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))

	type testCase struct {
		Name string

		IntegrationID string
		Header        http.Header
		App           func(t *testing.T, self *testCase) *mapp.App

		Code     int
		Response interface{}
		Error    error
	}

	testCases := []testCase{
		{
			Name:          "error/forbidden",
			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewString(),
					Tenant:  "123456789012345678901234",
				})},
			},

			App: func(t *testing.T, self *testCase) *mapp.App { return new(mapp.App) },

			Code:  http.StatusForbidden,
			Error: ErrMissingUserAuthentication,
		},
		{
			Name: "ok",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("RemoveIntegration", contextMatcher, integrationID).
					Return(nil)
				return appie
			},

			Code:     http.StatusNoContent,
			Response: nil,
		},
		{
			Name: "error, cannot remove integration",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("RemoveIntegration", contextMatcher, integrationID).
					Return(app.ErrCannotRemoveIntegration)
				return appie
			},

			Code: http.StatusConflict,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "test",
			},
			Error: app.ErrCannotRemoveIntegration,
		},
		{
			Name: "error, integration not found",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("RemoveIntegration", contextMatcher, integrationID).
					Return(app.ErrIntegrationNotFound)
				return appie
			},

			Code: http.StatusNotFound,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "test",
			},
			Error: app.ErrIntegrationNotFound,
		},
		{
			Name: "error, internal server error",

			IntegrationID: integrationID.String(),
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("RemoveIntegration", contextMatcher, integrationID).
					Return(errors.New("Internal Server Error"))
				return appie
			},

			Code: http.StatusInternalServerError,
			Response: rest.Error{
				Err:       app.ErrIntegrationNotFound.Error(),
				RequestID: "test",
			},
			Error: errors.New("Internal Server Error"),
		},
		{
			Name: "error, cannot parse path param",

			IntegrationID: "invalid_uuid",
			Header: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					Subject: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}).String(),
					Tenant:  "123456789012345678901234",
					IsUser:  true,
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},
			App: func(t *testing.T, self *testCase) *mapp.App { return new(mapp.App) },

			Code: http.StatusBadRequest,
			Response: rest.Error{
				Err:       "integration ID must be a valid UUID: invalid UUID length: 12",
				RequestID: "test",
			},
			Error: errors.New("integration ID must be a valid UUID"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t, &tc)
			defer app.AssertExpectations(t)
			repl := strings.NewReplacer(":id", tc.IntegrationID)
			req, _ := http.NewRequest(
				http.MethodDelete,
				"http://localhost"+APIURLManagement+
					repl.Replace(APIURLIntegration),
				nil,
			)
			for k, v := range tc.Header {
				req.Header[k] = v
			}

			w := httptest.NewRecorder()
			handler := NewRouter(app)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code, "invalid HTTP status code")

			if tc.Error != nil {
				var erro rest.Error
				if assert.NotNil(t, w.Body) {
					err := json.Unmarshal(w.Body.Bytes(), &erro)
					require.NoError(t, err)
					assert.Regexp(t, tc.Error.Error(), erro.Error())
				}
			} else {
				assert.Empty(t, w.Body.Bytes())
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func validateAuthz(t *testing.T, key []byte, hdr string) bool {
	if !assert.True(t, strings.HasPrefix(hdr, "SharedAccessSignature")) {
		return false
	}
	hdr = strings.TrimPrefix(hdr, "SharedAccessSignature")
	hdr = strings.TrimLeft(hdr, " ")
	q, err := url.ParseQuery(hdr)
	if !assert.NoError(t, err) {
		return false
	}
	for _, key := range []string{"sr", "se", "sig"} {
		if !assert.Contains(t, q, key, "missing signature parameters") {
			return false
		}
	}
	msg := fmt.Sprintf("%s\n%s", url.QueryEscape(q.Get("sr")), q.Get("se"))
	digest := hmac.New(sha256.New, key)
	digest.Write([]byte(msg))
	expected := digest.Sum(nil)
	return assert.Equal(t, base64.StdEncoding.EncodeToString(expected), q.Get("sig"))
}

type neverExpireContext struct {
	context.Context
}

func (neverExpireContext) Deadline() (time.Time, bool) {
	return time.Now().Add(time.Hour), true
}

func TestGetEvents(t *testing.T) {
	t.Parallel()
	integrationId := uuid.New().String()
	testCases := []struct {
		Name string

		Headers http.Header

		Url string

		App func(t *testing.T) *mapp.App

		StatusCode int
		Response   interface{}
	}{
		{
			Name: "ok",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			Url: "http://localhost" + APIURLManagement + APIURLEvents,

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetEvents", contextMatcher, model.EventsFilter{Limit: 20}).
					Return([]model.Event{{
						WebhookEvent: model.WebhookEvent{
							ID:      uuid.Nil,
							Type:    model.EventTypeDeviceProvisioned,
							Data:    model.DeviceEvent{ID: uuid.Nil.String()},
							EventTS: time.Time{},
						},
						DeliveryStatus: []model.DeliveryStatus{{
							IntegrationID: uuid.Nil,
							Success:       true,
							StatusCode: func() *int {
								i := 200
								return &i
							}(),
						}},
					}}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response: []map[string]interface{}{{
				"id":   uuid.Nil,
				"data": map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
				"delivery_statuses": []map[string]interface{}{{
					"integration_id": uuid.Nil,
					"success":        true,
					"status_code":    200,
				}},
				"time": "0001-01-01T00:00:00Z",
				"type": "device-provisioned",
			}},
		},
		{
			Name: "ok with integration id",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			Url: "http://localhost" + APIURLManagement + APIURLEvents + "/" + integrationId,

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetEvents", contextMatcher, model.EventsFilter{Limit: 20, IntegrationID: &integrationId}).
					Return([]model.Event{{
						WebhookEvent: model.WebhookEvent{
							ID:      uuid.MustParse(integrationId),
							Type:    model.EventTypeDeviceProvisioned,
							Data:    model.DeviceEvent{ID: uuid.Nil.String()},
							EventTS: time.Time{},
						},
						DeliveryStatus: []model.DeliveryStatus{{
							IntegrationID: uuid.Nil,
							Success:       true,
							StatusCode: func() *int {
								i := 200
								return &i
							}(),
						}},
					}}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response: []map[string]interface{}{{
				"id":   uuid.MustParse(integrationId),
				"data": map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
				"delivery_statuses": []map[string]interface{}{{
					"integration_id": uuid.Nil,
					"success":        true,
					"status_code":    200,
				}},
				"time": "0001-01-01T00:00:00Z",
				"type": "device-provisioned",
			}},
		},
		{
			Name: "ok, with query params",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			Url: "http://localhost" + APIURLManagement + APIURLEvents + "?page=3&per_page=500",

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetEvents", contextMatcher, model.EventsFilter{Skip: 1000, Limit: 500}).
					Return([]model.Event{{
						WebhookEvent: model.WebhookEvent{
							ID:      uuid.Nil,
							Type:    model.EventTypeDeviceProvisioned,
							Data:    model.DeviceEvent{ID: uuid.Nil.String()},
							EventTS: time.Time{},
						},
						DeliveryStatus: []model.DeliveryStatus{{
							IntegrationID: uuid.Nil,
							Success:       true,
							StatusCode: func() *int {
								i := 200
								return &i
							}(),
						}},
					}}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response: []map[string]interface{}{{
				"id":   uuid.Nil,
				"data": map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
				"delivery_statuses": []map[string]interface{}{{
					"integration_id": uuid.Nil,
					"success":        true,
					"status_code":    200,
				}},
				"time": "0001-01-01T00:00:00Z",
				"type": "device-provisioned",
			}},
		},
		{
			Name: "bad request",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{"test"},
			},

			Url: "http://localhost" + APIURLManagement + APIURLEvents + "?page=foo",

			StatusCode: http.StatusBadRequest,
			Response: map[string]interface{}{
				"error":      "invalid page query: \"foo\"",
				"request_id": "test",
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
			req, _ := http.NewRequest("GET",
				tc.Url,
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
