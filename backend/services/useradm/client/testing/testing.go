// Copyright 2021 Northern.tech AS
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
package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type TestReqData struct {
	Url    *url.URL
	Method string
}

// return mock http server returning status code 'status' and response 'body'
func NewMockServer(status int, body interface{}) (*httptest.Server, *TestReqData) {
	rdata := &TestReqData{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		rdata.Url = r.URL
		rdata.Method = r.Method
		json, err := json.Marshal(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(status)
		if body != nil {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(json)
		}
	}))
	return srv, rdata
}
