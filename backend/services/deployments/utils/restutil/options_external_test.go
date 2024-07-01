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
package restutil_test

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"

	. "github.com/mendersoftware/mender-server/services/deployments/utils/restutil"
)

func TestOptionsHandle(t *testing.T) {

	t.Parallel()

	router, err := rest.MakeRouter(rest.Options("/r", NewOptionsHandler(http.MethodGet, http.MethodGet)))
	if err != nil {
		t.FailNow()
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest(http.MethodOptions, "http://1.2.3.4/r", nil))

	recorded.CodeIs(http.StatusOK)

	if len(recorded.Recorder.Header()[HttpHeaderAllow]) != 2 {
		t.FailNow()
	}

	for _, method := range recorded.Recorder.Header()[HttpHeaderAllow] {
		switch method {
		case http.MethodGet:
			continue
		case http.MethodOptions:
			continue
		default:
			t.FailNow()
		}
	}
}

type MockResponseWriter struct {
	methods []string
}

func (m *MockResponseWriter) Header() http.Header                      { return nil }
func (m *MockResponseWriter) WriteJson(v interface{}) error            { return nil }
func (m *MockResponseWriter) EncodeJson(v interface{}) ([]byte, error) { return nil, nil }
func (m *MockResponseWriter) WriteHeader(int)                          {}

func NewMockOptionsHandler(methods ...string) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		sort.Strings(methods)
		mockWriter, _ := w.(*MockResponseWriter)
		mockWriter.methods = methods
	}
}

type RouteList []*rest.Route

func (slice RouteList) Less(i, j int) bool {
	return slice[i].HttpMethod+slice[i].PathExp < slice[j].HttpMethod+slice[j].PathExp
}

func (slice RouteList) Len() int {
	return len(slice)
}

func (slice RouteList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func TestAutogenOptionsRoutes(t *testing.T) {

	t.Parallel()

	testList := []struct {
		out []*rest.Route
		in  []*rest.Route
	}{
		{nil, nil},
		{[]*rest.Route{}, []*rest.Route{}},
		{
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Options("/path", NewMockOptionsHandler("GET")),
			},
			[]*rest.Route{
				rest.Get("/path", nil),
			},
		},
		{
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Post("/path", nil),
				rest.Options("/path", NewMockOptionsHandler("GET", "POST")),
			},
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Post("/path", nil),
			},
		},
		{
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Options("/path", NewMockOptionsHandler("GET")),
				rest.Post("/path/path", nil),
				rest.Options("/path/path", NewMockOptionsHandler("POST")),
			},
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Post("/path/path", nil),
			},
		},
		{
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Put("/path", nil),
				rest.Options("/path", NewMockOptionsHandler("GET", "PUT")),
				rest.Post("/path/path", nil),
				rest.Options("/path/path", NewMockOptionsHandler("POST")),
			},
			[]*rest.Route{
				rest.Get("/path", nil),
				rest.Post("/path/path", nil),
				rest.Put("/path", nil),
			},
		},
	}

	for _, test := range testList {
		out := RouteList(AutogenOptionsRoutes(NewMockOptionsHandler, test.in...))

		if len(test.out) != len(out) {
			t.FailNow()
		}

		// Search requires sorted input
		sort.Sort(out)

		for _, expectedRoute := range test.out {

			idx := sort.Search(len(out), func(i int) bool {
				if out[i].HttpMethod+out[i].PathExp >= expectedRoute.HttpMethod+expectedRoute.PathExp {
					return true
				}
				return false
			})

			if idx == len(out) {
				t.FailNow()
			}

			if expectedRoute.HttpMethod == "OPTIONS" {
				exp := &MockResponseWriter{}
				expectedRoute.Func(exp, nil)

				recived := &MockResponseWriter{}
				out[idx].Func(recived, nil)

				if !reflect.DeepEqual(exp.methods, recived.methods) {
					t.FailNow()
				}
			}

		}
	}
}
