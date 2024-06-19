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
package testing

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type TestReqData struct {
	ReqBody []byte
	Headers http.Header
	Err     error
	Url     *url.URL
}

// return mock http server returning status code 'status'
func NewMockServer(status int, body []byte) (*httptest.Server, *TestReqData) {
	rdata := &TestReqData{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		rdata.ReqBody, rdata.Err = ioutil.ReadAll(r.Body)
		rdata.Headers = r.Header
		rdata.Url = r.URL
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	return srv, rdata
}
