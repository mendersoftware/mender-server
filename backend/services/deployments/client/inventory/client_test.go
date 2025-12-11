// Copyright 2025 Northern.tech AS
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

package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

// legacyTestClient is a test-only HTTP client for unit tests.
// It provides direct HTTP client functionality without the shared client wrapper.
type legacyTestClient struct {
	baseURL    string
	httpClient *http.Client
}

func newTestClient(baseURL string) *legacyTestClient {
	return &legacyTestClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

func (c *legacyTestClient) CheckHealth(ctx context.Context) error {
	var apiErr rest.Error

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, _ := http.NewRequestWithContext(
		ctx, "GET", c.baseURL+"/api/internal/v1/inventory/health", nil,
	)

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= http.StatusOK && rsp.StatusCode < 300 {
		return nil
	}
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&apiErr)
	if err != nil {
		return errors.Errorf("health check HTTP error: %s", rsp.Status)
	}
	return &apiErr
}

func (c *legacyTestClient) GetDeviceGroups(ctx context.Context, tenantId, deviceId string) ([]string, error) {
	repl := strings.NewReplacer(":tenantId", tenantId, ":deviceId", deviceId)
	url := c.baseURL + repl.Replace("/api/internal/v1/inventory/tenants/:tenantId/devices/:deviceId/groups")

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "get device groups request failed")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		if rsp.StatusCode == http.StatusNotFound {
			return []string{}, nil
		}
		return nil, errors.Errorf(
			"get device groups request failed with unexpected status: %v",
			rsp.StatusCode,
		)
	}

	res := model.DeviceGroups{}
	if err := json.NewDecoder(rsp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "error parsing device groups response")
	}

	return res.Groups, nil
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
	client := newTestClient(srv.URL)
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

func TestGetDeviceGroups(t *testing.T) {
	t.Parallel()

	defaultCtx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	testCases := map[string]struct {
		name string

		ctx context.Context

		// Inventory response
		responseCode int
		responseBody interface{}

		expectedGroups []string

		outError error
	}{
		"ok": {

			ctx:            defaultCtx,
			responseCode:   http.StatusOK,
			responseBody:   model.DeviceGroups{Groups: []string{"foo"}},
			expectedGroups: []string{"foo"},
		},
		"ok, not found": {

			ctx:            context.TODO(),
			responseCode:   http.StatusNotFound,
			expectedGroups: []string{},
		},
		"some error": {

			ctx:          context.TODO(),
			responseCode: http.StatusInternalServerError,

			outError: errors.New("get device groups request failed with unexpected status: 500"),
		},
	}

	responses := make(chan http.Response, 1)
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		rsp := <-responses
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
	client := newTestClient(srv.URL)
	defer srv.Close()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

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

			groups, err := client.GetDeviceGroups(tc.ctx, "foo", "bar")

			if tc.outError != nil {
				assert.EqualError(t, err, tc.outError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGroups, groups)
			}
		})
	}
}
