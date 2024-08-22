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
package requestid

import (
	"net/http"
)

type ApiRequester interface {
	Do(r *http.Request) (*http.Response, error)
}

// TrackingApiClient wrapper for http.Client
// for sending http requests to outside services with a given request id
type TrackingApiClient struct {
	http.Client
	reqid string
}

func NewTrackingApiClient(reqid string) *TrackingApiClient {
	return &TrackingApiClient{
		http.Client{},
		reqid,
	}
}

// do send a request with a request id
func (a *TrackingApiClient) Do(r *http.Request) (*http.Response, error) {
	if r.Header.Get(RequestIdHeader) == "" {
		r.Header.Set(RequestIdHeader, a.reqid)
	}
	return a.Client.Do(r)
}
