// Copyright 2025 Northern.tech AS
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
	"net/http"
	"net/url"
)

const (
	HeaderForwardedURI    = "X-Forwarded-Uri"
	HeaderForwardedHost   = "X-Forwarded-Host"
	HeaderForwardedMethod = "X-Forwarded-Method"
)

// RewriteForwardedRequest makes a shallow clone of request and replaces
// the URL and Method with X-Forwarded-* headers.
func RewriteForwardedRequest(request *http.Request) *http.Request {
	if request == nil {
		return nil
	}
	newRequest := new(http.Request)
	*newRequest = *request
	uri := request.Header.Get(HeaderForwardedURI)
	if uri != "" {
		var err error
		newRequest.URL, err = url.ParseRequestURI(uri)
		if err != nil {
			panic(err)
		}
		newRequest.RequestURI = uri
	} else {
		newRequest.URL = new(url.URL)
		*newRequest.URL = *request.URL
	}
	host := request.Header.Get(HeaderForwardedHost)
	if host != "" {
		newRequest.URL.Host = host
	}
	method := request.Header.Get(HeaderForwardedMethod)
	if method != "" {
		newRequest.Method = method
	}
	return newRequest
}
