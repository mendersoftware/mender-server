// Copyright 2022 Northern.tech AS
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
package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	common "github.com/mendersoftware/mender-server/services/iot-manager/client"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

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
		ResponseBody: rest.Error{
			Err:       "internal error",
			RequestID: "test",
		},

		Error: errors.New("internal error"),
	}, {
		Name: "error, bad response",

		Ctx: context.TODO(),

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: "foobar",

		Error: errors.New("health check HTTP error: 503 Service Unavailable"),
	}}

	responses := make(chan http.Response, 1)
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		rsp := <-responses
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
	client := NewClient(srv.URL)
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestProvisionExternalDevice(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX            context.Context
		DeviceID       string
		Config         map[string]string
		RoundTripError error

		URLNoise string

		ResponseCode int
		Error        error
	}{{
		Name: "ok",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "123456789012345678901234",
		}),
		DeviceID: "60131e78-5c31-43bf-9fab-2aaa3b422d13",
		Config: map[string]string{
			"foo": "bar",
			"bar": "baz",
		},

		ResponseCode: http.StatusOK,
	}, {
		Name: "ok, no tenant id",

		CTX:      context.Background(),
		DeviceID: "60131e78-5c31-43bf-9fab-2aaa3b422d13",
		Config: map[string]string{
			"foo": "bar",
			"bar": "baz",
		},

		ResponseCode: http.StatusOK,
	}, {
		Name: "error/bad status code",

		CTX:      context.Background(),
		DeviceID: "60131e78-5c31-43bf-9fab-2aaa3b422d13",
		Config: map[string]string{
			"foo": "bar",
			"bar": "baz",
		},

		ResponseCode: http.StatusServiceUnavailable,
		Error:        common.NewHTTPError(http.StatusServiceUnavailable),
	}, {
		Name: "error/round trip error",

		CTX:      context.Background(),
		DeviceID: "60131e78-5c31-43bf-9fab-2aaa3b422d13",
		Config: map[string]string{
			"foo": "bar",
			"bar": "baz",
		},
		RoundTripError: errors.New("internal error"),

		Error: errors.New("workflows: failed to execute request:.*internal error"),
	}, {
		Name:     "error/fail to prepare request",
		URLNoise: "%%%",

		Error: errors.New("workflows: failed to prepare request"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			htClient := &http.Client{Transport: roundTripperFunc(func(
				r *http.Request,
			) (*http.Response, error) {
				defer r.Body.Close()
				if tc.RoundTripError != nil {
					return nil, tc.RoundTripError
				}
				var req map[string]interface{}
				dec := json.NewDecoder(r.Body)
				err := dec.Decode(&req)
				if assert.NoError(t, err) {
					IShouldContinue := true
					for _, key := range []string{
						"tenant_id", "device_id",
						"request_id", "configuration",
					} {
						IShouldContinue = IShouldContinue &&
							assert.Contains(t, req, key)
					}
					if IShouldContinue {
						var (
							tenantID  string
							requestID string
						)
						if id := identity.FromContext(
							r.Context(),
						); id != nil {
							tenantID = id.Tenant
						}
						requestID = requestid.FromContext(
							r.Context(),
						)
						assert.Equal(t,
							tenantID,
							req["tenant_id"],
						)
						assert.Equal(t,
							requestID,
							req["request_id"],
						)
						assert.Equal(t,
							tc.DeviceID,
							req["device_id"],
						)
						// convert tc.Config to json map
						tcConfig := make(
							map[string]interface{},
							len(tc.Config),
						)
						for k, v := range tc.Config {
							tcConfig[k] = v
						}
						assert.EqualValues(t,
							tcConfig,
							req["configuration"],
						)

					}
				}

				w.WriteHeader(tc.ResponseCode)
				return w.Result(), nil
			})}
			opts := NewOptions(nil).
				SetClient(htClient)
			client := NewClient("http://localhost:6969"+tc.URLNoise, opts)

			err := client.ProvisionExternalDevice(tc.CTX, tc.DeviceID, tc.Config)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
