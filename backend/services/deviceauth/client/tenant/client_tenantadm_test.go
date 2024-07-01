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
package tenant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/pkg/rest_utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	ct "github.com/mendersoftware/mender-server/services/deviceauth/client/testing"
	"github.com/mendersoftware/mender-server/pkg/ratelimits"
)

const testServerAddrRegex = `https?://(([0-9]{1,3}\.?){4}|localhost)(:[0-9]{1,5})?`

// newTestServer creates a new mock server that responds with the responses
// pushed onto the rspChan and pushes any requests received onto reqChan if
// the requests are consumed in the other end.
func newTestServer(
	rspChan <-chan *http.Response,
	reqChan chan<- *http.Request,
) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var rsp *http.Response
		select {
		case rsp = <-rspChan:
		default:
			panic("[PROG ERR] I don't know what to respond!")
		}
		if reqChan != nil {
			bodyClone := bytes.NewBuffer(nil)
			_, _ = io.Copy(bodyClone, r.Body)
			req := r.Clone(context.TODO())
			req.Body = ioutil.NopCloser(bodyClone)
			select {
			case reqChan <- req:
				// Only push request if test function is
				// popping from the channel.
			default:
			}
		}
		hdrs := w.Header()
		for k, v := range rsp.Header {
			for _, vv := range v {
				hdrs.Add(k, vv)
			}
		}
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestClientGet(t *testing.T) {
	t.Parallel()

	c := NewClient(Config{TenantAdmAddr: "http://foo"})
	assert.NotNil(t, c)
}

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	expiredCtx, cancel := context.WithDeadline(
		context.TODO(), time.Now().Add(-1*time.Second))
	defer cancel()
	defaultCtx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	testCases := []struct {
		Name string

		Ctx context.Context

		// Workflows response
		ResponseCode int
		ResponseBody interface{}

		Error error
	}{{
		Name: "ok",

		Ctx:          defaultCtx,
		ResponseCode: http.StatusOK,
	}, {
		Name: "error, expired deadline",

		Ctx:   expiredCtx,
		Error: errors.New(context.DeadlineExceeded.Error()),
	}, {
		Name: "error, workflows unhealthy",

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: rest_utils.ApiError{
			Err:   "internal error",
			ReqId: "test",
		},

		Error: errors.New("internal error"),
	}, {
		Name: "error, bad response",

		Ctx: context.TODO(),

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: "potato",

		Error: errors.New("health check HTTP error: 503 Service Unavailable"),
	}}

	responses := make(chan http.Response, 1)
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		rsp := <-responses
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
	client := NewClient(Config{TenantAdmAddr: srv.URL})
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			if tc.ResponseCode > 0 {
				rsp := http.Response{
					StatusCode: tc.ResponseCode,
				}
				if tc.ResponseBody != nil {
					b, _ := json.Marshal(tc.ResponseBody)
					rsp.Body = ioutil.NopCloser(bytes.NewReader(b))
				}
				responses <- rsp
			}

			err := client.CheckHealth(tc.Ctx)

			if tc.Error != nil {
				assert.Contains(t, err.Error(), tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientVerifyToken(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		tadmStatus int
		tadmBody   interface{}
		token      string

		tenant *Tenant
		err    error
	}{
		{
			tadmStatus: http.StatusBadRequest,
			err:        errors.New("token verification request returned unexpected status 400"),
		},
		{
			// try some bogus tadmStatus
			tadmStatus: http.StatusNotAcceptable,
			err:        errors.New("token verification request returned unexpected status 406"),
		},
		{
			// token verified ok
			tadmStatus: http.StatusOK,
			tadmBody: &Tenant{
				ID:     "foo",
				Name:   "foo-name",
				Status: "active",
				Plan:   "enterprise",
				ApiLimits: TenantApiLimits{
					MgmtLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
					DeviceLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
				},
			},
			tenant: &Tenant{
				ID:     "foo",
				Name:   "foo-name",
				Status: "active",
				Plan:   "enterprise",
				ApiLimits: TenantApiLimits{
					MgmtLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
					DeviceLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
				},
			},
		},
		{
			tadmStatus: http.StatusUnauthorized,
			tadmBody:   restError("account suspended"),
			err:        errors.New("tenant token verification failed: account suspended"),
		},
	}

	for i := range tcs {
		tc := tcs[i]
		t.Run(fmt.Sprintf("status %v", tc.tadmStatus), func(t *testing.T) {
			t.Parallel()

			var body []byte

			body, err := json.Marshal(tc.tadmBody)
			assert.NoError(t, err)

			s, rd := ct.NewMockServer(tc.tadmStatus, body)

			c := NewClient(Config{
				TenantAdmAddr: s.URL,
			})

			tenant, err := c.VerifyToken(context.Background(), tc.token)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, TenantVerifyUri, rd.Url.Path)
				assert.Equal(t, tc.tenant, tenant)
			}
			s.Close()
		})
	}
}

func TestClientGetTenant(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		tid        string
		tadmStatus int
		tadmBody   interface{}
		token      string

		tenant *Tenant
		err    error
	}{
		{
			tid:        "foo",
			tadmStatus: http.StatusOK,
			tadmBody: &Tenant{
				ID:     "foo",
				Name:   "tenant-foo",
				Status: "active",
				Plan:   "enterprise",
				ApiLimits: TenantApiLimits{
					MgmtLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
					DeviceLimits: ratelimits.ApiLimits{
						ApiQuota: ratelimits.ApiQuota{
							MaxCalls:    100,
							IntervalSec: 60,
						},
						ApiBursts: []ratelimits.ApiBurst{
							{
								Uri:            "/foo",
								MinIntervalSec: 5,
							},
						},
					},
				},
			},
			tenant: &Tenant{
				ID:     "foo",
				Name:   "tenant-foo",
				Status: "active",
				Plan:   "enterprise",
				ApiLimits: TenantApiLimits{
					MgmtLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
					DeviceLimits: ratelimits.ApiLimits{
						ApiQuota: ratelimits.ApiQuota{
							MaxCalls:    100,
							IntervalSec: 60,
						},
						ApiBursts: []ratelimits.ApiBurst{
							{
								Uri:            "/foo",
								MinIntervalSec: 5,
							},
						},
					},
				},
			},
		},
		{
			tadmStatus: http.StatusNotFound,
			tadmBody:   restError(""),
			tenant:     nil,
			err:        nil,
		},
		{
			tadmStatus: http.StatusInternalServerError,
			tadmBody:   restError("internal error"),
			tenant:     nil,
			err:        errors.New("getting tenant resulted in unexpected code: 500"),
		},
	}

	for i := range tcs {
		tc := tcs[i]
		t.Run(fmt.Sprintf("status %v", tc.tadmStatus), func(t *testing.T) {
			t.Parallel()

			var body []byte

			body, err := json.Marshal(tc.tadmBody)
			assert.NoError(t, err)

			s, rd := ct.NewMockServer(tc.tadmStatus, body)

			c := NewClient(Config{
				TenantAdmAddr: s.URL,
			})

			tenant, err := c.GetTenant(context.Background(), tc.tid)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				repl := strings.NewReplacer("#tid", tc.tid)
				uri := repl.Replace(TenantGetUri)
				assert.Equal(t, uri, rd.Url.Path)
				assert.Equal(t, tc.tenant, tenant)
			}
			s.Close()
		})
	}
}

func restError(msg string) interface{} {
	return map[string]interface{}{"error": msg, "request_id": "test"}
}

func TestGetTenantUsers(t *testing.T) {
	testCases := []struct {
		Name string

		CTX      context.Context
		TenantID string

		URLNoise     string
		HTTPResponse *http.Response

		Users []User
		Error error
	}{{
		Name: "ok",

		CTX:      context.Background(),
		TenantID: "foobar",

		HTTPResponse: func() *http.Response {
			rsp := &http.Response{
				StatusCode: http.StatusOK,
			}
			b, _ := json.Marshal([]User{{
				ID:       "123456789012345678901234",
				Email:    "foo@bar.com",
				TenantID: "foobar",
			}})
			body := ioutil.NopCloser(bytes.NewReader(b))
			rsp.Body = body
			return rsp
		}(),

		Users: []User{{
			ID:       "123456789012345678901234",
			Email:    "foo@bar.com",
			TenantID: "foobar",
		}},
	}, {
		Name: "error, tenant id cannot be empty",

		Error: errors.New(`tenantadm: \[internal\] bad argument ` +
			`tenantID: cannot be empty`,
		),
	}, {
		Name: "error, bad server url",

		TenantID: "foobar",
		URLNoise: "%%%",
		CTX:      context.Background(),

		Error: errors.New(`tenantadm: failed to prepare request: ` +
			`parse "http://(([0-9]{1,3}\.?){4}|localhost):[0-9]+%%%` +
			TenantUsersURI + `": invalid port ":[0-9]+%%%" after host`),
	}, {
		Name: "error, nil context",

		TenantID: "foobar",

		Error: errors.New(`tenantadm: failed to prepare request: ` +
			`net/http: nil Context`,
		),
	}, {
		Name: "error, context already canceled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),
		TenantID: "foobar",
		Error: errors.Wrap(context.Canceled,
			`tenantadm: error sending user request: `+
				`Get "`+testServerAddrRegex+TenantUsersURI+
				`\?tenant_id=foobar"`),
	}, {
		Name: "error, api error from server",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",

		HTTPResponse: func() *http.Response {
			rsp := &http.Response{
				StatusCode: http.StatusBadRequest,
			}
			b, _ := json.Marshal(rest_utils.ApiError{
				Err:   "internal error",
				ReqId: "test",
			})
			body := ioutil.NopCloser(bytes.NewReader(b))
			rsp.Body = body
			return rsp
		}(),
		Error: errors.New(`tenantadm: HTTP error \(400 Bad Request\) ` +
			`on user request: internal error`),
	}, {
		Name: "error, unknown error from server",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",

		HTTPResponse: &http.Response{
			StatusCode: http.StatusInternalServerError,
		},
		Error: errors.New(`tenantadm: unexpected HTTP status: ` +
			`500 Internal Server Error`),
	}, {
		Name: "error, glitchy response from tenantadm",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",

		HTTPResponse: func() *http.Response {
			rsp := &http.Response{
				StatusCode: http.StatusOK,
			}
			body := ioutil.NopCloser(bytes.NewReader([]byte(
				`plain text body`,
			)))
			rsp.Body = body
			return rsp
		}(),
		Error: errors.New(
			`tenantadm: error decoding response payload: ` +
				`invalid character 'p' looking for beginning ` +
				`of value`,
		),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			rspChan := make(chan *http.Response, 1)
			reqChan := make(chan *http.Request, 1)
			srv := newTestServer(rspChan, reqChan)
			client := NewClient(Config{
				TenantAdmAddr: srv.URL + tc.URLNoise,
			})

			if tc.HTTPResponse != nil {
				rspChan <- tc.HTTPResponse
			}

			users, err := client.GetTenantUsers(tc.CTX, tc.TenantID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
					)
				}
			} else {
				req := <-reqChan
				assert.Equal(t,
					tc.TenantID,
					req.URL.Query().Get("tenant_id"),
				)
				assert.NoError(t, err)
				assert.Equal(t, tc.Users, users)
			}
		})
	}
}
