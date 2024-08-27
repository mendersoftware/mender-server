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

package devauth

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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func NewEchoServer(
	rspChan <-chan *http.Response,
	reqChan chan<- *http.Request,
) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		rsp, ok := <-rspChan
		if !ok {
			return
		}
		hdr := w.Header()
		for key, values := range rsp.Header {
			for _, value := range values {
				hdr.Add(key, value)
			}
		}
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			io.Copy(w, rsp.Body)
			rsp.Body.Close()
		}
		select {
		case reqChan <- r:

		default:

		}
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		DevauthAddress string
		Error          error
	}{{
		Name:           "ok",
		DevauthAddress: "docker.mender.io:1234",
	}, {
		Name:           "error, empty url",
		DevauthAddress: "",

		Error: errors.New("DevauthAddress: cannot be blank."),
	}, {
		Name:           "error, invalid url",
		DevauthAddress: "%/",

		Error: errors.New("DevauthAddress: invalid URL format."),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := Config{
				DevauthAddress: tc.DevauthAddress,
			}
			c, err := NewClient(config)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				addr := strings.TrimRight(tc.DevauthAddress, "/")
				if !strings.Contains(addr, "://") {
					addr = "http://" + addr
				}
				if assert.NoError(t, err) {
					assert.Equal(t,
						addr,
						c.(*client).uri,
					)
				}
			}
		})
	}
}

func TestPing(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX context.Context

		Response *http.Response
		Error    error
	}{{
		Name: "ok",

		Response: &http.Response{
			StatusCode: 204,
		},
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, cancel := context.WithDeadline(context.TODO(), time.Now())
			cancel()
			return ctx
		}(),

		Error: errors.Errorf(
			"error checking deviceauth liveliness:.*%s",
			context.DeadlineExceeded,
		),
	}, {
		Name: "error, bad status code",

		CTX: context.Background(),

		Response: &http.Response{
			StatusCode: http.StatusTeapot,
		},

		Error: errors.Errorf(
			"received bad status code from deviceauth liveliness probe: %d %s",
			http.StatusTeapot, http.StatusText(http.StatusTeapot),
		),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			rspChan := make(chan *http.Response, 1)
			reqChan := make(chan *http.Request, 1)
			defer close(rspChan)
			defer close(reqChan)

			server := NewEchoServer(rspChan, reqChan)
			defer server.Close()

			config := Config{
				DevauthAddress: server.URL,
			}
			client, err := NewClient(config)
			require.NoError(t, err)

			// Push response from server
			select {
			case rspChan <- tc.Response:
			default:
				panic("[TEST ERROR]: Blocked on response channel.")
			}

			err = client.Ping(tc.CTX)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDevices(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX       context.Context
		TenantID  string
		DeviceIDs []string

		Response *http.Response
		Result   []Device
		Error    error
	}{{
		Name: "ok",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "9e87c3c0-5393-4f25-8c71-5639a264e970",
		}),
		TenantID: "9e87c3c0-5393-4f25-8c71-5639a264e970",
		DeviceIDs: []string{
			"ba6e7ae4-679b-471a-8cb3-55df03b21767",
			"ba6e7ae4-679b-471a-8cb3-55df03b21768",
		},
		Response: func() *http.Response {
			var body = &bytes.Buffer{}
			encoder := json.NewEncoder(body)
			err := encoder.Encode([]Device{{
				ID:     "ba6e7ae4-679b-471a-8cb3-55df03b21767",
				Status: "accepted",
			}})
			if err != nil {
				panic(fmt.Sprintf("[TEST ERROR]: %s", err))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(body),
			}
		}(),
		Result: []Device{{
			ID:     "ba6e7ae4-679b-471a-8cb3-55df03b21767",
			Status: "accepted",
		}},
	}, {
		Name: "error, canceled context",
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),
		Error: errors.New("devauth: error performing request"),
	}, {
		Name: "error, bad URL",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "%%%%%%%%-####-$$$$-^^^^-&&&&&&&&&&&&",
		}),
		TenantID:  "%%%%%%%%-####-$$$$-^^^^-&&&&&&&&&&&&",
		DeviceIDs: []string{"%%%%%%%%-!!!!-$$$$-^^^^-############"},

		Error: errors.New(`devauth: error preparing request`),
	}, {
		Name: "error, bogus response from devauth",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "9e87c3c0-5393-4f25-8c71-5639a264e970",
		}),
		TenantID:  "9e87c3c0-5393-4f25-8c71-5639a264e970",
		DeviceIDs: []string{"ba6e7ae4-679b-471a-8cb3-55df03b21767"},

		Response: func() *http.Response {
			var body = &bytes.Buffer{}
			_, err := body.Write([]byte("bogus"))
			if err != nil {
				panic(fmt.Sprintf("[TEST ERROR]: %s", err))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(body),
			}
		}(),
		Error: errors.New("devauth: error decoding HTTP response body"),
	}, {
		Name: "error, rest.Error from server",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "9e87c3c0-5393-4f25-8c71-5639a264e970",
		}),
		TenantID:  "9e87c3c0-5393-4f25-8c71-5639a264e970",
		DeviceIDs: []string{"ba6e7ae4-679b-471a-8cb3-55df03b21767"},
		Response: func() *http.Response {
			var body = &bytes.Buffer{}
			encoder := json.NewEncoder(body)
			err := encoder.Encode(rest.Error{
				Err: "internal server error",
			})
			if err != nil {
				panic(fmt.Sprintf("[TEST ERROR]: %s", err))
			}
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(body),
			}
		}(),
		Error: errors.New(
			"client: unexpected status code from API: 500: internal server error",
		),
	}, {
		Name: "error, unexpected error response",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "9e87c3c0-5393-4f25-8c71-5639a264e970",
		}),
		TenantID:  "9e87c3c0-5393-4f25-8c71-5639a264e970",
		DeviceIDs: []string{"ba6e7ae4-679b-471a-8cb3-55df03b21767"},
		Response: &http.Response{
			StatusCode: http.StatusInternalServerError,
		},
		Error: errors.New("client: unexpected status code from API: 500"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			rspChan := make(chan *http.Response, 1)
			reqChan := make(chan *http.Request, 1)
			defer close(rspChan)
			defer close(reqChan)

			server := NewEchoServer(rspChan, reqChan)
			defer server.Close()

			config := Config{
				DevauthAddress: server.URL,
			}
			client, err := NewClient(config)
			require.NoError(t, err)

			// Push response from server
			select {
			case rspChan <- tc.Response:
			default:
				panic("[TEST ERROR]: Blocked on response channel.")
			}

			dev, err := client.GetDevices(tc.CTX, tc.DeviceIDs)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Contains(t,
						err.Error(),
						tc.Error.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
				if tc.Result != nil {
					assert.Equal(t, tc.Result, dev)
				} else {
					assert.Nil(t, dev)
				}
			}
			select {
			case req := <-reqChan:
				repl := strings.NewReplacer(
					":tenant", tc.TenantID,
				)
				assert.Equal(t,
					repl.Replace(URIInternalDevices),
					req.URL.Path,
				)
			default:
			}
		})
	}
}
