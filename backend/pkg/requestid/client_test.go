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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrackingApiClient(t *testing.T) {
	c := NewTrackingApiClient("1234")
	assert.NotNil(t, c)
}

func TestTrackingApiClientDo(t *testing.T) {
	s := newMockServer()
	defer s.Close()

	c := NewTrackingApiClient("1234")
	assert.NotNil(t, c)

	req, err := http.NewRequest(
		http.MethodGet, s.URL, nil)
	assert.NoError(t, err)

	rsp, err := c.Do(req)
	assert.NotNil(t, rsp)

	reqid := rsp.Header.Get(RequestIdHeader)
	assert.Equal(t, "1234", reqid)
}

// mock server pings back the Request ID header
func newMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get(RequestIdHeader)
		w.Header().Set(RequestIdHeader, hdr)
		w.WriteHeader(200)
	}))
}
