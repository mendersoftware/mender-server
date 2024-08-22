// Copyright 2024 Northern.tech AS
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

package netutils

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPFromXFFDepth(_t *testing.T) {

	type testCase struct {
		Request    *http.Request
		ProxyDepth int

		Expected net.IP
	}
	for name, _tc := range map[string]testCase{
		"no proxy": {
			Request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				return req
			}(),
			ProxyDepth: 0,

			Expected: net.IPv4(127, 0, 0, 1),
		},
		"simple one proxy": {
			Request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				req.Header.Add(headerXForwardedFor, "127.0.0.2")
				return req
			}(),
			ProxyDepth: 1,

			Expected: net.IPv4(127, 0, 0, 2),
		},
		"multiple xff headers": {
			Request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				req.Header.Add(headerXForwardedFor, "127.0.0.5")
				req.Header.Add(headerXForwardedFor, "127.0.0.4, 127.0.0.3")
				req.Header.Add(headerXForwardedFor, "127.0.0.2")
				return req
			}(),
			ProxyDepth: 2,

			Expected: net.IPv4(127, 0, 0, 3),
		},
		"no xff header": {
			Request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				return req
			}(),
			ProxyDepth: 1,

			Expected: nil,
		},
		"scan past xff": {
			Request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				req.Header.Add(headerXForwardedFor, "127.0.0.2")
				return req
			}(),
			ProxyDepth: 2,

			Expected: nil,
		},
	} {
		tc := _tc
		_t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := GetIPFromXFFDepth(tc.Request, tc.ProxyDepth)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
