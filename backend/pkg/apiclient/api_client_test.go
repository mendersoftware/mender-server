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
package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/requestid"
)

func TestApiClient(t *testing.T) {

	c := HttpApi{}

	// request that came to test server
	var inreq *http.Request

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inreq = r
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r, _ := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	ctx := r.Context()
	ctx = requestid.WithContext(ctx, "123-456")
	ctx = ctxhttpheader.WithContext(ctx,
		http.Header{
			"Authorization":     []string{"Bearer of-bad-news"},
			"X-My-First-Header": []string{"none"},
			"No-Override":       []string{"override"},
		},
		"Authorization", "No-Override")

	// make sure that the client will not override already-set headers
	r.Header.Add("No-Override", "original")

	_, err := c.Do(r.WithContext(ctx))
	assert.NoError(t, err)

	assert.NotNil(t, inreq)
	assert.Equal(t, "Bearer of-bad-news", inreq.Header.Get("Authorization"))
	assert.Equal(t, "123-456", inreq.Header.Get(requestid.RequestIdHeader))
	assert.Equal(t, "original", inreq.Header.Get("No-Override"))
}
