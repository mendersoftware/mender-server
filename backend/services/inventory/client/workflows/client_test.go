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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

func mockServerReindex(t *testing.T, tenant, device, reqid string, code int) (*httptest.Server, error) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if code != http.StatusOK {
			w.WriteHeader(code)
			return
		}
		defer r.Body.Close()

		request := []ReindexWorkflow{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		assert.NoError(t, err)

		assert.Equal(t, reqid, request[0].RequestID)
		assert.Equal(t, tenant, request[0].TenantID)
		assert.Equal(t, device, request[0].DeviceID)
		assert.Equal(t, ServiceInventory, request[0].Service)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}

	srv := httptest.NewServer(http.HandlerFunc(h))
	return srv, nil
}

func TestReindex(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		tenant string
		device string
		reqid  string

		url  string
		code int

		err error
	}{
		{
			name:   "ok",
			tenant: "tenant1",
			device: "device2",
			reqid:  "reqid1",

			code: http.StatusOK,
		},
		{
			name:   "error, connection refused",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			url: "http://127.0.0.1:12345",
			err: errors.New(`workflows: failed to submit reindex job: Post "http://127.0.0.1:12345/api/v1/workflow/reindex_reporting/batch": dial tcp 127.0.0.1:12345: connect: connection refused`),
		},
		{
			name:   "error, 404",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			code: http.StatusNotFound,
			err:  errors.New(`workflows: workflow "reindex_reporting" not defined`),
		},
		{
			name:   "error, 500",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			code: http.StatusInternalServerError,
			err:  errors.New(`workflows: unexpected HTTP status from workflows service: 500 Internal Server Error`),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv, err := mockServerReindex(t, tc.tenant, tc.device, tc.reqid, tc.code)
			assert.NoError(t, err)

			defer srv.Close()

			ctx := context.Background()
			ctx = requestid.WithContext(ctx, tc.reqid)
			ctx = identity.WithContext(ctx,
				&identity.Identity{
					Tenant: tc.tenant,
				})

			url := tc.url
			if url == "" {
				url = srv.URL
			}
			client := NewClient(url)

			err = client.StartReindex(ctx, []model.DeviceID{model.DeviceID(tc.device)})
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
	client := NewClient(srv.URL, ClientOptions{Client: &http.Client{}}).(*client)
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
