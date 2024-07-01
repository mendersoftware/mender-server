// Copyright 2023 Northern.tech AS
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

package rest_utils

import (
	"net/http"
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParmStr(t *testing.T) {

	testCases := map[string]struct {
		url      string
		username string
	}{
		"test 1": {
			url:      "/test?username=demo+user@mender.io",
			username: "demo user@mender.io",
		},
		"test 2": {
			url:      "/test?username=demo%2Buser@mender.io",
			username: "demo+user@mender.io",
		},
		"test 3": {
			url:      "/test?username=demo%2Fuser@mender.io",
			username: "demo/user@mender.io",
		},
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			httpReq := test.MakeSimpleRequest("POST", tc.url, "")
			req := &rest.Request{Request: httpReq}

			value, err := ParseQueryParmStr(req, "username", false, nil)
			assert.Equal(t, tc.username, value)
			assert.Nil(t, err)
		})
	}
}

func TestMakePageLinkHdrs(t *testing.T) {
	testCases := []struct {
		Name string

		HasNext       bool
		Page, PerPage uint64
		Path          string

		Expected []string
	}{{
		Name: "First page",

		HasNext: true,
		Path:    "/root",
		Page:    1,
		PerPage: 20,

		Expected: []string{
			`</root?page=2&per_page=20>; rel="next"`,
			`</root?page=1&per_page=20>; rel="first"`,
		},
	}, {
		Name: "Second page",

		HasNext: true,
		Path:    "/root/child",
		Page:    3,
		PerPage: 10,

		Expected: []string{
			`</root/child?page=1&per_page=10>; rel="first"`,
			`</root/child?page=2&per_page=10>; rel="prev"`,
			`</root/child?page=4&per_page=10>; rel="next"`,
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req, _ := http.NewRequest(
				"GET",
				"http://localhost/"+
					strings.TrimPrefix(tc.Path, "/"),
				nil,
			)
			res := MakePageLinkHdrs(
				&rest.Request{Request: req},
				tc.Page, tc.PerPage, tc.HasNext,
			)
			for _, link := range tc.Expected {
				assert.Contains(t, res, link)
			}
		})
	}
}
