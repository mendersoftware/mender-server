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
package inventory

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

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
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
	client := NewClient(srv.URL, false)
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

func TestClientSetDeviceStatus(t *testing.T) {
	cases := map[string]struct {
		tid     string
		devices []model.DeviceInventoryUpdate
		status  string

		code             int
		errCheckPrefix   bool
		doNotStartServer bool
		err              error
	}{
		"ok": {
			devices: []model.DeviceInventoryUpdate{
				{Id: "dev1"},
				{Id: "dev2"},
				{Id: "dev3"},
			},
			tid:    "tenant",
			status: "accepted",

			code: http.StatusOK,
		},
		"ok, no tenant": {
			devices: []model.DeviceInventoryUpdate{
				{Id: "dev1"},
				{Id: "dev2"},
				{Id: "dev3"},
			},
			status: "accepted",

			code: http.StatusOK,
		},
		"error: inventory": {
			devices: []model.DeviceInventoryUpdate{
				{Id: "dev1"},
				{Id: "dev2"},
				{Id: "dev3"},
			},
			status: "accepted",

			code: http.StatusBadRequest,
		},
		"error: no devices to update": {
			status: "accepted",

			err: errors.New("no devices to update"),
		},
		"error: not a valid url": {
			devices: []model.DeviceInventoryUpdate{
				{Id: "dev1"},
				{Id: "dev2"},
				{Id: "dev3"},
			},
			status: "accepted",
			tid:    "/well, leads to % no / good url/",

			errCheckPrefix: true,
			err:            errors.New("failed to create request: parse"),
		},
		"error: connection refused": {
			devices: []model.DeviceInventoryUpdate{
				{Id: "dev1"},
				{Id: "dev2"},
				{Id: "dev3"},
			},
			status: "accepted",
			tid:    "tenant",

			doNotStartServer: true,
			errCheckPrefix:   true,
			err:              errors.New("failed to create request: parse"),
		},
	}

	for d := range cases {
		tc := cases[d]
		t.Run(fmt.Sprintf("case: %s", d), func(t *testing.T) {
			t.Parallel()

			if tc.doNotStartServer {
				c := NewClient("http://this.does.not.exists/url/also/", true)
				err := c.SetDeviceStatus(context.TODO(),
					tc.tid,
					tc.devices,
					tc.status)
				assert.True(t, strings.HasPrefix(err.Error(), "failed to submit POST"))
				return
			}
			s := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(tc.code)
						if tc.code != http.StatusOK {
							return
						}

						url := urlUpdateDeviceStatus + tc.status
						url = strings.Replace(url, "#tid", tc.tid, 1)
						assert.Equal(t,
							r.URL.Path,
							url)
						assert.Equal(t, "deviceauth", r.Header.Get("X-MEN-Source"))

						defer r.Body.Close()
						_, err := ioutil.ReadAll(r.Body)
						assert.NoError(t, err)
					}))

			c := NewClient(s.URL, true)
			err := c.SetDeviceStatus(context.TODO(),
				tc.tid,
				tc.devices,
				tc.status)
			if tc.err == nil {
				if tc.code == 0 {
					if tc.errCheckPrefix {
						assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
					} else {
						assert.NoError(t, err)
					}
				}
			} else {
				if tc.errCheckPrefix {
					assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
				} else {
					assert.EqualError(t, err, tc.err.Error())
				}
				return
			}

			if tc.code == http.StatusOK {
				assert.NoError(t, err)
			} else {
				url := urlUpdateDeviceStatus + tc.status
				url = strings.Replace(url, "#tid", tc.tid, 1)
				s := fmt.Sprintf("POST %s request failed with status %d %s", s.URL+url, tc.code, http.StatusText(tc.code))
				assert.EqualError(t, err, s)
			}
		})
	}
}

func TestClientSetDeviceIdentity(t *testing.T) {
	cases := map[string]struct {
		tid     string
		did     string
		didData map[string]interface{}

		code             int
		errCheckPrefix   bool
		doNotStartServer bool
		err              error
	}{
		"ok": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",
			didData: map[string]interface{}{
				"serial":   "vcsdfgsadt7678dfswr543",
				"eMMC-CID": "d0 27 01 32 0f 59 03 ff f6 db ff ef 8a 40 40 00",
				"CPUID":    "0x8000 0008",
			},
			tid: "tenant",

			code: http.StatusOK,
		},
		"ok, no tenant": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",
			didData: map[string]interface{}{
				"serial":   "vcsdfgsadt7678dfswr543",
				"eMMC-CID": "d0 27 01 32 0f 59 03 ff f6 db ff ef 8a 40 40 00",
				"CPUID":    "0x8000 0008",
			},

			code: http.StatusOK,
		},
		"error: inventory": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",
			didData: map[string]interface{}{
				"serial":   "vcsdfgsadt7678dfswr543",
				"eMMC-CID": "d0 27 01 32 0f 59 03 ff f6 db ff ef 8a 40 40 00",
				"CPUID":    "0x8000 0008",
			},

			code: http.StatusBadRequest,
		},
		"error: status attribute is reserved": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",
			didData: map[string]interface{}{
				"status": "accepted",
			},

			err: errors.New("no attributes to update"),
		},
		"error: no device id": {

			err: errors.New("device id is needed"),
		},
		"error: no attributes": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",

			err: errors.New("no attributes to update"),
		},
		"error: not a valid url": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",
			didData: map[string]interface{}{
				"serial":   "vcsdfgsadt7678dfswr543",
				"eMMC-CID": "d0 27 01 32 0f 59 03 ff f6 db ff ef 8a 40 40 00",
				"CPUID":    "0x8000 0008",
			},
			tid: "/well, leads to % no / good url/",

			errCheckPrefix: true,
			err:            errors.New("failed to create request: parse"),
		},
		"error: connection refused": {
			did: "dsfgr32r23-dfgst34gsdf-34gs-sdgf34",
			didData: map[string]interface{}{
				"serial":   "vcsdfgsadt7678dfswr543",
				"eMMC-CID": "d0 27 01 32 0f 59 03 ff f6 db ff ef 8a 40 40 00",
				"CPUID":    "0x8000 0008",
			},
			tid: "tenant",

			doNotStartServer: true,
			errCheckPrefix:   true,
			err:              errors.New("failed to create request: parse"),
		},
	}

	for d := range cases {
		tc := cases[d]
		t.Run(fmt.Sprintf("case: %s", d), func(t *testing.T) {
			t.Parallel()

			if tc.doNotStartServer {
				c := NewClient("http://this.does.not.exists/url/also/", true)
				err := c.SetDeviceIdentity(context.TODO(),
					tc.tid,
					tc.did,
					tc.didData)
				assert.True(t, strings.HasPrefix(err.Error(), "failed to submit PATCH"))
				return
			}
			s := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(tc.code)
						if tc.code != http.StatusOK {
							return
						}

						url := urlSetDeviceAttribute
						url = strings.Replace(url, "#tid", tc.tid, 1)
						url = strings.Replace(url, "#did", tc.did, 1)
						url = strings.Replace(url, "#scope", "identity", 1)
						assert.Equal(t,
							r.URL.Path,
							url)
						assert.Equal(t, "deviceauth", r.Header.Get("X-MEN-Source"))

						defer r.Body.Close()
						_, err := ioutil.ReadAll(r.Body)
						assert.NoError(t, err)
					}))

			c := NewClient(s.URL, true)
			err := c.SetDeviceIdentity(context.TODO(),
				tc.tid,
				tc.did,
				tc.didData)
			if tc.err == nil {
				if tc.code == 0 {
					if tc.errCheckPrefix {
						assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
					} else {
						assert.NoError(t, err)
					}
				}
			} else {
				if tc.errCheckPrefix {
					assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
				} else {
					assert.EqualError(t, err, tc.err.Error())
				}
				return
			}

			if tc.code == http.StatusOK {
				assert.NoError(t, err)
			} else {
				url := urlSetDeviceAttribute
				url = strings.Replace(url, "#tid", tc.tid, 1)
				url = strings.Replace(url, "#did", tc.did, 1)
				url = strings.Replace(url, "#scope", "identity", 1)
				s := fmt.Sprintf("PATCH %s request failed with status %d %s", s.URL+url, tc.code, http.StatusText(tc.code))
				assert.EqualError(t, err, s)
			}
		})
	}
}
