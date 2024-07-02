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

package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestParsePagingParameters(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		URL  url.URL

		ExpectedPage    int64
		ExpectedPerPage int64
		ExpectedError   error
	}{{
		Name: "ok",
		URL: url.URL{
			Path:     "/foobar",
			RawQuery: "page=2&per_page=32",
		},
		ExpectedPage:    2,
		ExpectedPerPage: 32,
	}, {
		Name:            "defaults",
		URL:             url.URL{Path: "/"},
		ExpectedPage:    1,
		ExpectedPerPage: PerPageDefault,
	}, {
		Name:            "error, bad page parameter",
		URL:             url.URL{Path: "/", RawQuery: "page=two"},
		ExpectedPage:    -1,
		ExpectedPerPage: -1,
		ExpectedError: errors.New(
			"invalid page query: \"two\"",
		),
	}, {
		Name:            "error, bad per_page parameter",
		URL:             url.URL{Path: "/", RawQuery: "per_page=thirty"},
		ExpectedPage:    -1,
		ExpectedPerPage: -1,
		ExpectedError: errors.New(
			"invalid per_page query: \"thirty\"",
		),
	}, {
		Name:            "error, negative page parameter",
		URL:             url.URL{Path: "/", RawQuery: "page=-12345"},
		ExpectedPage:    -1,
		ExpectedPerPage: -1,
		ExpectedError: errors.New(
			"invalid page query: " +
				"value must be a non-zero positive integer",
		),
	}, {
		Name:            "error, zero per_page parameter",
		URL:             url.URL{Path: "/", RawQuery: "per_page=0"},
		ExpectedPage:    -1,
		ExpectedPerPage: -1,
		ExpectedError: errors.New(
			"invalid per_page query: " +
				"value must be a non-zero positive integer",
		),
	}, {
		Name: "error, per_page above limit",
		URL: url.URL{
			Path:     "/",
			RawQuery: fmt.Sprintf("per_page=%d", PerPageMax+1),
		},
		ExpectedPage:    1,
		ExpectedPerPage: PerPageMax + 1,
		ExpectedError:   ErrPerPageLimit,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			req := &http.Request{
				URL: &tc.URL,
			}
			page, perPage, err := ParsePagingParameters(req)
			if tc.ExpectedError != nil {
				assert.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.ExpectedPage, page)
			assert.Equal(t, tc.ExpectedPerPage, perPage)
		})
	}
}

func TestMakePagingHeaders(t *testing.T) {
	testCases := []struct {
		Name string

		// Inputs
		URL   url.URL
		Hints *PagingHints

		// Expected
		Links []string
		Error error
	}{{
		Name:  "ok",
		URL:   url.URL{Path: "/foobar", RawQuery: "page=3&per_page=10"},
		Hints: NewPagingHints().SetTotalCount(120),

		Links: []string{
			`</foobar?page=1&per_page=10>; rel="first"`,
			`</foobar?page=2&per_page=10>; rel="prev"`,
			`</foobar?page=4&per_page=10>; rel="next"`,
			`</foobar?page=12&per_page=10>; rel="last"`,
		},
	}, {
		Name:  "ok, up-by-one",
		URL:   url.URL{Path: "/foobar", RawQuery: "page=3&per_page=10"},
		Hints: NewPagingHints().SetTotalCount(121),

		Links: []string{
			`</foobar?page=1&per_page=10>; rel="first"`,
			`</foobar?page=2&per_page=10>; rel="prev"`,
			`</foobar?page=4&per_page=10>; rel="next"`,
			`</foobar?page=13&per_page=10>; rel="last"`,
		},
	}, {
		Name: "ok, defaults",
		URL:  url.URL{Path: "/foobar"},

		Links: []string{`</foobar?page=1&per_page=20>; rel="first"`},
	}, {
		Name: "ok, default has next",
		URL:  url.URL{Path: "/foobar"},
		Hints: NewPagingHints().
			SetHasNext(true).
			SetPage(1).
			SetPerPage(20),

		Links: []string{
			`</foobar?page=1&per_page=20>; rel="first"`,
			`</foobar?page=2&per_page=20>; rel="next"`,
		},
	}, {
		Name: "error parsing paging parameters",
		URL:  url.URL{Path: "/foobar", RawQuery: "page=badvalue"},

		Error: errors.New("invalid page query: \"badvalue\""),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			req := &http.Request{
				URL: &tc.URL,
			}
			links, err := MakePagingHeaders(req, tc.Hints)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Links, links)
			}
		})
	}
}
