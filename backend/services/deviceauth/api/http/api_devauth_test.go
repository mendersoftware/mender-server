// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package http

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"

	mt "github.com/mendersoftware/mender-server/pkg/testing"

	"github.com/mendersoftware/mender-server/services/deviceauth/devauth"
	"github.com/mendersoftware/mender-server/services/deviceauth/devauth/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	mtest "github.com/mendersoftware/mender-server/services/deviceauth/utils/testing"
)

func RestError(status string) string {
	msg, _ := json.Marshal(map[string]interface{}{"error": status, "request_id": "test"})
	return string(msg)
}

func runTestRequest(t *testing.T, handler http.Handler, req *http.Request, code int, body string) *httptest.ResponseRecorder {
	req.Header.Add(requestid.RequestIdHeader, "test")
	r := httptest.NewRecorder()
	handler.ServeHTTP(r, req)
	assert.Equal(t, code, r.Code)
	if body != "" {
		assert.Equal(t, body, r.Body.String())
	} else {
		assert.Empty(t, r.Body.String())
	}
	return r
}

func runSimpleTestRequest(t *testing.T, handler http.Handler, req *http.Request) *mt.Recorded {
	req.Header.Add(requestid.RequestIdHeader, "test")
	r := httptest.NewRecorder()
	handler.ServeHTTP(r, req)
	return &mt.Recorded{
		T:        t,
		Recorder: r,
	}
}

func makeMockApiHandler(t *testing.T, da devauth.App, db store.DataStore) http.Handler {
	t.Helper()
	router := NewRouter(da, db)
	assert.NotNil(t, router)
	return router
}

// create an auth req that's optionally:
// - signed with an actual key
// - signed with a bogus test value
// - not signed at all
func makeAuthReq(payload interface{}, key crypto.PrivateKey, signature string, t *testing.T) *http.Request {
	r := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "POST",
		Path:   "http://localhost/api/devices/v1/authentication/auth_requests",
		Body:   payload,
		Auth:   true,
	})

	b, err := json.Marshal(payload)
	if err != nil {
		t.FailNow()
	}

	if signature != "" {
		r.Header.Set(HdrAuthReqSign, signature)
	} else if key != nil {
		sign := mtest.AuthReqSign(b, key, t)
		r.Header.Set(HdrAuthReqSign, string(sign))
	}

	return r
}

func TestNewRouter_EntityTooLarge(t *testing.T) {
	da := &mocks.App{}
	apih := NewRouter(da, nil, SetMaxRequestSize(1024))
	body := bytes.NewReader(bytes.Repeat([]byte("4KiB"), 1024))
	req, _ := http.NewRequest("POST", "http://localhost/", body)
	err := &http.MaxBytesError{Limit: 1024}
	b, _ := json.Marshal(rest.Error{Err: err.Error(), RequestID: "test"})
	runTestRequest(t, apih, req, http.StatusRequestEntityTooLarge, string(b))
}

func TestAliveHandler(t *testing.T) {
	da := &mocks.App{}
	apih := makeMockApiHandler(t, da, nil)
	req, _ := http.NewRequest("GET", "http://localhost"+apiUrlInternalV1+uriAlive, nil)
	runTestRequest(t, apih, req, http.StatusNoContent, "")
}

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		Name string

		AppError     error
		ResponseCode int
		ResponseBody string
	}{{
		Name:         "ok",
		ResponseCode: http.StatusNoContent,
	}, {
		Name: "error, service unhealthy",

		AppError:     errors.New("connection error"),
		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: RestError("connection error"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			uadm := &mocks.App{}
			uadm.On("HealthCheck", mock.MatchedBy(
				func(ctx interface{}) bool {
					if _, ok := ctx.(context.Context); ok {
						return true
					}
					return false
				},
			)).Return(tc.AppError)

			api := makeMockApiHandler(t, uadm, nil)
			req, _ := http.NewRequest(
				"GET",
				"http://localhost"+apiUrlInternalV1+uriHealth,
				nil,
			)
			req.Header.Set("X-MEN-RequestID", "test")
			runTestRequest(t, api, req, tc.ResponseCode, tc.ResponseBody)
		})
	}
}

func TestApiDevAuthSubmitAuthReq(t *testing.T) {
	t.Parallel()

	privkey := mtest.LoadPrivKey("testdata/private.pem")
	pubkeyStr := mtest.LoadPubKeyStr("testdata/public.pem")

	testCases := []struct {
		req *http.Request

		devAuthToken string
		devAuthErr   error

		code int
		body string
	}{
		{
			//empty body
			makeAuthReq(nil, nil, "dontcare", t),
			"",
			nil,
			400,
			RestError("failed to decode auth request: empty request body"),
		},
		{
			//incomplete body
			makeAuthReq(
				map[string]interface{}{
					"pubkey":       pubkeyStr,
					"tenant_token": "tenant-0001",
				},
				privkey,
				"",
				t),
			"",
			nil,
			400,
			RestError("invalid auth request: id_data must be provided"),
		},
		{
			//incomplete body
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `{"sn":"0001"}`,
					"tenant_token": "tenant-0001",
				},
				privkey,
				"",
				t),
			"",
			nil,
			400,
			RestError("invalid auth request: pubkey must be provided"),
		},
		{
			//complete body, missing signature header
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `{"sn":"0001"}`,
					"pubkey":       pubkeyStr,
					"tenant_token": "tenant-0001",
				},
				nil,
				"",
				t),
			"",
			nil,
			400,
			RestError("missing request signature header"),
		},
		{
			//complete body, invalid signature header
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `{"sn":"0001"}`,
					"pubkey":       pubkeyStr,
					"tenant_token": "tenant-0001",
				},
				nil,
				"invalidsignature",
				t),
			"",
			nil,
			401,
			RestError("signature verification failed"),
		},
		{
			//complete body + signature, auth error
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `{"sn":"0001"}`,
					"pubkey":       pubkeyStr,
					"tenant_token": "tenant-0001",
				},
				privkey,
				"",
				t),
			"",
			devauth.MakeErrDevAuthUnauthorized(
				errors.New("account suspended"),
			),
			401,
			RestError("account suspended"),
		},
		{
			//invalid id data (not json)
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `"sn":"0001"`,
					"pubkey":       pubkeyStr,
					"tenant_token": "tenant-0001",
				},
				privkey,
				"",
				t),
			"",
			nil,
			400,
			RestError("invalid auth request: invalid character ':' after top-level value"),
		},
		{
			//complete body + signature, auth ok
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `{"sn":"0001"}`,
					"pubkey":       pubkeyStr,
					"tenant_token": "tenant-0001",
				},
				privkey,
				"",
				t),
			rtest.DEFAULT_AUTH,
			nil,
			200,
			rtest.DEFAULT_AUTH,
		},
		{
			//complete body + signature, auth ok, tenant token empty
			makeAuthReq(
				map[string]interface{}{
					"id_data": `{"sn":"0001"}`,
					"pubkey":  pubkeyStr,
				},
				privkey,
				"",
				t),
			rtest.DEFAULT_AUTH,
			nil,
			200,
			rtest.DEFAULT_AUTH,
		},
		{
			//complete body, invalid public key
			makeAuthReq(
				map[string]interface{}{
					"id_data":      `{"sn":"0001"}`,
					"pubkey":       "invalid",
					"tenant_token": "tenant-0001",
				},
				privkey,
				"",
				t),
			rtest.DEFAULT_AUTH,
			nil,
			400,
			RestError("invalid auth request: cannot decode public key"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			da := &mocks.App{}
			da.On("SubmitAuthRequest",
				mtest.ContextMatcher(),
				mock.AnythingOfType("*model.AuthReq")).
				Return(
					func(_ context.Context, r *model.AuthReq) string {
						if tc.devAuthErr != nil {
							return ""
						}
						return tc.devAuthToken
					},
					tc.devAuthErr)

			apih := makeMockApiHandler(t, da, nil)

			recorded := runTestRequest(t, apih, tc.req, tc.code, tc.body)
			if tc.code == http.StatusOK {
				assert.Equal(t, "application/jwt",
					recorded.Result().Header.Get("Content-Type"))
			}
		})
	}
}

// Custom checker for the Location header in a preauth response
type DevicePreauthReturnID struct {
	mt.JSONResponse
}

func NewJSONResponseIDChecker(status int, headers map[string]string, body interface{}) *DevicePreauthReturnID {
	return &DevicePreauthReturnID{
		mt.JSONResponse{
			BaseResponse: mt.BaseResponse{
				Status:      status,
				ContentType: "application/json",
				Headers:     headers,
				Body:        body,
			},
		},
	}
}

func (d *DevicePreauthReturnID) CheckHeaders(t *testing.T, recorded *mt.Recorded) {
	assert.Contains(t, recorded.Recorder.Result().Header, "Location")
	assert.Contains(t, recorded.Recorder.Result().Header["Location"][0], "devices/")
}

func TestApiV2DevAuthPreauthDevice(t *testing.T) {
	t.Parallel()

	pubkeyStr := mtest.LoadPubKeyStr("testdata/public.pem")

	type brokenPreAuthReq struct {
		IdData string `json:"identity_data"`
		PubKey string `json:"pubkey"`
	}

	testCases := map[string]struct {
		body interface{}

		devAuthErr error
		outDev     *model.Device

		callApp bool

		checker mt.ResponseChecker
	}{
		"ok": {
			body: &preAuthReq{
				IdData: map[string]interface{}{
					"sn": "0001",
				},
				PubKey: pubkeyStr,
			},
			callApp: true,
			checker: mt.NewJSONResponse(
				http.StatusCreated,
				nil,
				nil),
		},
		"ok - verify Location header": {
			body: &preAuthReq{
				IdData: map[string]interface{}{
					"sn": "0001",
				},
				PubKey: pubkeyStr,
			},
			callApp: true,
			checker: NewJSONResponseIDChecker(
				http.StatusCreated,
				map[string]string{"Location": "devices/somegeneratedid"},
				nil),
		},
		"invalid: id data is not json": {
			body: &brokenPreAuthReq{
				IdData: `"sn":"0001"`,
				PubKey: pubkeyStr,
			},
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("failed to decode preauth request: json: cannot unmarshal string into Go struct field preAuthReq.identity_data of type map[string]interface {}")),
		},
		"invalid: no id data": {
			body: &preAuthReq{
				PubKey: pubkeyStr,
			},
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("failed to decode preauth request: identity_data: cannot be blank.")),
		},
		"invalid: no pubkey": {
			body: &preAuthReq{
				IdData: map[string]interface{}{
					"sn": "0001",
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("failed to decode preauth request: pubkey: cannot be blank.")),
		},
		"invalid: no body": {
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("failed to decode preauth request: empty request body")),
		},
		"invalid public key": {
			body: &preAuthReq{
				IdData: map[string]interface{}{
					"sn": "0001",
				},
				PubKey: "invalid",
			},
			devAuthErr: devauth.ErrDeviceExists,
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("failed to decode preauth request: cannot decode public key")),
		},
		"devauth: device exists": {
			body: &preAuthReq{
				IdData: map[string]interface{}{
					"sn": "0001",
				},
				PubKey: pubkeyStr,
			},
			devAuthErr: devauth.ErrDeviceExists,
			outDev:     &model.Device{Id: "foo"},
			callApp:    true,
			checker: mt.NewJSONResponse(
				http.StatusConflict,
				nil,
				model.Device{Id: "foo"}),
		},
		"devauth: generic error": {
			body: &preAuthReq{
				IdData: map[string]interface{}{
					"sn": "0001",
				},
				PubKey: pubkeyStr,
			},
			callApp:    true,
			devAuthErr: errors.New("generic"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				restError("internal error")),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("tc %s", name), func(t *testing.T) {
			da := &mocks.App{}
			if tc.callApp {
				da.On("PreauthorizeDevice",
					mtest.ContextMatcher(),
					mock.AnythingOfType("*model.PreAuthReq")).
					Return(tc.outDev, tc.devAuthErr)
			}

			apih := makeMockApiHandler(t, da, nil)
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/management/v2/devauth/devices",
				Body:   tc.body,
				Auth:   true,
			})
			recorded := runSimpleTestRequest(t, apih, req)

			mt.CheckHTTPResponse(t, tc.checker, recorded)
			da.AssertExpectations(t)
		})
	}
}

func TestApiV2DevAuthUpdateStatusDevice(t *testing.T) {
	t.Parallel()

	devs := map[string]struct {
		dev *model.Device
		err error
	}{
		"123,456": {
			dev: &model.Device{
				Id:     "foo",
				Status: "accepted",
				IdData: "deadcafe",
			},
			err: nil,
		},
		"234,567": {
			dev: nil,
			err: devauth.ErrDevIdAuthIdMismatch,
		},
		"345,678": {
			dev: nil,
			err: errors.New("processing failed"),
		},
		"567,890": {
			dev: &model.Device{
				Id:     "foo",
				Status: "pending",
				IdData: "deadcafe",
			},
			err: devauth.ErrMaxDeviceCountReached,
		},
	}

	mockaction := func(_ context.Context, dev_id string, auth_id string) error {
		d, ok := devs[dev_id+","+auth_id]
		if ok == false {
			return store.ErrDevNotFound
		}
		if d.err != nil {
			return d.err
		}
		return nil
	}
	da := &mocks.App{}
	da.On("AcceptDeviceAuth",
		mtest.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(mockaction)
	da.On("RejectDeviceAuth",
		mtest.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(mockaction)
	da.On("ResetDeviceAuth",
		mtest.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(mockaction)

	apih := makeMockApiHandler(t, da, nil)

	accstatus := DevAuthApiStatus{"accepted"}
	rejstatus := DevAuthApiStatus{"rejected"}
	penstatus := DevAuthApiStatus{"pending"}

	tcases := []struct {
		req  *http.Request
		code int
		body string
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/123/auth/456/status",
				Auth:   true,
			}),
			code: http.StatusBadRequest,
			body: RestError("failed to decode status data: invalid request"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/123/auth/456/status",
				Auth:   true,
				Body:   DevAuthApiStatus{"foo"},
			}),
			code: http.StatusBadRequest,
			body: RestError("incorrect device status"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/123/auth/456/status",
				Auth:   true,
				Body:   accstatus,
			}),
			code: http.StatusNoContent,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/345/auth/678/status",
				Auth:   true,
				Body:   accstatus,
			}),
			code: http.StatusInternalServerError,
			body: RestError("internal error"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/999/auth/123/status",
				Auth:   true,
				Body:   accstatus,
			}),
			code: http.StatusNotFound,
			body: RestError(store.ErrDevNotFound.Error()),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/123/auth/456/status",
				Auth:   true,
				Body:   rejstatus,
			}),
			code: http.StatusNoContent,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/123/auth/456/status",
				Auth:   true,
				Body:   penstatus,
			}),
			code: http.StatusNoContent,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/234/auth/567/status",
				Auth:   true,
				Body:   penstatus,
			}),
			code: http.StatusBadRequest,
			body: RestError("dev auth: dev ID and auth ID mismatch"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/management/v2/devauth/devices/567/auth/890/status",
				Auth:   true,
				Body:   accstatus,
			}),
			code: http.StatusUnprocessableEntity,
			body: RestError("maximum number of accepted devices reached"),
		},
	}

	for idx := range tcases {
		tc := tcases[idx]
		t.Run(fmt.Sprintf("tc %d", idx), func(t *testing.T) {
			t.Parallel()

			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}

}

func TestApiDevAuthVerifyToken(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req     *http.Request
		code    int
		body    string
		headers map[string]string
		err     error
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: http.StatusUnauthorized,
			body: RestError(ErrNoAuthHeader.Error()),
			err:  nil,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: 200,
			headers: map[string]string{
				"authorization": rtest.DEFAULT_AUTH,
			},
			err: nil,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: 200,
			headers: map[string]string{
				"authorization": rtest.DEFAULT_AUTH,
			},
			err: nil,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: http.StatusForbidden,
			headers: map[string]string{
				"authorization": rtest.DEFAULT_AUTH,
			},
			err: jwt.ErrTokenExpired,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: http.StatusUnauthorized,
			headers: map[string]string{
				"authorization": rtest.DEFAULT_AUTH,
			},
			err: jwt.ErrTokenInvalid,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: http.StatusUnauthorized,
			headers: map[string]string{
				"authorization": rtest.DEFAULT_AUTH,
			},
			err: store.ErrAuthSetNotFound,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tokens/verify",
			}),
			code: 500,
			body: RestError("internal error"),
			headers: map[string]string{
				"authorization": rtest.DEFAULT_AUTH,
			},
			err: errors.New("some error that will only be logged"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("VerifyToken",
				mtest.ContextMatcher(),
				mock.AnythingOfType("string")).
				Return(tc.err)

			apih := makeMockApiHandler(t, da, nil)
			if len(tc.headers) > 0 {
				tc.req.Header.Set("authorization", tc.headers["authorization"])
			}

			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}

}

func TestApiV2DevAuthDeleteToken(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req  *http.Request
		code int
		body string
		err  error
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/tokens/foo",
				Auth:   true,
			}),
			code: http.StatusNoContent,
			err:  nil,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/tokens/foo",
				Auth:   true,
			}),
			code: http.StatusNotFound,
			err:  store.ErrTokenNotFound,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/tokens/foo",
				Auth:   true,
			}),
			code: http.StatusInternalServerError,
			body: RestError("internal error"),
			err:  errors.New("some error that will only be logged"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("RevokeToken",
				mtest.ContextMatcher(),
				mock.AnythingOfType("string")).
				Return(tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}

}

func TestApiV2GetDevice(t *testing.T) {
	t.Parallel()

	dev := &model.Device{
		Id:     "foo",
		IdData: `{"mac": "00:00:00:01"}`,
		IdDataStruct: map[string]interface{}{
			"mac": "00:00:00:01",
		},
		Status: model.DevStatusPending,
		AuthSets: []model.AuthSet{
			{
				Id:       "1",
				DeviceId: "foo",
				IdData:   `{"mac": "00:00:00:01"}`,
				IdDataStruct: map[string]interface{}{
					"mac": "00:00:00:01",
				},
			},
		},
	}

	tcases := []struct {
		req *http.Request

		device *model.Device
		err    error

		code int
		body string
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo",
				Auth:   true,
			}),
			device: dev,
			err:    nil,

			code: http.StatusOK,
			body: string(asJSON(dev)),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices/bar",
				Auth:   true,
			}),
			device: nil,
			err:    store.ErrDevNotFound,

			code: http.StatusNotFound,
			body: RestError("device not found"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices/bar",
				Auth:   true,
			}),
			device: nil,
			err:    errors.New("generic error"),

			code: http.StatusInternalServerError,
			body: RestError("internal error"),
		},
	}

	for i := range tcases {
		tc := tcases[i]

		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("GetDevice",
				mtest.ContextMatcher(),
				mock.AnythingOfType("string")).
				Return(tc.device, tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}

func TestSearchDevices(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Request      *http.Request
		DeviceFilter model.DeviceFilter
		AppDevices   []model.Device
		AppError     error

		// Response
		StatusCode int
		Headers    http.Header
		Body       []byte
	}{{
		Name: "ok, single device",

		Request: func() *http.Request {
			body := []byte(`{"id":"123456789012345678901234"}`)
			req, _ := http.NewRequest("POST",
				"http://localhost/api/management/v2/devauth/devices/search",
				bytes.NewReader(body),
			)
			req.Header.Add("X-MEN-RequestID", "test")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", rtest.DEFAULT_AUTH)
			return req
		}(),
		DeviceFilter: model.DeviceFilter{
			IDs: []string{"123456789012345678901234"},
		},
		AppDevices: []model.Device{{
			Id:        "123456789012345678901234",
			Status:    "accepted",
			CreatedTs: time.Unix(1606942069, 0),
		}},

		StatusCode: http.StatusOK,
		Headers:    http.Header{"X-Men-Requestid": []string{"test"}},
		Body: func() []byte {
			dev := []model.Device{{
				Id:        "123456789012345678901234",
				Status:    "accepted",
				CreatedTs: time.Unix(1606942069, 0),
			}}
			b, _ := json.Marshal(dev)
			return b
		}(),
	}, {
		Name: "ok, multiple devices",

		Request: func() *http.Request {
			body := []byte(`{"status":"accepted"}`)
			req, _ := http.NewRequest("POST",
				"http://localhost/api/management/v2/devauth/devices/search?per_page=1&page=2",
				bytes.NewReader(body),
			)
			req.Header.Add("X-MEN-RequestID", "test")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", rtest.DEFAULT_AUTH)
			return req
		}(),
		DeviceFilter: model.DeviceFilter{
			Status: []string{"accepted"},
		},
		AppDevices: []model.Device{{
			Id:        "123456789012345678901234",
			Status:    "accepted",
			CreatedTs: time.Unix(1606942069, 0),
		}, {
			Id:        "123456789012345678901235",
			Status:    "accepted",
			CreatedTs: time.Unix(1606942069, 0),
		}, {
			Id:        "123456789012345678901236",
			Status:    "accepted",
			CreatedTs: time.Unix(1606942069, 0),
		}},

		StatusCode: http.StatusOK,
		Headers: http.Header{"X-Men-Requestid": []string{"test"},
			"Link": []string{
				"</api/management/v2/devauth/devices/search?page=1&per_page=1>; rel=\"first\"",
				"</api/management/v2/devauth/devices/search?page=1&per_page=1>; rel=\"prev\"",
				"</api/management/v2/devauth/devices/search?page=3&per_page=1>; rel=\"next\"",
			}},
		Body: func() []byte {
			dev := []model.Device{{
				Id:        "123456789012345678901234",
				Status:    "accepted",
				CreatedTs: time.Unix(1606942069, 0),
			}, {
				Id:        "123456789012345678901235",
				Status:    "accepted",
				CreatedTs: time.Unix(1606942069, 0),
			}}
			b, _ := json.Marshal(dev)
			return b
		}(),
	}, {
		Name: "ok, single device url-encoded post-form",

		Request: func() *http.Request {
			body := []byte(`{"id":"123456789012345678901234","status":"accepted"}`)
			req, _ := http.NewRequest("POST",
				"http://localhost/api/management/v2/devauth/devices/search",
				bytes.NewReader(body),
			)
			req.Header.Add("X-MEN-RequestID", "test")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", rtest.DEFAULT_AUTH)
			return req
		}(),
		DeviceFilter: model.DeviceFilter{
			IDs:    []string{"123456789012345678901234"},
			Status: []string{"accepted"},
		},
		AppDevices: []model.Device{{
			Id:        "123456789012345678901234",
			Status:    "accepted",
			CreatedTs: time.Unix(1606942069, 0),
		}},

		StatusCode: http.StatusOK,
		Headers:    http.Header{"X-Men-Requestid": []string{"test"}},
		Body: func() []byte {
			dev := []model.Device{{
				Id:        "123456789012345678901234",
				Status:    "accepted",
				CreatedTs: time.Unix(1606942069, 0),
			}}
			b, _ := json.Marshal(dev)
			return b
		}(),
	}, {
		Name: "error, bad paging params",

		Request: func() *http.Request {
			body := []byte(`{"id":"123456789012345678901234","status":"accepted"}`)
			req, _ := http.NewRequest("POST",
				"http://localhost/api/management/v2/devauth/devices/search?per_page=many",
				bytes.NewReader(body),
			)
			req.Header.Add("X-MEN-RequestID", "test")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", rtest.DEFAULT_AUTH)
			return req
		}(),

		StatusCode: http.StatusBadRequest,
		Headers:    http.Header{"X-Men-Requestid": []string{"test"}},
		Body: func() []byte {
			err := rest.Error{
				Err:       "invalid per_page query: \"many\"",
				RequestID: "test",
			}
			b, _ := json.Marshal(err)
			return b
		}(),
	}, {
		Name: "error, bad JSON",

		Request: func() *http.Request {
			body := []byte(`{{"id":123456789012345678901234,"status":"accepted"}`)
			req, _ := http.NewRequest("POST",
				"http://localhost/api/management/v2/devauth/devices/search",
				bytes.NewReader(body),
			)
			req.Header.Add("X-MEN-RequestID", "test")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", rtest.DEFAULT_AUTH)
			return req
		}(),

		StatusCode: http.StatusBadRequest,
		Headers:    http.Header{"X-Men-Requestid": []string{"test"}},
		Body: func() []byte {
			err := rest.Error{
				Err: "api: malformed request body: " +
					"invalid character '{' looking for " +
					"beginning of object key string",
				RequestID: "test",
			}
			b, _ := json.Marshal(err)
			return b
		}(),
	},
	/* This test case is irrelevant because the Content-Type is checked in the
	   contenttype middleware.
		{
			Name: "error, bad Content-Type",

			Request: func() *http.Request {
				body := []byte(`id: 123456789012345678901234\nstatus: accepted`)
				req, _ := http.NewRequest("POST",
					"http://localhost/api/management/v2/devauth/devices/search",
					bytes.NewReader(body),
				)
				req.Header.Add("X-MEN-RequestID", "test")
				req.Header.Set("Content-Type", "application/yæml")
				req.Header.Set("Authorization", rtest.DEFAULT_AUTH)
				return req
			}(),

			StatusCode: http.StatusUnsupportedMediaType,
			Headers:    http.Header{"X-Men-Requestid": []string{"test"}},
			Body: func() []byte {
				err := rest.Error{
					Err: "Bad Content-Type or charset, expected 'application/json'",
				}
				b, _ := json.Marshal(err)
				return b
			}(),
		},
	*/
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := &mocks.App{}
			if tc.AppDevices != nil || tc.AppError != nil {
				app.On("GetDevices",
					mtest.ContextMatcher(),
					mock.AnythingOfType("uint"),
					mock.AnythingOfType("uint"),
					tc.DeviceFilter,
				).Return(tc.AppDevices, tc.AppError)
			}
			apih := makeMockApiHandler(t, app, nil)

			w := runTestRequest(t, apih, tc.Request, tc.StatusCode, string(tc.Body))

			rspHeader := w.Header()
			for key, values := range tc.Headers {
				if assert.Contains(t, rspHeader, key) {
					for _, value := range values {
						assert.Contains(t, rspHeader[key], value)
					}
				}
			}
		})
	}
}

func TestApiV2GetDevices(t *testing.T) {
	t.Parallel()

	devs := []model.Device{
		{
			Id:     "id1",
			Status: model.DevStatusPending,
		},
		{
			Id:     "id2",
			Status: model.DevStatusRejected,
		},
		{
			Id:     "id3",
			Status: model.DevStatusRejected,
		},
		{
			Id:     "id4",
			Status: model.DevStatusAccepted,
		},
		{
			Id:     "id5",
			Status: model.DevStatusPreauth,
		},
	}

	tcases := map[string]struct {
		req     *http.Request
		code    int
		body    string
		devices []model.Device
		err     error
		skip    uint
		limit   uint
	}{
		"ok": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices",
				Auth:   true,
			}),
			code:    http.StatusOK,
			devices: devs,
			err:     nil,
			skip:    0,
			limit:   rest.PerPageDefault + 1,
			body:    string(asJSON(devs)),
		},
		"no devices": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices",
				Auth:   true,
			}),
			code:    http.StatusOK,
			devices: []model.Device{},
			skip:    0,
			limit:   rest.PerPageDefault + 1,
			err:     nil,
			body:    "[]",
		},
		// this test does not check if the devices were skipped
		// it is only checking if endpoint limits number of devices in the response
		"limit number of devices": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices?page=2&per_page=2",
				Auth:   true,
			}),
			devices: devs[:2],
			skip:    2,
			limit:   3,
			code:    http.StatusOK,
			// reqquested 2 devices per page, so expect only 2
			body: string(asJSON(devs[:2])),
		},
		"internal error": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/devices?page=2&per_page=2",
				Auth:   true,
			}),
			skip:  2,
			limit: 3,
			code:  http.StatusInternalServerError,
			err:   errors.New("failed"),
			body:  RestError("internal error"),
		},
	}

	for name := range tcases {
		tc := tcases[name]
		t.Run(fmt.Sprintf("tc %s", name), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("GetDevices",
				mtest.ContextMatcher(),
				tc.skip, tc.limit, mock.AnythingOfType("model.DeviceFilter")).Return(
				tc.devices, tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}

func asJSON(sth interface{}) []byte {
	data, _ := json.Marshal(sth)
	return data
}

func TestApiV2DevAuthDecommissionDevice(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req  *http.Request
		code int
		body string
		err  error
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo",
				Auth:   true,
			}),
			code: http.StatusNoContent,
			err:  nil,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo",
				Auth:   true,
			}),
			code: http.StatusNotFound,
			err:  store.ErrDevNotFound,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo",
				Auth:   true,
			}),
			code: http.StatusInternalServerError,
			body: RestError("internal error"),
			err:  errors.New("some error that will only be logged"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("DecommissionDevice",
				mtest.ContextMatcher(),
				mock.AnythingOfType("string")).
				Return(tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}

func TestApiDevAuthPutTenantLimit(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req    *http.Request
		code   int
		body   string
		tenant string
		limit  model.Limit
		err    error
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/max_devices",
				Body: map[string]int{
					"limit": 123,
				},
			}),
			limit: model.Limit{
				Name:  model.LimitMaxDeviceCount,
				Value: 123,
			},
			tenant: "foo",
			code:   http.StatusNoContent,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/max_devices",
				Body:   []string{"garbage"},
			}),
			code: http.StatusBadRequest,
			body: RestError("failed to decode limit request: json: cannot unmarshal array into Go value of type http.LimitValue"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/bogus-limit",
				Body: map[string]int{
					"limit": 123,
				},
			}),
			code: http.StatusBadRequest,
			body: RestError("unsupported limit bogus-limit"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "PUT",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/max_devices",
				Body: map[string]int{
					"limit": 123,
				},
			}),
			tenant: "foo",
			limit:  model.Limit{Name: model.LimitMaxDeviceCount, Value: 123},
			code:   http.StatusInternalServerError,
			err:    errors.New("failed"),
			body:   RestError("internal error"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("SetTenantLimit",
				mtest.ContextMatcher(),
				tc.tenant,
				tc.limit).
				Return(tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}

func TestApiDevAuthDeleteTenantLimit(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req    *http.Request
		code   int
		body   string
		tenant string
		limit  string
		err    error
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/max_devices",
			}),
			limit:  model.LimitMaxDeviceCount,
			tenant: "foo",
			code:   http.StatusNoContent,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/bogus-limit",
			}),
			code: http.StatusBadRequest,
			body: RestError("unsupported limit bogus-limit"),
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/internal/v1/devauth/tenant/foo/limits/max_devices",
			}),
			tenant: "foo",
			limit:  model.LimitMaxDeviceCount,
			code:   http.StatusInternalServerError,
			err:    errors.New("failed"),
			body:   RestError("internal error"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("DeleteTenantLimit",
				mtest.ContextMatcher(),
				tc.tenant,
				tc.limit).
				Return(tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}

func TestApiV2DevAuthGetLimit(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		limit string

		daLimit *model.Limit
		daErr   error

		code int
		body string
	}{
		{
			limit: "max_devices",

			daLimit: &model.Limit{
				Name:  model.LimitMaxDeviceCount,
				Value: 123,
			},
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				LimitValue{
					Limit: 123,
				},
			)),
		},
		{
			limit: "bogus",

			code: http.StatusBadRequest,
			body: RestError("unsupported limit bogus"),
		},
		{
			limit: "max_devices",

			daLimit: nil,
			daErr:   errors.New("generic error"),

			code: http.StatusInternalServerError,
			body: RestError("internal error"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/management/v2/devauth/limits/" + tc.limit,
				Auth:   true,
			})

			da := &mocks.App{}
			da.On("GetLimit",
				mtest.ContextMatcher(),
				tc.limit).
				Return(tc.daLimit, tc.daErr)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, req, tc.code, tc.body)
		})
	}
}

func TestApiDevAuthGetTenantLimit(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		limit    string
		tenantId string

		daLimit *model.Limit
		daErr   error

		code int
		body string
	}{
		{
			limit:    "max_devices",
			tenantId: "tenant-foo",

			daLimit: &model.Limit{
				Name:  model.LimitMaxDeviceCount,
				Value: 123,
			},
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				LimitValue{
					Limit: 123,
				},
			)),
		},
		{
			limit:    "bogus",
			tenantId: "tenant-foo",

			code: http.StatusBadRequest,
			body: RestError("unsupported limit bogus"),
		},
		{
			limit:    "max_devices",
			tenantId: "tenant-foo",

			daLimit: nil,
			daErr:   errors.New("generic error"),

			code: http.StatusInternalServerError,
			body: RestError("internal error"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path: "http://localhost/api/internal/v1/devauth/tenant/" +
					tc.tenantId +
					"/limits/" +
					tc.limit,
			})
			da := &mocks.App{}
			da.On("GetTenantLimit",
				mtest.ContextMatcher(),
				tc.limit,
				tc.tenantId).
				Return(tc.daLimit, tc.daErr)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, req, tc.code, tc.body)
		})
	}
}

func TestApiV2DevAuthGetDevicesCount(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req    *http.Request
		status string

		daCnt int
		daErr error

		code int
		body string
	}{
		{
			status: "pending",

			daCnt: 5,
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				model.Count{
					Count: 5,
				},
			)),
		},
		{
			status: "noauth",

			daCnt: 5,
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				model.Count{
					Count: 5,
				},
			)),
		},
		{
			status: "accepted",

			daCnt: 0,
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				model.Count{
					Count: 0,
				},
			)),
		},
		{
			status: "rejected",

			daCnt: 4,
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				model.Count{
					Count: 4,
				},
			)),
		},
		{
			status: model.DevStatusPreauth,

			daCnt: 7,
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				model.Count{
					Count: 7,
				},
			)),
		},
		{
			status: "",

			daCnt: 10,
			daErr: nil,

			code: http.StatusOK,
			body: string(asJSON(
				model.Count{
					Count: 10,
				},
			)),
		},
		{
			status: "bogus",

			code: http.StatusBadRequest,
			body: RestError("status must be one of: pending, accepted, rejected, preauthorized, noauth"),
		},
		{
			status: "accepted",

			daErr: errors.New("generic error"),

			code: http.StatusInternalServerError,
			body: RestError("internal error"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			url := "http://localhost/api/management/v2/devauth/devices/count"
			if tc.status != "" {
				url += "?status=" + tc.status
			}
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   url,
				Auth:   true,
			})
			da := &mocks.App{}
			da.On("GetDevCountByStatus",
				mtest.ContextMatcher(),
				tc.status).
				Return(tc.daCnt, tc.daErr)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, req, tc.code, tc.body)
		})
	}
}

func TestApiDevAuthPostTenants(t *testing.T) {
	testCases := map[string]struct {
		req        *http.Request
		devAuthErr error
		respCode   int
		respBody   string
	}{
		"ok": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   "http://localhost/api/internal/v1/devauth/tenants",
				Body:   model.NewTenant{TenantId: "foo"},
			}),
			respCode: 201,
			respBody: "",
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		da := &mocks.App{}

		da.On("ProvisionTenant",
			mock.MatchedBy(func(c context.Context) bool { return true }),
			mock.AnythingOfType("string")).Return(tc.devAuthErr)

		apih := makeMockApiHandler(t, da, nil)

		runTestRequest(t, apih, tc.req, tc.respCode, tc.respBody)
	}
}

func restError(status string) map[string]interface{} {
	return map[string]interface{}{"error": status, "request_id": "test"}
}

func TestApiV2DevAuthDeleteDeviceAuthSet(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		req  *http.Request
		code int
		body string
		err  error
	}{
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo/auth/bar",
				Auth:   true,
			}),
			code: http.StatusNoContent,
			err:  nil,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo/auth/bar",
				Auth:   true,
			}),
			code: http.StatusNotFound,
			err:  store.ErrAuthSetNotFound,
		},
		{
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   "http://localhost/api/management/v2/devauth/devices/foo/auth/bar",
				Auth:   true,
			}),
			code: http.StatusInternalServerError,
			body: RestError("internal error"),
			err:  errors.New("some error that will only be logged"),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("DeleteAuthSet",
				mtest.ContextMatcher(),
				"foo",
				"bar").
				Return(tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}

func TestApiDeleteTokens(t *testing.T) {
	t.Parallel()

	tcases := map[string]struct {
		tenantId string
		deviceId string

		devAuthErr error

		checker mt.ResponseChecker
	}{
		"ok, all tokens": {
			tenantId: "foo",
			checker: mt.NewJSONResponse(
				http.StatusNoContent,
				nil,
				nil),
		},
		"ok, device's tokens": {
			tenantId: "foo",
			deviceId: "dev-foo",
			checker: mt.NewJSONResponse(
				http.StatusNoContent,
				nil,
				nil),
		},
		"error, no tenant id": {
			deviceId: "dev-foo",
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("tenant_id must be provided")),
		},
		"error, devauth": {
			tenantId:   "foo",
			devAuthErr: errors.New("generic error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				restError("internal error")),
		},
	}

	for n := range tcases {
		tc := tcases[n]
		t.Run(fmt.Sprintf("tc %s", n), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("DeleteTokens",
				mtest.ContextMatcher(),
				tc.tenantId,
				tc.deviceId).
				Return(tc.devAuthErr)

			//make request
			url := fmt.Sprintf("http://localhost/api/internal/v1/devauth/tokens?tenant_id=%v&device_id=%v",
				tc.tenantId,
				tc.deviceId)

			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "DELETE",
				Path:   url,
			})

			apih := makeMockApiHandler(t, da, nil)
			recorded := runSimpleTestRequest(t, apih, req)
			mt.CheckHTTPResponse(t, tc.checker, recorded)
		})
	}
}

func TestApiDevAuthGetTenantDeviceStatus(t *testing.T) {
	t.Parallel()

	tcases := map[string]struct {
		tid string
		did string

		daStatus *model.Status
		daErr    error

		checker mt.ResponseChecker
	}{
		"ok": {
			tid: "foo",
			did: "bar",

			daStatus: &model.Status{Status: "accepted"},

			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				model.Status{Status: "accepted"}),
		},
		"ok: tenant id empty": {
			did: "bar",

			daStatus: &model.Status{Status: "accepted"},

			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				model.Status{Status: "accepted"}),
		},
		"error: device id empty": {
			tid: "foo",

			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				restError("device id (did) cannot be empty")),
		},
		"error: not found": {
			tid: "foo",
			did: "bar",

			daErr: devauth.ErrDeviceNotFound,

			checker: mt.NewJSONResponse(
				http.StatusNotFound,
				nil,
				restError("device not found")),
		},
		"error: generic": {
			tid: "foo",
			did: "bar",

			daErr: errors.New("generic error"),

			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				restError("internal error")),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path: fmt.Sprintf("http://localhost/api/internal/v1/devauth/tenants/%s/devices/%s/status",
					tc.tid, tc.did),
				Auth: true,
			})

			da := &mocks.App{}
			da.On("GetTenantDeviceStatus",
				mtest.ContextMatcher(),
				tc.tid,
				tc.did,
			).Return(tc.daStatus, tc.daErr)

			apih := makeMockApiHandler(t, da, nil)

			recorded := runSimpleTestRequest(t, apih, req)
			mt.CheckHTTPResponse(t, tc.checker, recorded)
		})
	}
}

func TestApiDevAuthGetTenantDeviceCount(t *testing.T) {
	t.Parallel()

	tcases := map[string]struct {
		tid    string
		status string

		count    int
		countErr error

		checker mt.ResponseChecker
	}{
		"ok": {
			tid: "foo",

			count: 1,

			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				model.Count{Count: 1}),
		},
		"ok, empty tenant ID": {
			count: 1,

			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				model.Count{Count: 1}),
		},
		"ok, with status": {
			tid:    "foo",
			status: "accepted",

			count: 1,

			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				model.Count{Count: 1}),
		},
		"error: generic": {
			tid: "foo",

			countErr: errors.New("generic error"),

			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				restError("internal error")),
		},
	}

	for i := range tcases {
		tc := tcases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path: fmt.Sprintf("http://localhost/api/internal/v1/devauth/tenants/%s/devices/count?status=%s",
					tc.tid, tc.status),
				Auth: true,
			})

			da := &mocks.App{}
			da.On("GetDevCountByStatus",
				mtest.ContextMatcher(),
				tc.status,
			).Return(tc.count, tc.countErr)

			apih := makeMockApiHandler(t, da, nil)

			recorded := runSimpleTestRequest(t, apih, req)
			mt.CheckHTTPResponse(t, tc.checker, recorded)
		})
	}
}

func ExtractHeader(hdr, val string, r *mt.Recorded) string {
	rec := r.Recorder
	for _, v := range rec.Header()[hdr] {
		if v == val {
			return v
		}
	}

	return ""
}

func TestApiGetTenantDevicesV2(t *testing.T) {
	t.Parallel()

	devs := []model.Device{
		{
			Id:     "id1",
			Status: model.DevStatusPending,
		},
		{
			Id:     "id2",
			Status: model.DevStatusRejected,
		},
		{
			Id:     "id3",
			Status: model.DevStatusRejected,
		},
		{
			Id:     "id4",
			Status: model.DevStatusAccepted,
		},
		{
			Id:     "id5",
			Status: model.DevStatusPreauth,
		},
	}

	tcases := map[string]struct {
		req       *http.Request
		code      int
		body      string
		devices   []model.Device
		err       error
		skip      uint
		limit     uint
		tenant_id string

		filterMatch interface{}
	}{
		"ok": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tenants/powerpuff123/devices",
			}),
			code:      http.StatusOK,
			devices:   devs,
			err:       nil,
			skip:      0,
			limit:     rest.PerPageDefault + 1,
			body:      string(asJSON(devs)),
			tenant_id: "powerpuff123",

			filterMatch: mock.AnythingOfType("model.DeviceFilter"),
		},
		"ok with empty tenant ID": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tenants//devices",
			}),
			code:      http.StatusOK,
			devices:   devs,
			err:       nil,
			skip:      0,
			limit:     rest.PerPageDefault + 1,
			body:      string(asJSON(devs)),
			tenant_id: "powerpuff123",

			filterMatch: mock.AnythingOfType("model.DeviceFilter"),
		},
		"ok with IDs": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tenants/powerpuff123/devices?id=id1&id=id2",
			}),
			code:      http.StatusOK,
			devices:   devs[:2],
			err:       nil,
			skip:      0,
			limit:     rest.PerPageDefault + 1,
			body:      string(asJSON(devs[:2])),
			tenant_id: "powerpuff123",

			filterMatch: mock.MatchedBy(func(filter model.DeviceFilter) bool {
				assert.Equal(t, filter.IDs, []string{"id1", "id2"})

				return true
			}),
		},
		"no devices": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tenants/powerpuff123/devices",
			}),
			code:      http.StatusOK,
			devices:   []model.Device{},
			skip:      0,
			limit:     rest.PerPageDefault + 1,
			err:       nil,
			body:      "[]",
			tenant_id: "powerpuff123",

			filterMatch: mock.AnythingOfType("model.DeviceFilter"),
		},
		// this test does not check if the devices were skipped
		// it is only checking if endpoint limits number of devices in the response
		"limit number of devices": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tenants/powerpuff123/devices?page=2&per_page=2",
			}),
			devices: devs[:2],
			skip:    2,
			limit:   3,
			code:    http.StatusOK,
			// requested 2 devices per page, so expect only 2
			body:      string(asJSON(devs[:2])),
			tenant_id: "powerpuff123",

			filterMatch: mock.AnythingOfType("model.DeviceFilter"),
		},
		"internal error": {
			req: rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/internal/v1/devauth/tenants/powerpuff123/devices?page=2&per_page=2",
			}),
			skip:      2,
			limit:     3,
			code:      http.StatusInternalServerError,
			err:       errors.New("failed"),
			body:      RestError("internal error"),
			tenant_id: "powerpuff123",

			filterMatch: mock.AnythingOfType("model.DeviceFilter"),
		},
	}

	for name := range tcases {
		tc := tcases[name]
		t.Run(fmt.Sprintf("tc %s", name), func(t *testing.T) {
			t.Parallel()

			da := &mocks.App{}
			da.On("GetDevices",
				mock.MatchedBy(func(c context.Context) bool {
					if id := identity.FromContext(c); id != nil && id.Tenant != tc.tenant_id {
						assert.FailNow(t, "Tenant ID from request mismatch", identity.FromContext(c).Tenant)
						return false
					}
					return true
				}),
				tc.skip,
				tc.limit,
				tc.filterMatch,
			).Return(tc.devices, tc.err)

			apih := makeMockApiHandler(t, da, nil)
			runTestRequest(t, apih, tc.req, tc.code, tc.body)
		})
	}
}
