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
package orchestrator

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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	ct "github.com/mendersoftware/mender-server/services/deviceauth/client/testing"
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

func TestGetClient(t *testing.T) {
	t.Parallel()

	c := NewClient(Config{
		OrchestratorAddr: "localhost:6666",
	})
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
		ResponseBody: rest.Error{
			Err:       "internal error",
			RequestID: "test",
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
	client := NewClient(Config{OrchestratorAddr: srv.URL})
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

func TestClientReqSuccess(t *testing.T) {
	t.Parallel()

	s, rd := ct.NewMockServer(http.StatusOK, nil)
	defer s.Close()

	c := NewClient(Config{
		OrchestratorAddr: s.URL,
	})

	ctx := context.Background()

	err := c.SubmitDeviceDecommisioningJob(ctx, DecommissioningReq{})
	assert.NoError(t, err, "expected no errors")
	assert.Equal(t, DeviceDecommissioningOrchestratorUri, rd.Url.Path)
}

func TestClientReqFail(t *testing.T) {
	t.Parallel()

	s, rd := ct.NewMockServer(http.StatusBadRequest, nil)
	defer s.Close()

	c := NewClient(Config{
		OrchestratorAddr: s.URL,
	})

	ctx := context.Background()

	err := c.SubmitDeviceDecommisioningJob(ctx, DecommissioningReq{})
	assert.Error(t, err, "expected an error")
	assert.Equal(t, DeviceDecommissioningOrchestratorUri, rd.Url.Path)
}

func TestClientReqNoHost(t *testing.T) {
	t.Parallel()

	c := NewClient(Config{
		OrchestratorAddr: "http://somehost:1234",
	})

	ctx := context.Background()

	err := c.SubmitDeviceDecommisioningJob(ctx, DecommissioningReq{})

	assert.Error(t, err, "expected an error")
}

func TestSubmitDeviceLimitWarning(t *testing.T) {
	testCases := []struct {
		Name string

		CTX context.Context
		DeviceLimitWarning

		URLNoise     string
		HTTPResponse *http.Response

		Error error
	}{{
		Name: "ok",

		CTX: context.Background(),
		DeviceLimitWarning: DeviceLimitWarning{
			RecipientEmail: "user@acme.io",

			Subject: "Your approaching your device limit.",
			Body:    "Please fix...",
			BodyHTML: `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body><script>alert("Øpgrade plan pls");alert("ok thænks, bye!")</script></body>
</html>
`,
			RemainingDevices: func() *uint {
				ret := uint(2)
				return &ret
			}(),
		},

		HTTPResponse: &http.Response{
			StatusCode: http.StatusCreated,
		},
	}, {
		Name: "error, bad request argument",

		DeviceLimitWarning: DeviceLimitWarning{},

		Error: errors.New(
			`workflows: \[internal\] invalid request argument: ` +
				`invalid device limit request: ` +
				`missing parameter "to"`),
	}, {
		Name: "error, bad URL",

		DeviceLimitWarning: DeviceLimitWarning{
			RecipientEmail: "user@acme.io",

			Subject:  "Your approaching your device limit.",
			Body:     "Please fix...",
			BodyHTML: `<!DOCTYPE html><html> hello! </html>`,
			RemainingDevices: func() *uint {
				ret := uint(2)
				return &ret
			}(),
		},
		CTX:      context.Background(),
		URLNoise: "%%%",
		Error: errors.New(
			`workflows: error preparing device limit warning ` +
				`request: parse "http://[0-9.:]+%%%` +
				DeviceLimitWarningURI),
	}, {
		Name: "error, context canceled",

		DeviceLimitWarning: DeviceLimitWarning{
			RecipientEmail: "user@acme.io",

			Subject:  "Your approaching your device limit.",
			Body:     "Please fix...",
			BodyHTML: `<!DOCTYPE html><html> hello! </html>`,
			RemainingDevices: func() *uint {
				ret := uint(2)
				return &ret
			}(),
		},
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),
		Error: errors.Wrapf(context.Canceled,
			`workflows: error sending device limit warning `+
				`request: Post "http://[0-9.:].+%s"`,
			DeviceLimitWarningURI),
	}, {
		Name: "error, API error returned from workflow",

		CTX: context.Background(),
		DeviceLimitWarning: DeviceLimitWarning{
			RecipientEmail: "user@acme.io",

			Subject:  "Your approaching your device limit.",
			Body:     "Please fix...",
			BodyHTML: `<!DOCTYPE html><html>Check device limit</html>`,
			RemainingDevices: func() *uint {
				ret := uint(2)
				return &ret
			}(),
		},

		HTTPResponse: func() *http.Response {
			rsp := &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			}
			b, _ := json.Marshal(rest.Error{
				Err:       "internal error",
				RequestID: "foobar",
			})
			rsp.Body = ioutil.NopCloser(bytes.NewReader(b))
			return rsp
		}(),
		Error: errors.New("internal error"),
	}, {
		Name: "error, bad status from workflows - content not understood",

		CTX: context.Background(),
		DeviceLimitWarning: DeviceLimitWarning{
			RecipientEmail: "user@acme.io",

			Subject:  "Your approaching your device limit.",
			Body:     "Please fix...",
			BodyHTML: `<!DOCTYPE html><html>Check device limit</html>`,
			RemainingDevices: func() *uint {
				ret := uint(2)
				return &ret
			}(),
		},

		HTTPResponse: &http.Response{
			StatusCode: http.StatusInternalServerError,
		},
		Error: errors.New("workflows: unexpected HTTP response: " +
			"500 Internal Server Error"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			rspChan := make(chan *http.Response, 1)
			srv := newTestServer(rspChan, nil)

			client := NewClient(Config{
				OrchestratorAddr: srv.URL + tc.URLNoise,
				Timeout:          time.Minute,
			})
			if tc.HTTPResponse != nil {
				rspChan <- tc.HTTPResponse
			}

			err := client.SubmitDeviceLimitWarning(
				tc.CTX, tc.DeviceLimitWarning,
			)
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
