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

package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSAllowedOrigins(t *testing.T) {
	// This test cannot run in parallel!
	oldChecker := wsUpgrader.CheckOrigin
	defer func() {
		wsUpgrader.CheckOrigin = oldChecker
	}()
	testCases := []struct {
		Name string

		Origins []string
		Input   string

		Result bool
	}{{
		Name: "ok, single domain",

		Origins: []string{"localhost"},
		Input:   "localhost",

		Result: true,
	}, {
		Name: "ok, list of domains",

		Origins: []string{"https://localhost", "wss://localhost", "localhost"},
		Input:   "wss://localhost",

		Result: true,
	}, {
		Name: "ok, no header/single domain",

		Origins: []string{"localhost"},

		Result: true,
	}, {
		Name: "ok, no header/list of domains",

		Origins: []string{"https://localhost", "wss://localhost", "localhost"},

		Result: true,
	}, {
		Name: "ok, allow all",

		Input:  "localhost",
		Result: true,
	}, {
		Name: "origin mismatch",

		Origins: []string{"https://localhost", "wss://localhost", "localhost"},
		Input:   "remotehost",

		Result: false,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			SetAcceptedOrigins(tc.Origins)

			req, err := http.NewRequest("GET", "http://localhost", nil)
			if err != nil {
				panic(err)
			}
			if tc.Input != "" {
				req.Header.Set("Origin", tc.Input)
			}

			assert.Equal(t, tc.Result, wsUpgrader.CheckOrigin(req))
		})
	}
}
