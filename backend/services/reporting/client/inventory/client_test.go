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

package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

// legacyTestClient is a test-only HTTP client for unit tests.
type legacyTestClient struct {
	client  *http.Client
	urlBase string
}

func newTestClient(urlBase string) *legacyTestClient {
	return &legacyTestClient{
		client:  &http.Client{},
		urlBase: urlBase,
	}
}

func (c *legacyTestClient) GetDevices(
	ctx context.Context,
	tid string,
	deviceIDs []string,
) ([]Device, error) {
	perPage := uint(len(deviceIDs))
	getReq := &GetDevsReq{
		DeviceIDs: deviceIDs,
		Page:      1,
		PerPage:   perPage,
	}

	body, err := json.Marshal(getReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to serialize get devices request")
	}

	rd := bytes.NewReader(body)

	url := c.urlBase + "/api/internal/v2/inventory/tenants/" + tid + "/filters/search"

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, rd)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to submit %s %s", req.Method, req.URL)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(
			"%s %s request failed with status %v", req.Method, req.URL, rsp.Status)
	}

	dec := json.NewDecoder(rsp.Body)
	var invDevs []Device
	if err = dec.Decode(&invDevs); err != nil {
		return nil, errors.Wrap(err, "failed to parse request body")
	}

	return invDevs, nil
}

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
			req.Body = io.NopCloser(bodyClone)
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

func TestGetDevices(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX      context.Context
		TenantID string
		DeviceID []string

		URLNoise     string
		ResponseCode int
		ResponseBody interface{}

		Error error
	}{{
		Name: "ok, no devices",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: []Device{},
	}, {
		Name: "ok",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: []string{
			"9acfe595-78ff-456a-843a-0fa08bfd7c7a",
			"c5e37ef5-160e-401a-aec3-9dbef94855c0",
		},

		ResponseCode: http.StatusOK,
		ResponseBody: []Device{{
			ID: DeviceID("9acfe595-78ff-456a-843a-0fa08bfd7c7a"),
			Attributes: DeviceAttributes{{
				Name:  "foo",
				Value: "bar",
				Scope: "baz",
			}},
			UpdatedTs: time.Now().Add(-time.Minute).UTC().Round(0),
		}, {
			ID: DeviceID("c5e37ef5-160e-401a-aec3-9dbef94855c0"),
			Attributes: DeviceAttributes{{
				Name:  "lorem",
				Value: "ipsum",
				Scope: "questionmark",
			}},
			UpdatedTs: time.Now().Add(-time.Minute * 5).UTC().Round(0),
		}},
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		Error: context.Canceled,
	}, {
		Name:     "error, nil context",
		CTX:      context.Background(),
		URLNoise: "#%%%",

		Error: errors.New("failed to create request"),
	}, {
		Name: "error, invalid response schema",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: []byte("bad response"),
		Error:        errors.New("failed to parse request body"),
	}, {
		Name: "error, unexpected status code",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusInternalServerError,
		ResponseBody: rest.Error{Err: "something went wrong..."},
		Error:        errors.New(`^POST [A-Za-z:0-9/\.]+ request failed with status 500`),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			rspChan := make(chan *http.Response, 1)
			srv := newTestServer(rspChan, nil)
			defer srv.Close()

			client := newTestClient(srv.URL + tc.URLNoise)

			rsp := &http.Response{
				StatusCode: tc.ResponseCode,
			}

			switch typ := tc.ResponseBody.(type) {
			case []Device:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case rest.Error:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case []byte:
				rsp.Body = io.NopCloser(bytes.NewReader(typ))

			case nil:
				// pass

			default:
				panic("[PROG ERR] invalid ResponseBody type")
			}
			rspChan <- rsp
			devs, err := client.GetDevices(tc.CTX, tc.TenantID, tc.DeviceID)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error message does not match expected pattern",
					)
				}
			} else {
				assert.NoError(t, err)
				if typ, ok := tc.ResponseBody.([]Device); ok {
					assert.Equal(t, typ, devs)
				} else {
					panic("[PROG ERR] bad test case: " +
						"expected no error but response " +
						"body contains invalid schema")
				}
			}

		})
	}
}
