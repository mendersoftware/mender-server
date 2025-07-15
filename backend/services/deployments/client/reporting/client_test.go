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

package reporting

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
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

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

		// inventory response
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
		Name: "error, inventory unhealthy",

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: rest.Error{
			Err:       "internal error",
			RequestID: "test",
		},

		Error: errors.New("internal error"),
	}, {
		Name: "error, bad response",

		Ctx:          context.TODO(),
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
	client := NewClient("").(*client)
	client.baseURL = srv.URL
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

func TestSearch(t *testing.T) {
	t.Parallel()

	defaultCtx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	testCases := map[string]struct {
		name string

		ctx context.Context

		// Inventory response
		responseCode  int
		responseBody  interface{}
		responseCount string

		expectedDevices []model.InvDevice

		outError error
	}{
		"ok": {
			ctx:             defaultCtx,
			responseCode:    http.StatusOK,
			responseCount:   "1",
			responseBody:    []model.InvDevice{},
			expectedDevices: []model.InvDevice{},
		},
		"ko, wrong payload": {
			ctx:             defaultCtx,
			responseCode:    http.StatusOK,
			responseCount:   "1",
			responseBody:    "dummy",
			expectedDevices: []model.InvDevice{},

			outError: errors.New("error parsing search devices response: json: cannot unmarshal string into Go value of type []model.InvDevice"),
		},
		"ko, no count": {
			ctx:             defaultCtx,
			responseCode:    http.StatusOK,
			responseCount:   "xyz",
			responseBody:    []model.InvDevice{},
			expectedDevices: []model.InvDevice{},

			outError: errors.New("error parsing " + hdrTotalCount + " header: strconv.Atoi: parsing \"xyz\": invalid syntax"),
		},
		"ko, not found": {

			ctx:          context.TODO(),
			responseCode: http.StatusNotFound,

			outError: errors.New("search devices request failed with unexpected status: 404"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			responses := make(chan http.Response, 1)
			serveHTTP := func(w http.ResponseWriter, r *http.Request) {
				rsp := <-responses
				if tc.responseCount != "" {
					w.Header().Add(hdrTotalCount, tc.responseCount)
				}
				w.WriteHeader(rsp.StatusCode)
				if rsp.Body != nil {
					_, _ = io.Copy(w, rsp.Body)
				}
			}
			srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
			client := NewClient("").(*client)
			client.baseURL = srv.URL
			defer srv.Close()

			if tc.responseCode > 0 {
				rsp := http.Response{
					StatusCode: tc.responseCode,
				}
				if tc.responseBody != nil {
					b, _ := json.Marshal(tc.responseBody)
					rsp.Body = ioutil.NopCloser(bytes.NewReader(b))
				}
				responses <- rsp
			}

			expectedDevices, _, err := client.Search(tc.ctx, "foo", model.SearchParams{})

			if tc.outError != nil {
				assert.EqualError(t, err, tc.outError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDevices, expectedDevices)
			}
		})
	}
}
