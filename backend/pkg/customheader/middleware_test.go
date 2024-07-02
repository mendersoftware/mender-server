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

package customheader

import (
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestCustomHeaderMiddleware(t *testing.T) {

	testCases := map[string]struct {
		Name  string
		Value string
	}{
		"empty": {},
		"no value": {
			Name: "MyName",
		},
		"both": {
			Name:  "MyName",
			Value: "Lala",
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {

			api := rest.NewApi()

			api.Use(&CustomHeaderMiddleware{
				HeaderName:  tc.Name,
				HeaderValue: tc.Value,
			})

			api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			}))

			handler := api.MakeHandler()

			req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
			recorded := test.RunRequest(t, handler, req)
			recorded.CodeIs(200)
			recorded.ContentTypeIsJson()
			recorded.HeaderIs(tc.Name, tc.Value)
		})
	}
}
